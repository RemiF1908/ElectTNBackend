package database

import (
	"fmt"

	"elect-tn-backend/internal/config"
	"elect-tn-backend/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewConnection creates a new database connection
func NewConnection(cfg *config.Config) (*gorm.DB, error) {
	// Configure GORM logger based on application mode
	var logLevel logger.LogLevel
	if cfg.IsDevelopment() {
		logLevel = logger.Warn
	} else {
		logLevel = logger.Error
	}

	// Open database connection using PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Election{},
		&models.Candidacy{},
		&models.UserVoteTracking{},
		&models.VoteCount{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return db, nil
}

// Ping tests the database connection
func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	return sqlDB.Close()
}
