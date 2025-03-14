package middleware

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http/status"
	"google.golang.org/grpc/codes"
)

// Redacter defines how to log an object
type Redacter interface {
	Redact() string
}

// Logging extends the go-kratos built-in middleware `logging.Server`.
// It enhances logging capabilities by adding request ID support and allowing selective argument redaction.
//
// Parameters:
// - logger: The logger instance used for logging request information.
// - maskedOperations: A list of operation names where request arguments (`args`) should be redacted for security reasons.
//
// Example usage:
//
//	logger := log.NewStdLogger(os.Stdout)
//	middleware := Logging(logger, []string{"auth.Login", "user.UpdatePassword"})
//
// In this example, the `args` field in logs for "auth.Login" and "user.UpdatePassword" operations
// will be replaced with "***" to prevent sensitive data exposure.
func Logging(logger log.Logger, maskedOperations ...string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (reply any, err error) {
			var (
				code      int32
				reason    string
				kind      string
				operation string
			)

			// default code
			code = int32(status.FromGRPCCode(codes.OK))

			startTime := time.Now()
			if tr, ok := transport.FromServerContext(ctx); ok {
				kind = tr.Kind().String()
				operation = tr.Operation()
			}
			reply, err = handler(ctx, req)
			if se := errors.FromError(err); se != nil {
				code = se.Code
				reason = se.Reason
			}
			level, stack := extractError(err)
			args := extractArgs(req)
			if slices.Contains(maskedOperations, operation) {
				args = "***"
			}
			log.NewHelper(log.WithContext(ctx, logger)).Log(level,
				"kind", "server",
				"component", kind,
				"operation", operation,
				"args", args,
				"code", code,
				"reason", reason,
				"stack", stack,
				"latency", time.Since(startTime).Seconds(),
			)
			return
		}
	}
}

// extractArgs returns the string of the req
func extractArgs(req any) string {
	if redacter, ok := req.(Redacter); ok {
		return redacter.Redact()
	}
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	return fmt.Sprintf("%+v", req)
}

// extractError returns the string of the error
func extractError(err error) (log.Level, string) {
	if err != nil {
		return log.LevelError, fmt.Sprintf("%+v", err)
	}
	return log.LevelInfo, ""
}
