package crawler

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kbariotis/go-discover/internal/cache"
	"github.com/kbariotis/go-discover/internal/model"
	"github.com/kbariotis/go-discover/internal/provider"
	"github.com/kbariotis/go-discover/internal/queue"
	"github.com/kbariotis/go-discover/internal/store"
)

// Crawler is our main orchestrating service
type Crawler struct {
	followerPollInterval time.Duration

	graphStore      store.GraphStore
	suggestionStore store.SuggestionStore // Rename because it includes all SQL store
	cache           cache.Cache
	provider        provider.Provider

	userOnboardingQueue       queue.Queue
	userFolloweeQueue         queue.Queue
	suggestionExtractionQueue queue.Queue
	userQueue                 queue.Queue
	repositoryQueue           queue.Queue
}

// New constructs a Github crawler
func New(
	followerPollInterval time.Duration,
	graphStore store.GraphStore,
	suggestionStore store.SuggestionStore,
	cache cache.Cache,
	provider provider.Provider,
	suggestionExtractionQueue queue.Queue,
	userOnboardingQueue queue.Queue,
	userFolloweeQueue queue.Queue,
	userQueue queue.Queue,
	repositoryQueue queue.Queue,
) (*Crawler, error) {

	crw := &Crawler{
		graphStore:                graphStore,
		suggestionStore:           suggestionStore,
		cache:                     cache,
		provider:                  provider,
		followerPollInterval:      followerPollInterval,
		suggestionExtractionQueue: suggestionExtractionQueue,
		userOnboardingQueue:       userOnboardingQueue,
		userFolloweeQueue:         userFolloweeQueue,
		userQueue:                 userQueue,
		repositoryQueue:           repositoryQueue,
	}

	return crw, nil
}

// processRegisteredUsers processes our own followers
func (c *Crawler) processRegisteredUsers() error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.processRegisteredUsers",
	})

	logger.Info("processing registered users")

	// get users and push them to the userOnboarding queue
	users, err := c.suggestionStore.GetAllUsers()
	if err != nil {
		return errors.Wrap(err, "could not retrieve users")
	}

	for _, user := range users {
		logger.
			WithField("user", user).
			Debug("got user, pushing to userOnboardingQueue")

		userOnboardingTask := &model.UserOnboardingTask{
			Name: user.Name,
		}

		if err := c.userOnboardingQueue.Push(userOnboardingTask); err != nil {
			return errors.Wrap(err, "could not add user task to queue")
		}
	}

	return nil
}

// handleSuggestionExtractionTask extracts suggestions for each user
func (c *Crawler) handleSuggestionExtractionTask(task *model.SuggestionExtractionTask) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleSuggestionExtractionTask",
	})

	logger.Info("extracting suggestions")

	// get users and push them to the suggestionExtraction queue
	user, err := c.suggestionStore.GetUser(task.UserName)
	if err != nil {
		return errors.Wrap(err, "could not retrieve user")
	}

	suggestion, err := c.graphStore.GetUserSuggestion(user)
	if err != nil {
		return errors.Wrap(err, "could not extract suggestions")
	}

	if err := c.suggestionStore.PutSuggestion(suggestion); err != nil {
		return errors.Wrap(err, "could not put suggestion")
	}

	return nil
}

// extractSuggestions feed the suggestionExtractionQueue
func (c *Crawler) extractSuggestions() error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleSuggestionExtractionTask",
	})

	logger.Info("processing registered users")

	// get users and push them to the suggestionExtraction queue
	users, err := c.suggestionStore.GetAllUsers()
	if err != nil {
		return errors.Wrap(err, "could not retrieve users")
	}

	for _, user := range users {
		logger.
			WithField("user", user).
			Debug("got user, pushing to suggestionExtractionQueue")

		suggestionExtractionTask := &model.SuggestionExtractionTask{
			UserName: user.Name,
		}

		if err := c.suggestionExtractionQueue.Push(suggestionExtractionTask); err != nil {
			return errors.Wrap(err, "could not add user task to queue")
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

	// upsert the user to the c.graphStore
	user := &model.User{
		Name:      task.Name,
		Followees: followees,
	}

	if err := c.graphStore.PutUser(user); err != nil {
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

	// check if user is in the cache
	if err := c.cache.LockUser(task.Name); err != nil {
		if err == cache.ErrAlreadyLocked {
			logger.Info("User's cached, skipping")
			return nil
		}

		return errors.Wrap(err, "could not cache user")
	}

	// follow back user
	// if err := c.provider.FollowUser(ctx, task.Name); err != nil {
	// TODO return errors.Wrap(err, "could not follow back user")
	// }

	// fetch user's starred repos
	stars, err := c.provider.GetUserStars(ctx, task.Name)
	if err != nil {
		return errors.Wrap(err, "could not get user's stars")
	}

	// push all starred repos to repository queue
	for _, star := range stars {
		c.repositoryQueue.Push(&model.RepositoryTask{
			Name: star.Repository,
		})
	}

	// TODO fetch user's repositories

	// upsert the user to the c.graphStore
	user := &model.User{
		Name:  task.Name,
		Stars: stars,
	}

	if err := c.graphStore.PutUser(user); err != nil {
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
	ctx := context.Background()

	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleRepositoryTask",
		"task":   task,
	})

	logger.Info("handling model.RepositoryTask")

	// check if repository is in the cache
	if err := c.cache.LockRepository(task.Name); err != nil {
		if err == cache.ErrAlreadyLocked {
			logger.Info("Repository's cached, skipping")
			return nil
		}
		return errors.Wrap(err, "could not cache repository")
	}

	// get repository
	repository, err := c.provider.GetRepository(ctx, task.Name)
	if err != nil {
		return errors.Wrap(err, "could not get repository")
	}

	// upsert the repository
	if err := c.graphStore.PutRepository(repository); err != nil {
		return errors.Wrap(err, "could not store repository")
	}

	return nil
}

// Start crawling
func (c *Crawler) Start(ctx context.Context) error {
	cctx, _ := context.WithCancel(ctx)
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.Start",
	})

	suggestionExtractionTasks := make(chan *model.SuggestionExtractionTask, 10000)
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

	// pop tasks from suggestionExtractionTasks and push them to a local channel
	go func() {
		logger.Info("starting to pop tasks from suggestionExtractionTasks")
		for {
			task, _ := c.suggestionExtractionQueue.Pop() // TODO handle error
			if task == nil {
				time.Sleep(time.Second)
				continue
			}
			if okTask, ok := task.(*model.SuggestionExtractionTask); ok {
				suggestionExtractionTasks <- okTask
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

	followerPollTicker := time.NewTicker(c.followerPollInterval)
	extractSuggestionsTicker := time.NewTicker(time.Hour * 24 * 7)

	for {
		select {
		case <-cctx.Done():
			return nil

		case <-extractSuggestionsTicker.C:
			if err := c.extractSuggestions(); err != nil {
				logger.WithError(err).Warn("extractSuggestions failed")
			}

		case <-followerPollTicker.C:
			if err := c.processRegisteredUsers(); err != nil {
				logger.WithError(err).Warn("processRegisteredUsers failed")
			}

		case task := <-userOnboardingTasks:
			if err := c.handleUserOnboardingTask(task); err != nil {
				logger.WithError(err).Warn("failed to handle model.UserOnboardingTask")
			}

		case task := <-suggestionExtractionTasks:
			if err := c.handleSuggestionExtractionTask(task); err != nil {
				logger.WithError(err).Warn("failed to handle model.SuggestionExtractionTask")
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
