package main

import (
	"time"

	"github.com/caarlos0/env"
)

// Config contains various configuration settings
type Config struct {
	LogLevel             string `env:"LOG_LEVEL" envDefault:"info"`
	GithubToken          string `env:"GITHUB_TOKEN"`
	SuggestionsStoreType string `env:"SUGGESTION_STORE_TYPE" envDefault:"sqlite3"`
	SuggestionsStoreDSN  string `env:"SUGGESTION_STORE_DSN" envDefault:"${HOME}/go-discover/suggestions.db"`
	QueueStoreDir        string `env:"QUEUE_STORE_DIR" envDefault:"${HOME}/go-discover" envExpand:"true"`
	NeoHost              string `env:"NEO4J_HOST" envDefault:"http://localhost:7474/db/data"`
	RedisHost            string `env:"REDIS_HOST" envDefault:"localhost:6379"`
	APIBindAddress       string `env:"API_BIND_ADDRESS" envDefault:"0.0.0.0:8080"`

	LockUserDuration       time.Duration `env:"LOCK_USER_DURATION" envDefault:"12h"`
	LockRepositoryDuration time.Duration `env:"LOCK_REPOSITORY_DURATION" envDefault:"24h"`
}

// loadConfig parses environment variables returning configuration
func loadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
