package core

import (
	"context"
	"net"

	"github.com/michael-diggin/yass/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// YassServer wraps up the listener, grpc server and cache service
type YassServer struct {
	lis net.Listener
	srv *grpc.Server
}

// New sets up the server
func New(lis net.Listener, leader, follower model.Service) YassServer {
	s := grpc.NewServer()
	pb.RegisterCacheServer(s, server{Leader: leader, Follower: follower})
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
	Leader   model.Service
	Follower model.Service
}

// Ping serves the healthcheck endpoint for the server
// It checks if the storage is serving and responds accordingly
func (s server) Ping(context.Context, *pb.Null) (*pb.PingResponse, error) {
	logrus.Debug("Serving Ping request")
	err := s.Leader.Ping()
	if err != nil {
		resp := &pb.PingResponse{Status: pb.PingResponse_NOT_SERVING}
		return resp, status.Error(codes.Unavailable, err.Error())
	}
	err = s.Follower.Ping()
	if err != nil {
		resp := &pb.PingResponse{Status: pb.PingResponse_NOT_SERVING}
		return resp, status.Error(codes.Unavailable, err.Error())
	}
	resp := &pb.PingResponse{Status: pb.PingResponse_SERVING}
	return resp, nil
}

// Set takes a key/value pair and adds it to the storage
// It returns an error if the key is already set
func (s server) Set(ctx context.Context, req *pb.Pair) (*pb.Key, error) {
	logrus.Debug("Serving Set request")
	if req.Key == "" || len(req.Value) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Cannot set an empty key or value")
	}
	pair, err := req.ToModel()
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-s.Leader.Set(pair.Key, pair.Value):
		if cacheResp.Err != nil {
			return nil, status.Error(codes.AlreadyExists, cacheResp.Err.Error())
		}
		output := &pb.Key{Key: cacheResp.Key}
		logrus.Debug("Set request succeeded")
		return output, nil
	}
}

// Get returns the value of a key
// It returns an error if the key is not in the storage
func (s server) Get(ctx context.Context, req *pb.Key) (*pb.Pair, error) {
	logrus.Debug("Serving Get request")
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot get an empty key")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case cacheResp := <-s.Leader.Get(req.Key):
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
func (s server) Delete(ctx context.Context, req *pb.Key) (*pb.Null, error) {
	logrus.Info("Serving Delete request")
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot delete a zero key")
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case <-s.Leader.Delete(req.Key):
		return &pb.Null{}, nil
	}
}
