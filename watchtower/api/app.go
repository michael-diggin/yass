package api

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/michael-diggin/yass/common/hashring"
	"github.com/michael-diggin/yass/common/models"
	"github.com/sirupsen/logrus"
)

// WatchTower holds the router and the grpc clients
type WatchTower struct {
	router        *mux.Router
	clientFactory models.ClientFactory
	Clients       map[string]models.ClientInterface
	numServers    int
	mu            sync.RWMutex
	hashRing      models.HashRing
	replicas      int
}

// NewWatchTower will initialize the application
func NewWatchTower(numServers, weight int, factory models.ClientFactory) *WatchTower {
	wt := WatchTower{}

	wt.router = mux.NewRouter()
	wt.router.HandleFunc("/get/{key}", wt.Get).Methods("GET")
	wt.router.HandleFunc("/set", wt.Set).Methods("POST")
	wt.router.HandleFunc("/register", wt.RegisterCacheServer).Methods("POST")

	wt.clientFactory = factory

	wt.Clients = make(map[string]models.ClientInterface)
	wt.numServers = numServers
	wt.mu = sync.RWMutex{}
	wt.hashRing = hashring.New(weight)
	wt.replicas = 2
	return &wt
}

//ServeHTTP will serve and route a request
func (wt *WatchTower) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wt.router.ServeHTTP(w, r)
}

// Stop will close all grpc connections the watchtower holds
func (wt *WatchTower) Stop() {
	wt.mu.Lock()
	for num, client := range wt.Clients {
		serverNum := num
		client.Close()
		delete(wt.Clients, serverNum)
	}
	logrus.Info("Closed connections to storage servers")
}
