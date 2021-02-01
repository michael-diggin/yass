package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
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
	var fileName *string
	var loglevel *string
	port = flag.Int("p", 8010, "port for server to listen on")
	numServers = flag.Int("s", 3, "the initial number of storage servers")
	weight = flag.Int("w", 10, "the number of virtual nodes for each node on the hash ring")
	fileName = flag.String("f", "/usr/yass/node_data", "location of the file for storing node addresses")
	loglevel = flag.String("v", "info", "the logging level verbosity")
	flag.Parse()

	level, err := parseLogLevel(*loglevel)
	if err != nil {
		logrus.Warningf("Could not parse log level %s, defaulting to info", *loglevel)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	wt := api.NewWatchTower(*numServers, *weight, client.Factory{}, *fileName)

	err = wt.LoadData()
	if err != nil {
		logrus.Fatalf("failed to load node data from file: %v", err)
	}

	srv := grpc.NewServer()

	pb.RegisterWatchTowerServer(srv, wt)
	grpc_health_v1.RegisterHealthServer(srv, wt)

	defer func() {
		srv.GracefulStop()
		logrus.Info("server stopped")
		wt.Stop()
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("Cannot listen on port: %v", err)
	}

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

func parseLogLevel(level string) (logrus.Level, error) {
	level = strings.ToLower(level)
	var lev logrus.Level
	switch level {
	case "trace":
		lev = logrus.TraceLevel
	case "debug":
		lev = logrus.DebugLevel
	case "info":
		lev = logrus.InfoLevel
	case "warning":
		lev = logrus.WarnLevel
	case "error":
		lev = logrus.ErrorLevel
	case "fatal":
		lev = logrus.FatalLevel
	default:
		return logrus.InfoLevel, fmt.Errorf("cannot parse level %s", level)
	}

	return lev, nil
}
