package db

import (
	"usermanage/gen/proto/conf"
	"usermanage/internal/pkg/rdb"

	"github.com/redis/go-redis/v9"
)

// NewRedis creates a new Redis client.
func NewRedis(c *conf.Data) (redis.UniversalClient, error) {
	opts := &redis.UniversalOptions{
		Addrs:        c.Redis.Addrs,
		Password:     c.Redis.Password,
		DB:           int(c.Redis.Db),
		DialTimeout:  c.Redis.DialTimeout.AsDuration(),
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	}
	return rdb.NewClient(opts)
}
