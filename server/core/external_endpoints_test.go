package core

import (
	"context"
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
	value := "test-value"
	pair := &models.Pair{Key: key, Hash: hashkey, Value: value}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRing := mocks.NewMockHashRing(ctrl)
		mockRing.EXPECT().Hash(key).Return(hashkey)
		mockRing.EXPECT().GetN(hashkey, 2).Return(
			[]models.Node{
				models.Node{ID: "node-0", Idx: 0},
				models.Node{ID: "node-1", Idx: 1},
			}, nil)

		mockClientOne := mocks.NewMockClientInterface(ctrl)
		mockClientTwo := mocks.NewMockClientInterface(ctrl)

		mockClientOne.EXPECT().SetValue(gomock.Any(), pair, 0).Return(nil)
		mockClientTwo.EXPECT().SetValue(gomock.Any(), pair, 1).Return(nil)

		srv := newServer(nil, nil)
		srv.hashRing = mockRing
		srv.nodeClients["node-0"] = mockClientOne
		srv.nodeClients["node-1"] = mockClientTwo
		srv.minServers = 2

		req := &pb.Pair{Key: key, Value: []byte(`"test-value"`)}
		_, err := srv.Put(context.Background(), req)

		require.NoError(t, err)

	})

	t.Run("already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRing := mocks.NewMockHashRing(ctrl)
		mockRing.EXPECT().Hash(key).Return(hashkey)
		mockRing.EXPECT().GetN(hashkey, 2).Return(
			[]models.Node{
				models.Node{ID: "node-0", Idx: 0},
				models.Node{ID: "node-1", Idx: 1},
			}, nil)

		mockClientOne := mocks.NewMockClientInterface(ctrl)
		mockClientTwo := mocks.NewMockClientInterface(ctrl)

		errMock := status.Error(codes.AlreadyExists, "key in cache already")
		mockClientOne.EXPECT().SetValue(gomock.Any(), pair, 0).Return(nil).AnyTimes()
		mockClientOne.EXPECT().DelValue(gomock.Any(), pair.Key, 0).Return(nil).AnyTimes()
		mockClientTwo.EXPECT().SetValue(gomock.Any(), pair, 1).Return(errMock)

		srv := newServer(nil, nil)
		srv.hashRing = mockRing
		srv.nodeClients["node-0"] = mockClientOne
		srv.nodeClients["node-1"] = mockClientTwo
		srv.minServers = 2

		req := &pb.Pair{Key: key, Value: []byte(`"test-value"`)}
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
	value := "test-value"
	pair := &models.Pair{Key: key, Value: value}

	mockRing := mocks.NewMockHashRing(ctrl)
	mockRing.EXPECT().Hash(key).Return(uint32(100))
	mockRing.EXPECT().GetN(uint32(100), 2).Return(
		[]models.Node{
			models.Node{ID: "node-0", Idx: 0},
			models.Node{ID: "node-1", Idx: 1},
		}, nil)

	mockClientOne := mocks.NewMockClientInterface(ctrl)
	mockClientTwo := mocks.NewMockClientInterface(ctrl)

	mockClientOne.EXPECT().GetValue(gomock.Any(), key, 0).Return(pair, nil).AnyTimes()
	mockClientTwo.EXPECT().GetValue(gomock.Any(), key, 1).Return(pair, nil).AnyTimes()

	srv := newServer(nil, nil)
	srv.hashRing = mockRing
	srv.nodeClients["node-0"] = mockClientOne
	srv.nodeClients["node-1"] = mockClientTwo
	srv.minServers = 2

	req := &pb.Key{Key: key}
	retPair, err := srv.Retrieve(context.Background(), req)

	require.NoError(t, err)
	require.Equal(t, key, retPair.Key)
	require.Equal(t, []byte(`"test-value"`), retPair.Value)
}

func TestGatewayGetNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "test-get-key"

	mockRing := mocks.NewMockHashRing(ctrl)
	mockRing.EXPECT().Hash(key).Return(uint32(100))
	mockRing.EXPECT().GetN(uint32(100), 2).Return(
		[]models.Node{
			models.Node{ID: "node-0", Idx: 0},
			models.Node{ID: "node-1", Idx: 1},
		}, nil)

	mockClientOne := mocks.NewMockClientInterface(ctrl)
	mockClientTwo := mocks.NewMockClientInterface(ctrl)

	errMock := status.Error(codes.NotFound, "key not found in cache")

	mockClientOne.EXPECT().GetValue(gomock.Any(), key, 0).Return(nil, errMock).AnyTimes()
	mockClientTwo.EXPECT().GetValue(gomock.Any(), key, 1).Return(nil, errMock).AnyTimes()

	srv := newServer(nil, nil)
	srv.hashRing = mockRing
	srv.nodeClients["node-0"] = mockClientOne
	srv.nodeClients["node-1"] = mockClientTwo
	srv.minServers = 2

	req := &pb.Key{Key: key}
	retPair, err := srv.Retrieve(context.Background(), req)

	require.Error(t, err)
	require.Nil(t, retPair)
	require.Equal(t, codes.NotFound, status.Code(err))
}

func TestGatewayGetOneSuccessOneFailure(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientOne := mocks.NewMockClientInterface(ctrl)
	mockClientTwo := mocks.NewMockClientInterface(ctrl)
	mockClientThree := mocks.NewMockClientInterface(ctrl)

	key := "test-get-key"
	value := "test-value"
	pair := &models.Pair{Key: key, Value: value}
	errMock := status.Error(codes.Canceled, "request timed out")

	mockRing := mocks.NewMockHashRing(ctrl)
	mockRing.EXPECT().Hash(key).Return(uint32(100))
	mockRing.EXPECT().GetN(uint32(100), 2).Return(
		[]models.Node{
			models.Node{ID: "node-3", Idx: 0},
			models.Node{ID: "node-1", Idx: 1},
		}, nil)

	mockClientThree.EXPECT().GetValue(gomock.Any(), key, 0).Return(nil, errMock).AnyTimes()
	mockClientOne.EXPECT().GetValue(gomock.Any(), key, 1).
		DoAndReturn(func(...interface{}) (interface{}, error) {
			time.Sleep(50 * time.Millisecond)
			return pair, nil
		}).AnyTimes()

	srv := newServer(nil, nil)
	srv.hashRing = mockRing
	srv.nodeClients["node-1"] = mockClientOne
	srv.nodeClients["node-2"] = mockClientTwo
	srv.nodeClients["node-3"] = mockClientThree

	req := &pb.Key{Key: key}
	retPair, err := srv.Retrieve(context.Background(), req)

	require.NoError(t, err)
	require.Equal(t, key, retPair.Key)
	require.Equal(t, []byte(`"test-value"`), retPair.Value)
}
