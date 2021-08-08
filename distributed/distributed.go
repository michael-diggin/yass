package distributed

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb"
	"github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/kv"
	"google.golang.org/protobuf/proto"
)

type RequestType uint8

const (
	SetRequestType RequestType = iota
)

type Config struct {
	kvConfig kv.Config
	Raft     struct {
		raft.Config
		StreamLayer *StreamLayer
		Bootstrap   bool
	}
}

type YassDB struct {
	config Config
	db     *kv.DB
	raft   *raft.Raft
}

func NewYassDB(datadir string, config Config) (*YassDB, error) {
	ydb := &YassDB{config: config}
	if err := ydb.setUpDB(datadir); err != nil {
		return nil, err
	}
	if err := ydb.setUpRaft(datadir); err != nil {
		return nil, err
	}
	return ydb, nil
}

func (ydb *YassDB) setUpDB(datadir string) (err error) {
	plogDir := filepath.Join(datadir, "plog")
	if err := os.MkdirAll(plogDir, 0755); err != nil {
		return err
	}
	ydb.db, err = kv.NewDB(plogDir, ydb.config.kvConfig)
	return err
}

func (ydb *YassDB) setUpRaft(datadir string) (err error) {
	fsm := &fsm{db: ydb.db}

	logDir := filepath.Join(datadir, "raft")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return err
	}

	raftStore, err := raftboltdb.NewBoltStore(filepath.Join(datadir, "raft", "log"))
	if err != nil {
		return err
	}
	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(datadir, "raft", "stable"))
	if err != nil {
		return err
	}
	snapshotStore, err := raft.NewFileSnapshotStore(
		filepath.Join(datadir, "raft"), 1, os.Stderr)
	if err != nil {
		return err
	}
	maxPool := 5
	timeout := 10 * time.Second
	transport := raft.NewNetworkTransport(ydb.config.Raft.StreamLayer,
		maxPool, timeout, os.Stderr)

	config := raft.DefaultConfig()
	config.LocalID = ydb.config.Raft.LocalID
	if ydb.config.Raft.HeartbeatTimeout != 0 {
		config.HeartbeatTimeout = ydb.config.Raft.HeartbeatTimeout
	}
	if ydb.config.Raft.ElectionTimeout != 0 {
		config.ElectionTimeout = ydb.config.Raft.ElectionTimeout
	}
	if ydb.config.Raft.LeaderLeaseTimeout != 0 {
		config.LeaderLeaseTimeout = ydb.config.Raft.LeaderLeaseTimeout
	}
	if ydb.config.Raft.CommitTimeout != 0 {
		config.CommitTimeout = ydb.config.Raft.CommitTimeout
	}

	ydb.raft, err = raft.NewRaft(
		config,
		fsm,
		raftStore,
		stableStore,
		snapshotStore,
		transport,
	)
	if err != nil {
		return err
	}

	hasState, err := raft.HasExistingState(raftStore, stableStore, snapshotStore)
	if err != nil {
		return err
	}
	if ydb.config.Raft.Bootstrap && !hasState {
		config := raft.Configuration{
			Servers: []raft.Server{{ID: config.LocalID, Address: transport.LocalAddr()}},
		}
		err = ydb.raft.BootstrapCluster(config).Error()
	}

	return err
}

func (ydb *YassDB) Set(record *api.Record) error {
	_, err := ydb.Apply(SetRequestType, &api.SetRequest{Record: record})
	return err
}

func (ydb *YassDB) Get(id string) (*api.Record, error) {
	return ydb.db.Get(id)
}

func (ydb *YassDB) Apply(reqType RequestType, req proto.Message) (interface{}, error) {
	var buf bytes.Buffer
	_, err := buf.Write([]byte{byte(reqType)})
	if err != nil {
		return nil, err
	}
	b, err := proto.Marshal(req)
	if err != nil {
		return nil, err
	}
	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}

	future := ydb.raft.Apply(buf.Bytes(), 10*time.Second)
	if future.Error() != nil {
		return nil, future.Error()
	}
	res := future.Response()
	if err, ok := res.(error); ok {
		return nil, err
	}
	return res, nil
}

func (ydb *YassDB) Join(id, addr string) error {
	confFuture := ydb.raft.GetConfiguration()
	if err := confFuture.Error(); err != nil {
		return err
	}
	serverID := raft.ServerID(id)
	serverAddr := raft.ServerAddress(addr)
	for _, srv := range confFuture.Configuration().Servers {
		if srv.ID == serverID || srv.Address == serverAddr {
			return nil
		}
		removeFuture := ydb.raft.RemoveServer(serverID, 0, 0)
		if removeFuture.Error() != nil {
			return removeFuture.Error()
		}
	}
	addFuture := ydb.raft.AddVoter(serverID, serverAddr, 0, 0)
	if addFuture.Error() != nil {
		return addFuture.Error()
	}
	return nil
}

func (ydb *YassDB) Leave(id string) error {
	removeFuture := ydb.raft.RemoveServer(raft.ServerID(id), 0, 0)
	return removeFuture.Error()
}

func (ydb *YassDB) WaitForLeader(timeout time.Duration) error {
	timeoutCh := time.After(timeout)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-timeoutCh:
			return fmt.Errorf("timed out")
		case <-ticker.C:
			if ydb.raft.Leader() != "" {
				return nil
			}
		}
	}
}

func (ydb *YassDB) Close() error {
	f := ydb.raft.Shutdown()
	if f.Error() != nil {
		return f.Error()
	}
	return ydb.db.Close()
}
