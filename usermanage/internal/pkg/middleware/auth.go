package middleware

import (
	"context"
	"strings"
	"time"
	userv1 "usermanage/gen/proto/api/user/v1"
	"usermanage/internal/biz"
	"usermanage/internal/pkg/auth"
	"usermanage/internal/pkg/jwt"
	"usermanage/internal/pkg/tracingx"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

var skipAuthPaths = map[string]bool{
	// health
	"/health.v1.HealthService/Probe": true,
	"/health.v1.HealthService/Check": true,

	// auth
	"/auth.v1.AuthService/Login":       true,
	"/auth.v1.AuthService/GetUserInfo": true,
}

// JWTAuth is a middleware that authenticates the user using JWT.
func JWTAuth(authUseCase *biz.AuthUseCase) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			logger := log.WithContext(ctx, log.GetLogger())
			md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

			if tr, ok := transport.FromServerContext(ctx); ok {
				operation := tr.Operation()
				if skipAuthPaths[operation] {
					return handler(ctx, req)
				}

				// Extract the JWT token from the request
				token := tr.RequestHeader().Get("Authorization")
				if token == "" {
					logger.Log(log.LevelError, "msg", "missing Authorization header")
					err := errors.Unauthorized("MISSING_TOKEN", "Missing token").
						WithMetadata(md)
					return nil, err
				}

				token = strings.TrimPrefix(token, "Bearer ")
				claims, err := jwt.ParseToken(token)
				if err != nil {
					logger.Log(log.LevelError, "msg", "failed to parse token", "error", err)
					err = errors.Unauthorized("INVALID_TOKEN", "Invalid or expired token").
						WithMetadata(md)
					return nil, err
				}

				// Check if the token exists
				if exists, err := authUseCase.TokenExists(ctx, token); err != nil || !exists {
					logger.Log(log.LevelError, "msg", "failed to check token", "error", err, "exists", exists)
					err = errors.Unauthorized("INVALID_TOKEN", "Invalid or expired token").
						WithMetadata(md)
					return nil, err
				}

				user, err := authUseCase.GetUserByUsername(ctx, claims.Username)
				if err != nil {
					logger.Log(log.LevelError, "msg", "failed to get user by username", "error", err)
					err = errors.InternalServer("GET_USER_BY_USERNAME", "Failed to get user by username").
						WithMetadata(md)
					return nil, err
				}

				// Verify user status
				if !user.Status.IsNormal() {
					logger.Log(log.LevelError, "msg", "invalid user status", "status", user.Status)
					err = errors.Forbidden("INVALID_USER_STATUS", "Invalid user status").
						WithMetadata(md)
					return nil, err
				}

				// Set the user role in the token claims
				claims.Role(int32(user.Role))
				ctx = jwt.WithContext(ctx, claims)
				resp, err := handler(ctx, req)
				if shouldExtendToken(claims) {
					if err := authUseCase.ExtendTokenExpiry(ctx, token); err != nil {
						logger.Log(log.LevelError, "msg", "failed to extend token expiry", "error", err)
						err = errors.InternalServer("EXTEND_TOKEN_EXPIRY", "Failed to extend token expiry").
							WithMetadata(md)
						return nil, err
					}
				}

				return resp, err
			}
			return handler(ctx, req)
		}
	}
}

// RequireRole is a middleware that checks if the user has the required role.
func RequireRole(roles ...userv1.UserRole) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}
			if err := auth.CheckRole(ctx, roles...); err != nil {
				err := errors.Forbidden("INSUFFICIENT_PERMISSIONS", "Insufficient permissions").
					WithMetadata(md)
				return nil, err
			}
			return handler(ctx, req)
		}
	}
}

// Check if the token should be extended.
//
// The token should be extended if the time until the token expires is less than 30% of the total lifetime.
func shouldExtendToken(claims *jwt.Claims) bool {
	if claims.ExpiresAt == nil || claims.IssuedAt == nil {
		return false
	}

	expiresAt := claims.ExpiresAt.Time
	issuedAt := claims.IssuedAt.Time
	totalLifetime := expiresAt.Sub(issuedAt)
	threshold := totalLifetime * 30 / 100
	return time.Until(expiresAt) < threshold
}
