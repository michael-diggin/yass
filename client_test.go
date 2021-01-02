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

func TestClientSetValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "test"
	hash := uint32(100)
	val := "value"

	mockgRPC := mocks.NewMockStorageClient(ctrl)
	pair := &pb.Pair{Key: key, Hash: hash, Value: []byte(`"value"`)}
	mockgRPC.EXPECT().Set(gomock.Any(), &pb.SetRequest{Replica: pb.Replica_MAIN, Pair: pair}).Return(nil, nil)
	cc := StorageClient{grpcClient: mockgRPC, conn: nil}

	err := cc.SetValue(context.Background(), &models.Pair{Key: key, Hash: hash, Value: val}, models.MainReplica)
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
			mockgRPC := mocks.NewMockStorageClient(ctrl)
			mockgRPC.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: pb.Replica_MAIN, Key: tc.key}).
				Return(&pb.Pair{Key: tc.key, Value: []byte(`"value"`)}, tc.err)

			cc := StorageClient{grpcClient: mockgRPC, conn: nil}
			resp, err := cc.GetValue(context.Background(), tc.key, models.MainReplica)

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
	mockgRPC := mocks.NewMockStorageClient(ctrl)
	key := "test"
	mockgRPC.EXPECT().Delete(gomock.Any(), &pb.DeleteRequest{Replica: pb.Replica_MAIN, Key: key}).
		Return(&pb.Null{}, nil)
	cc := StorageClient{grpcClient: mockgRPC, conn: nil}
	err := cc.DelValue(context.Background(), key, models.MainReplica)

	require.NoError(t, err)
}

func TestBatchGet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockgRPC := mocks.NewMockStorageClient(ctrl)
	testPair := &pb.Pair{Key: "test-key", Value: []byte(`"value"`)}
	mockgRPC.EXPECT().BatchGet(gomock.Any(), &pb.BatchGetRequest{Replica: 0}).
		Return(&pb.BatchGetResponse{Replica: 0, Data: []*pb.Pair{testPair}}, nil)
	cc := StorageClient{grpcClient: mockgRPC, conn: nil}
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
	mockgRPC := mocks.NewMockStorageClient(ctrl)
	testPair := &pb.Pair{Key: "test-key", Value: []byte(`"value"`)}
	mockgRPC.EXPECT().BatchSet(gomock.Any(), &pb.BatchSetRequest{Replica: 1, Data: []*pb.Pair{testPair}}).
		Return(&pb.Null{}, nil)
	cc := StorageClient{grpcClient: mockgRPC, conn: nil}
	err := cc.BatchSet(context.Background(), 1, []*pb.Pair{testPair})
	require.NoError(t, err)
}
