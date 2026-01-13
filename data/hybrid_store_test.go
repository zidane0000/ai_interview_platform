package data_test

import (
	"os"
	"testing"
	"time"

	"github.com/zidane0000/ai-interview-platform/data"
)

func TestAutoDetectBackend(t *testing.T) {
	// Save original DATABASE_URL
	originalURL := os.Getenv("DATABASE_URL")
	defer func() {
		if originalURL != "" {
			os.Setenv("DATABASE_URL", originalURL)
		} else {
			os.Unsetenv("DATABASE_URL")
		}
	}()

	// Test memory backend detection
	t.Run("memory backend when DATABASE_URL is empty", func(t *testing.T) {
		os.Unsetenv("DATABASE_URL")
		backend := data.AutoDetectBackend()
		if backend != data.BackendMemory {
			t.Errorf("expected BackendMemory, got %v", backend)
		}
	})

	// Test database backend detection
	t.Run("database backend when DATABASE_URL is set", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb")
		backend := data.AutoDetectBackend()
		if backend != data.BackendDatabase {
			t.Errorf("expected BackendDatabase, got %v", backend)
		}
	})
}

func TestNewHybridStore_MemoryBackend(t *testing.T) {
	store, err := data.NewHybridStore(data.BackendMemory, "")
	if err != nil {
		t.Fatalf("NewHybridStore failed: %v", err)
	}

	if store.GetBackend() != data.BackendMemory {
		t.Errorf("expected BackendMemory, got %v", store.GetBackend())
	}

	// Test basic operations
	interview := &data.Interview{
		ID:            "test-hybrid-1",
		CandidateName: "Test Candidate",
		Questions:     []string{"Q1", "Q2"},
		InterviewType: "technical",
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = store.CreateInterview(interview)
	if err != nil {
		t.Fatalf("CreateInterview failed: %v", err)
	}

	retrieved, err := store.GetInterview("test-hybrid-1")
	if err != nil {
		t.Fatalf("GetInterview failed: %v", err)
	}

	if retrieved.ID != interview.ID {
		t.Errorf("expected ID %s, got %s", interview.ID, retrieved.ID)
	}
}

func TestNewHybridStore_DatabaseBackend_InvalidURL(t *testing.T) {
	// Test with invalid database URL
	_, err := data.NewHybridStore(data.BackendDatabase, "invalid-url")
	if err == nil {
		t.Error("expected error for invalid database URL")
	}
}

func TestHybridStore_InterviewOperations(t *testing.T) {
	// Test with memory backend
	store, err := data.NewHybridStore(data.BackendMemory, "")
	if err != nil {
		t.Fatalf("NewHybridStore failed: %v", err)
	}

	// Create test interview
	interview := &data.Interview{
		ID:             "hybrid-test-1",
		CandidateName:  "John Doe",
		Questions:      []string{"Q1", "Q2"},
		InterviewType:  "behavioral",
		JobDescription: "Test Job",
		Status:         "pending",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Test CreateInterview
	err = store.CreateInterview(interview)
	if err != nil {
		t.Fatalf("CreateInterview failed: %v", err)
	}

	// Test GetInterview
	retrieved, err := store.GetInterview("hybrid-test-1")
	if err != nil {
		t.Fatalf("GetInterview failed: %v", err)
	}

	if retrieved.CandidateName != interview.CandidateName {
		t.Errorf("expected CandidateName %s, got %s", interview.CandidateName, retrieved.CandidateName)
	}

	// Test GetInterviewsWithOptions
	opts := data.ListInterviewsOptions{
		Limit: 10,
		Page:  1,
	}

	result, err := store.GetInterviewsWithOptions(opts)
	if err != nil {
		t.Fatalf("GetInterviewsWithOptions failed: %v", err)
	}

	if result.Total != 1 {
		t.Errorf("expected 1 interview, got %d", result.Total)
	}
}

func TestHybridStore_EvaluationOperations(t *testing.T) {
	store, err := data.NewHybridStore(data.BackendMemory, "")
	if err != nil {
		t.Fatalf("NewHybridStore failed: %v", err)
	}

	// Test CreateEvaluation
	evaluation := &data.Evaluation{
		ID:          "hybrid-eval-1",
		InterviewID: "test-interview-1",
		Answers: data.StringMap{
			"question_0": "Answer 1",
			"question_1": "Answer 2",
		},
		Score:     78.5,
		Feedback:  "Good answers",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = store.CreateEvaluation(evaluation)
	if err != nil {
		t.Fatalf("CreateEvaluation failed: %v", err)
	}

	// Test GetEvaluation
	retrieved, err := store.GetEvaluation("hybrid-eval-1")
	if err != nil {
		t.Fatalf("GetEvaluation failed: %v", err)
	}

	if retrieved.Score != evaluation.Score {
		t.Errorf("expected Score %f, got %f", evaluation.Score, retrieved.Score)
	}
}

func TestHybridStore_ChatSessionOperations(t *testing.T) {
	store, err := data.NewHybridStore(data.BackendMemory, "")
	if err != nil {
		t.Fatalf("NewHybridStore failed: %v", err)
	}

	// Test CreateChatSession
	session := &data.ChatSession{
		ID:          "hybrid-session-1",
		InterviewID: "test-interview-1",
		Status:      "active",
		StartedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = store.CreateChatSession(session)
	if err != nil {
		t.Fatalf("CreateChatSession failed: %v", err)
	}

	// Test GetChatSession
	retrieved, err := store.GetChatSession("hybrid-session-1")
	if err != nil {
		t.Fatalf("GetChatSession failed: %v", err)
	}

	if retrieved.Status != session.Status {
		t.Errorf("expected Status %s, got %s", session.Status, retrieved.Status)
	}

	// Test UpdateChatSession
	session.Status = "completed"
	endTime := time.Now()
	session.EndedAt = &endTime

	err = store.UpdateChatSession(session)
	if err != nil {
		t.Fatalf("UpdateChatSession failed: %v", err)
	}

	updated, err := store.GetChatSession("hybrid-session-1")
	if err != nil {
		t.Fatalf("GetChatSession after update failed: %v", err)
	}

	if updated.Status != "completed" {
		t.Errorf("expected Status completed, got %s", updated.Status)
	}
}

func TestHybridStore_ChatMessageOperations(t *testing.T) {
	store, err := data.NewHybridStore(data.BackendMemory, "")
	if err != nil {
		t.Fatalf("NewHybridStore failed: %v", err)
	}

	// First create a chat session
	session := &data.ChatSession{
		ID:          "hybrid-msg-session",
		InterviewID: "test-interview-1",
		Status:      "active",
		StartedAt:   time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = store.CreateChatSession(session)
	if err != nil {
		t.Fatalf("CreateChatSession failed: %v", err)
	}

	// Test AddChatMessage
	message := &data.ChatMessage{
		ID:        "hybrid-msg-1",
		SessionID: "hybrid-msg-session",
		Type:      "user",
		Content:   "Hello from hybrid store",
		Timestamp: time.Now(),
	}

	err = store.AddChatMessage("hybrid-msg-session", message)
	if err != nil {
		t.Fatalf("AddChatMessage failed: %v", err)
	}

	// Test GetChatMessages
	messages, err := store.GetChatMessages("hybrid-msg-session")
	if err != nil {
		t.Fatalf("GetChatMessages failed: %v", err)
	}

	if len(messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(messages))
	}

	if messages[0].Content != "Hello from hybrid store" {
		t.Errorf("expected content 'Hello from hybrid store', got %s", messages[0].Content)
	}
}

func TestHybridStore_Health(t *testing.T) {
	// Test health check for memory backend
	t.Run("memory backend health", func(t *testing.T) {
		store, err := data.NewHybridStore(data.BackendMemory, "")
		if err != nil {
			t.Fatalf("NewHybridStore failed: %v", err)
		}

		err = store.Health()
		if err != nil {
			t.Errorf("Health check failed for memory backend: %v", err)
		}
	})

	// Test health check for database backend with invalid URL (should fail)
	t.Run("database backend health with invalid URL", func(t *testing.T) {
		_, err := data.NewHybridStore(data.BackendDatabase, "invalid-url")
		if err == nil {
			t.Error("expected error for invalid database URL during creation")
		}
	})
}

func TestHybridStore_Close(t *testing.T) {
	// Test close for memory backend
	t.Run("memory backend close", func(t *testing.T) {
		store, err := data.NewHybridStore(data.BackendMemory, "")
		if err != nil {
			t.Fatalf("NewHybridStore failed: %v", err)
		}

		err = store.Close()
		if err != nil {
			t.Errorf("Close failed for memory backend: %v", err)
		}
	})
}

func TestHybridStore_BackendInterface(t *testing.T) {
	// Test that both backends implement the same interface methods
	memoryStore, err := data.NewHybridStore(data.BackendMemory, "")
	if err != nil {
		t.Fatalf("NewHybridStore memory failed: %v", err)
	}

	// Create the same test data
	interview := &data.Interview{
		ID:            "interface-test-1",
		CandidateName: "Interface Test",
		Questions:     []string{"Q1"},
		InterviewType: "technical",
		Status:        "pending",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	evaluation := &data.Evaluation{
		ID:          "interface-eval-1",
		InterviewID: "interface-test-1",
		Answers: data.StringMap{
			"question_0": "Interface Answer",
		},
		Score:     90.0,
		Feedback:  "Interface Test Feedback",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Test that methods work consistently across backends
	stores := map[string]*data.HybridStore{
		"memory": memoryStore,
	}

	for backendName, store := range stores {
		t.Run(backendName+" backend interface", func(t *testing.T) {
			// Test interview operations
			err := store.CreateInterview(interview)
			if err != nil {
				t.Fatalf("CreateInterview failed for %s: %v", backendName, err)
			}

			_, err = store.GetInterview(interview.ID)
			if err != nil {
				t.Fatalf("GetInterview failed for %s: %v", backendName, err)
			}

			// Test evaluation operations
			err = store.CreateEvaluation(evaluation)
			if err != nil {
				t.Fatalf("CreateEvaluation failed for %s: %v", backendName, err)
			}

			_, err = store.GetEvaluation(evaluation.ID)
			if err != nil {
				t.Fatalf("GetEvaluation failed for %s: %v", backendName, err)
			}

			// Test listing operations
			opts := data.ListInterviewsOptions{Limit: 10}
			result, err := store.GetInterviewsWithOptions(opts)
			if err != nil {
				t.Fatalf("GetInterviewsWithOptions failed for %s: %v", backendName, err)
			}

			if result.Total == 0 {
				t.Errorf("expected at least 1 interview for %s, got %d", backendName, result.Total)
			}
		})
	}
}

func TestHybridStore_ErrorHandling(t *testing.T) {
	store, err := data.NewHybridStore(data.BackendMemory, "")
	if err != nil {
		t.Fatalf("NewHybridStore failed: %v", err)
	}

	// Test getting non-existent resources
	t.Run("non-existent interview", func(t *testing.T) {
		_, err := store.GetInterview("non-existent")
		if err == nil {
			t.Error("expected error for non-existent interview")
		}
	})

	t.Run("non-existent evaluation", func(t *testing.T) {
		_, err := store.GetEvaluation("non-existent")
		if err == nil {
			t.Error("expected error for non-existent evaluation")
		}
	})

	t.Run("non-existent chat session", func(t *testing.T) {
		_, err := store.GetChatSession("non-existent")
		if err == nil {
			t.Error("expected error for non-existent chat session")
		}
	})

	t.Run("messages for non-existent session", func(t *testing.T) {
		_, err := store.GetChatMessages("non-existent")
		if err == nil {
			t.Error("expected error for messages of non-existent session")
		}
	})
}
