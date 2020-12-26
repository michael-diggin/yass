package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

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

	defer gateway.ShutDown()

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		logrus.Infof("Running the server on port %s", *port)
		select {
		case <-ctx.Done():
		case err := <-gateway.Serve(*port):
			logrus.Errorf("encountered error when running gateway: %v", err)
		}
		wg.Done()
	}()

	go func() {
		// Check for a closing signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-ctx.Done():
		case sig := <-sigChan:
			logrus.Infof("OS signal caught: %+v", sig)
			cancel()
			return
		}
	}()

	wg.Wait()
}
