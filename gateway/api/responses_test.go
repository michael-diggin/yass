package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestGrpcToHTTP(t *testing.T) {
	r := require.New(t)

	tt := []struct {
		name    string
		err     error
		code    int
		message string
	}{
		{"already exists", status.Error(codes.AlreadyExists, "already exists"), http.StatusAlreadyReported, "exists"},
		{"invalid argument", status.Error(codes.InvalidArgument, "key not set"), http.StatusNotAcceptable, "not set"},
		{"canceled", status.Error(codes.Canceled, "conext timedout"), http.StatusRequestTimeout, "cancelled"},
		{"not found", status.Error(codes.NotFound, "key doesn't exist"), http.StatusNotFound, "found"},
		{"internal error", status.Error(codes.Internal, "server issue"), http.StatusInternalServerError, "wrong"},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			e, ok := status.FromError(tc.err)
			r.True(ok)
			code, message := grpcToHTTP(e)
			r.Equal(code, tc.code)
			r.Contains(message, tc.message)
		})
	}
}

func TestRespondWithJson(t *testing.T) {
	w := httptest.NewRecorder()
	code := http.StatusOK
	payload := map[string]interface{}{
		"key":   "test-key",
		"value": "test-value",
	}

	respondWithJSON(w, code, payload)

	require.Equal(t, w.Code, code)
	var result kv
	json.Unmarshal(w.Body.Bytes(), &result)
	require.Equal(t, result.Key, "test-key")
	require.Equal(t, result.Value, "test-value")
}
