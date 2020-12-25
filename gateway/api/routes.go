package api

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/status"

	"github.com/gorilla/mux"
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
			code, message := grpcToHTTP(e)
			respondWithError(w, code, message)
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := map[string]interface{}{
		"key":   key,
		"value": value,
	}

	respondWithJSON(w, http.StatusOK, resp)
	return
}

// Set handles the Setting of a key value pair
func (g *Gateway) Set(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var data kv
	err := decoder.Decode(&data)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "data could not be decoded")
		return
	}

	key := data.Key
	value := data.Value

	err = g.Client.SetValue(r.Context(), key, value)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			code, message := grpcToHTTP(e)
			respondWithError(w, code, message)
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, "key and value successfully added to cache")
	return
}

// Delete handles the removal of a value for a given key
func (g *Gateway) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "no key supplied with request")
		return
	}

	err := g.Client.DelValue(r.Context(), key)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			code, message := grpcToHTTP(e)
			respondWithError(w, code, message)
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, "key deleted")
	return
}
