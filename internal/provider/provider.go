package provider

import (
	"context"
)

type StarredRepository struct {
	repository string
	starredAt  int64
}

// Provider represents a backend for our crawler
// Even though currenly only Github is supported this is separated to help out
// with testing.
type Provider interface {
	GetUserStars(context.Context, string) ([]StarredRepository, error)
	GetUserFollowers(context.Context, string) ([]string, error)
	GetUserFollowees(context.Context, string) ([]string, error)
	GetUserRepositories(context.Context, string) ([]string, error)
	FollowUser(context.Context, string) error
}
