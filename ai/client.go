// Client for communicating with AI service/model
package ai

import (
	"context"
	"fmt"
)

// AIClient provides a high-level interface for AI operations
// All instances should be created through the AIClientFactory
type AIClient struct {
	enhancedClient *EnhancedAIClient
}

// GenerateChatResponse generates AI response for conversational interviews
func (c *AIClient) GenerateChatResponse(sessionID string, conversationHistory []map[string]string, userMessage string) (string, error) {
	return c.GenerateChatResponseWithLanguage(sessionID, conversationHistory, userMessage, "en")
}

// GenerateChatResponseWithLanguage generates AI response with language support
func (c *AIClient) GenerateChatResponseWithLanguage(sessionID string, conversationHistory []map[string]string, userMessage string, language string) (string, error) {
	// Build context for the AI including conversation history and language
	contextMap := map[string]interface{}{
		"interview_type":       "general",
		"job_title":            "Software Engineer",
		"context":              "Interview in progress",
		"conversation_history": conversationHistory,
		"language":             language,
	}

	return c.enhancedClient.GenerateInterviewResponse(sessionID, userMessage, contextMap)
}

// GenerateClosingMessage generates a closing AI response for ending interviews
func (c *AIClient) GenerateClosingMessage(sessionID string, conversationHistory []map[string]string, userMessage string) (string, error) {
	return c.GenerateClosingMessageWithLanguage(sessionID, conversationHistory, userMessage, "en")
}

// GenerateClosingMessageWithLanguage generates a closing AI response with language support
func (c *AIClient) GenerateClosingMessageWithLanguage(sessionID string, conversationHistory []map[string]string, userMessage string, language string) (string, error) {
	// Build context for the AI to indicate this is the final message
	contextMap := map[string]interface{}{
		"interview_type":       "general",
		"job_title":            "Software Engineer",
		"context":              "This is the final message - wrap up the interview professionally and thank the candidate",
		"conversation_history": conversationHistory,
		"closing_interview":    true,
		"language":             language,
	}

	return c.enhancedClient.GenerateInterviewResponse(sessionID, userMessage, contextMap)
}

// ShouldEndInterview determines if the interview should end
func (c *AIClient) ShouldEndInterview(messageCount int) bool {
	return messageCount >= 8 // End after 8 user messages
}

// EvaluateAnswers evaluates chat conversation and generates score and feedback
func (c *AIClient) EvaluateAnswers(questions []string, answers []string, language string) (float64, string, error) {
	// Use the context version with default job info
	return c.EvaluateAnswersWithContext(questions, answers, "General interview evaluation", language)
}

// EvaluateAnswersWithContext evaluates chat conversation with interview context
func (c *AIClient) EvaluateAnswersWithContext(questions []string, answers []string, jobDesc, language string) (float64, string, error) {
	if len(answers) == 0 {
		return 0.0, "No answers provided.", nil
	}

	// Use the enhanced AI client for real evaluation with context
	ctx := context.Background()

	// Create evaluation request with proper context including language
	req := &EvaluationRequest{
		Questions:   questions,
		Answers:     answers,
		JobDesc:     jobDesc,
		Criteria:    []string{"communication", "technical_knowledge", "problem_solving", "clarity", "cultural_fit"},
		DetailLevel: "detailed",
		Language:    language, // Pass language for evaluation
		Context: map[string]interface{}{
			"interview_type":  "conversational",
			"evaluation_type": "chat_based",
			"language":        language, // Also include in context map
		},
	}

	// Call enhanced client for evaluation
	resp, err := c.enhancedClient.EvaluateAnswers(ctx, req)
	if err != nil {
		return 0.0, "Evaluation failed", err
	}

	return resp.OverallScore, resp.Feedback, nil
}

// GenerateQuestionsFromResume generates interview questions based on resume and job description
func (c *AIClient) GenerateQuestionsFromResume(resumeText, jobDescription string) ([]InterviewQuestion, error) {
	ctx := context.Background()

	req := &QuestionGenerationRequest{
		JobDescription:  jobDescription,
		ResumeContent:   resumeText,
		InterviewType:   "mixed",
		NumQuestions:    8,
		ExperienceLevel: "mid",
		Difficulty:      "medium",
	}

	resp, err := c.enhancedClient.GenerateQuestions(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Questions, nil
}

// GenerateInterviewQuestions generates questions for a specific interview setup
func (c *AIClient) GenerateInterviewQuestions(jobDesc string, questionCount int) ([]InterviewQuestion, error) {
	ctx := context.Background()

	req := &QuestionGenerationRequest{
		JobDescription:  jobDesc,
		InterviewType:   "general",
		NumQuestions:    questionCount,
		ExperienceLevel: "mid",
		Difficulty:      "medium",
	}

	resp, err := c.enhancedClient.GenerateQuestions(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Questions, nil
}

// GetProviderInfo returns information about available AI providers
func (c *AIClient) GetProviderInfo() map[string]interface{} {
	info := make(map[string]interface{})
	providers := c.enhancedClient.GetAvailableProviders()

	for _, providerName := range providers {
		info[providerName] = GetProviderInfo(providerName)
	}

	return info
}

// SwitchProvider changes the active AI provider
func (c *AIClient) SwitchProvider(providerName string) error {
	c.enhancedClient.mu.Lock()
	defer c.enhancedClient.mu.Unlock()

	if _, exists := c.enhancedClient.providers[providerName]; !exists {
		return fmt.Errorf("provider not available: %s", providerName)
	}

	c.enhancedClient.config.DefaultProvider = providerName
	return nil
}

// GetCurrentProvider returns the currently configured AI provider
func (c *AIClient) GetCurrentProvider() string {
	return c.enhancedClient.config.DefaultProvider
}

// GetCurrentModel returns the currently configured AI model
func (c *AIClient) GetCurrentModel() string {
	return c.enhancedClient.config.DefaultModel
}
