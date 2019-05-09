package cache

import "github.com/pkg/errors"

// ErrAlreadyLocked is returned on Lock* when the key is already locked
var ErrAlreadyLocked = errors.New("key already locked")

// Cache defines the interface for the cache implementations
type Cache interface {
	LockUser(user string) error
}
