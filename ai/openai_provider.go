// OpenAI provider implementation
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenAIProvider implements the AIProvider interface for OpenAI API
type OpenAIProvider struct {
	apiKey     string
	config     *AIConfig
	httpClient *http.Client
	baseURL    string
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
	return &OpenAIProvider{
		apiKey:  apiKey,
		config:  config,
		baseURL: "https://api.openai.com/v1",
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}
}

// GenerateResponse generates a chat completion using OpenAI API
func (p *OpenAIProvider) GenerateResponse(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	// Convert our request to OpenAI format
	openAIReq := &openAIRequest{
		Model:       p.getModelName(req.Model),
		Messages:    p.convertMessages(req.Messages),
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		Stream:      req.Stream,
	}

	// Make HTTP request to OpenAI
	respData, err := p.makeRequest(ctx, "/chat/completions", openAIReq)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}

	// Parse OpenAI response
	var openAIResp openAIResponse
	if err := json.Unmarshal(respData, &openAIResp); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI response: %w", err)
	}

	// Check for API errors
	if openAIResp.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s (%s)", openAIResp.Error.Message, openAIResp.Error.Type)
	}

	// Convert to our response format
	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices returned from OpenAI")
	}

	choice := openAIResp.Choices[0]
	response := &ChatResponse{
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
	}

	return response, nil
}

// GenerateStreamResponse generates a streaming response (placeholder for now)
func (p *OpenAIProvider) GenerateStreamResponse(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error) {
	// TODO: Implement streaming support
	return nil, fmt.Errorf("streaming not yet implemented for OpenAI provider")
}

// GenerateInterviewQuestions generates interview questions using OpenAI
func (p *OpenAIProvider) GenerateInterviewQuestions(ctx context.Context, req *QuestionGenerationRequest) (*QuestionGenerationResponse, error) {
	// Build prompt for question generation
	systemPrompt := p.buildQuestionGenerationPrompt(req)

	chatReq := &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Generate %d interview questions based on this job description: %s", req.NumQuestions, req.JobDescription),
			},
		},
		Model:       p.getModelName(""),
		MaxTokens:   2000,
		Temperature: 0.7,
	}

	response, err := p.GenerateResponse(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate questions: %w", err)
	}

	// Parse the response to extract questions
	questions := p.parseQuestionResponse(response.Content)

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
	// Build evaluation prompt
	systemPrompt := p.buildEvaluationPrompt(req)

	// Combine questions and answers for evaluation
	userContent := p.formatAnswersForEvaluation(req.Questions, req.Answers)

	chatReq := &ChatRequest{
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "user",
				Content: userContent,
			},
		},
		Model:       p.getModelName(""),
		MaxTokens:   3000,
		Temperature: 0.3, // Lower temperature for more consistent evaluation
	}

	response, err := p.GenerateResponse(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate answers: %w", err)
	}

	// Parse evaluation response
	evaluation := p.parseEvaluationResponse(response.Content)
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
	// Make a simple request to validate credentials
	testReq := &openAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []openAIMessage{
			{Role: "user", Content: "Hello"},
		},
		MaxTokens: 5,
	}

	_, err := p.makeRequest(ctx, "/chat/completions", testReq)
	return err
}

// IsHealthy checks if the provider is healthy
func (p *OpenAIProvider) IsHealthy(ctx context.Context) bool {
	err := p.ValidateCredentials(ctx)
	return err == nil
}

// GetUsageStats returns usage statistics (placeholder)
func (p *OpenAIProvider) GetUsageStats(ctx context.Context) (map[string]interface{}, error) {
	// TODO: Implement usage statistics retrieval
	return map[string]interface{}{
		"provider": ProviderOpenAI,
		"status":   "healthy",
	}, nil
}

// Helper methods

func (p *OpenAIProvider) getModelName(model string) string {
	if model == "" {
		return p.config.DefaultModel
	}
	return model
}

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

func (p *OpenAIProvider) makeRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (p *OpenAIProvider) buildQuestionGenerationPrompt(req *QuestionGenerationRequest) string {
	return fmt.Sprintf(`You are an expert interviewer tasked with generating high-quality interview questions.

Experience Level: %s
Interview Type: %s
Difficulty: %s

Job Description:
%s

Candidate Resume:
%s

Generate %d relevant interview questions that:
1. Assess the candidate's skills and experience based on the job description
2. Are appropriate for the %s level
3. Focus on %s aspects
4. Match the difficulty level: %s

Format each question as:
Question: [question text]
Category: [technical/behavioral/situational]
Difficulty: [easy/medium/hard]
Expected Time: [minutes]

Provide diverse questions that thoroughly evaluate the candidate for this role.`,
		req.ExperienceLevel, req.InterviewType, req.Difficulty,
		req.JobDescription, req.ResumeContent, req.NumQuestions,
		req.ExperienceLevel, req.InterviewType, req.Difficulty)
}

func (p *OpenAIProvider) buildEvaluationPrompt(req *EvaluationRequest) string {
	criteriaText := strings.Join(req.Criteria, ", ")

	return fmt.Sprintf(`You are an expert interview evaluator. Evaluate the candidate's answers objectively and provide detailed feedback.

Job Description: %s
Evaluation Criteria: %s
Detail Level: %s

Provide evaluation in this format:
Overall Score: [0.0-1.0]
Category Scores: 
- Technical Skills: [0.0-1.0]
- Communication: [0.0-1.0]
- Problem Solving: [0.0-1.0]
- Experience: [0.0-1.0]

Feedback: [comprehensive feedback paragraph]

Strengths:
- [strength 1]
- [strength 2]

Areas for Improvement:
- [area 1]
- [area 2]

Recommendations:
- [specific recommendation 1]
- [specific recommendation 2]

Be specific, constructive, and fair in your evaluation.`,
		req.JobDesc, criteriaText, req.DetailLevel)
}

func (p *OpenAIProvider) formatAnswersForEvaluation(questions, answers []string) string {
	var content strings.Builder
	content.WriteString("Interview Questions and Candidate Answers:\n\n")

	for i := 0; i < len(questions) && i < len(answers); i++ {
		content.WriteString(fmt.Sprintf("Q%d: %s\n", i+1, questions[i]))
		content.WriteString(fmt.Sprintf("A%d: %s\n\n", i+1, answers[i]))
	}

	return content.String()
}

func (p *OpenAIProvider) parseQuestionResponse(content string) []InterviewQuestion {
	// Simple parsing - in production, you might want more sophisticated parsing
	lines := strings.Split(content, "\n")
	var questions []InterviewQuestion

	var currentQuestion InterviewQuestion
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Question:") {
			currentQuestion.Question = strings.TrimSpace(line[9:])
		} else if strings.HasPrefix(line, "Category:") {
			currentQuestion.Category = strings.TrimSpace(line[9:])
		} else if strings.HasPrefix(line, "Difficulty:") {
			currentQuestion.Difficulty = strings.TrimSpace(line[11:])
		} else if strings.HasPrefix(line, "Expected Time:") {
			// Parse expected time - simplified
			currentQuestion.ExpectedTime = 5 // Default 5 minutes

			// If we have all required fields, add the question
			if currentQuestion.Question != "" {
				questions = append(questions, currentQuestion)
				currentQuestion = InterviewQuestion{} // Reset
			}
		}
	}

	return questions
}

func (p *OpenAIProvider) parseEvaluationResponse(content string) *EvaluationResponse {
	// Simple parsing - in production, you might want more sophisticated parsing or structured output
	evaluation := &EvaluationResponse{
		OverallScore:    0.7, // Default fallback
		CategoryScores:  make(map[string]float64),
		Strengths:       []string{},
		Weaknesses:      []string{},
		Recommendations: []string{},
	}

	// Extract feedback (everything between "Feedback:" and "Strengths:")
	lines := strings.Split(content, "\n")
	var feedbackLines []string
	inFeedback := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Feedback:") {
			inFeedback = true
			feedbackText := strings.TrimSpace(line[9:])
			if feedbackText != "" {
				feedbackLines = append(feedbackLines, feedbackText)
			}
			continue
		}
		if strings.HasPrefix(line, "Strengths:") || strings.HasPrefix(line, "Areas for Improvement:") {
			inFeedback = false
			continue
		}
		if inFeedback && line != "" {
			feedbackLines = append(feedbackLines, line)
		}

		// Parse strengths and weaknesses
		if strings.HasPrefix(line, "- ") && len(line) > 2 {
			item := strings.TrimSpace(line[2:])
			// This is a simplified parser - you'd want better logic to categorize
			evaluation.Strengths = append(evaluation.Strengths, item)
		}
	}

	evaluation.Feedback = strings.Join(feedbackLines, " ")

	// Set some default category scores
	evaluation.CategoryScores["technical"] = 0.7
	evaluation.CategoryScores["communication"] = 0.8
	evaluation.CategoryScores["problem_solving"] = 0.6
	evaluation.CategoryScores["experience"] = 0.7

	return evaluation
}
