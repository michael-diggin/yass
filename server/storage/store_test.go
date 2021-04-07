package storage

import (
	"testing"

	"github.com/michael-diggin/yass/common/yasserrors"
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
		{"not-serving", &Service{}, yasserrors.NotServing{}},
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
	_ = <-ser.Set("test-key", uint32(100), "test-value", true)

	tt := []struct {
		name  string
		key   string
		value interface{}
		hash  uint32
		err   error
	}{
		{"valid", "key", "value", uint32(77), nil},
		{"already set", "test-key", "test-value", uint32(100), nil},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			resp := <-ser.Set(tc.key, tc.hash, tc.value, true)
			r.Equal(tc.err, resp.Err)
			r.Equal(tc.key, resp.Key)
		})
	}
}

func TestSetInCacheWithCommit(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()
	_ = <-ser.Set("test-key", uint32(100), "test-value", true)

	resp := <-ser.Set("test-key", uint32(100), "new-value", false)
	r.NoError(resp.Err)
	r.Equal("test-key", resp.Key)

	// check it hasn't changed the commit data
	r.Equal("test-value", ser.db["test-key"].Value)
	r.Equal("new-value", ser.proposed["test-key"].Value)

	resp = <-ser.Set("test-key", uint32(100), "new-value", true)
	r.NoError(resp.Err)
	r.Equal("test-key", resp.Key)

	// now it should overwrite the data
	r.Equal("new-value", ser.db["test-key"].Value)
	r.Empty(ser.proposed["test-key"])
}

func TestSetInCacheWithTransacionError(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()
	_ = <-ser.Set("test-key", uint32(100), "test-value", true)

	resp := <-ser.Set("test-key", uint32(100), "new-value", false)
	r.NoError(resp.Err)
	r.Equal("test-key", resp.Key)

	// check it hasn't changed the commit data
	r.Equal("test-value", ser.db["test-key"].Value)
	r.Equal("new-value", ser.proposed["test-key"].Value)

	resp = <-ser.Set("test-key", uint32(100), "different-value", false)
	r.Error(resp.Err)
	r.Equal(yasserrors.TransactionError{Key: "test-key"}, resp.Err)

	// check it hasn't changed the commit data
	r.Equal("test-value", ser.db["test-key"].Value)
	r.Equal("new-value", ser.proposed["test-key"].Value)
}

func TestGetFromCache(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()
	_ = <-ser.Set("test-key", uint32(100), "test-value", true)

	tt := []struct {
		name  string
		key   string
		value interface{}
		err   error
	}{
		{"valid", "test-key", "test-value", nil},
		{"not found", "key", nil, yasserrors.NotFound{Key: "key"}},
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
	_ = <-ser.Set("test-key", uint32(100), "test-value", true)

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
	ser.db["test-key-1"] = model.Data{Value: "value-1", Hash: uint32(100)}
	ser.db["test-key-2"] = model.Data{Value: 2, Hash: uint32(201)}

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
	r.Len(ser.db, 2)
	r.Equal(ser.db["test-key-1"], model.Data{Value: "value-1", Hash: uint32(100)})
	r.Equal(ser.db["test-key-2"], model.Data{Value: 2, Hash: uint32(101)})
}

func TestBatchDelete(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()

	ser.db["test-key-1"] = model.Data{Value: "value-1", Hash: uint32(100)}
	ser.db["test-key-2"] = model.Data{Value: 2, Hash: uint32(101)}
	ser.db["test-key-3"] = model.Data{Value: 12, Hash: uint32(201)}

	err := <-ser.BatchDelete(uint32(50), uint32(150))
	r.NoError(err)
	r.Len(ser.db, 1)
	r.Equal(ser.db["test-key-3"], model.Data{Value: 12, Hash: uint32(201)})
}
