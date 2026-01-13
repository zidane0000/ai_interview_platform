// AI types and interfaces for LLM integration
package ai

import (
	"context"
	"time"
)

// Provider constants
const (
	ProviderOpenAI = "openai"
	ProviderGemini = "gemini"
	ProviderMock   = "mock"
)

// Message represents a chat message in the conversation
type Message struct {
	Role      string                 `json:"role"`      // "system", "user", "assistant"
	Content   string                 `json:"content"`   // Message content
	Metadata  map[string]interface{} `json:"metadata"`  // Additional metadata
	Timestamp time.Time              `json:"timestamp"` // When the message was created
}

// ChatRequest represents a request to generate a chat response
type ChatRequest struct {
	Messages     []Message              `json:"messages"`      // Conversation history
	Model        string                 `json:"model"`         // Model to use
	MaxTokens    int                    `json:"max_tokens"`    // Maximum tokens in response
	Temperature  float64                `json:"temperature"`   // Randomness (0.0-1.0)
	TopP         float64                `json:"top_p"`         // Nucleus sampling
	Stream       bool                   `json:"stream"`        // Whether to stream response
	SystemPrompt string                 `json:"system_prompt"` // System instruction
	Context      map[string]interface{} `json:"context"`       // Additional context
	SessionID    string                 `json:"session_id"`    // Session identifier
}

// ChatResponse represents a response from the AI
type ChatResponse struct {
	Content      string                 `json:"content"`       // Generated content
	FinishReason string                 `json:"finish_reason"` // Why generation stopped
	TokensUsed   TokenUsage             `json:"tokens_used"`   // Token consumption
	Model        string                 `json:"model"`         // Model used
	Provider     string                 `json:"provider"`      // Provider used
	Metadata     map[string]interface{} `json:"metadata"`      // Additional response data
	ResponseTime time.Duration          `json:"response_time"` // Time taken to generate
	Timestamp    time.Time              `json:"timestamp"`     // When response was generated
}

// TokenUsage represents token consumption metrics
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`     // Tokens in input
	CompletionTokens int `json:"completion_tokens"` // Tokens in output
	TotalTokens      int `json:"total_tokens"`      // Total tokens used
}

// EvaluationRequest represents a request to evaluate interview answers
type EvaluationRequest struct {
	Questions   []string               `json:"questions"`    // Interview questions
	Answers     []string               `json:"answers"`      // Candidate answers
	JobDesc     string                 `json:"job_desc"`     // Job description (AI will extract job title from this)
	Criteria    []string               `json:"criteria"`     // Evaluation criteria
	Context     map[string]interface{} `json:"context"`      // Additional context
	DetailLevel string                 `json:"detail_level"` // "brief", "detailed", "comprehensive"
	Language    string                 `json:"language"`     // Language for evaluation ("en", "zh-TW")
}

// EvaluationResponse represents an AI evaluation result
type EvaluationResponse struct {
	OverallScore    float64            `json:"overall_score"`   // 0.0-1.0
	CategoryScores  map[string]float64 `json:"category_scores"` // Scores by category
	Feedback        string             `json:"feedback"`        // General feedback
	Strengths       []string           `json:"strengths"`       // Identified strengths
	Weaknesses      []string           `json:"weaknesses"`      // Areas for improvement
	Recommendations []string           `json:"recommendations"` // Specific recommendations
	TokensUsed      TokenUsage         `json:"tokens_used"`     // Token consumption
	Provider        string             `json:"provider"`        // Provider used
	Model           string             `json:"model"`           // Model used
	Timestamp       time.Time          `json:"timestamp"`       // When evaluation was done
}

// QuestionGenerationRequest represents a request to generate interview questions
type QuestionGenerationRequest struct {
	JobDescription  string                 `json:"job_description"`  // Job requirements (AI will extract job title from this)
	ResumeContent   string                 `json:"resume_content"`   // Candidate's resume
	ExperienceLevel string                 `json:"experience_level"` // "junior", "mid", "senior"
	InterviewType   string                 `json:"interview_type"`   // "technical", "behavioral", "mixed"
	NumQuestions    int                    `json:"num_questions"`    // Number of questions to generate
	Difficulty      string                 `json:"difficulty"`       // "easy", "medium", "hard"
	Context         map[string]interface{} `json:"context"`          // Additional context
}

// QuestionGenerationResponse represents generated interview questions
type QuestionGenerationResponse struct {
	Questions  []InterviewQuestion `json:"questions"`   // Generated questions
	Rationale  string              `json:"rationale"`   // Why these questions were chosen
	TokensUsed TokenUsage          `json:"tokens_used"` // Token consumption
	Provider   string              `json:"provider"`    // Provider used
	Model      string              `json:"model"`       // Model used
	Timestamp  time.Time           `json:"timestamp"`   // When questions were generated
}

// InterviewQuestion represents a single interview question with metadata
type InterviewQuestion struct {
	Question     string   `json:"question"`      // The question text
	Category     string   `json:"category"`      // Question category
	Difficulty   string   `json:"difficulty"`    // Difficulty level
	ExpectedTime int      `json:"expected_time"` // Expected answer time in minutes
	Keywords     []string `json:"keywords"`      // Key concepts to look for
	FollowUp     []string `json:"follow_up"`     // Potential follow-up questions
}

// AIProvider interface defines the contract for AI providers
type AIProvider interface {
	// Basic chat completion
	GenerateResponse(ctx context.Context, req *ChatRequest) (*ChatResponse, error)

	// Streaming chat completion
	GenerateStreamResponse(ctx context.Context, req *ChatRequest) (<-chan *ChatResponse, error)

	// Interview-specific methods
	GenerateInterviewQuestions(ctx context.Context, req *QuestionGenerationRequest) (*QuestionGenerationResponse, error)
	EvaluateAnswers(ctx context.Context, req *EvaluationRequest) (*EvaluationResponse, error)

	// Provider info
	GetProviderName() string
	GetSupportedModels() []string
	ValidateCredentials(ctx context.Context) error

	// Health and monitoring
	IsHealthy(ctx context.Context) bool
	GetUsageStats(ctx context.Context) (map[string]interface{}, error)
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	Content      string    `json:"content"`       // Partial content
	Delta        string    `json:"delta"`         // New content since last chunk
	IsComplete   bool      `json:"is_complete"`   // Whether this is the final chunk
	FinishReason string    `json:"finish_reason"` // Reason for completion (if complete)
	TokensUsed   int       `json:"tokens_used"`   // Tokens used so far
	Timestamp    time.Time `json:"timestamp"`     // Chunk timestamp
}

// AIConfig represents configuration for AI providers
type AIConfig struct {
	// API Keys
	OpenAIAPIKey string `json:"openai_api_key"`
	GeminiAPIKey string `json:"gemini_api_key"`

	// Custom endpoints (for OpenAI-compatible providers)
	OpenAIBaseURL string `json:"openai_base_url,omitempty"` // e.g., "https://api.together.ai/v1"
	GeminiBaseURL string `json:"gemini_base_url,omitempty"` // Custom Gemini endpoint

	// Provider settings
	DefaultProvider string `json:"default_provider"`
	DefaultModel    string `json:"default_model"`

	// Request settings
	MaxRetries       int           `json:"max_retries"`
	RequestTimeout   time.Duration `json:"request_timeout"`
	DefaultMaxTokens int           `json:"default_max_tokens"`
	DefaultTemp      float64       `json:"default_temperature"`

	// Feature flags
	EnableCaching   bool `json:"enable_caching"`
	EnableMetrics   bool `json:"enable_metrics"`
	EnableStreaming bool `json:"enable_streaming"`

	// Rate limiting
	RateLimitRPM int `json:"rate_limit_rpm"` // Requests per minute
	RateLimitTPM int `json:"rate_limit_tpm"` // Tokens per minute

	// Costs and quotas
	DailyTokenLimit int     `json:"daily_token_limit"`
	CostPerToken    float64 `json:"cost_per_token"`
	MaxCostPerDay   float64 `json:"max_cost_per_day"`
}

// InterviewContext contains context for interview-related AI operations
type InterviewContext struct {
	JobDescription  string            `json:"job_description"` // Job description (AI will extract job title from this)
	CandidateName   string            `json:"candidate_name"`
	InterviewType   string            `json:"interview_type"`
	ExperienceLevel string            `json:"experience_level"`
	InterviewStage  string            `json:"interview_stage"`  // "introduction", "questions", "conclusion"
	CurrentQuestion int               `json:"current_question"` // Current question number
	TotalQuestions  int               `json:"total_questions"`  // Total expected questions
	TimeElapsed     time.Duration     `json:"time_elapsed"`     // Time since interview start
	CustomContext   map[string]string `json:"custom_context"`   // Additional custom context
}

// PromptTemplate represents a reusable prompt template
type PromptTemplate struct {
	Name        string            `json:"name"`
	Template    string            `json:"template"`
	Variables   []string          `json:"variables"`
	Category    string            `json:"category"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
}
