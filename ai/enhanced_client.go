// Enhanced AI client with support for multiple providers
package ai

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zidane0000/ai-interview-platform/utils"
)

// EnhancedAIClient provides a unified interface to multiple AI providers
type EnhancedAIClient struct {
	config    *AIConfig
	providers map[string]AIProvider
	metrics   *AIMetrics
	cache     *ResponseCache
	mu        sync.RWMutex
}

// AIMetrics tracks usage and performance metrics
type AIMetrics struct {
	TotalRequests   int64                     `json:"total_requests"`
	SuccessfulReqs  int64                     `json:"successful_requests"`
	FailedRequests  int64                     `json:"failed_requests"`
	TotalTokensUsed int64                     `json:"total_tokens_used"`
	TotalCost       float64                   `json:"total_cost"`
	AvgResponseTime time.Duration             `json:"avg_response_time"`
	LastRequestTime time.Time                 `json:"last_request_time"`
	ProviderStats   map[string]*ProviderStats `json:"provider_stats"`
	mu              sync.RWMutex
}

// ProviderStats tracks metrics per provider
type ProviderStats struct {
	Requests   int64         `json:"requests"`
	Successes  int64         `json:"successes"`
	Failures   int64         `json:"failures"`
	TokensUsed int64         `json:"tokens_used"`
	Cost       float64       `json:"cost"`
	AvgLatency time.Duration `json:"avg_latency"`
	LastUsed   time.Time     `json:"last_used"`
}

// ResponseCache provides caching for AI responses
type ResponseCache struct {
	cache map[string]*CacheEntry
	mu    sync.RWMutex
}

// CacheEntry represents a cached response
type CacheEntry struct {
	Response  *ChatResponse `json:"response"`
	ExpiresAt time.Time     `json:"expires_at"`
	HitCount  int           `json:"hit_count"`
}

// NewEnhancedAIClient creates a new enhanced AI client
func NewEnhancedAIClient(config *AIConfig) *EnhancedAIClient {
	client := &EnhancedAIClient{
		config:    config,
		providers: make(map[string]AIProvider),
		metrics: &AIMetrics{
			ProviderStats: make(map[string]*ProviderStats),
		},
		cache: &ResponseCache{
			cache: make(map[string]*CacheEntry),
		},
	}

	// Initialize providers based on configuration
	if config.OpenAIAPIKey != "" {
		client.registerProvider(ProviderOpenAI, NewOpenAIProvider(config.OpenAIAPIKey, config))
	}
	if config.GeminiAPIKey != "" {
		client.registerProvider(ProviderGemini, NewGeminiProvider(config.GeminiAPIKey, config))
	}
	// Always register mock provider for fallback/testing
	client.registerProvider("mock", NewMockProvider())

	return client
}

// registerProvider registers a new AI provider
func (c *EnhancedAIClient) registerProvider(name string, provider AIProvider) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.providers[name] = provider
	c.metrics.ProviderStats[name] = &ProviderStats{}
}

// GetProvider returns the specified provider or default
func (c *EnhancedAIClient) GetProvider(providerName string) (AIProvider, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if providerName == "" {
		providerName = c.config.DefaultProvider
	}

	provider, exists := c.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider %s not found or not configured", providerName)
	}

	return provider, nil
}

// GenerateInterviewResponse generates an AI response for interview conversation
func (c *EnhancedAIClient) GenerateInterviewResponse(sessionID, userMessage string, contextMap map[string]interface{}) (string, error) {
	ctx := context.Background()
	if ctxVal, ok := contextMap["ctx"]; ok {
		if ctxTyped, ok := ctxVal.(context.Context); ok {
			ctx = ctxTyped
		}
	}

	// Build interview-specific prompt
	systemPrompt := c.buildInterviewSystemPrompt(contextMap)

	// Start with system message
	messages := []Message{
		{
			Role:      "system",
			Content:   systemPrompt,
			Timestamp: time.Now(),
		},
	}
	// Add conversation history if available
	if historyVal, exists := contextMap["conversation_history"]; exists {
		if history, ok := historyVal.([]map[string]string); ok {
			// Add conversation history with proper roles
			for _, msg := range history {
				role := msg["role"]
				content := msg["content"]

				// Convert message type to proper role for AI
				if role == "user" {
					role = "user"
				} else if role == "ai" {
					role = "assistant"
				}

				messages = append(messages, Message{
					Role:      role,
					Content:   content,
					Timestamp: time.Now(),
				})
			}
		}
	}

	// Add current user message
	messages = append(messages, Message{
		Role:      "user",
		Content:   userMessage,
		Timestamp: time.Now(),
	})

	// Create chat request
	req := &ChatRequest{
		Messages:    messages,
		Model:       c.config.DefaultModel,
		MaxTokens:   c.config.DefaultMaxTokens,
		Temperature: c.config.DefaultTemp,
		SessionID:   sessionID,
		Context:     contextMap,
	}

	// Generate response
	response, err := c.GenerateResponse(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to generate interview response: %w", err)
	}

	return response.Content, nil
}

// GenerateResponse generates a response using the configured provider
func (c *EnhancedAIClient) GenerateResponse(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	startTime := time.Now()

	// Check cache first if enabled
	if c.config.EnableCaching {
		if cached := c.getCachedResponse(req); cached != nil {
			c.updateMetrics("cache_hit", startTime, nil, 0)
			return cached, nil
		}
	}

	// Get provider
	var providerName string
	if v, ok := req.Context["provider"]; ok {
		if s, ok := v.(string); ok && s != "" {
			providerName = s
		}
	}
	provider, err := c.GetProvider(providerName)
	if err != nil {
		// Fallback to default provider
		provider, err = c.GetProvider("")
		if err != nil {
			return nil, fmt.Errorf("no available AI provider: %w", err)
		}
	}

	// Set defaults
	if req.MaxTokens == 0 {
		req.MaxTokens = c.config.DefaultMaxTokens
	}
	if req.Temperature == 0 {
		req.Temperature = c.config.DefaultTemp
	}
	if req.Model == "" {
		req.Model = c.config.DefaultModel
	}

	// Generate response with retries
	var response *ChatResponse
	var lastErr error

	for i := 0; i <= c.config.MaxRetries; i++ {
		response, lastErr = provider.GenerateResponse(ctx, req)
		if lastErr == nil {
			break
		}
		if i < c.config.MaxRetries {
			// Exponential backoff with overflow protection
			// Use safe integer arithmetic to avoid gosec warnings
			backoffSeconds := 1
			for shift := 0; shift < i && shift < 10; shift++ {
				backoffSeconds *= 2 // 2^shift, capped at 2^10 = 1024 seconds
			}
			backoffDuration := time.Duration(backoffSeconds) * time.Second
			utils.Errorf("AI request failed (attempt %d/%d), retrying in %v: %v",
				i+1, c.config.MaxRetries+1, backoffDuration, lastErr)
			time.Sleep(backoffDuration)
		}
	}

	if lastErr != nil {
		c.updateMetrics("error", startTime, lastErr, 0)
		return nil, fmt.Errorf("AI request failed after %d retries: %w", c.config.MaxRetries, lastErr)
	}

	// Update metrics
	c.updateMetrics("success", startTime, nil, response.TokensUsed.TotalTokens)

	// Cache response if enabled
	if c.config.EnableCaching {
		c.cacheResponse(req, response)
	}

	return response, nil
}

// GenerateQuestions generates interview questions using AI
func (c *EnhancedAIClient) GenerateQuestions(ctx context.Context, req *QuestionGenerationRequest) (*QuestionGenerationResponse, error) {
	provider, err := c.GetProvider("")
	if err != nil {
		return nil, fmt.Errorf("no available AI provider for question generation: %w", err)
	}

	return provider.GenerateInterviewQuestions(ctx, req)
}

// EvaluateAnswers evaluates interview answers using AI
func (c *EnhancedAIClient) EvaluateAnswers(ctx context.Context, req *EvaluationRequest) (*EvaluationResponse, error) {
	provider, err := c.GetProvider("")
	if err != nil {
		return nil, fmt.Errorf("no available AI provider for evaluation: %w", err)
	}

	return provider.EvaluateAnswers(ctx, req)
}

// buildInterviewSystemPrompt creates a system prompt for interview context
func (c *EnhancedAIClient) buildInterviewSystemPrompt(context map[string]interface{}) string {
	jobDescription := getStringFromContext(context, "job_description", "")
	interviewType := getStringFromContext(context, "interview_type", "general")
	language := getStringFromContext(context, "language", "en")

	// Build language-specific instructions
	var languageInstructions string
	switch language {
	case "zh-TW":
		languageInstructions = `CRITICAL LANGUAGE REQUIREMENT: 
- You MUST respond ONLY in Traditional Chinese (繁體中文)
- Do NOT use English in your responses
- Use Traditional Chinese characters for ALL communication
- All questions, acknowledgments, and follow-ups must be in Traditional Chinese
- This is a Traditional Chinese interview - maintain language consistency`
	case "en":
		languageInstructions = "IMPORTANT: You must respond ONLY in English."
	default:
		languageInstructions = "IMPORTANT: You must respond ONLY in English."
	}

	// Build job context from description
	var jobContext string
	if jobDescription != "" {
		jobContext = fmt.Sprintf("Job Description: %s", jobDescription)
	} else {
		jobContext = "This is a general interview assessment"
	}

	basePrompt := fmt.Sprintf(`%sYou are an experienced interviewer conducting a %s interview.

%s

Your role:
- Ask thoughtful, relevant questions that assess the candidate's skills and experience
- Provide a professional and friendly interview experience
- Ask follow-up questions based on the candidate's responses
- Keep questions focused and clear
- Maintain a conversational but professional tone

Guidelines:
- Ask one question at a time
- Wait for the candidate's response before asking the next question
- Provide brief acknowledgments of good answers
- Ask follow-up questions to dive deeper into interesting topics
- Keep the conversation flowing naturally

Remember: You are evaluating the candidate's technical skills, problem-solving ability, and cultural fit.

%s`, languageInstructions, interviewType, jobContext, languageInstructions)

	return basePrompt
}

// Helper function to get string from context map
func getStringFromContext(context map[string]interface{}, key, defaultValue string) string {
	if val, exists := context[key]; exists {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// getCachedResponse retrieves a cached response if available and not expired
func (c *EnhancedAIClient) getCachedResponse(req *ChatRequest) *ChatResponse {
	if !c.config.EnableCaching {
		return nil
	}

	cacheKey := c.generateCacheKey(req)

	c.cache.mu.RLock()
	defer c.cache.mu.RUnlock()

	entry, exists := c.cache.cache[cacheKey]
	if !exists || time.Now().After(entry.ExpiresAt) {
		return nil
	}

	entry.HitCount++
	return entry.Response
}

// cacheResponse stores a response in the cache
func (c *EnhancedAIClient) cacheResponse(req *ChatRequest, response *ChatResponse) {
	if !c.config.EnableCaching {
		return
	}

	cacheKey := c.generateCacheKey(req)
	expiresAt := time.Now().Add(1 * time.Hour) // Cache for 1 hour

	c.cache.mu.Lock()
	defer c.cache.mu.Unlock()

	c.cache.cache[cacheKey] = &CacheEntry{
		Response:  response,
		ExpiresAt: expiresAt,
		HitCount:  0,
	}
}

// generateCacheKey creates a cache key for the request
func (c *EnhancedAIClient) generateCacheKey(req *ChatRequest) string {
	// Build a comprehensive cache key that includes:
	// 1. Model name
	// 2. Language context
	// 3. System prompt content (first 100 chars to avoid overly long keys)
	// 4. Conversation history length

	var keyParts []string
	keyParts = append(keyParts, req.Model)

	// Include language from context if available
	if languageVal, exists := req.Context["language"]; exists {
		if language, ok := languageVal.(string); ok {
			keyParts = append(keyParts, "lang:"+language)
		}
	}

	// Include system prompt (truncated for key length)
	if len(req.Messages) > 0 {
		systemMessage := req.Messages[0]
		if systemMessage.Role == "system" {
			// Use first 100 characters of system prompt to differentiate languages
			promptPreview := systemMessage.Content
			if len(promptPreview) > 100 {
				promptPreview = promptPreview[:100]
			}
			keyParts = append(keyParts, "system:"+promptPreview)
		}

		// Include conversation length
		keyParts = append(keyParts, fmt.Sprintf("len:%d", len(req.Messages)))
	}

	// Join all parts with colons
	cacheKey := strings.Join(keyParts, ":")

	// Replace problematic characters that might break cache keys
	cacheKey = strings.ReplaceAll(cacheKey, "\n", "\\n")
	cacheKey = strings.ReplaceAll(cacheKey, "\r", "\\r")

	return cacheKey
}

// updateMetrics updates client metrics
func (c *EnhancedAIClient) updateMetrics(eventType string, startTime time.Time, err error, tokensUsed int) {
	if !c.config.EnableMetrics {
		return
	}

	duration := time.Since(startTime)

	c.metrics.mu.Lock()
	defer c.metrics.mu.Unlock()

	c.metrics.TotalRequests++
	c.metrics.LastRequestTime = time.Now()

	if err != nil {
		c.metrics.FailedRequests++
	} else {
		c.metrics.SuccessfulReqs++
		c.metrics.TotalTokensUsed += int64(tokensUsed)
		c.metrics.TotalCost += float64(tokensUsed) * c.config.CostPerToken
	}

	// Update average response time
	if c.metrics.SuccessfulReqs > 0 {
		totalTime := time.Duration(c.metrics.SuccessfulReqs-1)*c.metrics.AvgResponseTime + duration
		c.metrics.AvgResponseTime = totalTime / time.Duration(c.metrics.SuccessfulReqs)
	}
}

// GetMetrics returns current client metrics
func (c *EnhancedAIClient) GetMetrics() *AIMetrics {
	c.metrics.mu.RLock()
	defer c.metrics.mu.RUnlock()

	// Return a copy to avoid race conditions
	return &AIMetrics{
		TotalRequests:   c.metrics.TotalRequests,
		SuccessfulReqs:  c.metrics.SuccessfulReqs,
		FailedRequests:  c.metrics.FailedRequests,
		TotalTokensUsed: c.metrics.TotalTokensUsed,
		TotalCost:       c.metrics.TotalCost,
		AvgResponseTime: c.metrics.AvgResponseTime,
		LastRequestTime: c.metrics.LastRequestTime,
		ProviderStats:   make(map[string]*ProviderStats),
	}
}

// IsHealthy checks if the client and providers are healthy
func (c *EnhancedAIClient) IsHealthy(ctx context.Context) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, provider := range c.providers {
		if provider.IsHealthy(ctx) {
			return true // At least one provider is healthy
		}
	}

	return false
}

// GetAvailableProviders returns list of available providers
func (c *EnhancedAIClient) GetAvailableProviders() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	providers := make([]string, 0, len(c.providers))
	for name := range c.providers {
		providers = append(providers, name)
	}

	return providers
}
