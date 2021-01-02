package storage

import (
	"testing"

	"github.com/michael-diggin/yass/server/errors"
	"github.com/michael-diggin/yass/server/model"
	"github.com/stretchr/testify/require"
)

func TestPingStorage(t *testing.T) {
	tt := []struct {
		name string
		ser  *Service
		err  error
	}{
		{"serving", New(), nil},
		{"not-serving", &Service{}, errors.NotServing{}},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			defer tc.ser.Close()
			err := tc.ser.Ping()
			require.Equal(t, tc.err, err)
		})
	}
}

func TestSetInCache(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()
	_ = <-ser.Set("test-key", uint32(100), "test-value")

	tt := []struct {
		name  string
		key   string
		value interface{}
		hash  uint32
		err   error
	}{
		{"valid", "key", "value", uint32(77), nil},
		{"already set", "test-key", "test-value", uint32(100), errors.AlreadySet{Key: "test-key"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			resp := <-ser.Set(tc.key, tc.hash, tc.value)
			r.Equal(tc.err, resp.Err)
			r.Equal(tc.key, resp.Key)
		})
	}
}

func TestGetFromCache(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()
	_ = <-ser.Set("test-key", uint32(100), "test-value")

	tt := []struct {
		name  string
		key   string
		value interface{}
		err   error
	}{
		{"valid", "test-key", "test-value", nil},
		{"not found", "key", nil, errors.NotFound{Key: "key"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp := <-ser.Get(tc.key)
			r.Equal(tc.err, resp.Err)
			r.Equal(tc.value, resp.Value)
		})
	}

}
func TestDelFromCache(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()
	_ = <-ser.Set("test-key", uint32(100), "test-value")

	tt := []struct {
		name string
		key  string
		err  error
	}{
		{"valid", "test-key", nil},
		{"not in cache", "key", nil},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			resp := <-ser.Delete(tc.key)
			r.Equal(tc.err, resp.Err)
		})
	}
}

func TestBatchGet(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()
	ser.store["test-key-1"] = model.Data{Value: "value-1", Hash: uint32(100)}
	ser.store["test-key-2"] = model.Data{Value: 2, Hash: uint32(201)}

	t.Run("get all", func(t *testing.T) {
		data := <-ser.BatchGet(uint32(0), uint32(1000))
		r.Len(data, 2)
		r.Equal(data["test-key-1"], model.Data{Value: "value-1", Hash: uint32(100)})
		r.Equal(data["test-key-2"], model.Data{Value: 2, Hash: uint32(201)})
	})

	t.Run("filters correctly", func(t *testing.T) {
		data := <-ser.BatchGet(uint32(0), uint32(150))
		r.Len(data, 1)
		r.Equal(data["test-key-1"], model.Data{Value: "value-1", Hash: uint32(100)})
	})

	t.Run("wrap around case", func(t *testing.T) {
		// this is the edge case where high < low due to wrapping around the hash ring
		data := <-ser.BatchGet(uint32(150), uint32(50))
		r.Len(data, 1)
		r.Equal(data["test-key-2"], model.Data{Value: 2, Hash: uint32(201)})
	})

}

func TestBatchSet(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()

	data := make(map[string]model.Data)
	data["test-key-1"] = model.Data{Value: "value-1", Hash: uint32(100)}
	data["test-key-2"] = model.Data{Value: 2, Hash: uint32(101)}

	err := <-ser.BatchSet(data)
	r.NoError(err)
	r.Len(ser.store, 2)
	r.Equal(ser.store["test-key-1"], model.Data{Value: "value-1", Hash: uint32(100)})
	r.Equal(ser.store["test-key-2"], model.Data{Value: 2, Hash: uint32(101)})
}

func TestBatchDelete(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()

	ser.store["test-key-1"] = model.Data{Value: "value-1", Hash: uint32(100)}
	ser.store["test-key-2"] = model.Data{Value: 2, Hash: uint32(101)}
	ser.store["test-key-3"] = model.Data{Value: 12, Hash: uint32(201)}

	err := <-ser.BatchDelete([]string{"test-key-1", "test-key-2"})
	r.NoError(err)
	r.Len(ser.store, 1)
	r.Equal(ser.store["test-key-3"], model.Data{Value: 12, Hash: uint32(201)})
}
