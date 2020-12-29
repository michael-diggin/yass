// Package yass implements the client side of
// the gRPC service interface
package yass

import (
	"context"

	"github.com/michael-diggin/yass/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// StorageClient is a struct containing the grpc client
type StorageClient struct {
	grpcClient   pb.StorageClient
	healthClient grpc_health_v1.HealthClient
	conn         *grpc.ClientConn
}

// NewClient returns a new client that connects to the cache server
func NewClient(ctx context.Context, addr string) (*StorageClient, error) {
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure()) //TODO: add security and credentials
	if err != nil {
		return nil, err
	}
	client := pb.NewStorageClient(conn)
	health := grpc_health_v1.NewHealthClient(conn)
	return &StorageClient{grpcClient: client, healthClient: health, conn: conn}, nil
}

//Close tears down the underlying connection to the server
func (c StorageClient) Close() error {
	return c.conn.Close()
}

// Check performs a health check on the server
func (c StorageClient) Check(ctx context.Context) (bool, error) {
	resp, err := c.healthClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "Cache"})
	if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
		return true, nil
	}
	return false, err
}

// SetValue sets a key/value pair in the cache
func (c *StorageClient) SetValue(ctx context.Context, pair *models.Pair, rep models.Replica) error {
	pbPair, err := pb.ToPair(pair)
	if err != nil {
		return err
	}
	replica := pb.ToReplica(rep)
	req := &pb.SetRequest{Replica: replica, Pair: pbPair}
	_, err = c.grpcClient.Set(ctx, req)
	return err
}

// GetValue returns the value of a given key
func (c *StorageClient) GetValue(ctx context.Context, key string, rep models.Replica) (*models.Pair, error) {
	replica := pb.ToReplica(rep)
	req := &pb.GetRequest{Replica: replica, Key: key}
	pbPair, err := c.grpcClient.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return pbPair.ToModel()
}

// DelValue deletes a key/value pair
func (c *StorageClient) DelValue(ctx context.Context, key string, rep models.Replica) error {
	replica := pb.ToReplica(rep)
	req := &pb.DeleteRequest{Replica: replica, Key: key}
	_, err := c.grpcClient.Delete(ctx, req)
	return err
}

// BatchSet sets a batch of data into the storage server
func (c *StorageClient) BatchSet(ctx context.Context, replica int, data interface{}) error {
	reqData, ok := data.([]*pb.Pair)
	if !ok {
		return errors.New("invalid input")
	}
	req := &pb.BatchSetRequest{Replica: int32(replica), Data: reqData}
	_, err := c.grpcClient.BatchSet(ctx, req)
	return err
}

// BatchGet returns a batch of data
func (c *StorageClient) BatchGet(ctx context.Context, replica int) (interface{}, error) {
	req := &pb.BatchGetRequest{Replica: int32(replica)}
	resp, err := c.grpcClient.BatchGet(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
