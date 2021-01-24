package api

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// PingStorageServers will health check the storage servers that are registered with the gateway every
// `freq` seconds until the context is cancelled
func (wt *WatchTower) PingStorageServers(ctx context.Context, freq time.Duration) {

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(freq):
			for addr, client := range wt.Clients {
				serverAddr := addr
				logrus.Infof("Checking storage server %s", serverAddr)
				tCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
				resp, err := client.Check(tCtx, &grpc_health_v1.HealthCheckRequest{Service: "Storage"})
				if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING || err != nil {
					logrus.Warningf("Storage server %s not serving. Response %v, error: %v", serverAddr, resp, err)
				}
				cancel()
			}
		}
	}

}
