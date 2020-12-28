package yass

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/michael-diggin/yass/mocks"
	"github.com/michael-diggin/yass/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/stretchr/testify/require"
)

func TestClientPing(t *testing.T) {
	tt := []struct {
		name    string
		status  pb.PingResponse_ServingStatus
		serving bool
	}{
		{"serving", pb.PingResponse_SERVING, true},
		{"not serving", pb.PingResponse_NOT_SERVING, false},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockgRPC := mocks.NewMockCacheClient(ctrl)
			mockgRPC.EXPECT().Ping(gomock.Any(), gomock.Any()).
				Return(&pb.PingResponse{Status: tc.status}, nil)
			cc := CacheClient{grpcClient: mockgRPC, conn: nil}
			ok, _ := cc.Ping(context.Background())

			require.Equal(t, tc.serving, ok)
		})
	}
}

func TestClientSetValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "test"
	val := "value"

	mockgRPC := mocks.NewMockCacheClient(ctrl)
	mockgRPC.EXPECT().Set(gomock.Any(), &pb.Pair{Key: key, Value: []byte(`"value"`)}).Return(nil, nil)
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}

	err := cc.SetValue(context.Background(), &models.Pair{Key: key, Value: val})
	require.NoError(t, err)
}

func TestClientGetValue(t *testing.T) {
	errTest := errors.New("Not in storage")

	tt := []struct {
		name  string
		key   string
		value interface{}
		err   error
	}{
		{"valid case", "test", "value", nil},
		{"err case", "bad", nil, errTest},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockgRPC := mocks.NewMockCacheClient(ctrl)
			mockgRPC.EXPECT().Get(gomock.Any(), &pb.Key{Key: tc.key}).
				Return(&pb.Pair{Key: tc.key, Value: []byte(`"value"`)}, tc.err)

			cc := CacheClient{grpcClient: mockgRPC, conn: nil}
			resp, err := cc.GetValue(context.Background(), tc.key)

			require.Equal(t, err, tc.err)
			if tc.value != nil {
				require.Equal(t, tc.value, resp.Value)
			}
		})
	}
}

func TestClientDelValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockgRPC := mocks.NewMockCacheClient(ctrl)
	key := "test"
	mockgRPC.EXPECT().Delete(gomock.Any(), &pb.Key{Key: key}).Return(&pb.Null{}, nil)
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	err := cc.DelValue(context.TODO(), key)

	require.NoError(t, err)
}

func TestClientSetFollowerValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockgRPC := mocks.NewMockCacheClient(ctrl)

	key := "test"
	val := "value"
	mockgRPC.EXPECT().SetFollower(gomock.Any(), &pb.Pair{Key: key, Value: []byte(`"value"`)}).
		Return(&pb.Key{Key: key}, nil)
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}

	err := cc.SetFollowerValue(context.Background(), key, val)
	require.NoError(t, err)
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
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockgRPC := mocks.NewMockCacheClient(ctrl)
			mockgRPC.EXPECT().GetFollower(gomock.Any(), &pb.Key{Key: tc.key}).
				Return(&pb.Pair{Key: tc.key, Value: []byte(`"value"`)}, tc.err)

			cc := CacheClient{grpcClient: mockgRPC, conn: nil}
			val, err := cc.GetFollowerValue(context.Background(), tc.key)

			require.Equal(t, err, tc.err)
			require.Equal(t, tc.value, val)
		})
	}
}

func TestClientDelFollowerValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockgRPC := mocks.NewMockCacheClient(ctrl)
	key := "test"
	mockgRPC.EXPECT().DeleteFollower(gomock.Any(), &pb.Key{Key: key}).Return(&pb.Null{}, nil)
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	err := cc.DelFollowerValue(context.TODO(), key)

	require.NoError(t, err)
}

func TestBatchGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockgRPC := mocks.NewMockCacheClient(ctrl)
	testPair := &pb.Pair{Key: "test-key", Value: []byte(`"value"`)}
	mockgRPC.EXPECT().BatchGet(gomock.Any(), &pb.BatchGetRequest{Replica: 0}).
		Return(&pb.BatchGetResponse{Replica: 0, Data: []*pb.Pair{testPair}}, nil)
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	resp, err := cc.BatchGet(context.Background(), 0)
	require.NoError(t, err)

	respData, ok := resp.([]*pb.Pair)
	require.True(t, ok)

	require.Len(t, respData, 1)
	require.Equal(t, respData[0], testPair)
}

func TestBatchSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockgRPC := mocks.NewMockCacheClient(ctrl)
	testPair := &pb.Pair{Key: "test-key", Value: []byte(`"value"`)}
	mockgRPC.EXPECT().BatchSet(gomock.Any(), &pb.BatchSetRequest{Replica: 1, Data: []*pb.Pair{testPair}}).
		Return(&pb.Null{}, nil)
	cc := CacheClient{grpcClient: mockgRPC, conn: nil}
	err := cc.BatchSet(context.Background(), 1, []*pb.Pair{testPair})
	require.NoError(t, err)
}
