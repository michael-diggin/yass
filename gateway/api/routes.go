package api

import (
	"context"
	"encoding/json"
	"hash/fnv"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/michael-diggin/yass/models"
	"github.com/sirupsen/logrus"
)

// Get handles the Retrieve of a value for a given key
func (g *Gateway) Get(w http.ResponseWriter, r *http.Request) {
	if len(g.Clients) != g.numServers {
		respondWithErrorCode(w, http.StatusServiceUnavailable, "server is not ready yet")
		return
	}
	logrus.Debug("Serving Get request")
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		respondWithErrorCode(w, http.StatusBadRequest, "No key supplied with request")
		return
	}

	hash := getHashOfKey(key)
	server := hash % g.numServers
	follower := (server + 1) % g.numServers
	g.mu.RLock()
	client := g.Clients[server]
	followerClient := g.Clients[follower]
	g.mu.RUnlock()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	var err error
	var value interface{}

	type internalResponse struct {
		value interface{}
		err   error
	}

	resps := make(chan internalResponse, 2)
	go func() {
		pair, err := client.GetValue(ctx, key, models.MainReplica)
		if err != nil {
			resps <- internalResponse{err: err}
			return
		}
		resps <- internalResponse{value: pair.Value, err: err}
	}()
	go func() {
		pair, err := followerClient.GetValue(ctx, key, models.BackupReplica)
		if err != nil {
			resps <- internalResponse{err: err}
			return
		}
		resps <- internalResponse{value: pair.Value, err: err}
	}()

	for i := 0; i < 2; i++ {
		r := <-resps
		if r.value != nil {
			value = r.value
			cancel()
			break
		}
		err = r.err
	}

	if value == nil && err != nil {
		respondWithError(w, err)
		return
	}

	resp := models.Pair{Key: key, Value: value}
	respondWithJSON(w, http.StatusOK, resp)
	return
}

// Set handles the Setting of a key value pair
func (g *Gateway) Set(w http.ResponseWriter, r *http.Request) {
	if len(g.Clients) != g.numServers {
		respondWithErrorCode(w, http.StatusServiceUnavailable, "server is not ready yet")
		return
	}
	logrus.Debug("Serving Set request")
	decoder := json.NewDecoder(r.Body)
	var pair models.Pair
	err := decoder.Decode(&pair)
	if err != nil {
		respondWithErrorCode(w, http.StatusBadRequest, "data could not be decoded")
		return
	}

	hash := getHashOfKey(pair.Key)
	server := hash % g.numServers
	follower := (server + 1) % g.numServers
	g.mu.RLock()
	client := g.Clients[server]
	followerClient := g.Clients[follower]
	g.mu.RUnlock()

	err = client.SetValue(r.Context(), &pair, models.MainReplica)
	if err != nil {
		respondWithError(w, err)
		return
	}
	// TODO: This should be done async
	err = followerClient.SetValue(r.Context(), &pair, models.BackupReplica)
	if err != nil {
		respondWithError(w, err)
		return
	}

	respondWithJSON(w, http.StatusCreated, "key and value successfully added")
	return
}

// Delete handles the removal of a value for a given key
func (g *Gateway) Delete(w http.ResponseWriter, r *http.Request) {
	if len(g.Clients) != g.numServers {
		respondWithErrorCode(w, http.StatusServiceUnavailable, "server is not ready yet")
		return
	}
	logrus.Debug("Serving Delete request")
	vars := mux.Vars(r)
	key, ok := vars["key"]
	if !ok {
		respondWithErrorCode(w, http.StatusBadRequest, "no key supplied with request")
		return
	}

	hash := getHashOfKey(key)
	server := hash % g.numServers
	follower := (server + 1) % g.numServers
	g.mu.RLock()
	client := g.Clients[server]
	followerClient := g.Clients[follower]
	g.mu.RUnlock()

	ctx := r.Context()

	err := client.DelValue(ctx, key, models.MainReplica)
	if err != nil {
		respondWithError(w, err)
		return
	}
	err = followerClient.DelValue(ctx, key, models.BackupReplica)
	if err != nil {
		respondWithError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, "key deleted")
	return
}

func getHashOfKey(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32())
}
