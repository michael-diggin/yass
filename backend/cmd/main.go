package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend"
	"github.com/michael-diggin/yass/backend/pkg/redis"
	"github.com/michael-diggin/yass/backend/pkg/server"
	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"
)

func main() {

	port := flag.Int("p", 8080, "port for server to listen on")
	addr := flag.String("r", "localhost:6379", "address of redis cache")
	flag.Parse()

	// set up redis
	username := os.Getenv("REDIS_USER") //defaults to ""
	password := os.Getenv("REDIS_PASS") //defaults to ""
	cache, err := redis.New(username, password, *addr)
	if err != nil {
		logrus.Fatalf("Could not connect to redis cache: %v", err)
	}

	// set up listener and grpc server

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("Could not listen on port %d: %v", *port, err)
	}

	srv := YassServer{}
	srv.Init(lis, cache)
	defer srv.ShutDown()

	// Check for a closing signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	logrus.Infof("Starting server on port: %d", *port)
	select {
	case err = <-srv.SpinUp():
		logrus.Errorf("Server error: %v", err)
	case sig := <-sigChan:
		logrus.Infof("Caught interrupt signal: %+v", sig)
	}
}

// YassServer wraps up the listener, grpc server and cache service
type YassServer struct {
	lis   net.Listener
	srv   *grpc.Server
	cache backend.Service
}

// Init sets up the server
func (y *YassServer) Init(lis net.Listener, cache backend.Service) {
	s := grpc.NewServer()
	pb.RegisterCacheServer(s, server.New(cache))
	y.lis = lis
	y.srv = s
	y.cache = cache
}

// SpinUp starts the grpc server
func (y YassServer) SpinUp() <-chan error {
	errChan := make(chan error)
	go func() {
		err := y.srv.Serve(y.lis)
		if err != nil {
			errChan <- err
			close(errChan)
		}
	}()
	return errChan
}

// ShutDown closes the server resources
func (y YassServer) ShutDown() {
	logrus.Infof("Shutting down server resources")
	y.cache.Close()
	logrus.Infof("Redis connection closed")
	y.srv.GracefulStop()
	logrus.Infof("gRPC server stopped")
}
