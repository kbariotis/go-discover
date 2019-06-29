package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
	"github.com/google/go-github/v25/github"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"

	"github.com/kbariotis/go-discover/internal/model"
	"github.com/kbariotis/go-discover/internal/store"
)

const (
	githubCallbackQueryParam = "code"
)

// API -
type API struct {
	suggestionStore store.SuggestionStore
	githubClient    *github.Client
	oauthConfig     *oauth2.Config
}

// NewAPI -
func NewAPI(
	suggestionStore store.SuggestionStore,
	githubClientID string,
	githubClientSecret string,
	githubCallbackURL string,
) *API {
	oauthCfg := &oauth2.Config{
		ClientID:     githubClientID,
		ClientSecret: githubClientSecret,
		Endpoint:     githuboauth.Endpoint,
		RedirectURL:  githubCallbackURL,
		Scopes: []string{
			"user:email",
		},
	}

	api := &API{
		suggestionStore: suggestionStore,
		oauthConfig:     oauthCfg,
	}

	return api
}

// HandleHealth -
func (api *API) HandleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
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

// HandleGetRoot -
func (api *API) HandleGetRoot(c *gin.Context) {
	state, _ := uuid.NewV4()
	url := api.oauthConfig.AuthCodeURL(state.String())
	c.HTML(http.StatusOK, "index.html", struct {
		GithubLoginURL string
	}{
		GithubLoginURL: url,
	})
}

// HandleGetGithubCallback -
func (api *API) HandleGetGithubCallback(c *gin.Context) {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "api/api.HandleGetGithubCallback",
	})

	values := &struct {
		ErrorMessage string
		User         *model.User
	}{}

	token, err := api.oauthConfig.Exchange(
		oauth2.NoContext,
		c.Query(githubCallbackQueryParam),
	)
	if err != nil {
		values.ErrorMessage = "Could not exchange token"
		c.String(http.StatusBadRequest, "github_callback.html", values)
		logger.WithError(err).Warn("Could not exchange token")
		return
	}

	if !token.Valid() {
		values.ErrorMessage = "Invalid token"
		c.String(http.StatusBadRequest, "github_callback.html", values)
		logger.Warn("Invalid token")
		return
	}

	user, err := api.registerUser(token.AccessToken)
	if err != nil {
		values.ErrorMessage = "Could not register user"
		c.String(http.StatusBadRequest, "github_callback.html", values)
		logger.WithError(err).Warn("Could not register user")
		return
	}

	values.User = user
	c.HTML(http.StatusOK, "github_callback.html", values)
}

// HandleGetUserSuggestions -
func (api *API) HandleGetUserSuggestions(c *gin.Context) {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "api/api.HandleGetUserSuggestions",
	})

	user, err := api.suggestionStore.GetUser("kbariotis")
	if err != nil {
		logger.WithError(err).Warn("Could not get user")
		c.JSON(200, gin.H{
			"message": "pong",
		})
		return
	}

	suggestion, err := api.suggestionStore.GetLatestSuggestionForUser(user.Name)
	if err != nil {
		logger.WithError(err).Warn("Could not get suggestions")
		c.JSON(200, gin.H{
			"message": "pong",
		})
		return
	}

	c.JSON(200, gin.H{
		"response": suggestion.Items,
	})
}

// registerUser -
func (api *API) registerUser(githubToken string) (*model.User, error) {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "api/api.registerUser",
	})

	logger.Info("trying to register user")

	ctx, cf := context.WithTimeout(
		context.Background(),
		time.Second*5,
	)
	defer cf()

	githubClient := newGithubClientWithUserToken(ctx, githubToken)

	ghUser, _, err := githubClient.Users.Get(ctx, "")
	if err != nil {
		return nil, errors.Wrap(err, "could not get user profile")
	}

	user := &model.User{
		Name:  ghUser.GetLogin(),
		Email: ghUser.GetEmail(),
	}

	if user.Email == "" {
		ghEmails, _, err := githubClient.Users.
			ListEmails(
				ctx,
				&github.ListOptions{},
			)
		if err != nil {
			return nil, errors.Wrap(err, "could not list user's emails")
		}

		for _, email := range ghEmails {
			if email.GetPrimary() && email.GetVerified() {
				user.Email = email.GetEmail()
				break
			}
		}
	}

	if user.Name == "" || user.Email == "" {
		return nil, errors.Wrap(err, "missing name or email")
	}

	if err := api.suggestionStore.PutUser(user); err != nil {
		return nil, errors.Wrap(err, "could not persist user")
	}

	return user, nil
}

// Serve API endpoints
func (api *API) Serve(address string) error {
	r := gin.Default()

	// load templates
	r.LoadHTMLGlob("templates/*")

	// ops endpoints
	r.GET("/healthz", api.HandleHealth)

	// frontend endpoits
	r.GET("/", api.HandleGetRoot)
	r.GET("/suggestions/latest", api.HandleGetUserSuggestions)
	r.GET("/github/callback", api.HandleGetGithubCallback)

	return r.Run(address)
}
