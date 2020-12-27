package core

import (
	"context"
	"net"
	"testing"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/stretchr/testify/require"
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
	l := &mocks.TestStorage{
		PingFn: func() error {
			return nil
		},
	}
	f := &mocks.TestStorage{
		PingFn: func() error {
			return nil
		},
	}

	srv := New(lis, l, f)
	srv.Start()
	defer srv.ShutDown()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(getBufDialer(lis)), grpc.WithInsecure())
	require.NoError(t, err)

	client := pb.NewCacheClient(conn)
	resp, err := client.Ping(ctx, &pb.Null{})
	require.NoError(t, err)
	require.Equal(t, pb.PingResponse_SERVING, resp.Status)
	require.True(t, l.PingInvoked)
}
