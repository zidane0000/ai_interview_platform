package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zidane0000/ai-interview-platform/config"
	"github.com/zidane0000/ai-interview-platform/data"
)

// Test utilities and helpers

// setupTestRouter creates a test router with mock configuration
func setupTestRouter() http.Handler {
	testConfig := &config.Config{
		Port:            "8080",
		OpenAIAPIKey:    "test-openai-key",
		GeminiAPIKey:    "test-gemini-key",
		ShutdownTimeout: 30 * time.Second,
	}
	// No frontend handler needed for tests (nil)
	return SetupRouter(testConfig, nil)
}

// clearMemoryStore clears all data from the memory store for test isolation
func clearMemoryStore() {
	var err error
	data.GlobalStore, err = data.NewHybridStore(data.BackendMemory, "")
	if err != nil {
		panic("Failed to initialize test store: " + err.Error())
	}
}

// createTestInterview creates a test interview and returns the response
func createTestInterview(t *testing.T, router http.Handler, req CreateInterviewRequestDTO) InterviewResponseDTO {
	t.Helper()
	b, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/interviews", bytes.NewReader(b))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to create interview, got %d: %s", w.Code, w.Body.String())
	}

	var resp InterviewResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal interview response: %v", err)
	}
	return resp
}

// startChatSession starts a chat session for an interview
func startChatSession(t *testing.T, router http.Handler, interviewID string, req *StartChatSessionRequestDTO) ChatInterviewSessionDTO {
	t.Helper()
	var body []byte
	if req != nil {
		body, _ = json.Marshal(req)
	}

	httpReq := httptest.NewRequest("POST", "/api/interviews/"+interviewID+"/chat/start", bytes.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Fatalf("failed to start chat session, got %d: %s", w.Code, w.Body.String())
	}

	var resp ChatInterviewSessionDTO
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal chat session response: %v", err)
	}
	return resp
}

// sendMessage sends a message in a chat session
func sendMessage(t *testing.T, router http.Handler, sessionID, message string) SendMessageResponseDTO {
	t.Helper()
	req := SendMessageRequestDTO{Message: message}
	b, _ := json.Marshal(req)

	httpReq := httptest.NewRequest("POST", "/api/chat/"+sessionID+"/message", bytes.NewReader(b))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Fatalf("failed to send message, got %d: %s", w.Code, w.Body.String())
	}

	var resp SendMessageResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal message response: %v", err)
	}
	return resp
}

// expectHTTPError expects a specific HTTP error status
func expectHTTPError(t *testing.T, router http.Handler, method, path string, body []byte, expectedStatus int) {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != expectedStatus {
		t.Errorf("expected %d, got %d: %s", expectedStatus, w.Code, w.Body.String())
	}
}

// create interview and start chat session for tests
func createTestInterviewAndSession(t *testing.T, router http.Handler) struct {
	InterviewID string
	SessionID   string
} {
	t.Helper()

	// Create interview using helper
	interview := createTestInterview(t, router, CreateInterviewRequestDTO{
		CandidateName: "Test User",
		Questions:     []string{"Q1", "Q2"},
		InterviewType: "general",
	})

	// Start chat session using helper
	session := startChatSession(t, router, interview.ID, nil)

	return struct {
		InterviewID string
		SessionID   string
	}{
		InterviewID: interview.ID,
		SessionID:   session.ID,
	}
}

// ============================================
// INTERVIEW CRUD TESTS
// ============================================

func TestCreateInterviewHandler_Success(t *testing.T) {
	clearMemoryStore()

	req := CreateInterviewRequestDTO{
		CandidateName: "Alice",
		Questions:     []string{"Q1", "Q2"},
		InterviewType: "general",
	}

	b, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/interviews", bytes.NewReader(b))
	w := httptest.NewRecorder()
	CreateInterviewHandler(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 Created, got %d", w.Code)
	}
}

func TestCreateInterviewHandler_BadRequest(t *testing.T) {
	router := setupTestRouter()

	// Invalid JSON
	expectHTTPError(t, router, "POST", "/api/interviews", []byte("{"), http.StatusBadRequest)

	// Missing fields
	emptyReq := CreateInterviewRequestDTO{}
	b, _ := json.Marshal(emptyReq)
	expectHTTPError(t, router, "POST", "/api/interviews", b, http.StatusBadRequest)
}

func TestCreateInterviewHandler_EdgeCases(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	tests := []struct {
		name           string
		body           CreateInterviewRequestDTO
		expectedStatus int
	}{
		{
			name: "missing candidate name",
			body: CreateInterviewRequestDTO{
				Questions:     []string{"Q1"},
				InterviewType: "general",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty questions",
			body: CreateInterviewRequestDTO{
				CandidateName: "Test",
				Questions:     []string{},
				InterviewType: "general",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing interview type",
			body: CreateInterviewRequestDTO{
				CandidateName: "Test",
				Questions:     []string{"Q1"},
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid interview type",
			body: CreateInterviewRequestDTO{
				CandidateName: "Test",
				Questions:     []string{"Q1"},
				InterviewType: "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, _ := json.Marshal(tt.body)
			expectHTTPError(t, router, "POST", "/api/interviews", b, tt.expectedStatus)
		})
	}
}

func TestListInterviewsHandler_Empty(t *testing.T) {
	clearMemoryStore() // Clear store for test isolation
	router := setupTestRouter()
	req := httptest.NewRequest("GET", "/api/interviews", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	var resp ListInterviewsResponseDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Interviews) != 0 {
		t.Errorf("expected empty interviews list, got %d interviews", len(resp.Interviews))
	}
	if resp.Total != 0 {
		t.Errorf("expected total count of 0, got %d", resp.Total)
	}
}

func TestListInterviewsHandler_WithData(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	// Create multiple test interviews using helper
	interviews := []CreateInterviewRequestDTO{
		{CandidateName: "Alice Johnson", Questions: []string{"Q1", "Q2"}, InterviewType: "general"},
		{CandidateName: "Bob Smith", Questions: []string{"Q3", "Q4"}, InterviewType: "technical"},
		{CandidateName: "Charlie Brown", Questions: []string{"Q5", "Q6"}, InterviewType: "behavioral"},
	}

	for _, interview := range interviews {
		createTestInterview(t, router, interview)
	}

	// Test listing all interviews
	req := httptest.NewRequest("GET", "/api/interviews", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	var resp ListInterviewsResponseDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Interviews) != 3 {
		t.Errorf("expected 3 interviews, got %d", len(resp.Interviews))
	}
	if resp.Total != 3 {
		t.Errorf("expected total count of 3, got %d", resp.Total)
	}
}

func TestListInterviewsHandler_Pagination(t *testing.T) {
	clearMemoryStore() // Clear store for test isolation
	router := setupTestRouter()

	// Create 5 test interviews
	for i := 1; i <= 5; i++ {
		interview := CreateInterviewRequestDTO{
			CandidateName: fmt.Sprintf("Candidate %d", i),
			Questions:     []string{"Q1", "Q2"},
			InterviewType: "general", // Add required interview type
		}
		b, _ := json.Marshal(interview)
		req := httptest.NewRequest("POST", "/api/interviews", bytes.NewReader(b))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("failed to create interview %d, got %d", i, w.Code)
		}
	}

	// Test pagination with limit=2
	req := httptest.NewRequest("GET", "/api/interviews?limit=2", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	var resp ListInterviewsResponseDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Interviews) != 2 {
		t.Errorf("expected 2 interviews with limit=2, got %d", len(resp.Interviews))
	}
	if resp.Total != 5 {
		t.Errorf("expected total count of 5, got %d", resp.Total)
	}

	// Test pagination with offset
	req = httptest.NewRequest("GET", "/api/interviews?limit=2&offset=2", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Interviews) != 2 {
		t.Errorf("expected 2 interviews with offset=2, got %d", len(resp.Interviews))
	}
	if resp.Total != 5 {
		t.Errorf("expected total count of 5, got %d", resp.Total)
	}
}

func TestListInterviewsHandler_Filtering(t *testing.T) {
	clearMemoryStore() // Clear store for test isolation
	router := setupTestRouter()

	// Create test interviews with different names
	interviews := []CreateInterviewRequestDTO{
		{CandidateName: "Alice Johnson", Questions: []string{"Q1", "Q2"}, InterviewType: "general"},
		{CandidateName: "Bob Alice", Questions: []string{"Q3", "Q4"}, InterviewType: "technical"},
		{CandidateName: "Charlie Brown", Questions: []string{"Q5", "Q6"}, InterviewType: "behavioral"},
	}

	for _, interview := range interviews {
		b, _ := json.Marshal(interview)
		req := httptest.NewRequest("POST", "/api/interviews", bytes.NewReader(b))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("failed to create interview for %s, got %d", interview.CandidateName, w.Code)
		}
	}

	// Test filtering by candidate name
	req := httptest.NewRequest("GET", "/api/interviews?candidate_name=Alice", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	var resp ListInterviewsResponseDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Interviews) != 2 {
		t.Errorf("expected 2 interviews containing 'Alice', got %d", len(resp.Interviews))
	}
	if resp.Total != 2 {
		t.Errorf("expected total count of 2, got %d", resp.Total)
	}

	// Verify the filtered results contain "Alice"
	for _, interview := range resp.Interviews {
		if !strings.Contains(strings.ToLower(interview.CandidateName), "alice") {
			t.Errorf("expected interview name to contain 'alice', got %s", interview.CandidateName)
		}
	}
}

func TestListInterviewsHandler_Sorting(t *testing.T) {
	clearMemoryStore() // Clear store for test isolation
	router := setupTestRouter()

	// Create test interviews in a specific order
	interviews := []CreateInterviewRequestDTO{
		{CandidateName: "Charlie Brown", Questions: []string{"Q1", "Q2"}, InterviewType: "general"},
		{CandidateName: "Alice Johnson", Questions: []string{"Q3", "Q4"}, InterviewType: "technical"},
		{CandidateName: "Bob Smith", Questions: []string{"Q5", "Q6"}, InterviewType: "behavioral"},
	}

	for _, interview := range interviews {
		b, _ := json.Marshal(interview)
		req := httptest.NewRequest("POST", "/api/interviews", bytes.NewReader(b))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("failed to create interview for %s, got %d", interview.CandidateName, w.Code)
		}
		// Add small delay to ensure different creation times
		time.Sleep(1 * time.Millisecond)
	}

	// Test sorting by name ascending
	req := httptest.NewRequest("GET", "/api/interviews?sort_by=name&sort_order=asc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}

	var resp ListInterviewsResponseDTO
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(resp.Interviews) != 3 {
		t.Errorf("expected 3 interviews, got %d", len(resp.Interviews))
	}

	// Verify the sorting order
	expectedOrder := []string{"Alice Johnson", "Bob Smith", "Charlie Brown"}
	for i, interview := range resp.Interviews {
		if interview.CandidateName != expectedOrder[i] {
			t.Errorf("expected interview %d to be %s, got %s", i, expectedOrder[i], interview.CandidateName)
		}
	}
}

func TestGetInterviewHandler_BadRequest(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest("GET", "/api/interviews/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
}

func TestGetInterviewHandler_Success(t *testing.T) {
	clearMemoryStore() // Clear store for test isolation
	router := setupTestRouter()

	// Step 1: Create an interview
	createBody := CreateInterviewRequestDTO{
		CandidateName: "Test User",
		Questions:     []string{"Q1", "Q2"},
		InterviewType: "general", // Add required interview type
	}
	b, _ := json.Marshal(createBody)
	createReq := httptest.NewRequest("POST", "/api/interviews", bytes.NewReader(b))
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("failed to create interview, got %d", createW.Code)
	}
	var createdResp InterviewResponseDTO
	if err := json.Unmarshal(createW.Body.Bytes(), &createdResp); err != nil {
		t.Fatalf("failed to decode create response: %v", err)
	}

	// Step 2: Use the real ID for GET
	req := httptest.NewRequest("GET", "/api/interviews/"+createdResp.ID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
}

func TestSubmitEvaluationHandler_Success(t *testing.T) {
	clearMemoryStore() // Clear store for test isolation
	// First create a valid interview
	interview := &data.Interview{
		ID:            "test-interview-123",
		CandidateName: "Test Candidate",
		Questions:     []string{"What is your experience?", "Tell me about yourself"}, CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := data.GlobalStore.CreateInterview(interview); err != nil {
		t.Fatalf("failed to create interview: %v", err)
	}

	body := SubmitEvaluationRequestDTO{
		InterviewID: "test-interview-123",
		Answers:     map[string]string{"question_0": "5 years of experience", "question_1": "I am a developer"},
	}
	b, _ := json.Marshal(body)

	router := setupTestRouter()
	req := httptest.NewRequest("POST", "/api/evaluation", bytes.NewReader(b))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
}

func TestSubmitEvaluationHandler_BadRequest(t *testing.T) {
	router := setupTestRouter()

	// Invalid JSON
	req := httptest.NewRequest("POST", "/api/evaluation", bytes.NewReader([]byte("{")))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}

	// Missing fields
	body := SubmitEvaluationRequestDTO{}
	b, _ := json.Marshal(body)
	req = httptest.NewRequest("POST", "/api/evaluation", bytes.NewReader(b))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request for missing fields, got %d", w.Code)
	}
}

func TestGetEvaluationHandler_BadRequest(t *testing.T) {
	router := setupTestRouter()
	req := httptest.NewRequest("GET", "/api/evaluation/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 Method Not Allowed, got %d", w.Code)
	}
}

func TestGetEvaluationHandler_Success(t *testing.T) {
	clearMemoryStore() // Clear store for test isolation
	// First create a valid evaluation
	evaluation := &data.Evaluation{
		ID:          "test-evaluation-456",
		InterviewID: "test-interview-456",
		Answers:     map[string]string{"question_0": "Test answer"},
		Score:       0.8,
		Feedback:    "Good performance",
		CreatedAt:   time.Now(), UpdatedAt: time.Now(),
	}
	if err := data.GlobalStore.CreateEvaluation(evaluation); err != nil {
		t.Fatalf("failed to create evaluation: %v", err)
	}

	router := setupTestRouter()
	req := httptest.NewRequest("GET", "/api/evaluation/test-evaluation-456", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
	}
}

// ============================================
// CHAT SESSION HANDLER TESTS
// ============================================

func TestStartChatSessionHandler_Success(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	// Create interview using helper
	interview := createTestInterview(t, router, CreateInterviewRequestDTO{
		CandidateName: "Test User",
		Questions:     []string{"Q1", "Q2"},
		InterviewType: "general",
	})

	// Start chat session using helper
	session := startChatSession(t, router, interview.ID, nil)

	// Verify response structure
	if session.ID == "" {
		t.Error("expected session ID to be set")
	}
	if session.InterviewID != interview.ID {
		t.Errorf("expected interview ID %s, got %s", interview.ID, session.InterviewID)
	}
	if session.Status != "active" {
		t.Errorf("expected status 'active', got %s", session.Status)
	}
	if len(session.Messages) == 0 {
		t.Error("expected at least one initial AI message")
	}
}

func TestStartChatSessionHandler_WithLanguage(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	// Create interview with specific language
	interview := createTestInterview(t, router, CreateInterviewRequestDTO{
		CandidateName:     "Test User",
		Questions:         []string{"Q1", "Q2"},
		InterviewType:     "general",
		InterviewLanguage: "zh-TW",
	})
	// Start chat session with language override
	sessionReq := &StartChatSessionRequestDTO{
		SessionLanguage: "en",
	}
	session := startChatSession(t, router, interview.ID, sessionReq)
	// Should use the overridden language
	if session.SessionLanguage != "en" {
		t.Errorf("expected language 'en', got %s", session.SessionLanguage)
	}
}

func TestStartChatSessionHandler_InvalidInterview(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	expectHTTPError(t, router, "POST", "/api/interviews/nonexistent/chat/start", nil, http.StatusNotFound)
}

func TestStartChatSessionHandler_MissingInterviewID(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	expectHTTPError(t, router, "POST", "/api/interviews//chat/start", nil, http.StatusBadRequest)
}

func TestSendMessageHandler_Success(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	interview := createTestInterviewAndSession(t, router)

	// Send a message using helper
	response := sendMessage(t, router, interview.SessionID, "Hello, this is my test message")

	// Should have user message and AI response
	if response.Message.Content != "Hello, this is my test message" {
		t.Errorf("expected user message content to match, got %s", response.Message.Content)
	}
	if response.AIResponse == nil {
		t.Error("expected AI response to be present")
	}
}

func TestSendMessageHandler_EmptyMessage(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	interview := createTestInterviewAndSession(t, router)

	emptyReq := SendMessageRequestDTO{Message: ""}
	b, _ := json.Marshal(emptyReq)
	expectHTTPError(t, router, "POST", "/api/chat/"+interview.SessionID+"/message", b, http.StatusBadRequest)
}

func TestSendMessageHandler_InvalidSession(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	req := SendMessageRequestDTO{Message: "Hello"}
	b, _ := json.Marshal(req)
	expectHTTPError(t, router, "POST", "/api/chat/nonexistent/message", b, http.StatusNotFound)
}

func TestSendMessageHandler_InvalidJSON(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	interview := createTestInterviewAndSession(t, router)
	expectHTTPError(t, router, "POST", "/api/chat/"+interview.SessionID+"/message", []byte("{"), http.StatusBadRequest)
}

func TestGetChatSessionHandler_Success(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	interview := createTestInterviewAndSession(t, router)

	req := httptest.NewRequest("GET", "/api/chat/"+interview.SessionID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
		return
	}

	var response ChatInterviewSessionDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal chat session response: %v", err)
	}

	if response.ID != interview.SessionID {
		t.Errorf("expected session ID %s, got %s", interview.SessionID, response.ID)
	}
}

func TestGetChatSessionHandler_NotFound(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	req := httptest.NewRequest("GET", "/api/chat/nonexistent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 Not Found, got %d", w.Code)
	}
}

func TestEndChatSessionHandler_Success(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	interview := createTestInterviewAndSession(t, router)

	// Send a test message before ending to ensure there's content to evaluate
	sendMessage(t, router, interview.SessionID, "I have 5 years of experience in software development")

	// End the session
	req := httptest.NewRequest("POST", "/api/chat/"+interview.SessionID+"/end", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", w.Code)
		return
	}

	var response EvaluationResponseDTO
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal evaluation response: %v", err)
	}

	if response.Score <= 0 {
		t.Errorf("expected score > 0, got %f", response.Score)
	}
	if response.Feedback == "" {
		t.Error("expected feedback to be present")
	}
	if response.ID == "" {
		t.Error("expected evaluation ID to be present")
	}
}

func TestEndChatSessionHandler_NotFound(t *testing.T) {
	clearMemoryStore()
	router := setupTestRouter()

	expectHTTPError(t, router, "POST", "/api/chat/nonexistent/end", nil, http.StatusNotFound)
}

// ============================================
// ADDITIONAL EDGE CASE TESTS
// ============================================

func TestParseIntQuery_EdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		queryValue   string
		defaultValue int
		expected     int
	}{
		{"empty string", "", 10, 10},
		{"valid positive", "5", 10, 5},
		{"zero", "0", 10, 0},
		{"negative", "-1", 10, 10}, // Should use default for negative
		{"invalid string", "abc", 10, 10},
		{"float", "5.5", 10, 10}, // Should use default for non-int
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?test="+tt.queryValue, nil)
			result := parseIntQuery(req, "test", tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}
