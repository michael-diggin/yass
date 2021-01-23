package api

import (
	"net/http"
	"sync"

	"github.com/michael-diggin/yass/common/hashring"
	"github.com/michael-diggin/yass/common/models"
	"github.com/sirupsen/logrus"

	pb "github.com/michael-diggin/yass/proto"
)

// WatchTower holds the router and the grpc clients
type WatchTower struct {
	clientFactory models.ClientFactory
	Clients       map[string]pb.StorageClient
	numServers    int
	mu            sync.RWMutex
	hashRing      models.HashRing
	replicas      int
	nodeFile      string
}

// NewWatchTower will initialize the application
func NewWatchTower(numServers, weight int, factory models.ClientFactory, nodeFile string) *WatchTower {
	wt := WatchTower{}

	wt.clientFactory = factory

	wt.Clients = make(map[string]pb.StorageClient)
	wt.numServers = numServers
	wt.mu = sync.RWMutex{}
	wt.hashRing = hashring.New(weight)
	wt.replicas = 2
	wt.nodeFile = nodeFile
	return &wt
}

//ServeHTTP will serve and route a request
func (wt *WatchTower) ServeHTTP(w http.ResponseWriter, r *http.Request) {}

// Stop will close all grpc connections the watchtower holds
func (wt *WatchTower) Stop() {
	wt.mu.Lock()
	for num := range wt.Clients {
		serverNum := num
		// client.Close() TODO: should use an interface with a close method
		delete(wt.Clients, serverNum)
	}
	logrus.Info("Closed connections to storage servers")
}
