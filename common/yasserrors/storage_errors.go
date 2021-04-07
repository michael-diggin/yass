package yasserrors

import (
	"fmt"
)

// NotServing is error when the storage is not ready to serve
type NotServing struct{}

func (e NotServing) Error() string {
	return fmt.Sprint("Storage service is not serving")
}

// AlreadySet is the error when a key is already in the storage
type AlreadySet struct {
	Key string
}

func (e AlreadySet) Error() string {
	return fmt.Sprintf("Key '%s' is already set", e.Key)
}

// AlreadySet is the error when a key is already in the storage
type TransactionError struct {
	Key string
}

func (e TransactionError) Error() string {
	return fmt.Sprintf("Transaction ongoing for key '%s'", e.Key)
}

// NotFound is error for when key is not in the storage
type NotFound struct {
	Key string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("Key '%s' not found in storage", e.Key)
}
