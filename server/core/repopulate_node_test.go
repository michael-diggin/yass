package core

import (
	"context"
	"math"
	"testing"

	"github.com/golang/mock/gomock"
	cmocks "github.com/michael-diggin/yass/common/mocks"
	"github.com/michael-diggin/yass/common/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/michael-diggin/yass/server/model"
)

func TestRepopulateNode(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMain := mocks.NewMockService(ctrl)
	mockBackup := mocks.NewMockService(ctrl)

	respMain := make(chan map[string]model.Data, 1)
	//respBackup := make(chan map[string]model.Data, 1)
	data := map[string]model.Data{
		"key-1": {Value: "value-1", Hash: uint32(100)},
		"key-2": {Value: 2, Hash: uint32(101)},
	}
	respMain <- data
	close(respMain)
	//	respBackup <- data
	//	close(respBackup)

	mockMain.EXPECT().BatchGet(uint32(0), uint32(math.MaxUint32)).Return(respMain)
	//	mockBackup.EXPECT().BatchGet(uint32(0), uint32(math.MaxUint32)).Return(respBackup)

	cc := cmocks.NewMockStorageClient(ctrl)
	cc.EXPECT().BatchSet(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ ...interface{}) (*pb.Null, error) {
			cancel()
			return &pb.Null{}, nil
		})

	newClient := &models.StorageClient{
		StorageClient: cc,
	}

	factory := cmocks.NewMockClientFactory(ctrl)

	srv := newServer(factory, mockMain, mockBackup)
	srv.nodeClients["yass-0.yassdb:8080"] = newClient

	podName := "yass-1.yassdb:8080"

	srv.repopulateChan <- "yass-0.yassdb:8080"
	srv.RepopulateNodes(ctx, podName)

}
