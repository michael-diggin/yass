package mocks

import (
	"context"
)

// MockGrpcClient implements the pb client interface
type MockGrpcClient struct {
	SetFn      func(context.Context, string, interface{}) error
	SetInvoked bool
	GetFn      func(context.Context, string) (interface{}, error)
	GetInvoked bool
	DelFn      func(context.Context, string) error
	DelInvoked bool
}

// SetValue calls the mocked set fn
func (m MockGrpcClient) SetValue(ctx context.Context, key string, value interface{}) error {
	m.SetInvoked = true
	return m.SetFn(ctx, key, value)
}

// GetValue calls the mocked get fn
func (m MockGrpcClient) GetValue(ctx context.Context, key string) (interface{}, error) {
	m.GetInvoked = true
	return m.GetFn(ctx, key)
}

// DelValue calls the mocked delete fn
func (m MockGrpcClient) DelValue(ctx context.Context, key string) error {
	m.DelInvoked = true
	return m.DelFn(ctx, key)
}
