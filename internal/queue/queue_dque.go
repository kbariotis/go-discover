package queue

import (
	"github.com/joncrlsn/dque"
	"github.com/pkg/errors"
)

// DQueue implements a Queue using a DQueue as the underlying provider
type DQueue struct {
	dque *dque.DQue
}

// NewDQueue constrcuts a new Queue with an underlying DQueue provider
func NewDQueue(name, dir string, task interface{}) (Queue, error) {
	d, err := dque.NewOrOpen(name, dir, 50, func() interface{} {
		return task
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to construct new DQue")
	}

	q := &DQueue{
		dque: d,
	}

	return q, nil
}

// Push item to the end of the queue
func (q *DQueue) Push(o interface{}) error {
	return q.dque.Enqueue(o)
}

// Pop item from top of the queue
func (q *DQueue) Pop() (interface{}, error) {
	return q.dque.Dequeue()
}
