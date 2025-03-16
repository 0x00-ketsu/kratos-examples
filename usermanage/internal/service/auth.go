package service

import (
	"context"
	authv1 "usermanage/gen/proto/api/auth/v1"
	userv1 "usermanage/gen/proto/api/user/v1"
	"usermanage/internal/biz"
	"usermanage/internal/pkg/auth"
	"usermanage/internal/pkg/jwt"
	"usermanage/internal/pkg/tracingx"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// AtuhService is a service for authentication.
type AuthService struct {
	authv1.UnimplementedAuthServiceServer
	uc  *biz.AuthUseCase
	log *log.Helper
}

// NewAuthService creates a new auth service.
func NewAuthService(uc *biz.AuthUseCase, logger log.Logger) *AuthService {
	return &AuthService{uc: uc, log: log.NewHelper(logger)}
}

// Login logs in a user.
func (s *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := s.validate(ctx, req); err != nil {
		return nil, err
	}

	logger.Info("auth login")
	user, token, expiresAt, err := s.uc.Login(ctx, req.Username, req.Password)
	if err != nil {
		logger.Errorw("msg", "failed to login", "error", err)
		err = errors.Unauthorized("LOGIN_FAILED", "Failed to login").
			WithMetadata(md)
		return nil, err
	}
	logger.Infow("msg", "token generated", "user.name", user.Username)

	return &authv1.LoginResponse{Token: token, ExpiresAt: timestamppb.New(expiresAt)}, nil
}

// Logout logs out a user.
func (s *AuthService) Logout(ctx context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	token, err := jwt.ExtractToken(ctx)
	if err != nil {
		logger.Errorw("msg", "failed to extract token", "error", err)
		err = errors.InternalServer("EXTRACT_TOKEN_FAILED", "Failed to extract token").
			WithMetadata(md)
		return nil, err
	}

	username, err := s.uc.Logout(ctx, token)
	if err != nil {
		logger.Errorw("msg", "failed to logout", "error", err)
		err = errors.InternalServer("LOGOUT_FAILED", "Failed to logout").
			WithMetadata(md)
		return nil, err
	}
	logger.Infow("msg", "logout success", "user.name", username)

	return &emptypb.Empty{}, nil
}

// GetUserInfo gets a user's information.
func (s *AuthService) GetUserInfo(ctx context.Context, req *authv1.UserInfoRequest) (*authv1.UserInfoResponse, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := s.validate(ctx, req); err != nil {
		return nil, err
	}

	logger.Info("auth get user info")
	user, err := s.uc.VerifyToken(ctx, req.Token)
	if err != nil {
		logger.Errorw("msg", "failed to verify token", "error", err)
		err = errors.Unauthorized("INVALID_TOKEN", "Invalid token").WithMetadata(md)
		return nil, err
	}
	logger.Infow("msg", "successfully get user info", "user.name", user.Username)

	resp := &authv1.UserInfoResponse{
		Id:        user.ID,
		Username:  user.Username,
		Role:      userv1.UserRole(user.Role),
		Status:    userv1.UserStatus(user.Status),
		Creator:   user.Creator,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedBy: user.UpdatedBy,
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
	return resp, nil
}

// ChangePassword changes a user's password.
func (s *AuthService) ChangePassword(ctx context.Context, req *authv1.ChangePasswordRequest) (*emptypb.Empty, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := s.validate(ctx, req); err != nil {
		return nil, err
	}

	username := auth.Username(ctx)
	if username == "" {
		logger.Errorw("msg", "failed to get current username")
		err := errors.InternalServer("GET_CURRENT_USERNAME_FAILED", "Failed to get current username").
			WithMetadata(md)
		return nil, err
	}

	err := s.uc.ChangePassword(ctx, username, req.OldPassword, req.NewPassword)
	if err != nil {
		logger.Errorw("msg", "failed to change password", "error", err)
		err := errors.InternalServer("CHANGE_PASSWORD_FAILED", "Failed to change password").
			WithMetadata(md)
		return nil, err
	}
	logger.Infow("msg", "password changed successfully", "user", username)

	return &emptypb.Empty{}, nil
}

// A helper method to validate a request.
func (s *AuthService) validate(ctx context.Context, req interface{ Validate() error }) error {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := req.Validate(); err != nil {
		logger.Errorw("msg", "invalid request", "error", err)
		return errors.BadRequest("INVALID_REQUEST", "Invalid request").
			WithMetadata(md)
	}
	return nil
}
