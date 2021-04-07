package core

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/michael-diggin/yass/server/model"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSettoStorage(t *testing.T) {
	r := require.New(t)
	tt := []struct {
		name    string
		key     string
		hash    uint32
		value   []byte
		errCode codes.Code
		timeout time.Duration
	}{
		{"valid case", "newKey", uint32(100), []byte(`"newValue"`), codes.OK, time.Second},
		{"already set key", "testKey", uint32(101), []byte(`"testVal"`), codes.AlreadyExists, time.Second},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockMainStore := mocks.NewMockService(ctrl)
			resp := make(chan *model.StorageResponse, 1)

			resp <- &model.StorageResponse{
				Key:   tc.key,
				Value: string(tc.value),
				Err:   status.Error(tc.errCode, "")}
			close(resp)

			mockMainStore.EXPECT().Set(tc.key, tc.hash, gomock.Any(), true).Return(resp)

			srv := server{DataStores: []model.Service{mockMainStore}}
			testKV := &pb.Pair{Key: tc.key, Hash: tc.hash, Value: tc.value}
			req := &pb.SetRequest{Replica: 0, Pair: testKV}

			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			_, err := srv.Set(ctx, req)
			defer cancel()
			if e, ok := status.FromError(err); ok {
				r.Equal(tc.errCode, e.Code())
			}
		})
	}
}

func TestGetFromStorage(t *testing.T) {
	r := require.New(t)
	tt := []struct {
		name    string
		key     string
		value   interface{}
		errCode codes.Code
		timeout time.Duration
	}{
		{"valid case", "testKey", "testValue", codes.OK, 100 * time.Millisecond},
		{"key not found", "newKey", "not found", codes.NotFound, 100 * time.Millisecond},
		{"empty key", "", nil, codes.InvalidArgument, 100 * time.Millisecond},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockMainStore := mocks.NewMockService(ctrl)
			resp := make(chan *model.StorageResponse, 1)

			resp <- &model.StorageResponse{
				Key:   tc.key,
				Value: tc.value,
				Err:   status.Error(tc.errCode, "")}
			close(resp)

			if tc.key != "" && tc.value != nil {
				mockMainStore.EXPECT().Get(tc.key).Return(resp)
			}

			srv := server{DataStores: []model.Service{mockMainStore}}
			req := &pb.GetRequest{Replica: 0, Key: tc.key}
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			res, err := srv.Get(ctx, req)
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

func TestDeleteKeyValue(t *testing.T) {
	r := require.New(t)
	tt := []struct {
		name    string
		key     string
		errCode codes.Code
		timeout time.Duration
	}{
		{"valid case", "testKey", codes.OK, 100 * time.Millisecond},
		{"empty key", "", codes.InvalidArgument, 100 * time.Millisecond},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockMainStore := mocks.NewMockService(ctrl)
			resp := make(chan *model.StorageResponse, 1)

			resp <- &model.StorageResponse{
				Key: tc.key,
				Err: status.Error(tc.errCode, "")}
			close(resp)

			if tc.key != "" {
				mockMainStore.EXPECT().Delete(tc.key).Return(resp)
			}

			srv := server{DataStores: []model.Service{mockMainStore}}

			req := &pb.DeleteRequest{Replica: 0, Key: tc.key}
			ctx, cancel := context.WithTimeout(context.Background(), tc.timeout)
			_, err := srv.Delete(ctx, req)
			cancel()
			e, ok := status.FromError(err)
			r.True(ok)
			r.Equal(tc.errCode, e.Code())

		})
	}
}

func TestDeleteKeyValueFromBackup(t *testing.T) {
	r := require.New(t)
	key := "test-del-key"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockMainStore := mocks.NewMockService(ctrl)
	mockStore := mocks.NewMockService(ctrl)
	resp := make(chan *model.StorageResponse, 1)

	resp <- &model.StorageResponse{
		Key: key,
		Err: status.Error(codes.OK, "")}
	close(resp)

	mockStore.EXPECT().Delete(key).Return(resp)

	srv := server{DataStores: []model.Service{mockMainStore, mockStore}}

	req := &pb.DeleteRequest{Replica: 1, Key: key}
	ctx := context.Background()
	_, err := srv.Delete(ctx, req)
	e, ok := status.FromError(err)
	r.True(ok)
	r.Equal(codes.OK, e.Code())
}
