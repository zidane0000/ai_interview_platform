package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/zidane0000/ai-interview-platform/api"
)

// Use API DTOs directly instead of mirroring them
type InterviewResponseDTO = api.InterviewResponseDTO
type ChatMessageDTO = api.ChatMessageDTO
type ChatInterviewSessionDTO = api.ChatInterviewSessionDTO
type SendMessageRequestDTO = api.SendMessageRequestDTO
type SendMessageResponseDTO = api.SendMessageResponseDTO
type EvaluationResponseDTO = api.EvaluationResponseDTO

// Test helper functions for E2E tests

func GetAPIBaseURL() string {
	if v := os.Getenv("API_BASE_URL"); v != "" {
		return v
	}
	return "http://localhost:8080/api"
}

// CreateTestInterview creates a test interview and returns the response
func CreateTestInterview(t *testing.T, candidateName string, questions []string) InterviewResponseDTO {
	return CreateTestInterviewWithTypeAndLanguage(t, candidateName, questions, "general", "")
}

// CreateTestInterviewWithLanguage creates a test interview with specified language
func CreateTestInterviewWithLanguage(t *testing.T, candidateName string, questions []string, language string) InterviewResponseDTO {
	return CreateTestInterviewWithTypeAndLanguage(t, candidateName, questions, "general", language)
}

// CreateTestInterviewWithType creates a test interview with specified type
func CreateTestInterviewWithType(t *testing.T, candidateName string, questions []string, interviewType string) InterviewResponseDTO {
	return CreateTestInterviewWithTypeAndLanguage(t, candidateName, questions, interviewType, "")
}

// CreateTestInterviewWithTypeAndLanguage creates a test interview with specified type and language
func CreateTestInterviewWithTypeAndLanguage(t *testing.T, candidateName string, questions []string, interviewType string, language string) InterviewResponseDTO {
	return CreateTestInterviewWithFullDetails(t, candidateName, questions, interviewType, language, "")
}

// CreateTestInterviewWithFullDetails creates a test interview with all optional fields
func CreateTestInterviewWithFullDetails(t *testing.T, candidateName string, questions []string, interviewType string, language string, jobDescription string) InterviewResponseDTO {
	t.Helper()
	baseURL := GetAPIBaseURL()

	createReq := map[string]interface{}{
		"candidate_name": candidateName,
		"questions":      questions,
		"interview_type": interviewType,
	}

	// Add optional fields if specified
	if language != "" {
		createReq["interview_language"] = language
	}
	if jobDescription != "" {
		createReq["job_description"] = jobDescription
	}
	// TODO: Resume file support will be added later

	reqBody, _ := json.Marshal(createReq)
	resp, err := http.Post(baseURL+"/interviews", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create interview: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	var interview InterviewResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&interview); err != nil {
		t.Fatalf("Failed to decode interview response: %v", err)
	}
	if interview.ID == "" {
		t.Fatalf("Interview ID is empty")
	}

	return interview
}

// CreateTestInterviewWithJobDescription creates a test interview with job description but no resume
func CreateTestInterviewWithJobDescription(t *testing.T, candidateName string, questions []string, interviewType string, language string, jobDescription string) InterviewResponseDTO {
	t.Helper()
	baseURL := GetAPIBaseURL()

	createReq := map[string]interface{}{
		"candidate_name": candidateName,
		"questions":      questions,
		"interview_type": interviewType,
	}

	// Add optional fields if specified
	if language != "" {
		createReq["interview_language"] = language
	}
	if jobDescription != "" {
		createReq["job_description"] = jobDescription
	}

	reqBody, _ := json.Marshal(createReq)
	resp, err := http.Post(baseURL+"/interviews", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to create interview: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	var interview InterviewResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&interview); err != nil {
		t.Fatalf("Failed to decode interview response: %v", err)
	}
	if interview.ID == "" {
		t.Fatalf("Interview ID is empty")
	}

	return interview
}

// StartChatSession starts a chat session for the given interview
func StartChatSession(t *testing.T, interviewID string) ChatInterviewSessionDTO {
	t.Helper()
	baseURL := GetAPIBaseURL()

	resp, err := http.Post(fmt.Sprintf("%s/interviews/%s/chat/start", baseURL, interviewID), "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to start chat session: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	var chatSession ChatInterviewSessionDTO
	if err := json.NewDecoder(resp.Body).Decode(&chatSession); err != nil {
		t.Fatalf("Failed to decode chat session: %v", err)
	}

	return chatSession
}

// SendMessage sends a message in a chat session
func SendMessage(t *testing.T, sessionID, message string) SendMessageResponseDTO {
	t.Helper()
	baseURL := GetAPIBaseURL()

	sendMsgReq := SendMessageRequestDTO{
		Message: message,
		// Note: InterviewID is not needed for /chat/{sessionID}/message endpoint
		// as the session already knows which interview it belongs to
	}
	reqBody, _ := json.Marshal(sendMsgReq)
	resp, err := http.Post(fmt.Sprintf("%s/chat/%s/message", baseURL, sessionID), "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var msgResponse SendMessageResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&msgResponse); err != nil {
		t.Fatalf("Failed to decode send message response: %v", err)
	}

	return msgResponse
}

// GetChatSession retrieves chat session state
func GetChatSession(t *testing.T, sessionID string) ChatInterviewSessionDTO {
	t.Helper()
	baseURL := GetAPIBaseURL()

	resp, err := http.Get(fmt.Sprintf("%s/chat/%s", baseURL, sessionID))
	if err != nil {
		t.Fatalf("Failed to get chat session: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var session ChatInterviewSessionDTO
	if err := json.NewDecoder(resp.Body).Decode(&session); err != nil {
		t.Fatalf("Failed to decode chat session: %v", err)
	}

	return session
}

// EndChatSession ends a chat session and returns evaluation
func EndChatSession(t *testing.T, sessionID string) EvaluationResponseDTO {
	t.Helper()
	baseURL := GetAPIBaseURL()

	resp, err := http.Post(fmt.Sprintf("%s/chat/%s/end", baseURL, sessionID), "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to end chat session: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var evaluation EvaluationResponseDTO
	if err := json.NewDecoder(resp.Body).Decode(&evaluation); err != nil {
		t.Fatalf("Failed to decode evaluation: %v", err)
	}

	return evaluation
}

// AssertErrorResponse checks if response contains expected error
func AssertErrorResponse(t *testing.T, resp *http.Response, expectedStatus int, expectedMessage string) {
	t.Helper()
	if resp.StatusCode != expectedStatus {
		t.Errorf("Expected status %d, got %d", expectedStatus, resp.StatusCode)
	}

	var errorResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errorMsg, ok := errorResp["error"].(string); ok {
		if errorMsg != expectedMessage {
			t.Errorf("Expected error message '%s', got '%s'", expectedMessage, errorMsg)
		}
	} else {
		t.Errorf("Error response missing 'error' field")
	}
}

// Sample test data generators
func GetSampleQuestions() []string {
	return []string{
		"Tell me about yourself",
		"What are your strengths?",
		"Describe a challenging project you worked on",
		"Where do you see yourself in 5 years?",
	}
}

func GetSampleTechnicalQuestions() []string {
	return []string{
		"Tell me about your technical background and experience",
		"Describe a challenging technical problem you solved recently",
		"How do you approach debugging and troubleshooting?",
		"What technologies are you most excited about learning?",
		"Walk me through your development process for a new feature",
	}
}

func GetSampleBehavioralQuestions() []string {
	return []string{
		"Tell me about a time when you had to work under pressure",
		"Describe a situation where you had to resolve a conflict with a colleague",
		"Give me an example of when you showed leadership",
		"Tell me about a time you failed and what you learned from it",
		"How do you handle feedback and criticism?",
	}
}

func GetSampleJobDescription() string {
	return "We are looking for a Senior Software Engineer to join our dynamic team. " +
		"The ideal candidate will have experience with backend development, database design, " +
		"and API development. Strong problem-solving skills and ability to work in an agile " +
		"environment are essential."
}

// TODO: Resume file upload support will be added in future iteration
// func GetSampleResumeText() string { ... }

func GetLongMessage() string {
	return "This is a very long message that contains a lot of text to test how the system handles longer inputs. " +
		"It includes multiple sentences and should test the limits of message processing. " +
		"The message continues with more content to ensure we test various edge cases related to message length and content processing. " +
		"This helps us verify that the chat system can handle realistic user inputs of varying lengths."
}

func GetSpecialCharacterMessage() string {
	return "Message with special chars: 你好 🚀 @#$%^&*() \"quotes\" 'apostrophes' and\nnewlines"
}
