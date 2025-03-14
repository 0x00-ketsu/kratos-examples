package tracingx

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// GetTraceID returns the traceID from the context.
func GetTraceID(ctx context.Context) string {
	var traceID string
	if span := trace.SpanContextFromContext(ctx); span.HasTraceID() {
		traceID = span.TraceID().String()
	}
	return traceID
}
