package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v25/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/kbariotis/go-discover/internal/model"
	"github.com/kbariotis/go-discover/internal/store"
)

const (
	queryParamGithubToken = "github_token"
)

// API -
type API struct {
	suggestionStore store.SuggestionStore
}

// HandleHealth -
func (api *API) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// NewAPI -
func NewAPI(suggestionStore store.SuggestionStore) *API {
	api := &API{
		suggestionStore: suggestionStore,
	}

	return api
}

func newGithubClientWithUserToken(
	ctx context.Context,
	token string,
) *github.Client {
	ghTokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{
			AccessToken: token,
		},
	)

	ghClient := github.NewClient(
		oauth2.NewClient(ctx, ghTokenSource),
	)

	return ghClient
}

// HandlePostUsers -
func (api *API) HandlePostUsers(c *gin.Context) {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "api/api.HandlePostUsers",
	})

	logger.Info("handling")

	ctx, cf := context.WithTimeout(
		context.Background(),
		time.Second*5,
	)
	defer cf()

	githubToken := c.Query(queryParamGithubToken)
	githubClient := newGithubClientWithUserToken(ctx, githubToken)

	ghUser, res, err := githubClient.Users.Get(ctx, "")
	if err != nil {
		logger.WithError(err).Info("could not list user's profile")
		// TODO don't use the code we got from github, map it to an internal
		c.JSON(res.StatusCode, nil)
		return
	}

	user := &model.User{
		Name:  ghUser.GetLogin(),
		Email: ghUser.GetEmail(),
	}

	if user.Email == "" {
		ghEmails, res, err := githubClient.Users.
			ListEmails(
				ctx,
				&github.ListOptions{},
			)
		if err != nil {
			logger.WithError(err).Info("could not list user's emails")
			// TODO don't use the code we got from github, map it to an internal
			c.JSON(res.StatusCode, nil)
			return
		}

		for _, email := range ghEmails {
			if email.GetPrimary() && email.GetVerified() {
				user.Email = email.GetEmail()
				break
			}
		}
	}

	if user.Name == "" || user.Email == "" {
		logger.WithField("user", user).Info("missing login or email")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "missing login or email",
		})
		return
	}

	if err := api.suggestionStore.PutUser(user); err != nil {
		logger.WithError(err).Error("could not put user")
		c.JSON(http.StatusInternalServerError, nil)
		return
	}

	c.JSON(http.StatusOK, nil)
}

// Serve API endpoints
func (api *API) Serve(address string) error {
	r := gin.Default()

	// ops endpoints
	r.GET("/healthz", api.HandleHealth)

	// public endpoints
	r.POST("/users", api.HandlePostUsers)

	return r.Run(address)
}
