package model

import (
	"bytes"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

const (
	newSuggestionEmailTemplate = `
		<html>
			<head></head>
			<body>
				<br/>
				This is you weekly report from GitHub.
				<br/>
				<ul>
				{{range .Suggestion.Items}}
					<li>
						Repository: {{.Value}} because {{.Reason}}
					</li>
				{{end}}
				</ul>
			</body>
		</html>
	`
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
	// gorm.Model
	ID           uint `gorm:"primary_key"`
	Type         string
	Value        string
	SuggestionID uint
	Reason       string
}

// Suggestion contains a list of suggestions
type Suggestion struct {
	// gorm.Model
	ID       uint `gorm:"primary_key"`
	UserID   User
	DateTime time.Time
	Items    []SuggestionItem
}

func (s *Suggestion) ToHTML() (string, error) {
	// create template for query
	newSuggestionEmail, err := template.
		New("newSuggestionEmail").
		Parse(newSuggestionEmailTemplate)
	if err != nil {
		return "", errors.Wrap(err, "could not parse template")
	}

	// render mailTemplate
	mailTemplate := &bytes.Buffer{}
	if err := newSuggestionEmail.Execute(mailTemplate, map[string]interface{}{
		"Suggestion": s,
	}); err != nil {
		return "", errors.Wrap(err, "could not execute template")
	}

	return mailTemplate.String(), nil
}
