package storage

import (
	"testing"

	"github.com/michael-diggin/yass/server/errors"
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
	_ = <-ser.Set("test-key", "test-value")

	tt := []struct {
		name  string
		key   string
		value interface{}
		err   error
	}{
		{"valid", "key", "value", nil},
		{"already set", "test-key", "test-value", errors.AlreadySet{Key: "test-key"}},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			resp := <-ser.Set(tc.key, tc.value)
			r.Equal(tc.err, resp.Err)
			r.Equal(tc.key, resp.Key)
		})
	}
}

func TestGetFromCache(t *testing.T) {
	r := require.New(t)
	ser := New()
	defer ser.Close()
	_ = <-ser.Set("test-key", "test-value")

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
	_ = <-ser.Set("test-key", "test-value")

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
