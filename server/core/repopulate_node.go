package core

import (
	"context"
	"encoding/json"
	"time"

	"github.com/michael-diggin/yass/common/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/model"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// RepopulateFromNodes will request all of the data from the other nodes in the cluster
// in the event that a node goes down and needs to be brought back up to the current state
func (s *server) RepopulateFromNodes(nodes ...string) error {
	for _, node := range nodes {
		client := s.nodeClients[node]
		for i, store := range s.DataStores {
			time.Sleep(100 * time.Millisecond)
			go func(client *models.StorageClient, store model.Service, i int) {
				err := s.getDataForStore(client, store, i)
				if err != nil {
					logrus.Errorf("unable to get data from node %s, store %d: %v", node, i, err)
				}
			}(client, store, i)
		}
	}
	return nil
}

func (s *server) getDataForStore(client *models.StorageClient, store model.Service, replica int) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req := &pb.BatchGetRequest{
		Replica: int32(replica),
	}
	resp, err := client.BatchGet(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to get data for store")
	}
	data := resp.Data
	err = setProtoDataToStore(store, data)
	return err
}

func setProtoDataToStore(store model.Service, newData []*pb.Pair) error {
	mapData := make(map[string]model.Data)
	for _, pair := range newData {
		var value interface{}
		err := json.Unmarshal(pair.Value, &value)
		if err != nil {
			return err
		}
		mapData[pair.Key] = model.Data{Value: value, Hash: pair.Hash}
	}
	err := <-store.BatchSet(mapData)
	return err
}
