package core

import (
	"context"
	"encoding/json"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BatchGet returns all of the stored data in a given replica
func (s server) BatchGet(ctx context.Context, req *pb.BatchGetRequest) (*pb.BatchGetResponse, error) {
	logrus.Info("Serving BatchGet request")
	stores := map[int32]model.Service{0: s.Leader, 1: s.Follower}
	store := stores[req.GetReplica()]
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case storedData := <-store.BatchGet():
		data := make([]*pb.Pair, 0, len(storedData))
		for k, v := range storedData {
			valueBytes, err := json.Marshal(v)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to marshal data")
			}
			data = append(data, &pb.Pair{Key: k, Value: valueBytes})
		}
		resp := &pb.BatchGetResponse{Replica: req.GetReplica(), Data: data}
		logrus.Info("BatchGet request succeeded")
		return resp, nil
	}

}

// BatchSet sets the values into the store
func (s server) BatchSet(ctx context.Context, req *pb.BatchSetRequest) (*pb.Null, error) {
	logrus.Info("Serving BatchSet request")
	stores := map[int32]model.Service{0: s.Leader, 1: s.Follower}
	store := stores[req.GetReplica()]
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case err := <-batchSet(store, req.GetData()):
		if err != nil {
			return nil, status.Error(codes.Internal, "failed to set data")
		}
		logrus.Info("BatchSet request succeeded")
		return &pb.Null{}, nil
	}
}

func batchSet(store model.Service, data []*pb.Pair) <-chan error {
	resp := make(chan error)
	go func() {
		defer close(resp)
		mapData := make(map[string]interface{})
		for _, pair := range data {
			var value interface{}
			err := json.Unmarshal(pair.Value, &value)
			if err != nil {
				resp <- err
				return
			}
			mapData[pair.Key] = value
		}
		err := <-store.BatchSet(mapData)
		resp <- err
		return
	}()
	return resp
}
