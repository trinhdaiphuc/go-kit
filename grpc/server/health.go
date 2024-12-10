package grpcserver

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Service interface {
	Check(ctx context.Context) error
}

type HealthController struct {
	healthService Service
}

var (
	_ grpc_health_v1.HealthServer = (*HealthController)(nil)
)

func (h *HealthController) Check(ctx context.Context, request *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	if h.healthService == nil {
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_SERVING,
		}, nil
	}

	err := h.healthService.Check(ctx)
	if err != nil {
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, nil
	}

	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

func (h *HealthController) Watch(request *grpc_health_v1.HealthCheckRequest, g grpc.ServerStreamingServer[grpc_health_v1.HealthCheckResponse]) error {
	return nil
}

func NewHealthController(healthService Service) *HealthController {
	return &HealthController{healthService: healthService}
}
