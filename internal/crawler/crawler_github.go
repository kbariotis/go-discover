package crawler

import (
	"context"
	"net/http"
	"time"

	"github.com/google/go-github/v25/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kbariotis/go-discover/internal/queue"
	"github.com/kbariotis/go-discover/internal/store"
)

type (
	// Github crawler
	Github struct {
		followerPollInterval time.Duration

		store  store.Store
		client *github.Client

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

// NewGithub constructs a Github crawler
func NewGithub(
	followerPollInterval time.Duration,
	store store.Store,
	client *github.Client,
	userOnboardingQueue queue.Queue,
	userFollowerQueue queue.Queue,
	userQueue queue.Queue,
	repositoryQueue queue.Queue,
) (*Github, error) {

	sch := &Github{
		store:                store,
		client:               client,
		followerPollInterval: followerPollInterval,
		userOnboardingQueue:  userOnboardingQueue,
		userFollowerQueue:    userFollowerQueue,
		userQueue:            userQueue,
		repositoryQueue:      repositoryQueue,
	}

	return sch, nil
}

// processOwnFollowers processes our own followers
func (g *Github) processOwnFollowers() error {
	ctx := context.Background()

	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.processOwnFollowers",
	})

	logger.Info("processing own followers, listing own followers")

	// get own followers and push them to the userOnboarding queue
	currentPage := 1
	for currentPage != 0 {
		opts := &github.ListOptions{
			Page:    currentPage,
			PerPage: 100,
		}

		followers, res, err := g.client.Users.ListFollowers(ctx, "", opts)
		if err != nil {
			return errors.Wrap(err, "could not retrieve own followers")
		}

		logger.
			WithFields(logrus.Fields{
				"current_page":  currentPage,
				"count":         len(followers),
				"res.code":      res.StatusCode,
				"res.next_page": res.NextPage,
			}).
			Debug("got followers")

		for _, follower := range followers {
			logger.
				WithField("follower", follower.GetLogin()).
				Debug("got follower, pushing to userOnboardingQueue")

			followerTask := &UserOnboardingTask{
				Name: follower.GetLogin(),
			}

			if err := g.userFollowerQueue.Push(followerTask); err != nil {
				return errors.Wrap(err, "could not add follower task to queue")
			}
		}

		currentPage = res.NextPage
	}

	return nil
}

func (g *Github) handleUserOnboardingTask(task *UserOnboardingTask) error {
	ctx := context.Background()

	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleUserOnboardingTask",
		"task":   task,
	})

	logger.Info("handling UserOnboardingTask")

	// TODO check if we've already processed this user in last n hours

	logger.Info("listing user's followers")

	// follow back user
	res, err := g.client.Users.Follow(ctx, task.Name)
	if err != nil {
		return errors.Wrap(err, "could not follow back user")
	}

	if res.StatusCode != http.StatusOK {
		logger.
			WithFields(logrus.Fields{
				"res.status":  res.StatusCode,
				"res.headers": res.Header,
			}).
			Warn("following back a user returned a non-ok status code")
	}

	// get user's followers and push them to the userFollower queue
	currentPage := 1
	for currentPage != 0 {
		opts := &github.ListOptions{
			Page:    currentPage,
			PerPage: 100,
		}

		followers, res, err := g.client.Users.ListFollowers(ctx, "", opts)
		if err != nil {
			return errors.Wrap(err, "could not retrieve user's followers")
		}

		logger.
			WithFields(logrus.Fields{
				"current_page":  currentPage,
				"count":         len(followers),
				"res.code":      res.StatusCode,
				"res.next_page": res.NextPage,
			}).
			Debug("got followers")

		for _, follower := range followers {
			logger.
				WithField("follower", follower.GetLogin()).
				Debug("got follower, pushing to handleUserFollowerTask")

			followerTask := &UserFollowerTask{
				Name: follower.GetLogin(),
			}

			if err := g.userFollowerQueue.Push(followerTask); err != nil {
				return errors.Wrap(err, "could not add follower task to queue")
			}
		}

		currentPage = res.NextPage
	}

	return nil
}

func (g *Github) handleUserFollowerTask(task *UserFollowerTask) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleUserFollowerTask",
		"task":   task,
	})

	logger.Info("handling UserFollowerTask")

	// TODO check if we've already processed this user in last n hours

	// TODO fetch user's starred repos
	// TODO fetch user's repositories
	// TODO upsert the user to the g.store

	return nil
}

func (g *Github) handleUserTask(task *UserTask) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/handleUserTask.handleUserOnboardingTask",
		"task":   task,
	})

	logger.Info("handling UserTask")

	// TODO handle UserTask

	return nil
}

func (g *Github) handleRepositoryTask(task *RepositoryTask) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "crawler/Github.handleRepositoryTask",
		"task":   task,
	})

	logger.Info("handling RepositoryTask")

	// TODO handle RepositoryTask

	return nil
}

// Start crawling
func (g *Github) Start(ctx context.Context) error {
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
			task, _ := g.userOnboardingQueue.Pop() // TODO handle error
			if task == nil {
				// time.Sleep(time.Second)
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
			task, _ := g.userFollowerQueue.Pop() // TODO handle error
			if task == nil {
				// time.Sleep(time.Second)
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
			task, _ := g.userQueue.Pop() // TODO handle error
			if task == nil {
				// time.Sleep(time.Second)
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
			task, _ := g.repositoryQueue.Pop() // TODO handle error
			if task == nil {
				// time.Sleep(time.Second)
				continue
			}
			if okTask, ok := task.(*RepositoryTask); ok {
				repositoryTasks <- okTask
			}
		}
	}()

	// serially process events to reduce strain on the github api
	// TODO introduce workers and proper request throttling for the api
	followerPollTicker := time.NewTicker(g.followerPollInterval)
	followerPollFirstPoll := time.After(0)
	for {
		select {
		case <-cctx.Done():
			return nil

		case <-followerPollFirstPoll:
			if err := g.processOwnFollowers(); err != nil {
				return errors.Wrap(err, "first time processOwnFollowers failed")
			}

		case <-followerPollTicker.C:
			if err := g.processOwnFollowers(); err != nil {
				return errors.Wrap(err, "processOwnFollowers failed")
			}

		case task := <-userOnboardingTasks:
			if err := g.handleUserOnboardingTask(task); err != nil {
				return errors.Wrap(err, "failed to handle UserOnboardingTask")
			}

		case task := <-userFollowerTasks:
			if err := g.handleUserFollowerTask(task); err != nil {
				return errors.Wrap(err, "failed to handle UserFollowerTask")
			}

		case task := <-userTasks:
			if err := g.handleUserTask(task); err != nil {
				return errors.Wrap(err, "failed to handle UserTask")
			}

		case task := <-repositoryTasks:
			if err := g.handleRepositoryTask(task); err != nil {
				return errors.Wrap(err, "failed to handle RepositoryTask")
			}

		}
	}
}
