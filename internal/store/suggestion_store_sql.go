package store

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"

	"github.com/kbariotis/go-discover/internal/model"
)

// SuggestionSQL store implementation
type SuggestionSQL struct {
	db *gorm.DB
}

var (
	suggestionModels = []interface{}{
		&model.User{},
		&model.Suggestion{},
		&model.SuggestionItem{},
	}
)

// NewSuggestionSQL constrcuts a new SuggestionSQL store given a neoism db
func NewSuggestionSQL(db *gorm.DB) (*SuggestionSQL, error) {
	neo := &SuggestionSQL{
		db: db,
	}

	return neo, nil
}

// Setup -
func (s *SuggestionSQL) Setup() error {
	res := s.db.AutoMigrate(suggestionModels...)
	return errors.Wrap(res.Error, "could not drop tables")
}

// Cleanup -
func (s *SuggestionSQL) Cleanup() error {
	res := s.db.DropTableIfExists(suggestionModels...)
	return errors.Wrap(res.Error, "could not drop tables")
}

// GetAllUsers -
func (s *SuggestionSQL) GetAllUsers() ([]*model.User, error) {
	users := []*model.User{}
	res := s.db.Find(&users)
	return users, errors.Wrap(res.Error, "could not get all users")
}

// GetUser -
func (s *SuggestionSQL) GetUser(name string) (*model.User, error) {
	user := &model.User{}
	res := s.db.First(user, model.User{Name: name})
	return user, errors.Wrap(res.Error, "could not get user")
}

// PutUser -
func (s *SuggestionSQL) PutUser(user *model.User) error {
	selectedUser := model.User{
		Name: user.Name,
	}
	updatedUser := model.User{
		Email: user.Email,
	}
	res := s.db.
		Where(selectedUser).
		Assign(updatedUser).
		FirstOrCreate(user)
	return errors.Wrap(res.Error, "could not put user")
}

// GetSuggestion -
func (s *SuggestionSQL) GetSuggestion(id uint) (*model.Suggestion, error) {
	suggestion := &model.Suggestion{}
	res := s.db.
		Preload("Items").
		First(suggestion, model.Suggestion{ID: id})
	return suggestion, errors.Wrap(res.Error, "could not get suggestion")
}

// PutSuggestion -
func (s *SuggestionSQL) PutSuggestion(suggestion *model.Suggestion) error {
	selectedSuggestion := model.Suggestion{
		ID: suggestion.ID,
	}
	updatedSuggestion := model.Suggestion{
		UserID:   suggestion.UserID,
		DateTime: suggestion.DateTime,
		Items:    suggestion.Items,
	}
	res := s.db.
		Where(selectedSuggestion).
		Assign(updatedSuggestion).
		FirstOrCreate(suggestion)
	return errors.Wrap(res.Error, "could not put suggestions")
}
