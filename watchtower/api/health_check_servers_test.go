package api

import (
	"testing"
)

func TestPingStorageServers(t *testing.T) {
	t.Skip("skipping health check tests for now")
	//ctx, cancel := context.WithCancel(context.Background())
	//ctrl := gomock.NewController(t)
	//defer ctrl.Finish()

	//mockClientOne := mocks.NewMockStorageClient(ctrl)
	//mockClientTwo := mocks.NewMockStorageClient(ctrl)

	//mockClientOne.EXPECT().Check(gomock.Any()).Return(true, nil)
	//mockClientTwo.EXPECT().Check(gomock.Any()).
	//	DoAndReturn(func(...interface{}) (bool, error) {
	//		cancel()
	//		return true, nil
	//	})

	//wt := setUpTestWatchTower(mockClientOne, mockClientTwo)

	//wt.PingStorageServers(ctx, 50*time.Millisecond)
}
