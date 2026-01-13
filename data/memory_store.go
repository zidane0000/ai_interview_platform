package data

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// MemoryStore provides in-memory storage for development and testing
// TODO: Replace with proper database implementation
type MemoryStore struct {
	interviews   map[string]*Interview
	evaluations  map[string]*Evaluation
	chatSessions map[string]*ChatSession
	chatMessages map[string][]*ChatMessage
	mu           sync.RWMutex
}

// NewMemoryStore creates a new in-memory store
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		interviews:   make(map[string]*Interview),
		evaluations:  make(map[string]*Evaluation),
		chatSessions: make(map[string]*ChatSession),
		chatMessages: make(map[string][]*ChatMessage),
	}
}

// Global memory store instance
var Store = NewMemoryStore()

// Interview operations
func (ms *MemoryStore) CreateInterview(interview *Interview) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.interviews[interview.ID] = interview
	return nil
}

func (ms *MemoryStore) GetInterview(id string) (*Interview, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	interview, exists := ms.interviews[id]
	if !exists {
		return nil, fmt.Errorf("interview not found")
	}
	return interview, nil
}

func (ms *MemoryStore) GetInterviews() ([]*Interview, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	interviews := make([]*Interview, 0, len(ms.interviews))
	for _, interview := range ms.interviews {
		interviews = append(interviews, interview)
	}
	return interviews, nil
}

// ListInterviewsOptions defines options for listing interviews with pagination, filtering and sorting
type ListInterviewsOptions struct {
	Limit         int       // Page size (default: 10)
	Offset        int       // Number of records to skip (default: 0)
	Page          int       // Page number (1-based, used to calculate offset if provided)
	CandidateName string    // Filter by candidate name (case-insensitive partial match)
	Status        string    // Filter by status
	DateFrom      time.Time // Filter interviews created after this date
	DateTo        time.Time // Filter interviews created before this date
	SortBy        string    // Sort field: "date", "name", "status" (default: "date")
	SortOrder     string    // Sort order: "asc", "desc" (default: "desc")
}

// ListInterviewsResult contains the result of listing interviews with pagination info
type ListInterviewsResult struct {
	Interviews []*Interview
	Total      int
	Page       int
	Limit      int
	TotalPages int
}

// GetInterviewsWithOptions returns interviews with pagination, filtering, and sorting
func (ms *MemoryStore) GetInterviewsWithOptions(opts ListInterviewsOptions) (*ListInterviewsResult, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	// Set defaults
	if opts.Limit <= 0 {
		opts.Limit = 10
	}
	if opts.Page > 0 {
		opts.Offset = (opts.Page - 1) * opts.Limit
	}
	if opts.SortBy == "" {
		opts.SortBy = "date"
	}
	if opts.SortOrder == "" {
		opts.SortOrder = "desc"
	}

	// Get all interviews and apply filters
	allInterviews := make([]*Interview, 0)
	for _, interview := range ms.interviews {
		// Apply filters
		if opts.CandidateName != "" {
			if !strings.Contains(strings.ToLower(interview.CandidateName), strings.ToLower(opts.CandidateName)) {
				continue
			}
		}

		if opts.Status != "" && interview.Status != opts.Status {
			continue
		}

		if !opts.DateFrom.IsZero() && interview.CreatedAt.Before(opts.DateFrom) {
			continue
		}

		if !opts.DateTo.IsZero() && interview.CreatedAt.After(opts.DateTo) {
			continue
		}

		allInterviews = append(allInterviews, interview)
	}

	// Sort interviews
	sort.Slice(allInterviews, func(i, j int) bool {
		switch opts.SortBy {
		case "name":
			if opts.SortOrder == "asc" {
				return strings.ToLower(allInterviews[i].CandidateName) < strings.ToLower(allInterviews[j].CandidateName)
			}
			return strings.ToLower(allInterviews[i].CandidateName) > strings.ToLower(allInterviews[j].CandidateName)
		case "status":
			if opts.SortOrder == "asc" {
				return allInterviews[i].Status < allInterviews[j].Status
			}
			return allInterviews[i].Status > allInterviews[j].Status
		default: // "date"
			if opts.SortOrder == "asc" {
				return allInterviews[i].CreatedAt.Before(allInterviews[j].CreatedAt)
			}
			return allInterviews[i].CreatedAt.After(allInterviews[j].CreatedAt)
		}
	})

	total := len(allInterviews)
	totalPages := (total + opts.Limit - 1) / opts.Limit

	// Apply pagination
	start := opts.Offset
	if start < 0 {
		start = 0
	}
	if start >= total {
		// Return empty result if offset is beyond total
		return &ListInterviewsResult{
			Interviews: []*Interview{},
			Total:      total,
			Page:       opts.Page,
			Limit:      opts.Limit,
			TotalPages: totalPages,
		}, nil
	}

	end := start + opts.Limit
	if end > total {
		end = total
	}

	pagedInterviews := allInterviews[start:end]

	return &ListInterviewsResult{
		Interviews: pagedInterviews,
		Total:      total,
		Page:       opts.Page,
		Limit:      opts.Limit,
		TotalPages: totalPages,
	}, nil
}

// Evaluation operations
func (ms *MemoryStore) CreateEvaluation(evaluation *Evaluation) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.evaluations[evaluation.ID] = evaluation
	return nil
}

func (ms *MemoryStore) GetEvaluation(id string) (*Evaluation, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	evaluation, exists := ms.evaluations[id]
	if !exists {
		return nil, fmt.Errorf("evaluation not found")
	}
	return evaluation, nil
}

// Chat session operations
func (ms *MemoryStore) CreateChatSession(session *ChatSession) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.chatSessions[session.ID] = session
	ms.chatMessages[session.ID] = []*ChatMessage{}
	return nil
}

func (ms *MemoryStore) GetChatSession(id string) (*ChatSession, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	session, exists := ms.chatSessions[id]
	if !exists {
		return nil, fmt.Errorf("chat session not found")
	}
	return session, nil
}

func (ms *MemoryStore) UpdateChatSession(session *ChatSession) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if _, exists := ms.chatSessions[session.ID]; !exists {
		return fmt.Errorf("chat session not found")
	}
	session.UpdatedAt = time.Now()
	ms.chatSessions[session.ID] = session
	return nil
}

// Chat message operations
func (ms *MemoryStore) AddChatMessage(message *ChatMessage) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	if _, exists := ms.chatMessages[message.SessionID]; !exists {
		return fmt.Errorf("chat session not found")
	}
	ms.chatMessages[message.SessionID] = append(ms.chatMessages[message.SessionID], message)
	return nil
}

func (ms *MemoryStore) GetChatMessages(sessionID string) ([]*ChatMessage, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	messages, exists := ms.chatMessages[sessionID]
	if !exists {
		return nil, fmt.Errorf("chat session not found")
	}
	return messages, nil
}
