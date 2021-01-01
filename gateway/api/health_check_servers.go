package api

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// PingStorageServers will health check the storage servers that are registered with the gateway every
// `freq` seconds until the context is cancelled
func (g *Gateway) PingStorageServers(ctx context.Context, freq time.Duration) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(freq):
			for addr, client := range g.Clients {
				serverAddr := addr
				logrus.Infof("Checking storage server %s", serverAddr)
				tCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				ok, err := client.Check(tCtx)
				if !ok || err != nil {
					logrus.Warningf("Storage server %s not serving", serverAddr)
					delete(g.Clients, serverAddr)
					g.hashRing.RemoveNode(serverAddr)
				}
				cancel()
				// TODO: strategy for dealing with dropping node
				// idea: post `instruction` for replicate data from replica A to new replica
				// hashRimg.RemoveNode returns instruction (ranges of hash, nodes that they are on, need to go too?)
				// and the nodes themselves can request the data from other nodes?
			}
		}
	}
}
