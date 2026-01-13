// Chat session data access (CRUD operations)
package data

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// ChatSessionFilters defines filter options for chat session queries
type ChatSessionFilters struct {
	InterviewID   string
	Status        string
	CreatedAfter  time.Time
	CreatedBefore time.Time
}

// ChatSessionRepository interface defines the contract for chat session data access
type ChatSessionRepository interface {
	Create(session *ChatSession) error
	GetByID(id string) (*ChatSession, error)
	GetByInterviewID(interviewID string) (*ChatSession, error)
	List(limit, offset int, filters ChatSessionFilters) ([]*ChatSession, int64, error)
	Update(id string, updates map[string]interface{}) error
	Delete(id string) error
	AddMessage(sessionID string, message *ChatMessage) error
	GetMessages(sessionID string) ([]*ChatMessage, error)
}

// chatSessionRepository implements ChatSessionRepository interface
type chatSessionRepository struct {
	db *gorm.DB
}

// NewChatSessionRepository creates a new chat session repository
func NewChatSessionRepository(db *gorm.DB) ChatSessionRepository {
	return &chatSessionRepository{db: db}
}

// Create creates a new chat session
func (r *chatSessionRepository) Create(session *ChatSession) error {
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	return r.db.Create(session).Error
}

// GetByID retrieves a chat session by ID
func (r *chatSessionRepository) GetByID(id string) (*ChatSession, error) {
	var session ChatSession
	err := r.db.Where("id = ?", id).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("chat session not found")
	}
	return &session, err
}

// GetByInterviewID retrieves a chat session by interview ID
func (r *chatSessionRepository) GetByInterviewID(interviewID string) (*ChatSession, error) {
	var session ChatSession
	err := r.db.Where("interview_id = ?", interviewID).First(&session).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("chat session not found")
	}
	return &session, err
}

// List retrieves chat sessions with pagination and filtering
func (r *chatSessionRepository) List(limit, offset int, filters ChatSessionFilters) ([]*ChatSession, int64, error) {
	var sessions []*ChatSession
	var total int64

	query := r.db.Model(&ChatSession{})

	// Apply filters
	if filters.InterviewID != "" {
		query = query.Where("interview_id = ?", filters.InterviewID)
	}
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if !filters.CreatedAfter.IsZero() {
		query = query.Where("created_at >= ?", filters.CreatedAfter)
	}
	if !filters.CreatedBefore.IsZero() {
		query = query.Where("created_at <= ?", filters.CreatedBefore)
	}

	// Get total count
	query.Count(&total)

	// Apply pagination and ordering
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&sessions).Error
	return sessions, total, err
}

// Update updates a chat session
func (r *chatSessionRepository) Update(id string, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return r.db.Model(&ChatSession{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes a chat session
func (r *chatSessionRepository) Delete(id string) error {
	// Also delete associated messages
	r.db.Where("session_id = ?", id).Delete(&ChatMessage{})
	return r.db.Where("id = ?", id).Delete(&ChatSession{}).Error
}

// AddMessage adds a message to a chat session
func (r *chatSessionRepository) AddMessage(sessionID string, message *ChatMessage) error {
	// Verify session exists
	var session ChatSession
	if err := r.db.Where("id = ?", sessionID).First(&session).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("chat session not found")
		}
		return err
	}

	message.SessionID = sessionID
	message.CreatedAt = time.Now()
	return r.db.Create(message).Error
}

// GetMessages retrieves all messages for a chat session
func (r *chatSessionRepository) GetMessages(sessionID string) ([]*ChatMessage, error) {
	var messages []*ChatMessage
	err := r.db.Where("session_id = ?", sessionID).Order("timestamp ASC").Find(&messages).Error
	return messages, err
}
