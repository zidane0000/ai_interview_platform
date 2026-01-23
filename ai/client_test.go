package ai

import (
	"testing"
	"time"
)

// Test NewAIClient with various configurations
func TestNewAIClient(t *testing.T) {
	tests := []struct {
		name        string
		config      *AIConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid mock provider",
			config: &AIConfig{
				DefaultProvider:  ProviderMock,
				DefaultModel:     "mock-model",
				MaxRetries:       2,
				RequestTimeout:   60 * time.Second,
				DefaultMaxTokens: 1000,
				DefaultTemp:      0.7,
			},
			expectError: false,
		},
		{
			name: "valid OpenAI provider",
			config: &AIConfig{
				OpenAIAPIKey:     "sk-test-key",
				DefaultProvider:  ProviderOpenAI,
				DefaultModel:     "gpt-4",
				MaxRetries:       2,
				RequestTimeout:   60 * time.Second,
				DefaultMaxTokens: 1000,
				DefaultTemp:      0.7,
			},
			expectError: false,
		},
		{
			name: "valid Gemini provider",
			config: &AIConfig{
				GeminiAPIKey:     "test-gemini-key",
				DefaultProvider:  ProviderGemini,
				DefaultModel:     "gemini-pro",
				MaxRetries:       2,
				RequestTimeout:   60 * time.Second,
				DefaultMaxTokens: 1000,
				DefaultTemp:      0.7,
			},
			expectError: false,
		},
		{
			name: "OpenAI with custom base URL",
			config: &AIConfig{
				OpenAIAPIKey:     "sk-test-key",
				OpenAIBaseURL:    "https://api.custom.com/v1",
				DefaultProvider:  ProviderOpenAI,
				DefaultModel:     "gpt-4",
				MaxRetries:       2,
				RequestTimeout:   60 * time.Second,
				DefaultMaxTokens: 1000,
				DefaultTemp:      0.7,
			},
			expectError: false,
		},
		{
			name: "missing OpenAI key",
			config: &AIConfig{
				DefaultProvider:  ProviderOpenAI,
				DefaultModel:     "gpt-4",
				MaxRetries:       2,
				RequestTimeout:   60 * time.Second,
				DefaultMaxTokens: 1000,
				DefaultTemp:      0.7,
			},
			expectError: true,
			errorMsg:    "at least one AI provider API key must be configured",
		},
		{
			name: "missing Gemini key",
			config: &AIConfig{
				DefaultProvider:  ProviderGemini,
				DefaultModel:     "gemini-pro",
				MaxRetries:       2,
				RequestTimeout:   60 * time.Second,
				DefaultMaxTokens: 1000,
				DefaultTemp:      0.7,
			},
			expectError: true,
			errorMsg:    "at least one AI provider API key must be configured",
		},
		{
			name: "invalid provider",
			config: &AIConfig{
				OpenAIAPIKey:     "sk-test",  // Add key so validation reaches provider check
				DefaultProvider:  "invalid-provider",
				DefaultModel:     "model",
				MaxRetries:       2,
				RequestTimeout:   60 * time.Second,
				DefaultMaxTokens: 1000,
				DefaultTemp:      0.7,
			},
			expectError: true,
			errorMsg:    "invalid default provider",
		},
		{
			name: "negative max retries",
			config: &AIConfig{
				DefaultProvider:  ProviderMock,
				DefaultModel:     "mock",
				MaxRetries:       -1,
				RequestTimeout:   60 * time.Second,
				DefaultMaxTokens: 1000,
				DefaultTemp:      0.7,
			},
			expectError: true,
			errorMsg:    "max retries cannot be negative",
		},
		{
			name: "zero timeout",
			config: &AIConfig{
				DefaultProvider:  ProviderMock,
				DefaultModel:     "mock",
				MaxRetries:       2,
				RequestTimeout:   0,
				DefaultMaxTokens: 1000,
				DefaultTemp:      0.7,
			},
			expectError: true,
			errorMsg:    "request timeout must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewAIClient(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', but got no error", tt.errorMsg)
					return
				}
				if !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got: %s", tt.errorMsg, err.Error())
				}
				if client != nil {
					t.Error("Expected nil client on error")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
					return
				}
				if client == nil {
					t.Error("Expected client to be created")
					return
				}
				// Verify provider was created correctly
				if client.provider == nil {
					t.Error("Expected provider to be initialized")
				}
				if client.config == nil {
					t.Error("Expected config to be stored")
				}
				// Verify provider name matches config
				if client.GetCurrentProvider() != tt.config.DefaultProvider {
					t.Errorf("Provider name = %s, expected %s", client.GetCurrentProvider(), tt.config.DefaultProvider)
				}
			}
		})
	}
}

// Helper to create valid test config
func createTestConfig(provider string) *AIConfig {
	cfg := &AIConfig{
		DefaultProvider:  provider,
		DefaultModel:     "test-model",
		MaxRetries:       2,
		RequestTimeout:   60 * 1000000000, // 60 seconds in nanoseconds
		DefaultMaxTokens: 1000,
		DefaultTemp:      0.7,
	}

	if provider == ProviderOpenAI {
		cfg.OpenAIAPIKey = "sk-test-key-for-testing"
	} else if provider == ProviderGemini {
		cfg.GeminiAPIKey = "test-gemini-key-for-testing"
	}

	return cfg
}

// Test ShouldEndInterview logic
func TestShouldEndInterview(t *testing.T) {
	tests := []struct {
		name         string
		messageCount int
		expected     bool
	}{
		{"less than threshold", 5, false},
		{"at threshold", 8, true},
		{"above threshold", 10, true},
		{"zero messages", 0, false},
		{"one message", 1, false},
		{"seven messages", 7, false},
	}

	client, err := NewAIClient(createTestConfig(ProviderMock))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.ShouldEndInterview(tt.messageCount)
			if result != tt.expected {
				t.Errorf("ShouldEndInterview(%d) = %v, expected %v", tt.messageCount, result, tt.expected)
			}
		})
	}
}

// Test GetCurrentProvider
func TestGetCurrentProvider(t *testing.T) {
	tests := []struct {
		name     string
		provider string
	}{
		{"mock provider", ProviderMock},
		{"openai provider", ProviderOpenAI},
		{"gemini provider", ProviderGemini},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewAIClient(createTestConfig(tt.provider))
			if err != nil {
				t.Fatalf("Failed to create client: %v", err)
			}

			provider := client.GetCurrentProvider()
			if provider != tt.provider {
				t.Errorf("GetCurrentProvider() = %s, expected %s", provider, tt.provider)
			}
		})
	}
}

// Test GetCurrentModel
func TestGetCurrentModel(t *testing.T) {
	cfg := createTestConfig(ProviderMock)
	cfg.DefaultModel = "test-model-v1"

	client, err := NewAIClient(cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	model := client.GetCurrentModel()
	if model != "test-model-v1" {
		t.Errorf("GetCurrentModel() = %s, expected test-model-v1", model)
	}
}

// Test buildSystemPrompt variations
func TestBuildSystemPrompt(t *testing.T) {
	tests := []struct {
		name            string
		language        string
		isClosing       bool
		expectedContains []string
		notContains     []string
	}{
		{
			name:             "English, not closing",
			language:         "en",
			isClosing:        false,
			expectedContains: []string{"professional interviewer", "Ask one clear question", "Respond in English"},
			notContains:      []string{"wrap up", "繁體中文"},
		},
		{
			name:             "English, closing",
			language:         "en",
			isClosing:        true,
			expectedContains: []string{"professional interviewer", "wrap up the interview", "thank the candidate", "Respond in English"},
			notContains:      []string{"Ask one clear question", "繁體中文"},
		},
		{
			name:             "Traditional Chinese, not closing",
			language:         "zh-TW",
			isClosing:        false,
			expectedContains: []string{"professional interviewer", "Ask one clear question", "繁體中文"},
			notContains:      []string{"wrap up", "Respond in English"},
		},
		{
			name:             "Traditional Chinese, closing",
			language:         "zh-TW",
			isClosing:        true,
			expectedContains: []string{"professional interviewer", "wrap up", "thank the candidate", "繁體中文"},
			notContains:      []string{"Ask one clear question", "Respond in English"},
		},
		{
			name:             "lowercase zh-tw variant",
			language:         "zh-tw",
			isClosing:        false,
			expectedContains: []string{"繁體中文"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := buildSystemPrompt(tt.language, tt.isClosing)

			// Check expected strings are present
			for _, expected := range tt.expectedContains {
				if !contains(prompt, expected) {
					t.Errorf("buildSystemPrompt() missing expected string: %s", expected)
				}
			}

			// Check unwanted strings are absent
			for _, notExpected := range tt.notContains {
				if contains(prompt, notExpected) {
					t.Errorf("buildSystemPrompt() contains unexpected string: %s", notExpected)
				}
			}
		})
	}
}

// Test buildChatMessages with role conversion
func TestBuildChatMessages(t *testing.T) {
	tests := []struct {
		name            string
		history         []map[string]string
		userMessage     string
		language        string
		isClosing       bool
		expectedMsgCount int
		checkRoles      map[int]string  // index -> expected role
	}{
		{
			name:             "empty history",
			history:          []map[string]string{},
			userMessage:      "Hello",
			language:         "en",
			isClosing:        false,
			expectedMsgCount: 2,  // system + user
			checkRoles:       map[int]string{0: "system", 1: "user"},
		},
		{
			name: "history with ai role conversion",
			history: []map[string]string{
				{"role": "ai", "content": "Hi there!"},  // Should convert to "assistant"
				{"role": "user", "content": "Hello"},
			},
			userMessage:      "How are you?",
			language:         "en",
			isClosing:        false,
			expectedMsgCount: 4,  // system + ai + user + current user
			checkRoles:       map[int]string{0: "system", 1: "assistant", 2: "user", 3: "user"},
		},
		{
			name: "history without new message",
			history: []map[string]string{
				{"role": "user", "content": "Question"},
				{"role": "ai", "content": "Answer"},  // Should convert to "assistant"
			},
			userMessage:      "",  // Empty
			language:         "zh-TW",
			isClosing:        false,
			expectedMsgCount: 3,  // system + user + ai (no new message)
			checkRoles:       map[int]string{0: "system", 1: "user", 2: "assistant"},
		},
		{
			name:             "closing message in Chinese",
			history:          []map[string]string{},
			userMessage:      "謝謝",
			language:         "zh-TW",
			isClosing:        true,
			expectedMsgCount: 2,
			checkRoles:       map[int]string{0: "system", 1: "user"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			messages := buildChatMessages(tt.history, tt.userMessage, tt.language, tt.isClosing)

			// Check message count
			if len(messages) != tt.expectedMsgCount {
				t.Errorf("Expected %d messages, got %d", tt.expectedMsgCount, len(messages))
			}

			// Check role conversions
			for index, expectedRole := range tt.checkRoles {
				if index < len(messages) {
					if messages[index].Role != expectedRole {
						t.Errorf("Message[%d] role = %s, expected %s", index, messages[index].Role, expectedRole)
					}
				}
			}

			// Verify system prompt has correct language instruction
			if len(messages) > 0 {
				systemPrompt := messages[0].Content
				if tt.language == "zh-TW" || tt.language == "zh-tw" {
					if !contains(systemPrompt, "繁體中文") {
						t.Error("Expected Chinese language instruction in system prompt")
					}
				} else {
					if !contains(systemPrompt, "English") {
						t.Error("Expected English language instruction in system prompt")
					}
				}
			}
		})
	}
}

// Test GenerateChatResponse (calls GenerateChatResponseWithLanguage)
func TestGenerateChatResponse(t *testing.T) {
	client, err := NewAIClient(createTestConfig(ProviderMock))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	response, err := client.GenerateChatResponse("session1", []map[string]string{}, "Hello")
	if err != nil {
		t.Fatalf("GenerateChatResponse failed: %v", err)
	}

	// Verify mock provider response
	if !contains(response, "[MOCK]") {
		t.Errorf("Expected mock response to contain [MOCK], got: %s", response)
	}
}

// Test GenerateChatResponseWithLanguage
func TestGenerateChatResponseWithLanguage(t *testing.T) {
	client, err := NewAIClient(createTestConfig(ProviderMock))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name    string
		history []map[string]string
		message string
		lang    string
	}{
		{
			name:    "English conversation",
			history: []map[string]string{},
			message: "Tell me about yourself",
			lang:    "en",
		},
		{
			name: "Chinese conversation with history",
			history: []map[string]string{
				{"role": "ai", "content": "你好"},
				{"role": "user", "content": "我是工程師"},
			},
			message: "請介紹自己",
			lang:    "zh-TW",
		},
		{
			name:    "empty message",
			history: []map[string]string{},
			message: "",
			lang:    "en",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.GenerateChatResponseWithLanguage("session1", tt.history, tt.message, tt.lang)

			if err != nil {
				t.Errorf("GenerateChatResponseWithLanguage failed: %v", err)
				return
			}

			// Verify mock provider was called and returned response
			if response == "" {
				t.Error("Expected non-empty response from mock provider")
			}

			if !contains(response, "[MOCK]") && !contains(response, "[模擬]") {
				t.Errorf("Expected mock response marker, got: %s", response)
			}
		})
	}
}

// Test GenerateClosingMessage (calls GenerateClosingMessageWithLanguage)
func TestGenerateClosingMessage(t *testing.T) {
	client, err := NewAIClient(createTestConfig(ProviderMock))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	history := []map[string]string{
		{"role": "user", "content": "Hello"},
		{"role": "ai", "content": "Hi there"},
	}

	response, err := client.GenerateClosingMessage("session1", history, "Thank you")
	if err != nil {
		t.Fatalf("GenerateClosingMessage failed: %v", err)
	}

	// Verify closing message generated
	if response == "" {
		t.Error("Expected non-empty closing message")
	}
}

// Test GenerateClosingMessageWithLanguage
func TestGenerateClosingMessageWithLanguage(t *testing.T) {
	client, err := NewAIClient(createTestConfig(ProviderMock))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name string
		lang string
	}{
		{"English closing", "en"},
		{"Chinese closing", "zh-TW"},
	}

	history := []map[string]string{
		{"role": "user", "content": "Test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.GenerateClosingMessageWithLanguage("session1", history, "Goodbye", tt.lang)

			if err != nil {
				t.Errorf("GenerateClosingMessageWithLanguage failed: %v", err)
				return
			}

			if response == "" {
				t.Error("Expected non-empty closing message")
			}
		})
	}
}

// Test EvaluateAnswers (calls EvaluateAnswersWithContext)
func TestEvaluateAnswers(t *testing.T) {
	client, err := NewAIClient(createTestConfig(ProviderMock))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	questions := []string{"Tell me about yourself", "What are your strengths?"}
	answers := []string{"I am a developer", "Problem solving"}

	score, feedback, err := client.EvaluateAnswers(questions, answers, "en")
	if err != nil {
		t.Fatalf("EvaluateAnswers failed: %v", err)
	}

	// Verify score and feedback returned
	if score < 0 || score > 100 {
		t.Errorf("Score %f out of valid range [0-100]", score)
	}

	if feedback == "" {
		t.Error("Expected non-empty feedback")
	}
}

// Test EvaluateAnswersWithContext
func TestEvaluateAnswersWithContext(t *testing.T) {
	client, err := NewAIClient(createTestConfig(ProviderMock))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	tests := []struct {
		name      string
		questions []string
		answers   []string
		jobDesc   string
		lang      string
		wantScore float64  // For empty answers case
		wantErr   bool
	}{
		{
			name:      "normal evaluation",
			questions: []string{"Q1", "Q2"},
			answers:   []string{"A1", "A2"},
			jobDesc:   "Software Engineer",
			lang:      "en",
			wantErr:   false,
		},
		{
			name:      "Chinese evaluation",
			questions: []string{"問題一", "問題二"},
			answers:   []string{"答案一", "答案二"},
			jobDesc:   "軟體工程師",
			lang:      "zh-TW",
			wantErr:   false,
		},
		{
			name:      "empty answers edge case",
			questions: []string{"Q1", "Q2"},
			answers:   []string{},
			jobDesc:   "Engineer",
			lang:      "en",
			wantScore: 0.0,
			wantErr:   false,
		},
		{
			name:      "long job description",
			questions: []string{"Q1"},
			answers:   []string{"A1"},
			jobDesc:   "Senior Full-Stack Software Engineer with 10 years experience in distributed systems",
			lang:      "en",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, feedback, err := client.EvaluateAnswersWithContext(tt.questions, tt.answers, tt.jobDesc, tt.lang)

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			// For empty answers case
			if len(tt.answers) == 0 {
				if score != 0.0 {
					t.Errorf("Empty answers should return 0.0 score, got %f", score)
				}
				if feedback != "No answers provided." {
					t.Errorf("Empty answers should return specific message, got: %s", feedback)
				}
				return
			}

			// For normal cases
			if score < 0 || score > 100 {
				t.Errorf("Score %f out of valid range [0-100]", score)
			}

			if feedback == "" {
				t.Error("Expected non-empty feedback")
			}
		})
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
