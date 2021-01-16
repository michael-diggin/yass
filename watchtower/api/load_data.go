package api

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// LoadData loads the data from a file containing the existing node address
// and populates the WT hash ring and client map
func (wt *WatchTower) LoadData() error {
	r, err := os.Open(wt.nodeFile)
	if err != nil {
		return err
	}
	defer r.Close()
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	str := string(b)
	nodes := strings.Split(str, "\n")
	wt.mu.Lock()
	defer wt.mu.Unlock()
	for _, nodeAddr := range nodes {
		if nodeAddr == "" {
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		dbClient, err := wt.clientFactory.New(ctx, nodeAddr)
		cancel()
		if err != nil {
			return fmt.Errorf("failed to dial node %s: %v", nodeAddr, err)
		}
		wt.Clients[nodeAddr] = dbClient
		wt.hashRing.AddNode(nodeAddr)
		logrus.Infof("Loaded node data for %s", nodeAddr)
	}
	return nil
}
