package storage

import (
	"context"

	"github.com/michael-diggin/yass/backend/model"
	"github.com/sirupsen/logrus"

	"github.com/gomodule/redigo/redis"
)

// RedisService implements the storage service interface
type RedisService struct {
	conn redis.Conn
}

// NewRedisService returns an instance of the redis service
func NewRedisService(username, password, addr string) (RedisService, error) {
	userOpt := redis.DialUsername(username)
	passOpt := redis.DialPassword(password)
	conn, err := redis.Dial("tcp", addr, userOpt, passOpt)

	return RedisService{conn}, err
}

//Close terminates the connection
func (r RedisService) Close() error {
	return r.conn.Close()
}

// Set is redis implementation of service set
func (r RedisService) Set(ctx context.Context, key, value string) <-chan *model.CacheResponse {
	respChan := make(chan *model.CacheResponse, 1)
	go func() {
		_, err := r.conn.Do("SET", key, value)
		if err != nil {
			logrus.Errorf("Error trying to set key %s: %v", key, err)
			respChan <- &model.CacheResponse{Err: err}
		}
		respChan <- &model.CacheResponse{Key: key, Err: nil}
	}()
	return respChan
}

// Get is redis implementation of service get
func (r RedisService) Get(ctx context.Context, key string) <-chan *model.CacheResponse {
	respChan := make(chan *model.CacheResponse, 1)
	go func() {
		value, err := redis.String(r.conn.Do("GET", key))
		if err != nil {
			logrus.Errorf("Error trying to get key %s: %v", key, err)
			respChan <- &model.CacheResponse{Err: err}
		}
		respChan <- &model.CacheResponse{Key: key, Value: value}
	}()
	return respChan
}

// Delete is redis implementation of service delete
func (r RedisService) Delete(ctx context.Context, key string) <-chan *model.CacheResponse {
	respChan := make(chan *model.CacheResponse)
	go func() {
		_, err := r.conn.Do("DEL", key)
		respChan <- &model.CacheResponse{Err: err}
	}()
	return respChan
}
