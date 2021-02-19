package core

import (
	"testing"

	"github.com/michael-diggin/yass/common/models"

	"github.com/golang/mock/gomock"
	cmocks "github.com/michael-diggin/yass/common/mocks"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/stretchr/testify/require"
)

func TestRegisterNodeWithWatchTower(t *testing.T) {
	r := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMain := mocks.NewMockService(ctrl)
	mockBackup := mocks.NewMockService(ctrl)

	nodeStrings := []string{"yass-db-0:8080", "yass-db-1:8080"}

	mockRing := cmocks.NewMockHashRing(ctrl)
	factory := cmocks.NewMockClientFactory(ctrl)

	for _, node := range nodeStrings {
		cc := cmocks.NewMockStorageClient(ctrl)
		cc.EXPECT().AddNode(gomock.Any(), gomock.Any())
		newClient := &models.StorageClient{
			StorageClient: cc,
		}

		factory.EXPECT().NewProtoClient(gomock.Any(), node).Return(newClient, nil)
		mockRing.EXPECT().AddNode(node)
	}

	srv := newServer(factory, mockMain, mockBackup)
	srv.hashRing = mockRing

	mockWT := mocks.NewMockWatchTowerClient(ctrl)
	mockWT.EXPECT().RegisterNode(gomock.Any(), &pb.RegisterNodeRequest{Address: "yass-db-2:8080"}).
		Return(&pb.RegisterNodeResponse{ExistingNodes: nodeStrings}, nil)

	err := srv.RegisterNodeWithWatchTower(mockWT, "yass-db-2:8080")
	r.NoError(err)
}
