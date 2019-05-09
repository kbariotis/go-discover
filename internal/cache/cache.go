package cache

import (
	"github.com/kbariotis/go-discover/internal/model"
)

// Cache defines the interface for the cache implementations
type Cache interface {
	UserExists(user string) (bool, error)
	SetUser(*model.User) error
}
