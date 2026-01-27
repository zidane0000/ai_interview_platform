package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewBaseProvider verifies BaseProvider initialization
func TestNewBaseProvider(t *testing.T) {
	config := &AIConfig{
		DefaultModel:   "test-model",
		RequestTimeout: 30 * time.Second,
	}

	bp := NewBaseProvider(config, "https://api.example.com", 10*time.Second)

	if bp.config != config {
		t.Error("Expected config to be set")
	}
	if bp.baseURL != "https://api.example.com" {
		t.Errorf("Expected baseURL to be 'https://api.example.com', got '%s'", bp.baseURL)
	}
	if bp.httpClient == nil {
		t.Error("Expected httpClient to be initialized")
	}
	if bp.httpClient.Timeout != 10*time.Second {
		t.Errorf("Expected timeout to be 10s, got %v", bp.httpClient.Timeout)
	}
}

// TestGetModelName tests the model fallback/precedence logic
func TestGetModelName(t *testing.T) {
	testCases := []struct {
		name          string
		model         string
		defaultModel  string
		configDefault string
		expected      string
	}{
		{
			name:          "use provided model",
			model:         "model-a",
			defaultModel:  "model-b",
			configDefault: "config-default",
			expected:      "model-a",
		},
		{
			name:          "use default when model empty",
			model:         "",
			defaultModel:  "model-b",
			configDefault: "config-default",
			expected:      "model-b",
		},
		{
			name:          "use config default when both empty",
			model:         "",
			defaultModel:  "",
			configDefault: "config-default",
			expected:      "config-default",
		},
		{
			name:          "return empty when all empty",
			model:         "",
			defaultModel:  "",
			configDefault: "",
			expected:      "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &AIConfig{
				DefaultModel: tc.configDefault,
			}
			bp := NewBaseProvider(config, "https://api.example.com", 10*time.Second)

			result := bp.GetModelName(tc.model, tc.defaultModel)

			if result != tc.expected {
				t.Errorf("Expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

// mockAdapter implements ProviderAdapter for testing
type mockAdapter struct {
	authHeader string
	baseURL    string
}

func (m *mockAdapter) SetAuth(req *http.Request) {
	if m.authHeader != "" {
		req.Header.Set("Authorization", m.authHeader)
	}
}

func (m *mockAdapter) GetEndpointURL(endpoint string) string {
	return m.baseURL + endpoint
}

// TestMakeRequest tests HTTP request handling
func TestMakeRequest(t *testing.T) {
	testCases := []struct {
		name           string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
	}{
		{
			name:           "success 200 OK",
			serverResponse: `{"status": "ok"}`,
			serverStatus:   http.StatusOK,
			expectError:    false,
		},
		{
			name:           "server error 500",
			serverResponse: `{"error": "internal error"}`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
			errorContains:  "status 500",
		},
		{
			name:           "bad request 400",
			serverResponse: `{"error": "bad request"}`,
			serverStatus:   http.StatusBadRequest,
			expectError:    true,
			errorContains:  "status 400",
		},
		{
			name:           "unauthorized 401",
			serverResponse: `{"error": "unauthorized"}`,
			serverStatus:   http.StatusUnauthorized,
			expectError:    true,
			errorContains:  "status 401",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				// Verify content type
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				w.WriteHeader(tc.serverStatus)
				w.Write([]byte(tc.serverResponse))
			}))
			defer server.Close()

			config := &AIConfig{}
			bp := NewBaseProvider(config, server.URL, 10*time.Second)

			adapter := &mockAdapter{
				authHeader: "Bearer test-key",
				baseURL:    server.URL,
			}

			payload := map[string]string{"test": "data"}
			body, err := bp.MakeRequest(context.Background(), adapter, "/test", payload)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
				if string(body) != tc.serverResponse {
					t.Errorf("Expected body '%s', got '%s'", tc.serverResponse, string(body))
				}
			}
		})
	}
}

// TestMakeRequest_Timeout tests context timeout handling
func TestMakeRequest_Timeout(t *testing.T) {
	// Create a server that sleeps longer than timeout
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &AIConfig{}
	bp := NewBaseProvider(config, server.URL, 50*time.Millisecond)

	adapter := &mockAdapter{baseURL: server.URL}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := bp.MakeRequest(ctx, adapter, "/test", map[string]string{})

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

// TestMakeRequest_AuthHeader verifies auth header is set
func TestMakeRequest_AuthHeader(t *testing.T) {
	var receivedAuthHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuthHeader = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	config := &AIConfig{}
	bp := NewBaseProvider(config, server.URL, 10*time.Second)

	adapter := &mockAdapter{
		authHeader: "Bearer my-secret-key",
		baseURL:    server.URL,
	}

	_, err := bp.MakeRequest(context.Background(), adapter, "/test", map[string]string{})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if receivedAuthHeader != "Bearer my-secret-key" {
		t.Errorf("Expected auth header 'Bearer my-secret-key', got '%s'", receivedAuthHeader)
	}
}

// TestBuildQuestionGenerationPrompt verifies prompt construction
func TestBuildQuestionGenerationPrompt(t *testing.T) {
	req := &QuestionGenerationRequest{
		ExperienceLevel: "senior",
		InterviewType:   "technical",
		Difficulty:      "hard",
		JobDescription:  "Build scalable distributed systems",
		ResumeContent:   "10 years experience in backend development",
		NumQuestions:    5,
	}

	prompt := BuildQuestionGenerationPrompt(req)

	// Verify all fields are included in prompt
	expectedContains := []string{
		"senior",
		"technical",
		"hard",
		"Build scalable distributed systems",
		"10 years experience in backend development",
		"5",
	}

	for _, expected := range expectedContains {
		if !strings.Contains(prompt, expected) {
			t.Errorf("Expected prompt to contain '%s'", expected)
		}
	}

	// Verify formatting instructions are present
	if !strings.Contains(prompt, "Question:") {
		t.Error("Expected prompt to contain 'Question:' format instruction")
	}
	if !strings.Contains(prompt, "Expected Time:") {
		t.Error("Expected prompt to contain 'Expected Time:' format instruction")
	}
}

// TestBuildEvaluationPrompt verifies evaluation prompt construction
func TestBuildEvaluationPrompt(t *testing.T) {
	testCases := []struct {
		name             string
		criteria         []string
		expectedContains []string
	}{
		{
			name:     "with criteria",
			criteria: []string{"technical skills", "communication", "problem solving"},
			expectedContains: []string{
				"technical skills, communication, problem solving",
				"Overall Score:",
				"Feedback:",
			},
		},
		{
			name:             "empty criteria",
			criteria:         []string{},
			expectedContains: []string{"Overall Score:", "Feedback:"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := &EvaluationRequest{
				JobDesc:     "Software Engineer position",
				Criteria:    tc.criteria,
				DetailLevel: "detailed",
			}

			prompt := BuildEvaluationPrompt(req)

			for _, expected := range tc.expectedContains {
				if !strings.Contains(prompt, expected) {
					t.Errorf("Expected prompt to contain '%s'", expected)
				}
			}
		})
	}
}

// TestFormatAnswersForEvaluation tests Q&A formatting
func TestFormatAnswersForEvaluation(t *testing.T) {
	testCases := []struct {
		name             string
		questions        []string
		answers          []string
		expectedContains []string
		unexpectedParts  []string
	}{
		{
			name:             "normal paired Q&A",
			questions:        []string{"What is Go?", "Explain concurrency"},
			answers:          []string{"A programming language", "Running multiple tasks"},
			expectedContains: []string{"Q1:", "A1:", "Q2:", "A2:", "What is Go?", "A programming language"},
		},
		{
			name:             "empty arrays",
			questions:        []string{},
			answers:          []string{},
			expectedContains: []string{"Interview Questions and Candidate Answers:"},
		},
		{
			name:             "more questions than answers",
			questions:        []string{"Q1", "Q2", "Q3"},
			answers:          []string{"A1", "A2"},
			expectedContains: []string{"Q1:", "A1:", "Q2:", "A2:"},
			unexpectedParts:  []string{"Q3:"},
		},
		{
			name:             "more answers than questions",
			questions:        []string{"Q1"},
			answers:          []string{"A1", "A2"},
			expectedContains: []string{"Q1:", "A1:"},
			unexpectedParts:  []string{"Q2:", "A2:"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := FormatAnswersForEvaluation(tc.questions, tc.answers)

			for _, expected := range tc.expectedContains {
				if !strings.Contains(result, expected) {
					t.Errorf("Expected result to contain '%s', got: %s", expected, result)
				}
			}

			for _, unexpected := range tc.unexpectedParts {
				if strings.Contains(result, unexpected) {
					t.Errorf("Expected result NOT to contain '%s', got: %s", unexpected, result)
				}
			}
		})
	}
}

// TestParseQuestionResponse tests question parsing from AI response
func TestParseQuestionResponse(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedCount int
		checkFirst    func(t *testing.T, q InterviewQuestion)
	}{
		{
			name: "single question full format",
			input: `Question: What is a goroutine?
Category: technical
Difficulty: medium
Expected Time: 5`,
			expectedCount: 1,
			checkFirst: func(t *testing.T, q InterviewQuestion) {
				if q.Question != "What is a goroutine?" {
					t.Errorf("Expected question 'What is a goroutine?', got '%s'", q.Question)
				}
				if q.Category != "technical" {
					t.Errorf("Expected category 'technical', got '%s'", q.Category)
				}
				if q.Difficulty != "medium" {
					t.Errorf("Expected difficulty 'medium', got '%s'", q.Difficulty)
				}
			},
		},
		{
			name: "multiple questions",
			input: `Question: First question?
Category: behavioral
Difficulty: easy
Expected Time: 3

Question: Second question?
Category: technical
Difficulty: hard
Expected Time: 10`,
			expectedCount: 2,
		},
		{
			name:          "empty content",
			input:         "",
			expectedCount: 0,
		},
		{
			name: "incomplete question (no Expected Time)",
			input: `Question: Incomplete question
Category: technical
Difficulty: easy`,
			expectedCount: 0,
		},
		{
			name: "question with extra whitespace",
			input: `  Question:   Trimmed question?
  Category:   technical
  Difficulty:   hard
Expected Time: 5`,
			expectedCount: 1,
			checkFirst: func(t *testing.T, q InterviewQuestion) {
				if q.Question != "Trimmed question?" {
					t.Errorf("Expected trimmed question, got '%s'", q.Question)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			questions := ParseQuestionResponse(tc.input)

			if len(questions) != tc.expectedCount {
				t.Errorf("Expected %d questions, got %d", tc.expectedCount, len(questions))
			}

			if tc.checkFirst != nil && len(questions) > 0 {
				tc.checkFirst(t, questions[0])
			}
		})
	}
}

// TestParseEvaluationResponse tests evaluation parsing from AI response
func TestParseEvaluationResponse(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		checkEvaluation func(t *testing.T, e *EvaluationResponse)
	}{
		{
			name: "full evaluation format with all sections",
			input: `Overall Score: 0.85
Category Scores:
- Technical Skills: 0.9
- Communication: 0.8

Feedback: The candidate demonstrated strong technical skills and clear communication.

Strengths:
- Excellent problem-solving
- Clear explanations

Areas for Improvement:
- Could improve time management
- Needs more depth in answers

Recommendations:
- Practice system design questions
- Review data structures`,
			checkEvaluation: func(t *testing.T, e *EvaluationResponse) {
				if e.Feedback == "" {
					t.Error("Expected feedback to be parsed")
				}
				if !strings.Contains(e.Feedback, "strong technical skills") {
					t.Error("Expected feedback to contain 'strong technical skills'")
				}

				// Verify strengths parsed correctly
				if len(e.Strengths) != 2 {
					t.Errorf("Expected 2 strengths, got %d", len(e.Strengths))
				}
				if len(e.Strengths) > 0 && e.Strengths[0] != "Excellent problem-solving" {
					t.Errorf("Expected first strength 'Excellent problem-solving', got '%s'", e.Strengths[0])
				}

				// Verify weaknesses parsed correctly
				if len(e.Weaknesses) != 2 {
					t.Errorf("Expected 2 weaknesses, got %d", len(e.Weaknesses))
				}
				if len(e.Weaknesses) > 0 && e.Weaknesses[0] != "Could improve time management" {
					t.Errorf("Expected first weakness 'Could improve time management', got '%s'", e.Weaknesses[0])
				}

				// Verify recommendations parsed correctly
				if len(e.Recommendations) != 2 {
					t.Errorf("Expected 2 recommendations, got %d", len(e.Recommendations))
				}
				if len(e.Recommendations) > 0 && e.Recommendations[0] != "Practice system design questions" {
					t.Errorf("Expected first recommendation 'Practice system design questions', got '%s'", e.Recommendations[0])
				}
			},
		},
		{
			name:  "feedback only",
			input: `Feedback: Basic feedback text here.`,
			checkEvaluation: func(t *testing.T, e *EvaluationResponse) {
				if !strings.Contains(e.Feedback, "Basic feedback") {
					t.Error("Expected feedback to contain 'Basic feedback'")
				}
			},
		},
		{
			name:  "empty content returns defaults",
			input: "",
			checkEvaluation: func(t *testing.T, e *EvaluationResponse) {
				if e.OverallScore != 0.7 {
					t.Errorf("Expected default overall score 0.7, got %f", e.OverallScore)
				}
				if e.CategoryScores["technical"] != 0.7 {
					t.Error("Expected default technical score 0.7")
				}
			},
		},
		{
			name: "multiline feedback",
			input: `Feedback: First line of feedback.
Second line continues the feedback.
Third line as well.

Strengths:
- Good work`,
			checkEvaluation: func(t *testing.T, e *EvaluationResponse) {
				// Feedback lines should be joined with spaces
				if !strings.Contains(e.Feedback, "First line") {
					t.Error("Expected feedback to contain 'First line'")
				}
				if !strings.Contains(e.Feedback, "Second line") {
					t.Error("Expected multiline feedback to be joined")
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			evaluation := ParseEvaluationResponse(tc.input)

			if evaluation == nil {
				t.Fatal("Expected evaluation to not be nil")
			}

			// Verify category scores map is initialized
			if evaluation.CategoryScores == nil {
				t.Error("Expected CategoryScores map to be initialized")
			}

			if tc.checkEvaluation != nil {
				tc.checkEvaluation(t, evaluation)
			}
		})
	}
}

// TestParseEvaluationResponse_SectionSeparation tests correct categorization of items
func TestParseEvaluationResponse_SectionSeparation(t *testing.T) {
	input := `Feedback: Good interview performance.

Strengths:
- Strong technical knowledge
- Good communication

Areas for Improvement:
- Time management
- Code organization

Recommendations:
- Practice more algorithms
- Review best practices`

	evaluation := ParseEvaluationResponse(input)

	// Verify strengths
	if len(evaluation.Strengths) != 2 {
		t.Errorf("Expected 2 strengths, got %d: %v", len(evaluation.Strengths), evaluation.Strengths)
	}
	expectedStrengths := []string{"Strong technical knowledge", "Good communication"}
	for i, expected := range expectedStrengths {
		if i >= len(evaluation.Strengths) || evaluation.Strengths[i] != expected {
			t.Errorf("Strength %d: expected '%s', got '%v'", i, expected, evaluation.Strengths)
		}
	}

	// Verify weaknesses
	if len(evaluation.Weaknesses) != 2 {
		t.Errorf("Expected 2 weaknesses, got %d: %v", len(evaluation.Weaknesses), evaluation.Weaknesses)
	}
	expectedWeaknesses := []string{"Time management", "Code organization"}
	for i, expected := range expectedWeaknesses {
		if i >= len(evaluation.Weaknesses) || evaluation.Weaknesses[i] != expected {
			t.Errorf("Weakness %d: expected '%s', got '%v'", i, expected, evaluation.Weaknesses)
		}
	}

	// Verify recommendations
	if len(evaluation.Recommendations) != 2 {
		t.Errorf("Expected 2 recommendations, got %d: %v", len(evaluation.Recommendations), evaluation.Recommendations)
	}
	expectedRecommendations := []string{"Practice more algorithms", "Review best practices"}
	for i, expected := range expectedRecommendations {
		if i >= len(evaluation.Recommendations) || evaluation.Recommendations[i] != expected {
			t.Errorf("Recommendation %d: expected '%s', got '%v'", i, expected, evaluation.Recommendations)
		}
	}
}

// TestMakeRequest_MarshalError tests handling of unmarshalable payloads
func TestMakeRequest_MarshalError(t *testing.T) {
	config := &AIConfig{}
	bp := NewBaseProvider(config, "https://api.example.com", 10*time.Second)

	adapter := &mockAdapter{baseURL: "https://api.example.com"}

	// Create a payload that can't be marshaled (channel type)
	type unmarshalable struct {
		Ch chan int `json:"ch"`
	}
	payload := unmarshalable{Ch: make(chan int)}

	_, err := bp.MakeRequest(context.Background(), adapter, "/test", payload)

	if err == nil {
		t.Error("Expected marshal error, got nil")
	}
	if !strings.Contains(err.Error(), "marshal") {
		t.Errorf("Expected error to mention 'marshal', got: %v", err)
	}
}

// TestParseQuestionResponse_EdgeCases tests edge cases in question parsing
func TestParseQuestionResponse_EdgeCases(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedCount int
	}{
		{
			name:          "only whitespace",
			input:         "   \n\n   \t\t   ",
			expectedCount: 0,
		},
		{
			name: "partial question at end without Expected Time",
			input: `Question: Complete question?
Category: technical
Difficulty: easy
Expected Time: 5

Question: Incomplete question
Category: technical`,
			expectedCount: 1,
		},
		{
			name:          "random text without structure",
			input:         "This is just some random text without any question format.",
			expectedCount: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			questions := ParseQuestionResponse(tc.input)

			if len(questions) != tc.expectedCount {
				t.Errorf("Expected %d questions, got %d", tc.expectedCount, len(questions))
			}
		})
	}
}

// TestMakeRequest_JSONPayload verifies JSON payload is sent correctly
func TestMakeRequest_JSONPayload(t *testing.T) {
	var receivedPayload map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedPayload)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result": "success"}`))
	}))
	defer server.Close()

	config := &AIConfig{}
	bp := NewBaseProvider(config, server.URL, 10*time.Second)

	adapter := &mockAdapter{baseURL: server.URL}

	payload := map[string]interface{}{
		"model":       "test-model",
		"temperature": 0.7,
		"messages":    []string{"hello", "world"},
	}

	_, err := bp.MakeRequest(context.Background(), adapter, "/test", payload)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if receivedPayload["model"] != "test-model" {
		t.Errorf("Expected model 'test-model', got '%v'", receivedPayload["model"])
	}
	if receivedPayload["temperature"] != 0.7 {
		t.Errorf("Expected temperature 0.7, got '%v'", receivedPayload["temperature"])
	}
}

// TestParseEvaluationResponse_DefaultScores verifies all default category scores
func TestParseEvaluationResponse_DefaultScores(t *testing.T) {
	evaluation := ParseEvaluationResponse("")

	expectedScores := map[string]float64{
		"technical":       0.7,
		"communication":   0.8,
		"problem_solving": 0.6,
		"experience":      0.7,
	}

	for category, expectedScore := range expectedScores {
		if score, ok := evaluation.CategoryScores[category]; !ok {
			t.Errorf("Expected category '%s' to exist in scores", category)
		} else if score != expectedScore {
			t.Errorf("Expected %s score %.1f, got %.1f", category, expectedScore, score)
		}
	}
}

// TestBuildQuestionGenerationPrompt_AllFieldsUsed verifies prompt uses all request fields
func TestBuildQuestionGenerationPrompt_AllFieldsUsed(t *testing.T) {
	req := &QuestionGenerationRequest{
		ExperienceLevel: "junior",
		InterviewType:   "behavioral",
		Difficulty:      "easy",
		JobDescription:  "Entry-level position",
		ResumeContent:   "Fresh graduate",
		NumQuestions:    3,
	}

	prompt := BuildQuestionGenerationPrompt(req)

	// Check experience level appears multiple times (as it should per the template)
	count := strings.Count(prompt, "junior")
	if count < 2 {
		t.Errorf("Expected experience level to appear multiple times, found %d", count)
	}

	// Check interview type appears multiple times
	count = strings.Count(prompt, "behavioral")
	if count < 2 {
		t.Errorf("Expected interview type to appear multiple times, found %d", count)
	}

	// Check difficulty appears multiple times
	count = strings.Count(prompt, "easy")
	if count < 2 {
		t.Errorf("Expected difficulty to appear multiple times, found %d", count)
	}
}

// TestBuildEvaluationPrompt_JobDescIncluded verifies job description is in prompt
func TestBuildEvaluationPrompt_JobDescIncluded(t *testing.T) {
	req := &EvaluationRequest{
		JobDesc:     "Senior Software Engineer at Tech Company",
		Criteria:    []string{"coding", "design"},
		DetailLevel: "comprehensive",
	}

	prompt := BuildEvaluationPrompt(req)

	if !strings.Contains(prompt, "Senior Software Engineer at Tech Company") {
		t.Error("Expected job description to be in prompt")
	}
	if !strings.Contains(prompt, "comprehensive") {
		t.Error("Expected detail level to be in prompt")
	}
}

// TestFormatAnswersForEvaluation_Numbering verifies Q&A numbering is correct
func TestFormatAnswersForEvaluation_Numbering(t *testing.T) {
	questions := []string{"Q1", "Q2", "Q3"}
	answers := []string{"A1", "A2", "A3"}

	result := FormatAnswersForEvaluation(questions, answers)

	// Verify numbering format
	if !strings.Contains(result, "Q1: Q1") {
		t.Error("Expected 'Q1: Q1' format")
	}
	if !strings.Contains(result, "A1: A1") {
		t.Error("Expected 'A1: A1' format")
	}
	if !strings.Contains(result, "Q3: Q3") {
		t.Error("Expected 'Q3: Q3' format")
	}
	if !strings.Contains(result, "A3: A3") {
		t.Error("Expected 'A3: A3' format")
	}
}

// TestMakeRequest_InvalidURL tests handling of invalid URLs
func TestMakeRequest_InvalidURL(t *testing.T) {
	config := &AIConfig{}
	bp := NewBaseProvider(config, "not-a-valid-url", 10*time.Second)

	adapter := &mockAdapter{baseURL: "not-a-valid-url"}

	_, err := bp.MakeRequest(context.Background(), adapter, "/test", map[string]string{})

	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

// Benchmark tests
func BenchmarkParseQuestionResponse(b *testing.B) {
	input := `Question: What is a goroutine?
Category: technical
Difficulty: medium
Expected Time: 5

Question: Explain channels in Go
Category: technical
Difficulty: hard
Expected Time: 10`

	for i := 0; i < b.N; i++ {
		ParseQuestionResponse(input)
	}
}

func BenchmarkBuildQuestionGenerationPrompt(b *testing.B) {
	req := &QuestionGenerationRequest{
		ExperienceLevel: "senior",
		InterviewType:   "technical",
		Difficulty:      "hard",
		JobDescription:  "Build scalable distributed systems",
		ResumeContent:   "10 years experience",
		NumQuestions:    5,
	}

	for i := 0; i < b.N; i++ {
		BuildQuestionGenerationPrompt(req)
	}
}

func BenchmarkFormatAnswersForEvaluation(b *testing.B) {
	questions := []string{"Q1", "Q2", "Q3", "Q4", "Q5"}
	answers := []string{"A1", "A2", "A3", "A4", "A5"}

	for i := 0; i < b.N; i++ {
		FormatAnswersForEvaluation(questions, answers)
	}
}

