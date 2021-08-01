package server

import (
	"context"
	"io/ioutil"
	"net"
	"testing"

	"github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/config"
	"github.com/michael-diggin/yass/kv"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
)

func TestSetAndGetFromServer(t *testing.T) {
	client, teardown := setupTest(t)
	defer teardown()

	ctx := context.Background()
	setRec := &api.Record{Id: "test-key", Value: []byte("hello world")}
	_, err := client.Set(ctx, &api.SetRequest{Record: setRec})
	require.NoError(t, err)

	resp, err := client.Get(ctx, &api.GetRequest{Id: setRec.Id})
	require.NoError(t, err)
	require.Equal(t, setRec.Value, resp.Record.Value)
}

func TestGetNotFoundFromServer(t *testing.T) {
	client, teardown := setupTest(t)
	defer teardown()

	ctx := context.Background()

	resp, err := client.Get(ctx, &api.GetRequest{Id: "test-key"})
	require.Error(t, err)
	require.Nil(t, resp)
	require.Equal(t, codes.NotFound, grpc.Code(err))
}

func setupTest(t *testing.T) (api.StorageClient, func()) {
	t.Helper()

	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	clientTLSConfig, err := config.SetUpTLSConfig(config.TLSConfig{
		CAFile:   config.CAFile,
		CertFile: config.ClientCertFile,
		KeyFile:  config.ClientKeyFile,
	})
	require.NoError(t, err)
	clientCreds := credentials.NewTLS(clientTLSConfig)

	cc, err := grpc.Dial(l.Addr().String(), grpc.WithTransportCredentials(clientCreds))
	require.NoError(t, err)

	serverTLSConfig, err := config.SetUpTLSConfig(config.TLSConfig{
		CAFile:        config.CAFile,
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		ServerAddress: l.Addr().String(),
		Server:        true,
	})
	require.NoError(t, err)
	serverCreds := credentials.NewTLS(serverTLSConfig)

	dir, err := ioutil.TempDir("", "server-test")
	require.NoError(t, err)
	db, err := kv.NewDB(dir, kv.Config{})
	require.NoError(t, err)

	cfg := &Config{DB: db}
	server, err := NewGRPCServer(cfg, grpc.Creds(serverCreds))
	require.NoError(t, err)
	go func() {
		server.Serve(l)
	}()

	return api.NewStorageClient(cc), func() {
		server.Stop()
		cc.Close()
		l.Close()
		db.Clear()
	}

}
