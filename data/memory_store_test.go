package data_test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/zidane0000/ai-interview-platform/data"
)

func TestMemoryStore_InterviewOperations(t *testing.T) {
	store := data.NewMemoryStore()

	// Test CreateInterview
	interview := &data.Interview{
		ID:             "test-interview-1",
		CandidateName:  "John Doe",
		Questions:      []string{"Q1", "Q2"},
		InterviewType:  "technical",
		JobDescription: "Software Engineer",
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err := store.CreateInterview(interview)
	if err != nil {
		t.Fatalf("CreateInterview failed: %v", err)
	}

	// Test GetInterview
	retrieved, err := store.GetInterview("test-interview-1")
	if err != nil {
		t.Fatalf("GetInterview failed: %v", err)
	}

	if retrieved.ID != interview.ID {
		t.Errorf("expected ID %s, got %s", interview.ID, retrieved.ID)
	}
	if retrieved.CandidateName != interview.CandidateName {
		t.Errorf("expected CandidateName %s, got %s", interview.CandidateName, retrieved.CandidateName)
	}
	if retrieved.InterviewType != interview.InterviewType {
		t.Errorf("expected InterviewType %s, got %s", interview.InterviewType, retrieved.InterviewType)
	}

	// Test GetInterview with non-existent ID
	_, err = store.GetInterview("non-existent")
	if err == nil {
		t.Error("expected error for non-existent interview")
	}

	// Test GetInterviews
	interviews, err := store.GetInterviews()
	if err != nil {
		t.Fatalf("GetInterviews failed: %v", err)
	}

	if len(interviews) != 1 {
		t.Errorf("expected 1 interview, got %d", len(interviews))
	}
}

func TestMemoryStore_GetInterviewsWithOptions(t *testing.T) {
	store := data.NewMemoryStore()

	// Create test interviews
	interviews := []*data.Interview{
		{
			ID:            "interview-1",
			CandidateName: "Alice Johnson",
			Questions:     []string{"Q1"},
			InterviewType: "technical",
			Status:        "pending",
			CreatedAt:     time.Now().Add(-2 * time.Hour),
			UpdatedAt:     time.Now().Add(-2 * time.Hour),
		},
		{
			ID:            "interview-2",
			CandidateName: "Bob Smith",
			Questions:     []string{"Q2"},
			InterviewType: "behavioral",
			Status:        "completed",
			CreatedAt:     time.Now().Add(-1 * time.Hour),
			UpdatedAt:     time.Now().Add(-1 * time.Hour),
		},
		{
			ID:            "interview-3",
			CandidateName: "Charlie Brown",
			Questions:     []string{"Q3"},
			InterviewType: "general",
			Status:        "pending",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	for _, interview := range interviews {
		err := store.CreateInterview(interview)
		if err != nil {
			t.Fatalf("failed to create interview %s: %v", interview.ID, err)
		}
	}

	// Test pagination
	t.Run("pagination", func(t *testing.T) {
		opts := data.ListInterviewsOptions{
			Limit: 2,
			Page:  1,
		}

		result, err := store.GetInterviewsWithOptions(opts)
		if err != nil {
			t.Fatalf("GetInterviewsWithOptions failed: %v", err)
		}

		if result.Total != 3 {
			t.Errorf("expected total 3, got %d", result.Total)
		}
		if len(result.Interviews) != 2 {
			t.Errorf("expected 2 interviews in page, got %d", len(result.Interviews))
		}
		if result.TotalPages != 2 {
			t.Errorf("expected 2 total pages, got %d", result.TotalPages)
		}
	})

	// Test filtering by candidate name
	t.Run("filter by candidate name", func(t *testing.T) {
		opts := data.ListInterviewsOptions{
			CandidateName: "alice",
			Limit:         10,
		}

		result, err := store.GetInterviewsWithOptions(opts)
		if err != nil {
			t.Fatalf("GetInterviewsWithOptions failed: %v", err)
		}

		if result.Total != 1 {
			t.Errorf("expected 1 interview, got %d", result.Total)
		}
		if len(result.Interviews) > 0 && result.Interviews[0].CandidateName != "Alice Johnson" {
			t.Errorf("expected Alice Johnson, got %s", result.Interviews[0].CandidateName)
		}
	})

	// Test filtering by status
	t.Run("filter by status", func(t *testing.T) {
		opts := data.ListInterviewsOptions{
			Status: "pending",
			Limit:  10,
		}

		result, err := store.GetInterviewsWithOptions(opts)
		if err != nil {
			t.Fatalf("GetInterviewsWithOptions failed: %v", err)
		}

		if result.Total != 2 {
			t.Errorf("expected 2 pending interviews, got %d", result.Total)
		}
	})

	// Test sorting by name
	t.Run("sort by name asc", func(t *testing.T) {
		opts := data.ListInterviewsOptions{
			SortBy:    "name",
			SortOrder: "asc",
			Limit:     10,
		}

		result, err := store.GetInterviewsWithOptions(opts)
		if err != nil {
			t.Fatalf("GetInterviewsWithOptions failed: %v", err)
		}

		if len(result.Interviews) >= 2 {
			first := result.Interviews[0].CandidateName
			second := result.Interviews[1].CandidateName
			if first > second {
				t.Errorf("expected ascending order, but %s > %s", first, second)
			}
		}
	})

	// Test date filtering
	t.Run("filter by date range", func(t *testing.T) {
		opts := data.ListInterviewsOptions{
			DateFrom: time.Now().Add(-90 * time.Minute),
			DateTo:   time.Now().Add(-30 * time.Minute),
			Limit:    10,
		}

		result, err := store.GetInterviewsWithOptions(opts)
		if err != nil {
			t.Fatalf("GetInterviewsWithOptions failed: %v", err)
		}

		if result.Total != 1 {
			t.Errorf("expected 1 interview in date range, got %d", result.Total)
		}
	})
}

func TestMemoryStore_EvaluationOperations(t *testing.T) {
	store := data.NewMemoryStore()

	// Test CreateEvaluation
	evaluation := &data.Evaluation{
		ID:          "test-eval-1",
		InterviewID: "test-interview-1",
		Answers: data.StringMap{
			"question_0": "Answer 1",
			"question_1": "Answer 2",
		},
		Score:     85.5,
		Feedback:  "Good performance",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := store.CreateEvaluation(evaluation)
	if err != nil {
		t.Fatalf("CreateEvaluation failed: %v", err)
	}

	// Test GetEvaluation
	retrieved, err := store.GetEvaluation("test-eval-1")
	if err != nil {
		t.Fatalf("GetEvaluation failed: %v", err)
	}

	if retrieved.ID != evaluation.ID {
		t.Errorf("expected ID %s, got %s", evaluation.ID, retrieved.ID)
	}
	if retrieved.Score != evaluation.Score {
		t.Errorf("expected Score %f, got %f", evaluation.Score, retrieved.Score)
	}

	// Test GetEvaluation with non-existent ID
	_, err = store.GetEvaluation("non-existent")
	if err == nil {
		t.Error("expected error for non-existent evaluation")
	}
}

func TestMemoryStore_ChatSessionOperations(t *testing.T) {
	store := data.NewMemoryStore()

	// Test CreateChatSession
	session := &data.ChatSession{
		ID:          "test-session-1",
		InterviewID: "test-interview-1",
		Status:      "active",
		StartedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := store.CreateChatSession(session)
	if err != nil {
		t.Fatalf("CreateChatSession failed: %v", err)
	}

	// Test GetChatSession
	retrieved, err := store.GetChatSession("test-session-1")
	if err != nil {
		t.Fatalf("GetChatSession failed: %v", err)
	}

	if retrieved.ID != session.ID {
		t.Errorf("expected ID %s, got %s", session.ID, retrieved.ID)
	}
	if retrieved.Status != session.Status {
		t.Errorf("expected Status %s, got %s", session.Status, retrieved.Status)
	}

	// Test UpdateChatSession
	session.Status = "completed"
	session.EndedAt = &time.Time{}
	*session.EndedAt = time.Now()

	err = store.UpdateChatSession(session)
	if err != nil {
		t.Fatalf("UpdateChatSession failed: %v", err)
	}

	updated, err := store.GetChatSession("test-session-1")
	if err != nil {
		t.Fatalf("GetChatSession after update failed: %v", err)
	}

	if updated.Status != "completed" {
		t.Errorf("expected Status completed, got %s", updated.Status)
	}

	// Test GetChatSession with non-existent ID
	_, err = store.GetChatSession("non-existent")
	if err == nil {
		t.Error("expected error for non-existent chat session")
	}

	// Test UpdateChatSession with non-existent ID
	nonExistent := &data.ChatSession{ID: "non-existent"}
	err = store.UpdateChatSession(nonExistent)
	if err == nil {
		t.Error("expected error for updating non-existent chat session")
	}
}

func TestMemoryStore_ChatMessageOperations(t *testing.T) {
	store := data.NewMemoryStore()

	// First create a chat session
	session := &data.ChatSession{
		ID:          "test-session-1",
		InterviewID: "test-interview-1",
		Status:      "active",
		StartedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := store.CreateChatSession(session)
	if err != nil {
		t.Fatalf("CreateChatSession failed: %v", err)
	}

	// Test AddChatMessage
	message1 := &data.ChatMessage{
		ID:        "test-msg-1",
		SessionID: "test-session-1",
		Type:      "user",
		Content:   "Hello",
		Timestamp: time.Now(),
	}

	err = store.AddChatMessage(message1)
	if err != nil {
		t.Fatalf("AddChatMessage failed: %v", err)
	}

	message2 := &data.ChatMessage{
		ID:        "test-msg-2",
		SessionID: "test-session-1",
		Type:      "ai",
		Content:   "Hi there!",
		Timestamp: time.Now(),
	}

	err = store.AddChatMessage(message2)
	if err != nil {
		t.Fatalf("AddChatMessage failed: %v", err)
	}

	// Test GetChatMessages
	messages, err := store.GetChatMessages("test-session-1")
	if err != nil {
		t.Fatalf("GetChatMessages failed: %v", err)
	}

	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}

	// Test GetChatMessages with non-existent session
	_, err = store.GetChatMessages("non-existent")
	if err == nil {
		t.Error("expected error for non-existent chat session")
	}

	// Test AddChatMessage to non-existent session
	invalidMessage := &data.ChatMessage{
		ID:        "test-msg-3",
		SessionID: "non-existent",
		Type:      "user",
		Content:   "Invalid",
		Timestamp: time.Now(),
	}

	err = store.AddChatMessage(invalidMessage)
	if err == nil {
		t.Error("expected error for adding message to non-existent session")
	}
}

func TestMemoryStore_ConcurrentAccess(t *testing.T) {
	store := data.NewMemoryStore()

	// Test concurrent interview creation
	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent interview creation
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			interview := &data.Interview{
				ID:            fmt.Sprintf("concurrent-interview-%d", id),
				CandidateName: fmt.Sprintf("Candidate %d", id),
				Questions:     []string{"Q1"},
				InterviewType: "technical",
				Status:        "pending",
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			err := store.CreateInterview(interview)
			if err != nil {
				t.Errorf("CreateInterview failed in goroutine %d: %v", id, err)
			}
		}(i)
	}

	// Concurrent interview reading
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := store.GetInterviews()
			if err != nil {
				t.Errorf("GetInterviews failed: %v", err)
			}
		}()
	}

	wg.Wait()

	// Verify all interviews were created
	interviews, err := store.GetInterviews()
	if err != nil {
		t.Fatalf("Final GetInterviews failed: %v", err)
	}

	if len(interviews) != numGoroutines {
		t.Errorf("expected %d interviews, got %d", numGoroutines, len(interviews))
	}
}

func TestMemoryStore_EdgeCases(t *testing.T) {
	store := data.NewMemoryStore()

	// Test empty options for GetInterviewsWithOptions
	t.Run("empty options", func(t *testing.T) {
		result, err := store.GetInterviewsWithOptions(data.ListInterviewsOptions{})
		if err != nil {
			t.Fatalf("GetInterviewsWithOptions with empty options failed: %v", err)
		}

		// Should have default values applied
		if result.Limit != 10 {
			t.Errorf("expected default limit 10, got %d", result.Limit)
		}
	})

	// Test large offset
	t.Run("large offset", func(t *testing.T) {
		opts := data.ListInterviewsOptions{
			Offset: 1000,
			Limit:  10,
		}

		result, err := store.GetInterviewsWithOptions(opts)
		if err != nil {
			t.Fatalf("GetInterviewsWithOptions with large offset failed: %v", err)
		}

		if len(result.Interviews) != 0 {
			t.Errorf("expected 0 interviews with large offset, got %d", len(result.Interviews))
		}
		if result.Total != 0 {
			t.Errorf("expected total 0, got %d", result.Total)
		}
	})

	// Test invalid sort parameters
	t.Run("invalid sort parameters", func(t *testing.T) {
		// Create a test interview
		interview := &data.Interview{
			ID:            "test-sort",
			CandidateName: "Test Candidate",
			Questions:     []string{"Q1"},
			InterviewType: "technical",
			Status:        "pending",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		err := store.CreateInterview(interview)
		if err != nil {
			t.Fatalf("CreateInterview failed: %v", err)
		}

		opts := data.ListInterviewsOptions{
			SortBy:    "invalid_field",
			SortOrder: "invalid_order",
			Limit:     10,
		}

		// Should not crash and should fall back to defaults
		result, err := store.GetInterviewsWithOptions(opts)
		if err != nil {
			t.Fatalf("GetInterviewsWithOptions with invalid sort failed: %v", err)
		}

		if result.Total != 1 {
			t.Errorf("expected 1 interview, got %d", result.Total)
		}
	})
}
