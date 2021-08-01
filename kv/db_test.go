package kv

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/michael-diggin/yass/api"
	"github.com/stretchr/testify/require"
)

func TestKVDBSetAndGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "store-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	c := Config{}
	c.logConfig.Segment.MaxStoreBytes = 32
	db, err := NewDB(dir, c)
	require.NoError(t, err)

	key := "test-key"
	append := &api.Record{
		Id:    key,
		Value: []byte("hello world"),
	}
	err = db.Set(append)
	require.NoError(t, err)

	read, err := db.Get(key)
	require.NoError(t, err)
	require.Equal(t, append.Value, read.Value)

	err = db.Close()
	require.NoError(t, err)
}

func TestKVDBNotFound(t *testing.T) {
	dir, err := ioutil.TempDir("", "store-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	c := Config{}
	c.logConfig.Segment.MaxStoreBytes = 32
	db, err := NewDB(dir, c)
	require.NoError(t, err)

	key := "test-key"

	read, err := db.Get(key)
	require.Error(t, err)
	require.Nil(t, read)
	require.True(t, errors.As(err, &api.ErrNotFound{}))
}
