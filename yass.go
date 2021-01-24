package yass

import (
	"context"

	"google.golang.org/grpc"

	"github.com/michael-diggin/yass/common/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/proto/convert"
)

// Client is a struct containing the grpc client
type Client struct {
	GrpcClient pb.YassServiceClient
	conn       *grpc.ClientConn
}

// NewClient returns a new client that connects to the Yass data nodes
func NewClient(ctx context.Context, addr string) (*Client, error) {
	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure()) //TODO: add security and credentials
	if err != nil {
		return nil, err
	}
	client := pb.NewYassServiceClient(conn)
	return &Client{GrpcClient: client, conn: conn}, nil
}

// Close tears down the underlying connection to the server
func (c *Client) Close() error {
	return c.conn.Close()
}

// Put sets a key/value pair
func (c *Client) Put(ctx context.Context, key string, value interface{}) error {
	pair := &models.Pair{Key: key, Value: value}
	pbPair, err := convert.ToPair(pair)
	if err != nil {
		return err
	}
	_, err = c.GrpcClient.Put(ctx, pbPair)
	return err
}

// Fetch returns the value of a given key
func (c *Client) Fetch(ctx context.Context, key string) (interface{}, error) {
	req := &pb.Key{Key: key}
	pbPair, err := c.GrpcClient.Retrieve(ctx, req)
	if err != nil {
		return nil, err
	}
	pair, err := convert.ToModel(pbPair)
	if err != nil {
		return nil, err
	}
	return pair.Value, nil
}
