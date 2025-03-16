package server

import (
	"context"
	authv1 "usermanage/gen/proto/api/auth/v1"
	healthv1 "usermanage/gen/proto/api/health/v1"
	userv1 "usermanage/gen/proto/api/user/v1"
	"usermanage/gen/proto/conf"
	"usermanage/internal/biz"
	"usermanage/internal/pkg/middleware"
	"usermanage/internal/service"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	ctx context.Context,
	c *conf.Server,
	health *service.HealthService,
	user *service.UserService,
	auth *service.AuthService,
	authUseCase *biz.AuthUseCase,
	logger log.Logger,
) *grpc.Server {
	if err := initTracer(ctx, c); err != nil {
		panic(err)
	}

	opts := []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			middleware.Logging(logger, generateMaskedOperations(c)...),
			middleware.JWTAuth(authUseCase),
		),
	}
	if c.Grpc.Network != "" {
		opts = append(opts, grpc.Network(c.Grpc.Network))
	}
	if c.Grpc.Addr != "" {
		opts = append(opts, grpc.Address(c.Grpc.Addr))
	}
	if c.Grpc.Timeout != nil {
		opts = append(opts, grpc.Timeout(c.Grpc.Timeout.AsDuration()))
	}
	srv := grpc.NewServer(opts...)
	healthv1.RegisterHealthServiceServer(srv, health)
	userv1.RegisterUserServiceServer(srv, user)
	authv1.RegisterAuthServiceServer(srv, auth)
	return srv
}
