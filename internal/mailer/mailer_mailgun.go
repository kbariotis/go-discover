package mailer

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// Mailgun mailer
type Mailgun struct {
	client *mailgun.MailgunImpl
}

// NewMailgun constructs a new Mailgun provider
func NewMailgun(client *mailgun.MailgunImpl) (Mailer, error) {
	ml := &Mailgun{
		client: client,
	}

	return ml, nil
}

// Mail to send a given to the user
func (m *Mailgun) Mail(email string, html string) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger":     "mailers/Mailgun.Mail",
		"user.email": email,
	})

	logger.Info("sending suggestion to the user")

	msg := m.client.NewMessage(
		"Kostas Bariotis <newsletter@sandboxbd7bb1a1171e4f41abb06173d239f07d.mailgun.org>",
		"Your weekly GitHub newsletter",
		"",
		email,
	)
	msg.SetHtml(html)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, _, err := m.client.Send(ctx, msg)

	if err != nil {
		return errors.Wrap(err, "could not parse template")
	}

	return nil
}
