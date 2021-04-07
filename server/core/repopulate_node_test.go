package core

import (
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	cmocks "github.com/michael-diggin/yass/common/mocks"
	"github.com/michael-diggin/yass/common/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/michael-diggin/yass/server/model"
	"github.com/stretchr/testify/require"
)

func TestRepopulateNode(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMain := mocks.NewMockService(ctrl)
	mockBackup := mocks.NewMockService(ctrl)

	respMain := make(chan error, 1)
	respMain <- nil
	close(respMain)
	respBackup := make(chan error, 1)
	respBackup <- nil
	close(respBackup)

	wg := sync.WaitGroup{}
	wg.Add(2)

	mockMain.EXPECT().BatchSet(gomock.Any()).DoAndReturn(
		func(d map[string]model.Data) <-chan error {
			if len(d) != 2 {
				t.FailNow()
			}
			wg.Done()
			return respMain
		})
	mockBackup.EXPECT().BatchSet(gomock.Any()).DoAndReturn(
		func(d map[string]model.Data) <-chan error {
			if len(d) != 2 {
				t.FailNow()
			}
			wg.Done()
			return respBackup
		})

	pairData := []*pb.Pair{
		{Key: "key-1", Hash: uint32(100), Value: []byte(`"value-1"`)},
		{Key: "key-2", Hash: uint32(101), Value: []byte(`2`)},
	}

	cc := cmocks.NewMockStorageClient(ctrl)
	resp := &pb.BatchGetResponse{Data: pairData}
	cc.EXPECT().BatchGet(gomock.Any(), &pb.BatchGetRequest{Replica: int32(0)}).DoAndReturn(
		func(_ ...interface{}) (*pb.BatchGetResponse, error) {
			return resp, nil
		})
	cc.EXPECT().BatchGet(gomock.Any(), &pb.BatchGetRequest{Replica: int32(1)}).DoAndReturn(
		func(_ ...interface{}) (*pb.BatchGetResponse, error) {
			return resp, nil
		})

	newClient := &models.StorageClient{
		StorageClient: cc,
	}

	factory := cmocks.NewMockClientFactory(ctrl)

	srv := newServer(factory, "yass-0", "yass-0", mockMain, mockBackup)
	srv.nodeClients["yass-0.yassdb:8080"] = newClient

	podName := "yass-0.yassdb:8080"

	err := srv.RepopulateFromNodes(podName)
	require.NoError(t, err)
	wg.Wait()

}
