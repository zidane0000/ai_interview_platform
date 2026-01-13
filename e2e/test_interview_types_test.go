package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
)

// TestInterviewTypeValidation tests interview type validation and error handling
func TestInterviewTypeValidation(t *testing.T) {
	baseURL := GetAPIBaseURL()

	t.Run("CreateInterview_InvalidType_ShouldFail", func(t *testing.T) {
		createReq := map[string]interface{}{
			"candidate_name": "Invalid Type Test",
			"questions":      GetSampleQuestions(),
			"interview_type": "invalid_type",
		}

		reqBody, _ := json.Marshal(createReq)
		resp, err := http.Post(baseURL+"/interviews", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Should return 400 Bad Request for invalid interview type
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid interview type, got %d", resp.StatusCode)
		}
	})

	t.Run("CreateInterview_MissingType_ShouldFail", func(t *testing.T) {
		createReq := map[string]interface{}{
			"candidate_name": "Missing Type Test",
			"questions":      GetSampleQuestions(),
			// interview_type is missing
		}

		reqBody, _ := json.Marshal(createReq)
		resp, err := http.Post(baseURL+"/interviews", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Should return 400 Bad Request for missing interview type
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for missing interview type, got %d", resp.StatusCode)
		}
	})

	t.Run("CreateInterview_EmptyType_ShouldFail", func(t *testing.T) {
		createReq := map[string]interface{}{
			"candidate_name": "Empty Type Test",
			"questions":      GetSampleQuestions(),
			"interview_type": "",
		}

		reqBody, _ := json.Marshal(createReq)
		resp, err := http.Post(baseURL+"/interviews", "application/json", bytes.NewBuffer(reqBody))
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		// Should return 400 Bad Request for empty interview type
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400 for empty interview type, got %d", resp.StatusCode)
		}
	})
}

// TestInterviewTypeQuestionDefaults tests that different interview types have appropriate questions
func TestInterviewTypeQuestionDefaults(t *testing.T) {
	t.Run("GeneralType_HasAppropriateQuestions", func(t *testing.T) {
		questions := GetSampleQuestions()
		interview := CreateTestInterviewWithType(t, "Test General", questions, "general")

		// Verify general questions are present
		foundGeneral := false
		for _, q := range interview.Questions {
			if q == "Tell me about yourself" {
				foundGeneral = true
				break
			}
		}
		if !foundGeneral {
			t.Error("General interview should contain appropriate general questions")
		}
	})

	t.Run("TechnicalType_HasAppropriateQuestions", func(t *testing.T) {
		questions := GetSampleTechnicalQuestions()
		interview := CreateTestInterviewWithType(t, "Test Technical", questions, "technical")

		// Verify technical questions are present
		foundTechnical := false
		for _, q := range interview.Questions {
			if q == "Tell me about your technical background and experience" {
				foundTechnical = true
				break
			}
		}
		if !foundTechnical {
			t.Error("Technical interview should contain appropriate technical questions")
		}
	})

	t.Run("BehavioralType_HasAppropriateQuestions", func(t *testing.T) {
		questions := GetSampleBehavioralQuestions()
		interview := CreateTestInterviewWithType(t, "Test Behavioral", questions, "behavioral")

		// Verify behavioral questions are present
		foundBehavioral := false
		for _, q := range interview.Questions {
			if q == "Tell me about a time when you had to work under pressure" {
				foundBehavioral = true
				break
			}
		}
		if !foundBehavioral {
			t.Error("Behavioral interview should contain appropriate behavioral questions")
		}
	})
}
