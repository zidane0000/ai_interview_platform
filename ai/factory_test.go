package ai

import (
	"testing"
	"time"
)

// Test model parsing functionality
func TestParseModel(t *testing.T) {
	testCases := []struct {
		input            string
		expectedProvider string
		expectedModel    string
		isValid          bool
	}{
		// Valid formats
		{"openai/gpt-4o", "openai", "gpt-4o", true},
		{"google/gemini-pro", "google", "gemini-pro", true},
		{"mock/test", "mock", "test", true},
		{"anthropic/claude-3.5-sonnet", "anthropic", "claude-3.5-sonnet", true},

		// Invalid formats
		{"invalid-format", "", "", false},
		{"", "", "", false},
		{"openai/", "", "", false},
		{"/gpt-4o", "", "", false},
		{"openai/gpt-4o/extra", "", "", false},
		{" openai/gpt-4o ", "", "", false}, // with spaces
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			provider, model, err := parseModel(tc.input)

			if tc.isValid {
				// Valid input should not return an error
				if err != nil {
					t.Errorf("Expected no error for input %s, but got: %v", tc.input, err)
				}
				if provider != tc.expectedProvider {
					t.Errorf("Expected provider %s, got %s", tc.expectedProvider, provider)
				}
				if model != tc.expectedModel {
					t.Errorf("Expected model %s, got %s", tc.expectedModel, model)
				}
			} else {
				// Invalid input should return an error
				if err == nil {
					t.Errorf("Expected error for input %s, but got none", tc.input)
				}
			}
		})
	}
}

// Test provider factory functionality
func TestProviderFactory(t *testing.T) {
	testCases := []struct {
		model        string
		expectedType string
		isValid      bool
	}{
		{"openai/gpt-4o", "OpenAI", true},
		{"gemini/gemini-pro", "Gemini", true},
		{"mock/test", "Mock", true},
		{"unsupported/model", "", false},
		{"invalid-format", "", false},
	}

	// Create a test config with API keys
	config := &AIConfig{
		OpenAIAPIKey:   "test-openai-key",
		GeminiAPIKey:   "test-gemini-key",
		RequestTimeout: 30 * time.Second,
	}

	for _, tc := range testCases {
		t.Run(tc.model, func(t *testing.T) {
			provider, err := CreateProvider(tc.model, config)

			if tc.isValid {
				// Valid model should not return an error
				if err != nil {
					t.Errorf("Expected no error for model %s, but got: %v", tc.model, err)
				}
				if provider == nil {
					t.Errorf("Expected provider instance for model %s, but got nil", tc.model)
				}

				// Verify provider implements AIProvider interface
				if provider != nil {
					var _ AIProvider = provider
				}
			} else {
				// Invalid model should return an error
				if err == nil {
					t.Errorf("Expected error for model %s, but got none", tc.model)
				}
			}
		})
	}
}

// Test provider interface compatibility
func TestProviderInterface(t *testing.T) {
	// This test ensures all providers implement the AIProvider interface correctly
	// once the factory pattern is implemented

	models := []string{
		"openai/gpt-4o",
		"gemini/gemini-pro",
		"mock/test",
	}

	// Create a test config with API keys
	config := &AIConfig{
		OpenAIAPIKey:   "test-openai-key",
		GeminiAPIKey:   "test-gemini-key",
		RequestTimeout: 30 * time.Second,
	}

	for _, model := range models {
		t.Run(model, func(t *testing.T) {
			provider, err := CreateProvider(model, config)
			if err != nil {
				t.Fatalf("Failed to create provider for %s: %v", model, err)
			}

			// Test that provider implements AIProvider interface
			var _ AIProvider = provider

			// Test that basic provider info methods work
			providerName := provider.GetProviderName()
			if providerName == "" {
				t.Errorf("Provider name should not be empty for %s", model)
			}

			supportedModels := provider.GetSupportedModels()
			if len(supportedModels) == 0 {
				t.Errorf("Provider should support at least one model for %s", model)
			}

			// Note: We skip health check in unit tests as it requires real API calls
			// Health check functionality should be tested separately in integration tests
			t.Logf("Provider %s implements health check interface (not tested in unit tests)", model)

			t.Logf("Provider %s created successfully with name: %s", model, providerName)
		})
	}
}
