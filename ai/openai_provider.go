// OpenAI provider implementation
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OpenAIProvider implements the AIProvider interface for OpenAI API
type OpenAIProvider struct {
	BaseProvider
	apiKey string
}

// OpenAI API request/response structures
type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Temperature float64         `json:"temperature,omitempty"`
	TopP        float64         `json:"top_p,omitempty"`
	Stream      bool            `json:"stream,omitempty"`
	Stop        []string        `json:"stop,omitempty"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Usage   openAIUsage    `json:"usage"`
	Choices []openAIChoice `json:"choices"`
	Error   *openAIError   `json:"error,omitempty"`
}

type openAIChoice struct {
	Index        int           `json:"index"`
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type openAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string, config *AIConfig) *OpenAIProvider {
	// Use custom base URL if provided, otherwise default to OpenAI
	baseURL := config.OpenAIBaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &OpenAIProvider{
		BaseProvider: NewBaseProvider(config, baseURL, config.RequestTimeout),
		apiKey:       apiKey,
	}
}

// --- ProviderAdapter interface implementation ---

// SetAuth sets OpenAI authentication header
func (p *OpenAIProvider) SetAuth(req *http.Request) {
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
}

// GetEndpointURL returns the full URL for OpenAI endpoints
func (p *OpenAIProvider) GetEndpointURL(endpoint string) string {
	return p.baseURL + endpoint
}

// --- Provider-specific message conversion ---

func (p *OpenAIProvider) convertMessages(messages []Message) []openAIMessage {
	converted := make([]openAIMessage, len(messages))
	for i, msg := range messages {
		converted[i] = openAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}
	return converted
}

// --- AIProvider interface implementation ---

// GenerateResponse generates a chat completion using OpenAI API
func (p *OpenAIProvider) GenerateResponse(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	openAIReq := &openAIRequest{
		Model:       p.GetModelName(req.Model, ""),
		Messages:    p.convertMessages(req.Messages),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
	}

	respData, err := p.MakeRequest(ctx, p, "/chat/completions", openAIReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}

	var openAIResp openAIResponse
	if err := json.Unmarshal(respData, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s (%s)", openAIResp.Error.Message, openAIResp.Error.Type)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	choice := openAIResp.Choices[0]
	return &ChatResponse{
		Content:      choice.Message.Content,
		FinishReason: choice.FinishReason,
		TokensUsed: TokenUsage{
			PromptTokens:     openAIResp.Usage.PromptTokens,
			CompletionTokens: openAIResp.Usage.CompletionTokens,
			TotalTokens:      openAIResp.Usage.TotalTokens,
		},
		Model:        openAIResp.Model,
		Provider:     ProviderOpenAI,
		ResponseTime: time.Since(startTime),
		Timestamp:    time.Now(),
		Metadata: map[string]interface{}{
			"id":      openAIResp.ID,
			"created": openAIResp.Created,
		},
	}, nil
}

// GenerateStreamResponse generates a streaming response (placeholder for now)
func (p *OpenAIProvider) GenerateStreamResponse(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error) {
	return nil, fmt.Errorf("streaming not yet implemented for OpenAI provider")
}

// GenerateInterviewQuestions generates interview questions using OpenAI
func (p *OpenAIProvider) GenerateInterviewQuestions(ctx context.Context, req *QuestionGenerationRequest) (*QuestionGenerationResponse, error) {
	systemPrompt := BuildQuestionGenerationPrompt(req)

	chatReq := &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fmt.Sprintf("Generate %d interview questions based on this job description: %s", req.NumQuestions, req.JobDescription)},
		},
		Model:       p.GetModelName("", ""),
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	response, err := p.GenerateResponse(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate questions: %w", err)
	}

	questions := ParseQuestionResponse(response.Content)

	return &QuestionGenerationResponse{
		Questions:  questions,
		Rationale:  "Questions generated based on job requirements and candidate experience",
		TokensUsed: response.TokensUsed,
		Provider:   ProviderOpenAI,
		Model:      response.Model,
		Timestamp:  time.Now(),
	}, nil
}

// EvaluateAnswers evaluates interview answers using OpenAI
func (p *OpenAIProvider) EvaluateAnswers(ctx context.Context, req *EvaluationRequest) (*EvaluationResponse, error) {
	systemPrompt := BuildEvaluationPrompt(req)
	userContent := FormatAnswersForEvaluation(req.Questions, req.Answers)

	chatReq := &ChatRequest{
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userContent},
		},
		Model:       p.GetModelName("", ""),
		MaxTokens:   3000,
		Temperature: 0.3,
	}

	response, err := p.GenerateResponse(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate answers: %w", err)
	}

	evaluation := ParseEvaluationResponse(response.Content)
	evaluation.TokensUsed = response.TokensUsed
	evaluation.Provider = ProviderOpenAI
	evaluation.Model = response.Model
	evaluation.Timestamp = time.Now()

	return evaluation, nil
}

// GetProviderName returns the provider name
func (p *OpenAIProvider) GetProviderName() string {
	return ProviderOpenAI
}

// GetSupportedModels returns list of supported OpenAI models
func (p *OpenAIProvider) GetSupportedModels() []string {
	return []string{
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4-turbo-preview",
		"gpt-3.5-turbo",
		"gpt-3.5-turbo-16k",
	}
}

// ValidateCredentials validates the API key
func (p *OpenAIProvider) ValidateCredentials(ctx context.Context) error {
	testReq := &openAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openAIMessage{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 5,
	}

	_, err := p.MakeRequest(ctx, p, "/chat/completions", testReq)
	return err
}

// IsHealthy checks if the provider is healthy
func (p *OpenAIProvider) IsHealthy(ctx context.Context) bool {
	return p.ValidateCredentials(ctx) == nil
}

// GetUsageStats returns usage statistics (placeholder)
func (p *OpenAIProvider) GetUsageStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"provider": ProviderOpenAI,
		"status":   "healthy",
	}, nil
}
