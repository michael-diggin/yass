package backend

// CacheResponse encodes key/values and the errors from the cache
type CacheResponse struct {
	Key, Value string
	Err        error
}

// Service defines the interface for getting and setting cache key/values
type Service interface {
	Ping() error
	Get(string) <-chan *CacheResponse
	Set(string, string) <-chan *CacheResponse
	Delete(string) <-chan *CacheResponse
	Close()
}
