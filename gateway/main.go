package main

import (
	"flag"

	"github.com/michael-diggin/yass/gateway/api"
	"github.com/sirupsen/logrus"
)

func main() {
	var port *string
	port = flag.String("p", ":8010", "port for server to listen on")
	flag.Parse()

	gateway := api.NewGateway()

	//ctx, cancel := context.WithCancel(context.Background())
	// TODO: Add signal catching and graceful shutdown here
	logrus.Infof("Running the server on port %s", *port)
	gateway.Run(*port)
}
