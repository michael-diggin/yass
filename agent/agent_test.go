package agent

import (
	"context"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	"github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/config"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func TestAgent(t *testing.T) {
	serverTLSConfig, err := config.SetUpTLSConfig(config.TLSConfig{
		CertFile:      config.ServerCertFile,
		KeyFile:       config.ServerKeyFile,
		CAFile:        config.CAFile,
		Server:        true,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)

	peerTLSConfig, err := config.SetUpTLSConfig(config.TLSConfig{
		CertFile:      config.ClientCertFile,
		KeyFile:       config.ClientKeyFile,
		CAFile:        config.CAFile,
		Server:        false,
		ServerAddress: "127.0.0.1",
	})
	require.NoError(t, err)

	var agents []*Agent
	for i := 0; i < 3; i++ {
		ports := []int{getFreePort(), getFreePort()}
		bindAddr := fmt.Sprintf("%s:%d", "127.0.0.1", ports[0])
		rpcPort := ports[1]
		datadir, err := ioutil.TempDir("", "agent-test-plog")
		require.NoError(t, err)

		var startJoinAddrs []string
		if i > 0 {
			startJoinAddrs = append(startJoinAddrs, agents[0].Config.BindAddr)
		}
		agent, err := New(Config{
			NodeName:        fmt.Sprintf("%d", i),
			StartJoinAddrs:  startJoinAddrs,
			BindAddr:        bindAddr,
			RPCPort:         rpcPort,
			DataDir:         datadir,
			ServerTLSConfig: serverTLSConfig,
			PeerTLSConfig:   peerTLSConfig,
		})
		require.NoError(t, err)
		agents = append(agents, agent)
	}

	defer func() {
		for _, agent := range agents {
			err := agent.Shutdown()
			require.NoError(t, err)
			require.NoError(t, os.RemoveAll(agent.Config.DataDir))
		}
	}()

	time.Sleep(3 * time.Second)

	leaderClient := client(t, agents[0], peerTLSConfig)
	_, err = leaderClient.Set(
		context.Background(),
		&api.SetRequest{
			Record: &api.Record{Id: "test-key", Value: []byte("hello world")},
		},
	)
	require.NoError(t, err)
	getResp, err := leaderClient.Get(
		context.Background(),
		&api.GetRequest{Id: "test-key"},
	)
	require.NoError(t, err)
	require.Equal(t, []byte("hello world"), getResp.Record.Value)

}

func client(t *testing.T, agent *Agent, tlsConfig *tls.Config) api.StorageClient {
	tlsCreds := credentials.NewTLS(tlsConfig)
	rpcAddr, err := agent.RPCAddr()
	require.NoError(t, err)
	conn, err := grpc.Dial(rpcAddr, grpc.WithTransportCredentials(tlsCreds))
	require.NoError(t, err)
	return api.NewStorageClient(conn)
}

func getFreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}
