package cache

import (
	"time"

	"github.com/go-redis/redis"
)

// Redis cache implementation
type Redis struct {
	client *redis.Client

	userLockDuration       time.Duration
	repositoryLockDuration time.Duration
}

// NewRedis constructs a new Redis cache given a Redis client
func NewRedis(
	client *redis.Client,
	lockUserDuration time.Duration,
	lockRepositoryDuration time.Duration,
) (Cache, error) {
	red := &Redis{
		client: client,
	}

	return red, nil
}

// LockUser locks a user for an x amount of time
func (red *Redis) LockUser(user string) error {
	key := "user/" + user
	duration := red.userLockDuration
	if err := red.client.Set(key, "value", duration).Err(); err != nil {
		return ErrAlreadyLocked
	}
	return nil
}

// LockRepository locks a repository for an x amount of time
func (red *Redis) LockRepository(name string) error {
	key := "repository/" + name
	duration := red.repositoryLockDuration
	if err := red.client.Set(key, "value", duration).Err(); err != nil {
		return ErrAlreadyLocked
	}
	return nil
}
