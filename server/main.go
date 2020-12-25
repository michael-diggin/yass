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

	if err := run(os.Args, os.Getenv); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

}

// Run is the entry point of the main function
func run(args []string, envFunc func(string) string) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	port := flags.Int("p", 8080, "port for cache server to listen on")
	gateway := flags.String("g", "localhost:8010", "location of the gateway server")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		return err
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

	// register cache server with the gateway
	localIP, err := GetLocalIP()
	if err != nil {
		logrus.Fatalf("Could not get local IP: %v", err)
	}

	addr := location{IP: localIP, Port: fmt.Sprintf("%d", *port)}
	payload, err := json.Marshal(addr)
	if err != nil {
		logrus.Fatalf("could not convert local address to json")
	}
	resp, err := http.Post("http://"+*gateway+"/register", "application/json", bytes.NewBuffer(payload))
	if err != nil || resp.StatusCode != http.StatusCreated {
		logrus.Fatalf("could not register server with gateway: %v", err)
	}

	wg.Wait()
	return nil
}

type location struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// GetLocalIP returns the non loopback local IP of the host
// Used for debugging in docker/wsl
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
