package api

import (
	"context"
	"net/http"
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
	g := NewGateway(2, 1, &http.Server{}, nil)

	mockClientOne.EXPECT().Check(gomock.Any()).Return(true, nil)
	mockClientTwo.EXPECT().Check(gomock.Any()).
		DoAndReturn(func(...interface{}) (bool, error) {
			cancel()
			return true, nil
		})

	g.Clients["0"] = mockClientOne
	g.Clients["1"] = mockClientTwo

	g.PingStorageServers(ctx, 50*time.Millisecond)
}
