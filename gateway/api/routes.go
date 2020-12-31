package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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

	clientIDs, err := g.hashRing.GetN(key, g.replicas)
	if err != nil {
		respondWithErrorCode(w, http.StatusInternalServerError, "something went wrong")
	}

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	var value interface{}

	type internalResponse struct {
		value interface{}
		err   error
	}

	resps := make(chan internalResponse, len(clientIDs))
	for _, addr := range clientIDs {
		g.mu.RLock()
		client := g.Clients[addr]
		g.mu.RUnlock()
		go func() {
			subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			pair, err := client.GetValue(subctx, key, models.MainReplica)
			if err != nil {
				resps <- internalResponse{err: err}
				return
			}
			resps <- internalResponse{value: pair.Value, err: err}
		}()
	}

	var retErr error
	// valMap := make(map[interface{}]int)
	for i := 0; i < len(clientIDs); i++ {
		r := <-resps
		if r.err != nil && retErr == nil {
			retErr = r.err
		}
		if r.value != nil {
			value = r.value
			cancel()
			break
		}
	}

	if value == nil && retErr != nil {
		respondWithError(w, retErr)
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

	clientIDs, err := g.hashRing.GetN(pair.Key, g.replicas)
	if err != nil {
		respondWithErrorCode(w, http.StatusInternalServerError, "something went wrong")
	}

	ctx := r.Context()

	resps := make(chan error, len(clientIDs))
	for _, addr := range clientIDs {
		g.mu.RLock()
		client := g.Clients[addr]
		g.mu.RUnlock()
		go func() {
			subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			resps <- client.SetValue(subctx, &pair, models.MainReplica)
			cancel()
		}()
	}

	var retErr error
	var nilErrs int
	for i := 0; i < len(clientIDs); i++ {
		err := <-resps
		if err != nil && retErr == nil {
			retErr = err
			continue
		}
		nilErrs++
		if nilErrs > (g.replicas-1)/2 {
			retErr = nil
			break
		}
	}
	if retErr != nil {
		respondWithError(w, retErr)
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

	clientIDs, err := g.hashRing.GetN(key, g.replicas)
	if err != nil {
		respondWithErrorCode(w, http.StatusInternalServerError, "something went wrong")
	}

	ctx := r.Context()

	resps := make(chan error, len(clientIDs))
	for _, addr := range clientIDs {
		g.mu.RLock()
		client := g.Clients[addr]
		g.mu.RUnlock()
		go func() {
			subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			resps <- client.DelValue(subctx, key, models.MainReplica)
			cancel()
		}()
	}

	var retErr error
	var nilErrs int
	for i := 0; i < len(clientIDs); i++ {
		err := <-resps
		if err != nil && retErr == nil {
			retErr = err
			continue
		}
		nilErrs++
		if nilErrs > (g.replicas-1)/2 {
			retErr = nil
			break
		}
	}

	if retErr != nil {
		respondWithError(w, retErr)
		return
	}

	respondWithJSON(w, http.StatusOK, "key deleted")
	return
}
