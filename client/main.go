package main

import (
	"context"
	"flag"

	"google.golang.org/grpc"

	pb "github.com/michael-diggin/yass/api"
	"github.com/sirupsen/logrus"
)

func main() {
	backend := flag.String("b", "localhost:8080", "address of the server")
	flag.Parse()

	conn, err := grpc.Dial(*backend, grpc.WithInsecure())
	if err != nil {
		logrus.Fatalf("Could not connect to %s: %v", *backend, err)
	}
	defer conn.Close()

	client := pb.NewCacheClient(conn)

	pair := &pb.Pair{Key: "New key", Value: "New Value"}
	_, err = client.Add(context.Background(), pair)
	if err != nil {
		logrus.Fatalf("Could not add key %s: %v", pair.Key, err)
	}
	pair = &pb.Pair{Key: "key2", Value: "value2"}
	_, err = client.Add(context.Background(), pair)
	if err != nil {
		logrus.Fatalf("Could not add key %s: %v", pair.Key, err)
	}

	keys := []string{
		"New key",
		"key2",
	}

	for _, str := range keys {
		k := &pb.Key{Key: str}
		res, err := client.Get(context.Background(), k)
		if err != nil {
			logrus.Infof("Could not get value for %s: %v", k.Key, err)
		} else {
			logrus.Infof(res.Value)
		}
	}
}
