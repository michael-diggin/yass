package api

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func respondWithError(w http.ResponseWriter, err error) {
	logrus.Errorf("encountered error: %v", err)
	if e, ok := status.FromError(err); ok {
		code, message := grpcToHTTP(e)
		respondWithErrorCode(w, code, message)
		return
	}
	respondWithErrorCode(w, http.StatusInternalServerError, err.Error())
	return
}

func respondWithErrorCode(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		code = http.StatusInternalServerError
		payload = map[string]interface{}{"error returning output": err.Error()}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func grpcToHTTP(e *status.Status) (int, string) {
	switch e.Code() {
	case codes.AlreadyExists:
		return http.StatusAlreadyReported, "key already exists"
	case codes.InvalidArgument:
		return http.StatusNotAcceptable, e.Message()
	case codes.Canceled:
		return http.StatusRequestTimeout, "context cancelled"
	case codes.NotFound:
		return http.StatusNotFound, "key not found"
	default:
		return http.StatusInternalServerError, "something went wrong"
	}
}
