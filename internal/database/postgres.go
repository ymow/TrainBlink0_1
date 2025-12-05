package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/ymow/messenger_protocol_research/internal/admin"
	"github.com/ymow/messenger_protocol_research/internal/config"
	"github.com/ymow/messenger_protocol_research/internal/user"
)

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	// Use DATABASE_URL if available, otherwise build DSN from config
	dsn := cfg.GetDatabaseURL()

	// Configure GORM logger
	gormLogger := logger.Default
	if cfg.SSLMode == "disable" {
		// Development mode - more verbose logging
		gormLogger = logger.Default.LogMode(logger.Info)
	} else {
		// Production mode - only errors
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Connect to database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true, // Better performance
		PrepareStmt:            true, // Prepared statement caching
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying *sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)                  // Maximum number of open connections
	sqlDB.SetMaxIdleConns(5)                   // Maximum number of idle connections
	sqlDB.SetConnMaxLifetime(5 * time.Minute)  // Maximum connection lifetime
	sqlDB.SetConnMaxIdleTime(10 * time.Minute) // Maximum idle time

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("✅ PostgreSQL connected: %s:%s/%s", cfg.Host, cfg.Port, cfg.DBName)

	return db, nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AutoMigrate runs GORM auto-migrations for all models
func AutoMigrate(db *gorm.DB) error {
	// Enable UUID extension
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		log.Printf("⚠️  Warning: Could not create uuid-ossp extension: %v", err)
	}

	// Run migrations for all models
	if err := db.AutoMigrate(
		&user.User{},
		&admin.Admin{},
		&admin.Role{},
	); err != nil {
		return fmt.Errorf("failed to run auto-migrations: %w", err)
	}

	log.Println("✅ Auto-migrations completed successfully")
	return nil
}
