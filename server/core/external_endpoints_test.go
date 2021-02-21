package core

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/michael-diggin/yass/common/mocks"
	"github.com/michael-diggin/yass/common/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerPut(t *testing.T) {
	key := "test"
	hashkey := uint32(100)
	valueBytes := []byte(`"test-value"`)
	pair := &pb.Pair{Key: key, Hash: hashkey, Value: valueBytes}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRing := mocks.NewMockHashRing(ctrl)
		mockRing.EXPECT().Hash(key).Return(hashkey)
		mockRing.EXPECT().Get(hashkey).Return(models.Node{Idx: 0})

		mockClientOne := mocks.NewMockStorageClient(ctrl)
		mockClientTwo := mocks.NewMockStorageClient(ctrl)
		mockClientThree := mocks.NewMockStorageClient(ctrl)

		mockClientOne.EXPECT().Set(gomock.Any(), &pb.SetRequest{Replica: 0, Pair: pair}).Return(nil, nil).AnyTimes()
		mockClientTwo.EXPECT().Set(gomock.Any(), &pb.SetRequest{Replica: 0, Pair: pair}).Return(nil, nil).AnyTimes()
		mockClientThree.EXPECT().Set(gomock.Any(), &pb.SetRequest{Replica: 0, Pair: pair}).Return(nil, nil).AnyTimes()

		srv := newServer(nil, nil)
		srv.hashRing = mockRing
		srv.nodeClients["node-0"] = &models.StorageClient{StorageClient: mockClientOne}
		srv.nodeClients["node-1"] = &models.StorageClient{StorageClient: mockClientTwo}
		srv.nodeClients["node-2"] = &models.StorageClient{StorageClient: mockClientTwo}

		req := &pb.Pair{Key: key, Value: valueBytes}
		_, err := srv.Put(context.Background(), req)

		require.NoError(t, err)
	})

	t.Run("quorum reached 2/3", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRing := mocks.NewMockHashRing(ctrl)
		mockRing.EXPECT().Hash(key).Return(hashkey)
		mockRing.EXPECT().Get(hashkey).Return(models.Node{Idx: 0})

		mockClientOne := mocks.NewMockStorageClient(ctrl)
		mockClientTwo := mocks.NewMockStorageClient(ctrl)
		mockClientThree := mocks.NewMockStorageClient(ctrl)

		transientErr := errors.New("transient error")
		mockClientOne.EXPECT().Set(gomock.Any(), &pb.SetRequest{Replica: 0, Pair: pair}).Return(nil, nil)
		mockClientTwo.EXPECT().Set(gomock.Any(), &pb.SetRequest{Replica: 0, Pair: pair}).Return(nil, transientErr).AnyTimes()
		mockClientThree.EXPECT().Set(gomock.Any(), &pb.SetRequest{Replica: 0, Pair: pair}).Return(nil, nil)

		srv := newServer(nil, nil)
		srv.hashRing = mockRing
		srv.nodeClients["node-0"] = &models.StorageClient{StorageClient: mockClientOne}
		srv.nodeClients["node-1"] = &models.StorageClient{StorageClient: mockClientTwo}
		srv.nodeClients["node-2"] = &models.StorageClient{StorageClient: mockClientThree}

		req := &pb.Pair{Key: key, Value: valueBytes}
		_, err := srv.Put(context.Background(), req)

		require.NoError(t, err)
	})

	t.Run("already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRing := mocks.NewMockHashRing(ctrl)
		mockRing.EXPECT().Hash(key).Return(hashkey)
		mockRing.EXPECT().Get(hashkey).Return(models.Node{Idx: 1})

		mockClientOne := mocks.NewMockStorageClient(ctrl)
		mockClientTwo := mocks.NewMockStorageClient(ctrl)
		mockClientThree := mocks.NewMockStorageClient(ctrl)

		errMock := status.Error(codes.AlreadyExists, "key in cache already")
		mockClientOne.EXPECT().Set(gomock.Any(), &pb.SetRequest{Replica: 1, Pair: pair}).Return(nil, nil).AnyTimes()
		mockClientTwo.EXPECT().Set(gomock.Any(), &pb.SetRequest{Replica: 1, Pair: pair}).Return(nil, errMock)
		mockClientThree.EXPECT().Set(gomock.Any(), &pb.SetRequest{Replica: 1, Pair: pair}).Return(nil, errMock)

		mockClientOne.EXPECT().Delete(gomock.Any(), &pb.DeleteRequest{Replica: 1, Key: key}).Return(nil, nil).AnyTimes()
		mockClientTwo.EXPECT().Delete(gomock.Any(), &pb.DeleteRequest{Replica: 1, Key: key}).Return(nil, nil).AnyTimes()
		mockClientThree.EXPECT().Delete(gomock.Any(), &pb.DeleteRequest{Replica: 1, Key: key}).Return(nil, nil).AnyTimes()

		srv := newServer(nil, nil)
		srv.hashRing = mockRing
		srv.nodeClients["node-0"] = &models.StorageClient{StorageClient: mockClientOne}
		srv.nodeClients["node-1"] = &models.StorageClient{StorageClient: mockClientTwo}
		srv.nodeClients["node-2"] = &models.StorageClient{StorageClient: mockClientThree}

		req := &pb.Pair{Key: key, Value: valueBytes}
		_, err := srv.Put(context.Background(), req)

		require.Error(t, err)
		require.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	t.Run("no key specified", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		srv := newServer(nil, nil)
		srv.minServers = 0

		req := &pb.Pair{Key: "", Value: []byte(`"test-value"`)}
		_, err := srv.Put(context.Background(), req)

		require.Error(t, err)
		require.Equal(t, codes.InvalidArgument, status.Code(err))
	})
}

func TestGatewayGetSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "test-get-key"
	value := []byte(`"test-value"`)
	pair := &pb.Pair{Key: key, Value: value}

	mockRing := mocks.NewMockHashRing(ctrl)
	mockRing.EXPECT().Hash(key).Return(uint32(100))
	mockRing.EXPECT().GetN(uint32(100), 2).Return(
		[]models.Node{
			{ID: "node-0", Idx: 0},
			{ID: "node-1", Idx: 1},
		}, nil)

	mockClientOne := mocks.NewMockStorageClient(ctrl)
	mockClientTwo := mocks.NewMockStorageClient(ctrl)

	mockClientOne.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(pair, nil).AnyTimes()
	mockClientTwo.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 1, Key: key}).Return(pair, nil).AnyTimes()

	srv := newServer(nil, nil)
	srv.hashRing = mockRing
	srv.nodeClients["node-0"] = &models.StorageClient{StorageClient: mockClientOne}
	srv.nodeClients["node-1"] = &models.StorageClient{StorageClient: mockClientTwo}
	srv.minServers = 2

	req := &pb.Key{Key: key}
	retPair, err := srv.Fetch(context.Background(), req)

	require.NoError(t, err)
	require.Equal(t, key, retPair.Key)
	require.Equal(t, value, retPair.Value)
}

func TestGatewayGetNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "test-get-key"

	mockRing := mocks.NewMockHashRing(ctrl)
	mockRing.EXPECT().Hash(key).Return(uint32(100))
	mockRing.EXPECT().GetN(uint32(100), 2).Return(
		[]models.Node{
			{ID: "node-0", Idx: 0},
			{ID: "node-1", Idx: 1},
		}, nil)

	mockClientOne := mocks.NewMockStorageClient(ctrl)
	mockClientTwo := mocks.NewMockStorageClient(ctrl)

	errMock := status.Error(codes.NotFound, "key not found in cache")

	mockClientOne.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(nil, errMock).AnyTimes()
	mockClientTwo.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 1, Key: key}).Return(nil, errMock).AnyTimes()

	srv := newServer(nil, nil)
	srv.hashRing = mockRing
	srv.nodeClients["node-0"] = &models.StorageClient{StorageClient: mockClientOne}
	srv.nodeClients["node-1"] = &models.StorageClient{StorageClient: mockClientTwo}
	srv.minServers = 2

	req := &pb.Key{Key: key}
	retPair, err := srv.Fetch(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, retPair)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestGatewayGetOneSuccessOneFailure(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientOne := mocks.NewMockStorageClient(ctrl)
	mockClientTwo := mocks.NewMockStorageClient(ctrl)
	mockClientThree := mocks.NewMockStorageClient(ctrl)

	key := "test-get-key"
	value := []byte(`"test-value"`)
	pair := &pb.Pair{Key: key, Value: value}
	errMock := status.Error(codes.Canceled, "request timed out")

	mockRing := mocks.NewMockHashRing(ctrl)
	mockRing.EXPECT().Hash(key).Return(uint32(100))
	mockRing.EXPECT().GetN(uint32(100), 2).Return(
		[]models.Node{
			{ID: "node-3", Idx: 0},
			{ID: "node-1", Idx: 1},
		}, nil)

	mockClientThree.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(nil, errMock).AnyTimes()
	mockClientOne.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 1, Key: key}).
		DoAndReturn(func(...interface{}) (*pb.Pair, error) {
			time.Sleep(50 * time.Millisecond)
			return pair, nil
		}).AnyTimes()

	srv := newServer(nil, nil)
	srv.hashRing = mockRing
	srv.nodeClients["node-1"] = &models.StorageClient{StorageClient: mockClientOne}
	srv.nodeClients["node-2"] = &models.StorageClient{StorageClient: mockClientTwo}
	srv.nodeClients["node-3"] = &models.StorageClient{StorageClient: mockClientThree}

	req := &pb.Key{Key: key}
	retPair, err := srv.Fetch(context.Background(), req)

	require.NoError(t, err)
	require.Equal(t, key, retPair.Key)
	require.Equal(t, value, retPair.Value)
}
