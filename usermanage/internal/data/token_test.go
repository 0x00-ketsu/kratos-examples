package data

import (
	"context"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedisTokenRepo_StoreToken(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewRedisTokenRepo(client, log.DefaultLogger)

	ctx := context.Background()
	token := "test-token"
	username := "testuser"
	expiration := 15 * time.Minute

	// Mock expectations
	mock.ExpectSet("token:"+token, username, expiration).SetVal("OK")
	mock.ExpectSAdd("user_tokens:"+username, token).SetVal(1)

	// Execute the method
	err := repo.StoreToken(ctx, token, username, expiration)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisTokenRepo_GetUsernameByToken(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewRedisTokenRepo(client, log.DefaultLogger)

	ctx := context.Background()
	token := "test-token"
	username := "testuser"

	t.Run("Token exists", func(t *testing.T) {
		mock.ExpectGet("token:" + token).SetVal(username)

		result, err := repo.GetUsernameByToken(ctx, token)
		assert.NoError(t, err)
		assert.Equal(t, username, result)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Token not found", func(t *testing.T) {
		mock.ExpectGet("token:" + token).SetErr(redis.Nil)

		result, err := repo.GetUsernameByToken(ctx, token)
		assert.Error(t, err)
		assert.Equal(t, "", result)
		assert.Contains(t, err.Error(), "token not found")
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRedisTokenRepo_DeleteToken(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewRedisTokenRepo(client, log.DefaultLogger)

	ctx := context.Background()
	token := "test-token"
	username := "testuser"

	mock.ExpectGet("token:" + token).SetVal(username)
	mock.ExpectDel("token:" + token).SetVal(1)
	mock.ExpectSRem("user_tokens:"+username, token).SetVal(1)

	err := repo.DeleteToken(ctx, token)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisTokenRepo_DeleteTokensByUsername(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewRedisTokenRepo(client, log.DefaultLogger)

	ctx := context.Background()
	username := "testuser"
	tokens := []string{"token1", "token2"}

	t.Run("With tokens", func(t *testing.T) {
		mock.ExpectSMembers("user_tokens:" + username).SetVal(tokens)
		mock.ExpectDel("token:token1", "token:token2", "user_tokens:"+username).SetVal(3)

		err := repo.DeleteTokensByUsername(ctx, username)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No tokens", func(t *testing.T) {
		mock.ExpectSMembers("user_tokens:" + username).SetVal([]string{})

		err := repo.DeleteTokensByUsername(ctx, username)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRedisTokenRepo_ExtendTokenExpiry(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewRedisTokenRepo(client, log.DefaultLogger)

	ctx := context.Background()
	token := "test-token"
	duration := 30 * time.Minute

	mock.ExpectExpire("token:"+token, duration).SetVal(true)

	err := repo.ExtendTokenExpiry(ctx, token, duration)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedisTokenRepo_TokenExists(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewRedisTokenRepo(client, log.DefaultLogger)

	ctx := context.Background()
	token := "test-token"

	t.Run("Token exists", func(t *testing.T) {
		mock.ExpectExists("token:" + token).SetVal(1)

		exists, err := repo.TokenExists(ctx, token)
		assert.NoError(t, err)
		assert.True(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("Token does not exist", func(t *testing.T) {
		mock.ExpectExists("token:" + token).SetVal(0)

		exists, err := repo.TokenExists(ctx, token)
		assert.NoError(t, err)
		assert.False(t, exists)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRedisTokenRepo_UserHasActiveSession(t *testing.T) {
	client, mock := redismock.NewClientMock()
	repo := NewRedisTokenRepo(client, log.DefaultLogger)

	ctx := context.Background()
	username := "testuser"
	tokens := []string{"token1", "token2"}

	t.Run("User has active session", func(t *testing.T) {
		mock.ExpectSMembers("user_tokens:" + username).SetVal(tokens)
		mock.ExpectExists("token:" + tokens[0]).SetVal(1)

		hasSession, err := repo.UserHasActiveSession(ctx, username)
		assert.NoError(t, err)
		assert.True(t, hasSession)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No active session", func(t *testing.T) {
		mock.ExpectSMembers("user_tokens:" + username).SetVal(tokens)
		mock.ExpectExists("token:" + tokens[0]).SetVal(0)
		mock.ExpectExists("token:" + tokens[1]).SetVal(0)

		hasSession, err := repo.UserHasActiveSession(ctx, username)
		assert.NoError(t, err)
		assert.False(t, hasSession)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("No tokens", func(t *testing.T) {
		mock.ExpectSMembers("user_tokens:" + username).SetVal([]string{})

		hasSession, err := repo.UserHasActiveSession(ctx, username)
		assert.NoError(t, err)
		assert.False(t, hasSession)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
