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
	store, err := s.getStoreForRequest(int(req.GetReplica()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case storedData := <-store.BatchGet(uint32(0), uint32(1)):
		data := make([]*pb.Pair, 0, len(storedData))
		for k, d := range storedData {
			valueBytes, err := json.Marshal(d.Value)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to marshal data")
			}
			data = append(data, &pb.Pair{Key: k, Hash: d.Hash, Value: valueBytes})
		}
		resp := &pb.BatchGetResponse{Replica: req.GetReplica(), Data: data}
		logrus.Info("BatchGet request succeeded")
		return resp, nil
	}

}

// BatchSet sets the values into the store
func (s server) BatchSet(ctx context.Context, req *pb.BatchSetRequest) (*pb.Null, error) {
	logrus.Info("Serving BatchSet request")
	store, err := s.getStoreForRequest(int(req.GetReplica()))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
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

func batchSet(store model.Service, newData []*pb.Pair) <-chan error {
	resp := make(chan error)
	go func() {
		defer close(resp)
		mapData := make(map[string]model.Data)
		for _, pair := range newData {
			var value interface{}
			err := json.Unmarshal(pair.Value, &value)
			if err != nil {
				resp <- err
				return
			}
			mapData[pair.Key] = model.Data{Value: value, Hash: pair.Hash}
		}
		err := <-store.BatchSet(mapData)
		resp <- err
		return
	}()
	return resp
}
