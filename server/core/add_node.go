package core

import (
	"context"
	"time"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s server) AddNode(ctx context.Context, req *pb.AddNodeRequest) (*pb.Null, error) {
	logrus.Debug("Adding a new node")
	if req.Node == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot add a node without an address")
	}

	newCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	client, err := s.factory.New(newCtx, req.Node)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not connect to new node")
	}

	s.nodeClients[req.Node] = client
	s.hashRing.AddNode(req.Node)

	logrus.Debug("Successfully added a new node")
	return &pb.Null{}, nil
}
