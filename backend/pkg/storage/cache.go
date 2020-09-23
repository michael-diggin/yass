package storage

import (
	"sync"

	"github.com/michael-diggin/yass/backend"
)

// Service implements the backend.Service interface
type Service struct {
	cache map[string]string
	mu    sync.RWMutex
}

// New returns an instance of Service
func New() *Service {
	cache := make(map[string]string)
	return &Service{cache: cache}
}

// Ping performs healthcheck on service
func (s *Service) Ping() error {
	return nil
}

// Set adds a key/value pair to the cache
func (s *Service) Set(key, value string) <-chan *backend.CacheResponse {
	respChan := make(chan *backend.CacheResponse, 1)
	go func() {
		s.mu.Lock()
		err := setValue(s.cache, key, value)
		s.mu.Unlock()
		respChan <- &backend.CacheResponse{Key: key, Err: err}
		close(respChan)
	}()
	return respChan
}

func setValue(cache map[string]string, key, value string) error {
	if _, ok := cache[key]; ok {
		return AlreadySetError{key}
	}
	cache[key] = value
	return nil
}

// Get returns the value of a key in the cache
func (s *Service) Get(key string) <-chan *backend.CacheResponse {
	respChan := make(chan *backend.CacheResponse, 1)
	go func() {
		s.mu.RLock()
		val, err := getValue(s.cache, key)
		s.mu.RUnlock()
		respChan <- &backend.CacheResponse{Key: key, Value: val, Err: err}
		close(respChan)
	}()
	return respChan
}

func getValue(cache map[string]string, key string) (string, error) {
	val, ok := cache[key]
	if !ok {
		return "", NotFoundError{key}
	}
	return val, nil
}

// Delete removes a key from the cache
func (s *Service) Delete(key string) <-chan *backend.CacheResponse {
	respChan := make(chan *backend.CacheResponse, 1)
	go func() {
		s.mu.Lock()
		delete(s.cache, key)
		s.mu.Unlock()
		respChan <- &backend.CacheResponse{}
		close(respChan)
	}()
	return respChan
}

// Close clears the cache
func (s *Service) Close() {
	s = &Service{} // let GC handle the freeing up of memory
}
