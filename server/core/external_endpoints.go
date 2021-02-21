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
	node := s.hashRing.Get(hashkey)
	req.Hash = hashkey
	setReq := &pb.SetRequest{Replica: int32(node.Idx), Pair: req}
	subctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// propose write to all nodes
	respChan := make(chan error, 3)
	for _, client := range s.nodeClients {
		go func(c *models.StorageClient, errChan chan error) {
			_, err := c.Set(subctx, setReq)
			errChan <- err
		}(client, respChan)
	}

	// wait and listen for responses
	var commit, abort int = 0, 0
	var returnErr error
	for err := range respChan {
		if err == nil {
			commit++
		} else {
			abort++
			returnErr = err
		}
		if commit >= 2 || abort >= 2 {
			break
		}
	}

	if commit >= 2 {
		// write successful, exit here
		// in future edits, there will be a commit call made to each node
		logrus.Info("Committing")
		return &pb.Null{}, nil
	}

	// else, abort the write, delete
	cancel() // this is in case we got two errors, can cancel the 3rd request

	// revert any changes that were made before an error
	subctx, cancel = context.WithTimeout(ctx, 5*time.Second)
	delReq := &pb.DeleteRequest{Replica: int32(node.Idx), Key: req.Key}
	defer cancel()
	for id, client := range s.nodeClients {
		id := id
		c := client
		go func() {
			_, err := c.Delete(subctx, delReq)
			if err != nil {
				logrus.Errorf("err aborting write from node %s: %v", id, err)
			}
		}()
	}

	logrus.Errorf("Encountered error: %v", returnErr)
	return nil, status.Error(status.Code(returnErr), returnErr.Error())
}

// Fetch will return the value for a given key if it is in the data nodes
func (s *server) Fetch(ctx context.Context, req *pb.Key) (*pb.Pair, error) {
	if len(s.nodeClients) < s.minServers {
		return nil, status.Error(codes.Unavailable, "server is not ready yet")
	}
	logrus.Debug("Serving Fetch request")
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
