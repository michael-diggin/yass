package main

import (
	"flag"

	"github.com/michael-diggin/yass/gateway/api"
	"github.com/sirupsen/logrus"
)

func main() {
	var port *string
	var numServers *int
	port = flag.String("p", ":8010", "port for server to listen on")
	numServers = flag.Int("s", 3, "the initial number of storage servers")
	flag.Parse()

	gateway := api.NewGateway(*numServers)

	//ctx, cancel := context.WithCancel(context.Background())
	// TODO: Add signal catching and graceful shutdown here
	logrus.Infof("Running the server on port %s", *port)
	gateway.Run(*port)
}
