// Provider factory for creating AI provider instances
package ai

import (
	"fmt"
	"strings"
)

// parseModel parses a model string in "provider/model" format
// Returns provider, model, and error
func parseModel(model string) (provider, modelName string, err error) {
	if model == "" {
		return "", "", fmt.Errorf("model string cannot be empty")
	}

	// Check for leading/trailing whitespace - reject if present
	if strings.TrimSpace(model) != model {
		return "", "", fmt.Errorf("model string cannot have leading or trailing whitespace: '%s'", model)
	}

	// Split by "/"
	parts := strings.Split(model, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid model format, expected 'provider/model', got '%s'", model)
	}

	provider = parts[0]
	modelName = parts[1]

	// Validate non-empty parts
	if provider == "" {
		return "", "", fmt.Errorf("provider name cannot be empty in '%s'", model)
	}
	if modelName == "" {
		return "", "", fmt.Errorf("model name cannot be empty in '%s'", model)
	}

	return provider, modelName, nil
}

// CreateProvider creates an AI provider instance based on the model string
// Supports "provider/model" format and returns appropriate provider
func CreateProvider(model string, config *AIConfig) (AIProvider, error) {
	// If empty model, use default provider
	if model == "" {
		return NewMockProvider(), nil
	}

	// Parse the model string
	provider, modelName, err := parseModel(model)
	if err != nil {
		return nil, fmt.Errorf("failed to parse model '%s': %w", model, err)
	}

	// Store the model name for later use (avoid unused variable warning)
	_ = modelName

	// Create provider based on provider name
	switch provider {
	case ProviderOpenAI:
		apiKey := config.OpenAIAPIKey
		if apiKey == "" {
			return nil, fmt.Errorf("OpenAI API key is required")
		}
		return NewOpenAIProvider(apiKey, config), nil
	case ProviderGemini:
		apiKey := config.GeminiAPIKey
		if apiKey == "" {
			return nil, fmt.Errorf("Gemini API key is required")
		}
		return NewGeminiProvider(apiKey, config), nil
	case ProviderMock:
		return NewMockProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}
