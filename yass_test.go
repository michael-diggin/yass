package yass

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/michael-diggin/yass/mocks"
	"github.com/stretchr/testify/require"

	pb "github.com/michael-diggin/yass/proto"
)

func TestClientPutValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	key := "test"
	val := "value"

	mockgRPC := mocks.NewMockYassServiceClient(ctrl)
	pair := &pb.Pair{Key: key, Value: []byte(`"value"`)}
	mockgRPC.EXPECT().Put(gomock.Any(), pair).Return(nil, nil)
	cc := Client{GrpcClient: mockgRPC, conn: nil}

	err := cc.Put(context.Background(), key, val)
	require.NoError(t, err)
}

func TestClientGetValue(t *testing.T) {
	errTest := errors.New("Not in storage")

	tt := []struct {
		name  string
		key   string
		value interface{}
		err   error
	}{
		{"valid case", "test", "value", nil},
		{"err case", "bad", nil, errTest},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockgRPC := mocks.NewMockYassServiceClient(ctrl)
			mockgRPC.EXPECT().Fetch(gomock.Any(), &pb.Key{Key: tc.key}).
				Return(&pb.Pair{Key: tc.key, Value: []byte(`"value"`)}, tc.err)

			cc := Client{GrpcClient: mockgRPC, conn: nil}
			resp, err := cc.Fetch(context.Background(), tc.key)

			require.Equal(t, err, tc.err)
			if tc.value != nil {
				require.Equal(t, tc.value, resp)
			}
		})
	}
}
