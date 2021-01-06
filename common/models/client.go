package models

import (
	"context"
)

//go:generate mockgen -destination=../mocks/mock_client_interface.go -package=mocks . ClientInterface

// ClientInterface has the methods exposed for the client
// that wraps the internal Grpc Client
type ClientInterface interface {
	Check(context.Context) (bool, error)
	SetValue(context.Context, *Pair, int) error
	GetValue(context.Context, string, int) (*Pair, error)
	DelValue(context.Context, string, int) error

	BatchSend(context.Context, int, int, string, uint32, uint32) error
	BatchDelete(context.Context, int, uint32, uint32) error
	BatchGet(context.Context, int) (interface{}, error)
	BatchSet(context.Context, int, interface{}) error

	Close() error
}

//go:generate mockgen -destination=../mocks/mock_client_factory.go -package=mocks . ClientFactory

// ClientFactory is the interface for creating a new instance of the Client Interface
type ClientFactory interface {
	New(context.Context, string) (ClientInterface, error)
}
