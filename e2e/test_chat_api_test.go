package e2e

import (
	"testing"
)

func TestChatAPI(t *testing.T) {
	// Step 1: Create an interview first
	interview := CreateTestInterview(t, "測試候選人", []string{"Tell me about yourself", "What are your strengths?"})
	t.Logf("Created interview: %+v", interview)

	// Step 2: Start chat session
	chatSession := StartChatSession(t, interview.ID)
	if chatSession.ID == "" || len(chatSession.Messages) == 0 {
		t.Fatalf("Chat session not properly initialized")
	}

	// Step 3: Send a user message
	msgResponse := SendMessage(t, chatSession.ID, "I am a software engineer with 5 years of experience in web development.")
	if msgResponse.Message.Content == "" || msgResponse.AIResponse == nil {
		t.Errorf("Message or AI response missing in send message response")
	}

	// Step 4: Get chat session state
	updatedSession := GetChatSession(t, chatSession.ID)
	if updatedSession.Status == "" || len(updatedSession.Messages) < 2 {
		t.Errorf("Session status or messages not as expected")
	}

	// Step 5: End chat session
	evaluation := EndChatSession(t, chatSession.ID)
	if evaluation.ID == "" || evaluation.Score == 0 {
		t.Errorf("Evaluation ID or score missing")
	}
}
