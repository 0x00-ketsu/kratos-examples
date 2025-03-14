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
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(
	c *conf.Server,
	health *service.HealthService,
	user *service.UserService,
	auth *service.AuthService,
	authUseCase *biz.AuthUseCase,
	logger log.Logger,
) *http.Server {
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
	opts := []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(tracing.WithTracerProvider(tp)),
			middleware.Logging(logger, generateMaskedOperations(c)...),
			middleware.JWTAuth(authUseCase),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	healthv1.RegisterHealthServiceHTTPServer(srv, health)
	userv1.RegisterUserServiceHTTPServer(srv, user)
	authv1.RegisterAuthServiceHTTPServer(srv, auth)
	return srv
}
