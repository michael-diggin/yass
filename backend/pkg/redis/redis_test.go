package redis

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
	service := Service{conn: conn}
	defer service.Close()
	ctx := context.TODO()
	resp := <-service.Set(ctx, "testKey", "testValue")

	if resp.Err != nil {
		t.Fatalf("Failed to set: %v", resp.Err)
	}

	if resp.Key != "testKey" {
		t.Fatalf("Expected 'testKey', got %s", resp.Key)
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
	service := Service{conn: conn}
	defer service.Close()
	ctx := context.TODO()
	resp := <-service.Set(ctx, "testKey", "testValue")

	if resp.Err.Error() != "Redis Error thrown" {
		t.Fatalf("Expected 'Key not in cache' got: %v", resp.Err)
	}

	if resp.Key != "" {
		t.Fatalf("Expected empty string, got %s", resp.Key)
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
	service := Service{conn: conn}
	defer service.Close()
	ctx := context.TODO()
	resp := <-service.Get(ctx, "testKey")

	if resp.Err.Error() != "Key not in cache" {
		t.Fatalf("Expected 'Key not in cache' got: %v", resp.Err)
	}

	if resp.Value != "" {
		t.Fatalf("Expected empty string, got %s", resp.Value)
	}

	if conn.Stats(cmd) != 1 {
		t.Fatal("Command was not called!")
	}
}

func TestRedisGet(t *testing.T) {
	conn := redigomock.NewConn()
	conn.Clear()
	cmd := conn.Command("GET", "testKey").Expect("testValue")
	service := Service{conn: conn}
	defer service.Close()
	ctx := context.TODO()
	resp := <-service.Get(ctx, "testKey")

	if resp.Err != nil {
		t.Fatalf("Failed to get: %v", resp.Err)
	}

	if resp.Value != "testValue" {
		t.Fatalf("Expected 'testKey', got %s", resp.Value)
	}

	if conn.Stats(cmd) != 1 {
		t.Fatal("Command was not called!")
	}

}

func TestRedisDelete(t *testing.T) {
	conn := redigomock.NewConn()
	conn.Clear()
	cmd := conn.Command("DEL", "testKey").Expect("ok")
	service := Service{conn: conn}
	defer service.Close()
	ctx := context.TODO()
	resp := <-service.Delete(ctx, "testKey")

	if resp.Err != nil {
		t.Fatalf("Failed to set: %v", resp.Err)
	}

	if conn.Stats(cmd) != 1 {
		t.Fatal("Command was not called!")
	}

}
