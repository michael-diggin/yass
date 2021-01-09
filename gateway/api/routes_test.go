package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/michael-diggin/yass/common/hashring"
	"github.com/michael-diggin/yass/common/mocks"
	"github.com/michael-diggin/yass/common/models"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setUpTestGateway(clients ...*mocks.MockClientInterface) *Gateway {
	g := NewGateway(2, 2, nil)
	for i, c := range clients {
		addr := fmt.Sprintf("server%d", i)
		g.Clients[addr] = c
		g.hashRing.AddNode(addr)
	}
	return g
}

func TestGatewaySet(t *testing.T) {
	key := "test"
	hashkey := hashring.Hash(key)
	value := "test-value"
	pair := &models.Pair{Key: key, Hash: hashkey, Value: value}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClientOne := mocks.NewMockClientInterface(ctrl)
		mockClientTwo := mocks.NewMockClientInterface(ctrl)

		mockClientOne.EXPECT().SetValue(gomock.Any(), pair, 0).Return(nil)
		mockClientTwo.EXPECT().SetValue(gomock.Any(), pair, 1).Return(nil)

		g := setUpTestGateway(mockClientOne, mockClientTwo)

		var payload = []byte(`{"key":"test", "value": "test-value"}`)
		req, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)

		require.Equal(t, rec.Code, http.StatusCreated)

		var resp string
		json.Unmarshal(rec.Body.Bytes(), &resp)
		require.Contains(t, resp, "successfully added")
	})

	t.Run("already exists", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClientOne := mocks.NewMockClientInterface(ctrl)
		mockClientTwo := mocks.NewMockClientInterface(ctrl)

		errMock := status.Error(codes.AlreadyExists, "key in cache already")
		mockClientOne.EXPECT().SetValue(gomock.Any(), pair, 0).Return(nil).AnyTimes()
		mockClientOne.EXPECT().DelValue(gomock.Any(), pair.Key, 0).Return(nil).AnyTimes()
		mockClientTwo.EXPECT().SetValue(gomock.Any(), pair, 1).Return(errMock)
		g := setUpTestGateway(mockClientOne, mockClientTwo)

		var payload = []byte(`{"key":"test", "value": "test-value"}`)
		req, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)

		require.Equal(t, rec.Code, http.StatusAlreadyReported)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		require.Contains(t, resp["error"], "already exists")
	})

	t.Run("bad data", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := mocks.NewMockClientInterface(ctrl)

		g := setUpTestGateway(mockClient, mockClient)

		payload := []byte{}
		req, _ := http.NewRequest("POST", "/set", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)

		require.Equal(t, rec.Code, http.StatusBadRequest)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		require.Contains(t, resp["error"], "could not be decoded")
	})
}

func TestGatewayGetSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientOne := mocks.NewMockClientInterface(ctrl)
	mockClientTwo := mocks.NewMockClientInterface(ctrl)
	g := NewGateway(2, 2, nil)

	key := "test-get-key"
	value := "test-value"
	pair := &models.Pair{Key: key, Value: value}

	mockClientOne.EXPECT().GetValue(gomock.Any(), key, 0).Return(pair, nil).AnyTimes()
	mockClientTwo.EXPECT().GetValue(gomock.Any(), key, 1).Return(pair, nil).AnyTimes()

	g.Clients["server0"] = mockClientOne
	g.Clients["server1"] = mockClientTwo
	g.hashRing.AddNode("server0")
	g.hashRing.AddNode("server1")

	req, _ := http.NewRequest("GET", "/get/"+key, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, req)

	require.Equal(t, rec.Code, http.StatusOK)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	require.Contains(t, resp["key"], key)
	require.Contains(t, resp["value"], value)

}
func TestGatewayGetNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientOne := mocks.NewMockClientInterface(ctrl)
	mockClientTwo := mocks.NewMockClientInterface(ctrl)

	key := "test-get-key"
	errMock := status.Error(codes.NotFound, "key not found in cache")

	mockClientOne.EXPECT().GetValue(gomock.Any(), key, 0).Return(nil, errMock).AnyTimes()
	mockClientTwo.EXPECT().GetValue(gomock.Any(), key, 1).Return(nil, errMock).AnyTimes()

	g := setUpTestGateway(mockClientOne, mockClientTwo)

	req, _ := http.NewRequest("GET", "/get/"+key, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, req)

	require.Equal(t, rec.Code, http.StatusNotFound)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	require.Contains(t, resp["error"], "not found")

}

func TestGatewayGetTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientOne := mocks.NewMockClientInterface(ctrl)
	mockClientTwo := mocks.NewMockClientInterface(ctrl)

	key := "test-get-key"
	errMock := status.Error(codes.Canceled, "request timed out")

	mockClientOne.EXPECT().GetValue(gomock.Any(), key, 0).Return(nil, errMock).AnyTimes()
	mockClientTwo.EXPECT().GetValue(gomock.Any(), key, 1).Return(nil, errMock).AnyTimes()

	g := setUpTestGateway(mockClientOne, mockClientTwo)

	req, _ := http.NewRequest("GET", "/get/"+"test-get-key", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, req)

	require.Equal(t, rec.Code, http.StatusRequestTimeout)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	require.Contains(t, resp["error"], "cancelled")

}

func TestGatewayGetOneSuccessOneFailure(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClientOne := mocks.NewMockClientInterface(ctrl)
	mockClientTwo := mocks.NewMockClientInterface(ctrl)
	mockClientThree := mocks.NewMockClientInterface(ctrl)

	key := "test-get-key"
	value := "test-value"
	pair := &models.Pair{Key: key, Value: value}
	errMock := status.Error(codes.Canceled, "request timed out")

	mockClientThree.EXPECT().GetValue(gomock.Any(), key, 0).Return(nil, errMock).AnyTimes()
	mockClientOne.EXPECT().GetValue(gomock.Any(), key, 0).
		DoAndReturn(func(...interface{}) (interface{}, error) {
			time.Sleep(50 * time.Millisecond)
			return pair, nil
		}).AnyTimes()

	g := setUpTestGateway(mockClientOne, mockClientTwo, mockClientThree)

	req, _ := http.NewRequest("GET", "/get/"+key, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	require.Contains(t, resp["key"], key)
	require.Contains(t, resp["value"], value)

}
