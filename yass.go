// Package yass implements the client side of
// the gRPC service interface
package yass

import (
	"context"

	pb "github.com/michael-diggin/yass/api"
	"google.golang.org/grpc"
)

// CacheClient is a struct containing the grpc client
type CacheClient struct {
	grpcClient pb.CacheClient
	conn       *grpc.ClientConn
}

// NewClient returns a new client that connects to the cache server
func NewClient(addr string) (*CacheClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure()) //TODO: add security and credentials
	if err != nil {
		return &CacheClient{}, err
	}
	client := pb.NewCacheClient(conn)
	return &CacheClient{grpcClient: client, conn: conn}, nil
}

//Close tears down the underlying connection to the server
func (c CacheClient) Close() error {
	err := c.conn.Close()
	return err
}

// SetValue sets a key/value pair in the cache
func (c *CacheClient) SetValue(ctx context.Context, key, value string) error {
	pair := &pb.Pair{Key: key, Value: value}
	_, err := c.grpcClient.Add(ctx, pair)
	return err
}

// GetValue returns the value of a given key
func (c *CacheClient) GetValue(ctx context.Context, key string) (string, error) {
	pbKey := &pb.Key{Key: key}
	pbPair, err := c.grpcClient.Get(ctx, pbKey)
	if err != nil {
		return "", err
	}
	return pbPair.Value, nil
}

// DelValue deletes a key/value pair
func (c *CacheClient) DelValue(ctx context.Context, key string) error {
	pbKey := &pb.Key{Key: key}
	_, err := c.grpcClient.Delete(ctx, pbKey)
	return err
}
