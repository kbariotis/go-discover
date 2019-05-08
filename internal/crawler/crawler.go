package crawler

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kbariotis/go-discover/internal/model"
	"github.com/kbariotis/go-discover/internal/provider"
	"github.com/kbariotis/go-discover/internal/queue"
	"github.com/kbariotis/go-discover/internal/store"
)

// Crawler is our main orchestrating service
type Crawler struct {
	followerPollInterval time.Duration

	store    store.Store
	provider provider.Provider

	userOnboardingQueue queue.Queue
	userFollowerQueue   queue.Queue
	userQueue           queue.Queue
	repositoryQueue     queue.Queue
}

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

		followerTask := &model.UserOnboardingTask{
			Name: follower,
		}

		if err := c.userOnboardingQueue.Push(followerTask); err != nil {
			return errors.Wrap(err, "could not add follower task to queue")
		}
	}

	return nil
}

func (c *Crawler) handleUserOnboardingTask(task *model.UserOnboardingTask) error {
	ctx := context.Background()

	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleUserOnboardingTask",
		"task":   task,
	})

	logger.Info("handling model.UserOnboardingTask")

	// TODO check if we've already processed this user in last n hours

	logger.Info("listing user's followers")

	// follow back user
	if err := c.provider.FollowUser(ctx, task.Name); err != nil {
		// TODO return errors.Wrap(err, "could not follow back user")
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

		followerTask := &model.UserFollowerTask{
			Name: follower,
		}

		if err := c.userFollowerQueue.Push(followerTask); err != nil {
			return errors.Wrap(err, "could not add follower task to queue")
		}
	}

	return nil
}

func (c *Crawler) handleUserFollowerTask(task *model.UserFollowerTask) error {
	ctx := context.Background()

	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleUserFollowerTask",
		"task":   task,
	})

	logger.Info("handling model.UserFollowerTask")

	// TODO check if we've already processed this user in last n hours

	// fetch user's starred repos
	stars, err := c.provider.GetUserStars(ctx, task.Name)
	if err != nil {
		return errors.Wrap(err, "could not get user's stars")
	}

	// fetch user's followees
	followees, err := c.provider.GetUserFollowees(ctx, task.Name)
	if err != nil {
		return errors.Wrap(err, "could not get user's followees")
	}

	// TODO fetch user's repositories

	// upsert the user to the c.store
	user := &model.User{
		Name:      task.Name,
		Stars:     stars,
		Followees: followees,
	}

	if err := c.store.PutUser(user); err != nil {
		return errors.Wrap(err, "could not persist user")
	}

	return nil
}

func (c *Crawler) handleUserTask(task *model.UserTask) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/handleUserTask.handleUserOnboardingTask",
		"task":   task,
	})

	logger.Info("handling model.UserTask")

	// TODO handle model.UserTask

	return nil
}

func (c *Crawler) handleRepositoryTask(task *model.RepositoryTask) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleRepositoryTask",
		"task":   task,
	})

	logger.Info("handling model.RepositoryTask")

	// TODO handle model.RepositoryTask

	return nil
}

// Start crawling
func (c *Crawler) Start(ctx context.Context) error {
	cctx, _ := context.WithCancel(ctx)
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.Start",
	})

	userOnboardingTasks := make(chan *model.UserOnboardingTask, 10000)
	userFollowerTasks := make(chan *model.UserFollowerTask, 10000)
	userTasks := make(chan *model.UserTask, 10000)
	repositoryTasks := make(chan *model.RepositoryTask, 10000)

	// pop tasks from userOnboardingQueue and push them to a local channel
	go func() {
		logger.Info("starting to pop tasks from userOnboardingQueue")
		for {
			task, _ := c.userOnboardingQueue.Pop() // TODO handle error
			if task == nil {
				time.Sleep(time.Second)
				continue
			}
			if okTask, ok := task.(*model.UserOnboardingTask); ok {
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
			if okTask, ok := task.(*model.UserFollowerTask); ok {
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
			if okTask, ok := task.(*model.UserTask); ok {
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
			if okTask, ok := task.(*model.RepositoryTask); ok {
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
				logger.WithError(err).Warn("first time processOwnFollowers failed")
			}

		case <-followerPollTicker.C:
			if err := c.processOwnFollowers(); err != nil {
				logger.WithError(err).Warn("processOwnFollowers failed")
			}

		case task := <-userOnboardingTasks:
			if err := c.handleUserOnboardingTask(task); err != nil {
				logger.WithError(err).Warn("failed to handle model.UserOnboardingTask")
			}

		case task := <-userFollowerTasks:
			if err := c.handleUserFollowerTask(task); err != nil {
				logger.WithError(err).Warn("failed to handle model.UserFollowerTask")
			}

		case task := <-userTasks:
			if err := c.handleUserTask(task); err != nil {
				logger.WithError(err).Warn("failed to handle model.UserTask")
			}

		case task := <-repositoryTasks:
			if err := c.handleRepositoryTask(task); err != nil {
				logger.WithError(err).Warn("failed to handle model.RepositoryTask")
			}

		}
	}
}
