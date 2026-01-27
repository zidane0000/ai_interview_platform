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

// TestNewOpenAIProvider verifies provider initialization
func TestNewOpenAIProvider(t *testing.T) {
	testCases := []struct {
		name            string
		apiKey          string
		customBaseURL   string
		expectedBaseURL string
	}{
		{
			name:            "default base URL",
			apiKey:          "test-key",
			customBaseURL:   "",
			expectedBaseURL: "https://api.openai.com/v1",
		},
		{
			name:            "custom base URL",
			apiKey:          "test-key",
			customBaseURL:   "https://api.together.ai/v1",
			expectedBaseURL: "https://api.together.ai/v1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &AIConfig{
				OpenAIBaseURL:  tc.customBaseURL,
				RequestTimeout: 30 * time.Second,
			}

			provider := NewOpenAIProvider(tc.apiKey, config)

			if provider == nil {
				t.Fatal("Expected provider to be created")
			}
			if provider.baseURL != tc.expectedBaseURL {
				t.Errorf("Expected baseURL '%s', got '%s'", tc.expectedBaseURL, provider.baseURL)
			}
			if provider.apiKey != tc.apiKey {
				t.Errorf("Expected apiKey '%s', got '%s'", tc.apiKey, provider.apiKey)
			}
		})
	}
}

// TestOpenAIProvider_SetAuth verifies Authorization header is set correctly
func TestOpenAIProvider_SetAuth(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewOpenAIProvider("sk-test-key-12345", config)

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", nil)
	provider.SetAuth(req)

	authHeader := req.Header.Get("Authorization")
	expected := "Bearer sk-test-key-12345"

	if authHeader != expected {
		t.Errorf("Expected Authorization header '%s', got '%s'", expected, authHeader)
	}
}

// TestOpenAIProvider_GetEndpointURL verifies URL construction
func TestOpenAIProvider_GetEndpointURL(t *testing.T) {
	testCases := []struct {
		name        string
		baseURL     string
		endpoint    string
		expectedURL string
	}{
		{
			name:        "chat completions endpoint",
			baseURL:     "https://api.openai.com/v1",
			endpoint:    "/chat/completions",
			expectedURL: "https://api.openai.com/v1/chat/completions",
		},
		{
			name:        "custom base URL",
			baseURL:     "https://custom.api.com",
			endpoint:    "/chat/completions",
			expectedURL: "https://custom.api.com/chat/completions",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &AIConfig{
				OpenAIBaseURL:  tc.baseURL,
				RequestTimeout: 30 * time.Second,
			}
			provider := NewOpenAIProvider("test-key", config)

			url := provider.GetEndpointURL(tc.endpoint)

			if url != tc.expectedURL {
				t.Errorf("Expected URL '%s', got '%s'", tc.expectedURL, url)
			}
		})
	}
}

// TestOpenAIProvider_convertMessages verifies message conversion
func TestOpenAIProvider_convertMessages(t *testing.T) {
	testCases := []struct {
		name           string
		input          []Message
		expectedRoles  []string
		expectedTexts  []string
	}{
		{
			name: "single user message",
			input: []Message{
				{Role: "user", Content: "Hello"},
			},
			expectedRoles: []string{"user"},
			expectedTexts: []string{"Hello"},
		},
		{
			name: "multiple messages",
			input: []Message{
				{Role: "system", Content: "You are helpful"},
				{Role: "user", Content: "Hi"},
				{Role: "assistant", Content: "Hello!"},
			},
			expectedRoles: []string{"system", "user", "assistant"},
			expectedTexts: []string{"You are helpful", "Hi", "Hello!"},
		},
		{
			name:           "empty messages",
			input:          []Message{},
			expectedRoles:  []string{},
			expectedTexts:  []string{},
		},
	}

	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewOpenAIProvider("test-key", config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			converted := provider.convertMessages(tc.input)

			if len(converted) != len(tc.expectedRoles) {
				t.Errorf("Expected %d messages, got %d", len(tc.expectedRoles), len(converted))
				return
			}

			for i, msg := range converted {
				if msg.Role != tc.expectedRoles[i] {
					t.Errorf("Message %d: expected role '%s', got '%s'", i, tc.expectedRoles[i], msg.Role)
				}
				if msg.Content != tc.expectedTexts[i] {
					t.Errorf("Message %d: expected content '%s', got '%s'", i, tc.expectedTexts[i], msg.Content)
				}
			}
		})
	}
}

// TestOpenAIProvider_GenerateResponse tests the full response generation
func TestOpenAIProvider_GenerateResponse(t *testing.T) {
	testCases := []struct {
		name           string
		serverResponse string
		serverStatus   int
		expectError    bool
		errorContains  string
		checkResponse  func(t *testing.T, resp *ChatResponse)
	}{
		{
			name: "successful response",
			serverResponse: `{
				"id": "chatcmpl-123",
				"object": "chat.completion",
				"created": 1677652288,
				"model": "gpt-4",
				"choices": [{
					"index": 0,
					"message": {"role": "assistant", "content": "Hello! How can I help?"},
					"finish_reason": "stop"
				}],
				"usage": {
					"prompt_tokens": 10,
					"completion_tokens": 20,
					"total_tokens": 30
				}
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
			checkResponse: func(t *testing.T, resp *ChatResponse) {
				if resp.Content != "Hello! How can I help?" {
					t.Errorf("Expected content 'Hello! How can I help?', got '%s'", resp.Content)
				}
				if resp.TokensUsed.TotalTokens != 30 {
					t.Errorf("Expected 30 total tokens, got %d", resp.TokensUsed.TotalTokens)
				}
				if resp.Provider != ProviderOpenAI {
					t.Errorf("Expected provider '%s', got '%s'", ProviderOpenAI, resp.Provider)
				}
				if resp.FinishReason != "stop" {
					t.Errorf("Expected finish_reason 'stop', got '%s'", resp.FinishReason)
				}
			},
		},
		{
			name: "API error response",
			serverResponse: `{
				"error": {
					"message": "Invalid API key",
					"type": "invalid_request_error",
					"code": "invalid_api_key"
				}
			}`,
			serverStatus:  http.StatusOK,
			expectError:   true,
			errorContains: "Invalid API key",
		},
		{
			name: "no choices in response",
			serverResponse: `{
				"id": "chatcmpl-123",
				"choices": [],
				"usage": {"prompt_tokens": 10, "completion_tokens": 0, "total_tokens": 10}
			}`,
			serverStatus:  http.StatusOK,
			expectError:   true,
			errorContains: "no choices",
		},
		{
			name:           "invalid JSON response",
			serverResponse: `{invalid json`,
			serverStatus:   http.StatusOK,
			expectError:    true,
			errorContains:  "parse",
		},
		{
			name:           "server error",
			serverResponse: `{"error": "internal error"}`,
			serverStatus:   http.StatusInternalServerError,
			expectError:    true,
			errorContains:  "status 500",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != "POST" {
					t.Errorf("Expected POST, got %s", r.Method)
				}
				if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
					t.Errorf("Expected /chat/completions endpoint, got %s", r.URL.Path)
				}
				if r.Header.Get("Authorization") == "" {
					t.Error("Expected Authorization header")
				}

				w.WriteHeader(tc.serverStatus)
				w.Write([]byte(tc.serverResponse))
			}))
			defer server.Close()

			config := &AIConfig{
				OpenAIBaseURL:  server.URL,
				RequestTimeout: 10 * time.Second,
				DefaultModel:   "gpt-4",
			}
			provider := NewOpenAIProvider("test-key", config)

			req := &ChatRequest{
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				MaxTokens:   100,
				Temperature: 0.7,
			}

			resp, err := provider.GenerateResponse(context.Background(), req)

			if tc.expectError {
				if err == nil {
					t.Error("Expected error, got nil")
				} else if tc.errorContains != "" && !strings.Contains(err.Error(), tc.errorContains) {
					t.Errorf("Expected error to contain '%s', got '%s'", tc.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if tc.checkResponse != nil {
					tc.checkResponse(t, resp)
				}
			}
		})
	}
}

// TestOpenAIProvider_GenerateResponse_RequestFormat verifies request body format
func TestOpenAIProvider_GenerateResponse_RequestFormat(t *testing.T) {
	var receivedRequest map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedRequest)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "test",
			"choices": [{"message": {"content": "ok"}, "finish_reason": "stop"}],
			"usage": {"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		OpenAIBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
		DefaultModel:   "gpt-4",
	}
	provider := NewOpenAIProvider("test-key", config)

	req := &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: "Be helpful"},
			{Role: "user", Content: "Hello"},
		},
		Model:       "gpt-4-turbo",
		MaxTokens:   500,
		Temperature: 0.5,
		TopP:        0.9,
	}

	_, err := provider.GenerateResponse(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify request structure
	if receivedRequest["model"] != "gpt-4-turbo" {
		t.Errorf("Expected model 'gpt-4-turbo', got '%v'", receivedRequest["model"])
	}
	if receivedRequest["max_tokens"] != float64(500) {
		t.Errorf("Expected max_tokens 500, got '%v'", receivedRequest["max_tokens"])
	}
	if receivedRequest["temperature"] != 0.5 {
		t.Errorf("Expected temperature 0.5, got '%v'", receivedRequest["temperature"])
	}

	messages := receivedRequest["messages"].([]interface{})
	if len(messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(messages))
	}
}

// TestOpenAIProvider_GetProviderName verifies provider name
func TestOpenAIProvider_GetProviderName(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewOpenAIProvider("test-key", config)

	name := provider.GetProviderName()

	if name != ProviderOpenAI {
		t.Errorf("Expected '%s', got '%s'", ProviderOpenAI, name)
	}
}

// TestOpenAIProvider_GetSupportedModels verifies supported models list
func TestOpenAIProvider_GetSupportedModels(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewOpenAIProvider("test-key", config)

	models := provider.GetSupportedModels()

	if len(models) == 0 {
		t.Error("Expected at least one supported model")
	}

	// Verify some expected models are in the list
	expectedModels := []string{"gpt-4", "gpt-3.5-turbo"}
	for _, expected := range expectedModels {
		found := false
		for _, model := range models {
			if model == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected model '%s' to be in supported list", expected)
		}
	}
}

// TestOpenAIProvider_GenerateInterviewQuestions tests question generation
func TestOpenAIProvider_GenerateInterviewQuestions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a response with properly formatted questions
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "test",
			"model": "gpt-4",
			"choices": [{
				"message": {
					"content": "Question: What is your experience with Go?\nCategory: technical\nDifficulty: medium\nExpected Time: 5"
				},
				"finish_reason": "stop"
			}],
			"usage": {"prompt_tokens": 100, "completion_tokens": 50, "total_tokens": 150}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		OpenAIBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
		DefaultModel:   "gpt-4",
	}
	provider := NewOpenAIProvider("test-key", config)

	req := &QuestionGenerationRequest{
		JobDescription:  "Backend Engineer",
		ResumeContent:   "5 years Go experience",
		ExperienceLevel: "senior",
		InterviewType:   "technical",
		NumQuestions:    3,
		Difficulty:      "medium",
	}

	resp, err := provider.GenerateInterviewQuestions(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Provider != ProviderOpenAI {
		t.Errorf("Expected provider '%s', got '%s'", ProviderOpenAI, resp.Provider)
	}
	if len(resp.Questions) == 0 {
		t.Error("Expected at least one question to be parsed")
	}
}

// TestOpenAIProvider_EvaluateAnswers tests answer evaluation
func TestOpenAIProvider_EvaluateAnswers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "test",
			"model": "gpt-4",
			"choices": [{
				"message": {
					"content": "Overall Score: 0.8\n\nFeedback: Good answers overall.\n\nStrengths:\n- Clear communication\n- Technical knowledge"
				},
				"finish_reason": "stop"
			}],
			"usage": {"prompt_tokens": 200, "completion_tokens": 100, "total_tokens": 300}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		OpenAIBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
		DefaultModel:   "gpt-4",
	}
	provider := NewOpenAIProvider("test-key", config)

	req := &EvaluationRequest{
		Questions:   []string{"What is Go?", "Explain concurrency"},
		Answers:     []string{"A language", "Running tasks"},
		JobDesc:     "Backend Engineer",
		Criteria:    []string{"technical", "communication"},
		DetailLevel: "detailed",
	}

	resp, err := provider.EvaluateAnswers(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Provider != ProviderOpenAI {
		t.Errorf("Expected provider '%s', got '%s'", ProviderOpenAI, resp.Provider)
	}
	if resp.Feedback == "" {
		t.Error("Expected feedback to be present")
	}
}

// TestOpenAIProvider_ValidateCredentials tests credential validation
func TestOpenAIProvider_ValidateCredentials(t *testing.T) {
	testCases := []struct {
		name         string
		serverStatus int
		expectError  bool
	}{
		{
			name:         "valid credentials",
			serverStatus: http.StatusOK,
			expectError:  false,
		},
		{
			name:         "invalid credentials",
			serverStatus: http.StatusUnauthorized,
			expectError:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.serverStatus)
				if tc.serverStatus == http.StatusOK {
					w.Write([]byte(`{
						"choices": [{"message": {"content": "ok"}}],
						"usage": {"total_tokens": 1}
					}`))
				} else {
					w.Write([]byte(`{"error": {"message": "Invalid key"}}`))
				}
			}))
			defer server.Close()

			config := &AIConfig{
				OpenAIBaseURL:  server.URL,
				RequestTimeout: 10 * time.Second,
			}
			provider := NewOpenAIProvider("test-key", config)

			err := provider.ValidateCredentials(context.Background())

			if tc.expectError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestOpenAIProvider_IsHealthy tests health check
func TestOpenAIProvider_IsHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [{"message": {"content": "ok"}}],
			"usage": {"total_tokens": 1}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		OpenAIBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
	}
	provider := NewOpenAIProvider("test-key", config)

	healthy := provider.IsHealthy(context.Background())

	if !healthy {
		t.Error("Expected provider to be healthy")
	}
}

// TestOpenAIProvider_GetUsageStats tests usage stats retrieval
func TestOpenAIProvider_GetUsageStats(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewOpenAIProvider("test-key", config)

	stats, err := provider.GetUsageStats(context.Background())

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}
	if stats["provider"] != ProviderOpenAI {
		t.Errorf("Expected provider '%s', got '%v'", ProviderOpenAI, stats["provider"])
	}
}

// TestOpenAIProvider_GenerateStreamResponse tests streaming placeholder
func TestOpenAIProvider_GenerateStreamResponse(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewOpenAIProvider("test-key", config)

	_, err := provider.GenerateStreamResponse(context.Background(), &ChatRequest{})

	if err == nil {
		t.Error("Expected error for unimplemented streaming")
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("Expected 'not yet implemented' error, got: %v", err)
	}
}

// TestOpenAIProvider_ModelFallback tests model name fallback
func TestOpenAIProvider_ModelFallback(t *testing.T) {
	var receivedModel string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		json.NewDecoder(r.Body).Decode(&req)
		receivedModel = req["model"].(string)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"choices": [{"message": {"content": "ok"}, "finish_reason": "stop"}],
			"usage": {"total_tokens": 1}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		OpenAIBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
		DefaultModel:   "default-model",
	}
	provider := NewOpenAIProvider("test-key", config)

	// Request without model should use config default
	req := &ChatRequest{
		Messages: []Message{{Role: "user", Content: "test"}},
	}

	_, err := provider.GenerateResponse(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if receivedModel != "default-model" {
		t.Errorf("Expected default model 'default-model', got '%s'", receivedModel)
	}
}
