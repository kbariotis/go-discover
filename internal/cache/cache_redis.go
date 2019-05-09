package cache

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"

	"github.com/kbariotis/go-discover/internal/model"
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

// GetUser returns a user graph from red
func (red *Redis) UserExists(user string) (bool, error) {
	val, err := red.client.Get(user).Result()
	fmt.Println(val, err)
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, errors.Wrap(err, "could not find user")
	}

	return false, nil
}

// SetUser merges a user's graph in red
func (red *Redis) SetUser(user *model.User) error {
	if err := red.client.Set(user.Name, "value", time.Hour).Err(); err != nil {
		return errors.Wrap(err, "could not merge user")
	}
	return nil
}
