package core

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/golang/mock/gomock"

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

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLeader := mocks.NewMockService(ctrl)
	mockLeader.EXPECT().Ping().Return(nil)
	mockFollower := mocks.NewMockService(ctrl)
	mockFollower.EXPECT().Ping().Return(nil)

	srv := New(lis, mockLeader, mockFollower)
	srv.Start()
	defer srv.ShutDown()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "", grpc.WithContextDialer(getBufDialer(lis)), grpc.WithInsecure())
	require.NoError(t, err)

	health := grpc_health_v1.NewHealthClient(conn)

	resp, err := health.Check(ctx, &grpc_health_v1.HealthCheckRequest{Service: "Storage"})
	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, resp.Status)
}
