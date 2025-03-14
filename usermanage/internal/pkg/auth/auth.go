package auth

import (
	"context"
	"errors"
	"fmt"
	userv1 "usermanage/gen/proto/api/user/v1"
	"usermanage/internal/pkg/jwt"
)

// CurrentUsername returns the username of the current user from the context.
func CurrentUsername(ctx context.Context) string {
	if claims, ok := jwt.FromContext(ctx); ok {
		return claims.Username
	}
	return ""
}

// IsAdmin checks if the current user is an admin.
func IsAdmin(ctx context.Context) error {
	return CheckRole(ctx, userv1.UserRole_ADMIN)
}

// IsUser checks if the current user is a user.
func IsUser(ctx context.Context) error {
	return CheckRole(ctx, userv1.UserRole_USER)
}

// HasAnyRole checks if the current user has any of the required roles.
func HasAnyRole(ctx context.Context, roles ...userv1.UserRole) error {
	return CheckRole(ctx, roles...)
}

// CheckRole checks if the current user has the required role.
//
// Extract the token from the context and check if the role is in the required roles.
func CheckRole(ctx context.Context, roles ...userv1.UserRole) error {
	tokenClaims, ok := jwt.FromContext(ctx)
	if !ok {
		return errors.New("missing token claims")
	}

	role := tokenClaims.Role()
	hasRole := false
	for _, r := range roles {
		if role == int32(r) {
			hasRole = true
			break
		}
	}
	if !hasRole {
		return fmt.Errorf("insufficient permissions: required=%v, current=%v", roles, role)
	}
	return nil
}
