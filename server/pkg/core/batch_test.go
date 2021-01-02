package core

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/michael-diggin/yass/server/model"
	"github.com/stretchr/testify/require"
)

func TestBatchSettoStorage(t *testing.T) {
	r := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockStore := mocks.NewMockService(ctrl)
	resp := make(chan error, 1)
	resp <- nil
	close(resp)

	mockStore.EXPECT().BatchSet(gomock.Any()).
		DoAndReturn(func(data map[string]model.Data) <-chan error {
			if len(data) == 2 {
				return resp
			}
			t.FailNow()
			return nil
		})

	srv := server{MainReplica: mockStore}

	dataOne := &pb.Pair{Key: "key-1", Hash: uint32(100), Value: []byte(`"value-1"`)}
	dataTwo := &pb.Pair{Key: "key-2", Hash: uint32(101), Value: []byte(`2`)}
	req := &pb.BatchSetRequest{Replica: 0, Data: []*pb.Pair{dataOne, dataTwo}}

	ctx := context.Background()
	_, err := srv.BatchSet(ctx, req)
	r.NoError(err)
}

func TestBatchGetFromStorage(t *testing.T) {
	r := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockBackup := mocks.NewMockService(ctrl)
	resp := make(chan map[string]model.Data, 1)
	data := map[string]model.Data{
		"key-1": model.Data{Value: "value-1", Hash: uint32(100)},
		"key-2": model.Data{Value: 2, Hash: uint32(101)},
	}
	resp <- data
	close(resp)

	mockBackup.EXPECT().BatchGet().Return(resp)

	srv := server{BackupReplica: mockBackup}

	req := &pb.BatchGetRequest{Replica: 1}

	ctx := context.Background()
	res, err := srv.BatchGet(ctx, req)
	r.NoError(err)
	r.Len(res.Data, 2)
	keys := []string{"key-1", "key-2"}
	vals := []interface{}{[]byte(`"value-1"`), []byte(`2`)}
	hashes := []uint32{uint32(100), uint32(101)}
	r.Contains(keys, res.Data[0].Key)
	r.Contains(keys, res.Data[1].Key)
	r.Contains(vals, res.Data[0].Value)
	r.Contains(vals, res.Data[1].Value)
	r.Contains(hashes, res.Data[0].Hash)
	r.Contains(hashes, res.Data[1].Hash)
}
