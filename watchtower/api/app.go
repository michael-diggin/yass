package api

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/michael-diggin/yass/common/hashring"
	"github.com/michael-diggin/yass/common/models"
	"github.com/sirupsen/logrus"
)

// Gateway holds the router and the grpc clients
type Gateway struct {
	router        *mux.Router
	clientFactory models.ClientFactory
	Clients       map[string]models.ClientInterface
	numServers    int
	mu            sync.RWMutex
	hashRing      models.HashRing
	replicas      int
}

// NewGateway will initialize the application
func NewGateway(numServers, weight int, factory models.ClientFactory) *Gateway {
	g := Gateway{}

	g.router = mux.NewRouter()
	g.router.HandleFunc("/get/{key}", g.Get).Methods("GET")
	g.router.HandleFunc("/set", g.Set).Methods("POST")
	g.router.HandleFunc("/register", g.RegisterCacheServer).Methods("POST")

	g.clientFactory = factory

	g.Clients = make(map[string]models.ClientInterface)
	g.numServers = numServers
	g.mu = sync.RWMutex{}
	g.hashRing = hashring.New(weight)
	g.replicas = 2
	return &g
}

//ServeHTTP will serve and route a request
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.router.ServeHTTP(w, r)
}

// Stop will close all grpc connections the gateway holds
func (g *Gateway) Stop() {
	g.mu.Lock()
	for num, client := range g.Clients {
		serverNum := num
		client.Close()
		delete(g.Clients, serverNum)
	}
	logrus.Info("Closed connections to storage servers")
}
