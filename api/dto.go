package api

import "time"

// Data Transfer Objects (DTOs) for API request and response payloads:
// - CreateInterviewRequestDTO
// - InterviewResponseDTO
// - ListInterviewsResponseDTO
// - SubmitEvaluationRequestDTO
// - EvaluationResponseDTO
//
// These DTOs define the JSON structure for all RESTful API endpoints.
// Use these types for marshaling/unmarshaling and handler signatures.

// --- Interview DTOs ---
type CreateInterviewRequestDTO struct {
	CandidateName     string   `json:"candidate_name"`
	Questions         []string `json:"questions"`
	InterviewType     string   `json:"interview_type"`               // Required: "general", "technical", or "behavioral"
	InterviewLanguage string   `json:"interview_language,omitempty"` // Language preference: "en" or "zh-TW"
	JobDescription    string   `json:"job_description,omitempty"`    // Optional: Job description text
	// TODO: Resume file upload support will be added in future iteration
}

type InterviewResponseDTO struct {
	ID                string   `json:"id"`
	CandidateName     string   `json:"candidate_name"`
	Questions         []string `json:"questions"`
	InterviewType     string   `json:"interview_type"`            // "general", "technical", or "behavioral"
	InterviewLanguage string   `json:"interview_language"`        // Language preference: "en" or "zh-TW"
	JobDescription    string   `json:"job_description,omitempty"` // Optional: Job description text
	// TODO: Resume file support will be added in future iteration
	CreatedAt time.Time `json:"created_at"`
}

type ListInterviewsResponseDTO struct {
	Interviews []InterviewResponseDTO `json:"interviews"`
	// TODO: Add pagination support - Total field exists in frontend types but missing here
	Total int `json:"total"`
}

// --- Evaluation DTOs ---
type SubmitEvaluationRequestDTO struct {
	InterviewID string            `json:"interview_id"`
	Answers     map[string]string `json:"answers"`
}

type EvaluationResponseDTO struct {
	ID          string            `json:"id"`
	InterviewID string            `json:"interview_id"`
	Answers     map[string]string `json:"answers"` // TODO: Add answers field to match frontend expectations
	Score       float64           `json:"score"`
	Feedback    string            `json:"feedback"`
	CreatedAt   time.Time         `json:"created_at"`
}

// --- Chat DTOs ---
// TODO: Implement chat-based interview DTOs to support conversational interviews

type StartChatSessionRequestDTO struct {
	SessionLanguage string `json:"session_language,omitempty"` // Optional language override
}

type ChatMessageDTO struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "ai" or "user"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type ChatInterviewSessionDTO struct {
	ID              string           `json:"id"`
	InterviewID     string           `json:"interview_id"`
	SessionLanguage string           `json:"session_language"` // Session language: "en" or "zh-TW"
	Messages        []ChatMessageDTO `json:"messages"`
	Status          string           `json:"status"` // "active" or "completed"
	StartedAt       time.Time        `json:"started_at"`
	CreatedAt       time.Time        `json:"created_at"`
}

type SendMessageRequestDTO struct {
	Message string `json:"message"`
	Model   string `json:"model,omitempty"` // Optional: "openai/gpt-4o", "google/gemini-pro", defaults to configured provider
}

type SendMessageResponseDTO struct {
	Message       ChatMessageDTO  `json:"message"`
	AIResponse    *ChatMessageDTO `json:"ai_response,omitempty"`
	SessionStatus string          `json:"session_status"` // "active" or "completed"
}

// --- Error DTO ---
type ErrorResponseDTO struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}
