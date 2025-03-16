package server

import (
	"context"
	"fmt"
	"usermanage/gen/proto/conf"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var maskedOperations = []string{"/auth.v1.AuthService/Login"}

func generateMaskedOperations(c *conf.Server) []string {
	if c.Debug {
		return []string{}
	}
	return []string{
		"/auth.v1.AuthService/Login",
		"/auth.v1.AuthService/RegisterRequest",
		"/auth.v1.AuthService/ChangePasswordRequest",
	}
}

// Initialize opentelemetry trace provider.
func initTracer(ctx context.Context, c *conf.Server) error {
	var exporter sdktrace.SpanExporter
	var err error

	meta := c.Metadata
	telemetry := c.Telemetry
	if telemetry.OutputToConsole {
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	} else {
		grpcEndpoint := telemetry.Otlp.GrpcEndpoint
		if grpcEndpoint == "" {
			return fmt.Errorf("grpcEndpoint must be provided when debug is false")
		}

		opts := []otlptracegrpc.Option{
			otlptracegrpc.WithEndpoint(grpcEndpoint),
		}
		if telemetry.Otlp.Insecure {
			opts = append(opts, otlptracegrpc.WithInsecure())
		}
		exporter, err = otlptracegrpc.New(ctx, opts...)
	}
	if err != nil {
		return fmt.Errorf("failed to create exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(meta.Name),
				semconv.ServiceVersionKey.String(meta.Version),
				attribute.String("environment", meta.Env.String()),
			),
		),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	go func() {
		<-ctx.Done()
		if err := tp.Shutdown(ctx); err != nil {
			fmt.Printf("failed to shutdown trace provider: %v\n", err)
		}
	}()

	return nil
}
