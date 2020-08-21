package storage

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/gomodule/redigo/redis"
)

// RedisService implements the storage service interface
type RedisService struct {
	conn redis.Conn
}

// NewRedisService returns an instance of the redis service
func NewRedisService(conn redis.Conn) RedisService {
	return RedisService{conn: conn}

}

// Set is redis implementation of service set
func (r RedisService) Set(ctx context.Context, key, value string) (string, error) {
	_, err := r.conn.Do("SET", key, value)
	if err != nil {
		logrus.Errorf("Error trying to set key %s: %v", key, err)
		return "", err
	}
	return key, nil
}

// Get is redis implementation of service get
func (r RedisService) Get(ctx context.Context, key string) (string, error) {
	value, err := redis.String(r.conn.Do("GET", key))
	if err != nil {
		logrus.Errorf("Error trying to get key %s: %v", key, err)
		return "", err
	}
	return value, nil
}

// Delete is redis implementation of service delete
func (r RedisService) Delete(ctx context.Context, key string) error {
	_, err := r.conn.Do("DEL", key)
	return err
}
