package mocks

import (
	"github.com/michael-diggin/yass/server/model"
)

// TestStorage implements the Service interface
type TestStorage struct {
	PingFn      func() error
	PingInvoked bool
	SetFn       func(string, interface{}) *model.StorageResponse
	SetInvoked  bool
	GetFn       func(string) *model.StorageResponse
	GetInvoked  bool
	DelFn       func(string) *model.StorageResponse
	DelInvoked  bool
}

// Ping implements ping
func (c *TestStorage) Ping() error {
	c.PingInvoked = true
	err := c.PingFn()
	return err
}

// Set adds a key value pair to the in memmory cache service
func (c *TestStorage) Set(key string, value interface{}) <-chan *model.StorageResponse {
	c.SetInvoked = true
	resp := make(chan *model.StorageResponse, 1)
	go func() { resp <- c.SetFn(key, value) }()
	return resp
}

// Get returns the value from a key in the cache service
func (c *TestStorage) Get(key string) <-chan *model.StorageResponse {
	c.GetInvoked = true
	resp := make(chan *model.StorageResponse)
	go func() { resp <- c.GetFn(key) }()
	return resp
}

// Delete removes the key/value from the cache service
func (c *TestStorage) Delete(key string) <-chan *model.StorageResponse {
	c.DelInvoked = true
	resp := make(chan *model.StorageResponse)
	go func() { resp <- c.DelFn(key) }()
	return resp
}

// Close method so it satisfies the interface
func (c *TestStorage) Close() {

}
