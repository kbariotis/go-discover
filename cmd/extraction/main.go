package main

import (
	"context"
	"time"

	"github.com/Financial-Times/neoism"
	"github.com/jinzhu/gorm"
	"github.com/mailgun/mailgun-go/v3"
	"github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/kbariotis/go-discover/internal/config"
	"github.com/kbariotis/go-discover/internal/extraction"
	"github.com/kbariotis/go-discover/internal/mailer"
	"github.com/kbariotis/go-discover/internal/model"
	"github.com/kbariotis/go-discover/internal/queue"
	"github.com/kbariotis/go-discover/internal/store"
	"github.com/kbariotis/go-discover/internal/version"
)

// main initliases and starts the service
func main() {
	logger := logrus.WithFields(logrus.Fields{
		"logger":  "cmd/extraction",
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

	suggestionExtractionQueue, err := queue.NewDQueue(
		"suggestionExtraction.queue",
		cfg.QueueStoreDir,
		&model.SuggestionExtractionTask{},
	)
	if err != nil {
		logger.WithError(err).Fatal("could not create dqueue for suggestionExtraction")
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
		"host="+cfg.SuggestionsStoreHost+" port="+cfg.SuggestionsStorePort+" user="+cfg.SuggestionsStoreUser+" dbname="+cfg.SuggestionsStoreDb+" password="+cfg.SuggestionsStorePwd+" sslmode=disable",
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

	// setup mailgun
	mg := mailgun.NewMailgun(cfg.MailgunDomain, cfg.MailgunAPIKey)
	mailer, err := mailer.NewMailgun(mg, cfg.MailSenderAddress)
	if err != nil {
		logger.WithError(err).Fatal("could not create mailer")
	}

	// create extraction
	extr, err := extraction.New(
		time.Hour*24*7,
		graphStore,
		suggestionStore,
		suggestionExtractionQueue,
		mailer,
	)
	if err != nil {
		logger.WithError(err).Fatal("could not construct extraction")
	}

	logger.Info("starting extraction")

	// start extraction
	if err := extr.Start(ctx); err != nil {
		logger.WithError(err).Fatal("github extraction processing failed")
	}
}
