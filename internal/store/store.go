package store

import (
	"github.com/kbariotis/go-discover/internal/model"
)

// Store defines the interface for the store implementations
type Store interface {
	GetRepository(name string) (*model.Repository, error)
	PutRepository(*model.Repository) error

	GetUser(user string) (*model.User, error)
	PutUser(*model.User) error
}
