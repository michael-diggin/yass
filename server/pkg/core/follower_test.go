package core

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/michael-diggin/yass/server/model"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSettoFollower(t *testing.T) {
	r := require.New(t)
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
			follower := &mocks.TestStorage{
				SetFn: func(key string, value interface{}) *model.StorageResponse {
					return &model.StorageResponse{
						Key:   tc.key,
						Value: tc.value,
						Err:   status.Error(tc.errCode, "")}
				},
			}
			srv := server{Follower: follower}
			testKV := &pb.Pair{Key: tc.key, Value: tc.value}

			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			res, err := srv.SetFollower(ctx, testKV)
			cancel()
			if e, ok := status.FromError(err); ok {
				r.Equal(tc.errCode, e.Code())
			}
			if res != nil {
				r.Equal(testKV.Key, res.Key)
			}

		})
	}
}

func TestGetFollowerFromStorage(t *testing.T) {
	r := require.New(t)
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
			f := &mocks.TestStorage{
				GetFn: func(key string) *model.StorageResponse {
					return &model.StorageResponse{
						Key:   tc.key,
						Value: tc.value,
						Err:   status.Error(tc.errCode, "")}
				},
			}
			srv := server{Follower: f}
			testK := &pb.Key{Key: tc.key}
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			res, err := srv.GetFollower(ctx, testK)
			cancel()
			e, ok := status.FromError(err)
			r.True(ok)
			r.Equal(tc.errCode, e.Code())

			if res != nil {
				expectedBytes, _ := json.Marshal(tc.value)
				r.True(bytes.Equal(expectedBytes, res.Value))
			}
		})
	}
}

func TestDeleteFollowerKeyValue(t *testing.T) {
	r := require.New(t)
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
			follower := &mocks.TestStorage{
				DelFn: func(key string) *model.StorageResponse {
					return &model.StorageResponse{
						Key:   tc.key,
						Value: "",
						Err:   status.Error(tc.errCode, "")}
				},
			}
			srv := server{Follower: follower}
			testKV := &pb.Key{Key: tc.key}
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			_, err := srv.DeleteFollower(ctx, testKV)
			cancel()
			e, ok := status.FromError(err)
			r.True(ok)
			r.Equal(tc.errCode, e.Code())
		})
	}
}
