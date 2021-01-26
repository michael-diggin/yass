package models

import (
	"context"

	pb "github.com/michael-diggin/yass/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// StorageClient satisfies the internal proto interface
type StorageClient struct {
	pb.StorageClient
	grpc_health_v1.HealthClient
	*grpc.ClientConn
}

//go:generate mockgen -destination=../mocks/mock_client_factory.go -package=mocks . ClientFactory

// ClientFactory is the interface for creating a new instance of the Client Interface
type ClientFactory interface {
	NewProtoClient(context.Context, string) (*StorageClient, error)
}
