package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/michael-diggin/yass/common/mocks"
	"github.com/michael-diggin/yass/common/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/stretchr/testify/require"
)

func setUpTestWatchTower(clients ...*mocks.MockStorageClient) *WatchTower {
	wt := NewWatchTower(2, 2, nil, "")
	for i, c := range clients {
		sc := &models.StorageClient{StorageClient: c}
		addr := fmt.Sprintf("server%d", i)
		wt.Clients[addr] = sc
		wt.hashRing.AddNode(addr)
	}
	return wt
}

func TestRegisterNodeNoRebalancing(t *testing.T) {
	t.Run("add new node below limit", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tmpfile := setUpTestFile(t, "node-1")
		defer os.Remove(tmpfile.Name())

		newClient := &models.StorageClient{
			StorageClient: mocks.NewMockStorageClient(ctrl),
		}

		factory := mocks.NewMockClientFactory(ctrl)
		factory.EXPECT().NewProtoClient(gomock.Any(), "127.0.0.1:8080").Return(newClient, nil)

		wt := NewWatchTower(2, 2, factory, tmpfile.Name())

		mockHR := mocks.NewMockHashRing(ctrl)
		mockHR.EXPECT().AddNode("127.0.0.1:8080")
		wt.hashRing = mockHR

		req := &pb.RegisterNodeRequest{Address: "127.0.0.1:8080"}
		rec, err := wt.RegisterNode(context.Background(), req)

		require.NoError(t, err)
		require.Equal(t, rec.ExistingNodes, []string{})

		f, _ := os.Open(tmpfile.Name())
		b, err := ioutil.ReadAll(f)
		require.NoError(t, err)
		addr := strings.Split(string(b), "\n")
		require.Equal(t, []string{"node-1", "127.0.0.1:8080"}, addr)
	})

	t.Run("repopulate existing node", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		wg := &sync.WaitGroup{}
		wg.Add(1)

		mockClientOne := mocks.NewMockStorageClient(ctrl)

		req := &pb.BatchSendRequest{Replica: int32(3), Address: "127.0.0.1:8080",
			ToReplica: int32(7), Low: uint32(100), High: uint32(150)}
		mockClientOne.EXPECT().BatchSend(gomock.Any(), req).
			DoAndReturn(func(...interface{}) (*pb.Null, error) {
				wg.Done()
				return nil, nil
			})
		mockClientTwo := mocks.NewMockStorageClient(ctrl)
		factory := mocks.NewMockClientFactory(ctrl)

		wt := NewWatchTower(2, 10, factory, "")
		wt.Clients["ip:port"] = &models.StorageClient{StorageClient: mockClientOne}
		wt.Clients["127.0.0.1:8080"] = &models.StorageClient{StorageClient: mockClientTwo}

		mockHR := mocks.NewMockHashRing(ctrl)
		instrs := []models.Instruction{
			{
				FromNode: "ip:port",
				FromIdx:  3,
				ToIdx:    7,
				LowHash:  uint32(100),
				HighHash: uint32(150),
			},
		}
		mockHR.EXPECT().RebalanceInstructions("127.0.0.1:8080").Return(instrs)
		wt.hashRing = mockHR

		regReq := &pb.RegisterNodeRequest{Address: "127.0.0.1:8080"}
		rec, err := wt.RegisterNode(context.Background(), regReq)

		require.NoError(t, err)
		require.Equal(t, rec.ExistingNodes, []string{"ip:port"})

		wg.Wait() // wait for rebalance goroutine to execute
	})
}

func TestRebalanceDataToNewNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tmpfile := setUpTestFile(t, "ip:port\nserver:port")
	defer os.Remove(tmpfile.Name())

	wg := sync.WaitGroup{}
	wg.Add(4)

	mockClientOne := mocks.NewMockStorageClient(ctrl)
	mockClientTwo := mocks.NewMockStorageClient(ctrl)
	req := &pb.AddNodeRequest{Node: "127.0.0.1:8080"}
	mockClientOne.EXPECT().AddNode(gomock.Any(), req).
		DoAndReturn(func(...interface{}) (*pb.Null, error) {
			wg.Done()
			return nil, nil
		})
	mockClientTwo.EXPECT().AddNode(gomock.Any(), req).
		DoAndReturn(func(...interface{}) (*pb.Null, error) {
			wg.Done()
			return nil, nil
		})
	bsReq := &pb.BatchSendRequest{Replica: int32(3), Address: "127.0.0.1:8080",
		ToReplica: int32(7), Low: uint32(100), High: uint32(150)}
	bdReq := &pb.BatchDeleteRequest{Replica: int32(3), Low: uint32(100), High: uint32(150)}
	mockClientOne.EXPECT().BatchSend(gomock.Any(), bsReq).Return(nil, nil)
	mockClientOne.EXPECT().BatchDelete(gomock.Any(), bdReq).
		DoAndReturn(func(...interface{}) (*pb.Null, error) {
			wg.Done()
			return nil, nil
		})
	bsReq2 := &pb.BatchSendRequest{Replica: int32(6), Address: "127.0.0.1:8080",
		ToReplica: int32(1), Low: uint32(900), High: uint32(1500)}
	bdReq2 := &pb.BatchDeleteRequest{Replica: int32(6), Low: uint32(900), High: uint32(1500)}
	mockClientTwo.EXPECT().BatchSend(gomock.Any(), bsReq2).Return(nil, nil)
	mockClientTwo.EXPECT().BatchDelete(gomock.Any(), bdReq2).
		DoAndReturn(func(...interface{}) (*pb.Null, error) {
			wg.Done()
			return nil, nil
		})
	newClient := &models.StorageClient{
		StorageClient: mocks.NewMockStorageClient(ctrl),
	}
	factory := mocks.NewMockClientFactory(ctrl)
	factory.EXPECT().NewProtoClient(gomock.Any(), "127.0.0.1:8080").Return(newClient, nil)

	wt := NewWatchTower(2, 10, factory, tmpfile.Name())
	wt.Clients["ip:port"] = &models.StorageClient{StorageClient: mockClientOne}
	wt.Clients["server:port"] = &models.StorageClient{StorageClient: mockClientTwo}

	mockHR := mocks.NewMockHashRing(ctrl)
	mockHR.EXPECT().AddNode("127.0.0.1:8080")
	instrs := []models.Instruction{
		{
			FromNode: "ip:port",
			FromIdx:  3,
			ToIdx:    7,
			LowHash:  uint32(100),
			HighHash: uint32(150),
		},
		{
			FromNode: "server:port",
			FromIdx:  6,
			ToIdx:    1,
			LowHash:  uint32(900),
			HighHash: uint32(1500),
		},
	}
	mockHR.EXPECT().RebalanceInstructions("127.0.0.1:8080").Return(instrs)
	wt.hashRing = mockHR

	addReq := &pb.RegisterNodeRequest{Address: "127.0.0.1:8080"}
	rec, err := wt.RegisterNode(context.Background(), addReq)

	require.NoError(t, err)
	require.Contains(t, rec.ExistingNodes, "ip:port")
	require.Contains(t, rec.ExistingNodes, "server:port")

	f, _ := os.Open(tmpfile.Name())
	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	addr := strings.Split(string(b), "\n")
	require.Equal(t, []string{"ip:port", "server:port", "127.0.0.1:8080"}, addr)

	wg.Wait() // wait for rebalance goroutine to execute

}

func TestRebalanceData(t *testing.T) {
	instrs := []models.Instruction{
		{
			FromNode: "server0",
			FromIdx:  0,
			ToIdx:    1,
			LowHash:  uint32(100),
			HighHash: uint32(1000),
		},
		{
			FromNode: "server1",
			FromIdx:  1,
			ToIdx:    0,
			LowHash:  uint32(7000),
			HighHash: uint32(10),
		},
	}

	bsReq1 := &pb.BatchSendRequest{Replica: int32(0), Address: "server2",
		ToReplica: int32(1), Low: uint32(100), High: uint32(1000)}
	bdReq1 := &pb.BatchDeleteRequest{Replica: int32(0), Low: uint32(100), High: uint32(1000)}
	bsReq2 := &pb.BatchSendRequest{Replica: int32(1), Address: "server2",
		ToReplica: int32(0), Low: uint32(7000), High: uint32(10)}
	bdReq2 := &pb.BatchDeleteRequest{Replica: int32(1), Low: uint32(7000), High: uint32(10)}

	t.Run("repopulate failed node", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClientOne := mocks.NewMockStorageClient(ctrl)
		mockClientTwo := mocks.NewMockStorageClient(ctrl)
		mockClientThree := mocks.NewMockStorageClient(ctrl)

		mockClientOne.EXPECT().BatchSend(gomock.Any(), bsReq1).Return(nil, nil)
		mockClientTwo.EXPECT().BatchSend(gomock.Any(), bsReq2).Return(nil, nil)
		wt := setUpTestWatchTower(mockClientOne, mockClientTwo, mockClientThree)

		wt.rebalanceData("server2", instrs, false)
		time.Sleep(100 * time.Millisecond) // want to check the gorountines execute

	})
	t.Run("rebalance data to new node", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClientOne := mocks.NewMockStorageClient(ctrl)
		mockClientTwo := mocks.NewMockStorageClient(ctrl)
		mockClientThree := mocks.NewMockStorageClient(ctrl)

		mockClientOne.EXPECT().BatchSend(gomock.Any(), bsReq1).Return(nil, nil)
		mockClientTwo.EXPECT().BatchSend(gomock.Any(), bsReq2).Return(nil, nil)

		mockClientOne.EXPECT().BatchDelete(gomock.Any(), bdReq1).Return(nil, nil)
		mockClientTwo.EXPECT().BatchDelete(gomock.Any(), bdReq2).Return(nil, nil)

		wt := setUpTestWatchTower(mockClientOne, mockClientTwo, mockClientThree)

		wt.rebalanceData("server2", instrs, true)
		time.Sleep(100 * time.Millisecond) // want to check the gorountines execute
	})
}
