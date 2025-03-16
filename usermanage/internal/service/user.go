package service

import (
	"context"
	commonv1 "usermanage/gen/proto/api/common/v1"
	userv1 "usermanage/gen/proto/api/user/v1"
	"usermanage/internal/biz"
	"usermanage/internal/pkg/auth"
	"usermanage/internal/pkg/constants"
	"usermanage/internal/pkg/tracingx"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	userv1.UnimplementedUserServiceServer
	uc  *biz.UserUseCase
	log *log.Helper
}

// NewUserService creates a new user service.
func NewUserService(uc *biz.UserUseCase, logger log.Logger) *UserService {
	return &UserService{uc: uc, log: log.NewHelper(logger)}
}

// ListUsers lists users.
func (s *UserService) ListUsers(ctx context.Context, req *userv1.UserListRequest) (*userv1.UserListResponse, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := s.validateAdminAndRequest(ctx, req); err != nil {
		return nil, err
	}

	logger.Info("list users")
	page := func() int32 {
		if req.Page == 0 {
			return constants.DefaultPage
		}
		return req.Page
	}()
	pageSize := func() int32 {
		if req.PageSize == 0 {
			return constants.DefaultPageSize
		}
		return req.PageSize
	}()
	params := biz.UserListParams{
		Page:      page,
		PageSize:  pageSize,
		Username:  req.Username,
		Status:    int32(req.Status),
		SortBy:    req.SortBy,
		SortOrder: req.SortOrder,
	}
	logger.Infow("msg", "list users", "params", params.String())
	result, err := s.uc.ListUsers(ctx, params)
	if err != nil {
		logger.Errorw("msg", "failed to list users", "error", err)
		err = errors.InternalServer("LIST_USERS_FAILED", "Failed to list users").
			WithMetadata(md)
		return nil, err
	}
	logger.Infow("msg", "list users", "total_count", result.TotalCount)

	pagination := commonv1.PageResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: result.TotalCount,
	}
	data := make([]*userv1.UserPublic, 0, len(result.Users))
	for _, user := range result.Users {
		data = append(data, s.toUserPublic(user))
	}
	return &userv1.UserListResponse{Data: data, Pagination: &pagination}, nil
}

// GetUser gets a user by ID.
func (s *UserService) GetUser(ctx context.Context, req *userv1.UserRequest) (*userv1.UserResponse, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := s.validateAdminAndRequest(ctx, req); err != nil {
		return nil, err
	}

	userID := req.Id
	logger.Infow("msg", "get user", "user.id", userID)
	user, err := s.uc.GetUser(ctx, userID)
	if err != nil {
		logger.Errorw("msg", "failed to get user", "error", err)
		err = errors.InternalServer("GET_USER_FAILED", "Failed to get user").
			WithMetadata(md)
		return nil, err
	}
	return &userv1.UserResponse{Data: s.toUserPublic(user)}, nil
}

// CreateUser creates a user.
func (s *UserService) CreateUser(ctx context.Context, req *userv1.UserCreateRequest) (*userv1.UserResponse, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := s.validateAdminAndRequest(ctx, req); err != nil {
		return nil, err
	}

	params := biz.UserCreateParams{
		Username: req.Username,
		Role:     int32(req.Role),
		Status:   int32(req.Status),
		Creator:  auth.Username(ctx),
		UpdateBy: auth.Username(ctx),
	}
	logger.Infow("msg", "create user", "params", params.String())
	user, err := s.uc.CreateUser(ctx, params)
	if err != nil {
		logger.Errorw("msg", "failed to create user", "error", err)
		err = errors.InternalServer("CREATE_USER_FAILED", "Failed to create user").
			WithMetadata(md)
		return nil, err
	}
	logger.Info("successfully create user")
	return &userv1.UserResponse{Data: s.toUserPublic(user)}, nil
}

// UpdateUser performs a partial update on a user resource using the provided field mask.
// It only updates the fields specified in the UpdateMask of the request.
// If no `UpdateMask` is provided or it's empty, returns an InvalidArgument error.
func (s *UserService) UpdateUser(ctx context.Context, req *userv1.UserUpdateRequest) (*userv1.UserResponse, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := s.validateAdminAndRequest(ctx, req); err != nil {
		return nil, err
	}

	targetUserID := req.Id
	logger.Infow("msg", "partially update user", "target_user.id", targetUserID)
	if targetUserID == "" {
		logger.Error("user id is required")
		err := errors.BadRequest("USER_ID_REQUIRED", "User ID is required").
			WithMetadata(md)
		return nil, err
	}

	params := biz.UserUpdateParams{
		UpdatedBy: auth.Username(ctx),
	}
	for _, field := range req.UpdateMask.Paths {
		switch field {
		case "username":
			params.Username = &req.Username
		case "role":
			params.Role = (*int32)(&req.Role)
		case "status":
			params.Status = (*int32)(&req.Status)
		default:
			logger.Errorw("msg", "invalid update field", "field", field, "target_user.id", targetUserID)
			err := errors.BadRequest("INVALID_UPDATE_FIELD", "Invalid update field").
				WithMetadata(md)
			return nil, err
		}
	}

	logger.Infow("msg", "update user", "target_user.id", targetUserID, "params", params.String())
	updatedUser, err := s.uc.UpdateUser(ctx, targetUserID, params)
	if err != nil {
		logger.Errorw("msg", "failed to update user", "error", err)
		err = errors.InternalServer("UPDATE_USER_FAILED", "Failed to update user").
			WithMetadata(md)
		return nil, err
	}
	logger.Info("successfully update user")
	return &userv1.UserResponse{Data: s.toUserPublic(updatedUser)}, nil
}

// ReplaceUser performs a full replacement of a user resource.
// Unlike `UpdateUser`, this method replaces the entire user resource with the provided data,
// regardless of which fields are present in the request.
func (s *UserService) ReplaceUser(ctx context.Context, req *userv1.UserReplaceRequest) (*userv1.UserResponse, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := s.validateAdminAndRequest(ctx, req); err != nil {
		return nil, err
	}

	targetUserID := req.Id
	logger.Infow("msg", "replace user", "target_user.id", targetUserID)
	if targetUserID == "" {
		logger.Error("user id is required")
		err := errors.BadRequest("USER_ID_REQUIRED", "User ID is required").
			WithMetadata(md)
		return nil, err
	}

	params := biz.UserReplaceParams{
		Username:  req.Username,
		Role:      int32(req.Role),
		Status:    int32(req.Status),
		UpdatedBy: auth.Username(ctx),
	}
	logger.Infow("msg", "replace user", "target_user.id", targetUserID, "params", params.String())
	replacedUser, err := s.uc.ReplaceUser(ctx, targetUserID, params)
	if err != nil {
		logger.Errorw("msg", "failed to replace user", "error", err)
		err = errors.InternalServer("REPLACE_USER_FAILED", "Failed to replace user").
			WithMetadata(md)
		return nil, err
	}
	logger.Info("successfully replace user")
	return &userv1.UserResponse{Data: s.toUserPublic(replacedUser)}, nil
}

// DeleteUser deletes a user.
func (s *UserService) DeleteUser(ctx context.Context, req *userv1.UserDeleteRequest) (*emptypb.Empty, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := s.validateAdminAndRequest(ctx, req); err != nil {
		return nil, err
	}

	targetUserID := req.Id
	logger.Infow("msg", "delete user", "target_user.id", targetUserID)
	if err := s.uc.DeleteUser(ctx, targetUserID); err != nil {
		logger.Errorw("msg", "failed to delete user", "error", err)
		err = errors.InternalServer("DELETE_USER_FAILED", "Failed to delete user").
			WithMetadata(md)
		return nil, err
	}
	logger.Infow("msg", "successfully delete user", "target_user.id", targetUserID)
	return &emptypb.Empty{}, nil
}

// ResetUserPassword resets the user password.
func (s *UserService) ResetUserPassword(ctx context.Context, req *userv1.UserPasswordResetRequest) (*emptypb.Empty, error) {
	logger := s.log.WithContext(ctx)
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}

	if err := s.validateAdminAndRequest(ctx, req); err != nil {
		return nil, err
	}

	targetUserID := req.Id
	logger.Infow("msg", "reset user password", "target_user.id", targetUserID)
	user, err := s.uc.ResetUserPassword(ctx, targetUserID, req.NewPassword)
	if err != nil {
		logger.Errorw("msg", "failed to reset user password", "error", err)
		err = errors.InternalServer("RESET_USER_PASSWORD_FAILED", "Failed to reset user password").
			WithMetadata(md)
		return nil, err
	}
	logger.Infow("msg", "successfully reset user password", "target_user.id", targetUserID, "target_user.name", user.Username)
	return &emptypb.Empty{}, nil
}

// Validate admin permissions and request validation.
func (s *UserService) validateAdminAndRequest(ctx context.Context, req interface{ Validate() error }) error {
	md := map[string]string{"traceId": tracingx.GetTraceID(ctx)}
	logger := s.log.WithContext(ctx)

	if err := auth.IsAdmin(ctx); err != nil {
		logger.Errorw("msg", "insufficient permissions", "error", err)
		return errors.Forbidden("INSUFFICIENT_PERMISSIONS", "Insufficient permissions").
			WithMetadata(md)
	}

	if err := req.Validate(); err != nil {
		logger.Errorw("msg", "invalid request", "error", err)
		return errors.BadRequest("INVALID_REQUEST", "Invalid request").
			WithMetadata(md)
	}

	return nil
}

// Convert biz user to user public.
func (s *UserService) toUserPublic(u *biz.User) *userv1.UserPublic {
	if u == nil {
		return nil
	}

	return &userv1.UserPublic{
		Id:        u.ID,
		Username:  u.Username,
		Role:      userv1.UserRole(u.Role),
		Status:    userv1.UserStatus(u.Status),
		Creator:   u.Creator,
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedBy: u.UpdatedBy,
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}
