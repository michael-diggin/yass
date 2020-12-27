package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/michael-diggin/yass/gateway/mocks"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGatewaySet(t *testing.T) {
	mockClient := &mocks.MockGrpcClient{}
	g := NewGateway(1, &http.Server{})
	g.Clients[0] = mockClient

	t.Run("success", func(t *testing.T) {

		mockClient.SetFn = func(ctx context.Context, key string, value interface{}) error {
			return nil
		}
		mockClient.SetFollowerFn = func(ctx context.Context, key string, value interface{}) error {
			return nil
		}

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
		mockClient.SetFn = func(ctx context.Context, key string, value interface{}) error {
			return status.Error(codes.AlreadyExists, "key in cache already")
		}

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
	mockClientOne := &mocks.MockGrpcClient{}
	mockClientTwo := &mocks.MockGrpcClient{}
	g := NewGateway(2, &http.Server{})
	g.Clients[0] = mockClientOne
	g.Clients[1] = mockClientTwo

	mockClientOne.GetFn = func(ctx context.Context, key string) (interface{}, error) {
		if key == "test-get-key" {
			return "test-value", nil
		}
		return nil, errors.New("wrong key")
	}
	mockClientTwo.GetFollowerFn = func(ctx context.Context, key string) (interface{}, error) {
		if key == "test-get-key" {
			return "test-value", nil
		}
		return nil, errors.New("wrong key")
	}

	key := "test-get-key"
	req, _ := http.NewRequest("GET", "/get/"+key, nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, req)

	require.Equal(t, rec.Code, http.StatusOK)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	require.Contains(t, resp["key"], "test-get-key")
	require.Contains(t, resp["value"], "test-value")

}
func TestGatewayGetNotFound(t *testing.T) {
	mockClientOne := &mocks.MockGrpcClient{}
	mockClientTwo := &mocks.MockGrpcClient{}
	g := NewGateway(2, &http.Server{})
	g.Clients[0] = mockClientOne
	g.Clients[1] = mockClientTwo

	mockClientOne.GetFn = func(ctx context.Context, key string) (interface{}, error) {
		return nil, status.Error(codes.NotFound, "key not found in cache")
	}
	mockClientTwo.GetFollowerFn = func(ctx context.Context, key string) (interface{}, error) {
		return nil, status.Error(codes.NotFound, "key not found in cache")
	}

	req, _ := http.NewRequest("GET", "/get/"+"test-get-key", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, req)

	require.Equal(t, rec.Code, http.StatusNotFound)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	require.Contains(t, resp["error"], "not found")

}

func TestGatewayGetTimeout(t *testing.T) {
	mockClientOne := &mocks.MockGrpcClient{}
	mockClientTwo := &mocks.MockGrpcClient{}
	g := NewGateway(2, &http.Server{})
	g.Clients[0] = mockClientOne
	g.Clients[1] = mockClientTwo

	mockClientOne.GetFn = func(ctx context.Context, key string) (interface{}, error) {
		return nil, status.Error(codes.Canceled, "request timed out")
	}
	mockClientTwo.GetFollowerFn = func(ctx context.Context, key string) (interface{}, error) {
		return nil, status.Error(codes.Canceled, "request timed out")
	}

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
	mockClientOne := &mocks.MockGrpcClient{}
	mockClientTwo := &mocks.MockGrpcClient{}
	g := NewGateway(2, &http.Server{})
	g.Clients[0] = mockClientOne
	g.Clients[1] = mockClientTwo

	mockClientOne.GetFn = func(ctx context.Context, key string) (interface{}, error) {
		return nil, status.Error(codes.Canceled, "request timed out")
	}
	mockClientTwo.GetFollowerFn = func(ctx context.Context, key string) (interface{}, error) {
		time.Sleep(500 * time.Millisecond)
		return "test-value", nil
	}

	req, _ := http.NewRequest("GET", "/get/"+"test-get-key", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	g.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	require.Contains(t, resp["key"], "test-get-key")
	require.Contains(t, resp["value"], "test-value")

}

func TestGatewayDelete(t *testing.T) {
	mockClient := &mocks.MockGrpcClient{}
	g := NewGateway(1, &http.Server{})
	g.Clients[0] = mockClient

	t.Run("success", func(t *testing.T) {

		mockClient.DelFn = func(ctx context.Context, key string) error {
			return nil
		}
		mockClient.DelFollowerFn = func(ctx context.Context, key string) error {
			return nil
		}

		key := "test-key"
		req, _ := http.NewRequest("DELETE", "/delete/"+key, nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)

		require.Equal(t, rec.Code, http.StatusOK)

		var resp string
		json.Unmarshal(rec.Body.Bytes(), &resp)
		require.Contains(t, resp, "deleted")
	})

	t.Run("internal error", func(t *testing.T) {
		mockClient.DelFn = func(ctx context.Context, key string) error {
			return status.Error(codes.Internal, "internal server error")
		}

		req, _ := http.NewRequest("DELETE", "/delete/"+"test-key", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)

		require.Equal(t, rec.Code, http.StatusInternalServerError)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		require.Contains(t, resp["error"], "wrong")
	})

	t.Run("non grpc error", func(t *testing.T) {

		mockClient.DelFn = func(ctx context.Context, key string) error {
			return errors.New("network issues")
		}

		req, _ := http.NewRequest("DELETE", "/delete/"+"test-key", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)

		require.Equal(t, rec.Code, http.StatusInternalServerError)

		var resp map[string]interface{}
		json.Unmarshal(rec.Body.Bytes(), &resp)
		require.Contains(t, resp["error"], "network")
	})
}

func TestGatewayDeleteWithMultipleClients(t *testing.T) {
	mockClientOne := &mocks.MockGrpcClient{}
	mockClientTwo := &mocks.MockGrpcClient{}
	g := NewGateway(2, &http.Server{})
	g.Clients[0] = mockClientOne
	g.Clients[1] = mockClientTwo

	t.Run("success", func(t *testing.T) {

		mockClientTwo.DelFn = func(ctx context.Context, key string) error {
			return nil
		}
		mockClientOne.DelFollowerFn = func(ctx context.Context, key string) error {
			return nil
		}

		key := "test-key"
		req, _ := http.NewRequest("DELETE", "/delete/"+key, nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		g.ServeHTTP(rec, req)

		require.Equal(t, rec.Code, http.StatusOK)

		var resp string
		json.Unmarshal(rec.Body.Bytes(), &resp)
		require.Contains(t, resp, "deleted")
	})
}
