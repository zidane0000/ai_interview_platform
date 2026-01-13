// Factory for creating AI clients per request
package ai

import (
	"github.com/zidane0000/ai-interview-platform/config"
)

// AIClientFactory creates AI clients with proper configuration
type AIClientFactory struct {
	config config.Config
}

// NewAIClientFactory creates a new AI client factory with the given configuration
func NewAIClientFactory(cfg config.Config) *AIClientFactory {
	return &AIClientFactory{
		config: cfg,
	}
}

// CreateClient creates a new AI client instance with the specified provider and model
// If provider/model are empty, uses default configuration
func (f *AIClientFactory) CreateClient(provider, model string) (*AIClient, error) {
	// Create AI config based on the current configuration
	aiConfig := f.createAIConfig(provider, model)

	// Validate the configuration before creating the client
	if err := ValidateConfig(aiConfig); err != nil {
		return nil, err
	}

	// Create enhanced AI client with the config
	enhancedClient := NewEnhancedAIClient(aiConfig)

	// Create and return the AI client
	client := &AIClient{
		enhancedClient: enhancedClient,
	}

	return client, nil
}

// CreateDefaultClient creates a new AI client with default configuration
func (f *AIClientFactory) CreateDefaultClient() (*AIClient, error) {
	return f.CreateClient("", "")
}

// createAIConfig creates an AI configuration based on the factory's config and optional overrides
func (f *AIClientFactory) createAIConfig(provider, model string) *AIConfig {
	// Start with default configuration
	aiConfig := NewDefaultAIConfig()

	// Override with specific parameters if provided
	if provider != "" {
		aiConfig.DefaultProvider = provider
	}
	if model != "" {
		aiConfig.DefaultModel = model
	}

	// Set API keys from main config if available
	if f.config.OpenAIAPIKey != "" {
		aiConfig.OpenAIAPIKey = f.config.OpenAIAPIKey
	}
	if f.config.GeminiAPIKey != "" {
		aiConfig.GeminiAPIKey = f.config.GeminiAPIKey
	}
	return aiConfig
}
