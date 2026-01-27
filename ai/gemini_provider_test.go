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

// TestNewGeminiProvider verifies provider initialization
func TestNewGeminiProvider(t *testing.T) {
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
			expectedBaseURL: "https://generativelanguage.googleapis.com/v1beta",
		},
		{
			name:            "custom base URL",
			apiKey:          "test-key",
			customBaseURL:   "https://custom.gemini.api/v1",
			expectedBaseURL: "https://custom.gemini.api/v1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &AIConfig{
				GeminiBaseURL:  tc.customBaseURL,
				RequestTimeout: 30 * time.Second,
			}

			provider := NewGeminiProvider(tc.apiKey, config)

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

// TestGeminiProvider_SetAuth verifies no auth header is set (Gemini uses URL param)
func TestGeminiProvider_SetAuth(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewGeminiProvider("test-api-key", config)

	req, _ := http.NewRequest("POST", "https://api.example.com/test", nil)
	provider.SetAuth(req)

	// Gemini should NOT set Authorization header (uses URL param instead)
	authHeader := req.Header.Get("Authorization")
	if authHeader != "" {
		t.Errorf("Expected no Authorization header for Gemini, got '%s'", authHeader)
	}
}

// TestGeminiProvider_GetEndpointURL verifies URL construction with API key
func TestGeminiProvider_GetEndpointURL(t *testing.T) {
	testCases := []struct {
		name        string
		baseURL     string
		apiKey      string
		endpoint    string
		expectedURL string
	}{
		{
			name:        "default endpoint",
			baseURL:     "https://generativelanguage.googleapis.com/v1beta",
			apiKey:      "my-api-key",
			endpoint:    "/models/gemini-pro:generateContent",
			expectedURL: "https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent?key=my-api-key",
		},
		{
			name:        "custom base URL",
			baseURL:     "https://custom.api.com",
			apiKey:      "custom-key",
			endpoint:    "/test",
			expectedURL: "https://custom.api.com/test?key=custom-key",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &AIConfig{
				GeminiBaseURL:  tc.baseURL,
				RequestTimeout: 30 * time.Second,
			}
			provider := NewGeminiProvider(tc.apiKey, config)

			url := provider.GetEndpointURL(tc.endpoint)

			if url != tc.expectedURL {
				t.Errorf("Expected URL '%s', got '%s'", tc.expectedURL, url)
			}
		})
	}
}

// TestGeminiProvider_convertMessages verifies message conversion
func TestGeminiProvider_convertMessages(t *testing.T) {
	testCases := []struct {
		name          string
		input         []Message
		expectedRoles []string
		checkContent  func(t *testing.T, contents []geminiContent)
	}{
		{
			name: "user message only",
			input: []Message{
				{Role: "user", Content: "Hello"},
			},
			expectedRoles: []string{"user"},
			checkContent: func(t *testing.T, contents []geminiContent) {
				if contents[0].Parts[0].Text != "Hello" {
					t.Errorf("Expected 'Hello', got '%s'", contents[0].Parts[0].Text)
				}
			},
		},
		{
			name: "assistant becomes model",
			input: []Message{
				{Role: "user", Content: "Hi"},
				{Role: "assistant", Content: "Hello!"},
			},
			expectedRoles: []string{"user", "model"},
			checkContent: func(t *testing.T, contents []geminiContent) {
				if contents[1].Role != "model" {
					t.Errorf("Expected 'model' role, got '%s'", contents[1].Role)
				}
			},
		},
		{
			name: "system message prepended to first user message",
			input: []Message{
				{Role: "system", Content: "Be helpful"},
				{Role: "user", Content: "Hello"},
			},
			expectedRoles: []string{"user"},
			checkContent: func(t *testing.T, contents []geminiContent) {
				// System message should be prepended to user message
				if len(contents) != 1 {
					t.Errorf("Expected 1 content (system merged), got %d", len(contents))
					return
				}
				if !strings.Contains(contents[0].Parts[0].Text, "Be helpful") {
					t.Error("Expected system message to be prepended")
				}
				if !strings.Contains(contents[0].Parts[0].Text, "Hello") {
					t.Error("Expected user message to be present")
				}
			},
		},
		{
			name: "multiple system messages",
			input: []Message{
				{Role: "system", Content: "First instruction"},
				{Role: "system", Content: "Second instruction"},
				{Role: "user", Content: "Question"},
			},
			expectedRoles: []string{"user"},
			checkContent: func(t *testing.T, contents []geminiContent) {
				text := contents[0].Parts[0].Text
				if !strings.Contains(text, "First instruction") {
					t.Error("Expected first system message")
				}
				if !strings.Contains(text, "Second instruction") {
					t.Error("Expected second system message")
				}
				if !strings.Contains(text, "Question") {
					t.Error("Expected user question")
				}
			},
		},
		{
			name:          "empty messages",
			input:         []Message{},
			expectedRoles: []string{},
		},
	}

	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewGeminiProvider("test-key", config)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			converted := provider.convertMessages(tc.input)

			if len(converted) != len(tc.expectedRoles) {
				t.Errorf("Expected %d contents, got %d", len(tc.expectedRoles), len(converted))
				return
			}

			for i, content := range converted {
				if content.Role != tc.expectedRoles[i] {
					t.Errorf("Content %d: expected role '%s', got '%s'", i, tc.expectedRoles[i], content.Role)
				}
			}

			if tc.checkContent != nil {
				tc.checkContent(t, converted)
			}
		})
	}
}

// TestGeminiProvider_GenerateResponse tests full response generation
func TestGeminiProvider_GenerateResponse(t *testing.T) {
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
				"candidates": [{
					"content": {
						"parts": [{"text": "Hello from Gemini!"}],
						"role": "model"
					},
					"finishReason": "STOP",
					"index": 0
				}],
				"usageMetadata": {
					"promptTokenCount": 10,
					"candidatesTokenCount": 20,
					"totalTokenCount": 30
				}
			}`,
			serverStatus: http.StatusOK,
			expectError:  false,
			checkResponse: func(t *testing.T, resp *ChatResponse) {
				if resp.Content != "Hello from Gemini!" {
					t.Errorf("Expected content 'Hello from Gemini!', got '%s'", resp.Content)
				}
				if resp.TokensUsed.TotalTokens != 30 {
					t.Errorf("Expected 30 total tokens, got %d", resp.TokensUsed.TotalTokens)
				}
				if resp.Provider != ProviderGemini {
					t.Errorf("Expected provider '%s', got '%s'", ProviderGemini, resp.Provider)
				}
			},
		},
		{
			name: "API error response",
			serverResponse: `{
				"error": {
					"code": 400,
					"message": "Invalid API key",
					"status": "INVALID_ARGUMENT"
				}
			}`,
			serverStatus:  http.StatusOK,
			expectError:   true,
			errorContains: "Invalid API key",
		},
		{
			name: "no candidates in response",
			serverResponse: `{
				"candidates": [],
				"usageMetadata": {"totalTokenCount": 10}
			}`,
			serverStatus:  http.StatusOK,
			expectError:   true,
			errorContains: "no candidates",
		},
		{
			name: "no content parts",
			serverResponse: `{
				"candidates": [{
					"content": {"parts": [], "role": "model"},
					"finishReason": "STOP"
				}]
			}`,
			serverStatus:  http.StatusOK,
			expectError:   true,
			errorContains: "no content parts",
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
			serverResponse: `{"error": {"message": "internal error"}}`,
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
				// Verify API key in URL
				if !strings.Contains(r.URL.RawQuery, "key=") {
					t.Error("Expected API key in URL query")
				}

				w.WriteHeader(tc.serverStatus)
				w.Write([]byte(tc.serverResponse))
			}))
			defer server.Close()

			config := &AIConfig{
				GeminiBaseURL:  server.URL,
				RequestTimeout: 10 * time.Second,
				DefaultModel:   "gemini-1.5-flash",
			}
			provider := NewGeminiProvider("test-key", config)

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

// TestGeminiProvider_GenerateResponse_RequestFormat verifies request body format
func TestGeminiProvider_GenerateResponse_RequestFormat(t *testing.T) {
	var receivedRequest map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&receivedRequest)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"candidates": [{
				"content": {"parts": [{"text": "ok"}], "role": "model"},
				"finishReason": "STOP"
			}],
			"usageMetadata": {"totalTokenCount": 1}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		GeminiBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
		DefaultModel:   "gemini-1.5-flash",
	}
	provider := NewGeminiProvider("test-key", config)

	req := &ChatRequest{
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens:   500,
		Temperature: 0.5,
		TopP:        0.9,
	}

	_, err := provider.GenerateResponse(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify request structure
	genConfig := receivedRequest["generationConfig"].(map[string]interface{})
	if genConfig["maxOutputTokens"] != float64(500) {
		t.Errorf("Expected maxOutputTokens 500, got '%v'", genConfig["maxOutputTokens"])
	}
	if genConfig["temperature"] != 0.5 {
		t.Errorf("Expected temperature 0.5, got '%v'", genConfig["temperature"])
	}

	// Verify safety settings are included
	if receivedRequest["safetySettings"] == nil {
		t.Error("Expected safetySettings to be present")
	}
}

// TestGeminiProvider_GetProviderName verifies provider name
func TestGeminiProvider_GetProviderName(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewGeminiProvider("test-key", config)

	name := provider.GetProviderName()

	if name != ProviderGemini {
		t.Errorf("Expected '%s', got '%s'", ProviderGemini, name)
	}
}

// TestGeminiProvider_GetSupportedModels verifies supported models list
func TestGeminiProvider_GetSupportedModels(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewGeminiProvider("test-key", config)

	models := provider.GetSupportedModels()

	if len(models) == 0 {
		t.Error("Expected at least one supported model")
	}

	// Verify some expected models are in the list
	expectedModels := []string{"gemini-1.5-pro", "gemini-1.5-flash"}
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

// TestGeminiProvider_GenerateInterviewQuestions tests question generation
func TestGeminiProvider_GenerateInterviewQuestions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"candidates": [{
				"content": {
					"parts": [{"text": "Question: What is your experience?\nCategory: behavioral\nDifficulty: easy\nExpected Time: 3"}],
					"role": "model"
				},
				"finishReason": "STOP"
			}],
			"usageMetadata": {"totalTokenCount": 100}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		GeminiBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
		DefaultModel:   "gemini-1.5-flash",
	}
	provider := NewGeminiProvider("test-key", config)

	req := &QuestionGenerationRequest{
		JobDescription:  "Frontend Developer",
		ResumeContent:   "3 years React experience",
		ExperienceLevel: "mid",
		InterviewType:   "technical",
		NumQuestions:    5,
		Difficulty:      "medium",
	}

	resp, err := provider.GenerateInterviewQuestions(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Provider != ProviderGemini {
		t.Errorf("Expected provider '%s', got '%s'", ProviderGemini, resp.Provider)
	}
}

// TestGeminiProvider_EvaluateAnswers tests answer evaluation
func TestGeminiProvider_EvaluateAnswers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"candidates": [{
				"content": {
					"parts": [{"text": "Overall Score: 0.75\n\nFeedback: Solid performance.\n\nStrengths:\n- Good knowledge"}],
					"role": "model"
				},
				"finishReason": "STOP"
			}],
			"usageMetadata": {"totalTokenCount": 150}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		GeminiBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
		DefaultModel:   "gemini-1.5-flash",
	}
	provider := NewGeminiProvider("test-key", config)

	req := &EvaluationRequest{
		Questions:   []string{"Explain React hooks"},
		Answers:     []string{"Hooks allow state in functional components"},
		JobDesc:     "Frontend Developer",
		Criteria:    []string{"technical"},
		DetailLevel: "brief",
	}

	resp, err := provider.EvaluateAnswers(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.Provider != ProviderGemini {
		t.Errorf("Expected provider '%s', got '%s'", ProviderGemini, resp.Provider)
	}
}

// TestGeminiProvider_ValidateCredentials tests credential validation
func TestGeminiProvider_ValidateCredentials(t *testing.T) {
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
						"candidates": [{"content": {"parts": [{"text": "ok"}]}}],
						"usageMetadata": {"totalTokenCount": 1}
					}`))
				} else {
					w.Write([]byte(`{"error": {"message": "Invalid key"}}`))
				}
			}))
			defer server.Close()

			config := &AIConfig{
				GeminiBaseURL:  server.URL,
				RequestTimeout: 10 * time.Second,
			}
			provider := NewGeminiProvider("test-key", config)

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

// TestGeminiProvider_IsHealthy tests health check
func TestGeminiProvider_IsHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"candidates": [{"content": {"parts": [{"text": "ok"}]}}],
			"usageMetadata": {"totalTokenCount": 1}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		GeminiBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
	}
	provider := NewGeminiProvider("test-key", config)

	healthy := provider.IsHealthy(context.Background())

	if !healthy {
		t.Error("Expected provider to be healthy")
	}
}

// TestGeminiProvider_GetUsageStats tests usage stats retrieval
func TestGeminiProvider_GetUsageStats(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewGeminiProvider("test-key", config)

	stats, err := provider.GetUsageStats(context.Background())

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if stats == nil {
		t.Fatal("Expected stats, got nil")
	}
	if stats["provider"] != ProviderGemini {
		t.Errorf("Expected provider '%s', got '%v'", ProviderGemini, stats["provider"])
	}
}

// TestGeminiProvider_GenerateStreamResponse tests streaming placeholder
func TestGeminiProvider_GenerateStreamResponse(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewGeminiProvider("test-key", config)

	_, err := provider.GenerateStreamResponse(context.Background(), &ChatRequest{})

	if err == nil {
		t.Error("Expected error for unimplemented streaming")
	}
	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("Expected 'not yet implemented' error, got: %v", err)
	}
}

// TestGeminiProvider_ModelFallback tests model name fallback
func TestGeminiProvider_ModelFallback(t *testing.T) {
	var receivedEndpoint string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedEndpoint = r.URL.Path

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"candidates": [{"content": {"parts": [{"text": "ok"}]}, "finishReason": "STOP"}],
			"usageMetadata": {"totalTokenCount": 1}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		GeminiBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
		DefaultModel:   "custom-model",
	}
	provider := NewGeminiProvider("test-key", config)

	// Request without model should use "gemini-1.5-flash" as the Gemini default
	req := &ChatRequest{
		Messages: []Message{{Role: "user", Content: "test"}},
	}

	_, err := provider.GenerateResponse(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Gemini provider uses "gemini-1.5-flash" as its default
	if !strings.Contains(receivedEndpoint, "gemini-1.5-flash") {
		t.Errorf("Expected endpoint to contain 'gemini-1.5-flash', got '%s'", receivedEndpoint)
	}
}

// TestGeminiProvider_getDefaultSafetySettings tests safety settings
func TestGeminiProvider_getDefaultSafetySettings(t *testing.T) {
	config := &AIConfig{RequestTimeout: 30 * time.Second}
	provider := NewGeminiProvider("test-key", config)

	settings := provider.getDefaultSafetySettings()

	if len(settings) == 0 {
		t.Error("Expected safety settings to be present")
	}

	// Verify expected categories
	expectedCategories := []string{
		"HARM_CATEGORY_HARASSMENT",
		"HARM_CATEGORY_HATE_SPEECH",
		"HARM_CATEGORY_SEXUALLY_EXPLICIT",
		"HARM_CATEGORY_DANGEROUS_CONTENT",
	}

	for _, expected := range expectedCategories {
		found := false
		for _, setting := range settings {
			if setting.Category == expected {
				found = true
				if setting.Threshold != "BLOCK_MEDIUM_AND_ABOVE" {
					t.Errorf("Expected threshold 'BLOCK_MEDIUM_AND_ABOVE' for %s", expected)
				}
				break
			}
		}
		if !found {
			t.Errorf("Expected category '%s' in safety settings", expected)
		}
	}
}

// TestGeminiProvider_ResponseWithoutUsageMetadata tests handling missing usage data
func TestGeminiProvider_ResponseWithoutUsageMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Response without usageMetadata
		w.Write([]byte(`{
			"candidates": [{
				"content": {"parts": [{"text": "response"}], "role": "model"},
				"finishReason": "STOP"
			}]
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		GeminiBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
	}
	provider := NewGeminiProvider("test-key", config)

	req := &ChatRequest{
		Messages: []Message{{Role: "user", Content: "test"}},
	}

	resp, err := provider.GenerateResponse(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should succeed with zero token counts
	if resp.TokensUsed.TotalTokens != 0 {
		t.Errorf("Expected 0 tokens when no usage metadata, got %d", resp.TokensUsed.TotalTokens)
	}
}

// TestGeminiProvider_APIKeyInURL verifies API key is passed in URL
func TestGeminiProvider_APIKeyInURL(t *testing.T) {
	var receivedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedQuery = r.URL.RawQuery

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"candidates": [{"content": {"parts": [{"text": "ok"}]}}],
			"usageMetadata": {"totalTokenCount": 1}
		}`))
	}))
	defer server.Close()

	config := &AIConfig{
		GeminiBaseURL:  server.URL,
		RequestTimeout: 10 * time.Second,
	}
	provider := NewGeminiProvider("my-secret-api-key", config)

	req := &ChatRequest{
		Messages: []Message{{Role: "user", Content: "test"}},
	}

	_, err := provider.GenerateResponse(context.Background(), req)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(receivedQuery, "key=my-secret-api-key") {
		t.Errorf("Expected API key in query string, got '%s'", receivedQuery)
	}
}
