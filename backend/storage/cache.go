package storage

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
)

// CacheService implements the Service interface
type CacheService map[string]string

// NewCacheService returns an instance of the cache service
func NewCacheService() CacheService {
	return make(CacheService)
}

// Set adds a key value pair to the in memmory cache service
func (c CacheService) Set(ctx context.Context, key, value string) (string, error) {
	_, ok := c[key]
	if ok {
		logrus.Errorf("Tried to reset key: %s", key)
		return "", errors.New("Key is already in the cache")
	}
	c[key] = value
	return key, nil
}

// Get returns the value from a key in the cache service
func (c CacheService) Get(ctx context.Context, key string) (string, error) {
	res, ok := c[key]
	if !ok {
		logrus.Errorf("Tried to access non existent key: %s", key)
		return "", errors.New("Key is not in the cache")
	}
	return res, nil
}

// Delete removes the key/value from the cache service
func (c CacheService) Delete(ctx context.Context, key string) error {
	_, ok := c[key]
	if !ok {
		logrus.Errorf("Tried to reset key: %s", key)
		return errors.New("Key is not in the cache")
	}
	delete(c, key)
	return nil
}
