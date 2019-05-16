package main

import (
	"time"

	"github.com/caarlos0/env"
)

// Config contains various configuration settings
type Config struct {
	LogLevel             string `env:"LOG_LEVEL" envDefault:"info"`
	SuggestionsStoreType string `env:"SUGGESTION_STORE_TYPE" envDefault:"sqlite3"`
	SuggestionsStoreDSN  string `env:"SUGGESTION_STORE_DSN" envDefault:"./local/suggestions.db"`
	QueueStoreDir        string `env:"QUEUE_STORE_DIR" envDefault:"./local/queues" envExpand:"true"`
	NeoHost              string `env:"NEO4J_HOST" envDefault:"http://localhost:7474/db/data"`
	RedisHost            string `env:"REDIS_HOST" envDefault:"localhost:6379"`
	APIBindAddress       string `env:"API_BIND_ADDRESS" envDefault:"0.0.0.0:8080"`

	GithubToken        string `env:"GITHUB_TOKEN"`
	GithubClientSecret string `env:"GITHUB_CLIENT_SECRET"`
	GithubClientID     string `env:"GITHUB_CLIENT_ID"`
	GithubCallbackURL  string `env:"GITHUB_CALLBACK_URL" envDefault:"http://localhost:8080/github/callback"`

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
