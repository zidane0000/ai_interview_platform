// Simplified AI client for MVP - direct provider access without metrics/caching
package ai

import (
	"context"
	"fmt"
)

// AIClient provides a simple interface for AI operations
// Wraps a single AIProvider without enterprise features (metrics, caching, health checks)
type AIClient struct {
	provider AIProvider
	config   *AIConfig
}

// NewAIClient creates a new AI client with the specified configuration
func NewAIClient(cfg *AIConfig) (*AIClient, error) {
	if err := ValidateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// Create the provider based on default provider setting
	var provider AIProvider

	switch cfg.DefaultProvider {
	case ProviderOpenAI:
		if cfg.OpenAIAPIKey == "" {
			return nil, fmt.Errorf("OpenAI API key required")
		}
		provider = NewOpenAIProvider(cfg.OpenAIAPIKey, cfg)
	case ProviderGemini:
		if cfg.GeminiAPIKey == "" {
			return nil, fmt.Errorf("Gemini API key required")
		}
		provider = NewGeminiProvider(cfg.GeminiAPIKey, cfg)
	case ProviderMock:
		provider = NewMockProvider()
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.DefaultProvider)
	}

	return &AIClient{
		provider: provider,
		config:   cfg,
	}, nil
}

// GenerateChatResponse generates AI response for conversational interviews
func (c *AIClient) GenerateChatResponse(sessionID string, conversationHistory []map[string]string, userMessage string) (string, error) {
	return c.GenerateChatResponseWithLanguage(sessionID, conversationHistory, userMessage, "en")
}

// GenerateChatResponseWithLanguage generates AI response with language support
func (c *AIClient) GenerateChatResponseWithLanguage(sessionID string, conversationHistory []map[string]string, userMessage string, language string) (string, error) {
	ctx := context.Background()

	// Build messages for the AI provider
	messages := buildChatMessages(conversationHistory, userMessage, language, false)

	// Generate response using provider
	req := &ChatRequest{
		Messages:    messages,
		MaxTokens:   500,
		Temperature: 0.7,
		SessionID:   sessionID,
	}

	resp, err := c.provider.GenerateResponse(ctx, req)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	return resp.Content, nil
}

// GenerateClosingMessage generates a closing AI response for ending interviews
func (c *AIClient) GenerateClosingMessage(sessionID string, conversationHistory []map[string]string, userMessage string) (string, error) {
	return c.GenerateClosingMessageWithLanguage(sessionID, conversationHistory, userMessage, "en")
}

// GenerateClosingMessageWithLanguage generates a closing AI response with language support
func (c *AIClient) GenerateClosingMessageWithLanguage(sessionID string, conversationHistory []map[string]string, userMessage string, language string) (string, error) {
	ctx := context.Background()

	// Build messages with closing context
	messages := buildChatMessages(conversationHistory, userMessage, language, true)

	// Generate closing response
	req := &ChatRequest{
		Messages:    messages,
		MaxTokens:   300,
		Temperature: 0.7,
		SessionID:   sessionID,
	}

	resp, err := c.provider.GenerateResponse(ctx, req)
	if err != nil {
		return "", fmt.Errorf("AI generation failed: %w", err)
	}

	return resp.Content, nil
}

// ShouldEndInterview determines if the interview should end
func (c *AIClient) ShouldEndInterview(messageCount int) bool {
	return messageCount >= 8 // End after 8 user messages
}

// EvaluateAnswers evaluates chat conversation and generates score and feedback
func (c *AIClient) EvaluateAnswers(questions []string, answers []string, language string) (float64, string, error) {
	return c.EvaluateAnswersWithContext(questions, answers, "General interview evaluation", language)
}

// EvaluateAnswersWithContext evaluates chat conversation with interview context
func (c *AIClient) EvaluateAnswersWithContext(questions []string, answers []string, jobDesc, language string) (float64, string, error) {
	if len(answers) == 0 {
		return 0.0, "No answers provided.", nil
	}

	ctx := context.Background()

	// Create evaluation request using existing types
	req := &EvaluationRequest{
		Questions:   questions,
		Answers:     answers,
		JobDesc:     jobDesc,
		Criteria:    []string{"communication", "technical_knowledge", "problem_solving", "clarity", "cultural_fit"},
		DetailLevel: "detailed",
		Language:    language,
		Context: map[string]interface{}{
			"interview_type":  "conversational",
			"evaluation_type": "chat_based",
		},
	}

	// Use provider's EvaluateAnswers method
	resp, err := c.provider.EvaluateAnswers(ctx, req)
	if err != nil {
		return 0.0, "Evaluation failed", fmt.Errorf("AI evaluation failed: %w", err)
	}

	return resp.OverallScore, resp.Feedback, nil
}

// GetCurrentProvider returns the currently configured AI provider
func (c *AIClient) GetCurrentProvider() string {
	return c.provider.GetProviderName()
}

// GetCurrentModel returns the currently configured AI model
func (c *AIClient) GetCurrentModel() string {
	return c.config.DefaultModel
}

// buildChatMessages builds message array for chat generation
// Helper function (not a method to avoid parameter issues)
func buildChatMessages(history []map[string]string, userMessage, language string, isClosing bool) []Message {
	systemPrompt := buildSystemPrompt(language, isClosing)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
	}

	// Add conversation history
	for _, msg := range history {
		if role, ok := msg["role"]; ok {
			if content, ok := msg["content"]; ok {
				messages = append(messages, Message{
					Role:    role,
					Content: content,
				})
			}
		}
	}

	// Add current user message if provided
	if userMessage != "" {
		messages = append(messages, Message{
			Role:    "user",
			Content: userMessage,
		})
	}

	return messages
}

// buildSystemPrompt creates system prompt for chat
func buildSystemPrompt(language string, isClosing bool) string {
	basePrompt := "You are a professional interviewer conducting a job interview. "
	basePrompt += "Ask thoughtful questions, engage naturally with the candidate, "
	basePrompt += "and create a comfortable interview atmosphere. "

	if isClosing {
		basePrompt += "This is the final message - wrap up the interview professionally, "
		basePrompt += "thank the candidate for their time, and let them know next steps will follow."
	} else {
		basePrompt += "Ask one clear question at a time and listen carefully to responses."
	}

	// Add language instruction
	if language == "zh-TW" || language == "zh-tw" {
		basePrompt += " Respond in Traditional Chinese (繁體中文)."
	} else {
		basePrompt += " Respond in English."
	}

	return basePrompt
}
