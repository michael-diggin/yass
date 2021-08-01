package server

import (
	"context"

	"github.com/michael-diggin/yass/api"
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
