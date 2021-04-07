package core

import (
	"context"
	"time"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *server) AddNode(ctx context.Context, req *pb.AddNodeRequest) (*pb.Null, error) {
	logrus.Debug("Adding a new node")
	if req.Node == "" {
		return nil, status.Error(codes.InvalidArgument, "Cannot add a node without an address")
	}

	newCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	client, err := s.factory.NewProtoClient(newCtx, req.Node)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not connect to new node")
	}

	if _, ok := s.nodeClients[req.Node]; !ok {
		s.hashRing.AddNode(req.Node)
		s.nodeClients[req.Node] = client
		logrus.Infof("Successfully added a new node: %s", req.Node)
	} else {
		logrus.Infof("Reconnected to node %s", req.Node)
	}

	return &pb.Null{}, nil
}
