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

	logrus.Debug("Serving Put request as RaftLeader")
	if req.GetKey() == "" || req.GetValue() == nil {
		return nil, status.Error(codes.InvalidArgument, "data not present in request")
	}

	if !s.IsLeader() {
		logrus.Debug("Redirecting Put request to RaftLeader")
		leader := s.nodeClients[s.RaftLeader]
		return leader.Put(ctx, req)
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
	node := s.hashRing.Get(hashkey)

	newctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	getReq := &pb.GetRequest{Replica: int32(node.Idx), Key: req.Key}

	resps := make(chan internalResponse, s.minServers)
	for _, client := range s.nodeClients {
		go func(c *models.StorageClient, resps chan internalResponse) {
			pair, err := c.Get(newctx, getReq)
			if err != nil {
				resps <- internalResponse{err: err}
				return
			}
			resps <- internalResponse{value: pair.Value, err: err}
		}(client, resps)
	}

	value, retErr := getValueFromRequests(resps, s.minServers)
	cancel()

	if value == nil && retErr != nil {
		return nil, status.Error(status.Code(retErr), retErr.Error())
	}

	return &pb.Pair{Key: req.Key, Value: value}, nil
}

type internalResponse struct {
	value []byte
	err   error
}

func getValueFromRequests(resps chan internalResponse, n int) ([]byte, error) {
	var err error
	valMap := make(map[string]int)
	stringValMap := make(map[string][]byte)
	numErrs := 0
	responses := 0
	for r := range resps {
		responses++
		if r.err != nil {
			err = r.err
			numErrs++
		}
		if r.value != nil {
			str := string(r.value)
			stringValMap[str] = r.value
			valMap[str]++
			if valMap[str] >= 2 {
				return stringValMap[str], nil
			}
		}

		if numErrs >= 2 {
			return nil, err
		}
		if responses >= n {
			break
		}
	}
	// weird case where there's no consensus on value or error
	// should only happen when more than one node is down and they haven't caught up yet
	// for now respond with aborted: eVEntUaL COnsIsTeNCy
	// TODO: return value with highest txn key maybe?
	return nil, status.Error(codes.Aborted, "no quorum was reached")
}
