package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/michael-diggin/yass/server/pkg/core"
	"github.com/michael-diggin/yass/server/pkg/storage"
	"github.com/pkg/errors"

	"github.com/sirupsen/logrus"
)

func main() {
	port := flag.Int("p", 8080, "port for cache server to listen on")
	gateway := flag.String("g", "localhost:8010", "location of the gateway server")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("Cannot listen on port: %v", err)
	}

	// set up cache
	cache := storage.New()

	srv := core.New(lis, cache)
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

	logrus.Info("Registering cache server with api gateway")
	if err := RegisterServerWithGateway(*gateway, *port); err != nil {
		logrus.Fatalf("could not register server with gateway: %v", err)
	}

	wg.Wait()
}

// RegisterServerWithGateway will register the cache server with the api gateway so it can accept requests
func RegisterServerWithGateway(gateway string, port int) error {
	localIP, err := GetLocalIP()
	if err != nil {
		return err
	}

	addr := location{IP: localIP, Port: fmt.Sprintf("%d", port)}
	payload, err := json.Marshal(addr)
	if err != nil {
		return err
	}
	resp, err := http.Post("http://"+gateway+"/register", "application/json", bytes.NewBuffer(payload))
	if err != nil || resp.StatusCode != http.StatusCreated {
		return err
	}
	return nil
}

type location struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
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
