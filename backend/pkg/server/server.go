package server

import (
	"context"
	"net"

	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// YassServer wraps up the listener, grpc server and cache service
type YassServer struct {
	lis   net.Listener
	srv   *grpc.Server
	cache backend.Service
}

// New sets up the server
func New(lis net.Listener, cache backend.Service) YassServer {
	s := grpc.NewServer()
	pb.RegisterCacheServer(s, server{Cache: cache})
	return YassServer{lis: lis, srv: s, cache: cache}
}

// SpinUp starts the grpc server
func (y YassServer) SpinUp() <-chan error {
	errChan := make(chan error)
	go func() {
		err := y.srv.Serve(y.lis)
		if err != nil {
			errChan <- err
			close(errChan)
		}
	}()
	return errChan
}

// ShutDown closes the server resources
func (y YassServer) ShutDown() {
	logrus.Infof("Shutting down server resources")
	y.cache.Close()
	logrus.Infof("Cache closed")
	y.srv.GracefulStop()
	logrus.Infof("gRPC server stopped")
}

// server (unexported) implements the CacheServer interface
type server struct {
	Cache backend.Service
}

// Ping serves the healthcheck endpoint for the server
// It checks if the cache is serving too and responds accordingly
func (s server) Ping(context.Context, *pb.Null) (*pb.PingResponse, error) {
	err := s.Cache.Ping()
	if err != nil {
		resp := &pb.PingResponse{Status: pb.PingResponse_NOT_SERVING}
		return resp, status.Error(codes.Unavailable, err.Error())
	}
	resp := &pb.PingResponse{Status: pb.PingResponse_SERVING}
	return resp, nil
}

// Set takes a key/value pair and adds it to the cache storage
// It returns an error if the key is already set
func (s server) Set(ctx context.Context, pair *pb.Pair) (*pb.Key, error) {
	if pair.Key == "" || pair.Value == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot set a zero key or value")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-s.Cache.Set(pair.Key, pair.Value):
		if cacheResp.Err != nil {
			return nil, status.Error(codes.AlreadyExists, cacheResp.Err.Error())
		}
		output := &pb.Key{Key: cacheResp.Key}
		return output, nil
	}
}

// Get returns the value of a key
// It returns an error if the key is not in the cache
func (s server) Get(ctx context.Context, key *pb.Key) (*pb.Pair, error) {
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot get a zero key")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-s.Cache.Get(key.Key):
		if cacheResp.Err != nil {
			return nil, status.Error(codes.NotFound, cacheResp.Err.Error())
		}
		pair := &pb.Pair{Key: key.Key, Value: cacheResp.Value}
		return pair, nil
	}
}

// Delete is the endpoint to delete a key/value if it is already in the cache
func (s server) Delete(ctx context.Context, key *pb.Key) (*pb.Null, error) {
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot delete a zero key")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case <-s.Cache.Delete(key.Key):
		return &pb.Null{}, nil
	}
}
