package main

import (
	"context"
	"time"

	"github.com/google/go-github/v25/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/kbariotis/go-discover/internal/crawler"
	"github.com/kbariotis/go-discover/internal/provider"
	"github.com/kbariotis/go-discover/internal/queue"
	"github.com/kbariotis/go-discover/internal/model"
	"github.com/kbariotis/go-discover/internal/version"
)

// main initliases and starts the service
func main() {
	logger := logrus.WithFields(logrus.Fields{
		"logger":  "cmd/discover",
		"version": version.Version,
		"gitSHA":  version.GitSHA,
	})

	ctx := context.Background()

	logger.Debug("loading configuration")
	cfg, err := loadConfig()
	if err != nil {
		logger.WithError(err).Fatal("could not load configuration")
	}

	logLevel, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.WithError(err).Fatal("could not parse log level")
	}

	logrus.SetLevel(logLevel)

	ghTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: cfg.GithubToken,
		},
	)

	ghClient := github.NewClient(
		oauth2.NewClient(ctx, ghTokenSource),
	)

	userOnboardingQueue, err := queue.NewDQueue(
		"userOnboarding.queue",
		cfg.QueueStoreDir,
		&model.UserOnboardingTask{},
	)
	if err != nil {
		logger.WithError(err).Fatal("could not create dqueue for userOnboarding")
	}

	userFollowerQueue, err := queue.NewDQueue(
		"userFollower.queue",
		cfg.QueueStoreDir,
		&model.UserFollowerTask{},
	)
	if err != nil {
		logger.WithError(err).Fatal("could not create dqueue for userFollower")
	}

	userQueue, err := queue.NewDQueue(
		"user.queue",
		cfg.QueueStoreDir,
		&model.UserTask{},
	)
	if err != nil {
		logger.WithError(err).Fatal("could not create dqueue for user")
	}

	repositoryQueue, err := queue.NewDQueue(
		"repository.queue",
		cfg.QueueStoreDir,
		&model.RepositoryTask{},
	)
	if err != nil {
		logger.WithError(err).Fatal("could not create dqueue for repository")
	}

	prv, err := provider.NewGithub(ghClient)
	if err != nil {
		logger.WithError(err).Fatal("could not construct github provider")
	}

	crw, err := crawler.New(
		time.Minute*5,
		nil, // TODO use proper store implementation
		prv,
		userOnboardingQueue,
		userFollowerQueue,
		userQueue,
		repositoryQueue,
	)
	if err != nil {
		logger.WithError(err).Fatal("could not construct crawler")
	}

	logger.Info("starting crawler")

	if err := crw.Start(ctx); err != nil {
		logger.WithError(err).Fatal("github crawler processing failed")
	}
}
