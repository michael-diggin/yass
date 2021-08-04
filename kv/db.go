package kv

import (
	"errors"
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
	plog, err := log.NewLog(dir, c.logConfig)
	if err != nil {
		return nil, err
	}

	store, err := resetOnStartUp(plog)
	if err != nil {
		return nil, err
	}
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
	db.data[record.Id] = record
	return nil
}

func (db *DB) Get(key string) (*api.Record, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	record, ok := db.data[key]
	if !ok {
		return nil, api.ErrNotFound{Key: key}
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

func (db *DB) Clear() error {
	if err := db.plog.Remove(); err != nil {
		return err
	}
	db.data = nil
	return nil
}

func resetOnStartUp(plog *log.Log) (map[string]*api.Record, error) {
	store := make(map[string]*api.Record)
	i := 0
	done := false
	for !done {
		rec, err := plog.Read(uint64(i))
		if err != nil {
			if errors.As(err, &api.ErrOffsetOutOfRange{}) {
				done = true
				continue
			} else {
				return nil, err
			}
		}
		store[rec.Id] = rec
		i++
	}
	return store, nil
}
