// Google Gemini provider implementation
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Default model for Gemini API
const defaultGeminiModel = "gemini-1.5-flash"

// GeminiProvider implements the AIProvider interface for Google Gemini API
type GeminiProvider struct {
	BaseProvider
	apiKey string
}

// Gemini API structures
type geminiRequest struct {
	Contents         []geminiContent  `json:"contents"`
	GenerationConfig *geminiGenConfig `json:"generationConfig,omitempty"`
	SafetySettings   []geminiSafety   `json:"safetySettings,omitempty"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
	Role  string       `json:"role,omitempty"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiGenConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	TopK            int      `json:"topK,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type geminiSafety struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type geminiResponse struct {
	Candidates    []geminiCandidate `json:"candidates"`
	UsageMetadata *geminiUsage      `json:"usageMetadata"`
	Error         *geminiError      `json:"error,omitempty"`
}

type geminiCandidate struct {
	Content       geminiContent        `json:"content"`
	FinishReason  string               `json:"finishReason"`
	Index         int                  `json:"index"`
	SafetyRatings []geminiSafetyRating `json:"safetyRatings"`
}

type geminiSafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type geminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type geminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// NewGeminiProvider creates a new Gemini provider
func NewGeminiProvider(apiKey string, config *AIConfig) *GeminiProvider {
	// Use custom base URL if provided, otherwise default to Google
	baseURL := config.GeminiBaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}

	return &GeminiProvider{
		BaseProvider: NewBaseProvider(config, baseURL, config.RequestTimeout),
		apiKey:       apiKey,
	}
}

// --- ProviderAdapter interface implementation ---

// SetAuth is a no-op for Gemini (uses URL parameter instead)
func (p *GeminiProvider) SetAuth(req *http.Request) {
	// Gemini uses API key in URL, not header
}

// GetEndpointURL returns the full URL with API key for Gemini endpoints
func (p *GeminiProvider) GetEndpointURL(endpoint string) string {
	return p.baseURL + endpoint + "?key=" + p.apiKey
}

// --- Provider-specific methods ---

func (p *GeminiProvider) convertMessages(messages []Message) []geminiContent {
	var contents []geminiContent
	var systemMessages []string

	// Single pass: collect system messages and convert others
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg.Content)
			continue
		}

		// Gemini uses "user" and "model" roles, not "assistant"
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}

		contents = append(contents, geminiContent{
			Parts: []geminiPart{{Text: msg.Content}},
			Role:  role,
		})
	}

	// Prepend system messages to first user message
	if len(systemMessages) > 0 && len(contents) > 0 {
		systemPrompt := strings.Join(systemMessages, "\n\n")
		contents[0].Parts[0].Text = systemPrompt + "\n\n" + contents[0].Parts[0].Text
	}

	return contents
}

func (p *GeminiProvider) getDefaultSafetySettings() []geminiSafety {
	return []geminiSafety{
		{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
		{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
		{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
		{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
	}
}

// --- AIProvider interface implementation ---

// GenerateResponse generates a chat completion using Gemini API
func (p *GeminiProvider) GenerateResponse(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	geminiReq := &geminiRequest{
		Contents: p.convertMessages(req.Messages),
		GenerationConfig: &geminiGenConfig{
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			MaxOutputTokens: req.MaxTokens,
		},
		SafetySettings: p.getDefaultSafetySettings(),
	}

	model := p.GetModelName(req.Model, defaultGeminiModel)
	endpoint := fmt.Sprintf("/models/%s:generateContent", model)

	respData, err := p.MakeRequest(ctx, p, endpoint, geminiReq)
	if err != nil {
		return nil, fmt.Errorf("Gemini API request failed: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respData, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error: %s (code: %d)", geminiResp.Error.Message, geminiResp.Error.Code)
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from Gemini")
	}

	candidate := geminiResp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("no content parts in Gemini response")
	}

	content := candidate.Content.Parts[0].Text

	var tokensUsed TokenUsage
	if geminiResp.UsageMetadata != nil {
		tokensUsed = TokenUsage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		}
	}

	return &ChatResponse{
		Content:      content,
		FinishReason: candidate.FinishReason,
		TokensUsed:   tokensUsed,
		Model:        model,
		Provider:     ProviderGemini,
		ResponseTime: time.Since(startTime),
		Timestamp:    time.Now(),
		Metadata: map[string]interface{}{
			"index":          candidate.Index,
			"safety_ratings": candidate.SafetyRatings,
		},
	}, nil
}

// GenerateStreamResponse generates a streaming response (placeholder for now)
func (p *GeminiProvider) GenerateStreamResponse(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error) {
	return nil, fmt.Errorf("streaming not yet implemented for Gemini provider")
}

// GenerateInterviewQuestions generates interview questions using Gemini
func (p *GeminiProvider) GenerateInterviewQuestions(ctx context.Context, req *QuestionGenerationRequest) (*QuestionGenerationResponse, error) {
	systemPrompt := BuildQuestionGenerationPrompt(req)

	// Gemini handles system messages differently - combine into user message
	chatReq := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: systemPrompt + fmt.Sprintf("\n\nGenerate %d interview questions based on this job description: %s", req.NumQuestions, req.JobDescription),
			},
		},
		Model:       p.GetModelName("", defaultGeminiModel),
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
		Rationale:  "Questions generated based on job requirements and candidate experience using Gemini AI",
		TokensUsed: response.TokensUsed,
		Provider:   ProviderGemini,
		Model:      response.Model,
		Timestamp:  time.Now(),
	}, nil
}

// EvaluateAnswers evaluates interview answers using Gemini
func (p *GeminiProvider) EvaluateAnswers(ctx context.Context, req *EvaluationRequest) (*EvaluationResponse, error) {
	systemPrompt := BuildEvaluationPrompt(req)
	userContent := FormatAnswersForEvaluation(req.Questions, req.Answers)

	// Gemini handles system messages differently - combine into user message
	chatReq := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: systemPrompt + "\n\n" + userContent,
			},
		},
		Model:       p.GetModelName("", defaultGeminiModel),
		MaxTokens:   3000,
		Temperature: 0.3,
	}

	response, err := p.GenerateResponse(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate answers: %w", err)
	}

	evaluation := ParseEvaluationResponse(response.Content)
	evaluation.TokensUsed = response.TokensUsed
	evaluation.Provider = ProviderGemini
	evaluation.Model = response.Model
	evaluation.Timestamp = time.Now()

	return evaluation, nil
}

// GetProviderName returns the provider name
func (p *GeminiProvider) GetProviderName() string {
	return ProviderGemini
}

// GetSupportedModels returns list of supported Gemini models
func (p *GeminiProvider) GetSupportedModels() []string {
	return []string{
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"gemini-pro",
		"gemini-pro-vision",
	}
}

// ValidateCredentials validates the API key
func (p *GeminiProvider) ValidateCredentials(ctx context.Context) error {
	testReq := &geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{{Text: "Hello"}},
			},
		},
		GenerationConfig: &geminiGenConfig{
			MaxOutputTokens: 5,
		},
	}

	model := p.GetModelName("", defaultGeminiModel)
	endpoint := fmt.Sprintf("/models/%s:generateContent", model)
	_, err := p.MakeRequest(ctx, p, endpoint, testReq)
	return err
}

// IsHealthy checks if the provider is healthy
func (p *GeminiProvider) IsHealthy(ctx context.Context) bool {
	return p.ValidateCredentials(ctx) == nil
}

// GetUsageStats returns usage statistics (placeholder)
func (p *GeminiProvider) GetUsageStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"provider": ProviderGemini,
		"status":   "healthy",
	}, nil
}
