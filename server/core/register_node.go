package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/pkg/errors"
)

type location struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// RegisterServerWithGateway will register the storage server with the api gateway so it can accept requests
func (s server) RegisterServerWithGateway(gateway, nodeAddress string, port int) error {

	addr := location{IP: nodeAddress, Port: fmt.Sprintf("%d", port)}
	payload, err := json.Marshal(addr)
	if err != nil {
		return err
	}
	resp, err := http.Post("http://"+gateway+"/register", "application/json", bytes.NewBuffer(payload))
	if err != nil || resp.StatusCode != http.StatusCreated {
		return err
	}
	return nil
}

// RegisterNodeWithWatchTower will register the storage node with the watchtower
func (s server) RegisterNodeWithWatchTower(watchtowerClient pb.WatchTowerClient, nodeAddress string, port int) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	addr := fmt.Sprintf("%s:%d", nodeAddress, port)
	req := &pb.RegisterNodeRequest{Address: addr}
	resp, err := watchtowerClient.RegisterNode(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to register node")
	}
	otherNodes := resp.GetExistingNodes()
	for _, node := range otherNodes {
		newCtx, newCancel := context.WithTimeout(ctx, 3*time.Second)
		defer newCancel()
		client, err := s.factory.New(newCtx, node)
		if err != nil {
			return errors.Wrap(err, "could not connect to existing node")
		}

		s.nodeClients[node] = client
		s.hashRing.AddNode(node)
	}
	return nil
}
