package main

import (
	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/kbariotis/go-discover/internal/api"
	"github.com/kbariotis/go-discover/internal/config"
	"github.com/kbariotis/go-discover/internal/store"
	"github.com/kbariotis/go-discover/internal/version"
)

// main initliases and starts the service
func main() {
	logger := logrus.WithFields(logrus.Fields{
		"logger":  "cmd/api",
		"version": version.Version,
		"gitSHA":  version.GitSHA,
	})

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

	// constrcut api
	api := api.NewAPI(
		suggestionStore,
		cfg.GithubClientID,
		cfg.GithubClientSecret,
		cfg.GithubCallbackURL,
	)

	// start api on the background
	api.Serve(cfg.APIBindAddress)
}
