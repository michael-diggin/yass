package api

import (
	"context"
	"os"
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

	wt.mu.RLock()
	// get all the other nodes in the hash ring
	existingNodes := []string{}
	for nodeAddr := range wt.Clients {
		if nodeAddr == addr {
			continue
		}
		existingNodes = append(existingNodes, nodeAddr)
	}
	// quick check to see if this is a restarted node
	if _, ok := wt.Clients[addr]; ok {
		wt.mu.RUnlock()
		// node failed and restarted -> repopulate from other nodes
		logrus.Info("Registering a failed node that restarted")
		instructions := wt.hashRing.RebalanceInstructions(addr)
		wt.rebalanceData(addr, instructions, false)
		return &pb.RegisterNodeResponse{ExistingNodes: existingNodes}, nil
	}

	// add node data to the file
	f, err := os.OpenFile(wt.nodeFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, status.Error(codes.Internal, "unable to open node file")
	}
	if _, err := f.Write([]byte("\n" + addr)); err != nil {
		return nil, status.Error(codes.Internal, "unable to append node to node file")
	}
	f.Close()

	wt.mu.RUnlock()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	dbClient, err := wt.clientFactory.NewProtoClient(ctx, addr)
	if err != nil {
		return nil, status.Error(codes.Aborted, "failed to dial server")
	}
	wt.mu.Lock()
	// send add Node instruction to existing nodes
	for _, nodeAddr := range existingNodes {
		otherClient := wt.Clients[nodeAddr]
		go func(address string, client pb.StorageClient) {
			subCtx, subCancel := context.WithTimeout(context.Background(), 3*time.Second)
			req := &pb.AddNodeRequest{Node: address}
			client.AddNode(subCtx, req)
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
			req := &pb.BatchSendRequest{
				Replica:   int32(instr.FromIdx),
				Address:   addr,
				ToReplica: int32(instr.ToIdx),
				Low:       instr.LowHash,
				High:      instr.HighHash,
			}
			_, err := dbClient.BatchSend(ctx, req)
			if err != nil {
				logrus.Errorf("failed to send data from node %v: %v", instr.FromNode, err)
				return
			}
			if delete {
				req := &pb.BatchDeleteRequest{
					Replica: int32(instr.FromIdx),
					Low:     instr.LowHash,
					High:    instr.HighHash,
				}
				dbClient.BatchDelete(ctx, req)
			}
			return
		}(instr)
	}
	return nil
}
