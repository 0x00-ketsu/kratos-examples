package biz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"usermanage/internal/pkg/constants"
)

// UserRepo is the repository for user.
type UserRepo interface {
	// GetUserByID gets the user by ID.
	GetUserByID(ctx context.Context, id string) (*User, error)

	// GetUserByUsername gets the user by username.
	GetUserByUsername(ctx context.Context, username string) (*User, error)

	// ExistsByID checks if the user exists by ID.
	ExistsByID(ctx context.Context, id string) (bool, error)

	// ExistsByUsername checks if the username exists.
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// FindByCredentials finds a user by credentials.
	// Returns user info if credentials are valid, error otherwise
	//
	// # Note
	//
	// The `password` parameter is plaintext
	FindByCredentials(ctx context.Context, username, password string) (*User, error)

	// ListUsers returns a paginated list of users.
	ListUsers(ctx context.Context, params UserListParams) (*UserListResult, error)

	// CreateUser creates a new user.
	CreateUser(ctx context.Context, params UserCreateParams) (*User, error)

	// UpdateUser performs a partial update on a user resource using the provided field mask.
	UpdateUser(ctx context.Context, id string, params UserUpdateParams) (*User, error)

	// ReplaceUser performs a full replacement of a user resource.
	// Unlike `UpdateUser`, this method replaces the entire user resource with the provided data,
	// regardless of which fields are present in the request.
	ReplaceUser(ctx context.Context, id string, params UserReplaceParams) (*User, error)

	// DeleteUser deletes a user by ID.
	DeleteUser(ctx context.Context, id string) error

	// ResetUserPassword resets the user password.
	//
	// # Note
	//
	// The `newPassword` parameter is plaintext.
	ResetUserPassword(ctx context.Context, id, newPassword string) (*User, error)

	// VerifyPassword verifies if the provided password matches the user's password.
	//
	// # Note
	//
	// The `password` parameter is plaintext.
	VerifyPassword(ctx context.Context, id, password string) (bool, error)
}

// User is the user entity.
type User struct {
	ID        string     `json:"id"`
	Username  string     `json:"username"`
	Role      UserRole   `json:"role"`
	Status    UserStatus `json:"status"`
	Creator   string     `json:"creator"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedBy string     `json:"updatedBy"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// UserListParams represents all parameters for user listing
type UserListParams struct {
	Page      int32  `json:"page"`       // current page number (1-based)
	PageSize  int32  `json:"page_size"`  // page size
	Username  string `json:"username"`   // filter by username (optional)
	Status    int32  `json:"status"`     // filter by status (optional)
	SortBy    string `json:"sort_by"`    // sort field (optional)
	SortOrder string `json:"sort_order"` // sort direction: asc/desc (optional)
}

// String implements fmt.Stringer interface
func (p *UserListParams) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf("error marshaling UserListParams: %v", err)
	}
	return string(data)
}

// GetPage returns the page and size.
func (p *UserListParams) GetPage() (page, size int32) {
	if p.Page <= 0 {
		p.Page = constants.DefaultPage
	}
	if p.PageSize <= 0 {
		p.PageSize = constants.DefaultPageSize
	}
	maxPageSize := constants.MaxPageSize
	if p.PageSize > maxPageSize {
		p.PageSize = maxPageSize
	}
	return p.Page, p.PageSize
}

// UserListResult represents the result of user listing
type UserListResult struct {
	TotalCount int64
	Users      []*User
}

// UserCreateParams represents the parameters for creating a user.
type UserCreateParams struct {
	Username string `json:"username"`
	Password string `json:"-"`
	Role     int32  `json:"role"`
	Status   int32  `json:"status"`
	Creator  string `json:"creator"`
	UpdateBy string `json:"updated_by"`
}

// String implements fmt.Stringer interface
func (p *UserCreateParams) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf("error marshaling UserCreateParams: %v", err)
	}
	return string(data)
}

// Validate() validates the user create params.
func (p *UserCreateParams) Validate() error {
	if p.Username == "" {
		return errors.New("username is required")
	}
	if p.Password == "" {
		return errors.New("password is required")
	}
	if !UserRole(p.Role).IsValid() {
		return errors.New("invalid role")
	}
	if !UserStatus(p.Status).IsValid() {
		return errors.New("invalid status")
	}
	return nil
}

// UserUpdateParams represents the parameters for updating a user.
type UserUpdateParams struct {
	Username  *string `json:"username,omitempty"`
	Role      *int32  `json:"role,omitempty"`
	Status    *int32  `json:"status,omitempty"`
	UpdatedBy string  `json:"updated_by,omitempty"`
}

// String implements fmt.Stringer interface
func (p *UserUpdateParams) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf("error marshaling UserUpdateParams: %v", err)
	}
	return string(data)
}

// Validate() validates the user update params.
func (p *UserUpdateParams) Validate() error {
	if p.Username != nil && *p.Username == "" {
		return errors.New("username cannot be empty")
	}
	if p.Role != nil && !UserRole(*p.Role).IsValid() {
		return errors.New("invalid role")
	}
	if p.Status != nil && !UserStatus(*p.Status).IsValid() {
		return errors.New("invalid status")
	}
	return nil
}

// UserReplaceParams represents the parameters for replacing a user.
type UserReplaceParams struct {
	Username  string `json:"username"`
	Role      int32  `json:"role"`
	Status    int32  `json:"status"`
	UpdatedBy string `json:"updated_by"`
}

// String implements fmt.Stringer interface
func (p *UserReplaceParams) String() string {
	data, err := json.Marshal(p)
	if err != nil {
		return fmt.Sprintf("error marshaling UserReplaceParams: %v", err)
	}
	return string(data)
}

// Validate() validates the user replace params.
func (p *UserReplaceParams) Validate() error {
	if p.Username == "" {
		return errors.New("username cannot be empty")
	}
	if !UserRole(p.Role).IsValid() {
		return errors.New("invalid role")
	}
	if !UserStatus(p.Status).IsValid() {
		return errors.New("invalid status")
	}
	return nil
}

// UserUseCase is the use case for user.
type UserUseCase struct {
	userRepo  UserRepo
	tokenRepo TokenRepo
}

// NewUserUseCase creates a new UserUseCase.
func NewUserUseCase(userRepo UserRepo, tokenRepo TokenRepo) *UserUseCase {
	return &UserUseCase{userRepo: userRepo, tokenRepo: tokenRepo}
}

// ListUsers lists users.
func (uc *UserUseCase) ListUsers(ctx context.Context, params UserListParams) (*UserListResult, error) {
	return uc.userRepo.ListUsers(ctx, params)
}

// GetUser gets a user by ID.
func (uc *UserUseCase) GetUser(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, errors.New("user id is required")
	}

	user, err := uc.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user[id=%s]: %w", id, err)
	}
	return user, nil
}

// CreateUser creates a user.
func (uc *UserUseCase) CreateUser(ctx context.Context, params UserCreateParams) (*User, error) {
	// NOTE: setup default password
	params.Password = DefaultUserPassword
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid create user params: %w", err)
	}

	user, err := uc.userRepo.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

// UpdateUser updates a user.
func (uc *UserUseCase) UpdateUser(ctx context.Context, id string, params UserUpdateParams) (*User, error) {
	if id == "" {
		return nil, errors.New("user id is required")
	}
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid update user params: %w", err)
	}

	user, err := uc.userRepo.UpdateUser(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user[id=%s]: %w", id, err)
	}
	return user, nil
}

// ReplaceUser replaces a user.
func (uc *UserUseCase) ReplaceUser(ctx context.Context, id string, params UserReplaceParams) (*User, error) {
	if id == "" {
		return nil, errors.New("user id is required")
	}
	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid replace user params: %w", err)
	}

	user, err := uc.userRepo.ReplaceUser(ctx, id, params)
	if err != nil {
		return nil, fmt.Errorf("failed to replace user[id=%s]: %w", id, err)
	}
	return user, nil
}

// DeleteUser deletes a user.
func (uc *UserUseCase) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return errors.New("user id is required")
	}

	user, err := uc.userRepo.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user[id=%s]: %w", id, err)
	}

	username := user.Username
	if err := uc.userRepo.DeleteUser(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user[id=%s]: %w", id, err)
	}

	if err := uc.tokenRepo.DeleteTokensByUsername(ctx, username); err != nil {
		return fmt.Errorf("failed to delete tokens by username[%s]: %w", username, err)
	}

	return nil
}

// ResetUserPassword resets the user password.
func (uc *UserUseCase) ResetUserPassword(ctx context.Context, id string, newPassword string) (*User, error) {
	if id == "" {
		return nil, errors.New("user id is required")
	}

	user, err := uc.userRepo.ResetUserPassword(ctx, id, newPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to reset user password[id=%s]: %w", id, err)
	}
	return user, nil
}
