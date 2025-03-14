package data

import (
	"context"
	"fmt"
	"time"
	"usermanage/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type redisTokenRepo struct {
	client redis.UniversalClient
	logger *log.Helper
}

// NewRedisTokenRepo returns a new instance of RedisTokenRepo.
func NewRedisTokenRepo(client redis.UniversalClient, logger log.Logger) biz.TokenRepo {
	return &redisTokenRepo{
		client: client,
		logger: log.NewHelper(logger),
	}
}

// DeleteToken implements biz.TokenRepo.
func (r *redisTokenRepo) DeleteToken(ctx context.Context, token string) error {
	username, err := r.GetUsernameByToken(ctx, token)
	if err != nil {
		return err
	}

	if err := r.client.Del(ctx, r.tokenKey(token)).Err(); err != nil {
		return err
	}

	return r.client.SRem(ctx, r.userTokensKey(username), token).Err()
}

// DeleteTokensByUsername implements biz.TokenRepo.
func (r *redisTokenRepo) DeleteTokensByUsername(ctx context.Context, username string) error {
	tokens, err := r.client.SMembers(ctx, r.userTokensKey(username)).Result()
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return nil
	}

	keys := make([]string, len(tokens)+1)
	for i, token := range tokens {
		keys[i] = r.tokenKey(token)
	}
	keys[len(tokens)] = r.userTokensKey(username)
	return r.client.Del(ctx, keys...).Err()
}

// GetUsernameByToken implements biz.TokenRepo.
func (r *redisTokenRepo) GetUsernameByToken(ctx context.Context, token string) (string, error) {
	username, err := r.client.Get(ctx, r.tokenKey(token)).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token not found")
	}
	return username, err
}

// ExtendTokenExpiry implements biz.TokenRepo.
func (r *redisTokenRepo) ExtendTokenExpiry(ctx context.Context, token string, duration time.Duration) error {
	return r.client.Expire(ctx, r.tokenKey(token), duration).Err()
}

// StoreToken implements biz.TokenRepo.
func (r *redisTokenRepo) StoreToken(ctx context.Context, token string, username string, expiration time.Duration) error {
	err := r.client.Set(ctx, r.tokenKey(token), username, expiration).Err()
	if err != nil {
		return err
	}
	return r.client.SAdd(ctx, r.userTokensKey(username), token).Err()
}

// TokenExists implements biz.TokenRepo.
func (r *redisTokenRepo) TokenExists(ctx context.Context, token string) (bool, error) {
	exists, err := r.client.Exists(ctx, r.tokenKey(token)).Result()
	return exists > 0, err
}

// UserHasActiveSession implements biz.TokenRepo.
func (r *redisTokenRepo) UserHasActiveSession(ctx context.Context, username string) (bool, error) {
	tokens, err := r.client.SMembers(ctx, r.userTokensKey(username)).Result()
	if err != nil {
		return false, err
	}

	for _, token := range tokens {
		exists, err := r.client.Exists(ctx, r.tokenKey(token)).Result()
		if err != nil {
			return false, err
		}
		if exists > 0 {
			return true, nil
		}
	}
	return false, nil
}

// Return the key for storing a token.
func (r *redisTokenRepo) tokenKey(token string) string {
	return "token:" + token
}

// Return the key for storing a user's tokens.
func (r *redisTokenRepo) userTokensKey(username string) string {
	return "user_tokens:" + username
}
