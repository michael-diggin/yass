package core

import (
	"context"
	"encoding/json"
	"net"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// YassServer wraps up the listener, grpc server and cache service
type YassServer struct {
	lis   net.Listener
	srv   *grpc.Server
	cache model.Service
}

// New sets up the server
func New(lis net.Listener, cache model.Service) YassServer {
	s := grpc.NewServer()
	pb.RegisterCacheServer(s, server{Cache: cache})
	return YassServer{lis: lis, srv: s, cache: cache}
}

// Start starts the grpc server
func (y YassServer) Start() <-chan error {
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
	Cache    model.Service
	Follower model.Service
}

// Ping serves the healthcheck endpoint for the server
// It checks if the cache is serving too and responds accordingly
func (s server) Ping(context.Context, *pb.Null) (*pb.PingResponse, error) {
	logrus.Debug("Serving Ping request")
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
	logrus.Debug("Serving Set request")
	if pair.Key == "" || len(pair.Value) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Cannot set an empty key or value")
	}
	var value interface{}
	err := json.Unmarshal(pair.Value, &value)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "failed to unmarshal value")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-s.Cache.Set(pair.Key, value):
		if cacheResp.Err != nil {
			return nil, status.Error(codes.AlreadyExists, cacheResp.Err.Error())
		}
		output := &pb.Key{Key: cacheResp.Key}
		logrus.Debug("Set request succeeded")
		return output, nil
	}
}

// Get returns the value of a key
// It returns an error if the key is not in the cache
func (s server) Get(ctx context.Context, key *pb.Key) (*pb.Pair, error) {
	logrus.Debug("Serving Get request")
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot get an empty key")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-s.Cache.Get(key.Key):
		if cacheResp.Err != nil {
			return nil, status.Error(codes.NotFound, cacheResp.Err.Error())
		}
		value, err := json.Marshal(cacheResp.Value)
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to marshal data")
		}
		pair := &pb.Pair{Key: key.Key, Value: value}
		logrus.Debug("Get request succeeded")
		return pair, nil
	}
}

// Delete is the endpoint to delete a key/value if it is already in the cache
func (s server) Delete(ctx context.Context, key *pb.Key) (*pb.Null, error) {
	logrus.Info("Serving Delete request")
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
