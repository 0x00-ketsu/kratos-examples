package rdb

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewClient creates a new Redis client and returns it.
func NewClient(opts *redis.UniversalOptions) (redis.UniversalClient, error) {
    if opts == nil {
        return nil, fmt.Errorf("redis options cannot be nil")
    }

	client := redis.NewUniversalClient(opts)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	return client, nil
}
