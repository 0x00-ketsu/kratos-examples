package jwt

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/transport"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const (
	// ClaimsKey is the context key for JWT claims
	ClaimsKey contextKey = "jwt_claims"
)

const (
	// AuthorizationHeader is the header key for authentication
	AuthorizationHeader = "Authorization"
	// BearerPrefix is the prefix for bearer tokens
	BearerPrefix = "Bearer "
)

var (
	jwtKey              []byte
	tokenExpireDuration time.Duration
)

// Initialize initializes the JWT package.
//
// key is the secret key used to sign the JWT token.
// expireduration is the duration after which the token will expire.
//
// # NOTE: This function must be called before any other function in this package.
func Initialize(key []byte, expireDuration time.Duration) error {
	if len(key) == 0 {
		return errors.New("JWT key cannot be empty")
	}
	if expireDuration <= 0 {
		return errors.New("token expire duration must be positive")
	}

	jwtKey = key
	tokenExpireDuration = expireDuration
	return nil
}

// TokenExpireDuration returns the token expire duration.
func TokenExpireDuration() time.Duration {
	return tokenExpireDuration
}

// Claims represents the claims in a JWT token.
type Claims struct {
	Username string `json:"username"`
	role     int32
	jwt.RegisteredClaims
}

// Role sets or returns the role of the user.
func (c *Claims) Role(role ...int32) int32 {
	if len(role) > 0 {
		c.role = role[0]
	}
	return c.role
}

// GenerateToken generates a JWT token for a user.
func GenerateToken(username string) (tokenString string, expiresAt time.Time, err error) {
	if len(strings.TrimSpace(username)) == 0 {
		return "", time.Time{}, errors.New("username is required")
	}

	now := time.Now()
	expiresAt = now.Add(tokenExpireDuration)
	claims := Claims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(jwtKey)
	return tokenString, expiresAt, err
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}

// ExtractToken extracts the token from both HTTP headers and gRPC metadata.
func ExtractToken(ctx context.Context) (token string, err error) {
	// Try to get token from gRPC metadata
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if tokens := md.Get(AuthorizationHeader); len(tokens) > 0 {
			return strings.TrimPrefix(tokens[0], BearerPrefix), nil
		}
	}

	// try to get token from HTTP tr
	if tr, ok := transport.FromServerContext(ctx); ok {
		if auth := tr.RequestHeader().Get(AuthorizationHeader); auth != "" {
			return strings.TrimPrefix(auth, BearerPrefix), nil
		}
	}

	return "", errors.New("token not found")
}

// WithContext returns a new context with the given username.
func WithContext(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ClaimsKey, claims)
}

// FromContext retrieves the username from the context.
// Returns empty string if username is not found in context.
func FromContext(ctx context.Context) (claims *Claims, ok bool) {
	claims, ok = ctx.Value(ClaimsKey).(*Claims)
	return
}
