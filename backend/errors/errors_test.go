package errors

import "testing"

func TestNotServingError(t *testing.T) {
	err := NotServing{}
	if err.Error() != "Cache service is not serving" {
		t.Fatalf("Unexpected error string: %s", err.Error())
	}
}
func TestAlreadySetError(t *testing.T) {
	key := "test-key"
	err := AlreadySet{key}
	if err.Error() != "Key 'test-key' is already set" {
		t.Fatalf("Unexpected error string: %s", err.Error())
	}
}

func TestNotFoundError(t *testing.T) {
	key := "test-key"
	err := NotFound{key}
	if err.Error() != "Key 'test-key' not found in cache" {
		t.Fatalf("Unexpected error string: %s", err.Error())
	}
}
