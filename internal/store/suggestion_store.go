package store

import (
	"github.com/kbariotis/go-discover/internal/model"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . SuggestionStore

// SuggestionStore defines the interface for the persistent store implementations
type SuggestionStore interface {
	GetAllUsers() ([]*model.User, error)
	GetUser(string) (*model.User, error)
	PutUser(*model.User) error
	GetSuggestion(uint) (*model.Suggestion, error)
	PutSuggestion(*model.Suggestion) error
}
