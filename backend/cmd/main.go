package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/michael-diggin/yass/backend/pkg/redis"

	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend/pkg/server"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
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

	s := grpc.NewServer()
	pb.RegisterCacheServer(s, server.New(cache))

	errChan := make(chan error)

	go func() {
		log.Printf("Starting server on port %d", *port)
		err = s.Serve(lis)
		if err != nil {
			errChan <- err
		}
	}()

	// Check for a closing signal
	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		logrus.Infof("Caught interrupt signal: %+v", sig)
		logrus.Infof("Gracefully shutting down server resources...")
	case err = <-errChan:
		logrus.Errorf("Server error: %v", err)
	}

	cache.Close()
	logrus.Infof("Redis connection closed")
	s.GracefulStop()
	logrus.Infof("gRPC server stopped")
}
