package api

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func respondWithError(w http.ResponseWriter, code int, message string) {
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
