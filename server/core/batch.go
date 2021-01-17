package core

import (
	"context"
	"encoding/json"
	"time"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/michael-diggin/yass/server/model"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BatchGet returns all of the stored data in a given replica
func (s *server) BatchGet(ctx context.Context, req *pb.BatchGetRequest) (*pb.BatchGetResponse, error) {
	logrus.Info("Serving BatchGet request")
	store, err := s.getStoreForRequest(req.GetReplica())
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
func (s *server) BatchSet(ctx context.Context, req *pb.BatchSetRequest) (*pb.Null, error) {
	logrus.Info("Serving BatchSet request")
	store, err := s.getStoreForRequest(req.GetReplica())
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

// BatchSend gets data in certain hash ranges and send it to another data node
func (s *server) BatchSend(ctx context.Context, req *pb.BatchSendRequest) (*pb.Null, error) {
	logrus.Info("Serving BatchSend request")
	store, err := s.getStoreForRequest(req.GetReplica())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case storedData := <-store.BatchGet(req.Low, req.High):
		data := make([]*pb.Pair, 0, len(storedData))
		keys := make([]string, 0, len(storedData))
		for k, d := range storedData {
			valueBytes, err := json.Marshal(d.Value)
			if err != nil {
				return nil, status.Error(codes.Internal, "failed to marshal data")
			}
			data = append(data, &pb.Pair{Key: k, Hash: d.Hash, Value: valueBytes})
			keys = append(keys, k)
		}

		newCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		client, ok := s.nodeClients[req.Address]
		if !ok {
			return nil, status.Error(codes.Internal, "could not connect to node")
		}
		err = client.BatchSet(newCtx, int(req.ToReplica), data)
		if err != nil {
			return nil, status.Error(codes.Internal, "could not set data on node")
		}

		logrus.Info("BatchSend request succeeded")
		return &pb.Null{}, nil
	}
}

// BatchDelete deletes data where the hash of the key lies in a hash range.
func (s *server) BatchDelete(ctx context.Context, req *pb.BatchDeleteRequest) (*pb.Null, error) {
	logrus.Info("Serving BatchDelete request")
	store, err := s.getStoreForRequest(req.GetReplica())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	select {
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "Context timeout")
	case err := <-store.BatchDelete(req.Low, req.High):
		if err != nil {
			return nil, status.Error(codes.Internal, "could not delete data from store")
		}
		logrus.Info("BatchDelete request succeeded")
		return &pb.Null{}, nil
	}
}
