package provider

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/go-github/v25/github"
	"github.com/kbariotis/go-discover/internal/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Github provider
type Github struct {
	client *github.Client
}

// NewGithub constrcuts a new Github provider
func NewGithub(client *github.Client) (Provider, error) {
	prv := &Github{
		client: client,
	}

	return prv, nil
}

// GetUserStars returns the user's starred repositories
func (g *Github) GetUserStars(ctx context.Context, name string) ([]model.StarredRepository, error) {
	logger := logrus.WithFields(logrus.Fields{
		"logger":     "providers/Github.GetUserStars",
		"user.login": name,
	})

	logger.Info("getting user's starred repositories")

	stars := []model.StarredRepository{}

	currentPage := 1
	for currentPage != 0 {
		opts := &github.ActivityListStarredOptions{
			ListOptions: github.ListOptions{
				Page:    currentPage,
				PerPage: 100,
			},
		}

		moreRepos, res, err := g.client.Activity.ListStarred(ctx, name, opts)
		if err != nil {
			return nil, errors.Wrap(err, "could not retrieve user's stars")
		}

		logger.
			WithFields(logrus.Fields{
				"current_page":  currentPage,
				"count":         len(stars),
				"res.code":      res.StatusCode,
				"res.next_page": res.NextPage,
			}).
			Debug("got stars")

		for _, repo := range moreRepos {
			stars = append(stars, model.StarredRepository{repo.Repository.GetFullName(), repo.StarredAt.Unix()})
		}

		currentPage = res.NextPage
	}

	return stars, nil
}

// GetUserFollowees returnes the user's followees
func (g *Github) GetUserFollowees(ctx context.Context, name string) ([]string, error) {
	logger := logrus.WithFields(logrus.Fields{
		"logger":     "providers/Github.GetUserFollowees",
		"user.login": name,
	})

	logger.Info("getting user's followers")

	followees := []string{}

	currentPage := 1
	for currentPage != 0 {
		opts := &github.ListOptions{
			Page:    currentPage,
			PerPage: 100,
		}

		moreFollowees, res, err := g.client.Users.ListFollowing(ctx, name, opts)
		if err != nil {
			return nil, errors.Wrap(err, "could not retrieve user's followees")
		}

		logger.
			WithFields(logrus.Fields{
				"current_page":  currentPage,
				"count":         len(followees),
				"res.code":      res.StatusCode,
				"res.next_page": res.NextPage,
			}).
			Debug("got followees")

		for _, follower := range moreFollowees {
			followees = append(followees, follower.GetLogin())
		}

		currentPage = res.NextPage
	}

	return followees, nil
}

// GetUserRepositories returns the user's repositories
func (g *Github) GetUserRepositories(ctx context.Context, name string) ([]string, error) {
	return nil, nil
}

// GetRepository returns a repository
func (g *Github) GetRepository(ctx context.Context, name string) (*model.Repository, error) {
	parts := strings.Split(name, "/")
	repoOwner := parts[len(parts)-2]
	repoName := parts[len(parts)-1]

	logger := logrus.WithFields(logrus.Fields{
		"logger":               "providers/Github.GetRepository",
		"repository.name":      name,
		"repository.repoOwner": repoOwner,
		"repository.repoName":  repoName,
	})

	repo, _, err := g.client.Repositories.Get(ctx, repoOwner, repoName)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve repository")
	}

	logger.Debug("got repository")

	stars := []model.UserStar{}

	currentPage := 1
	for currentPage != 0 {
		opts := &github.ListOptions{
			Page:    currentPage,
			PerPage: 100,
		}

		mporeStars, res, err := g.client.Activity.ListStargazers(ctx, repoOwner, repoName, opts)
		if err != nil {
			return nil, errors.Wrap(err, "could not retrieve repo's stars")
		}

		logger.
			WithFields(logrus.Fields{
				"current_page":  currentPage,
				"count":         len(stars),
				"res.code":      res.StatusCode,
				"res.next_page": res.NextPage,
			}).
			Debug("got repo's stars")

		for _, user := range mporeStars {
			stars = append(
				stars,
				model.UserStar{
					User:      user.GetUser().GetLogin(),
					StarredAt: user.StarredAt.Unix(),
				},
			)
		}

		currentPage = res.NextPage
	}

	topics, _, err := g.client.Repositories.ListAllTopics(ctx, repoOwner, repoName)
	if err != nil {
		return nil, errors.Wrap(err, "could not retrieve topics")
	}

	logger.Debug("got repository topics")

	mRepo := &model.Repository{
		Name:   name,
		Labels: topics,
		Stars:  stars,
		Languages: []string{
			repo.GetLanguage(),
		},
	}
	return mRepo, nil
}

// FollowUser follows a user give their login
func (g *Github) FollowUser(ctx context.Context, name string) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger":     "providers/Github.FollowUser",
		"user.login": name,
	})

	logger.Info("following user")

	res, err := g.client.Users.Follow(ctx, name)
	if err != nil {
		return errors.Wrap(err, "could not follow user")
	}

	if res.StatusCode != http.StatusOK {
		logger.
			WithFields(logrus.Fields{
				"res.status":  res.StatusCode,
				"res.headers": res.Header,
			}).
			Warn("following user returned a non-ok status code")
	}

	return nil
}
