package api

import (
	"context"
	"sync"
	"testing"

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

		newClient := mocks.NewMockClientInterface(ctrl)

		factory := mocks.NewMockClientFactory(ctrl)
		factory.EXPECT().New(gomock.Any(), "127.0.0.1:8080").Return(newClient, nil)

		wt := NewWatchTower(2, 2, factory)

		mockHR := mocks.NewMockHashRing(ctrl)
		mockHR.EXPECT().AddNode("127.0.0.1:8080")
		wt.hashRing = mockHR

		req := &pb.RegisterNodeRequest{Address: "127.0.0.1:8080"}
		rec, err := wt.RegisterNode(context.Background(), req)

		require.NoError(t, err)
		require.Equal(t, rec.ExistingNodes, []string{})
	})

	t.Run("repopulate existing node", func(t *testing.T) {

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		wg := &sync.WaitGroup{}
		wg.Add(2)

		mockClientOne := mocks.NewMockClientInterface(ctrl)
		mockClientOne.EXPECT().AddNode(gomock.Any(), gomock.Any()).
			DoAndReturn(func(...interface{}) error {
				wg.Done()
				return nil
			})
		mockClientOne.EXPECT().BatchSend(gomock.Any(), 3, 7, "127.0.0.1:8080", uint32(100), uint32(150)).
			DoAndReturn(func(...interface{}) error {
				wg.Done()
				return nil
			})
		newClient := mocks.NewMockClientInterface(ctrl)
		factory := mocks.NewMockClientFactory(ctrl)
		factory.EXPECT().New(gomock.Any(), "127.0.0.1:8080").Return(newClient, nil)

		wt := NewWatchTower(2, 10, factory)
		wt.Clients["ip:port"] = mockClientOne
		wt.Clients["127.0.0.1:8080"] = mocks.NewMockClientInterface(ctrl)

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
