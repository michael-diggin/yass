package core

import (
	"net"
	"sync"

	"github.com/michael-diggin/yass/common/client"
	"github.com/michael-diggin/yass/common/hashring"
	"github.com/michael-diggin/yass/common/models"
	"github.com/michael-diggin/yass/common/xid"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// YassServer wraps up the listener, grpc server and cache service
type YassServer struct {
	lis net.Listener
	srv *grpc.Server
	*server
}

// New sets up the server
func New(lis net.Listener, name, leader string, dataStores ...model.Service) YassServer {
	s := grpc.NewServer()
	srv := newServer(client.Factory{}, name, leader, dataStores...)
	pb.RegisterStorageServer(s, srv)
	pb.RegisterYassServiceServer(s, srv)
	grpc_health_v1.RegisterHealthServer(s, srv)
	return YassServer{lis: lis, srv: s, server: srv}
}

// Serve starts the grpc server
func (y YassServer) Serve() <-chan error {
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
	y.srv.GracefulStop()
	logrus.Infof("gRPC server stopped")
}

// server (unexported) implements the StorageServer interface
type server struct {
	DataStores  []model.Service
	factory     models.ClientFactory
	nodeClients map[string]*models.StorageClient
	mu          sync.RWMutex
	hashRing    models.HashRing
	minServers  int
	Name        string
	RaftLeader  string
	IDStore     *xid.IDStore
}

func newServer(factory models.ClientFactory, name, leader string, dataStores ...model.Service) *server {
	hashRing := hashring.New(len(dataStores))
	srv := server{
		DataStores:  dataStores,
		factory:     factory,
		nodeClients: make(map[string]*models.StorageClient),
		mu:          sync.RWMutex{},
		hashRing:    hashRing,
		minServers:  3,
		Name:        name,
		RaftLeader:  leader,
		IDStore:     xid.New(uint64(0)),
	}
	return &srv
}

func (srv *server) IsLeader() bool {
	return srv.Name == srv.RaftLeader
}
