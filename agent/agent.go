package agent

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/michael-diggin/yass/discovery"
	"github.com/michael-diggin/yass/distributed"
	"github.com/michael-diggin/yass/server"
	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Agent struct {
	Config

	db         *distributed.YassDB
	server     *grpc.Server
	membership *discovery.Membership
	mux        cmux.CMux

	shutdown     bool
	shutdowns    chan struct{}
	shutdownLock sync.Mutex
}

type Config struct {
	ServerTLSConfig *tls.Config
	PeerTLSConfig   *tls.Config
	DataDir         string
	BindAddr        string
	RPCPort         int
	NodeName        string
	StartJoinAddrs  []string
	Bootstrap       bool
}

func (c Config) RPCAddr() (string, error) {
	host, _, err := net.SplitHostPort(c.BindAddr)
	if err != nil {
		return "", nil
	}
	return fmt.Sprintf("%s:%d", host, c.RPCPort), nil
}

func New(config Config) (*Agent, error) {
	a := &Agent{
		Config:    config,
		shutdowns: make(chan struct{}),
	}

	setup := []func() error{
		a.setupLogger,
		a.setupMux,
		a.setupDB,
		a.setupServer,
		a.setupMembership,
	}

	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}
	go a.serve()
	return a, nil
}

func (a *Agent) setupLogger() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}

func (a *Agent) setupMux() error {
	rpcAddr := fmt.Sprintf(":%d", a.Config.RPCPort)
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}
	a.mux = cmux.New(ln)
	return nil
}

func (a *Agent) setupDB() (err error) {
	raftLn := a.mux.Match(func(reader io.Reader) bool {
		b := make([]byte, 1)
		if _, err := reader.Read(b); err != nil {
			return false
		}
		return bytes.Compare(b, []byte{byte(distributed.RaftRPC)}) == 0
	})

	conf := distributed.Config{}
	conf.Raft.StreamLayer = distributed.NewStreamLayer(
		raftLn, a.Config.ServerTLSConfig, a.Config.PeerTLSConfig,
	)
	conf.Raft.LocalID = raft.ServerID(a.Config.NodeName)
	conf.Raft.Bootstrap = a.Config.Bootstrap

	a.db, err = distributed.NewYassDB(a.Config.DataDir, conf)
	if err != nil {
		return err
	}
	if a.Config.Bootstrap {
		err = a.db.WaitForLeader(3 * time.Second)
	}
	return err
}

func (a *Agent) setupServer() (err error) {
	serverConfig := &server.Config{
		DB: a.db,
	}
	var opts []grpc.ServerOption
	if a.Config.ServerTLSConfig != nil {
		creds := credentials.NewTLS(a.Config.ServerTLSConfig)
		opts = append(opts, grpc.Creds(creds))
	}
	a.server, err = server.NewGRPCServer(serverConfig, opts...)
	if err != nil {
		return err
	}
	grpcLn := a.mux.Match(cmux.Any())
	go func() {
		if err := a.server.Serve(grpcLn); err != nil {
			a.Shutdown()
		}
	}()
	return nil
}

func (a *Agent) setupMembership() (err error) {
	rpcAddr, err := a.RPCAddr()
	if err != nil {
		return err
	}
	a.membership, err = discovery.New(a.db,
		discovery.Config{
			NodeName:       a.Config.NodeName,
			BindAddr:       a.Config.BindAddr,
			Tags:           map[string]string{"rpc_addr": rpcAddr},
			StartJoinAddrs: a.Config.StartJoinAddrs,
		},
	)
	return err
}

func (a *Agent) serve() error {
	if err := a.mux.Serve(); err != nil {
		a.Shutdown()
		return err
	}
	return nil
}

func (a *Agent) Shutdown() error {
	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()

	if a.shutdown {
		return nil
	}
	a.shutdown = true
	close(a.shutdowns)

	shutdown := []func() error{
		a.membership.Leave,
		func() error {
			a.server.GracefulStop()
			return nil
		},
		a.db.Close,
	}

	for _, fn := range shutdown {
		if err := fn(); err != nil {
			return err
		}
	}
	return nil
}
