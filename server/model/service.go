package model

// CacheResponse encodes key/values and the errors from the cache
type CacheResponse struct {
	Key   string
	Value interface{}
	Err   error
}

// Service defines the interface for getting and setting cache key/values
type Service interface {
	Ping() error
	Get(string) <-chan *CacheResponse
	Set(string, interface{}) <-chan *CacheResponse
	Delete(string) <-chan *CacheResponse
	Close()
}
