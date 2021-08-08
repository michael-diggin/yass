package distributed

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/raft"
	"github.com/michael-diggin/yass/api"
	"github.com/stretchr/testify/require"
)

func TestMultipleNodes(t *testing.T) {
	var dbs []*YassDB
	nodeCount := 3

	for i := 0; i < nodeCount; i++ {
		datadir, err := ioutil.TempDir("", "distributed-test")
		require.NoError(t, err)
		defer func(dir string) {
			os.RemoveAll(dir)
		}(datadir)

		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", getFreePort()))
		require.NoError(t, err)

		config := Config{}
		config.Raft.StreamLayer = NewStreamLayer(ln, nil, nil)
		config.Raft.LocalID = raft.ServerID(fmt.Sprintf("%d", i))
		config.Raft.HeartbeatTimeout = 50 * time.Millisecond
		config.Raft.ElectionTimeout = 50 * time.Millisecond
		config.Raft.LeaderLeaseTimeout = 50 * time.Millisecond
		config.Raft.CommitTimeout = 5 * time.Millisecond
		config.Raft.Bootstrap = (i == 0)

		db, err := NewYassDB(datadir, config)
		require.NoError(t, err, "failed on %d", i)
		if i != 0 {
			err = dbs[0].Join(fmt.Sprintf("%d", i), ln.Addr().String())
			require.NoError(t, err)
		} else {
			err = db.WaitForLeader(3 * time.Second)
			require.NoError(t, err)
		}
		dbs = append(dbs, db)
	}

	records := []*api.Record{
		{Id: "rec-1", Value: []byte("first")},
		{Id: "rec-2", Value: []byte("second")},
	}
	for _, record := range records {
		err := dbs[0].Set(record)
		require.NoError(t, err)
	}

	for _, record := range records {
		require.Eventually(t, func() bool {
			for j := 0; j < len(dbs); j++ {
				rec, err := dbs[j].Get(record.Id)
				if err != nil {
					return false
				}
				if !reflect.DeepEqual(rec.Value, record.Value) {
					return false
				}
			}
			return true
		}, 500*time.Millisecond, 50*time.Millisecond)
	}

	err := dbs[0].Leave("1")
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	err = dbs[0].Set(&api.Record{Id: "rec-3", Value: []byte("third")})
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	rc, err := dbs[1].Get("rec-3")
	require.Error(t, err)
	require.IsType(t, api.ErrNotFound{}, err)
	require.Nil(t, rc)

	rc, err = dbs[2].Get("rec-3")
	require.NoError(t, err)
	require.Equal(t, []byte("third"), rc.Value)
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
