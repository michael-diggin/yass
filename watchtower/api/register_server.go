package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type serverAddr struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// RegisterCacheServer registers a new cache server and grpc client to the API WatchTower
func (wt *WatchTower) RegisterCacheServer(w http.ResponseWriter, r *http.Request) {
	// TODO: add security to this endpoint
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
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	dbClient, err := wt.clientFactory.New(ctx, addr)
	if err != nil {
		logrus.Errorf("failed to dial server: %v", err)
		respondWithError(w, err)
		return
	}
	wt.mu.Lock()
	if len(wt.Clients) < wt.numServers {
		// new node for initial creation -> no rebalancing
		logrus.Info("Registering a new server for gateway")
		wt.Clients[addr] = dbClient
		wt.hashRing.AddNode(addr)
		wt.mu.Unlock()
		respondWithJSON(w, http.StatusCreated, "server registered with gateway")
		return
	}
	if _, ok := wt.Clients[addr]; ok {
		// node failed and restarted -> repopulate from other nodes
		logrus.Info("Registering a failed node that restarted")
		wt.Clients[addr] = dbClient
		instructions := wt.hashRing.RebalanceInstructions(addr)
		wt.mu.Unlock()
		wt.rebalanceData(addr, instructions, false)
		respondWithJSON(w, http.StatusCreated, "server registered with gateway")
		return
	}
	// must be a new node added for scaling -> rebalance data from other nodes
	logrus.Info("Registering a new node for scaling")
	wt.Clients[addr] = dbClient
	wt.mu.Unlock()
	wt.hashRing.AddNode(addr)
	instructions := wt.hashRing.RebalanceInstructions(addr)
	wt.rebalanceData(addr, instructions, true)
	respondWithJSON(w, http.StatusCreated, "server registered with gateway")
	return
}
