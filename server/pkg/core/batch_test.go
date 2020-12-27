package core

import (
	"context"
	"errors"
	"testing"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/mocks"
	"github.com/stretchr/testify/require"
)

func TestBatchSettoStorage(t *testing.T) {
	r := require.New(t)

	l := &mocks.TestStorage{
		BatchSetFn: func(data map[string]interface{}) error {
			if len(data) != 2 {
				return errors.New("Did not pass correct data")
			}
			return nil
		},
	}
	srv := server{Leader: l}
	dataOne := &pb.Pair{Key: "key-1", Value: []byte(`"value-1"`)}
	dataTwo := &pb.Pair{Key: "key-2", Value: []byte(`2`)}
	req := &pb.BatchSetRequest{Replica: 0, Data: []*pb.Pair{dataOne, dataTwo}}

	ctx := context.Background()
	_, err := srv.BatchSet(ctx, req)
	r.NoError(err)
}

func TestBatchGetFromStorage(t *testing.T) {
	r := require.New(t)

	f := &mocks.TestStorage{
		BatchGetFn: func() map[string]interface{} {
			return map[string]interface{}{
				"key-1": "value",
				"key-2": 2,
			}
		},
	}
	srv := server{Follower: f}
	req := &pb.BatchGetRequest{Replica: 1}

	ctx := context.Background()
	res, err := srv.BatchGet(ctx, req)
	r.NoError(err)
	r.Len(res.Data, 2)
	r.Equal(res.Data[0].Key, "key-1")
	r.Equal(res.Data[0].Value, []byte(`"value"`))
	r.Equal(res.Data[1].Key, "key-2")
	r.Equal(res.Data[1].Value, []byte(`2`))
}
