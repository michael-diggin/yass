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

	if err := run(os.Args, os.Getenv); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

}

// Run is the entry point of the main function
func run(args []string, envFunc func(string) string) error {
	flags := flag.NewFlagSet(args[0], flag.ExitOnError)
	port := flags.Int("p", 8080, "port for server to listen on")
	addr := flags.String("r", "localhost:6379", "address of redis cache")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		return err
	}

	// set up redis
	username := envFunc("REDIS_USER") //defaults to ""
	password := envFunc("REDIS_PASS") //defaults to ""
	cache, err := redis.New(username, password, *addr)
	if err != nil {
		return err
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
		return err
	case sig := <-sigChan:
		logrus.Infof("OS signal caught: %+v", sig)
		return nil
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
