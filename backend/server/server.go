package server

import (
	"context"

	pb "github.com/michael-diggin/yass/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type memory map[string]string

// New returns a new instance of the Server object
func New(cache memory) Server {
	return Server{Cache: cache}
}

// Server implements the CacheServer interface
type Server struct {
	Cache map[string]string
}

// Add takes a key/value pair and adds it to the cache storage
// It returns an error if the key is already set
func (s Server) Add(ctx context.Context, pair *pb.Pair) (*pb.Key, error) {
	if pair.Key == "" || pair.Value == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot set a zero key or value")
	}
	_, ok := s.Cache[pair.Key]
	if ok {
		logrus.Errorf("Tried to reset key: %s", pair.Key)
		return nil, status.Error(codes.AlreadyExists, "Key is already in the cache")
	}
	s.Cache[pair.Key] = pair.Value
	output := &pb.Key{Key: pair.Key}
	return output, nil
}

// Get returns the value of a key
// It returns an error if the key is not in the cache
func (s Server) Get(ctx context.Context, key *pb.Key) (*pb.Pair, error) {
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot get a zero key")
	}
	res, ok := s.Cache[key.Key]
	if !ok {
		logrus.Errorf("Tried to get value with not set key: %s", key.Key)
		return nil, status.Error(codes.NotFound, "Key is not in the cache")
	}
	pair := &pb.Pair{Key: key.Key, Value: res}
	return pair, nil
}

// Reset is the endpoint for reset a keys value if it is already in the cache
func (s Server) Reset(ctx context.Context, pair *pb.Pair) (*pb.Key, error) {
	if pair.Key == "" || pair.Value == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot set a zero key or value")
	}
	if _, ok := s.Cache[pair.Key]; !ok {
		logrus.Errorf("Could not reset key as does not exist: %s", pair.Key)
		return nil, status.Error(codes.NotFound, "Key is not in the cache")
	}
	s.Cache[pair.Key] = pair.Value
	output := &pb.Key{Key: pair.Key}
	return output, nil
}

// Delete is the endpoint to delete a key/value if it is already in the cache
func (s Server) Delete(ctx context.Context, key *pb.Key) (*pb.Null, error) {
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot delete a zero key")
	}
	if _, ok := s.Cache[key.Key]; !ok {
		logrus.Errorf("Could not delete key as does not exist: %s", key.Key)
		return nil, status.Error(codes.NotFound, "Key is not in the cache")
	}
	delete(s.Cache, key.Key)
	return &pb.Null{}, nil
}
