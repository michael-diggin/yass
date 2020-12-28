package core

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/stretchr/testify/require"
)

func TestBatchSettoStorage(t *testing.T) {
	r := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockLeader := mocks.NewMockService(ctrl)
	resp := make(chan error, 1)
	resp <- nil
	close(resp)

	mockLeader.EXPECT().BatchSet(gomock.Any()).
		DoAndReturn(func(data map[string]interface{}) <-chan error {
			if len(data) == 2 {
				return resp
			}
			t.FailNow()
			return nil
		})

	srv := server{Leader: mockLeader}

	dataOne := &pb.Pair{Key: "key-1", Value: []byte(`"value-1"`)}
	dataTwo := &pb.Pair{Key: "key-2", Value: []byte(`2`)}
	req := &pb.BatchSetRequest{Replica: 0, Data: []*pb.Pair{dataOne, dataTwo}}

	ctx := context.Background()
	_, err := srv.BatchSet(ctx, req)
	r.NoError(err)
}

func TestBatchGetFromStorage(t *testing.T) {
	r := require.New(t)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFollower := mocks.NewMockService(ctrl)
	resp := make(chan map[string]interface{}, 1)
	data := map[string]interface{}{
		"key-1": "value-1",
		"key-2": 2,
	}
	resp <- data
	close(resp)

	mockFollower.EXPECT().BatchGet().Return(resp)

	srv := server{Follower: mockFollower}

	req := &pb.BatchGetRequest{Replica: 1}

	ctx := context.Background()
	res, err := srv.BatchGet(ctx, req)
	r.NoError(err)
	r.Len(res.Data, 2)
	keys := []string{"key-1", "key-2"}
	vals := []interface{}{[]byte(`"value-1"`), []byte(`2`)}
	r.Contains(keys, res.Data[0].Key)
	r.Contains(keys, res.Data[1].Key)
	r.Contains(vals, res.Data[0].Value)
	r.Contains(vals, res.Data[1].Value)
}
