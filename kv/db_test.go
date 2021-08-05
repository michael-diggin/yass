package kv

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"

	"github.com/michael-diggin/yass/api"
	"github.com/michael-diggin/yass/log"
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

func TestResetOnStartUp(t *testing.T) {
	dir, err := ioutil.TempDir("", "store-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	c := log.Config{}
	c.Segment.MaxStoreBytes = 32
	plog, err := log.NewLog(dir, c)
	require.NoError(t, err)

	// append two records
	appendOne := &api.Record{
		Id:    "key-1",
		Value: []byte("hello world again"),
	}
	off, err := plog.Append(appendOne)
	require.NoError(t, err)
	require.Equal(t, uint64(0), off)

	appendTwo := &api.Record{
		Id:    "key-2",
		Value: []byte("hello world again"),
	}
	off, err = plog.Append(appendTwo)
	require.NoError(t, err)
	require.Equal(t, uint64(1), off)

	data, err := resetOnStartUp(plog)
	require.NoError(t, err)

	require.Len(t, data, 2)
	require.Equal(t, appendOne.Value, data[appendOne.Id].Value)
	require.Equal(t, appendTwo.Value, data[appendTwo.Id].Value)

}
