package data

import (
	"github.com/zidane0000/ai-interview-platform/utils"
	"gorm.io/gorm"
)

// AddPerformanceIndexes creates additional database indexes for better performance
func AddPerformanceIndexes(db *gorm.DB) error { // Index for interview queries
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_interviews_status ON interviews(status);").Error; err != nil {
		utils.Warningf("Could not create status index: %v\n", err)
	}

	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_interviews_created_at ON interviews(created_at);").Error; err != nil {
		utils.Warningf("Could not create created_at index: %v\n", err)
	}

	// Index for evaluation queries
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_evaluations_interview_id_created_at ON evaluations(interview_id, created_at);").Error; err != nil {
		utils.Warningf("Warning: Could not create evaluation composite index: %v\n", err)
	}

	// Index for chat session queries
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chat_sessions_status ON chat_sessions(status);").Error; err != nil {
		utils.Warningf("Warning: Could not create chat session status index: %v\n", err)
	}

	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chat_sessions_interview_id_status ON chat_sessions(interview_id, status);").Error; err != nil {
		utils.Warningf("Warning: Could not create chat session composite index: %v\n", err)
	}

	// Index for chat message queries
	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chat_messages_session_id_timestamp ON chat_messages(session_id, timestamp);").Error; err != nil {
		utils.Warningf("Warning: Could not create chat message composite index: %v\n", err)
	}

	if err := db.Exec("CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_chat_messages_type ON chat_messages(type);").Error; err != nil {
		utils.Warningf("Warning: Could not create chat message type index: %v\n", err)
	}

	return nil
}
