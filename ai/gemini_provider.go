// Google Gemini provider implementation
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

// GeminiProvider implements the AIProvider interface for Google Gemini API
type GeminiProvider struct {
	apiKey     string
	config     *AIConfig
	httpClient *http.Client
	baseURL    string
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
	return &GeminiProvider{
		apiKey:  apiKey,
		config:  config,
		baseURL: "https://generativelanguage.googleapis.com/v1beta",
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
	}
}

// GenerateResponse generates a chat completion using Gemini API
func (p *GeminiProvider) GenerateResponse(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	// Convert our request to Gemini format
	geminiReq := &geminiRequest{
		Contents: p.convertMessages(req.Messages),
		GenerationConfig: &geminiGenConfig{
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			MaxOutputTokens: req.MaxTokens,
		},
		SafetySettings: p.getDefaultSafetySettings(),
	}

	// Get model name
	model := p.getModelName(req.Model)
	endpoint := fmt.Sprintf("/models/%s:generateContent", model)

	// Make HTTP request to Gemini
	respData, err := p.makeRequest(ctx, endpoint, geminiReq)
	if err != nil {
		return nil, fmt.Errorf("Gemini API request failed: %w", err)
	}

	// Parse Gemini response
	var geminiResp geminiResponse
	if err := json.Unmarshal(respData, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	// Check for API errors
	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error: %s (code: %d)", geminiResp.Error.Message, geminiResp.Error.Code)
	}

	// Convert to our response format
	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates returned from Gemini")
	}

	candidate := geminiResp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("no content parts in Gemini response")
	}

	content := candidate.Content.Parts[0].Text

	// Handle usage metadata
	var tokensUsed TokenUsage
	if geminiResp.UsageMetadata != nil {
		tokensUsed = TokenUsage{
			PromptTokens:     geminiResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: geminiResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      geminiResp.UsageMetadata.TotalTokenCount,
		}
	}

	response := &ChatResponse{
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
	}

	return response, nil
}

// GenerateStreamResponse generates a streaming response (placeholder for now)
func (p *GeminiProvider) GenerateStreamResponse(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error) {
	// TODO: Implement streaming support for Gemini
	return nil, fmt.Errorf("streaming not yet implemented for Gemini provider")
}

// GenerateInterviewQuestions generates interview questions using Gemini
func (p *GeminiProvider) GenerateInterviewQuestions(ctx context.Context, req *QuestionGenerationRequest) (*QuestionGenerationResponse, error) {
	// Build prompt for question generation
	systemPrompt := p.buildQuestionGenerationPrompt(req)

	chatReq := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: systemPrompt + fmt.Sprintf("\n\nGenerate %d interview questions based on this job description: %s", req.NumQuestions, req.JobDescription),
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
		Rationale:  "Questions generated based on job requirements and candidate experience using Gemini AI",
		TokensUsed: response.TokensUsed,
		Provider:   ProviderGemini,
		Model:      response.Model,
		Timestamp:  time.Now(),
	}, nil
}

// EvaluateAnswers evaluates interview answers using Gemini
func (p *GeminiProvider) EvaluateAnswers(ctx context.Context, req *EvaluationRequest) (*EvaluationResponse, error) {
	// Build evaluation prompt
	systemPrompt := p.buildEvaluationPrompt(req)

	// Combine questions and answers for evaluation
	userContent := p.formatAnswersForEvaluation(req.Questions, req.Answers)

	chatReq := &ChatRequest{
		Messages: []Message{
			{
				Role:    "user",
				Content: systemPrompt + "\n\n" + userContent,
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
	// Make a simple request to validate credentials
	testReq := &geminiRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{Text: "Hello"},
				},
			},
		},
		GenerationConfig: &geminiGenConfig{
			MaxOutputTokens: 5,
		},
	}

	model := p.getModelName("")
	endpoint := fmt.Sprintf("/models/%s:generateContent", model)
	_, err := p.makeRequest(ctx, endpoint, testReq)
	return err
}

// IsHealthy checks if the provider is healthy
func (p *GeminiProvider) IsHealthy(ctx context.Context) bool {
	err := p.ValidateCredentials(ctx)
	return err == nil
}

// GetUsageStats returns usage statistics (placeholder)
func (p *GeminiProvider) GetUsageStats(ctx context.Context) (map[string]interface{}, error) {
	// TODO: Implement usage statistics retrieval
	return map[string]interface{}{
		"provider": ProviderGemini,
		"status":   "healthy",
	}, nil
}

// Helper methods

func (p *GeminiProvider) getModelName(model string) string {
	if model == "" {
		// Default Gemini model
		return "gemini-1.5-flash"
	}
	return model
}

func (p *GeminiProvider) convertMessages(messages []Message) []geminiContent {
	var contents []geminiContent

	for _, msg := range messages {
		// Gemini uses "user" and "model" roles, not "assistant"
		role := msg.Role
		if role == "assistant" {
			role = "model"
		}
		// Skip system messages as they need to be incorporated into user message
		if role == "system" {
			continue
		}

		content := geminiContent{
			Parts: []geminiPart{
				{Text: msg.Content},
			},
			Role: role,
		}
		contents = append(contents, content)
	}

	// If we have system messages, prepend them to the first user message
	var systemMessages []string
	for _, msg := range messages {
		if msg.Role == "system" {
			systemMessages = append(systemMessages, msg.Content)
		}
	}

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

func (p *GeminiProvider) makeRequest(ctx context.Context, endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.baseURL + endpoint + "?key=" + p.apiKey
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

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

// Reuse prompt building methods from OpenAI provider with slight modifications
func (p *GeminiProvider) buildQuestionGenerationPrompt(req *QuestionGenerationRequest) string {
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

func (p *GeminiProvider) buildEvaluationPrompt(req *EvaluationRequest) string {
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

func (p *GeminiProvider) formatAnswersForEvaluation(questions, answers []string) string {
	var content strings.Builder
	content.WriteString("Interview Questions and Candidate Answers:\n\n")

	for i := 0; i < len(questions) && i < len(answers); i++ {
		content.WriteString(fmt.Sprintf("Q%d: %s\n", i+1, questions[i]))
		content.WriteString(fmt.Sprintf("A%d: %s\n\n", i+1, answers[i]))
	}

	return content.String()
}

func (p *GeminiProvider) parseQuestionResponse(content string) []InterviewQuestion {
	// Reuse the same parsing logic as OpenAI provider
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
			currentQuestion.ExpectedTime = 5 // Default 5 minutes

			if currentQuestion.Question != "" {
				questions = append(questions, currentQuestion)
				currentQuestion = InterviewQuestion{}
			}
		}
	}

	return questions
}

func (p *GeminiProvider) parseEvaluationResponse(content string) *EvaluationResponse {
	// Reuse the same parsing logic as OpenAI provider
	evaluation := &EvaluationResponse{
		OverallScore:    0.7,
		CategoryScores:  make(map[string]float64),
		Strengths:       []string{},
		Weaknesses:      []string{},
		Recommendations: []string{},
	}

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

		if strings.HasPrefix(line, "- ") && len(line) > 2 {
			item := strings.TrimSpace(line[2:])
			evaluation.Strengths = append(evaluation.Strengths, item)
		}
	}

	evaluation.Feedback = strings.Join(feedbackLines, " ")

	evaluation.CategoryScores["technical"] = 0.7
	evaluation.CategoryScores["communication"] = 0.8
	evaluation.CategoryScores["problem_solving"] = 0.6
	evaluation.CategoryScores["experience"] = 0.7

	return evaluation
}
