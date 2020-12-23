package mocks

import (
	"context"

	pb "github.com/michael-diggin/yass/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockCacheClient implements the pb client interface
type MockCacheClient struct {
	PingFn      func() error
	PingInvoked bool
	SetFn       func(context.Context, string, []byte) error
	SetInvoked  bool
	GetFn       func(context.Context, string) ([]byte, error)
	GetInvoked  bool
	DelFn       func(context.Context, string) error
	DelInvoked  bool
}

// Ping implements ping
func (m *MockCacheClient) Ping(ctx context.Context, in *pb.Null, opts ...grpc.CallOption) (*pb.PingResponse, error) {
	m.PingInvoked = true
	err := m.PingFn()
	if err != nil {
		resp := &pb.PingResponse{Status: pb.PingResponse_NOT_SERVING}
		return resp, status.Error(codes.Unavailable, err.Error())
	}
	resp := &pb.PingResponse{Status: pb.PingResponse_SERVING}
	return resp, nil
}

// Set calls the mocked set fn
func (m *MockCacheClient) Set(ctx context.Context, in *pb.Pair, opts ...grpc.CallOption) (*pb.Key, error) {
	m.SetInvoked = true
	err := m.SetFn(ctx, in.Key, in.Value)
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
