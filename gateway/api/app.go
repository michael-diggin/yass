package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// Gateway holds the router and the grpc clients
type Gateway struct {
	Router *mux.Router
	Client GrpcClient
}

// NewGateway will initialize the application
func NewGateway() *Gateway {
	g := Gateway{}
	g.Router = mux.NewRouter()
	g.initializeAPIRoutes()
	g.Client = nil
	return &g
}

func (g *Gateway) initializeAPIRoutes() {
	g.Router.HandleFunc("/get/{key}", g.Get).Methods("GET")
	g.Router.HandleFunc("/set", g.Set).Methods("POST")
	g.Router.HandleFunc("/delete/{key}", g.Delete).Methods("DELETE")
	g.Router.HandleFunc("/register", g.RegisterCacheServer).Methods("POST")
}

//Run will start the application
func (g *Gateway) Run(addr string) {
	logrus.Fatal(http.ListenAndServe(addr, g))
}

//ServeHTTP will serve and route a request
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	g.Router.ServeHTTP(w, r)
}
