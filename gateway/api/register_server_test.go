package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/michael-diggin/yass/common/mocks"
	"github.com/michael-diggin/yass/gateway/models"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRegisterServerNoRebalancing(t *testing.T) {
	t.Skip("client factory not implemented yet")
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	g := NewGateway(2, 2, &http.Server{})

	var payload = []byte(`{"ip":"127.0.0.1", "port": "8080"}`)
	req, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, req)

	require.Equal(t, rec.Code, http.StatusCreated)

}

func TestRebalanceData(t *testing.T) {
	instrs := []models.Instruction{
		models.Instruction{
			FromNode: "server1",
			FromIdx:  0,
			ToIdx:    1,
			LowHash:  uint32(100),
			HighHash: uint32(1000),
		},
		models.Instruction{
			FromNode: "server2",
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
		g := NewGateway(2, 1, &http.Server{})

		mockClientOne.EXPECT().BatchSend(gomock.Any(), 0, 1, "server3", uint32(100), uint32(1000)).Return(nil)
		mockClientTwo.EXPECT().BatchSend(gomock.Any(), 1, 0, "server3", uint32(7000), uint32(10)).Return(nil)

		g.Clients["server1"] = mockClientOne
		g.Clients["server2"] = mockClientTwo

		g.rebalanceData("server3", instrs, false)
		time.Sleep(500 * time.Millisecond) // want to check the gorountines execute

	})
	t.Run("rebalance data to new node", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClientOne := mocks.NewMockClientInterface(ctrl)
		mockClientTwo := mocks.NewMockClientInterface(ctrl)
		g := NewGateway(2, 1, &http.Server{})

		mockClientOne.EXPECT().BatchSend(gomock.Any(), 0, 1, "server3", uint32(100), uint32(1000)).Return(nil)
		mockClientTwo.EXPECT().BatchSend(gomock.Any(), 1, 0, "server3", uint32(7000), uint32(10)).Return(nil)

		mockClientOne.EXPECT().BatchDelete(gomock.Any(), 0, uint32(100), uint32(1000)).Return(nil)
		mockClientTwo.EXPECT().BatchDelete(gomock.Any(), 1, uint32(7000), uint32(10)).Return(nil)

		g.Clients["server1"] = mockClientOne
		g.Clients["server2"] = mockClientTwo

		g.rebalanceData("server3", instrs, true)
		time.Sleep(500 * time.Millisecond) // want to check the gorountines execute

	})
}
