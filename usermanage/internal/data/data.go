package data

import (
	"context"
	"fmt"
	"usermanage/internal/data/model"
	"usermanage/internal/pkg/constants"
	"usermanage/internal/pkg/db"
	"usermanage/internal/pkg/password"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/redis/go-redis/v9"
)

type Data struct {
	db     *db.Database
	rdb    redis.UniversalClient
	logger *log.Helper
}

// NewData creates a new Data instance
func NewData(db *db.Database, rdb redis.UniversalClient, logger log.Logger) (*Data, error) {
	return &Data{
		db:     db,
		rdb:    rdb,
		logger: log.NewHelper(logger),
	}, nil
}

// Cleanup closes the resources.
func (d *Data) Cleanup() {
	// TODO: add resource cleanup
	d.logger.Info("closing the data resources")
}

// Migrate migrate database schema.
func (d *Data) Migrate() error {
	d.logger.Info("migrate database schema")
	models := []any{model.User{}}
	return d.db.AutoMigrate(models...)
}

// InitializeAdminAccount creates the root admin account if it does not exist.
func (d *Data) InitializeAdminAccount(ctx context.Context) error {
	d.logger.Info("checking if root account needs to be created")

	// Check if root username exists
	username := "admin"
	var count int64
	if err := d.db.WithContext(ctx).
		Model(&model.User{}).
		Where("username = ?", username).
		Count(&count).Error; err != nil {
		return fmt.Errorf("failed to check existing users: %w", err)
	}

	// If users already exist, skip initialization
	if count > 0 {
		d.logger.Info("user admin already exist, skipping root account creation")
		return nil
	}

	passwd, err := password.GeneratePassword(16)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	// Create root account
	adminUser := &model.User{
		Username: username,
		Password: passwd,
		Role:     constants.UserRoleAdmin,
	}

	d.logger.Info("creating admin account")
	if err := d.db.WithContext(ctx).
		Create(adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin account: %w", err)
	}

	d.logger.Infow("msg", "admin account created successfully", "credential", passwd)
	return nil
}
