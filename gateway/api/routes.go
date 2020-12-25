package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type kv struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// Get handles the Retrieve of a value for a given key
func (g *Gateway) Get(w http.ResponseWriter, r *http.Request) {
	if g.Client == nil {
		respondWithErrorCode(w, http.StatusServiceUnavailable, "cache server is not ready yet")
		return
	}
	logrus.Debug("Serving Get request")
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		respondWithErrorCode(w, http.StatusBadRequest, "No key supplied with request")
		return
	}

	value, err := g.Client.GetValue(r.Context(), key)
	if err != nil {
		respondWithError(w, err)
		return
	}

	resp := kv{Key: key, Value: value}

	respondWithJSON(w, http.StatusOK, resp)
	return
}

// Set handles the Setting of a key value pair
func (g *Gateway) Set(w http.ResponseWriter, r *http.Request) {
	if g.Client == nil {
		respondWithErrorCode(w, http.StatusServiceUnavailable, "cache server is not ready yet")
		return
	}
	logrus.Debug("Serving Set request")
	decoder := json.NewDecoder(r.Body)
	var data kv
	err := decoder.Decode(&data)
	if err != nil {
		respondWithErrorCode(w, http.StatusBadRequest, "data could not be decoded")
		return
	}

	key := data.Key
	value := data.Value

	err = g.Client.SetValue(r.Context(), key, value)
	if err != nil {
		respondWithError(w, err)
		return
	}
	respondWithJSON(w, http.StatusCreated, "key and value successfully added to cache")
	return
}

// Delete handles the removal of a value for a given key
func (g *Gateway) Delete(w http.ResponseWriter, r *http.Request) {
	if g.Client == nil {
		respondWithErrorCode(w, http.StatusServiceUnavailable, "cache server is not ready yet")
		return
	}
	logrus.Debug("Serving Delete request")
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		respondWithErrorCode(w, http.StatusBadRequest, "no key supplied with request")
		return
	}

	err := g.Client.DelValue(r.Context(), key)
	if err != nil {
		respondWithError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, "key deleted")
	return
}
