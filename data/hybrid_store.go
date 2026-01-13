// Hybrid store that can use either memory or database backend
//
// Architecture: Adapter Pattern
// HybridStore provides a unified interface that adapts between two different storage implementations:
// - MemoryStore: In-memory storage for development (simple map-based)
// - DatabaseService: PostgreSQL storage for production (repository-based)
//
// The adapter automatically detects which backend to use based on DATABASE_URL environment variable.
// This enables zero-configuration switching between development (no database) and production (PostgreSQL).
//
// The if/else routing in each method is intentional adapter logic, not code duplication.
package data

import (
	"fmt"
	"os"
)

// StoreBackend defines the type of backend storage
type StoreBackend string

const (
	BackendMemory   StoreBackend = "memory"
	BackendDatabase StoreBackend = "database"
)

// HybridStore provides a unified interface that can use either memory or database
type HybridStore struct {
	backend     StoreBackend
	memoryStore *MemoryStore
	dbService   *DatabaseService
}

// NewHybridStore creates a new hybrid store
func NewHybridStore(backend StoreBackend, databaseURL string) (*HybridStore, error) {
	store := &HybridStore{
		backend:     backend,
		memoryStore: NewMemoryStore(),
	}

	if backend == BackendDatabase {
		if databaseURL == "" {
			return nil, fmt.Errorf("database URL required for database backend")
		}

		err := InitDatabaseService(databaseURL)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database service: %w", err)
		}

		store.dbService = DBService
	}

	return store, nil
}

// AutoDetectBackend automatically detects which backend to use based on environment
func AutoDetectBackend() StoreBackend {
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		return BackendDatabase
	}
	return BackendMemory
}

// CreateInterview creates a new interview using the configured backend
func (h *HybridStore) CreateInterview(interview *Interview) error {
	if h.backend == BackendDatabase && h.dbService != nil {
		return h.dbService.InterviewRepo.Create(interview)
	}
	return h.memoryStore.CreateInterview(interview)
}

// GetInterview retrieves an interview by ID
func (h *HybridStore) GetInterview(id string) (*Interview, error) {
	if h.backend == BackendDatabase && h.dbService != nil {
		return h.dbService.InterviewRepo.GetByID(id)
	}
	return h.memoryStore.GetInterview(id)
}

// GetInterviewsWithOptions retrieves interviews with pagination, filtering, and sorting
func (h *HybridStore) GetInterviewsWithOptions(options ListInterviewsOptions) (*ListInterviewsResult, error) {
	if h.backend == BackendDatabase && h.dbService != nil {
		// Convert to database filters
		filters := InterviewFilters{
			CandidateName: options.CandidateName,
			Status:        options.Status,
		}
		if !options.DateFrom.IsZero() {
			filters.CreatedAfter = options.DateFrom
		}
		if !options.DateTo.IsZero() {
			filters.CreatedBefore = options.DateTo
		}

		interviews, total, err := h.dbService.InterviewRepo.List(options.Limit, options.Offset, filters)
		if err != nil {
			return nil, err
		}

		// Convert database result to ListInterviewsResult format
		totalPages := int(total) / options.Limit
		if int(total)%options.Limit > 0 {
			totalPages++
		}
		if totalPages == 0 {
			totalPages = 1
		}

		return &ListInterviewsResult{
			Interviews: interviews,
			Total:      int(total),
			Page:       (options.Offset / options.Limit) + 1,
			Limit:      options.Limit,
			TotalPages: totalPages,
		}, nil
	}

	// Fallback to memory store
	return h.memoryStore.GetInterviewsWithOptions(options)
}

// CreateEvaluation creates a new evaluation
func (h *HybridStore) CreateEvaluation(evaluation *Evaluation) error {
	if h.backend == BackendDatabase && h.dbService != nil {
		return h.dbService.EvaluationRepo.Create(evaluation)
	}
	return h.memoryStore.CreateEvaluation(evaluation)
}

// GetEvaluation retrieves an evaluation by ID
func (h *HybridStore) GetEvaluation(id string) (*Evaluation, error) {
	if h.backend == BackendDatabase && h.dbService != nil {
		return h.dbService.EvaluationRepo.GetByID(id)
	}
	return h.memoryStore.GetEvaluation(id)
}

// CreateChatSession creates a new chat session
func (h *HybridStore) CreateChatSession(session *ChatSession) error {
	if h.backend == BackendDatabase && h.dbService != nil {
		return h.dbService.ChatSessionRepo.Create(session)
	}
	return h.memoryStore.CreateChatSession(session)
}

// GetChatSession retrieves a chat session by ID
func (h *HybridStore) GetChatSession(id string) (*ChatSession, error) {
	if h.backend == BackendDatabase && h.dbService != nil {
		return h.dbService.ChatSessionRepo.GetByID(id)
	}
	return h.memoryStore.GetChatSession(id)
}

// UpdateChatSession updates a chat session
func (h *HybridStore) UpdateChatSession(session *ChatSession) error {
	if h.backend == BackendDatabase && h.dbService != nil {
		updates := map[string]interface{}{
			"status":   session.Status,
			"ended_at": session.EndedAt,
		}
		return h.dbService.ChatSessionRepo.Update(session.ID, updates)
	}
	return h.memoryStore.UpdateChatSession(session)
}

// AddChatMessage adds a message to a chat session
func (h *HybridStore) AddChatMessage(sessionID string, message *ChatMessage) error {
	if h.backend == BackendDatabase && h.dbService != nil {
		return h.dbService.ChatSessionRepo.AddMessage(sessionID, message)
	}
	// Memory store expects message with SessionID already set
	message.SessionID = sessionID
	return h.memoryStore.AddChatMessage(message)
}

// GetChatMessages retrieves all messages for a chat session
func (h *HybridStore) GetChatMessages(sessionID string) ([]*ChatMessage, error) {
	if h.backend == BackendDatabase && h.dbService != nil {
		return h.dbService.ChatSessionRepo.GetMessages(sessionID)
	}
	return h.memoryStore.GetChatMessages(sessionID)
}

// GetBackend returns the current backend type
func (h *HybridStore) GetBackend() StoreBackend {
	return h.backend
}

// Health checks the health of the current backend
func (h *HybridStore) Health() error {
	if h.backend == BackendDatabase && h.dbService != nil {
		return h.dbService.Health()
	}
	return nil // Memory store is always healthy
}

// Close closes the hybrid store and cleans up resources
func (h *HybridStore) Close() error {
	if h.backend == BackendDatabase && h.dbService != nil {
		return h.dbService.Close()
	}
	return nil // Memory store doesn't need cleanup
}

// Global hybrid store instance
var GlobalStore *HybridStore

// InitGlobalStore initializes the global store with auto-detected backend
func InitGlobalStore() error {
	backend := AutoDetectBackend()
	databaseURL := os.Getenv("DATABASE_URL")

	store, err := NewHybridStore(backend, databaseURL)
	if err != nil {
		return err
	}

	GlobalStore = store
	return nil
}
