package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend/server"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var cache map[string]string

func main() {

	cache = make(map[string]string)

	port := flag.Int("p", 8080, "port for server to listen on")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("Could not listen on port %d: %v", *port, err)
	}

	s := grpc.NewServer()
	pb.RegisterCacheServer(s, server.New(cache))
	log.Printf("Starting server on port %d", *port)
	err = s.Serve(lis)
	if err != nil {
		logrus.Fatalf("Could not serve: %v", err)
	}
}
