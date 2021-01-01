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
	if len(g.Clients) < g.numServers-1 {
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

	value, retErr := getValueFromRequests(resps, len(clientIDs), cancel)

	if value == nil && retErr != nil {
		respondWithError(w, retErr)
		return
	}

	resp := models.Pair{Key: key, Value: value}
	respondWithJSON(w, http.StatusOK, resp)
	return
}

type internalResponse struct {
	value interface{}
	err   error
}

func getValueFromRequests(resps chan internalResponse, n int, cancel context.CancelFunc) (interface{}, error) {
	var err error
	var value interface{}
	// valMap := make(map[interface{}]int)
	for i := 0; i < n; i++ {
		r := <-resps
		if r.err != nil && err == nil {
			err = r.err
		}
		if r.value != nil {
			value = r.value
			cancel()
			break
		}
	}
	return value, err
}

// Set handles the Setting of a key value pair
func (g *Gateway) Set(w http.ResponseWriter, r *http.Request) {
	if len(g.Clients) < g.numServers-1 {
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

	// get node Addrs from hash ring
	clientIDs, err := g.hashRing.GetN(pair.Key, g.replicas)
	if err != nil {
		respondWithErrorCode(w, http.StatusInternalServerError, "something went wrong")
	}

	ctx := r.Context()

	// synchronously set the key/value on the storage servers
	revertSetAddr := []string{}
	var returnErr error
	for _, addr := range clientIDs {
		g.mu.RLock()
		client := g.Clients[addr]
		g.mu.RUnlock()
		subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		err := client.SetValue(subctx, &pair, models.MainReplica)
		cancel()
		if err != nil {
			returnErr = err
			break
		}
		revertSetAddr = append(revertSetAddr, addr)
	}

	if returnErr != nil {
		logrus.Errorf("Encountered error: %v", returnErr)
		// revert any changes that were made before an error
		for _, addr := range revertSetAddr {
			g.mu.RLock()
			client := g.Clients[addr]
			g.mu.RUnlock()
			go func() {
				subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
				client.DelValue(subctx, pair.Key, models.MainReplica)
				cancel()
			}()
		}

		respondWithError(w, returnErr)
		return
	}

	respondWithJSON(w, http.StatusCreated, "key and value successfully added")
	return
}
