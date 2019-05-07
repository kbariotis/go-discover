package queue

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -p . Queue

// Queue represents the interface for our queue implementations
type Queue interface {
	Push(interface{}) error
	Pop() (interface{}, error)
}
