package kv

import (
	"sync"

	"github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/log"
)

// DB is a struct containing the in memory KV store
// as well as the persistent log
type DB struct {
	data map[string]*api.Record
	mu   sync.RWMutex
	plog *log.Log
}

type Config struct {
	logConfig log.Config
}

func NewDB(dir string, c Config) (*DB, error) {
	store := make(map[string]*api.Record)
	plog, err := log.NewLog(dir, c.logConfig)
	if err != nil {
		return nil, err
	}
	// TODO: restore data if plog is not empty
	// the log.Reader method can be used to read all the contents
	// api.Record contains the `key` as well
	return &DB{data: store, mu: sync.RWMutex{}, plog: plog}, nil
}

func (db *DB) Set(record *api.Record) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	off, err := db.plog.Append(record)
	if err != nil {
		return err
	}
	record.Offset = off
	db.data[record.Key] = record
	return nil
}

func (db *DB) Get(key string) (*api.Record, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	record, ok := db.data[key]
	if !ok {
		return nil, ErrNotFound{Key: key}
	}
	return record, nil
}

func (db *DB) Close() error {
	if err := db.plog.Close(); err != nil {
		return err
	}
	db.data = nil
	return nil
}
