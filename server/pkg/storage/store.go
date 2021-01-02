package storage

import (
	"sync"

	"github.com/michael-diggin/yass/server/errors"
	"github.com/michael-diggin/yass/server/model"
)

// Service implements the model.Service interface
type Service struct {
	store map[string]model.Data
	mu    sync.RWMutex
}

// New returns an instance of Service
func New() *Service {
	store := make(map[string]model.Data)
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
func (s *Service) Set(key string, hash uint32, value interface{}) <-chan *model.StorageResponse {
	respChan := make(chan *model.StorageResponse, 1)
	go func() {
		data := model.Data{Value: value, Hash: hash}
		s.mu.Lock()
		err := setValue(s.store, key, data)
		s.mu.Unlock()
		respChan <- &model.StorageResponse{Key: key, Err: err}
		close(respChan)
	}()
	return respChan
}

func setValue(store map[string]model.Data, key string, data model.Data) error {
	if _, ok := store[key]; ok {
		return errors.AlreadySet{Key: key}
	}
	store[key] = data
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

func getValue(store map[string]model.Data, key string) (interface{}, error) {
	data, ok := store[key]
	if !ok {
		return nil, errors.NotFound{Key: key}
	}
	return data.Value, nil
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

// BatchGet returns all of the stored data that lies in the region
// (low, high]
func (s *Service) BatchGet(low, high uint32) <-chan map[string]model.Data {
	resp := make(chan map[string]model.Data)
	var constraintFunc func(uint32) bool

	if low < high {
		constraintFunc = func(hash uint32) bool {
			return hash > low && hash <= high
		}
	} else {
		constraintFunc = func(hash uint32) bool {
			return hash <= high || hash > low
		}
	}

	go func() {
		data := make(map[string]model.Data)
		s.mu.RLock()
		for k, v := range s.store {
			if constraintFunc(v.Hash) {
				data[k] = v
			}
		}
		s.mu.RUnlock()
		resp <- data
		close(resp)
	}()
	return resp
}

// BatchSet sets all of the passed data to the data store
func (s *Service) BatchSet(newData map[string]model.Data) <-chan error {
	resp := make(chan error)
	go func() {
		s.mu.Lock()
		for key, data := range newData {
			if _, ok := s.store[key]; ok {
				continue
			}
			s.store[key] = data
		}
		s.mu.Unlock()
		resp <- nil
		close(resp)
	}()
	return resp
}
