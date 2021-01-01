package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/michael-diggin/yass"
	"github.com/sirupsen/logrus"
)

type serverAddr struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// RegisterCacheServer registers a new cache server and grpc client to the API Gateway
func (g *Gateway) RegisterCacheServer(w http.ResponseWriter, r *http.Request) {
	// TODO: add security to this endpoint
	logrus.Info("Registering a new cache server")
	decoder := json.NewDecoder(r.Body)
	var data serverAddr
	err := decoder.Decode(&data)
	if err != nil {
		logrus.Errorf("failed to register server: %v", err)
		respondWithErrorCode(w, http.StatusBadRequest, "data could not be decoded")
		return
	}

	ip := data.IP
	port := data.Port

	addr := ip + ":" + port
	client, err := yass.NewClient(context.Background(), addr)
	if err != nil {
		logrus.Errorf("failed to dial server: %v", err)
		respondWithError(w, err)
		return
	}
	g.mu.Lock()
	g.Clients[addr] = client
	g.hashRing.AddNode(addr)
	// TODO: add rebalance instruction when new node is added
	// new node requests appropriate data from other nodes
	g.mu.Unlock()
	respondWithJSON(w, http.StatusCreated, "key and value successfully added to cache")
	return
}
