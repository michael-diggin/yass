package storage

import (
	"context"
	"errors"
	"io/ioutil"
	"testing"

	"github.com/rafaeljusto/redigomock"
	"github.com/sirupsen/logrus"
)

func TestRedisSet(t *testing.T) {
	conn := redigomock.NewConn()
	conn.Clear()
	cmd := conn.Command("SET", "testKey", "testValue").Expect("ok")
	service := RedisService{conn: conn}
	defer service.Close()
	ctx := context.TODO()
	key, err := service.Set(ctx, "testKey", "testValue")

	if err != nil {
		t.Fatalf("Failed to set: %v", err)
	}

	if key != "testKey" {
		t.Fatalf("Expected 'testKey', got %s", key)
	}

	if conn.Stats(cmd) != 1 {
		t.Fatal("Command was not called!")
	}

}

func TestSetRedisWithError(t *testing.T) {
	logrus.SetOutput(ioutil.Discard) // Discard log output for test
	conn := redigomock.NewConn()
	conn.Clear()
	cmd := conn.Command("SET", "testKey", "testValue").ExpectError(errors.New("Redis Error thrown"))
	service := RedisService{conn: conn}
	defer service.Close()
	ctx := context.TODO()
	key, err := service.Set(ctx, "testKey", "testValue")

	if err.Error() != "Redis Error thrown" {
		t.Fatalf("Expected 'Key not in cache' got: %v", err)
	}

	if key != "" {
		t.Fatalf("Expected empty string, got %s", key)
	}

	if conn.Stats(cmd) != 1 {
		t.Fatal("Command was not called!")
	}
}

func TestGetRedisWithError(t *testing.T) {
	logrus.SetOutput(ioutil.Discard) // Discard log output for test
	conn := redigomock.NewConn()
	conn.Clear()
	cmd := conn.Command("GET", "testKey").ExpectError(errors.New("Key not in cache"))
	service := RedisService{conn: conn}
	defer service.Close()
	ctx := context.TODO()
	value, err := service.Get(ctx, "testKey")

	if err.Error() != "Key not in cache" {
		t.Fatalf("Expected 'Key not in cache' got: %v", err)
	}

	if value != "" {
		t.Fatalf("Expected empty string, got %s", value)
	}

	if conn.Stats(cmd) != 1 {
		t.Fatal("Command was not called!")
	}
}

func TestRedisGet(t *testing.T) {
	conn := redigomock.NewConn()
	conn.Clear()
	cmd := conn.Command("GET", "testKey").Expect("testValue")
	service := RedisService{conn: conn}
	defer service.Close()
	ctx := context.TODO()
	value, err := service.Get(ctx, "testKey")

	if err != nil {
		t.Fatalf("Failed to set: %v", err)
	}

	if value != "testValue" {
		t.Fatalf("Expected 'testKey', got %s", value)
	}

	if conn.Stats(cmd) != 1 {
		t.Fatal("Command was not called!")
	}

}

func TestRedisDelete(t *testing.T) {
	conn := redigomock.NewConn()
	conn.Clear()
	cmd := conn.Command("DEL", "testKey").Expect("ok")
	service := RedisService{conn: conn}
	defer service.Close()
	ctx := context.TODO()
	err := service.Delete(ctx, "testKey")

	if err != nil {
		t.Fatalf("Failed to set: %v", err)
	}

	if conn.Stats(cmd) != 1 {
		t.Fatal("Command was not called!")
	}

}
