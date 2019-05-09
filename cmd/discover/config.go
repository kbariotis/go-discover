package main

import (
	"github.com/caarlos0/env"
)

// Config contains various configuration settings
type Config struct {
	LogLevel      string `env:"LOG_LEVEL" envDefault:"info"`
	GithubToken   string `env:"GITHUB_TOKEN"`
	QueueStoreDir string `env:"QUEUE_STORE_DIR" envDefault:"${HOME}/go-discover" envExpand:"true"`
	NeoHost       string `env:"NEO4J_HOST" envDefault:"http://localhost:7474/db/data"`
	RedisHost     string `env:"REDIS_HOST" envDefault:"localhost:6379"`
}

// loadConfig parses environment variables returning configuration
func loadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
