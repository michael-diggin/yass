package storage

import (
	"testing"
)

func TestPingCache(t *testing.T) {
	ser := New()
	err := ser.Ping()
	if err != nil {
		t.Fatalf("Non nil err: %v", err)
	}
}

func TestSetInCache(t *testing.T) {
	ser := New()
	_ = <-ser.Set("test-key", "test-value")

	tt := []struct {
		name  string
		key   string
		value string
		err   error
	}{
		{"valid", "key", "value", nil},
		{"already set", "test-key", "test-value", AlreadySetError{key: "test-key"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			resp := <-ser.Set(tc.key, tc.value)
			if resp.Err != tc.err {
				t.Fatalf("Unexpected err: %v", resp.Err)
			}
			if resp.Key != tc.key {
				t.Fatalf("Expected '%s', got '%s'", tc.key, resp.Key)
			}
		})
	}
}

func TestGetFromCache(t *testing.T) {
	ser := New()
	_ = <-ser.Set("test-key", "test-value")

	tt := []struct {
		name  string
		key   string
		value string
		err   error
	}{
		{"valid", "test-key", "test-value", nil},
		{"not found", "key", "", NotFoundError{"key"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp := <-ser.Get(tc.key)
			if resp.Err != tc.err {
				t.Fatalf("Unexpected err: %v", resp.Err)
			}
			if resp.Value != tc.value {
				t.Fatalf("Expected '%s', got '%s'", tc.value, resp.Value)
			}
		})
	}

}
func TestDelFromCache(t *testing.T) {
	ser := New()
	_ = <-ser.Set("test-key", "test-value")

	tt := []struct {
		name string
		key  string
		err  error
	}{
		{"valid", "test-key", nil},
		{"not in cache", "key", nil},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp := <-ser.Delete(tc.key)
			if resp.Err != tc.err {
				t.Fatalf("Unexpected err: %v", resp.Err)
			}
		})
	}
}
