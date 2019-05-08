package main

import (
	"context"
	"time"

	"github.com/Financial-Times/neoism"
	"github.com/google/go-github/v25/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/kbariotis/go-discover/internal/crawler"
	"github.com/kbariotis/go-discover/internal/model"
	"github.com/kbariotis/go-discover/internal/provider"
	"github.com/kbariotis/go-discover/internal/queue"
	"github.com/kbariotis/go-discover/internal/store"
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

	// create github client
	ghTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: cfg.GithubToken,
		},
	)

	ghClient := github.NewClient(
		oauth2.NewClient(ctx, ghTokenSource),
	)

	// create queues
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

	// create neo db
	time.Sleep(time.Second * 30)
	db, err := neoism.Connect(cfg.NeoHost)
	if err != nil {
		logger.WithError(err).Fatal("could not create neo client")
	}

	// create neo store
	neo, err := store.NewNeo(db)
	if err != nil {
		logger.WithError(err).Fatal("could not create neo store")
	}

	// create github provider
	prv, err := provider.NewGithub(ghClient)
	if err != nil {
		logger.WithError(err).Fatal("could not construct github provider")
	}

	// create crawler
	crw, err := crawler.New(
		time.Minute*5,
		neo,
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

	// start crawler
	if err := crw.Start(ctx); err != nil {
		logger.WithError(err).Fatal("github crawler processing failed")
	}
}