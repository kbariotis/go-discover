package store

import (
	"github.com/kbariotis/go-discover/internal/model"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . GraphStore

// GraphStore defines the interface for the graph store implementations
type GraphStore interface {
	PutRepository(*model.Repository) error
	PutUser(*model.User) error
}
