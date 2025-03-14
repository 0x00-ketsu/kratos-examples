package server

import (
	"fmt"
	"io"
	authv1 "usermanage/gen/proto/api/auth/v1"
	healthv1 "usermanage/gen/proto/api/health/v1"
	userv1 "usermanage/gen/proto/api/user/v1"
	"usermanage/gen/proto/conf"
	"usermanage/internal/biz"
	"usermanage/internal/pkg/middleware"
	"usermanage/internal/service"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
)

var Name string

// NewGRPCServer new a gRPC server.
func NewGRPCServer(
	c *conf.Server,
	health *service.HealthService,
	user *service.UserService,
	auth *service.AuthService,
	authUseCase *biz.AuthUseCase,
	logger log.Logger,
) *grpc.Server {
	// TODO: change stdout to file or others
	exporter, err := stdouttrace.New(stdouttrace.WithWriter(io.Discard))
	if err != nil {
		fmt.Printf("creating stdout exporter: %v", err)
		panic(err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewSchemaless(
			semconv.ServiceNameKey.String(Name)),
		))
	opts := []grpc.ServerOption{
		grpc.Middleware(
			recovery.Recovery(),
			tracing.Server(tracing.WithTracerProvider(tp)),
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
