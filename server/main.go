package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/michael-diggin/yass/common/retry"
	"github.com/michael-diggin/yass/server/core"
	"github.com/michael-diggin/yass/server/model"
	"github.com/michael-diggin/yass/server/storage"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

func main() {
	port := flag.Int("p", 8080, "port for cache server to listen on")
	gateway := flag.String("g", "localhost:8010", "location of the gateway server")
	numStores := flag.Int("s", 10, "The number of data stores the server manages")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("Cannot listen on port: %v", err)
	}

	// set up cache
	stores := make([]model.Service, *numStores)
	for i := 0; i < *numStores; i++ {
		store := storage.New()
		stores[i] = store
	}

	srv := core.New(lis, stores...)
	defer srv.ShutDown()

	ctx, cancel := context.WithCancel(context.Background())
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		logrus.Infof("Starting cache server on port: %d", *port)
		select {
		case err = <-srv.Start():
			logrus.Errorf("error from server: %v", err)
		case <-ctx.Done():
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

	localIP, err := GetLocalIP()
	if err != nil {
		logrus.Fatal("Cannot get node IP")
	}
	logrus.Info("Registering cache server with api gateway")
	err = retry.WithBackOff(func() error {
		return srv.RegisterServerWithGateway(*gateway, localIP, *port)
	},
		5,
		1*time.Second,
		"register data server with gateway",
	)
	if err != nil {
		logrus.Fatalf("could not register server with gateway: %v", err)
	}

	wg.Wait()
}

// GetLocalIP returns the first non loopback local IP of the host
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", errors.Wrap(err, "failed to get local IP")
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback the display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", errors.New("No IP found")
}
