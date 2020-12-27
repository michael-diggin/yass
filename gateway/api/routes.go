package api

import (
	"encoding/json"
	"hash/fnv"
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
	g.mu.RLock()
	client := g.Clients[server]
	g.mu.RUnlock()

	value, err := client.GetValue(r.Context(), key)
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
	if len(g.Clients) != g.numServers {
		respondWithErrorCode(w, http.StatusServiceUnavailable, "server is not ready yet")
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

	hash := getHashOfKey(key)
	server := hash % g.numServers
	follower := (server + 1) % g.numServers
	g.mu.RLock()
	client := g.Clients[server]
	followerClient := g.Clients[follower]
	g.mu.RUnlock()

	err = client.SetValue(r.Context(), key, value)
	if err != nil {
		respondWithError(w, err)
		return
	}
	err = followerClient.SetFollowerValue(r.Context(), key, value)
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

	err := client.DelValue(ctx, key)
	if err != nil {
		respondWithError(w, err)
		return
	}
	err = followerClient.DelFollowerValue(ctx, key)
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
