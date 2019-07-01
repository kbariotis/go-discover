package mailer

// Mailer interface
type Mailer interface {
	SendSuggestion(email string, html string) error
}
