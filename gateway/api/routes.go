package api

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type kv struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Get handles the Retrieve of a value for a given key
func (g *Gateway) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "No key supplied with request")
		return
	}

	value, err := g.Client.GetValue(r.Context(), key)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				respondWithError(w, http.StatusNotFound, "Key not found")
			default:
				respondWithError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
	}
	resp := map[string]interface{}{
		"key":   key,
		"value": value,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// Set handles the Setting of a key value pair
func (g *Gateway) Set(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var data kv
	err := decoder.Decode(&data)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Data could not be decoded")
	}

	key := data.Key
	value := data.Value

	logrus.Infof("Setting %s, %v", key, value)

	err = g.Client.SetValue(r.Context(), key, value)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.AlreadyExists:
				respondWithError(w, http.StatusAlreadyReported, "key already exists")
			default:
				respondWithError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, "Key and value successfully added to cache")
}

// Delete handles the removal of a value for a given key
func (g *Gateway) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "No key supplied with request")
		return
	}

	err := g.Client.DelValue(r.Context(), key)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			switch e.Code() {
			case codes.NotFound:
				respondWithError(w, http.StatusNotFound, "Key not found")
			default:
				respondWithError(w, http.StatusInternalServerError, err.Error())
			}
			return
		}
	}

	respondWithJSON(w, http.StatusOK, "key deleted")
}
