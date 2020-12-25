package main

import (
	"context"
	"flag"
	"log"

	"github.com/michael-diggin/yass"
	"github.com/michael-diggin/yass/gateway/api"
	"github.com/sirupsen/logrus"
)

func main() {
	var port *string
	port = flag.String("p", ":8010", "port for server to listen on")
	cacheLocation := flag.String("l", "localhost:8080", "address of the cache server")
	flag.Parse()

	grpcClient, err := yass.NewClient(context.Background(), *cacheLocation)
	if err != nil {
		log.Fatal(err)
	}
	defer grpcClient.Close()

	gateway := api.NewGateway(grpcClient)

	//ctx, cancel := context.WithCancel(context.Background())
	// TODO: Add signal catching and graceful shutdown here
	logrus.Infof("Running the server on port %s", *port)
	gateway.Run(*port)
}
