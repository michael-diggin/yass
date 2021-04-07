// Package client implements the client side of
// the gRPC service interface
package client

import (
	"context"

	"github.com/michael-diggin/yass/common/models"
	pb "github.com/michael-diggin/yass/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Factory implements the clientFactory interface
type Factory struct{}

// NewProtoClient returns a new gprc client
func (f Factory) NewProtoClient(ctx context.Context, addr string) (*models.StorageClient, error) {
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure()) //TODO: add security and credentials
	if err != nil {
		return nil, err
	}
	protoClient := pb.NewStorageClient(conn)
	yassClient := pb.NewYassServiceClient(conn)
	healthClient := grpc_health_v1.NewHealthClient(conn)
	return &models.StorageClient{StorageClient: protoClient,
		YassServiceClient: yassClient,
		HealthClient:      healthClient,
		ClientConn:        conn,
	}, nil
}
