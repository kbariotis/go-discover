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
	userFolloweeQueue   queue.Queue
	userQueue           queue.Queue
	repositoryQueue     queue.Queue
}

// New constructs a Github crawler
func New(
	followerPollInterval time.Duration,
	store store.Store,
	provider provider.Provider,
	userOnboardingQueue queue.Queue,
	userFolloweeQueue queue.Queue,
	userQueue queue.Queue,
	repositoryQueue queue.Queue,
) (*Crawler, error) {

	crw := &Crawler{
		store:                store,
		provider:             provider,
		followerPollInterval: followerPollInterval,
		userOnboardingQueue:  userOnboardingQueue,
		userFolloweeQueue:    userFolloweeQueue,
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

		userOnboardingTask := &model.UserOnboardingTask{
			Name: follower,
		}

		if err := c.userOnboardingQueue.Push(userOnboardingTask); err != nil {
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

	// process the Bot's follower
	followeeTask := &model.UserFolloweeTask{
		Name: task.Name,
	}

	if err := c.userFolloweeQueue.Push(followeeTask); err != nil {
		return errors.Wrap(err, "could not add followee task to queue")
	}

	logger.Info("listing user's followees")

	// get user's followees and push them to the userFollowee queue
	followees, err := c.provider.GetUserFollowees(ctx, task.Name)
	if err != nil {
		return errors.Wrap(err, "could not retrieve own followees")
	}

	for _, followee := range followees {
		logger.
			WithField("followee", followee).
			Debug("got followee, pushing to handleUserFolloweeTask")

		followeeTask := &model.UserFolloweeTask{
			Name: followee,
		}

		if err := c.userFolloweeQueue.Push(followeeTask); err != nil {
			return errors.Wrap(err, "could not add followee task to queue")
		}
	}

	// upsert the user to the c.store
	user := &model.User{
		Name:      task.Name,
		Followees: followees,
	}

	if err := c.store.PutUser(user); err != nil {
		return errors.Wrap(err, "could not persist user")
	}

	return nil
}

func (c *Crawler) handleUserFolloweeTask(task *model.UserFolloweeTask) error {
	ctx := context.Background()

	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleUserFolloweeTask",
		"task":   task,
	})

	logger.Info("handling model.UserFolloweeTask")

	// TODO check if we've already processed this user in last n hours

	// follow back user
	// if err := c.provider.FollowUser(ctx, task.Name); err != nil {
	// TODO return errors.Wrap(err, "could not follow back user")
	// }

	// fetch user's starred repos
	stars, err := c.provider.GetUserStars(ctx, task.Name)
	if err != nil {
		return errors.Wrap(err, "could not get user's stars")
	}

	// TODO fetch user's repositories

	// upsert the user to the c.store
	user := &model.User{
		Name:  task.Name,
		Stars: stars,
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
	userFolloweeTasks := make(chan *model.UserFolloweeTask, 10000)
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

	// pop tasks from userFolloweeQueue and push them to a local channel
	go func() {
		logger.Info("starting to pop tasks from userFolloweeQueue")
		for {
			task, _ := c.userFolloweeQueue.Pop() // TODO handle error
			if task == nil {
				time.Sleep(time.Second)
				continue
			}
			if okTask, ok := task.(*model.UserFolloweeTask); ok {
				userFolloweeTasks <- okTask
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

		case task := <-userFolloweeTasks:
			if err := c.handleUserFolloweeTask(task); err != nil {
				logger.WithError(err).Warn("failed to handle model.UserFolloweeTask")
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
