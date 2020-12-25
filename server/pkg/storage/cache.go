package storage

import (
	"sync"

	"github.com/michael-diggin/yass/server/errors"
	"github.com/michael-diggin/yass/server/model"
)

// Service implements the model.Service interface
type Service struct {
	cache map[string]interface{}
	mu    sync.RWMutex
}

// New returns an instance of Service
func New() *Service {
	cache := make(map[string]interface{})
	return &Service{cache: cache}
}

// Ping performs healthcheck on service
func (s *Service) Ping() error {
	if s.cache == nil {
		return errors.NotServing{}
	}
	return nil
}

// Set adds a key/value pair to the cache
func (s *Service) Set(key string, value interface{}) <-chan *model.CacheResponse {
	respChan := make(chan *model.CacheResponse, 1)
	go func() {
		s.mu.Lock()
		err := setValue(s.cache, key, value)
		s.mu.Unlock()
		respChan <- &model.CacheResponse{Key: key, Err: err}
		close(respChan)
	}()
	return respChan
}

func setValue(cache map[string]interface{}, key string, value interface{}) error {
	if _, ok := cache[key]; ok {
		return errors.AlreadySet{Key: key}
	}
	cache[key] = value
	return nil
}

// Get returns the value of a key in the cache
func (s *Service) Get(key string) <-chan *model.CacheResponse {
	respChan := make(chan *model.CacheResponse, 1)
	go func() {
		s.mu.RLock()
		val, err := getValue(s.cache, key)
		s.mu.RUnlock()
		respChan <- &model.CacheResponse{Key: key, Value: val, Err: err}
		close(respChan)
	}()
	return respChan
}

func getValue(cache map[string]interface{}, key string) (interface{}, error) {
	val, ok := cache[key]
	if !ok {
		return nil, errors.NotFound{Key: key}
	}
	return val, nil
}

// Delete removes a key from the cache
func (s *Service) Delete(key string) <-chan *model.CacheResponse {
	respChan := make(chan *model.CacheResponse, 1)
	go func() {
		s.mu.Lock()
		delete(s.cache, key)
		s.mu.Unlock()
		respChan <- &model.CacheResponse{}
		close(respChan)
	}()
	return respChan
}

// Close clears the cache
func (s *Service) Close() {
	s = &Service{} // let GC handle the freeing up of memory
}
