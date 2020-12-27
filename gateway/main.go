package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/michael-diggin/yass/gateway/api"
	"github.com/sirupsen/logrus"
)

func main() {
	var port *int
	var numServers *int
	port = flag.Int("p", 8010, "port for server to listen on")
	numServers = flag.Int("s", 3, "the initial number of storage servers")
	flag.Parse()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	gateway := api.NewGateway(*numServers, srv)

	defer gateway.Stop()

	go func() {
		logrus.Infof("Running the server on port %d", *port)
		err := gateway.ListenAndServe()
		if err != nil {
			logrus.Errorf("encountered error when running gateway: %v", err)
		}
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		// Check for a closing signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		logrus.Infof("OS signal caught: %+v", sig)
		wg.Done()
	}()

	wg.Wait() // wait for sever go routine to exit
}
