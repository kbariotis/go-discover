package main

import (
	"context"
	"time"

	"github.com/Financial-Times/neoism"
	"github.com/go-redis/redis"
	"github.com/google/go-github/v25/github"
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	_ "github.com/jinzhu/gorm/dialects/sqlite" // required for sqlite

	"github.com/kbariotis/go-discover/internal/cache"
	"github.com/kbariotis/go-discover/internal/config"
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
		"logger":  "cmd/crawler",
		"version": version.Version,
		"gitSHA":  version.GitSHA,
	})

	ctx := context.Background()

	logger.Debug("loading configuration")
	cfg, err := config.LoadConfig()
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

	userFolloweeQueue, err := queue.NewDQueue(
		"userFollowee.queue",
		cfg.QueueStoreDir,
		&model.UserFolloweeTask{},
	)
	if err != nil {
		logger.WithError(err).Fatal("could not create dqueue for userFollowee")
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
	graphDB, err := neoism.Connect(cfg.NeoHost)
	if err != nil {
		logger.WithError(err).Fatal("could not create neo client")
	}

	// create graph store
	graphStore, err := store.NewNeo(graphDB)
	if err != nil {
		logger.WithError(err).Fatal("could not create graph store")
	}

	// setup graph store indices
	if err := graphStore.SetupIndices(); err != nil {
		logger.WithError(err).Fatal("could not setup graph indices")
	}

	// connect to suggestions store db
	db, err := gorm.Open(
		cfg.SuggestionsStoreType,
		cfg.SuggestionsStoreDSN,
	)
	if err != nil {
		logger.WithError(err).Fatal("could not connect to db")
	}
	defer db.Close()

	// create suggestions store
	suggestionStore, err := store.NewSuggestionSQL(db)
	if err != nil {
		logger.WithError(err).Fatal("could not create suggestion store")
	}

	// setup suggestion db
	if err := suggestionStore.Setup(); err != nil {
		logger.WithError(err).Fatal("could not setup suggestion db")
	}

	// create redis cache
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisHost,
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	_, err = redisClient.Ping().Result()
	if err != nil {
		logger.WithError(err).Fatal("could not connect to Redis")
	}
	redis, err := cache.NewRedis(
		redisClient,
		cfg.LockUserDuration,
		cfg.LockRepositoryDuration,
	)
	if err != nil {
		logger.WithError(err).Fatal("could not create Redis cache")
	}

	// create github provider
	prv, err := provider.NewGithub(ghClient)
	if err != nil {
		logger.WithError(err).Fatal("could not construct github provider")
	}

	// create crawler
	crw, err := crawler.New(
		time.Minute*5,
		graphStore,
		suggestionStore,
		redis,
		prv,
		userOnboardingQueue,
		userFolloweeQueue,
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
