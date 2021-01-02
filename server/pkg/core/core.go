package core

import (
	"context"
	"errors"
	"net"

	"github.com/michael-diggin/yass/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// YassServer wraps up the listener, grpc server and cache service
type YassServer struct {
	lis net.Listener
	srv *grpc.Server
}

// New sets up the server
func New(lis net.Listener, dataStores ...model.Service) YassServer {
	s := grpc.NewServer()
	srv := server{DataStores: dataStores}
	pb.RegisterStorageServer(s, srv)
	grpc_health_v1.RegisterHealthServer(s, &srv)
	return YassServer{lis: lis, srv: s}
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
	y.srv.GracefulStop()
	logrus.Infof("gRPC server stopped")
}

// server (unexported) implements the CacheServer interface
type server struct {
	DataStores []model.Service
}

// Set takes a key/value pair and adds it to the storage
// It returns an error if the key is already set
func (s server) Set(ctx context.Context, req *pb.SetRequest) (*pb.Null, error) {
	logrus.Debug("Serving Set request")
	pbPair := req.GetPair()
	pair, err := pbPair.ToModel()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if pair.Key == "" || pair.Value == nil {
		return nil, status.Error(codes.InvalidArgument, "Cannot set an empty key or value")
	}
	store, err := s.getStoreForRequest(int(req.GetReplica()))
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
func (s server) Get(ctx context.Context, req *pb.GetRequest) (*pb.Pair, error) {
	logrus.Debug("Serving Get request")
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot get an empty key")
	}
	store, err := s.getStoreForRequest(int(req.GetReplica()))
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
		pair, err := pb.ToPair(&models.Pair{Key: req.Key, Value: cacheResp.Value})
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to marshal data")
		}
		logrus.Debug("Get request succeeded")
		return pair, nil
	}
}

// Delete is the endpoint to delete a key/value if it is already in the storage
func (s server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.Null, error) {
	logrus.Debug("Serving Delete request")
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot delete a zero key")
	}
	store, err := s.getStoreForRequest(int(req.GetReplica()))
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

func (s server) getStoreForRequest(idx int) (model.Service, error) {
	if idx >= len(s.DataStores) {
		return nil, errors.New("requested a datastore that does not exist")
	}
	return s.DataStores[idx], nil
}
