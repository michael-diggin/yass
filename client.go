// Package yass implements the client side of
// the gRPC service interface
package yass

import (
	"context"
	"encoding/json"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// CacheClient is a struct containing the grpc client
type CacheClient struct {
	grpcClient pb.CacheClient
	conn       *grpc.ClientConn
}

// NewClient returns a new client that connects to the cache server
func NewClient(ctx context.Context, addr string) (*CacheClient, error) {
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure()) //TODO: add security and credentials
	if err != nil {
		return nil, err
	}
	client := pb.NewCacheClient(conn)
	return &CacheClient{grpcClient: client, conn: conn}, nil
}

//Close tears down the underlying connection to the server
func (c CacheClient) Close() error {
	return c.conn.Close()
}

// Ping checls if the server and cache are serving
func (c CacheClient) Ping(ctx context.Context) (bool, error) {
	resp, err := c.grpcClient.Ping(ctx, &pb.Null{})
	if resp.Status == pb.PingResponse_SERVING {
		return true, nil
	}
	return false, err
}

// SetValue sets a key/value pair in the cache
func (c *CacheClient) SetValue(ctx context.Context, key string, value interface{}) error {
	bytesValue, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "failed to marshal value")
	}
	pair := &pb.Pair{Key: key, Value: bytesValue}
	_, err = c.grpcClient.Set(ctx, pair)
	return err
}

// GetValue returns the value of a given key
func (c *CacheClient) GetValue(ctx context.Context, key string) (interface{}, error) {
	pbKey := &pb.Key{Key: key}
	pbPair, err := c.grpcClient.Get(ctx, pbKey)
	if err != nil {
		return "", err
	}
	var value interface{}
	err = json.Unmarshal(pbPair.Value, &value)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal result")
	}
	return value, nil
}

// DelValue deletes a key/value pair
func (c *CacheClient) DelValue(ctx context.Context, key string) error {
	pbKey := &pb.Key{Key: key}
	_, err := c.grpcClient.Delete(ctx, pbKey)
	return err
}
