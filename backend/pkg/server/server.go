package server

import (
	"context"

	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// New returns a new instance of the Server object
func New(cache backend.Service) Server {
	return Server{Cache: cache}
}

// Server implements the CacheServer interface
type Server struct {
	Cache backend.Service
}

// Ping serves the healthcheck endpoint for the server
// It checks if the cache is serving too and responds accordingly
func (s Server) Ping(context.Context, *pb.Null) (*pb.PingResponse, error) {
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
func (s Server) Set(ctx context.Context, pair *pb.Pair) (*pb.Key, error) {
	if pair.Key == "" || pair.Value == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot set a zero key or value")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-s.Cache.Set(ctx, pair.Key, pair.Value):
		if cacheResp.Err != nil {
			return nil, status.Error(codes.AlreadyExists, cacheResp.Err.Error())
		}
		output := &pb.Key{Key: cacheResp.Key}
		return output, nil
	}
}

// Get returns the value of a key
// It returns an error if the key is not in the cache
func (s Server) Get(ctx context.Context, key *pb.Key) (*pb.Pair, error) {
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot get a zero key")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-s.Cache.Get(ctx, key.Key):
		if cacheResp.Err != nil {
			logrus.Errorf("Tried to get value with not set key: %s", key.Key)
			return nil, status.Error(codes.NotFound, cacheResp.Err.Error())
		}
		pair := &pb.Pair{Key: key.Key, Value: cacheResp.Value}
		return pair, nil
	}
}

// Delete is the endpoint to delete a key/value if it is already in the cache
func (s Server) Delete(ctx context.Context, key *pb.Key) (*pb.Null, error) {
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot delete a zero key")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-s.Cache.Delete(ctx, key.Key):
		if cacheResp.Err != nil {
			logrus.Errorf("Could not delete key as does not exist: %s", key.Key)
			return nil, status.Error(codes.NotFound, cacheResp.Err.Error())
		}
		return &pb.Null{}, nil
	}
}
