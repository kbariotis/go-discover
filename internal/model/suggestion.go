package model

import (
	"time"
)

// SuggestionType defines the types for suggestions
type SuggestionType string

const (
	// SuggestionTypeNone probably due to an error
	SuggestionTypeNone SuggestionType = ""
	// SuggestionTypeStarRepository to star a repository
	SuggestionTypeStarRepository = "STAR_REPOSITORY"
	// SuggestionTypeFollowUser to follow a user
	SuggestionTypeFollowUser = "FOLLOW_USER"
)

// SuggestionItem is a single repository suggestion
type SuggestionItem struct {
	ID           uint `gorm:"primary_key"`
	SuggestionID uint
	Type         string
	Value        string
	Reason       string
}

// Suggestion contains a list of suggestions
type Suggestion struct {
	ID       uint   `gorm:"primary_key"`
	UserID   string `gorm:"index:user_id"`
	DateTime time.Time
	Items    []SuggestionItem `gorm:"foreignkey:SuggestionID"`
}
