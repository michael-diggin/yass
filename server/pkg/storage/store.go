package storage

import (
	"sync"

	"github.com/michael-diggin/yass/server/errors"
	"github.com/michael-diggin/yass/server/model"
)

// Service implements the model.Service interface
type Service struct {
	store map[string]interface{}
	mu    sync.RWMutex
}

// New returns an instance of Service
func New() *Service {
	store := make(map[string]interface{})
	return &Service{store: store, mu: sync.RWMutex{}}
}

// Ping performs healthcheck on service
func (s *Service) Ping() error {
	if s.store == nil {
		return errors.NotServing{}
	}
	return nil
}

// Set adds a key/value pair to the store
func (s *Service) Set(key string, value interface{}) <-chan *model.StorageResponse {
	respChan := make(chan *model.StorageResponse, 1)
	go func() {
		s.mu.Lock()
		err := setValue(s.store, key, value)
		s.mu.Unlock()
		respChan <- &model.StorageResponse{Key: key, Err: err}
		close(respChan)
	}()
	return respChan
}

func setValue(store map[string]interface{}, key string, value interface{}) error {
	if _, ok := store[key]; ok {
		return errors.AlreadySet{Key: key}
	}
	store[key] = value
	return nil
}

// Get returns the value of a key in the store
func (s *Service) Get(key string) <-chan *model.StorageResponse {
	respChan := make(chan *model.StorageResponse, 1)
	go func() {
		s.mu.RLock()
		val, err := getValue(s.store, key)
		s.mu.RUnlock()
		respChan <- &model.StorageResponse{Key: key, Value: val, Err: err}
		close(respChan)
	}()
	return respChan
}

func getValue(store map[string]interface{}, key string) (interface{}, error) {
	val, ok := store[key]
	if !ok {
		return nil, errors.NotFound{Key: key}
	}
	return val, nil
}

// Delete removes a key from the store
func (s *Service) Delete(key string) <-chan *model.StorageResponse {
	respChan := make(chan *model.StorageResponse, 1)
	go func() {
		s.mu.Lock()
		delete(s.store, key)
		s.mu.Unlock()
		respChan <- &model.StorageResponse{}
		close(respChan)
	}()
	return respChan
}

// Close clears the store
func (s *Service) Close() {
	s = &Service{} // let GC handle the freeing up of memory
}
