package biz

import (
	"context"
	"errors"
	"fmt"
	"time"
	"usermanage/internal/pkg/jwt"
)

// TokenRepo defines operations for managing tokens.
type TokenRepo interface {
	// StoreToken saves a token and associates it with a username.
	StoreToken(ctx context.Context, token string, username string, expiration time.Duration) error

	// GetUsernameByToken retrieves the username associated with a token.
	GetUsernameByToken(ctx context.Context, token string) (string, error)

	// DeleteTokensByUsername deletes all tokens associated with a username.
	DeleteTokensByUsername(ctx context.Context, username string) error

	// DeleteToken deletes a token.
	DeleteToken(ctx context.Context, token string) error

	// TokenExists checks whether a token exists.
	TokenExists(ctx context.Context, token string) (bool, error)

	// ExtendTokenExpiry extends the expiry of a token.
	ExtendTokenExpiry(ctx context.Context, token string, duration time.Duration) error

	// UserHasActiveSession checks whether a user has an active session.
	UserHasActiveSession(ctx context.Context, username string) (bool, error)
}

// AuthUseCase is the use case for auth.
type AuthUseCase struct {
	userRepo  UserRepo
	tokenRepo TokenRepo
}

// NewAuthUseCase creates a new AuthUseCase.
func NewAuthUseCase(userRepo UserRepo, tokenRepo TokenRepo) *AuthUseCase {
	return &AuthUseCase{userRepo: userRepo, tokenRepo: tokenRepo}
}

// Login logs in a user.
func (uc *AuthUseCase) Login(ctx context.Context, username, password string) (user *User, token string, expiresAt time.Time, err error) {
	user, err = uc.validateCredentials(ctx, username, password)
	if err != nil {
		err = fmt.Errorf("failed to validate credentials: %w", err)
		return
	}

	token, expiresAt, err = uc.generateToken(ctx, user.Username)
	if err != nil {
		err = fmt.Errorf("failed to generate token: %w", err)
		return
	}
	return
}

// Logout logs out a user.
func (uc *AuthUseCase) Logout(ctx context.Context, token string) (username string, err error) {
	user, err := uc.VerifyToken(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to verify token: %w", err)
	}

	if err := uc.DeleteToken(ctx, token); err != nil {
		return "", fmt.Errorf("failed to delete token: %w", err)
	}

	return user.Username, nil
}

func (uc *AuthUseCase) ChangePassword(ctx context.Context, username string, oldPassword, newPassword string) error {
	if username == "" {
		return errors.New("username is required")
	}

	if oldPassword == "" {
		return errors.New("old password is required")
	}
	if oldPassword == newPassword {
		return errors.New("new password must be different from old password")
	}

	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("failed to get user by username[%s]: %w", username, err)
	}

	userID := user.ID
	if ok, err := uc.userRepo.VerifyPassword(ctx, userID, oldPassword); err != nil || !ok {
		return fmt.Errorf("failed to verify old password: %w", err)
	}

	// TODO: validate the strength of the new password

	_, err = uc.userRepo.ResetUserPassword(ctx, userID, newPassword)
	if err != nil {
		return fmt.Errorf("failed to reset user password: %w", err)
	}

	if err := uc.tokenRepo.DeleteTokensByUsername(ctx, username); err != nil {
		return fmt.Errorf("failed to delete tokens by username[%s]: %w", username, err)
	}

	return nil
}

// DeleteToken deletes a token.
func (uc *AuthUseCase) DeleteToken(ctx context.Context, token string) error {
	return uc.tokenRepo.DeleteToken(ctx, token)
}

// VerifyToken verifies a token and returns the user associated with it.
func (uc *AuthUseCase) VerifyToken(ctx context.Context, token string) (*User, error) {
	if token == "" {
		return nil, errors.New("token is required")
	}

	username, err := uc.tokenRepo.GetUsernameByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get username by token: %w", err)
	}

	if username == "" {
		return nil, errors.New("empty username from token")
	}
	return uc.GetUserByUsername(ctx, username)
}

// TokenExists checks whether a token exists.
func (uc *AuthUseCase) TokenExists(ctx context.Context, token string) (bool, error) {
	return uc.tokenRepo.TokenExists(ctx, token)
}

// ExtendTokenExpiry extends the expiry of a token.
func (uc *AuthUseCase) ExtendTokenExpiry(ctx context.Context, token string) error {
	return uc.tokenRepo.ExtendTokenExpiry(ctx, token, jwt.TokenExpireDuration())
}

// GetUserByUsername retrieves a user by username.
func (uc *AuthUseCase) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	user, err := uc.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by username[%s]: %w", username, err)
	}
	return user, nil
}

// Validate the user's credentials.
//
// # Note
//
// The `password` parameter is plaintext
func (uc *AuthUseCase) validateCredentials(ctx context.Context, username, rawPassword string) (*User, error) {
	user, err := uc.userRepo.FindByCredentials(ctx, username, rawPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by credentials: %w", err)
	}
	return user, nil
}

// Generate a token for a user and stores it in Redis.
func (uc *AuthUseCase) generateToken(ctx context.Context, username string) (token string, expiresAt time.Time, err error) {
	token, expiresAt, err = jwt.GenerateToken(username)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to generate token: %w", err)
	}

	err = uc.tokenRepo.StoreToken(ctx, token, username, time.Until(expiresAt))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to store token: %w", err)
	}
	return
}
