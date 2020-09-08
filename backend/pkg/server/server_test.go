package server

import (
	"context"
	"io/ioutil"
	"testing"
	"time"

	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend"
	"github.com/michael-diggin/yass/backend/mocks"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSettoCache(t *testing.T) {
	logrus.SetOutput(ioutil.Discard) // Discard log output for test
	tt := []struct {
		name    string
		key     string
		value   string
		errCode codes.Code
		timeout time.Duration
	}{
		{"valid case", "newKey", "newValue", codes.OK, 100 * time.Millisecond},
		{"already set key", "testKey", "testVal", codes.AlreadyExists, 100 * time.Millisecond},
		{"empty key", "", "emptyVal", codes.InvalidArgument, 100 * time.Millisecond},
		{"empty value", "key", "", codes.InvalidArgument, 100 * time.Millisecond},
		{"timeout", "newKey", "newValue", codes.Canceled, 0 * time.Millisecond},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cache := mocks.TestCache{
				SetFn: func(ctx context.Context, key, value string) *backend.CacheResponse {
					return &backend.CacheResponse{
						Key:   tc.key,
						Value: tc.value,
						Err:   status.Error(tc.errCode, "")}
				},
			}
			srv := New(cache)

			testKV := &pb.Pair{Key: tc.key, Value: tc.value}

			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			res, err := srv.Set(ctx, testKV)
			cancel()
			if e, ok := status.FromError(err); ok {
				if e.Code() != tc.errCode {
					t.Fatalf("Expected error '%v', got: '%v'", tc.errCode, err)
				}
			}
			if res != nil {
				if res.Key != testKV.Key {
					t.Fatalf("Expected %s, got %s", testKV.Key, res.Key)
				}
			}

		})
	}
}

func TestGetFromCache(t *testing.T) {
	logrus.SetOutput(ioutil.Discard) // Discard log output for test
	tt := []struct {
		name    string
		key     string
		value   string
		errCode codes.Code
		timeout time.Duration
	}{
		{"valid case", "testKey", "testValue", codes.OK, 100 * time.Millisecond},
		{"key not found", "newKey", "", codes.NotFound, 100 * time.Millisecond},
		{"empty key", "", "", codes.InvalidArgument, 100 * time.Millisecond},
		{"timeout", "newKey", "newValue", codes.Canceled, 0 * time.Millisecond},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cache := mocks.TestCache{
				GetFn: func(ctx context.Context, key string) *backend.CacheResponse {
					return &backend.CacheResponse{
						Key:   tc.key,
						Value: tc.value,
						Err:   status.Error(tc.errCode, "")}
				},
			}
			srv := New(cache)

			testK := &pb.Key{Key: tc.key}
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			res, err := srv.Get(ctx, testK)
			cancel()
			e, ok := status.FromError(err)
			if !ok {
				t.Errorf("Could not get code from error: %v", err)
			}
			if e.Code() != tc.errCode {
				t.Fatalf("Expected error '%v', got: '%v'", tc.errCode, err)
			}

			if res != nil {
				if res.Value != tc.value {
					t.Fatalf("Expected %s, got %s", tc.value, res.Key)
				}
			}

		})
	}
}

func TestDeleteKeyValue(t *testing.T) {
	logrus.SetOutput(ioutil.Discard) // Discard log output for test
	tt := []struct {
		name    string
		key     string
		errCode codes.Code
		timeout time.Duration
	}{
		{"valid case", "testKey", codes.OK, 100 * time.Millisecond},
		{"key not found", "newKey", codes.NotFound, 100 * time.Millisecond},
		{"empty key", "", codes.InvalidArgument, 100 * time.Millisecond},
		{"timeout", "Key", codes.Canceled, 0 * time.Millisecond},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cache := mocks.TestCache{
				DelFn: func(ctx context.Context, key string) *backend.CacheResponse {
					return &backend.CacheResponse{
						Key:   tc.key,
						Value: "",
						Err:   status.Error(tc.errCode, "")}
				},
			}
			srv := New(cache)
			testKV := &pb.Key{Key: tc.key}
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			_, err := srv.Delete(ctx, testKV)
			cancel()
			e, ok := status.FromError(err)
			if !ok {
				t.Errorf("Could not get code from error: %v", err)
			}
			if e.Code() != tc.errCode {
				t.Fatalf("Expected error '%v', got: '%v'", tc.errCode, err)
			}

		})
	}
}
