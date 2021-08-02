package agent

import (
	"crypto/tls"
	"fmt"
	"net"
	"sync"

	"github.com/michael-diggin/yass/discovery"
	"github.com/michael-diggin/yass/kv"
	"github.com/michael-diggin/yass/server"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Agent struct {
	Config

	db         *kv.DB
	server     *grpc.Server
	membership *discovery.Membership

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
		a.setupDB,
		a.setupServer,
		a.setupMembership,
	}

	for _, fn := range setup {
		if err := fn(); err != nil {
			return nil, err
		}
	}
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

func (a *Agent) setupDB() (err error) {
	a.db, err = kv.NewDB(a.Config.DataDir, kv.Config{})
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
	rpcAddr, err := a.RPCAddr()
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}
	go func() {
		if err := a.server.Serve(ln); err != nil {
			a.Shutdown()
		}
	}()
	return err
}

func (a *Agent) setupMembership() (err error) {
	rpcAddr, err := a.RPCAddr()
	if err != nil {
		return err
	}
	a.membership, err = discovery.New(discovery.UnimplementedHandler{},
		discovery.Config{
			NodeName:       a.Config.NodeName,
			BindAddr:       a.Config.BindAddr,
			Tags:           map[string]string{"rpc_addr": rpcAddr},
			StartJoinAddrs: a.Config.StartJoinAddrs,
		},
	)
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
