package mocks

import (
	"context"

	pb "github.com/michael-diggin/yass/api"

	"google.golang.org/grpc"
)

// MockCacheClient implements the pb client interface
type MockCacheClient struct {
	AddFn      func(context.Context, string, string) error
	AddInvoked bool
	GetFn      func(context.Context, string) (string, error)
	GetInvoked bool
	DelFn      func(context.Context, string) error
	DelInvoked bool
}

// Add calls the mocked add fn
func (m *MockCacheClient) Add(ctx context.Context, in *pb.Pair, opts ...grpc.CallOption) (*pb.Key, error) {
	m.AddInvoked = true
	err := m.AddFn(ctx, in.Key, in.Value)
	return &pb.Key{Key: in.Key}, err
}

// Get calls the mocked get fn
func (m *MockCacheClient) Get(ctx context.Context, in *pb.Key, opts ...grpc.CallOption) (*pb.Pair, error) {
	m.GetInvoked = true
	val, err := m.GetFn(ctx, in.Key)
	return &pb.Pair{Key: in.Key, Value: val}, err
}

// Delete calls the mocked delete fn
func (m *MockCacheClient) Delete(ctx context.Context, in *pb.Key, opts ...grpc.CallOption) (*pb.Null, error) {
	m.DelInvoked = true
	err := m.DelFn(ctx, in.Key)
	return &pb.Null{}, err
}
