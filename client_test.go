package yass

import (
	"context"
	"errors"
	"testing"

	"github.com/michael-diggin/yass/mocks"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/stretchr/testify/require"
)

func TestClientPing(t *testing.T) {
	tt := []struct {
		name    string
		serving bool
	}{
		{"serving", true},
		{"not serving", false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mockgRPC := &mocks.MockClient{}
			mockgRPC.PingFn = func() error {
				if tc.name == "serving" {
					return nil
				}
				return errors.New("Not serving")
			}
			cc := CacheClient{grpcClient: mockgRPC, conn: nil}
			ok, _ := cc.Ping(context.Background())

			require.Equal(t, tc.serving, ok)
			require.True(t, mockgRPC.PingInvoked)
		})
	}
}

func TestClientSetValue(t *testing.T) {
	mockgRPC := &mocks.MockClient{}
	mockgRPC.SetFn = func(ctx context.Context, key string, value []byte) error {
		return nil
	}
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	key := "test"
	val := "value"
	err := cc.SetValue(context.Background(), key, val)

	require.NoError(t, err)
	require.True(t, mockgRPC.SetInvoked)
}

func TestClientGetValue(t *testing.T) {
	errTest := errors.New("Not in storage")

	tt := []struct {
		name  string
		key   string
		value string
		err   error
	}{
		{"valid case", "test", "value", nil},
		{"err case", "bad", "", errTest},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mockgRPC := &mocks.MockClient{}
			mockgRPC.GetFn = func(ctx context.Context, key string) ([]byte, error) {
				if key == "test" {
					return []byte(`"value"`), nil
				}
				return []byte{}, errTest
			}
			cc := CacheClient{grpcClient: mockgRPC, conn: nil}
			val, err := cc.GetValue(context.Background(), tc.key)

			require.Equal(t, err, tc.err)
			require.Equal(t, tc.value, val)
			require.True(t, mockgRPC.GetInvoked)
		})
	}
}

func TestClientDelValue(t *testing.T) {
	mockgRPC := &mocks.MockClient{}
	mockgRPC.DelFn = func(ctx context.Context, key string) error {
		return nil
	}
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	key := "test"
	err := cc.DelValue(context.TODO(), key)

	require.NoError(t, err)
	require.True(t, mockgRPC.DelInvoked)
}

func TestClientSetFollowerValue(t *testing.T) {
	mockgRPC := &mocks.MockClient{}
	mockgRPC.SetFollowerFn = func(ctx context.Context, key string, value []byte) error {
		return nil
	}
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	key := "test"
	val := "value"
	err := cc.SetFollowerValue(context.Background(), key, val)

	require.NoError(t, err)
	require.True(t, mockgRPC.SetFollowerInvoked)
}

func TestClientGetFollowerValue(t *testing.T) {
	errTest := errors.New("Not in storage")

	tt := []struct {
		name  string
		key   string
		value string
		err   error
	}{
		{"valid case", "test", "value", nil},
		{"err case", "bad", "", errTest},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mockgRPC := &mocks.MockClient{}
			mockgRPC.GetFollowerFn = func(ctx context.Context, key string) ([]byte, error) {
				if key == "test" {
					return []byte(`"value"`), nil
				}
				return []byte{}, errTest
			}
			cc := CacheClient{grpcClient: mockgRPC, conn: nil}
			val, err := cc.GetFollowerValue(context.Background(), tc.key)

			require.Equal(t, err, tc.err)
			require.Equal(t, tc.value, val)
			require.True(t, mockgRPC.GetFollowerInvoked)
		})
	}
}

func TestClientDelFollowerValue(t *testing.T) {
	mockgRPC := &mocks.MockClient{}
	mockgRPC.DelFollowerFn = func(ctx context.Context, key string) error {
		return nil
	}
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	key := "test"
	err := cc.DelFollowerValue(context.Background(), key)

	require.NoError(t, err)
	require.True(t, mockgRPC.DelFollowerInvoked)
}

func TestBatchGet(t *testing.T) {
	mockgRPC := &mocks.MockClient{}
	testPair := &pb.Pair{Key: "test-key", Value: []byte(`"value"`)}
	mockgRPC.BatchGetFn = func(ctx context.Context) ([]*pb.Pair, error) {
		return []*pb.Pair{testPair}, nil
	}
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	resp, err := cc.BatchGet(context.Background(), 0)
	require.NoError(t, err)
	require.True(t, mockgRPC.BatchGetInvoked)

	respData, ok := resp.([]*pb.Pair)
	require.True(t, ok)

	require.Len(t, respData, 1)
	require.Equal(t, respData[0], testPair)
}
