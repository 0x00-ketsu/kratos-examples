package biz

import (
	"context"
	"fmt"
	"usermanage/internal/pkg/db"

	"github.com/redis/go-redis/v9"
)

// HealthUseCase is a use case for health check.
type HealthUseCase struct {
	db  *db.Database
	rdb redis.UniversalClient
}

// NewHealthUseCase returns a new instance of the HealthUseCase.
func NewHealthUseCase(db *db.Database, rdb redis.UniversalClient) *HealthUseCase {
	return &HealthUseCase{db: db, rdb: rdb}
}

// PingDB checks the database connection.
func (uc *HealthUseCase) PingDB(ctx context.Context) error {
	if err := uc.db.Raw("SELECT 1").Error; err != nil {
		return fmt.Errorf("database ping error: %w", err)
	}
	return nil
}

// PingRedis checks the redis connection.
func (uc *HealthUseCase) PingRedis(ctx context.Context) error {
	if err := uc.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping error: %w", err)
	}
	return nil
}
