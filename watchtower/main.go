package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/michael-diggin/yass/common/client"
	"github.com/michael-diggin/yass/watchtower/api"
	"github.com/sirupsen/logrus"
)

func main() {
	var port *int
	var numServers *int
	var weight *int
	port = flag.Int("p", 8010, "port for server to listen on")
	numServers = flag.Int("s", 3, "the initial number of storage servers")
	weight = flag.Int("w", 10, "the number of virtual nodes for each node on the hash ring")
	flag.Parse()

	gateway := api.NewGateway(*numServers, *weight, client.Factory{})

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      gateway,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	defer func() {
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctxShutDown); err != nil {
			logrus.Fatalf("server shutdown failed:%v", err)
		}
		logrus.Info("server stopped")
		gateway.Stop()
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		logrus.Infof("Running the server on port %d", *port)
		err := srv.ListenAndServe()
		switch err {
		case nil:
			return
		case http.ErrServerClosed:
			logrus.Info("Server closed")
			return
		default:
			logrus.Errorf("encountered error when running gateway: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go gateway.PingStorageServers(ctx, 30*time.Second)

	go func() {
		// Check for a closing signal
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan
		logrus.Infof("OS signal caught: %+v", sig)
		wg.Done()
	}()

	wg.Wait() // wait for server go routine to exit
}
