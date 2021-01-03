package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/michael-diggin/yass/common/client"
	"github.com/michael-diggin/yass/gateway/models"
	"github.com/sirupsen/logrus"
)

type serverAddr struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// RegisterCacheServer registers a new cache server and grpc client to the API Gateway
func (g *Gateway) RegisterCacheServer(w http.ResponseWriter, r *http.Request) {
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
	dbClient, err := client.NewClient(ctx, addr)
	if err != nil {
		logrus.Errorf("failed to dial server: %v", err)
		respondWithError(w, err)
		return
	}
	g.mu.Lock()
	if len(g.Clients) < g.numServers {
		// new node for initial creation -> no rebalancing
		logrus.Info("Registering a new server for gateway")
		g.Clients[addr] = dbClient
		g.hashRing.AddNode(addr)
		g.mu.Unlock()
		respondWithJSON(w, http.StatusCreated, "key and value successfully added to cache")
		return
	}
	if _, ok := g.Clients[addr]; ok {
		// node failed and restarted -> repopulate from other nodes
		logrus.Info("Registering a failed node that restarted")
		g.Clients[addr] = dbClient
		instructions := g.hashRing.RebalanceInstructions(addr)
		g.mu.Unlock()
		g.rebalanceData(addr, instructions, false)
		respondWithJSON(w, http.StatusCreated, "key and value successfully added to cache")
		return
	}
	// must be a new node added for scaling -> rebalance data from other nodes
	logrus.Info("Registering a new node for scaling")
	g.Clients[addr] = dbClient
	g.mu.Unlock()
	g.hashRing.AddNode(addr)
	instructions := g.hashRing.RebalanceInstructions(addr)
	g.rebalanceData(addr, instructions, true)
	respondWithJSON(w, http.StatusCreated, "key and value successfully added to cache")
	return
}

func (g *Gateway) rebalanceData(addr string, instructions []models.Instruction, delete bool) error {
	for _, instr := range instructions {
		go func(instr models.Instruction) {
			g.mu.RLock()
			dbClient := g.Clients[instr.FromNode]
			g.mu.RUnlock()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			err := dbClient.BatchSend(ctx, instr.FromIdx, instr.ToIdx, addr, instr.LowHash, instr.HighHash)
			if err != nil {
				logrus.Errorf("failed to send data from node %v: %v", instr.FromNode, err)
				return
			}
			if delete {
				dbClient.BatchDelete(ctx, instr.FromIdx, instr.LowHash, instr.HighHash)
			}
			return
		}(instr)
	}
	return nil
}
