package storage

import (
	"fmt"
)

// AlreadySetError is the error when a key is already in the cache
type AlreadySetError struct {
	key string
}

func (a AlreadySetError) Error() string {
	return fmt.Sprintf("Key '%s' is already set", a.key)
}

// NotFoundError is error for when key is not in the cache
type NotFoundError struct {
	key string
}

func (a NotFoundError) Error() string {
	return fmt.Sprintf("Key '%s' not found in cache", a.key)
}
