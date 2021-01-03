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
				tCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
				ok, err := client.Check(tCtx)
				if !ok || err != nil {
					logrus.Warningf("Storage server %s not serving", serverAddr)
				}
				cancel()
			}
		}
	}
}
