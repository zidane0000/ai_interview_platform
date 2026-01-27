// Base provider with shared logic for all AI providers
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

// ProviderAdapter defines provider-specific behavior that each provider must implement
type ProviderAdapter interface {
	// SetAuth sets provider-specific authentication on the HTTP request
	SetAuth(req *http.Request)

	// GetEndpointURL returns the full URL for the given endpoint
	GetEndpointURL(endpoint string) string
}

// BaseProvider contains shared logic and configuration for all AI providers
type BaseProvider struct {
	config     *AIConfig
	httpClient *http.Client
	baseURL    string
}

// NewBaseProvider creates a new BaseProvider with the given configuration
func NewBaseProvider(config *AIConfig, baseURL string, timeout time.Duration) BaseProvider {
	return BaseProvider{
		config:  config,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// MakeRequest performs an HTTP request with provider-specific authentication
func (b *BaseProvider) MakeRequest(ctx context.Context, adapter ProviderAdapter, endpoint string, payload interface{}) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := adapter.GetEndpointURL(endpoint)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	adapter.SetAuth(req)

	resp, err := b.httpClient.Do(req)
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

// GetModelName returns the model name, using the default if not specified
func (b *BaseProvider) GetModelName(model, defaultModel string) string {
	if model == "" {
		if defaultModel != "" {
			return defaultModel
		}
		return b.config.DefaultModel
	}
	return model
}

// --- Shared Prompt Builders ---

// BuildQuestionGenerationPrompt creates the prompt for generating interview questions
func BuildQuestionGenerationPrompt(req *QuestionGenerationRequest) string {
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

// BuildEvaluationPrompt creates the prompt for evaluating interview answers
func BuildEvaluationPrompt(req *EvaluationRequest) string {
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

// FormatAnswersForEvaluation formats questions and answers for evaluation
func FormatAnswersForEvaluation(questions, answers []string) string {
	var content strings.Builder
	content.WriteString("Interview Questions and Candidate Answers:\n\n")

	for i := 0; i < len(questions) && i < len(answers); i++ {
		content.WriteString(fmt.Sprintf("Q%d: %s\n", i+1, questions[i]))
		content.WriteString(fmt.Sprintf("A%d: %s\n\n", i+1, answers[i]))
	}

	return content.String()
}

// --- Shared Response Parsers ---

// ParseQuestionResponse parses the AI response to extract interview questions
func ParseQuestionResponse(content string) []InterviewQuestion {
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

// ParseEvaluationResponse parses the AI response to extract evaluation data
func ParseEvaluationResponse(content string) *EvaluationResponse {
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
	currentSection := "" // Track which section we're in

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Handle section headers
		if strings.HasPrefix(line, "Feedback:") {
			inFeedback = true
			currentSection = ""
			feedbackText := strings.TrimSpace(line[9:])
			if feedbackText != "" {
				feedbackLines = append(feedbackLines, feedbackText)
			}
			continue
		}
		if strings.HasPrefix(line, "Strengths:") {
			inFeedback = false
			currentSection = "strengths"
			continue
		}
		if strings.HasPrefix(line, "Areas for Improvement:") {
			inFeedback = false
			currentSection = "weaknesses"
			continue
		}
		if strings.HasPrefix(line, "Recommendations:") {
			inFeedback = false
			currentSection = "recommendations"
			continue
		}

		// Handle feedback content
		if inFeedback && line != "" {
			feedbackLines = append(feedbackLines, line)
		}

		// Handle bullet points based on current section
		if strings.HasPrefix(line, "- ") && len(line) > 2 {
			item := strings.TrimSpace(line[2:])
			switch currentSection {
			case "strengths":
				evaluation.Strengths = append(evaluation.Strengths, item)
			case "weaknesses":
				evaluation.Weaknesses = append(evaluation.Weaknesses, item)
			case "recommendations":
				evaluation.Recommendations = append(evaluation.Recommendations, item)
			}
		}
	}

	evaluation.Feedback = strings.Join(feedbackLines, " ")

	// Set default category scores
	evaluation.CategoryScores["technical"] = 0.7
	evaluation.CategoryScores["communication"] = 0.8
	evaluation.CategoryScores["problem_solving"] = 0.6
	evaluation.CategoryScores["experience"] = 0.7

	return evaluation
}
