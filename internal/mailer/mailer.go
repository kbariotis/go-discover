package mailer

// Mailer interface
type Mailer interface {
	Mail(email string, html string) error
}
