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
	if resp == nil || err != nil {
		return false, err
	}
	if resp.Status == grpc_health_v1.HealthCheckResponse_SERVING {
		return true, nil
	}
	return false, err
}

// SetValue sets a key/value pair in the cache
func (c *StorageClient) SetValue(ctx context.Context, pair *models.Pair, rep int) error {
	pbPair, err := pb.ToPair(pair)
	if err != nil {
		return err
	}
	req := &pb.SetRequest{Replica: int32(rep), Pair: pbPair}
	_, err = c.grpcClient.Set(ctx, req)
	return err
}

// GetValue returns the value of a given key
func (c *StorageClient) GetValue(ctx context.Context, key string, rep int) (*models.Pair, error) {
	req := &pb.GetRequest{Replica: int32(rep), Key: key}
	pbPair, err := c.grpcClient.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return pbPair.ToModel()
}

// DelValue deletes a key/value pair
func (c *StorageClient) DelValue(ctx context.Context, key string, rep int) error {
	req := &pb.DeleteRequest{Replica: int32(rep), Key: key}
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

// BatchSend sends a batch of data where the keys lie between the hash values from
// node to another node
// it is used for rebalancing partitions and repopulating failed nodes
func (c *StorageClient) BatchSend(ctx context.Context, replica, toReplica int, addr string,
	low, high uint32) error {
	req := &pb.BatchSendRequest{
		Replica:   int32(replica),
		Address:   addr,
		ToReplica: int32(toReplica),
		Low:       low,
		High:      high,
		Delete:    true,
	}
	_, err := c.grpcClient.BatchSend(ctx, req)
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
