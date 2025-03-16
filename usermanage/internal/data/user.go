package data

import (
	"context"
	"errors"
	"fmt"
	"time"
	"usermanage/internal/biz"
	"usermanage/internal/data/model"
	"usermanage/internal/pkg/constants"
	"usermanage/internal/pkg/db"
	"usermanage/internal/pkg/password"

	"github.com/go-kratos/kratos/v2/log"
)

type userRepo struct {
	db     *db.Database
	logger *log.Helper
}

// NewUserRepo creates a new user repository.
func NewUserRepo(db *db.Database, logger log.Logger) biz.UserRepo {
	return &userRepo{
		db:     db,
		logger: log.NewHelper(logger),
	}
}

// GetUserByID implements user.UserRepo.
func (r *userRepo) GetUserByID(ctx context.Context, id string) (*biz.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id[%s]: %w", id, err)
	}
	return r.toBizUser(&user), nil
}

// GetUserByUsername implements biz.UserRepo.
func (r *userRepo) GetUserByUsername(ctx context.Context, username string) (*biz.User, error) {
	user := model.User{}
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username[%s]: %w", username, err)
	}
	return r.toBizUser(&user), nil
}

// ExistsByID implements biz.UserRepo.
func (r *userRepo) ExistsByID(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Select("COUNT(1) > 0").
		Where("id = ?", id).
		Find(&exists).Error
	if err != nil {
		return false, fmt.Errorf("failed to check user exists by id[%s]: %w", id, err)
	}
	return exists, nil
}

// ExistsByUsername implements biz.UserRepo.
func (r *userRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.db.WithContext(ctx).
		Model(&model.User{}).
		Select("COUNT(1) > 0").
		Where("username = ?", username).
		Find(&exists).Error
	if err != nil {
		return false, fmt.Errorf("failed to check user exists by username[%s]: %w", username, err)
	}
	return exists, nil
}

// FindByCredentials implements biz.UserRepo.
func (r *userRepo) FindByCredentials(ctx context.Context, username string, rawPassword string) (*biz.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("username = ?", username).
		First(&user).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username[%s]: %w", username, err)
	}

	if !user.VerifyPassword(rawPassword) {
		return nil, errors.New("invalid password")
	}

	return r.toBizUser(&user), nil
}

// DeleteUser implements biz.UserRepo.
func (r *userRepo) DeleteUser(ctx context.Context, id string) error {
	var user model.User
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&user).Error; err != nil {
		return fmt.Errorf("failed to delete user by id[%s]: %w", id, err)
	}
	return nil
}

// ListUsers implements biz.UserRepo.
func (r *userRepo) ListUsers(ctx context.Context, params biz.UserListParams) (*biz.UserListResult, error) {
	var totalCount int64
	var users []model.User

	// TODO: add query conditions

	query := r.db.WithContext(ctx).Model(&model.User{})
	if err := query.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	page, pageSize := params.GetPage()
	offset := (page - 1) * pageSize
	if err := query.
		Offset(int(offset)).
		Limit(int(params.PageSize)).
		Order("created_at DESC").
		Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to find users: %w", err)
	}

	bizUsers := make([]*biz.User, 0, len(users))
	for _, user := range users {
		bizUsers = append(bizUsers, r.toBizUser(&user))
	}
	return &biz.UserListResult{
		TotalCount: totalCount,
		Users:      bizUsers,
	}, nil
}

// CreateUser implements biz.UserRepo.
func (r *userRepo) CreateUser(ctx context.Context, params biz.UserCreateParams) (*biz.User, error) {
	username := params.Username
	exists, err := r.ExistsByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check user exists by username[%s]: %w", username, err)
	}
	if exists {
		return nil, fmt.Errorf("username[%s] already exists", username)
	}

	user := model.User{
		Username:  username,
		Password:  params.Password,
		Role:      constants.UserRole(params.Role),
		Status:    constants.UserStatus(params.Status),
		Creator:   params.Creator,
		UpdatedBy: params.UpdateBy,
	}
	if err := r.db.WithContext(ctx).
		Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return r.toBizUser(&user), nil
}

// UpdateUser implements biz.UserRepo.
func (r *userRepo) UpdateUser(ctx context.Context, id string, params biz.UserUpdateParams) (*biz.User, error) {
	// Check if the user exists
	existingUser, err := r.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id[%s]: %w", id, err)
	}

	updateUsername := params.Username
	if updateUsername != nil {
		if *updateUsername == existingUser.Username {
			params.Username = nil
		} else {
			exists, err := r.ExistsByUsername(ctx, *updateUsername)
			if err != nil {
				return nil, fmt.Errorf("failed to check user exists by username[%s]: %w", *updateUsername, err)
			}
			if exists {
				return nil, fmt.Errorf("username[%s] already exists", *updateUsername)
			}
		}
	}

	// Update the user
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Updates(params).
		Update("updated_at", time.Now())
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update user by id[%s]: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("user[id=%s] not found", id)
	}

	// Fetch the updated user
	user, err := r.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id[%s]: %w", id, err)
	}
	return user, nil
}

// ReplaceUser implements biz.UserRepo.
func (r *userRepo) ReplaceUser(ctx context.Context, id string, params biz.UserReplaceParams) (*biz.User, error) {
	// Check if the user exists
	replaceUsername := params.Username
	existingUser, err := r.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id[%s]: %w", id, err)
	}
	if replaceUsername != existingUser.Username {
		exists, err := r.ExistsByUsername(ctx, replaceUsername)
		if err != nil {
			return nil, fmt.Errorf("failed to check user exists by username[%s]: %w", replaceUsername, err)
		}
		if exists {
			return nil, fmt.Errorf("username[%s] already exists", replaceUsername)
		}
	}

	// Replace the user
	result := r.db.WithContext(ctx).
		Model(&model.User{}).
		Where("id = ?", id).
		Updates(params).
		Update("updated_at", time.Now())
	if result.Error != nil {
		return nil, fmt.Errorf("failed to replace user by id[%s]: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("user[id=%s] not found", id)
	}

	// Fetch the replaced user
	user, err := r.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by id[%s]: %w", id, err)
	}
	return user, nil
}

// ResetUserPassword implements biz.UserRepo.
func (r *userRepo) ResetUserPassword(ctx context.Context, id string, newPassword string) (*biz.User, error) {
	// Hash the password before saving
	hashedPassword, err := password.Hash(newPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var user model.User
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find user by id[%s]: %w", id, result.Error)
	}

	user.Password = hashedPassword
	user.UpdatedAt = time.Now()
	if err := r.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user password: %w", err)
	}
	return r.toBizUser(&user), nil
}

// VerifyPassword implements biz.UserRepo.
func (r *userRepo) VerifyPassword(ctx context.Context, id string, password string) (bool, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user).Error
	if err != nil {
		return false, fmt.Errorf("failed to find user by id[%s]: %w", id, err)
	}
	return user.VerifyPassword(password), nil
}

// Convert model User to biz User.
func (r *userRepo) toBizUser(u *model.User) *biz.User {
	if u == nil {
		return nil
	}

	return &biz.User{
		ID:        u.ID,
		Username:  u.Username,
		Role:      biz.UserRole(u.Role),
		Status:    biz.UserStatus(u.Status),
		Creator:   u.Creator,
		CreatedAt: u.CreatedAt,
		UpdatedBy: u.UpdatedBy,
		UpdatedAt: u.UpdatedAt,
	}
}
