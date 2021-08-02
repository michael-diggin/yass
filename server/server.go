package server

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/michael-diggin/yass/api"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
)

type Config struct {
	DB DB
}

type DB interface {
	Set(record *api.Record) error
	Get(id string) (*api.Record, error)
}

var _ api.StorageServer = (*grpcServer)(nil)

type grpcServer struct {
	api.UnimplementedStorageServer
	*Config
}

func newgrpcServer(config *Config) (*grpcServer, error) {
	return &grpcServer{Config: config}, nil
}

func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	logger := zap.L().Named("server")
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(func(duration time.Duration) zapcore.Field {
			return zap.Int64("grcp.time_ns", duration.Nanoseconds())
		}),
	}

	opts = append(opts, grpc.StreamInterceptor(
		grpc_middleware.ChainStreamServer(
			grpc_ctxtags.StreamServerInterceptor(),
			grpc_zap.StreamServerInterceptor(logger, zapOpts...),
		)), grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
		grpc_ctxtags.UnaryServerInterceptor(),
		grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
	)),
	)
	gsrv := grpc.NewServer(opts...)
	srv, err := newgrpcServer(config)
	if err != nil {
		return nil, err
	}
	api.RegisterStorageServer(gsrv, srv)
	return gsrv, nil
}

func (s *grpcServer) Set(ctx context.Context, req *api.SetRequest) (*api.SetResponse, error) {
	err := s.DB.Set(req.Record)
	if err != nil {
		return nil, err
	}
	return &api.SetResponse{}, nil
}

func (s *grpcServer) Get(ctx context.Context, req *api.GetRequest) (*api.GetResponse, error) {
	rec, err := s.DB.Get(req.Id)
	if err != nil {
		return nil, err
	}
	return &api.GetResponse{Record: rec}, nil
}
