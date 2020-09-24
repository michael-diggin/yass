package mocks

import (
	"github.com/michael-diggin/yass/backend"
)

// TestCache implements the Service interface
type TestCache struct {
	PingFn      func() error
	PingInvoked bool
	SetFn       func(string, string) *backend.CacheResponse
	SetInvoked  bool
	GetFn       func(string) *backend.CacheResponse
	GetInvoked  bool
	DelFn       func(string) *backend.CacheResponse
	DelInvoked  bool
}

// Ping implements ping
func (c *TestCache) Ping() error {
	c.PingInvoked = true
	err := c.PingFn()
	return err
}

// Set adds a key value pair to the in memmory cache service
func (c *TestCache) Set(key, value string) <-chan *backend.CacheResponse {
	c.SetInvoked = true
	resp := make(chan *backend.CacheResponse, 1)
	go func() { resp <- c.SetFn(key, value) }()
	return resp
}

// Get returns the value from a key in the cache service
func (c *TestCache) Get(key string) <-chan *backend.CacheResponse {
	c.GetInvoked = true
	resp := make(chan *backend.CacheResponse)
	go func() { resp <- c.GetFn(key) }()
	return resp
}

// Delete removes the key/value from the cache service
func (c *TestCache) Delete(key string) <-chan *backend.CacheResponse {
	c.DelInvoked = true
	resp := make(chan *backend.CacheResponse)
	go func() { resp <- c.DelFn(key) }()
	return resp
}

// Close method so it satisfies the interface
func (c *TestCache) Close() {

}
