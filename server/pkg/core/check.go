package core

import (
	"context"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// Check is the health check endpoint
func (s *server) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	logrus.Info("Serving the Check request for health check")

	errMain := s.MainReplica.Ping()
	errBackup := s.BackupReplica.Ping()
	if errMain != nil || errBackup != nil {
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, status.Error(codes.Unavailable, "replicas not serving")
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

// Watch is the streaming healthcheck endpoint
func (s *server) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	logrus.Debug("Serving the Watch request for health check")
	errMain := s.MainReplica.Ping()
	errBackup := s.BackupReplica.Ping()
	if errMain != nil || errBackup != nil {
		return server.Send(&grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		})
	}

	return server.Send(&grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	})
}
