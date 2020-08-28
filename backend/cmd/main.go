package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend/server"
	"github.com/michael-diggin/yass/backend/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {

	port := flag.Int("p", 8080, "port for server to listen on")
	addr := flag.String("r", "localhost:6379", "address of redis cache")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("Could not listen on port %d: %v", *port, err)
	}

	// set up redis
	username := os.Getenv("REDIS_USER") //defaults to ""
	password := os.Getenv("REDIS_PASS") //defaults to ""
	cache, err := storage.NewRedisService(username, password, *addr)
	if err != nil {
		logrus.Fatalf("Could not connect to redis cache: %v", err)
	}
	defer cache.Close()
	// set up grpc server
	s := grpc.NewServer()
	pb.RegisterCacheServer(s, server.New(cache))
	log.Printf("Starting server on port %d", *port)
	err = s.Serve(lis)
	if err != nil {
		logrus.Fatalf("Could not serve: %v", err)
	}
}
