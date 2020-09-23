package redis

import (
	"context"

	"github.com/michael-diggin/yass/backend"
	"github.com/sirupsen/logrus"

	"github.com/gomodule/redigo/redis"
)

// Service implements the storage service interface
type Service struct {
	conn redis.Conn
}

// New returns an instance of the redis service
func New(username, password, addr string) (Service, error) {
	userOpt := redis.DialUsername(username)
	passOpt := redis.DialPassword(password)
	conn, err := redis.Dial("tcp", addr, userOpt, passOpt)

	return Service{conn}, err
}

// Close terminates the connection
func (r Service) Close() error {
	return r.conn.Close()
}

// Ping checks if the redis connection is reachable
func (r Service) Ping() error {
	_, err := r.conn.Do("PING")
	return err
}

// Set is redis implementation of service set
func (r Service) Set(ctx context.Context, key, value string) <-chan *backend.CacheResponse {
	respChan := make(chan *backend.CacheResponse, 1)
	go func() {
		_, err := r.conn.Do("SET", key, value)
		if err != nil {
			logrus.Errorf("Error trying to set key %s: %v", key, err)
			respChan <- &backend.CacheResponse{Err: err}
		}
		respChan <- &backend.CacheResponse{Key: key, Err: nil}
		close(respChan)
	}()
	return respChan
}

// Get is redis implementation of service get
func (r Service) Get(ctx context.Context, key string) <-chan *backend.CacheResponse {
	respChan := make(chan *backend.CacheResponse, 1)
	go func() {
		value, err := redis.String(r.conn.Do("GET", key))
		if err != nil {
			logrus.Errorf("Error trying to get key %s: %v", key, err)
			respChan <- &backend.CacheResponse{Err: err}
		}
		respChan <- &backend.CacheResponse{Key: key, Value: value}
		close(respChan)
	}()
	return respChan
}

// Delete is redis implementation of service delete
func (r Service) Delete(ctx context.Context, key string) <-chan *backend.CacheResponse {
	respChan := make(chan *backend.CacheResponse)
	go func() {
		_, err := r.conn.Do("DEL", key)
		respChan <- &backend.CacheResponse{Err: err}
		close(respChan)
	}()
	return respChan
}
