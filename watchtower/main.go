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

	"github.com/michael-diggin/yass/common/client"
	"github.com/michael-diggin/yass/watchtower/api"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

func main() {
	var port *int
	var numServers *int
	var weight *int
	port = flag.Int("p", 8010, "port for server to listen on")
	numServers = flag.Int("s", 3, "the initial number of storage servers")
	weight = flag.Int("w", 10, "the number of virtual nodes for each node on the hash ring")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("Cannot listen on port: %v", err)
	}

	srv := grpc.NewServer()
	wt := api.NewWatchTower(*numServers, *weight, client.Factory{})
	pb.RegisterWatchTowerServer(srv, wt)
	grpc_health_v1.RegisterHealthServer(srv, wt)

	defer func() {
		srv.GracefulStop()
		logrus.Info("server stopped")
		wt.Stop()
	}()

	go func() {
		logrus.Infof("Running watchtower on port %d", *port)
		err = srv.Serve(lis)
		if err != nil {
			logrus.Errorf("encountered error when running watchtower: %v", err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go wt.PingStorageServers(ctx, 30*time.Second)

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

	wg.Wait()
}
