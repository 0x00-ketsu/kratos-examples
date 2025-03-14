package service

import (
	"context"
	"time"
	healthv1 "usermanage/gen/proto/api/health/v1"
	"usermanage/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HealthService struct {
	healthv1.UnimplementedHealthServiceServer
	uc  *biz.HealthUseCase
	log *log.Helper
}

func NewHealthService(uc *biz.HealthUseCase, logger log.Logger) *HealthService {
	return &HealthService{uc: uc, log: log.NewHelper(logger)}
}

func (s *HealthService) Probe(ctx context.Context, _ *emptypb.Empty) (*healthv1.ProbeResponse, error) {
	return &healthv1.ProbeResponse{
		Message: "success",
	}, nil
}

func (s *HealthService) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	logger := s.log.WithContext(ctx)
	startTime := time.Now()

	logger.Info("health check started")
	// Check database
	dbStart := time.Now()
	if err := s.uc.PingDB(ctx); err != nil {
		logger.Errorw(
			"msg", "database connection ping failed",
			"error", err,
			"duration_ms", time.Since(dbStart).Milliseconds(),
		)
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, nil
	}
	logger.Infow(
		"msg", "database connection check passed",
		"duration_ms", time.Since(dbStart).Milliseconds(),
	)

	// Check Redis
	redisStart := time.Now()
	if err := s.uc.PingRedis(ctx); err != nil {
		logger.Errorw(
			"msg", "redis connection ping failed",
			"error", err,
			"duration_ms", time.Since(redisStart).Milliseconds(),
		)
		return &grpc_health_v1.HealthCheckResponse{
			Status: grpc_health_v1.HealthCheckResponse_NOT_SERVING,
		}, nil
	}
	logger.Infow(
		"msg", "redis connection check passed",
		"duration_ms", time.Since(redisStart).Milliseconds(),
	)

	logger.Infow(
		"msg", "health check completed",
		"total_duration_ms", time.Since(startTime).Milliseconds(),
		"status", "serving",
	)
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}
