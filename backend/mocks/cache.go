package mocks

import (
	"context"

	"github.com/michael-diggin/yass/backend/model"
)

// TestCache implements the Service interface
type TestCache struct {
	SetFn      func(context.Context, string, string) *model.CacheResponse
	SetInvoked bool
	GetFn      func(context.Context, string) *model.CacheResponse
	GetInvoked bool
	DelFn      func(context.Context, string) *model.CacheResponse
	DelInvoked bool
}

// Set adds a key value pair to the in memmory cache service
func (c TestCache) Set(ctx context.Context, key, value string) <-chan *model.CacheResponse {
	c.SetInvoked = true
	resp := make(chan *model.CacheResponse, 1)
	go func() { resp <- c.SetFn(ctx, key, value) }()
	return resp
}

// Get returns the value from a key in the cache service
func (c TestCache) Get(ctx context.Context, key string) <-chan *model.CacheResponse {
	c.GetInvoked = true
	resp := make(chan *model.CacheResponse)
	go func() { resp <- c.GetFn(ctx, key) }()
	return resp
}

// Delete removes the key/value from the cache service
func (c TestCache) Delete(ctx context.Context, key string) <-chan *model.CacheResponse {
	c.DelInvoked = true
	resp := make(chan *model.CacheResponse)
	go func() { resp <- c.DelFn(ctx, key) }()
	return resp
}
