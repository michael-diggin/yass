package api

import (
	"context"
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

func TestRegisterNodeNoRebalancing(t *testing.T) {
	t.Run("add new node below limit", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tmpfile := setUpTestFile(t, "node-1")
		defer os.Remove(tmpfile.Name())

		newClient := mocks.NewMockClientInterface(ctrl)

		factory := mocks.NewMockClientFactory(ctrl)
		factory.EXPECT().New(gomock.Any(), "127.0.0.1:8080").Return(newClient, nil)

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

		mockClientOne := mocks.NewMockClientInterface(ctrl)

		mockClientOne.EXPECT().BatchSend(gomock.Any(), 3, 7, "127.0.0.1:8080", uint32(100), uint32(150)).
			DoAndReturn(func(...interface{}) error {
				wg.Done()
				return nil
			})
		mockClientTwo := mocks.NewMockClientInterface(ctrl)
		factory := mocks.NewMockClientFactory(ctrl)

		wt := NewWatchTower(2, 10, factory, "")
		wt.Clients["ip:port"] = mockClientOne
		wt.Clients["127.0.0.1:8080"] = mockClientTwo

		mockHR := mocks.NewMockHashRing(ctrl)
		instrs := []models.Instruction{
			models.Instruction{
				FromNode: "ip:port",
				FromIdx:  3,
				ToIdx:    7,
				LowHash:  uint32(100),
				HighHash: uint32(150),
			},
		}
		mockHR.EXPECT().RebalanceInstructions("127.0.0.1:8080").Return(instrs)
		wt.hashRing = mockHR

		req := &pb.RegisterNodeRequest{Address: "127.0.0.1:8080"}
		rec, err := wt.RegisterNode(context.Background(), req)

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

	mockClientOne := mocks.NewMockClientInterface(ctrl)
	mockClientTwo := mocks.NewMockClientInterface(ctrl)
	mockClientOne.EXPECT().AddNode(gomock.Any(), "127.0.0.1:8080").
		DoAndReturn(func(...interface{}) error {
			wg.Done()
			return nil
		})
	mockClientTwo.EXPECT().AddNode(gomock.Any(), "127.0.0.1:8080").
		DoAndReturn(func(...interface{}) error {
			wg.Done()
			return nil
		})
	mockClientOne.EXPECT().BatchSend(gomock.Any(), 3, 7, "127.0.0.1:8080", uint32(100), uint32(150)).Return(nil)
	mockClientOne.EXPECT().BatchDelete(gomock.Any(), 3, uint32(100), uint32(150)).
		DoAndReturn(func(...interface{}) error {
			wg.Done()
			return nil
		})
	mockClientTwo.EXPECT().BatchSend(gomock.Any(), 6, 1, "127.0.0.1:8080", uint32(900), uint32(1500)).Return(nil)
	mockClientTwo.EXPECT().BatchDelete(gomock.Any(), 6, uint32(900), uint32(1500)).
		DoAndReturn(func(...interface{}) error {
			wg.Done()
			return nil
		})
	newClient := mocks.NewMockClientInterface(ctrl)
	factory := mocks.NewMockClientFactory(ctrl)
	factory.EXPECT().New(gomock.Any(), "127.0.0.1:8080").Return(newClient, nil)

	wt := NewWatchTower(2, 10, factory, tmpfile.Name())
	wt.Clients["ip:port"] = mockClientOne
	wt.Clients["server:port"] = mockClientTwo

	mockHR := mocks.NewMockHashRing(ctrl)
	mockHR.EXPECT().AddNode("127.0.0.1:8080")
	instrs := []models.Instruction{
		models.Instruction{
			FromNode: "ip:port",
			FromIdx:  3,
			ToIdx:    7,
			LowHash:  uint32(100),
			HighHash: uint32(150),
		},
		models.Instruction{
			FromNode: "server:port",
			FromIdx:  6,
			ToIdx:    1,
			LowHash:  uint32(900),
			HighHash: uint32(1500),
		},
	}
	mockHR.EXPECT().RebalanceInstructions("127.0.0.1:8080").Return(instrs)
	wt.hashRing = mockHR

	req := &pb.RegisterNodeRequest{Address: "127.0.0.1:8080"}
	rec, err := wt.RegisterNode(context.Background(), req)

	require.NoError(t, err)
	require.Equal(t, rec.ExistingNodes, []string{"ip:port", "server:port"})

	f, _ := os.Open(tmpfile.Name())
	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	addr := strings.Split(string(b), "\n")
	require.Equal(t, []string{"ip:port", "server:port", "127.0.0.1:8080"}, addr)

	wg.Wait() // wait for rebalance goroutine to execute

}

func TestRebalanceData(t *testing.T) {
	instrs := []models.Instruction{
		models.Instruction{
			FromNode: "server0",
			FromIdx:  0,
			ToIdx:    1,
			LowHash:  uint32(100),
			HighHash: uint32(1000),
		},
		models.Instruction{
			FromNode: "server1",
			FromIdx:  1,
			ToIdx:    0,
			LowHash:  uint32(7000),
			HighHash: uint32(10),
		},
	}

	t.Run("repopulate failed node", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClientOne := mocks.NewMockClientInterface(ctrl)
		mockClientTwo := mocks.NewMockClientInterface(ctrl)
		mockClientThree := mocks.NewMockClientInterface(ctrl)

		mockClientOne.EXPECT().BatchSend(gomock.Any(), 0, 1, "server2", uint32(100), uint32(1000)).Return(nil)
		mockClientTwo.EXPECT().BatchSend(gomock.Any(), 1, 0, "server2", uint32(7000), uint32(10)).Return(nil)
		wt := setUpTestWatchTower(mockClientOne, mockClientTwo, mockClientThree)

		wt.rebalanceData("server2", instrs, false)
		time.Sleep(100 * time.Millisecond) // want to check the gorountines execute

	})
	t.Run("rebalance data to new node", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClientOne := mocks.NewMockClientInterface(ctrl)
		mockClientTwo := mocks.NewMockClientInterface(ctrl)
		mockClientThree := mocks.NewMockClientInterface(ctrl)

		mockClientOne.EXPECT().BatchSend(gomock.Any(), 0, 1, "server2", uint32(100), uint32(1000)).Return(nil)
		mockClientTwo.EXPECT().BatchSend(gomock.Any(), 1, 0, "server2", uint32(7000), uint32(10)).Return(nil)

		mockClientOne.EXPECT().BatchDelete(gomock.Any(), 0, uint32(100), uint32(1000)).Return(nil)
		mockClientTwo.EXPECT().BatchDelete(gomock.Any(), 1, uint32(7000), uint32(10)).Return(nil)

		wt := setUpTestWatchTower(mockClientOne, mockClientTwo, mockClientThree)

		wt.rebalanceData("server2", instrs, true)
		time.Sleep(100 * time.Millisecond) // want to check the gorountines execute
	})
}
