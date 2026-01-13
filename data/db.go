// Database connection and setup using GORM
package data

import (
	"fmt"
	"time"

	"github.com/zidane0000/ai-interview-platform/utils"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes and returns a PostgreSQL connection using GORM
func InitDB(databaseURL string) (*gorm.DB, error) {
	// Configure GORM for better performance
	config := &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent), // Reduce logging overhead in production
		NowFunc:                                  func() time.Time { return time.Now().UTC() },
		PrepareStmt:                              true, // Enable prepared statements for better performance
		DisableForeignKeyConstraintWhenMigrating: false,
	}

	db, err := gorm.Open(postgres.Open(databaseURL), config)
	if err != nil {
		return nil, err
	}

	// Configure connection pool settings for optimal performance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Connection pool optimization for concurrent operations
	sqlDB.SetMaxIdleConns(25)                  // Keep 25 idle connections ready
	sqlDB.SetMaxOpenConns(100)                 // Allow up to 100 concurrent connections
	sqlDB.SetConnMaxLifetime(time.Hour)        // Recycle connections every hour
	sqlDB.SetConnMaxIdleTime(15 * time.Minute) // Close idle connections after 15 minutes	// Run database migrations automatically
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// Add performance indexes for better concurrent query performance
	if err := AddPerformanceIndexes(db); err != nil {
		// Don't fail if indexes can't be created, just log warning
		utils.Warningf("Some performance indexes could not be created: %v\n", err)
	}

	// Add database health check
	if err := db.Exec("SELECT 1").Error; err != nil {
		return nil, fmt.Errorf("database health check failed: %w", err)
	}

	return db, nil
}

// Implement database migration function
func runMigrations(db *gorm.DB) error {
	return db.AutoMigrate(
		&Interview{},
		&Evaluation{},
		&ChatSession{},
		&ChatMessage{},
		// &File{}, // TODO: Uncomment when File model is implemented
	)
}

// Implement database seeding for development
func SeedDatabase(db *gorm.DB) error {
	// Add sample data for development/testing
	// This can be used to populate the database with test data
	return nil
}

// Implement database backup utilities
func BackupDatabase(db *gorm.DB, outputPath string) error {
	// TODO: Implement pg_dump wrapper or similar
	return nil
}

// CloseDB closes the provided database connection (if needed)
func CloseDB(db *gorm.DB) {
	if db == nil {
		return
	}

	// TODO: Add graceful connection cleanup
	// TODO: Wait for active transactions to complete
	// TODO: Log connection close status

	dbConn, err := db.DB()
	if err == nil {
		if closeErr := dbConn.Close(); closeErr != nil {
			utils.WarningIf(closeErr)
		}
	}
}

// TODO: Add database monitoring and metrics collection
// TODO: Implement connection retry logic with exponential backoff
// TODO: Add support for read replicas and write/read separation
// TODO: Add database transaction helpers and utilities
