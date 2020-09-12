package main

import (
	"context"
	"net"
	"testing"

	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend/pkg/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func getBufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestIntegrationServer(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	cache, err := redis.New("", "", "localhost:6379") //requires redis running - docker
	if err != nil {
		t.Fatalf("Cannot connect to redis for integration test: %v", err)
	}

	srv := YassServer{}
	srv.Init(lis, cache)
	srv.SpinUp()
	defer srv.ShutDown()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(getBufDialer(lis)), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Could not dial server: %err", err)
	}

	client := pb.NewCacheClient(conn)
	resp, err := client.Set(ctx, &pb.Pair{Key: "test-key", Value: "test-value"})
	if err != nil {
		t.Fatalf("Could not send Set command: %v", err)
	}
	if resp.Key != "test-key" {
		t.Fatalf("Set command returned '%s', not 'test-key'", resp.Key)
	}

}
