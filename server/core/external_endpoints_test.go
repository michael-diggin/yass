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

func testServer() *server {
	return newServer(nil, "name", "leader", nil)
}

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

		srv := testServer()
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

		srv := testServer()
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

		srv := testServer()
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

		srv := testServer()
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
	mockRing.EXPECT().Get(uint32(100)).Return(models.Node{Idx: 0})

	mockClientOne := mocks.NewMockStorageClient(ctrl)
	mockClientTwo := mocks.NewMockStorageClient(ctrl)
	mockClientThree := mocks.NewMockStorageClient(ctrl)

	mockClientOne.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(pair, nil).AnyTimes()
	mockClientTwo.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(pair, nil).AnyTimes()
	mockClientThree.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(pair, nil).AnyTimes()

	srv := testServer()
	srv.hashRing = mockRing
	srv.nodeClients["node-0"] = &models.StorageClient{StorageClient: mockClientOne}
	srv.nodeClients["node-1"] = &models.StorageClient{StorageClient: mockClientTwo}
	srv.nodeClients["node-2"] = &models.StorageClient{StorageClient: mockClientThree}

	req := &pb.Key{Key: key}
	retPair, err := srv.Fetch(context.Background(), req)

	require.NoError(t, err)
	require.Equal(t, key, retPair.Key)
	require.Equal(t, value, retPair.Value)
}

func TestGatewayFetchQuorumReached(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "test-get-key"
	value := []byte(`"test-value"`)
	pair := &pb.Pair{Key: key, Value: value}

	mockRing := mocks.NewMockHashRing(ctrl)
	mockRing.EXPECT().Hash(key).Return(uint32(100))
	mockRing.EXPECT().Get(uint32(100)).Return(models.Node{Idx: 0})

	mockClientOne := mocks.NewMockStorageClient(ctrl)
	mockClientTwo := mocks.NewMockStorageClient(ctrl)
	mockClientThree := mocks.NewMockStorageClient(ctrl)

	mockErr := errors.New("not found")
	mockClientOne.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).
		DoAndReturn(func(...interface{}) (*pb.Pair, error) {
			time.Sleep(50 * time.Millisecond)
			return pair, nil
		})
	mockClientTwo.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(pair, nil)
	mockClientThree.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(nil, mockErr).AnyTimes()

	srv := testServer()
	srv.hashRing = mockRing
	srv.nodeClients["node-0"] = &models.StorageClient{StorageClient: mockClientOne}
	srv.nodeClients["node-1"] = &models.StorageClient{StorageClient: mockClientTwo}
	srv.nodeClients["node-2"] = &models.StorageClient{StorageClient: mockClientThree}

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
	mockRing.EXPECT().Get(uint32(100)).Return(models.Node{Idx: 1})

	mockClientOne := mocks.NewMockStorageClient(ctrl)
	mockClientTwo := mocks.NewMockStorageClient(ctrl)
	mockClientThree := mocks.NewMockStorageClient(ctrl)

	errMock := status.Error(codes.NotFound, "key not found in cache")

	mockClientOne.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 1, Key: key}).Return(nil, errMock).AnyTimes()
	mockClientTwo.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 1, Key: key}).Return(nil, errMock).AnyTimes()
	mockClientThree.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 1, Key: key}).Return(nil, errMock).AnyTimes()

	srv := testServer()
	srv.hashRing = mockRing
	srv.nodeClients["node-0"] = &models.StorageClient{StorageClient: mockClientOne}
	srv.nodeClients["node-1"] = &models.StorageClient{StorageClient: mockClientTwo}
	srv.nodeClients["node-2"] = &models.StorageClient{StorageClient: mockClientThree}

	req := &pb.Key{Key: key}
	retPair, err := srv.Fetch(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, retPair)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestGatewayFetchNoQuorumReached(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "test-get-key"
	value := []byte(`"test-value"`)
	pair := &pb.Pair{Key: key, Value: value}
	pairTwo := &pb.Pair{Key: key, Value: []byte(`"different-value"`)}

	mockRing := mocks.NewMockHashRing(ctrl)
	mockRing.EXPECT().Hash(key).Return(uint32(100))
	mockRing.EXPECT().Get(uint32(100)).Return(models.Node{Idx: 0})

	mockClientOne := mocks.NewMockStorageClient(ctrl)
	mockClientTwo := mocks.NewMockStorageClient(ctrl)
	mockClientThree := mocks.NewMockStorageClient(ctrl)

	mockClientOne.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(pairTwo, nil).AnyTimes()
	mockClientTwo.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(pair, nil).AnyTimes()
	mockClientThree.EXPECT().Get(gomock.Any(), &pb.GetRequest{Replica: 0, Key: key}).Return(nil, errors.New("err")).AnyTimes()

	srv := testServer()
	srv.hashRing = mockRing
	srv.nodeClients["node-0"] = &models.StorageClient{StorageClient: mockClientOne}
	srv.nodeClients["node-1"] = &models.StorageClient{StorageClient: mockClientTwo}
	srv.nodeClients["node-2"] = &models.StorageClient{StorageClient: mockClientThree}

	req := &pb.Key{Key: key}
	retPair, err := srv.Fetch(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, retPair)
	require.Equal(t, codes.Aborted, status.Code(err))
}
