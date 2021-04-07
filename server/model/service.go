package model

// Data holds the value interface and the hash value of the key
type Data struct {
	Value interface{}
	Hash  uint32
}

// StorageResponse encodes key/values and the errors from the storage layer
type StorageResponse struct {
	Key   string
	Value interface{}
	Err   error
}

//go:generate mockgen -destination=../mocks/mock_storage_service.go -package=mocks . Service

// Service defines the interface for getting and setting cache key/values
type Service interface {
	Ping() error
	Get(string) <-chan *StorageResponse
	Set(string, uint32, interface{}, bool) <-chan *StorageResponse
	Delete(string) <-chan *StorageResponse

	BatchGet(low, high uint32) <-chan map[string]Data
	BatchSet(map[string]Data) <-chan error
	BatchDelete(low, high uint32) <-chan error

	Close()
}
