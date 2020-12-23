package core

import (
	"context"
	"net"
	"testing"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func getBufDialer(listener *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(ctx context.Context, url string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestRunAndPingServer(t *testing.T) {
	lis := bufconn.Listen(1024 * 1024)
	cache := &mocks.TestCache{
		PingFn: func() error {
			return nil
		},
	}

	srv := New(lis, cache)
	srv.SpinUp()
	defer srv.ShutDown()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(getBufDialer(lis)), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Could not dial server: %err", err)
	}

	client := pb.NewCacheClient(conn)
	resp, err := client.Ping(ctx, &pb.Null{})
	if err != nil {
		t.Fatalf("Could not send Ping command: %v", err)
	}
	if resp.Status != pb.PingResponse_SERVING {
		t.Fatalf("Server not serving")
	}
	if !cache.PingInvoked {
		t.Fatalf("mock cache ping fn not callled")
	}
}
