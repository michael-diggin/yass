package core

import (
	"context"
	"time"

	"github.com/michael-diggin/yass/common/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Put will set a key and value into the data nodes
func (s *server) Put(ctx context.Context, req *pb.Pair) (*pb.Null, error) {
	if len(s.nodeClients) < s.minServers {
		return nil, status.Error(codes.Unavailable, "server is not ready yet")
	}
	logrus.Debug("Serving Put request")
	if req.GetKey() == "" || req.GetValue() == nil {
		return nil, status.Error(codes.InvalidArgument, "data not present in request")
	}

	// get node Addrs from hash ring
	hashkey := s.hashRing.Hash(req.Key)
	nodes, err := s.hashRing.GetN(hashkey, 2)
	if err != nil {
		return nil, status.Error(codes.Internal, "something went wrong")
	}
	req.Hash = hashkey

	// synchronously set the key/value on the storage servers
	revertSetNodes := []models.Node{}
	var returnErr error
	for _, node := range nodes {
		s.mu.RLock()
		client := s.nodeClients[node.ID]
		s.mu.RUnlock()
		subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		setReq := &pb.SetRequest{Replica: int32(node.Idx), Pair: req}
		_, err := client.Set(subctx, setReq)
		cancel()
		if err != nil {
			returnErr = err
			break
		}
		revertSetNodes = append(revertSetNodes, node)
	}

	if returnErr != nil {
		logrus.Errorf("Encountered error: %v", returnErr)
		// revert any changes that were made before an error
		for _, node := range revertSetNodes {
			n := node
			s.mu.RLock()
			client := s.nodeClients[n.ID]
			s.mu.RUnlock()
			go func() {
				subctx, cancel := context.WithTimeout(ctx, 3*time.Second)
				client.Delete(subctx, &pb.DeleteRequest{Replica: int32(n.Idx), Key: req.Key})
				cancel()
			}()
		}

		return nil, status.Error(status.Code(returnErr), returnErr.Error())
	}

	return &pb.Null{}, nil
}

// Retrieve will return the value for a given key if it is in the data nodes
func (s *server) Retrieve(ctx context.Context, req *pb.Key) (*pb.Pair, error) {
	if len(s.nodeClients) < s.minServers {
		return nil, status.Error(codes.Unavailable, "server is not ready yet")
	}
	logrus.Debug("Serving Retrieve request")
	if req.GetKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "key not present in request")
	}

	hashkey := s.hashRing.Hash(req.Key)
	nodes, err := s.hashRing.GetN(hashkey, 2)
	if err != nil {
		return nil, status.Error(codes.Internal, "something went wrong")
	}

	newctx, cancel := context.WithCancel(ctx)
	defer cancel()

	resps := make(chan internalResponse, len(nodes))
	for _, node := range nodes {
		n := node
		s.mu.RLock()
		client := s.nodeClients[n.ID]
		s.mu.RUnlock()
		go func() {
			subctx, cancel := context.WithTimeout(newctx, 3*time.Second)
			defer cancel()
			getReq := &pb.GetRequest{Replica: int32(n.Idx), Key: req.Key}
			pair, err := client.Get(subctx, getReq)
			if err != nil {
				resps <- internalResponse{err: err}
				return
			}
			resps <- internalResponse{value: pair.Value, err: err}
		}()
	}

	value, retErr := getValueFromRequests(resps, len(nodes), cancel)

	if value == nil && retErr != nil {
		return nil, status.Error(status.Code(retErr), retErr.Error())
	}

	return &pb.Pair{Key: req.Key, Value: value}, nil
}

type internalResponse struct {
	value []byte
	err   error
}

func getValueFromRequests(resps chan internalResponse, n int, cancel context.CancelFunc) ([]byte, error) {
	var err error
	var value []byte
	for i := 0; i < n; i++ {
		r := <-resps
		if r.err != nil && err == nil {
			err = r.err
		}
		if r.value != nil {
			value = r.value
			cancel()
			break
		}
	}
	return value, err
}
