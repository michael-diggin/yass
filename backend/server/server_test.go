package server

import (
	"context"
	"io/ioutil"
	"testing"

	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func setUpCacheServer() Server {
	testCache := storage.NewCacheService()
	testCache["testKey"] = "testValue"
	srv := New(testCache)
	return srv
}

func TestAddtoCache(t *testing.T) {
	logrus.SetOutput(ioutil.Discard) // Discard log output for test
	srv := setUpCacheServer()
	tt := []struct {
		name    string
		key     string
		value   string
		errCode codes.Code
	}{
		{"valid case", "newKey", "newValue", codes.OK},
		{"already set key", "testKey", "testVal", codes.AlreadyExists},
		{"empty key", "", "emptyVal", codes.InvalidArgument},
		{"empty value", "key", "", codes.InvalidArgument},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			testKV := &pb.Pair{Key: tc.key, Value: tc.value}

			res, err := srv.Add(context.Background(), testKV)
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
	srv := setUpCacheServer()
	tt := []struct {
		name    string
		key     string
		value   string
		errCode codes.Code
	}{
		{"valid case", "testKey", "testValue", codes.OK},
		{"key not found", "newKey", "", codes.NotFound},
		{"empty key", "", "", codes.InvalidArgument},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			testK := &pb.Key{Key: tc.key}

			res, err := srv.Get(context.Background(), testK)
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
	srv := setUpCacheServer()
	tt := []struct {
		name    string
		key     string
		errCode codes.Code
	}{
		{"valid case", "testKey", codes.OK},
		{"key not found", "newKey", codes.NotFound},
		{"empty key", "", codes.InvalidArgument},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			testKV := &pb.Key{Key: tc.key}

			_, err := srv.Delete(context.Background(), testKV)
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
