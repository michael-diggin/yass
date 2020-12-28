// Package yass implements the client side of
// the gRPC service interface
package yass

import (
	"context"
	"encoding/json"

	"github.com/michael-diggin/yass/models"
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
func (c *CacheClient) SetValue(ctx context.Context, pair *models.Pair) error {
	pbPair, err := pb.ToPair(pair)
	if err != nil {
		return err
	}
	req := &pb.SetRequest{Replica: pb.Replica_MAIN, Pair: pbPair}
	_, err = c.grpcClient.Set(ctx, req)
	return err
}

// GetValue returns the value of a given key
func (c *CacheClient) GetValue(ctx context.Context, key string) (*models.Pair, error) {
	req := &pb.GetRequest{Replica: pb.Replica_MAIN, Key: key}
	pbPair, err := c.grpcClient.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	return pbPair.ToModel()
}

// DelValue deletes a key/value pair
func (c *CacheClient) DelValue(ctx context.Context, key string) error {
	req := &pb.DeleteRequest{Replica: pb.Replica_MAIN, Key: key}
	_, err := c.grpcClient.Delete(ctx, req)
	return err
}

// SetFollowerValue sets a key/value pair in the follower partition
func (c *CacheClient) SetFollowerValue(ctx context.Context, key string, value interface{}) error {
	bytesValue, err := json.Marshal(value)
	if err != nil {
		return errors.Wrap(err, "failed to marshal value")
	}
	pair := &pb.Pair{Key: key, Value: bytesValue}
	_, err = c.grpcClient.SetFollower(ctx, pair)
	return err
}

// GetFollowerValue returns the value of a given key
func (c *CacheClient) GetFollowerValue(ctx context.Context, key string) (interface{}, error) {
	pbKey := &pb.Key{Key: key}
	pbPair, err := c.grpcClient.GetFollower(ctx, pbKey)
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

// DelFollowerValue deletes a key/value pair
func (c *CacheClient) DelFollowerValue(ctx context.Context, key string) error {
	pbKey := &pb.Key{Key: key}
	_, err := c.grpcClient.DeleteFollower(ctx, pbKey)
	return err
}

// BatchSet sets a batch of data into the storage server
func (c *CacheClient) BatchSet(ctx context.Context, replica int, data interface{}) error {
	reqData, ok := data.([]*pb.Pair)
	if !ok {
		return errors.New("invalid input")
	}
	req := &pb.BatchSetRequest{Replica: int32(replica), Data: reqData}
	_, err := c.grpcClient.BatchSet(ctx, req)
	return err
}

// BatchGet returns a batch of data
func (c *CacheClient) BatchGet(ctx context.Context, replica int) (interface{}, error) {
	req := &pb.BatchGetRequest{Replica: int32(replica)}
	resp, err := c.grpcClient.BatchGet(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}
