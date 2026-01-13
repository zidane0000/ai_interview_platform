// Database service layer that coordinates between repositories
package data

import (
	"gorm.io/gorm"
)

// DatabaseService provides a unified interface for all database operations
type DatabaseService struct {
	db              *gorm.DB
	InterviewRepo   InterviewRepository
	EvaluationRepo  EvaluationRepository
	ChatSessionRepo ChatSessionRepository
}

// NewDatabaseService creates a new database service with all repositories
func NewDatabaseService(db *gorm.DB) *DatabaseService {
	return &DatabaseService{
		db:              db,
		InterviewRepo:   NewInterviewRepository(db),
		EvaluationRepo:  NewEvaluationRepository(db),
		ChatSessionRepo: NewChatSessionRepository(db),
	}
}

// DB returns the underlying GORM database instance for advanced operations
func (s *DatabaseService) DB() *gorm.DB {
	return s.db
}

// Transaction executes a function within a database transaction
func (s *DatabaseService) Transaction(fn func(*gorm.DB) error) error {
	return s.db.Transaction(fn)
}

// Health checks database connectivity
func (s *DatabaseService) Health() error {
	return s.db.Exec("SELECT 1").Error
}

// Close closes the database connection
func (s *DatabaseService) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Global database service instance (will be initialized when database is connected)
var DBService *DatabaseService

// InitDatabaseService initializes the global database service
func InitDatabaseService(databaseURL string) error {
	db, err := InitDB(databaseURL)
	if err != nil {
		return err
	}

	DBService = NewDatabaseService(db)
	return nil
}
