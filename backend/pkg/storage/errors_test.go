package storage

import "testing"

func TestAlreadySetError(t *testing.T) {
	key := "test-key"
	err := AlreadySetError{key}
	if err.Error() != "Key 'test-key' is already set" {
		t.Fatalf("Unexpected error string: %s", err.Error())
	}
}

func TestNotFoundError(t *testing.T) {
	key := "test-key"
	err := NotFoundError{key}
	if err.Error() != "Key 'test-key' not found in cache" {
		t.Fatalf("Unexpected error string: %s", err.Error())
	}
}
