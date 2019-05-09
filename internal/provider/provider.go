package provider

import (
	"context"

	"github.com/kbariotis/go-discover/internal/model"
)

// Provider represents a backend for our crawler
// Even though currenly only Github is supported this is separated to help out
// with testing.
type Provider interface {
	GetUserStars(context.Context, string) ([]model.StarredRepository, error)
	GetUserFollowers(context.Context, string) ([]string, error)
	GetUserFollowees(context.Context, string) ([]string, error)
	GetUserRepositories(context.Context, string) ([]string, error)
	GetRepository(context.Context, string) (*model.Repository, error)
	FollowUser(context.Context, string) error
}
