package model

import (
	"context"
)

// CacheResponse encodes key/values and the errors from the cache
type CacheResponse struct {
	Key, Value string
	Err        error
}

// Service defines the interface for getting and setting cache key/values
type Service interface {
	Get(context.Context, string) <-chan *CacheResponse
	Set(context.Context, string, string) <-chan *CacheResponse
	Delete(context.Context, string) <-chan *CacheResponse
}
