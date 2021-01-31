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

	pb "github.com/michael-diggin/yass/proto"

	"github.com/michael-diggin/yass/common/retry"
	"github.com/michael-diggin/yass/server/core"
	"github.com/michael-diggin/yass/server/model"
	"github.com/michael-diggin/yass/server/storage"
	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"
)

func main() {
	port := flag.Int("p", 8080, "port for storage server to listen on")
	gateway := flag.String("g", "localhost:8010", "location of the watchtower")
	weights := flag.Int("w", 10, "The number of weights/data stores the server manages")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("Cannot listen on port: %v", err)
	}

	// set up cache
	stores := make([]model.Service, *weights)
	for i := 0; i < *weights; i++ {
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
		case err = <-srv.Serve():
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

	// adding itself as a node to the hash ring
	_, err = srv.AddNode(ctx, &pb.AddNodeRequest{Node: fmt.Sprintf("localhost:%d", *port)})
	if err != nil {
		logrus.Fatalf("could not add own node: %v", err)
	}

	localDNS := os.Getenv("POD_NAME")
	if localDNS == "" {
		logrus.Warning("No pod name specified, defaulting to localhost")
		localDNS = "localhost"
	}
	logrus.Info("Registering storage server with watchtower")
	err = retry.WithBackOff(func() error {
		conn, err := grpc.DialContext(ctx, *gateway, grpc.WithInsecure()) //TODO: add security and credentials
		if err != nil {
			return err
		}
		gatewayClient := pb.NewWatchTowerClient(conn)
		return srv.RegisterNodeWithWatchTower(gatewayClient, localDNS, *port)
	},
		5,
		1*time.Second,
		"register data server with watchtower",
	)
	if err != nil {
		logrus.Fatalf("could not register server with watchtower: %v", err)
	}

	wg.Wait()
}
