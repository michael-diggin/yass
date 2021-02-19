package core

import (
	"context"
	"time"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/pkg/errors"
)

type location struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

// RegisterNodeWithWatchTower will register the storage node with the watchtower
func (s *server) RegisterNodeWithWatchTower(watchtowerClient pb.WatchTowerClient, nodeAddress string) error {

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req := &pb.RegisterNodeRequest{Address: nodeAddress}
	resp, err := watchtowerClient.RegisterNode(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to register node")
	}
	otherNodes := resp.GetExistingNodes()
	for _, node := range otherNodes {
		newCtx, newCancel := context.WithTimeout(ctx, 10*time.Second)
		defer newCancel()
		client, err := s.factory.NewProtoClient(newCtx, node)
		if err != nil {
			return errors.Wrap(err, "could not connect to existing node")
		}
		// register itself with the existing nodes in the cluster
		_, err = client.AddNode(newCtx, &pb.AddNodeRequest{Node: nodeAddress})
		if err != nil {
			return errors.Wrap(err, "could not add node to other nodes")
		}

		s.nodeClients[node] = client
		s.hashRing.AddNode(node)
	}
	return nil
}
