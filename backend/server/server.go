package server

import (
	"context"

	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// New returns a new instance of the Server object
func New(cache model.Service) Server {
	return Server{Cache: cache}
}

// Server implements the CacheServer interface
type Server struct {
	Cache model.Service
}

// Add takes a key/value pair and adds it to the cache storage
// It returns an error if the key is already set
func (s Server) Add(ctx context.Context, pair *pb.Pair) (*pb.Key, error) {
	if pair.Key == "" || pair.Value == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot set a zero key or value")
	}
	key, err := s.Cache.Set(ctx, pair.Key, pair.Value)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	output := &pb.Key{Key: key}
	return output, nil
}

// Get returns the value of a key
// It returns an error if the key is not in the cache
func (s Server) Get(ctx context.Context, key *pb.Key) (*pb.Pair, error) {
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot get a zero key")
	}
	value, err := s.Cache.Get(ctx, key.Key)
	if err != nil {
		logrus.Errorf("Tried to get value with not set key: %s", key.Key)
		return nil, status.Error(codes.NotFound, err.Error())
	}
	pair := &pb.Pair{Key: key.Key, Value: value}
	return pair, nil
}

// Delete is the endpoint to delete a key/value if it is already in the cache
func (s Server) Delete(ctx context.Context, key *pb.Key) (*pb.Null, error) {
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot delete a zero key")
	}
	err := s.Cache.Delete(ctx, key.Key)
	if err != nil {
		logrus.Errorf("Could not delete key as does not exist: %s", key.Key)
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &pb.Null{}, nil
}
