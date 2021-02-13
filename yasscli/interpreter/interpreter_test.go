package interpreter

import (
	"bytes"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/michael-diggin/yass"
	"github.com/michael-diggin/yass/mocks"
	"github.com/stretchr/testify/require"

	pb "github.com/michael-diggin/yass/proto"
)

func TestInterpreterProcessPUTCommands(t *testing.T) {
	t.Run("with correct command", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		key := "key"

		mockgRPC := mocks.NewMockYassServiceClient(ctrl)
		pair := &pb.Pair{Key: key, Value: []byte(`"value"`)}
		mockgRPC.EXPECT().Put(gomock.Any(), pair).Return(nil, nil)
		cc := yass.Client{GrpcClient: mockgRPC}

		interpreter := New(&cc)
		command := "PUT key value"
		b := &bytes.Buffer{}
		exit := interpreter.ProcessCommand(command, b)
		require.False(t, exit)
		require.Empty(t, b.String(), "no write when successful")
	})

	t.Run("with lower case command", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		key := "key"

		mockgRPC := mocks.NewMockYassServiceClient(ctrl)
		pair := &pb.Pair{Key: key, Value: []byte(`"value"`)}
		mockgRPC.EXPECT().Put(gomock.Any(), pair).Return(nil, nil)
		cc := yass.Client{GrpcClient: mockgRPC}

		interpreter := New(&cc)
		command := "put key value"
		b := &bytes.Buffer{}
		exit := interpreter.ProcessCommand(command, b)
		require.False(t, exit)
		require.Empty(t, b.String(), "no write when successful")
	})

	t.Run("with incorrect command", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockgRPC := mocks.NewMockYassServiceClient(ctrl)
		cc := yass.Client{GrpcClient: mockgRPC}

		interpreter := New(&cc)
		command := "put key value please"
		b := &bytes.Buffer{}
		exit := interpreter.ProcessCommand(command, b)
		require.False(t, exit)
		require.Contains(t, b.String(), "incorrect number of inputs")
	})

	t.Run("with error from server", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		key := "key"
		putError := errors.New("put error")

		mockgRPC := mocks.NewMockYassServiceClient(ctrl)
		pair := &pb.Pair{Key: key, Value: []byte(`"value"`)}
		mockgRPC.EXPECT().Put(gomock.Any(), pair).Return(nil, putError)
		cc := yass.Client{GrpcClient: mockgRPC}

		interpreter := New(&cc)
		command := "put key value"
		b := &bytes.Buffer{}
		exit := interpreter.ProcessCommand(command, b)
		require.False(t, exit)
		require.Contains(t, b.String(), putError.Error())
	})
}

func TestInterpreterProcessFetchCommands(t *testing.T) {
	t.Run("with correct command", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		key := "key"

		mockgRPC := mocks.NewMockYassServiceClient(ctrl)
		pbKey := &pb.Key{Key: key}
		mockgRPC.EXPECT().Fetch(gomock.Any(), pbKey).Return(&pb.Pair{Key: key, Value: []byte(`"value"`)}, nil)
		cc := yass.Client{GrpcClient: mockgRPC}

		interpreter := New(&cc)
		command := "FETCH key"
		b := &bytes.Buffer{}
		exit := interpreter.ProcessCommand(command, b)
		require.False(t, exit)
		require.Equal(t, b.String(), "value\n")
	})

	t.Run("with lower case command", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		key := "key"

		mockgRPC := mocks.NewMockYassServiceClient(ctrl)
		pbKey := &pb.Key{Key: key}
		mockgRPC.EXPECT().Fetch(gomock.Any(), pbKey).Return(&pb.Pair{Key: key, Value: []byte(`"value"`)}, nil)
		cc := yass.Client{GrpcClient: mockgRPC}

		interpreter := New(&cc)
		command := "fetch key"
		b := &bytes.Buffer{}
		exit := interpreter.ProcessCommand(command, b)
		require.False(t, exit)
		require.Equal(t, b.String(), "value\n")
	})

	t.Run("with incorrect command", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockgRPC := mocks.NewMockYassServiceClient(ctrl)
		cc := yass.Client{GrpcClient: mockgRPC}

		interpreter := New(&cc)
		command := "fetch key value"
		b := &bytes.Buffer{}
		exit := interpreter.ProcessCommand(command, b)
		require.False(t, exit)
		require.Contains(t, b.String(), "incorrect number of inputs")
	})

	t.Run("with server error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		key := "key"
		fetchError := errors.New("fetch error")

		mockgRPC := mocks.NewMockYassServiceClient(ctrl)
		pbKey := &pb.Key{Key: key}
		mockgRPC.EXPECT().Fetch(gomock.Any(), pbKey).Return(nil, fetchError)
		cc := yass.Client{GrpcClient: mockgRPC}

		interpreter := New(&cc)
		command := "fetch key"
		b := &bytes.Buffer{}
		exit := interpreter.ProcessCommand(command, b)
		require.False(t, exit)
		require.Contains(t, b.String(), fetchError.Error())
	})
}

func TestInterpreterProcessUnknownCommand(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockgRPC := mocks.NewMockYassServiceClient(ctrl)
	cc := yass.Client{GrpcClient: mockgRPC}

	interpreter := New(&cc)
	command := "get key"
	b := &bytes.Buffer{}
	exit := interpreter.ProcessCommand(command, b)
	require.False(t, exit)
	require.Contains(t, b.String(), "unknown request")
}

func TestInterpreterQuitsWithCorrectInput(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockgRPC := mocks.NewMockYassServiceClient(ctrl)
	cc := yass.Client{GrpcClient: mockgRPC}

	interpreter := New(&cc)
	command := "\\q"
	b := &bytes.Buffer{}
	exit := interpreter.ProcessCommand(command, b)
	require.True(t, exit)
	require.Contains(t, b.String(), "Quitting")
}
