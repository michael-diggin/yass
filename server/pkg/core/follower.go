package core

import (
	"context"
	"encoding/json"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SetFollower takes a key/value pair and adds it to the storage
// It returns an error if the key is already set
func (s server) SetFollower(ctx context.Context, pair *pb.Pair) (*pb.Key, error) {
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
	case cacheResp := <-s.Follower.Set(pair.Key, value):
		if cacheResp.Err != nil {
			return nil, status.Error(codes.AlreadyExists, cacheResp.Err.Error())
		}
		output := &pb.Key{Key: cacheResp.Key}
		logrus.Debug("Set request succeeded")
		return output, nil
	}
}

// GetFollower returns the value of a key
// It returns an error if the key is not in the storage
func (s server) GetFollower(ctx context.Context, key *pb.Key) (*pb.Pair, error) {
	logrus.Debug("Serving GetFollower request")
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot get an empty key")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-s.Follower.Get(key.Key):
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

// DeleteFollower is the endpoint to delete a key/value if it is already in the storage
func (s server) DeleteFollower(ctx context.Context, key *pb.Key) (*pb.Null, error) {
	logrus.Info("Serving Delete request")
	if key.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot delete a zero key")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case <-s.Follower.Delete(key.Key):
		return &pb.Null{}, nil
	}
}
