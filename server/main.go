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

	"github.com/michael-diggin/yass/common/retry"
	"github.com/michael-diggin/yass/server/core"
	"github.com/michael-diggin/yass/server/model"
	"github.com/michael-diggin/yass/server/storage"
	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"
)

func main() {
	port := flag.Int("p", 8080, "port for storage server to listen on")
	cluster := flag.String("join", "yass-0,yass-1,yass-2", "nodes in the cluster")
	weights := flag.Int("w", 10, "The number of weights/data stores the server manages")
	loglevel := flag.String("v", "info", "the logging level verbosity")
	flag.Parse()

	level, err := parseLogLevel(*loglevel)
	if err != nil {
		logrus.Warningf("Could not parse log level %s, defaulting to info", *loglevel)
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

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

	podName := os.Getenv("POD_NAME")
	if podName == "" {
		logrus.Fatalf("No pod name specified, cannot join cluster")
	}
	logrus.Infof("POD_NAME is %s", podName)

	// adding itself as a node to the hash ring on
	addReq := &pb.AddNodeRequest{Node: fmt.Sprintf("%s:%d", podName, *port)}
	_, err = srv.AddNode(ctx, addReq)
	if err != nil {
		logrus.Fatalf("could not add own node: %v", err)
	}

	allNodes := strings.Split(*cluster, ",")
	newNodes := []string{}
	for _, node := range allNodes {
		if strings.Contains(node, podName) {
			// this node is already added
			continue
		}
		newNodes = append(newNodes, node)

		logrus.Infof("Registering server with node %s", node)
		err = retry.WithBackOff(func() error {
			conn, err := grpc.DialContext(ctx, node, grpc.WithInsecure()) //TODO: add security and credentials
			if err != nil {
				return err
			}
			client := pb.NewStorageClient(conn)
			_, err = client.AddNode(ctx, addReq)
			return err
		},
			5,
			1*time.Second,
			"register data server with other nodes in cluster",
		)
		if err != nil {
			logrus.Fatalf("could not register server with node %s: %v", node, err)
		}
	}

	if false {
		// the AddNode response needs to be altered
		// use the response to decide if the node should request data from other nodes
		srv.RepopulateFromNodes(newNodes...)
	}

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
