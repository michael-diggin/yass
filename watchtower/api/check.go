package api

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// Check is the health check endpoint
func (wt *WatchTower) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	logrus.Info("Serving the Check request for health check")

	// TODO: these should be more indicative of how the system is opearting
	// ie, if it has failed to read the initial file it should not pass a health check
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

// Watch is the streaming healthcheck endpoint
func (wt *WatchTower) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	logrus.Debug("Serving the Watch request for health check")

	return server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}
