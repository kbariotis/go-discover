package crawler

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kbariotis/go-discover/internal/provider"
	"github.com/kbariotis/go-discover/internal/queue"
	"github.com/kbariotis/go-discover/internal/store"
)

type (
	// Crawler
	Crawler struct {
		followerPollInterval time.Duration

		store    store.Store
		provider provider.Provider

		userOnboardingQueue queue.Queue
		userFollowerQueue   queue.Queue
		userQueue           queue.Queue
		repositoryQueue     queue.Queue
	}
	// UserOnboardingTask represents a task in the userOnboarding queue
	UserOnboardingTask struct {
		Name string
	}
	// UserFollowerTask represents a task in the userFollower queue
	UserFollowerTask struct {
		Name string
	}
	// UserTask represents a task in the user queue
	UserTask struct {
		Name string
	}
	// RepositoryTask represents a task in the repository queue
	RepositoryTask struct {
		Name string
	}
)

// New constructs a Github crawler
func New(
	followerPollInterval time.Duration,
	store store.Store,
	provider provider.Provider,
	userOnboardingQueue queue.Queue,
	userFollowerQueue queue.Queue,
	userQueue queue.Queue,
	repositoryQueue queue.Queue,
) (*Crawler, error) {

	crw := &Crawler{
		store:                store,
		provider:             provider,
		followerPollInterval: followerPollInterval,
		userOnboardingQueue:  userOnboardingQueue,
		userFollowerQueue:    userFollowerQueue,
		userQueue:            userQueue,
		repositoryQueue:      repositoryQueue,
	}

	return crw, nil
}

// processOwnFollowers processes our own followers
func (c *Crawler) processOwnFollowers() error {
	ctx := context.Background()

	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.processOwnFollowers",
	})

	logger.Info("processing own followers, listing own followers")

	// get own followers and push them to the userOnboarding queue
	followers, err := c.provider.GetUserFollowers(ctx, "")
	if err != nil {
		return errors.Wrap(err, "could not retrieve own followers")
	}

	for _, follower := range followers {
		logger.
			WithField("follower", follower).
			Debug("got follower, pushing to userOnboardingQueue")

		followerTask := &UserOnboardingTask{
			Name: follower,
		}

		if err := c.userOnboardingQueue.Push(followerTask); err != nil {
			return errors.Wrap(err, "could not add follower task to queue")
		}
	}

	return nil
}

func (c *Crawler) handleUserOnboardingTask(task *UserOnboardingTask) error {
	ctx := context.Background()

	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleUserOnboardingTask",
		"task":   task,
	})

	logger.Info("handling UserOnboardingTask")

	// TODO check if we've already processed this user in last n hours

	logger.Info("listing user's followers")

	// follow back user
	if err := c.provider.FollowUser(ctx, task.Name); err != nil {
		return errors.Wrap(err, "could not follow back user")
	}

	// get user's followers and push them to the userFollower queue
	followers, err := c.provider.GetUserFollowers(ctx, task.Name)
	if err != nil {
		return errors.Wrap(err, "could not retrieve own followers")
	}

	for _, follower := range followers {
		logger.
			WithField("follower", follower).
			Debug("got follower, pushing to handleUserFollowerTask")

		followerTask := &UserFollowerTask{
			Name: follower,
		}

		if err := c.userFollowerQueue.Push(followerTask); err != nil {
			return errors.Wrap(err, "could not add follower task to queue")
		}
	}

	return nil
}

func (c *Crawler) handleUserFollowerTask(task *UserFollowerTask) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleUserFollowerTask",
		"task":   task,
	})

	logger.Info("handling UserFollowerTask")

	// TODO check if we've already processed this user in last n hours

	// TODO fetch user's starred repos

	// TODO fetch user's repositories

	// TODO upsert the user to the c.store

	return nil
}

func (c *Crawler) handleUserTask(task *UserTask) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/handleUserTask.handleUserOnboardingTask",
		"task":   task,
	})

	logger.Info("handling UserTask")

	// TODO handle UserTask

	return nil
}

func (c *Crawler) handleRepositoryTask(task *RepositoryTask) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleRepositoryTask",
		"task":   task,
	})

	logger.Info("handling RepositoryTask")

	// TODO handle RepositoryTask

	return nil
}

// Start crawling
func (c *Crawler) Start(ctx context.Context) error {
	cctx, _ := context.WithCancel(ctx)
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.Start",
	})

	userOnboardingTasks := make(chan *UserOnboardingTask, 10000)
	userFollowerTasks := make(chan *UserFollowerTask, 10000)
	userTasks := make(chan *UserTask, 10000)
	repositoryTasks := make(chan *RepositoryTask, 10000)

	// pop tasks from userOnboardingQueue and push them to a local channel
	go func() {
		logger.Info("starting to pop tasks from userOnboardingQueue")
		for {
			task, _ := c.userOnboardingQueue.Pop() // TODO handle error
			if task == nil {
				time.Sleep(time.Second)
				continue
			}
			if okTask, ok := task.(*UserOnboardingTask); ok {
				userOnboardingTasks <- okTask
			}
		}
	}()

	// pop tasks from userFollowerQueue and push them to a local channel
	go func() {
		logger.Info("starting to pop tasks from userFollowerQueue")
		for {
			task, _ := c.userFollowerQueue.Pop() // TODO handle error
			if task == nil {
				time.Sleep(time.Second)
				continue
			}
			if okTask, ok := task.(*UserFollowerTask); ok {
				userFollowerTasks <- okTask
			}
		}
	}()

	// pop tasks from userQueue and push them to a local channel
	go func() {
		logger.Info("starting to pop tasks from userQueue")
		for {
			task, _ := c.userQueue.Pop() // TODO handle error
			if task == nil {
				time.Sleep(time.Second)
				continue
			}
			if okTask, ok := task.(*UserTask); ok {
				userTasks <- okTask
			}
		}
	}()

	// pop tasks from repositoryQueue and push them to a local channel
	go func() {
		logger.Info("starting to pop tasks from repositoryQueue")
		for {
			task, _ := c.repositoryQueue.Pop() // TODO handle error
			if task == nil {
				time.Sleep(time.Second)
				continue
			}
			if okTask, ok := task.(*RepositoryTask); ok {
				repositoryTasks <- okTask
			}
		}
	}()

	// serially process events to reduce strain on the github api
	// TODO introduce workers and proper request throttling for the api
	followerPollTicker := time.NewTicker(c.followerPollInterval)
	followerPollFirstPoll := time.After(0)
	for {
		select {
		case <-cctx.Done():
			return nil

		case <-followerPollFirstPoll:
			if err := c.processOwnFollowers(); err != nil {
				return errors.Wrap(err, "first time processOwnFollowers failed")
			}

		case <-followerPollTicker.C:
			if err := c.processOwnFollowers(); err != nil {
				return errors.Wrap(err, "processOwnFollowers failed")
			}

		case task := <-userOnboardingTasks:
			if err := c.handleUserOnboardingTask(task); err != nil {
				return errors.Wrap(err, "failed to handle UserOnboardingTask")
			}

		case task := <-userFollowerTasks:
			if err := c.handleUserFollowerTask(task); err != nil {
				return errors.Wrap(err, "failed to handle UserFollowerTask")
			}

		case task := <-userTasks:
			if err := c.handleUserTask(task); err != nil {
				return errors.Wrap(err, "failed to handle UserTask")
			}

		case task := <-repositoryTasks:
			if err := c.handleRepositoryTask(task); err != nil {
				return errors.Wrap(err, "failed to handle RepositoryTask")
			}

		}
	}
}
