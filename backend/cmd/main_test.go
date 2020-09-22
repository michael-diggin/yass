package main

import (
	"context"
	"testing"
	"time"

	pb "github.com/michael-diggin/yass/api"
	"google.golang.org/grpc"
)

func TestRunServer(t *testing.T) {
	args := []string{"program", "-p", "8080"}
	envFunc := func(input string) string {
		return ""
	}

	go func() {
		run(args, envFunc)
	}()

	time.Sleep(50 * time.Millisecond) // Give server time to set up
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "localhost:8080", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Could not dial server: %err", err)
	}

	client := pb.NewCacheClient(conn)
	resp, err := client.Ping(ctx, &pb.Null{})
	if err != nil {
		t.Fatalf("Could not send Ping command: %v", err)
	}
	if resp.Status != pb.PingResponse_SERVING {
		t.Fatalf("Server not serving")
	}
}
