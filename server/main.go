package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	pb "github.com/michael-diggin/yass/api"
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
	pb.RegisterCacheServer(s, server{})
	log.Printf("Starting server on port %d", *port)
	err = s.Serve(lis)
	if err != nil {
		logrus.Fatalf("Could not serve: %v", err)
	}
}

type server struct{}

func (s server) Add(ctx context.Context, pair *pb.Pair) (*pb.Key, error) {
	_, ok := cache[pair.Key]
	if ok {
		logrus.Errorf("Tried to reset key: %s", pair.Key)
		return nil, fmt.Errorf("Key is already set")
	}
	cache[pair.Key] = pair.Value
	output := &pb.Key{Key: pair.Key}
	return output, nil
}

func (s server) Get(ctx context.Context, key *pb.Key) (*pb.Pair, error) {
	res, ok := cache[key.Key]
	if !ok {
		logrus.Errorf("Tried to get value with not set key: %s", key.Key)
		return nil, fmt.Errorf("No value stored for key %s", key.Key)
	}
	pair := &pb.Pair{Key: key.Key, Value: res}
	return pair, nil
}
