package cache

import (
	"time"

	"github.com/go-redis/redis"
)

// Redis cache implementation
type Redis struct {
	client *redis.Client
}

// Redis constructs a new Redis cache given a Redis client
func NewRedis(client *redis.Client) (Cache, error) {
	red := &Redis{
		client: client,
	}

	return red, nil
}

// LockUser merges a user's graph in red
func (red *Redis) LockUser(user string) error {
	if err := red.client.Set(user, "value", time.Hour).Err(); err != nil {
		return ErrAlreadyLocked
	}
	return nil
}
