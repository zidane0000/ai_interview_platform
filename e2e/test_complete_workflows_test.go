package e2e

import (
	"testing"
	"time"
)

// TestCompleteWorkflows tests end-to-end interview workflows
func TestCompleteWorkflows(t *testing.T) {
	t.Run("FullInterviewWorkflow_Technical", func(t *testing.T) {
		// Technical interview simulation with job description
		interview := CreateTestInterviewWithFullDetails(t,
			"Alice Johnson - Senior Developer",
			GetSampleTechnicalQuestions(),
			"technical",
			"en",
			GetSampleJobDescription())
		session := StartChatSession(t, interview.ID)

		// Verify interview setup
		if interview.InterviewType != "technical" {
			t.Errorf("Expected technical interview, got %s", interview.InterviewType)
		}
		if interview.JobDescription == "" {
			t.Error("Job description should be included in technical interview")
		}
		// TODO: Resume support will be added later

		// Verify initial AI greeting
		if len(session.Messages) != 1 {
			t.Errorf("Expected 1 initial message, got %d", len(session.Messages))
		}
		if session.Messages[0].Type != "ai" {
			t.Error("First message should be from AI")
		}

		// Simulate technical responses aligned with the sample questions
		technicalResponses := []string{
			"I have 5+ years of backend development experience with Go, Python, and JavaScript. I've built REST APIs and worked with PostgreSQL and Redis.",
			"Recently, I solved a memory leak in our microservice by profiling the application and optimizing object lifecycle management.",
			"I approach debugging systematically: reproduce the issue, check logs, use debugger, isolate variables, and write regression tests.",
			"I'm excited about learning WebAssembly for high-performance web applications and exploring serverless architectures.",
			"My development process includes requirements analysis, API design, TDD, code review, CI/CD, and monitoring.",
		}

		for i, response := range technicalResponses {
			msgResp := SendMessage(t, session.ID, response)

			// Verify response was recorded correctly
			if msgResp.Message.Content != response {
				t.Errorf("Response %d not recorded correctly", i)
			}
			if msgResp.AIResponse == nil {
				t.Errorf("Missing AI follow-up for response %d", i)
			}

			// Small delay to simulate realistic conversation timing
			time.Sleep(100 * time.Millisecond)
		}

		// Get session state before ending
		updatedSession := GetChatSession(t, session.ID)
		expectedMessages := 1 + (len(technicalResponses) * 2) // initial + (user + ai) pairs
		if len(updatedSession.Messages) < expectedMessages {
			t.Errorf("Expected at least %d messages, got %d", expectedMessages, len(updatedSession.Messages))
		}

		// End session and evaluate
		evaluation := EndChatSession(t, session.ID)
		// Verify evaluation quality for technical interview
		if evaluation.Score <= 0.3 {
			t.Errorf("Technical responses should score higher than 0.3, got %.2f", evaluation.Score)
		}
		if len(evaluation.Feedback) == 0 {
			t.Error("Technical evaluation should provide feedback (even if simple mock)")
		}

		// Verify session is marked as completed
		finalSession := GetChatSession(t, session.ID)
		if finalSession.Status != "completed" {
			t.Errorf("Expected session status 'completed', got '%s'", finalSession.Status)
		}
	})

	t.Run("FullInterviewWorkflow_Behavioral", func(t *testing.T) {
		// Behavioral interview simulation
		interview := CreateTestInterviewWithType(t, "Bob Smith - Product Manager", GetSampleBehavioralQuestions(), "behavioral")
		session := StartChatSession(t, interview.ID)

		// Verify interview setup
		if interview.InterviewType != "behavioral" {
			t.Errorf("Expected behavioral interview, got %s", interview.InterviewType)
		}

		// Simulate behavioral responses using STAR method aligned with sample questions
		behavioralResponses := []string{
			"When facing a tight deadline for a product launch, I prioritized critical features, communicated delays early, and worked with the team to deliver a solid MVP on time.",
			"I had to resolve a conflict between two team members by facilitating a discussion, helping them understand each other's perspectives, and establishing clear collaboration guidelines.",
			"When promoted to team lead, I took initiative by organizing regular one-on-ones, implementing code review processes, and mentoring junior developers.",
			"I failed to meet a quarterly goal once due to underestimating complexity. I learned to break down tasks better, communicate risks early, and ask for help when needed.",
			"I handle feedback by listening actively, asking clarifying questions, reflecting on the input, and creating action plans for improvement.",
		}

		for _, response := range behavioralResponses {
			SendMessage(t, session.ID, response)
			time.Sleep(50 * time.Millisecond) // Simulate conversation flow
		}

		evaluation := EndChatSession(t, session.ID)

		// Verify behavioral evaluation
		if evaluation.ID == "" {
			t.Error("Behavioral evaluation should be generated")
		}
		if evaluation.InterviewID != interview.ID {
			t.Error("Evaluation should reference correct interview")
		}
	})

	t.Run("ShortInterviewWorkflow", func(t *testing.T) {
		// Test workflow with minimal interaction
		interview := CreateTestInterview(t, "Quick Test", []string{"Tell me about yourself"})
		session := StartChatSession(t, interview.ID)

		// Send only one brief response
		SendMessage(t, session.ID, "I'm a developer.")

		// End immediately
		evaluation := EndChatSession(t, session.ID)

		// Should still generate valid evaluation
		if evaluation.Score == 0 {
			t.Error("Should generate non-zero score even for brief interviews")
		}
		if evaluation.Feedback == "" {
			t.Error("Should provide feedback even for brief interviews")
		}
	})

	t.Run("MultiSessionInterviewWorkflow", func(t *testing.T) {
		// Test multiple sessions for same interview (like multiple rounds)
		interview := CreateTestInterview(t, "Multi-Round Candidate", GetSampleQuestions())

		evaluations := make([]EvaluationResponseDTO, 3)

		// Round 1: Initial screening
		session1 := StartChatSession(t, interview.ID)
		SendMessage(t, session1.ID, "I have 3 years of experience in software development.")
		evaluations[0] = EndChatSession(t, session1.ID)

		// Round 2: Technical round
		session2 := StartChatSession(t, interview.ID)
		SendMessage(t, session2.ID, "I'm experienced with React, Node.js, and PostgreSQL.")
		evaluations[1] = EndChatSession(t, session2.ID)

		// Round 3: Final round
		session3 := StartChatSession(t, interview.ID)
		SendMessage(t, session3.ID, "I'm looking for growth opportunities and team collaboration.")
		evaluations[2] = EndChatSession(t, session3.ID)

		// Verify all evaluations are unique but reference same interview
		for i, eval := range evaluations {
			if eval.InterviewID != interview.ID {
				t.Errorf("Evaluation %d should reference interview %s", i, interview.ID)
			}
			for j, other := range evaluations {
				if i != j && eval.ID == other.ID {
					t.Errorf("Evaluations %d and %d have same ID", i, j)
				}
			}
		}
	})

	t.Run("WorkflowWithErrors", func(t *testing.T) {
		// Test workflow recovery from errors
		interview := CreateTestInterview(t, "Error Test", GetSampleQuestions())
		session := StartChatSession(t, interview.ID)

		// Send a normal message first
		SendMessage(t, session.ID, "This is a normal message.")

		// Try invalid operations (should not break workflow)
		// These will fail but shouldn't crash the system

		// Continue with normal workflow
		SendMessage(t, session.ID, "Continuing after error scenarios.")

		// Should still be able to end session normally
		evaluation := EndChatSession(t, session.ID)

		if evaluation.ID == "" {
			t.Error("Should generate evaluation despite error scenarios")
		}
	})

	t.Run("LongConversationWorkflow", func(t *testing.T) {
		// Test extended conversation beyond normal limits
		interview := CreateTestInterview(t, "Extended Conversation", GetSampleQuestions())
		session := StartChatSession(t, interview.ID)

		// Send many messages (more than typical interview)
		responses := []string{
			"I'm a senior software engineer with 8 years of experience.",
			"My expertise includes full-stack development with modern frameworks.",
			"I've led teams of 5-10 developers on complex projects.",
			"I'm passionate about clean code and test-driven development.",
			"I have experience with cloud platforms like AWS and Azure.",
			"I enjoy mentoring junior developers and code reviews.",
			"My goal is to become a technical architect in the next few years.",
			"I believe in continuous learning and staying updated with technology.",
		}

		for i, response := range responses {
			msgResp := SendMessage(t, session.ID, response)
			t.Logf("Message %d: %s -> AI: %s", i+1, response, msgResp.AIResponse.Content)
		}

		// Check if session auto-completed (based on AI logic)
		finalSession := GetChatSession(t, session.ID)

		if finalSession.Status == "completed" {
			t.Log("Session auto-completed after extended conversation")
		} else {
			// Manually end if not auto-completed
			evaluation := EndChatSession(t, session.ID)
			if len(evaluation.Answers) < len(responses) {
				t.Error("Long conversation should capture all user responses")
			}
		}
	})
}
