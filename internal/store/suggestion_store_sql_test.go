package store

import (
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"

	_ "github.com/jinzhu/gorm/dialects/sqlite" // required for sqlite

	"github.com/kbariotis/go-discover/internal/model"
)

func TestSuggestionSQL_Users(t *testing.T) {
	// connect to db
	db := getDB(t)
	user := &model.User{
		Name:  "foo",
		Email: "foo@bar.io",
	}

	// enable debugging
	db.LogMode(true)

	// construct store
	s := &SuggestionSQL{
		db: db,
	}

	// setup db
	require.NoError(t, s.Setup())

	// put user and check error
	gotPutErr := s.PutUser(user)
	require.NoError(t, gotPutErr)

	// get user and check response and error
	gotUser, gotGetErr := s.GetUser(user.Name)
	require.NoError(t, gotGetErr)
	require.Equal(t, user, gotUser)

	// get all users and check response and error
	gotUsers, gotUsersErr := s.GetAllUsers()
	require.NoError(t, gotUsersErr)
	require.Equal(t, []*model.User{user}, gotUsers)

	// update user
	user.Email = "not@bar.io"
	gotPutErr = s.PutUser(user)
	require.NoError(t, gotPutErr)

	// get updated user and check response and error
	gotUser, gotGetErr = s.GetUser(user.Name)
	require.NoError(t, gotGetErr)
	require.Equal(t, user, gotUser)

	// cleanup and close db
	require.NoError(t, s.Cleanup())
	require.NoError(t, db.Close())
}

func TestSuggestionSQL_Suggestions(t *testing.T) {
	// connect to db
	db := getDB(t)
	suggestion := &model.Suggestion{
		UserID:   model.User{Name: "kbariotis"},
		DateTime: time.Now().Round(time.Second).UTC(),
		Items: []model.SuggestionItem{
			{
				Type:   model.SuggestionTypeStarRepository,
				Value:  "github.com/kbariotis/go-discover",
				Reason: "because it's epic",
			},
			{
				Type:   model.SuggestionTypeFollowUser,
				Value:  "github.com/kbariotis",
				Reason: "because of reasons",
			},
		},
	}

	// enable debugging
	db.LogMode(true)

	// construct store
	s := &SuggestionSQL{
		db: db,
	}

	// setup db
	require.NoError(t, s.Setup())

	// put suggestion and check error
	gotPutErr := s.PutSuggestion(suggestion)
	require.NoError(t, gotPutErr)

	// get suggestion and check response and error
	gotSuggestions, gotGetErr := s.GetSuggestion(1)
	require.NoError(t, gotGetErr)
	require.Equal(t, suggestion, gotSuggestions)

	// update suggestion
	suggestion.Items = append(
		suggestion.Items,
		model.SuggestionItem{
			Type:   model.SuggestionTypeStarRepository,
			Value:  "github.com/geoah/go-discover",
			Reason: "just for good measure",
		},
	)
	gotPutErr = s.PutSuggestion(suggestion)
	require.NoError(t, gotPutErr)

	// get updated suggestion and check response and error
	gotSuggestions, gotGetErr = s.GetSuggestion(1)
	require.NoError(t, gotGetErr)
	require.Equal(t, suggestion, gotSuggestions)

	// cleanup and close db
	require.NoError(t, s.Cleanup())
	require.NoError(t, db.Close())
}

func getDB(t *testing.T) *gorm.DB {
	dir, err := ioutil.TempDir("", "go-discover-store")
	require.NoError(t, err)
	db, err := gorm.Open("sqlite3", filepath.Join(dir, "test.db"))
	require.NoError(t, err)
	return db
}
