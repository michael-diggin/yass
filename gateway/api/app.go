package api

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/michael-diggin/yass"
)

// Gateway holds the router and the grpc clients
type Gateway struct {
	Router *mux.Router
	Client *yass.CacheClient
}

// NewGateway will initialize the application
func NewGateway(grpcClient *yass.CacheClient) *Gateway {
	g := Gateway{}
	g.Router = mux.NewRouter()
	g.initializeRoutes()
	g.Client = grpcClient
	return &g
}

func (g *Gateway) initializeRoutes() {
	g.Router.HandleFunc("/get/{key}", g.Get).Methods("GET")
	g.Router.HandleFunc("/set", g.Set).Methods("POST")
	g.Router.HandleFunc("/delete/{key}", g.Delete).Methods("DELETE")
}

//Run will start the application
func (g *Gateway) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, g.Router))
}

//ServeHTTP will serve and route a request
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.Router.ServeHTTP(w, r)
}
