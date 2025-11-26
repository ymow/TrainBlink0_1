package database

import (
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(databaseURL string) (*gorm.DB, error) {
	// Extract SQLite file path from URL
	// sqlite://trainblink.db -> trainblink.db
	dbPath := databaseURL[9:] // Remove "sqlite://" prefix

	// Configure GORM logger for development
	gormLogger := logger.Default.LogMode(logger.Info)

	// Connect to SQLite database
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger:                 gormLogger,
		SkipDefaultTransaction: true, // Better performance
		PrepareStmt:            true, // Prepared statement caching
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}

	// Auto-migrate all models
	if err := AutoMigrateSQLite(db); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Printf("✅ SQLite connected: %s", dbPath)

	return db, nil
}

// AutoMigrateSQLite runs GORM auto-migrations for essential chat models
func AutoMigrateSQLite(db *gorm.DB) error {
	// For SQLite, we need to be more selective about which models we migrate
	// due to PostgreSQL-specific syntax in the models
	
	// First, create simple chat table manually for SQLite
	err := db.Exec(`
		CREATE TABLE IF NOT EXISTS chat_messages (
			id TEXT PRIMARY KEY,
			text TEXT NOT NULL,
			sender_id TEXT NOT NULL,
			receiver_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			is_encrypted BOOLEAN DEFAULT FALSE,
			is_read BOOLEAN DEFAULT FALSE,
			delivery_status TEXT DEFAULT 'PENDING',
			delivered_at DATETIME,
			read_at DATETIME,
			is_ephemeral BOOLEAN DEFAULT FALSE,
			expires_at DATETIME,
			is_expired BOOLEAN DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create chat_messages table: %w", err)
	}

	// Create trips table for basic functionality
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS trips (
			id TEXT PRIMARY KEY,
			user_id TEXT NOT NULL,
			route TEXT NOT NULL,
			train_number TEXT,
			departure_time DATETIME NOT NULL,
			estimated_arrival DATETIME NOT NULL,
			actual_end_time DATETIME,
			discovery_enabled BOOLEAN DEFAULT TRUE,
			ble_anonymous_id TEXT NOT NULL UNIQUE,
			status TEXT DEFAULT 'active',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			CHECK (status IN ('active', 'ended', 'cancelled'))
		)
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create trips table: %w", err)
	}

	// Create Matrix ephemeral rooms table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS matrix_ephemeral_rooms (
			id TEXT PRIMARY KEY,
			room_id TEXT NOT NULL UNIQUE,
			trip1_id TEXT NOT NULL,
			trip2_id TEXT NOT NULL,
			anonymous_id_1 TEXT NOT NULL,
			anonymous_id_2 TEXT NOT NULL,
			mls_group_id TEXT,
			expires_at DATETIME NOT NULL,
			auto_delete_queued BOOLEAN DEFAULT FALSE,
			deleted_at DATETIME,
			message_count INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create matrix_ephemeral_rooms table: %w", err)
	}

	// Create basic discoveries table
	err = db.Exec(`
		CREATE TABLE IF NOT EXISTS discoveries (
			id TEXT PRIMARY KEY,
			discovered_by TEXT NOT NULL,
			discovered_user TEXT NOT NULL,
			route TEXT NOT NULL,
			station TEXT NOT NULL,
			discovery_time DATETIME NOT NULL,
			distance_meters REAL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`).Error
	if err != nil {
		return fmt.Errorf("failed to create discoveries table: %w", err)
	}

	log.Println("✅ SQLite auto-migrations completed successfully")
	return nil
}