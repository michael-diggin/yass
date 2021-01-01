package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/michael-diggin/yass/gateway/mocks"
	"github.com/michael-diggin/yass/models"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGatewaySet(t *testing.T) {
	key := "test"
	value := "test-value"
	pair := &models.Pair{Key: key, Value: value}

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClientOne := mocks.NewMockGrpcClient(ctrl)
		mockClientTwo := mocks.NewMockGrpcClient(ctrl)
		g := NewGateway(2, &http.Server{})

		mockClientOne.EXPECT().SetValue(gomock.Any(), pair, models.MainReplica).Return(nil)
		mockClientTwo.EXPECT().SetValue(gomock.Any(), pair, models.MainReplica).Return(nil)
		g.Clients["0"] = mockClientOne
		g.hashRing.AddNode("0")
		g.Clients["1"] = mockClientTwo
		g.hashRing.AddNode("1")

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

		mockClientOne := mocks.NewMockGrpcClient(ctrl)
		mockClientTwo := mocks.NewMockGrpcClient(ctrl)
		g := NewGateway(2, &http.Server{})

		errMock := status.Error(codes.AlreadyExists, "key in cache already")
		mockClientOne.EXPECT().SetValue(gomock.Any(), pair, models.MainReplica).Return(nil).AnyTimes()
		mockClientOne.EXPECT().DelValue(gomock.Any(), key, models.MainReplica).Return(nil).AnyTimes()
		mockClientTwo.EXPECT().SetValue(gomock.Any(), pair, models.MainReplica).Return(errMock)
		g.Clients["0"] = mockClientOne
		g.hashRing.AddNode("0")
		g.Clients["1"] = mockClientTwo
		g.hashRing.AddNode("1")

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

		mockClient := mocks.NewMockGrpcClient(ctrl)
		g := NewGateway(1, &http.Server{})
		g.replicas = 1

		g.Clients["0"] = mockClient
		g.hashRing.AddNode("0")

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

	mockClientOne := mocks.NewMockGrpcClient(ctrl)
	mockClientTwo := mocks.NewMockGrpcClient(ctrl)
	g := NewGateway(2, &http.Server{})

	key := "test-get-key"
	value := "test-value"
	pair := &models.Pair{Key: key, Value: value}

	mockClientOne.EXPECT().GetValue(gomock.Any(), key, models.MainReplica).Return(pair, nil).AnyTimes()
	mockClientTwo.EXPECT().GetValue(gomock.Any(), key, models.MainReplica).Return(pair, nil).AnyTimes()

	g.Clients["0"] = mockClientOne
	g.Clients["1"] = mockClientTwo
	g.hashRing.AddNode("0")
	g.hashRing.AddNode("1")

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

	mockClientOne := mocks.NewMockGrpcClient(ctrl)
	mockClientTwo := mocks.NewMockGrpcClient(ctrl)
	g := NewGateway(2, &http.Server{})

	key := "test-get-key"
	errMock := status.Error(codes.NotFound, "key not found in cache")

	mockClientOne.EXPECT().GetValue(gomock.Any(), key, models.MainReplica).Return(nil, errMock).AnyTimes()
	mockClientTwo.EXPECT().GetValue(gomock.Any(), key, models.MainReplica).Return(nil, errMock).AnyTimes()

	g.Clients["0"] = mockClientOne
	g.Clients["1"] = mockClientTwo
	g.hashRing.AddNode("0")
	g.hashRing.AddNode("1")

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

	mockClientOne := mocks.NewMockGrpcClient(ctrl)
	mockClientTwo := mocks.NewMockGrpcClient(ctrl)
	g := NewGateway(2, &http.Server{})

	key := "test-get-key"
	errMock := status.Error(codes.Canceled, "request timed out")

	mockClientOne.EXPECT().GetValue(gomock.Any(), key, models.MainReplica).Return(nil, errMock).AnyTimes()
	mockClientTwo.EXPECT().GetValue(gomock.Any(), key, models.MainReplica).Return(nil, errMock).AnyTimes()

	g.Clients["0"] = mockClientOne
	g.Clients["1"] = mockClientTwo
	g.hashRing.AddNode("0")
	g.hashRing.AddNode("1")

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

	mockClientOne := mocks.NewMockGrpcClient(ctrl)
	mockClientTwo := mocks.NewMockGrpcClient(ctrl)
	mockClientThree := mocks.NewMockGrpcClient(ctrl)
	g := NewGateway(3, &http.Server{})

	key := "test-get-key"
	value := "test-value"
	pair := &models.Pair{Key: key, Value: value}
	errMock := status.Error(codes.Canceled, "request timed out")

	mockClientOne.EXPECT().GetValue(gomock.Any(), key, models.MainReplica).Return(nil, errMock).AnyTimes()
	mockClientTwo.EXPECT().GetValue(gomock.Any(), key, models.MainReplica).
		DoAndReturn(func(...interface{}) (interface{}, error) {
			time.Sleep(500 * time.Millisecond)
			return pair, nil
		}).AnyTimes()
	mockClientThree.EXPECT().GetValue(gomock.Any(), key, models.MainReplica).Return(pair, nil).AnyTimes()

	g.Clients["0"] = mockClientOne
	g.Clients["1"] = mockClientTwo
	g.Clients["2"] = mockClientThree
	g.hashRing.AddNode("0")
	g.hashRing.AddNode("1")
	g.hashRing.AddNode("2")

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
