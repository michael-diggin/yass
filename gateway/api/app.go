package api

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/michael-diggin/yass/gateway/hashring"
	"github.com/sirupsen/logrus"
)

// Gateway holds the router and the grpc clients
type Gateway struct {
	*http.Server
	Clients    map[string]GrpcClient
	numServers int
	mu         sync.RWMutex
	hashRing   *hashring.Ring
	replicas   int
}

// NewGateway will initialize the application
func NewGateway(numServers, weight int, srv *http.Server) *Gateway {
	g := Gateway{}

	router := mux.NewRouter()
	router.HandleFunc("/get/{key}", g.Get).Methods("GET")
	router.HandleFunc("/set", g.Set).Methods("POST")
	router.HandleFunc("/register", g.RegisterCacheServer).Methods("POST")

	g.Server = srv

	g.Handler = router

	g.Clients = make(map[string]GrpcClient)
	g.numServers = numServers
	g.mu = sync.RWMutex{}
	g.hashRing = hashring.New(weight)
	g.replicas = 2
	return &g
}

//ServeHTTP will serve and route a request
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.Handler.ServeHTTP(w, r)
}

// Stop will close all grpc connections the gateway holds
func (g *Gateway) Stop() {
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := g.Shutdown(ctxShutDown); err != nil {
		logrus.Fatalf("server shutdown failed:%v", err)
	}
	logrus.Info("gateway server stopped")

	g.mu.Lock()
	for num, client := range g.Clients {
		serverNum := num
		client.Close()
		delete(g.Clients, serverNum)
	}
	logrus.Info("Closed connections to storage servers")
}
