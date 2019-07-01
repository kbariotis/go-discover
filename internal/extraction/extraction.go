package extraction

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/kbariotis/go-discover/internal/mailer"
	"github.com/kbariotis/go-discover/internal/model"
	"github.com/kbariotis/go-discover/internal/queue"
	"github.com/kbariotis/go-discover/internal/store"
)

// Extraction is our main orchestrating service
type Extraction struct {
	extractionInterval time.Duration

	graphStore      store.GraphStore
	suggestionStore store.SuggestionStore // Rename because it includes all SQL store

	suggestionExtractionQueue queue.Queue

	mailer mailer.Mailer
}

// New constructs a Github extraction
func New(
	extractionInterval time.Duration,
	graphStore store.GraphStore,
	suggestionStore store.SuggestionStore,
	suggestionExtractionQueue queue.Queue,
	mailer mailer.Mailer,
) (*Extraction, error) {

	extraction := &Extraction{
		graphStore:                graphStore,
		suggestionStore:           suggestionStore,
		extractionInterval:        extractionInterval,
		suggestionExtractionQueue: suggestionExtractionQueue,
		mailer:                    mailer,
	}

	return extraction, nil
}

// handleSuggestionExtractionTask extracts suggestions for each user
func (e *Extraction) handleSuggestionExtractionTask(task *model.SuggestionExtractionTask) error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "extraction/Github.handleSuggestionExtractionTask",
	})

	logger.Info("extracting suggestions")

	// get users and push them to the suggestionExtraction queue
	user, err := e.suggestionStore.GetUser(task.UserName)
	if err != nil {
		return errors.Wrap(err, "could not retrieve user")
	}

	suggestion, err := e.graphStore.GetUserSuggestion(user)
	if err != nil {
		return errors.Wrap(err, "could not extract suggestions")
	}

	if err := e.suggestionStore.PutSuggestion(suggestion); err != nil {
		return errors.Wrap(err, "could not put suggestion")
	}

	html, err := suggestion.ToHTML()
	if err := e.suggestionStore.PutSuggestion(suggestion); err != nil {
		return errors.Wrap(err, "could not generate html")
	}

	if err := e.mailer.SendSuggestion(user.Email, html); err != nil {
		return errors.Wrap(err, "could not send suggestion email")
	}

	return nil
}

// extractSuggestions feed the suggestionExtractionQueue
func (e *Extraction) extractSuggestions() error {
	logger := logrus.WithFields(logrus.Fields{
		"logger": "extraction/Github.extractSuggestions",
	})

	// get users and push them to the suggestionExtraction queue
	users, err := e.suggestionStore.GetAllUsers()
	if err != nil {
		return errors.Wrap(err, "could not retrieve users")
	}

	for _, user := range users {
		logger.
			WithField("user", user).
			Debug("got user, pushing to suggestionExtractionQueue")

		suggestionExtractionTask := &model.SuggestionExtractionTask{
			UserName: user.Name,
		}

		if err := e.suggestionExtractionQueue.Push(suggestionExtractionTask); err != nil {
			return errors.Wrap(err, "could not add user task to queue")
		}
	}

	return nil
}

// Start extracing
func (e *Extraction) Start(ctx context.Context) error {
	cctx, _ := context.WithCancel(ctx)
	logger := logrus.WithFields(logrus.Fields{
		"logger": "extraction/Github.Start",
	})

	suggestionExtractionTasks := make(chan *model.SuggestionExtractionTask, 10000)

	// pop tasks from suggestionExtractionTasks and push them to a local channel
	go func() {
		logger.Info("starting to pop tasks from suggestionExtractionTasks")
		for {
			task, err := e.suggestionExtractionQueue.Pop()
			if err != nil {
				logger.WithError(err).Fatal("could not pop from suggestionExtractionQueue")
			}
			if task == nil {
				time.Sleep(time.Second)
				continue
			}
			if okTask, ok := task.(*model.SuggestionExtractionTask); ok {
				suggestionExtractionTasks <- okTask
			}
		}
	}()

	extractSuggestionsTicker := time.NewTicker(e.extractionInterval)

	for {
		select {
		case <-cctx.Done():
			return nil

		case <-extractSuggestionsTicker.C:
			if err := e.extractSuggestions(); err != nil {
				logger.WithError(err).Warn("extractSuggestions failed")
			}

		case task := <-suggestionExtractionTasks:
			if err := e.handleSuggestionExtractionTask(task); err != nil {
				logger.WithError(err).Warn("failed to handle model.SuggestionExtractionTask")
			}
		}
	}
}
