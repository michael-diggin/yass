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

// Opts gives the user the chance to alter the parameters
type Opts struct {
	hashRingWeights int
	replicaCount    int
}

var defaultOpts = Opts{hashRingWeights: 3, replicaCount: 2}

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
func NewGateway(numServers int, srv *http.Server, opts ...Opts) *Gateway {
	var opt = defaultOpts
	if len(opts) == 1 {
		opt = opts[0]
	}
	g := Gateway{}

	router := mux.NewRouter()
	router.HandleFunc("/get/{key}", g.Get).Methods("GET")
	router.HandleFunc("/set", g.Set).Methods("POST")
	router.HandleFunc("/delete/{key}", g.Delete).Methods("DELETE")
	router.HandleFunc("/register", g.RegisterCacheServer).Methods("POST")

	g.Server = srv

	g.Handler = router

	g.Clients = make(map[string]GrpcClient)
	g.numServers = numServers
	g.mu = sync.RWMutex{}
	g.hashRing = hashring.New(opt.hashRingWeights)
	g.replicas = numServers
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
