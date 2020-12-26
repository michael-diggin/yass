package api

import (
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Gateway holds the router and the grpc clients
type Gateway struct {
	Router     *mux.Router
	Clients    map[int]GrpcClient
	numServers int
	mu         sync.RWMutex
}

// NewGateway will initialize the application
func NewGateway(numServers int) *Gateway {
	g := Gateway{}
	g.Router = mux.NewRouter()
	g.initializeAPIRoutes()
	g.Clients = make(map[int]GrpcClient)
	g.numServers = numServers
	g.mu = sync.RWMutex{}
	return &g
}

func (g *Gateway) initializeAPIRoutes() {
	g.Router.HandleFunc("/get/{key}", g.Get).Methods("GET")
	g.Router.HandleFunc("/set", g.Set).Methods("POST")
	g.Router.HandleFunc("/delete/{key}", g.Delete).Methods("DELETE")
	g.Router.HandleFunc("/register", g.RegisterCacheServer).Methods("POST")
}

//Serve will start the application
func (g *Gateway) Serve(addr string) <-chan error {
	errChan := make(chan error, 1)
	errChan <- http.ListenAndServe(addr, g)
	return errChan
}

//ServeHTTP will serve and route a request
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.Router.ServeHTTP(w, r)
}

// ShutDown will close all grpc connections the gateway holds
func (g *Gateway) ShutDown() {
	logrus.Infof("Closing connections to storage servers")
	g.mu.Lock()
	for num, client := range g.Clients {
		serverNum := num
		client.Close()
		delete(g.Clients, serverNum)
	}
}
