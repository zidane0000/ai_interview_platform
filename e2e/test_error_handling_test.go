package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
)

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	baseURL := GetAPIBaseURL()

	t.Run("CreateInterview_MissingFields", func(t *testing.T) {
		// Test with missing candidate name
		createReq := map[string]interface{}{
			"questions": []string{"Test question"},
		}
		reqBody, _ := json.Marshal(createReq)
		resp, err := http.Post(baseURL+"/interviews", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		AssertErrorResponse(t, resp, http.StatusBadRequest, "Missing candidate_name or questions")
	})

	t.Run("CreateInterview_EmptyQuestions", func(t *testing.T) {
		// Test with empty questions array
		createReq := map[string]interface{}{
			"candidate_name": "Test User",
			"questions":      []string{},
		}
		reqBody, _ := json.Marshal(createReq)
		resp, err := http.Post(baseURL+"/interviews", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		AssertErrorResponse(t, resp, http.StatusBadRequest, "Missing candidate_name or questions")
	})

	t.Run("CreateInterview_InvalidJSON", func(t *testing.T) {
		// Test with malformed JSON
		resp, err := http.Post(baseURL+"/interviews", "application/json", strings.NewReader("{invalid json"))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		AssertErrorResponse(t, resp, http.StatusBadRequest, "Invalid JSON")
	})

	t.Run("StartChat_InvalidInterviewID", func(t *testing.T) {
		// Test starting chat with non-existent interview ID
		resp, err := http.Post(fmt.Sprintf("%s/interviews/invalid-id/chat/start", baseURL), "application/json", nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		AssertErrorResponse(t, resp, http.StatusNotFound, "Interview not found")
	})

	t.Run("SendMessage_InvalidSessionID", func(t *testing.T) {
		// Test sending message to non-existent session
		sendMsgReq := SendMessageRequestDTO{
			Message: "Test message",
		}
		reqBody, _ := json.Marshal(sendMsgReq)
		resp, err := http.Post(fmt.Sprintf("%s/chat/invalid-session/message", baseURL), "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		AssertErrorResponse(t, resp, http.StatusNotFound, "Chat session not found")
	})

	t.Run("SendMessage_EmptyMessage", func(t *testing.T) {
		// First create valid interview and chat session
		interview := CreateTestInterview(t, "Test User", GetSampleQuestions())
		session := StartChatSession(t, interview.ID)

		// Test sending empty message
		sendMsgReq := SendMessageRequestDTO{
			Message: "",
		}
		reqBody, _ := json.Marshal(sendMsgReq)
		resp, err := http.Post(fmt.Sprintf("%s/chat/%s/message", baseURL, session.ID), "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		AssertErrorResponse(t, resp, http.StatusBadRequest, "Message cannot be empty")
	})

	t.Run("GetChatSession_InvalidID", func(t *testing.T) {
		// Test getting non-existent chat session
		resp, err := http.Get(fmt.Sprintf("%s/chat/invalid-session", baseURL))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		AssertErrorResponse(t, resp, http.StatusNotFound, "Chat session not found")
	})

	t.Run("EndChatSession_InvalidID", func(t *testing.T) {
		// Test ending non-existent chat session
		resp, err := http.Post(fmt.Sprintf("%s/chat/invalid-session/end", baseURL), "application/json", nil)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		AssertErrorResponse(t, resp, http.StatusNotFound, "Chat session not found")
	})

	t.Run("SendMessage_CompletedSession", func(t *testing.T) {
		// Create interview and chat session
		interview := CreateTestInterview(t, "Test User", GetSampleQuestions())
		session := StartChatSession(t, interview.ID)

		// End the session first
		EndChatSession(t, session.ID)

		// Try to send message to completed session
		sendMsgReq := SendMessageRequestDTO{
			Message: "This should fail",
		}
		reqBody, _ := json.Marshal(sendMsgReq)
		resp, err := http.Post(fmt.Sprintf("%s/chat/%s/message", baseURL, session.ID), "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		defer resp.Body.Close()

		AssertErrorResponse(t, resp, http.StatusBadRequest, "Chat session is not active")
	})
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	t.Run("LongMessage", func(t *testing.T) {
		// Test with very long message
		interview := CreateTestInterview(t, "Test User", GetSampleQuestions())
		session := StartChatSession(t, interview.ID)

		longMessage := GetLongMessage()
		msgResponse := SendMessage(t, session.ID, longMessage)

		if msgResponse.Message.Content != longMessage {
			t.Errorf("Long message not preserved correctly")
		}
		if msgResponse.AIResponse == nil {
			t.Errorf("AI response missing for long message")
		}
	})

	t.Run("SpecialCharacters", func(t *testing.T) {
		// Test with special characters
		interview := CreateTestInterview(t, "Test User", GetSampleQuestions())
		session := StartChatSession(t, interview.ID)

		specialMessage := GetSpecialCharacterMessage()
		msgResponse := SendMessage(t, session.ID, specialMessage)

		if msgResponse.Message.Content != specialMessage {
			t.Errorf("Special characters not preserved correctly")
		}
		if msgResponse.AIResponse == nil {
			t.Errorf("AI response missing for special character message")
		}
	})

	t.Run("MultipleQuestions", func(t *testing.T) {
		// Test interview with many questions
		manyQuestions := make([]string, 20)
		for i := 0; i < 20; i++ {
			manyQuestions[i] = fmt.Sprintf("Question %d: Tell me about topic %d", i+1, i+1)
		}

		interview := CreateTestInterview(t, "Test User", manyQuestions)
		if len(interview.Questions) != 20 {
			t.Errorf("Expected 20 questions, got %d", len(interview.Questions))
		}
	})

	t.Run("UnicodeNames", func(t *testing.T) {
		// Test with Unicode candidate names
		unicodeNames := []string{
			"张三",
			"José García",
			"Müller",
			"محمد",
			"Александр",
		}

		for _, name := range unicodeNames {
			interview := CreateTestInterview(t, name, GetSampleQuestions())
			if interview.CandidateName != name {
				t.Errorf("Unicode name not preserved: expected %s, got %s", name, interview.CandidateName)
			}
		}
	})
}
