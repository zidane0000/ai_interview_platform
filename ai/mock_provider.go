// Mock AI provider for testing and CI environments
package ai

import (
	"context"
	"strings"
	"time"
)

// MockProvider implements the AIProvider interface with canned responses
type MockProvider struct{}

func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) GenerateResponse(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	// Detect language from system prompt
	var isTraditionalChinese bool
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			if strings.Contains(msg.Content, "Traditional Chinese") || strings.Contains(msg.Content, "繁體中文") {
				isTraditionalChinese = true
			}
		}
	}

	// Simple language-appropriate mock response
	var mockResponse string
	if isTraditionalChinese {
		mockResponse = "[模擬] 面試問題回應 - 這是測試用的模擬回應"
	} else {
		mockResponse = "[MOCK] Interview response - This is a test mock response"
	}

	return &ChatResponse{
		Content:      mockResponse,
		FinishReason: "stop",
		TokensUsed:   TokenUsage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
		Model:        "mock-model",
		Provider:     "mock",
		ResponseTime: 10 * time.Millisecond,
		Timestamp:    time.Now(),
	}, nil
}

func (m *MockProvider) GenerateStreamResponse(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error) {
	ch := make(chan *ChatResponse, 1)
	ch <- &ChatResponse{
		Content:      "[MOCK] Streaming response - This is a test mock streaming response",
		FinishReason: "stop",
		TokensUsed:   TokenUsage{PromptTokens: 10, CompletionTokens: 20, TotalTokens: 30},
		Model:        "mock-model",
		Provider:     "mock",
		ResponseTime: 5 * time.Millisecond,
		Timestamp:    time.Now(),
	}
	close(ch)
	return ch, nil
}

func (m *MockProvider) GenerateInterviewQuestions(ctx context.Context, req *QuestionGenerationRequest) (*QuestionGenerationResponse, error) {
	// Simple mock questions
	questions := []InterviewQuestion{
		{
			Question:   "[MOCK] Test question 1",
			Category:   "technical",
			Difficulty: "medium",
		},
		{
			Question:   "[MOCK] Test question 2",
			Category:   "behavioral",
			Difficulty: "medium",
		},
		{
			Question:   "[MOCK] Test question 3",
			Category:   "technical",
			Difficulty: "medium",
		},
	}
	return &QuestionGenerationResponse{
		Questions:  questions,
		Rationale:  "[MOCK] Simple test question rationale",
		TokensUsed: TokenUsage{PromptTokens: 20, CompletionTokens: 40, TotalTokens: 60},
		Provider:   "mock",
		Model:      "mock-model",
		Timestamp:  time.Now(),
	}, nil
}

func (m *MockProvider) EvaluateAnswers(ctx context.Context, req *EvaluationRequest) (*EvaluationResponse, error) {
	// Simple language-appropriate mock evaluation
	var feedback string
	var strengths, weaknesses, recommendations []string

	if req.Language == "zh-TW" {
		feedback = "[模擬] 測試用評估回饋"
		strengths = []string{"[模擬] 測試優勢1", "[模擬] 測試優勢2"}
		weaknesses = []string{"[模擬] 測試弱點1", "[模擬] 測試弱點2"}
		recommendations = []string{"[模擬] 測試建議1", "[模擬] 測試建議2"}
	} else {
		feedback = "[MOCK] Test evaluation feedback"
		strengths = []string{"[MOCK] Test strength 1", "[MOCK] Test strength 2"}
		weaknesses = []string{"[MOCK] Test weakness 1", "[MOCK] Test weakness 2"}
		recommendations = []string{"[MOCK] Test recommendation 1", "[MOCK] Test recommendation 2"}
	}

	return &EvaluationResponse{
		OverallScore:    0.8,
		CategoryScores:  map[string]float64{"technical": 0.8, "communication": 0.85, "problem_solving": 0.75},
		Feedback:        feedback,
		Strengths:       strengths,
		Weaknesses:      weaknesses,
		Recommendations: recommendations,
		TokensUsed:      TokenUsage{PromptTokens: 50, CompletionTokens: 150, TotalTokens: 200},
		Provider:        "mock",
		Model:           "mock-model",
		Timestamp:       time.Now(),
	}, nil
}

func (m *MockProvider) GetProviderName() string                       { return "mock" }
func (m *MockProvider) GetSupportedModels() []string                  { return []string{"mock-model"} }
func (m *MockProvider) ValidateCredentials(ctx context.Context) error { return nil }
func (m *MockProvider) IsHealthy(ctx context.Context) bool            { return true }
func (m *MockProvider) GetUsageStats(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{"mock": true}, nil
}
