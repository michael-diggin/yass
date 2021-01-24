package core

import (
	"context"
	"errors"

	"github.com/michael-diggin/yass/common/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/proto/convert"
	"github.com/michael-diggin/yass/server/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Set takes a key/value pair and adds it to the storage
// It returns an error if the key is already set
func (s *server) Set(ctx context.Context, req *pb.SetRequest) (*pb.Null, error) {
	logrus.Debug("Serving Set request")
	pbPair := req.GetPair()
	pair, err := convert.ToModel(pbPair)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if pair.Key == "" || pair.Value == nil {
		return nil, status.Error(codes.InvalidArgument, "Cannot set an empty key or value")
	}
	store, err := s.getStoreForRequest(req.GetReplica())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-store.Set(pair.Key, pair.Hash, pair.Value):
		if cacheResp.Err != nil {
			return nil, status.Error(codes.AlreadyExists, cacheResp.Err.Error())
		}
		logrus.Debug("Set request succeeded")
		return &pb.Null{}, nil
	}
}

// Get returns the value of a key
// It returns an error if the key is not in the storage
func (s *server) Get(ctx context.Context, req *pb.GetRequest) (*pb.Pair, error) {
	logrus.Debug("Serving Get request")
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot get an empty key")
	}
	store, err := s.getStoreForRequest(req.GetReplica())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-store.Get(req.Key):
		if cacheResp.Err != nil {
			return nil, status.Error(codes.NotFound, cacheResp.Err.Error())
		}
		pair, err := convert.ToPair(&models.Pair{Key: req.Key, Value: cacheResp.Value})
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to marshal data")
		}
		logrus.Debug("Get request succeeded")
		return pair, nil
	}
}

// Delete is the endpoint to delete a key/value if it is already in the storage
func (s *server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.Null, error) {
	logrus.Debug("Serving Delete request")
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot delete a zero key")
	}
	store, err := s.getStoreForRequest(req.GetReplica())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case <-store.Delete(req.Key):
		return &pb.Null{}, nil
	}
}

func (s *server) getStoreForRequest(idx int32) (model.Service, error) {
	if int(idx) >= len(s.DataStores) {
		return nil, errors.New("requested a datastore that does not exist")
	}
	return s.DataStores[int(idx)], nil
}
