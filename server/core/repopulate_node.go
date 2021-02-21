package core

import (
	"context"
	"math"
	"strconv"
	"time"

	pb "github.com/michael-diggin/yass/proto"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (s *server) RepopulateNodes(ctx context.Context, podName string) {
	currentOrdinalStr := podName[5] // yass-1.yassdb, want 1
	currOrd, err := strconv.Atoi(string(currentOrdinalStr))
	if err != nil {
		logrus.Errorf("could not get pod ordinal: %v", err)
		return
	}
	for {
		select {
		case <-ctx.Done():
			logrus.Info("Context cancelled, exitting")
			return
		case newNode := <-s.repopulateChan:
			newOrdinal := newNode[5]
			newOrd, err := strconv.Atoi(string(newOrdinal))
			if err != nil {
				logrus.Errorf("could not get new pod ordinal: %v", err)
			}
			err = s.sendData(newNode, currOrd, newOrd)
			if err != nil {
				logrus.Errorf("could not send data to node %s: %v", newNode, err)
			}
		}
	}
}

func (s *server) sendData(node string, currentOrdinal, newOrdinal int) error {
	parity := (currentOrdinal + 3 - newOrdinal) % 2
	for i := 0; i < len(s.DataStores); i++ {
		if (i % 2) != parity {
			continue
		}
		req := &pb.BatchSendRequest{
			Replica:   int32(i),
			Address:   node,
			ToReplica: int32(i),
			Low:       uint32(0),
			High:      math.MaxUint32,
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := s.BatchSend(ctx, req)
		cancel()
		if err != nil {
			return errors.Wrap(err, "failed to send data for store %d")
		}

	}
	return nil
}
