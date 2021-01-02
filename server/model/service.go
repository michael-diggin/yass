package model

// StorageResponse encodes key/values and the errors from the storage layer
type StorageResponse struct {
	Key   string
	Value interface{}
	Err   error
}

// Service defines the interface for getting and setting cache key/values
type Service interface {
	Ping() error
	Get(string) <-chan *StorageResponse
	Set(string, uint32, interface{}) <-chan *StorageResponse
	Delete(string) <-chan *StorageResponse

	BatchGet() <-chan map[string]interface{}
	BatchSet(map[string]interface{}) <-chan error

	Close()
}
