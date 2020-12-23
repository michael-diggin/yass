package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"testing"
	"time"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/michael-diggin/yass/server/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServerPing(t *testing.T) {
	tt := []struct {
		name    string
		errCode codes.Code
		status  pb.PingResponse_ServingStatus
	}{
		{"serving", codes.OK, pb.PingResponse_SERVING},
		{"not serving", codes.Unavailable, pb.PingResponse_NOT_SERVING},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cache := &mocks.TestCache{
				PingFn: func() error {
					if tc.name == "serving" {
						return nil
					}
					return errors.New("Cache not reachable")
				},
			}
			srv := server{cache}

			resp, err := srv.Ping(context.Background(), &pb.Null{})
			if grpc.Code(err) != tc.errCode {
				t.Fatalf("Expected code %v, got %v", tc.errCode, grpc.Code(err))
			}
			if resp.Status != tc.status {
				t.Fatalf("Incorrect status returned %v", resp.Status)
			}

		})
	}
}

func TestSettoCache(t *testing.T) {
	logrus.SetOutput(ioutil.Discard) // Discard log output for test
	tt := []struct {
		name    string
		key     string
		value   []byte
		errCode codes.Code
		timeout time.Duration
	}{
		{"valid case", "newKey", []byte(`"newValue"`), codes.OK, 100 * time.Millisecond},
		{"already set key", "testKey", []byte(`"testVal"`), codes.AlreadyExists, 100 * time.Millisecond},
		{"empty key", "", []byte(`"emptyVal"`), codes.InvalidArgument, 100 * time.Millisecond},
		{"empty value", "key", []byte{}, codes.InvalidArgument, 100 * time.Millisecond},
		{"timeout", "newKey", []byte(`"newValue"`), codes.Canceled, 0 * time.Millisecond},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cache := &mocks.TestCache{
				SetFn: func(key string, value interface{}) *model.CacheResponse {
					return &model.CacheResponse{
						Key:   tc.key,
						Value: tc.value,
						Err:   status.Error(tc.errCode, "")}
				},
			}
			srv := server{cache}
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
		value   []byte
		errCode codes.Code
		timeout time.Duration
	}{
		{"valid case", "testKey", []byte(`"testValue"`), codes.OK, 100 * time.Millisecond},
		{"key not found", "newKey", []byte{}, codes.NotFound, 100 * time.Millisecond},
		{"empty key", "", []byte{}, codes.InvalidArgument, 100 * time.Millisecond},
		{"timeout", "newKey", []byte(`"newValue"`), codes.Canceled, 0 * time.Millisecond},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cache := &mocks.TestCache{
				GetFn: func(key string) *model.CacheResponse {
					return &model.CacheResponse{
						Key:   tc.key,
						Value: tc.value,
						Err:   status.Error(tc.errCode, "")}
				},
			}
			srv := server{cache}
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
				expectedBytes, _ := json.Marshal(tc.value)
				if !bytes.Equal(expectedBytes, res.Value) {
					t.Fatalf("Expected %s, got %s", expectedBytes, res.Value)
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
		{"empty key", "", codes.InvalidArgument, 100 * time.Millisecond},
		{"timeout", "Key", codes.Canceled, 0 * time.Millisecond},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			cache := &mocks.TestCache{
				DelFn: func(key string) *model.CacheResponse {
					return &model.CacheResponse{
						Key:   tc.key,
						Value: "",
						Err:   status.Error(tc.errCode, "")}
				},
			}
			srv := server{cache}
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
