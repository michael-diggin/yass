package core

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	cmocks "github.com/michael-diggin/yass/common/mocks"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/stretchr/testify/require"
)

func TestServerAddNode(t *testing.T) {
	r := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMain := mocks.NewMockService(ctrl)
	mockBackup := mocks.NewMockService(ctrl)

	newClient := cmocks.NewMockClientInterface(ctrl)

	factory := cmocks.NewMockClientFactory(ctrl)
	factory.EXPECT().New(gomock.Any(), "localhost:8081").Return(newClient, nil)

	srv := newServer(factory, mockMain, mockBackup)

	mockRing := cmocks.NewMockHashRing(ctrl)
	mockRing.EXPECT().AddNode("localhost:8081")
	srv.hashRing = mockRing

	req := &pb.AddNodeRequest{Node: "localhost:8081"}

	ctx := context.Background()
	_, err := srv.AddNode(ctx, req)
	r.NoError(err)
}
