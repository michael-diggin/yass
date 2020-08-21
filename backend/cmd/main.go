package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/gomodule/redigo/redis"
	pb "github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/backend/server"
	"github.com/michael-diggin/yass/backend/storage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var cache map[string]string

func main() {

	port := flag.Int("p", 8080, "port for server to listen on")
	addr := flag.String("r", "localhost:6379", "address of redis cache")
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		logrus.Fatalf("Could not listen on port %d: %v", *port, err)
	}

	// set up redis
	conn, teardown := NewRedisConn(*addr)
	cache := storage.NewRedisService(conn)
	defer teardown()
	// set up grpc server
	s := grpc.NewServer()
	pb.RegisterCacheServer(s, server.New(cache))
	log.Printf("Starting server on port %d", *port)
	err = s.Serve(lis)
	if err != nil {
		logrus.Fatalf("Could not serve: %v", err)
	}
}

// NewRedisConn returns a new connection and a teardown function to close the conn
// addr default = localhost:6379
func NewRedisConn(addr string) (redis.Conn, func()) {
	conn, err := redis.Dial("tcp", addr)
	if err != nil {
		logrus.Fatal(err) // TODO: move this err handle up a level to main.go?
	}
	return conn, func() { conn.Close() }
}
