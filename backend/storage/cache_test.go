package storage

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestAddToCacheService(t *testing.T) {
	logrus.SetOutput(ioutil.Discard) // Discard log output for test
	cache := NewCacheService()
	cache["testKey"] = "testValue"
	tt := []struct {
		name   string
		key    string
		value  string
		errVal string
	}{
		{"valid case", "newKey", "newValue", ""},
		{"already set", "testKey", "testValue", "Key is already in the cache"},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			key, err := cache.Set(context.Background(), tc.key, tc.value)
			if err != nil {
				if err.Error() != tc.errVal {
					t.Fatalf("Expected %v, got: %v", tc.errVal, err)
				}
			}
			if key != "" {
				if key != "newKey" {
					t.Fatalf("Expected '%s', got '%s'", tc.key, key)
				}
				if cache[tc.key] != tc.value {
					t.Fatalf("Value not in the cache")
				}
			}
		})
	}
}

func TestGetFromCacheService(t *testing.T) {
	logrus.SetOutput(ioutil.Discard) // Discard log output for test
	cache := NewCacheService()
	cache["testKey"] = "testValue"
	tt := []struct {
		name   string
		key    string
		value  string
		errVal string
	}{
		{"valid case", "testKey", "testValue", ""},
		{"already set", "newKey", "", "Key is not in the cache"},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			value, err := cache.Get(context.Background(), tc.key)
			if err != nil {
				if err.Error() != tc.errVal {
					t.Fatalf("Expected %v, got: %v", tc.errVal, err)
				}
			}
			if value != tc.value {

			}
		})
	}
}

func TestDeleteFromCacheService(t *testing.T) {
	logrus.SetOutput(ioutil.Discard) // Discard log output for test
	cache := NewCacheService()
	cache["testKey"] = "testValue"
	tt := []struct {
		name   string
		key    string
		errVal string
	}{
		{"valid case", "testKey", ""},
		{"already set", "newKey", "Key is not in the cache"},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			err := cache.Delete(context.Background(), tc.key)
			if err != nil {
				if err.Error() != tc.errVal {
					t.Fatalf("Expected %v, got: %v", tc.errVal, err)
				}
			}
		})
	}
}
