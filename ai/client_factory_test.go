package ai

import (
	"testing"

	"github.com/zidane0000/ai-interview-platform/config"
)

func TestNewAIClientFactory(t *testing.T) {
	cfg := config.Config{
		OpenAIAPIKey: "test-openai-key",
		GeminiAPIKey: "test-gemini-key",
	}

	factory := NewAIClientFactory(cfg)

	if factory == nil {
		t.Fatal("Expected factory to be created, got nil")
	}

	if factory.config.OpenAIAPIKey != "test-openai-key" {
		t.Errorf("Expected OpenAI API key to be 'test-openai-key', got '%s'", factory.config.OpenAIAPIKey)
	}
}

func TestCreateClient(t *testing.T) {
	cfg := config.Config{
		OpenAIAPIKey: "test-openai-key",
		GeminiAPIKey: "test-gemini-key",
	}

	factory := NewAIClientFactory(cfg)

	// Test creating client with default provider/model
	client, err := factory.CreateClient("", "")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	// Check that the client has the enhanced client
	if client.enhancedClient == nil {
		t.Fatal("Expected enhanced client to be created, got nil")
	}
}

func TestCreateClientWithSpecificProvider(t *testing.T) {
	cfg := config.Config{
		OpenAIAPIKey: "test-openai-key",
		GeminiAPIKey: "test-gemini-key",
	}

	factory := NewAIClientFactory(cfg)

	// Test creating client with specific provider and model
	client, err := factory.CreateClient("openai", "gpt-4")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	// Check that the client has the enhanced client
	if client.enhancedClient == nil {
		t.Fatal("Expected enhanced client to be created, got nil")
	}

	// Check that the current provider is set correctly
	if client.GetCurrentProvider() != "openai" {
		t.Errorf("Expected provider to be 'openai', got '%s'", client.GetCurrentProvider())
	}
}

func TestCreateDefaultClient(t *testing.T) {
	cfg := config.Config{
		OpenAIAPIKey: "test-openai-key",
		GeminiAPIKey: "test-gemini-key",
	}

	factory := NewAIClientFactory(cfg)

	client, err := factory.CreateDefaultClient()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if client == nil {
		t.Fatal("Expected client to be created, got nil")
	}

	// Check that the client has the enhanced client
	if client.enhancedClient == nil {
		t.Fatal("Expected enhanced client to be created, got nil")
	}
}

func TestCreateClientWithInvalidConfiguration(t *testing.T) {
	// Test with empty config (no API keys)
	cfg := config.Config{}
	factory := NewAIClientFactory(cfg)

	// Test creating client with invalid provider
	client, err := factory.CreateClient("invalid-provider", "test-model")
	if err == nil {
		t.Error("Expected error for invalid provider, got nil")
	}
	if client != nil {
		t.Error("Expected nil client for invalid provider, got non-nil")
	}

	// Test creating client with valid provider but no API key
	client, err = factory.CreateClient("openai", "gpt-4")
	if err == nil {
		t.Error("Expected error for missing API key, got nil")
	}
	if client != nil {
		t.Error("Expected nil client for missing API key, got non-nil")
	}
}
