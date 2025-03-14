package db

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
	"usermanage/gen/proto/conf"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Database wraps the gorm.DB instance and provides custom methods
type Database struct {
	*gorm.DB
}

// NewDatabase creates a new Database instance.
func NewDatabase(c *conf.Data) (*Database, error) {
	dsn := c.Database.Dsn
	driver := c.Database.Driver
	dbName := c.Database.Name

	// First connect to the server without specifying a database
	serverDSN := removeDatabaseFromDSN(dsn, driver)

	var serverDB *gorm.DB
	var err error
	switch driver {
	case conf.DatabaseDriver_DATABASE_DRIVER_MYSQL:
		serverDB, err = gorm.Open(mysql.Open(serverDSN), &gorm.Config{})
	case conf.DatabaseDriver_DATABASE_DRIVER_POSTGRES:
		serverDB, err = gorm.Open(postgres.Open(serverDSN), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unknown database driver: %s", driver)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database server: %w", err)
	}

	// Create the database if it doesn't exist
	db := &Database{serverDB}
	if err := db.CreateDatabaseIfNotExists(context.Background(), dbName); err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Now connect to the specific database
	var appDB *gorm.DB
	switch driver {
	case conf.DatabaseDriver_DATABASE_DRIVER_MYSQL:
		appDB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case conf.DatabaseDriver_DATABASE_DRIVER_POSTGRES:
		appDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open application database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := appDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Set reasonable connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return &Database{appDB}, nil
}

// CreateDatabaseIfNotExists creates the specified database if it doesn't exist.
func (db *Database) CreateDatabaseIfNotExists(ctx context.Context, dbName string) error {
	exists, err := db.databaseExists(ctx, dbName)
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}
	if !exists {
		if err := db.createDatabase(ctx, dbName); err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}
	return nil
}

// Check if the specified database exists.
func (db *Database) databaseExists(ctx context.Context, dbName string) (bool, error) {
	var query string
	var result []string
	dialectName := db.Dialector.Name()
	switch dialectName {
	case "mysql":
		query = "SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA WHERE SCHEMA_NAME = ?"
		if err := db.WithContext(ctx).Raw(query, dbName).Scan(&result).Error; err != nil {
			return false, fmt.Errorf("failed to check if database exists: %w", err)
		}
	case "postgres":
		query = "SELECT datname FROM pg_database WHERE datname = $1"
		if err := db.WithContext(ctx).Raw(query, dbName).Scan(&result).Error; err != nil {
			return false, fmt.Errorf("failed to check if database exists: %w", err)
		}
	default:
		return false, fmt.Errorf("unsupported database dialect for checking database existence: %s", dialectName)
	}
	return len(result) > 0, nil
}

// Create a MySQL database if it doesn't exist.
// This method assumes that the connection has sufficient privileges to create databases.
func (db *Database) createDatabase(ctx context.Context, dbName string) error {
	// We need to use raw SQL for database creation since GORM doesn't have direct methods for this
	// First, check if we're working with MySQL
	dialectName := db.Dialector.Name()
	var sql string
	switch dialectName {
	case "mysql":
		sql = fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", dbName)
	case "postgres":
		sql = fmt.Sprintf("CREATE DATABASE %s WITH ENCODING 'UTF8'", dbName)
	default:
		return fmt.Errorf("unsupported database dialect: %s", dialectName)
	}

	// Execute the SQL using GORM's exec
	result := db.WithContext(ctx).Exec(sql)
	if result.Error != nil {
		return fmt.Errorf("failed to create database: %w", result.Error)
	}

	return nil
}

func removeDatabaseFromDSN(dsn string, driver conf.DatabaseDriver) string {
	switch driver {
	case conf.DatabaseDriver_DATABASE_DRIVER_MYSQL:
		// MySQL DSN format: username:password@protocol(host:port)/dbname?param=value
		parts := strings.Split(dsn, "/")
		if len(parts) <= 1 {
			return dsn // No database in DSN
		}

		// Remove database name but keep parameters if any
		beforeDB := parts[0] // username:password@protocol(host:port)
		afterDB := ""
		if strings.Contains(parts[1], "?") {
			dbAndParams := strings.SplitN(parts[1], "?", 2)
			afterDB = "?" + dbAndParams[1]
		}
		return beforeDB + "/" + afterDB
	case conf.DatabaseDriver_DATABASE_DRIVER_POSTGRES:
		// PostgreSQL DSN format: postgres://username:password@host:port/dbname?param=value
		// or: host=localhost port=5432 user=postgres password=secret dbname=mydb
		if strings.HasPrefix(dsn, "postgres://") {
			// URL format
			u, err := url.Parse(dsn)
			if err != nil {
				return dsn // Return original if cannot parse
			}
			u.Path = "" // Remove the path which contains the DB name
			return u.String()
		} else {
			// Key=value format
			regex := regexp.MustCompile(`\bdbname=\w+\b`)
			return regex.ReplaceAllString(dsn, "")
		}
	}
	return dsn
}
