// E2E Tests for Multi-Language Interview Support
// Tests the complete flow: Frontend language selection → Backend API → AI responses
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

// Helper function to make JSON requests with proper error handling
func makeJSONRequest(t *testing.T, method, endpoint string, payload interface{}) (*http.Response, []byte) {
	t.Helper()
	baseURL := GetAPIBaseURL()

	var reqBody *bytes.Buffer
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("Failed to marshal request payload: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, baseURL+endpoint, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Read response body
	responseBody := make([]byte, 0)
	if resp.Body != nil {
		defer resp.Body.Close()

		buf := make([]byte, 1024)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				responseBody = append(responseBody, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
	}

	return resp, responseBody
}

// Helper function to count Chinese characters in text
func countChineseCharacters(text string) int {
	count := 0
	for _, r := range text {
		if r >= 0x4e00 && r <= 0x9fff { // CJK Unified Ideographs range
			count++
		}
	}
	return count
}

// TestCreateInterviewWithLanguage tests creating interviews with language preference
func TestCreateInterviewWithLanguage(t *testing.T) {
	tests := []struct {
		name             string
		language         interface{} // Use interface{} to test both valid strings and invalid types
		expectedStatus   int
		expectedLanguage string
	}{
		// Valid Language Cases
		{
			name:             "Create interview with English language",
			language:         "en",
			expectedStatus:   http.StatusCreated,
			expectedLanguage: "en",
		},
		{
			name:             "Create interview with Traditional Chinese language",
			language:         "zh-TW",
			expectedStatus:   http.StatusCreated,
			expectedLanguage: "zh-TW",
		},
		{
			name:             "Create interview without language (should default to English)",
			language:         "",
			expectedStatus:   http.StatusCreated,
			expectedLanguage: "en",
		},
		// Invalid Language String Cases
		{
			name:             "Invalid language should fail",
			language:         "invalid-lang",
			expectedStatus:   http.StatusBadRequest,
			expectedLanguage: "",
		},
		{
			name:             "Simplified Chinese should fail (not supported)",
			language:         "zh-CN",
			expectedStatus:   http.StatusBadRequest,
			expectedLanguage: "",
		},
		{
			name:             "French language should fail (not supported)",
			language:         "fr",
			expectedStatus:   http.StatusBadRequest,
			expectedLanguage: "",
		},
		{
			name:             "Very long string should be rejected",
			language:         "this-is-way-too-long-to-be-a-valid-language-code-and-should-be-rejected",
			expectedStatus:   http.StatusBadRequest,
			expectedLanguage: "",
		},
		// Edge Cases
		{
			name:             "Number instead of string should be rejected",
			language:         123,
			expectedStatus:   http.StatusBadRequest,
			expectedLanguage: "",
		},
		{
			name:             "Object instead of string should be rejected",
			language:         map[string]string{"lang": "en"},
			expectedStatus:   http.StatusBadRequest,
			expectedLanguage: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Use map for flexible payload construction
			payload := map[string]interface{}{
				"candidate_name": "Test Candidate",
				"questions":      GetSampleQuestions(),
				"interview_type": "general", // Required field
			}

			// Add language field if provided (handles both strings and invalid types)
			if tt.language != "" && tt.language != nil {
				payload["interview_language"] = tt.language
			}

			// Act
			resp, body := makeJSONRequest(t, "POST", "/interviews", payload)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
				t.Logf("Response body: %s", string(body))
				return
			}

			if tt.expectedStatus == http.StatusCreated {
				// Parse as generic map to check for language field
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal interview response: %v", err)
				}

				// CRITICAL CHECK: Interview language field MUST exist in response
				languageField, hasLanguage := response["interview_language"]
				if !hasLanguage {
					t.Errorf("Response missing 'interview_language' field")
					t.Logf("Response: %+v", response)
					return
				}

				// CRITICAL CHECK: Language value MUST match expected
				actualLanguage, ok := languageField.(string)
				if !ok {
					t.Errorf("Language field should be string, got %T", languageField)
					return
				}

				if actualLanguage != tt.expectedLanguage {
					t.Errorf("Expected language '%s', got '%s'", tt.expectedLanguage, actualLanguage)
				}
			}
		})
	}

	// Additional tests for interview types
	t.Run("CreateInterview_TechnicalWithLanguage", func(t *testing.T) {
		// Test technical interview with Traditional Chinese
		interview := CreateTestInterviewWithTypeAndLanguage(t, "技術面試候選人", GetSampleTechnicalQuestions(), "technical", "zh-TW")

		// Verify interview type and language
		if interview.InterviewType != "technical" {
			t.Errorf("Expected interview type 'technical', got '%s'", interview.InterviewType)
		}
		if interview.InterviewLanguage != "zh-TW" {
			t.Errorf("Expected language 'zh-TW', got '%s'", interview.InterviewLanguage)
		}
		if interview.CandidateName != "技術面試候選人" {
			t.Errorf("Expected candidate name '技術面試候選人', got '%s'", interview.CandidateName)
		}
	})

	t.Run("CreateInterview_BehavioralWithLanguage", func(t *testing.T) {
		// Test behavioral interview with English
		interview := CreateTestInterviewWithTypeAndLanguage(t, "Behavioral Candidate", GetSampleBehavioralQuestions(), "behavioral", "en")

		// Verify interview type and language
		if interview.InterviewType != "behavioral" {
			t.Errorf("Expected interview type 'behavioral', got '%s'", interview.InterviewType)
		}
		if interview.InterviewLanguage != "en" {
			t.Errorf("Expected language 'en', got '%s'", interview.InterviewLanguage)
		}
	})
}

// TestAIQuestionGenerationLanguage tests AI question generation in different languages
func TestAIQuestionGenerationLanguage(t *testing.T) {
	tests := []struct {
		name              string
		language          string
		candidateName     string
		shouldHaveChinese bool
	}{
		{
			name:              "Generate questions in English",
			language:          "en",
			candidateName:     "Test Candidate",
			shouldHaveChinese: false,
		},
		{
			name:              "Generate questions in Traditional Chinese",
			language:          "zh-TW",
			candidateName:     "測試候選人",
			shouldHaveChinese: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create interview with specified language
			interview := CreateTestInterviewWithLanguage(t, tt.candidateName, GetSampleQuestions(), tt.language)

			// DEBUG: Log the created interview details
			t.Logf("Created interview - Language: %s, Expected: %s", interview.InterviewLanguage, tt.language)

			chatSession := StartChatSession(t, interview.ID)

			// Check if initial AI message language matches expectation
			if len(chatSession.Messages) > 0 {
				firstMessage := chatSession.Messages[0]
				actualHasChinese := countChineseCharacters(firstMessage.Content) > 0

				// DEBUG: Log message details
				t.Logf("AI Message: %s", firstMessage.Content)
				t.Logf("Chinese character count: %d", countChineseCharacters(firstMessage.Content))

				if actualHasChinese != tt.shouldHaveChinese {
					t.Errorf("Language mismatch - Expected Chinese: %v, Has Chinese: %v, Message: %s",
						tt.shouldHaveChinese, actualHasChinese, firstMessage.Content)
				}
			}
		})
	}
}

// TestChatSessionLanguagePersistence tests that chat sessions maintain language context
func TestChatSessionLanguagePersistence(t *testing.T) {
	// Arrange - Create interview with Traditional Chinese
	interview := CreateTestInterviewWithLanguage(t, "測試候選人", GetSampleQuestions(), "zh-TW")

	// Act - Start chat session (should inherit language from interview)
	payload := map[string]interface{}{}
	resp, body := makeJSONRequest(t, "POST", fmt.Sprintf("/interviews/%s/chat/start", interview.ID), payload)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d", resp.StatusCode)
	}

	// Parse response as generic map to check for language field
	var chatResponse map[string]interface{}
	if err := json.Unmarshal(body, &chatResponse); err != nil {
		t.Fatalf("Failed to decode chat session: %v", err)
	}
	// CRITICAL CHECK: Chat session must have language field
	languageField, hasLanguage := chatResponse["session_language"]
	if !hasLanguage {
		t.Errorf("Chat session missing 'session_language' field")
		t.Logf("Chat Response: %+v", chatResponse)
		return
	}

	// CRITICAL CHECK: Language value must match expected
	actualLanguage, ok := languageField.(string)
	if !ok {
		t.Errorf("Language field should be string, got %T", languageField)
		return
	}

	if actualLanguage != "zh-TW" {
		t.Errorf("Expected chat session language 'zh-TW', got '%s'", actualLanguage)
	}

	// Get session ID for message testing
	sessionID, exists := chatResponse["id"].(string)
	if !exists || sessionID == "" {
		t.Fatalf("Chat session ID missing or invalid")
	}

	// Test sending a message - AI response should be in Chinese
	msgResponse := SendMessage(t, sessionID, "你好，我是候選人")
	// CRITICAL CHECK: AI response must be in Traditional Chinese
	if msgResponse.AIResponse == nil {
		t.Errorf("AI response should not be nil")
		return
	}

	aiContent := msgResponse.AIResponse.Content
	if countChineseCharacters(aiContent) == 0 {
		t.Errorf("AI response should contain Chinese characters for zh-TW language")
		t.Logf("AI Response: %s", aiContent)
	}
}

// TestEvaluationLanguage tests that evaluations are generated in correct language
func TestEvaluationLanguage(t *testing.T) {
	// Create interview with Traditional Chinese language
	interview := CreateTestInterviewWithLanguage(t, "測試候選人", GetSampleQuestions(), "zh-TW")
	chatSession := StartChatSession(t, interview.ID)

	// Send some messages to have content for evaluation
	SendMessage(t, chatSession.ID, "我有五年的軟體開發經驗，主要使用 JavaScript 和 Python。")

	// End session and get evaluation
	evaluation := EndChatSession(t, chatSession.ID)

	// Assert - Evaluation should be created successfully
	if evaluation.ID == "" {
		t.Errorf("Evaluation ID should not be empty")
	}

	if evaluation.Feedback == "" {
		t.Errorf("Evaluation feedback should not be empty")
	}

	// Validate evaluation is in Traditional Chinese
	if countChineseCharacters(evaluation.Feedback) == 0 {
		t.Errorf("Evaluation feedback should be in Traditional Chinese for zh-TW interview")
		t.Logf("Actual feedback: %s", evaluation.Feedback)
	}
}

// TestCompleteLanguageWorkflow tests the complete workflow with language support
func TestCompleteLanguageWorkflow(t *testing.T) {
	// Test complete workflow: Create interview → Start chat → Send messages → End chat → Get evaluation
	// Create interview with Traditional Chinese
	interview := CreateTestInterviewWithLanguage(t, "測試候選人", GetSampleQuestions(), "zh-TW")
	t.Logf("Created interview with ID: %s", interview.ID)

	// Start chat session
	chatSession := StartChatSession(t, interview.ID)
	t.Logf("Started chat session with ID: %s", chatSession.ID)

	// Send multiple messages
	messages := []string{
		"你好，我是軟體工程師",
		"我有三年的開發經驗",
		"我熟悉 React 和 Node.js",
	}

	for _, msg := range messages {
		response := SendMessage(t, chatSession.ID, msg)
		if response.AIResponse == nil {
			t.Errorf("AI should respond to message: %s", msg)
		}
		t.Logf("Sent: %s, AI Response: %s", msg, response.AIResponse.Content)
	}
	// End session and get evaluation
	evaluation := EndChatSession(t, chatSession.ID)

	// Verify evaluation was created
	if evaluation.ID == "" {
		t.Errorf("Evaluation should be created")
	}

	if evaluation.Score == 0 {
		t.Errorf("Evaluation should have a score")
	}

	if evaluation.Feedback == "" {
		t.Errorf("Evaluation feedback should not be empty")
	}
	// CRITICAL: Validate evaluation is in Traditional Chinese
	if countChineseCharacters(evaluation.Feedback) == 0 {
		t.Errorf("Evaluation feedback should be in Traditional Chinese for zh-TW interview")
		t.Logf("Actual feedback: %s", evaluation.Feedback)
	}

	t.Logf("Evaluation Score: %.2f, Feedback: %s", evaluation.Score, evaluation.Feedback)
}
