package api

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/michael-diggin/yass/common/mocks"
)

func TestPingStorageServers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientOne := mocks.NewMockClientInterface(ctrl)
	mockClientTwo := mocks.NewMockClientInterface(ctrl)

	mockClientOne.EXPECT().Check(gomock.Any()).Return(true, nil)
	mockClientTwo.EXPECT().Check(gomock.Any()).
		DoAndReturn(func(...interface{}) (bool, error) {
			cancel()
			return true, nil
		})

	wt := setUpTestWatchTower(mockClientOne, mockClientTwo)

	wt.PingStorageServers(ctx, 50*time.Millisecond)
}
