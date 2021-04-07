package storage

import (
	"sync"

	"github.com/michael-diggin/yass/common/yasserrors"
	"github.com/michael-diggin/yass/server/model"
)

// Service implements the model.Service interface
type Service struct {
	db       map[string]model.Data
	proposed map[string]model.Data
	mu       sync.RWMutex
	pmu      sync.RWMutex
}

// New returns an instance of Service
func New() *Service {
	db := make(map[string]model.Data)
	proposed := make(map[string]model.Data)
	return &Service{db: db, proposed: proposed, mu: sync.RWMutex{}, pmu: sync.RWMutex{}}
}

// Ping performs healthcheck on service
func (s *Service) Ping() error {
	if s.db == nil {
		return yasserrors.NotServing{}
	}
	return nil
}

// Set adds a key/value pair to the db
func (s *Service) Set(key string, hash uint32, value interface{}, commit bool) <-chan *model.StorageResponse {
	respChan := make(chan *model.StorageResponse, 1)
	if commit {
		go func() {
			data := model.Data{Value: value, Hash: hash}
			s.mu.Lock()
			s.pmu.Lock()
			err := setValue(s.db, s.proposed, key, data)
			s.mu.Unlock()
			s.pmu.Unlock()
			respChan <- &model.StorageResponse{Key: key, Err: err}
			close(respChan)
		}()
	} else {
		go func() {
			data := model.Data{Value: value, Hash: hash}
			s.pmu.Lock()
			err := proposeValue(s.proposed, key, data)
			s.pmu.Unlock()
			respChan <- &model.StorageResponse{Key: key, Err: err}
			close(respChan)
		}()
	}
	return respChan
}

func setValue(db, proposed map[string]model.Data, key string, data model.Data) error {
	delete(proposed, key)
	db[key] = data
	return nil
}

func proposeValue(db map[string]model.Data, key string, data model.Data) error {
	if _, ok := db[key]; ok {
		return yasserrors.TransactionError{Key: key}
	}
	db[key] = data
	return nil
}

// Get returns the value of a key in the db
func (s *Service) Get(key string) <-chan *model.StorageResponse {
	respChan := make(chan *model.StorageResponse, 1)
	go func() {
		s.mu.RLock()
		val, err := getValue(s.db, key)
		s.mu.RUnlock()
		respChan <- &model.StorageResponse{Key: key, Value: val, Err: err}
		close(respChan)
	}()
	return respChan
}

func getValue(db map[string]model.Data, key string) (interface{}, error) {
	data, ok := db[key]
	if !ok {
		return nil, yasserrors.NotFound{Key: key}
	}
	return data.Value, nil
}

// Delete removes a key from the proposed db
func (s *Service) Delete(key string) <-chan *model.StorageResponse {
	respChan := make(chan *model.StorageResponse, 1)
	go func() {
		s.pmu.Lock()
		delete(s.proposed, key)
		s.pmu.Unlock()
		respChan <- &model.StorageResponse{}
		close(respChan)
	}()
	return respChan
}

// Close clears the db
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
		for k, v := range s.db {
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

// BatchSet sets all of the passed data to the data db
func (s *Service) BatchSet(newData map[string]model.Data) <-chan error {
	resp := make(chan error)
	go func() {
		s.mu.Lock()
		for key, data := range newData {
			s.db[key] = data
		}
		s.mu.Unlock()
		resp <- nil
		close(resp)
	}()
	return resp
}

// BatchDelete removes all the keys and data from the data db
func (s *Service) BatchDelete(low, high uint32) <-chan error {
	resp := make(chan error)

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
		s.mu.Lock()
		for k, v := range s.db {
			key := k
			if constraintFunc(v.Hash) {
				delete(s.db, key)
			}
		}
		s.mu.Unlock()
		resp <- nil
		close(resp)
	}()
	return resp
}
