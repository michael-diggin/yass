package api

import (
	"context"
	"time"

	"github.com/michael-diggin/yass/common/models"
	pb "github.com/michael-diggin/yass/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RegisterNode registers a new db server and grpc client to the API WatchTower
func (wt *WatchTower) RegisterNode(ctx context.Context, req *pb.RegisterNodeRequest) (*pb.RegisterNodeResponse, error) {
	// TODO: add security to this endpoint
	addr := req.GetAddress()
	if addr == "" {
		return nil, status.Error(codes.InvalidArgument, "no node address given")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	dbClient, err := wt.clientFactory.New(ctx, addr)
	if err != nil {
		return nil, status.Error(codes.Aborted, "failed to dial server")
	}
	wt.mu.Lock()
	// send add Node instruction to existing nodes
	existingNodes := []string{}
	for nodeAddr, otherClient := range wt.Clients {
		if nodeAddr == addr {
			continue
		}
		existingNodes = append(existingNodes, nodeAddr)
		go func(address string, client models.ClientInterface) {
			subCtx, subCancel := context.WithTimeout(context.Background(), 3*time.Second)
			client.AddNode(subCtx, address)
			subCancel()
		}(addr, otherClient)
	}
	if len(wt.Clients) < wt.numServers {
		// new node for initial creation -> no rebalancing
		logrus.Info("Registering a new server with watchtower")
		wt.Clients[addr] = dbClient
		wt.hashRing.AddNode(addr)
		wt.mu.Unlock()
		return &pb.RegisterNodeResponse{ExistingNodes: existingNodes}, nil
	}
	if _, ok := wt.Clients[addr]; ok {
		// node failed and restarted -> repopulate from other nodes
		logrus.Info("Registering a failed node that restarted")
		wt.Clients[addr] = dbClient
		instructions := wt.hashRing.RebalanceInstructions(addr)
		wt.mu.Unlock()
		wt.rebalanceData(addr, instructions, false)
		return &pb.RegisterNodeResponse{ExistingNodes: existingNodes}, nil
	}
	// must be a new node added for scaling -> rebalance data from other nodes
	logrus.Info("Registering a new node for scaling")
	wt.Clients[addr] = dbClient
	wt.mu.Unlock()
	wt.hashRing.AddNode(addr)
	instructions := wt.hashRing.RebalanceInstructions(addr)
	wt.rebalanceData(addr, instructions, true)
	return &pb.RegisterNodeResponse{ExistingNodes: existingNodes}, nil
}

func (wt *WatchTower) rebalanceData(addr string, instructions []models.Instruction, delete bool) error {
	for _, instr := range instructions {
		go func(instr models.Instruction) {
			wt.mu.RLock()
			dbClient := wt.Clients[instr.FromNode]
			wt.mu.RUnlock()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			err := dbClient.BatchSend(ctx, instr.FromIdx, instr.ToIdx, addr, instr.LowHash, instr.HighHash)
			if err != nil {
				logrus.Errorf("failed to send data from node %v: %v", instr.FromNode, err)
				return
			}
			if delete {
				dbClient.BatchDelete(ctx, instr.FromIdx, instr.LowHash, instr.HighHash)
			}
			return
		}(instr)
	}
	return nil
}
