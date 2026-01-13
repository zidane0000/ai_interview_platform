package e2e

import (
	"testing"
	"time"
)

// TestEvaluationWorkflow tests the complete evaluation workflow with different interview types
func TestEvaluationWorkflow(t *testing.T) {
	t.Run("CompleteEvaluationWorkflow_GeneralInterview", func(t *testing.T) {
		// Create general interview and chat session
		interview := CreateTestInterview(t, "Evaluation Test User", GetSampleQuestions())
		session := StartChatSession(t, interview.ID)

		// Verify the interview type is general
		if interview.InterviewType != "general" {
			t.Errorf("Expected interview type 'general', got '%s'", interview.InterviewType)
		}

		// Simulate general interview responses
		responses := []string{
			"I'm a software engineer with 3 years of experience in web development and API design.",
			"My main strengths are problem-solving and communication. I work well in teams and enjoy mentoring junior developers.",
			"I worked on a challenging e-commerce platform where I optimized the checkout process and reduced cart abandonment by 15%.",
			"In 5 years, I see myself as a senior engineer leading technical decisions and contributing to architecture design.",
		}

		// Send responses and verify AI interaction
		for i, response := range responses {
			msgResp := SendMessage(t, session.ID, response)
			if msgResp.AIResponse == nil {
				t.Errorf("No AI response for general interview message %d", i)
			}
			// Small delay to simulate realistic conversation
			time.Sleep(50 * time.Millisecond)
		}

		// End session and get evaluation
		evaluation := EndChatSession(t, session.ID)

		// Verify evaluation fields
		if evaluation.ID == "" {
			t.Error("Evaluation ID should not be empty")
		}
		if evaluation.InterviewID != interview.ID {
			t.Errorf("Expected interview ID %s, got %s", interview.ID, evaluation.InterviewID)
		}
		if evaluation.Score <= 0 || evaluation.Score > 1 {
			t.Errorf("Score should be between 0 and 1, got %f", evaluation.Score)
		}
		if evaluation.Feedback == "" {
			t.Error("Feedback should not be empty")
		}
		if len(evaluation.Answers) == 0 {
			t.Error("Answers should not be empty")
		}
		if evaluation.CreatedAt.IsZero() {
			t.Error("CreatedAt should not be zero")
		}
	})

	t.Run("CompleteEvaluationWorkflow_TechnicalInterview", func(t *testing.T) {
		// Create technical interview with job description
		interview := CreateTestInterviewWithJobDescription(t,
			"Technical Candidate",
			GetSampleTechnicalQuestions(),
			"technical",
			"en",
			GetSampleJobDescription())
		session := StartChatSession(t, interview.ID)

		// Verify the interview setup
		if interview.InterviewType != "technical" {
			t.Errorf("Expected interview type 'technical', got '%s'", interview.InterviewType)
		}
		if interview.JobDescription == "" {
			t.Error("Expected job description to be set for technical interview")
		}

		// Technical interview responses
		responses := []string{
			"I have 5+ years of backend development experience with Go, Python, and JavaScript.",
			"Recently, I solved a performance issue by optimizing database queries and implementing Redis caching.",
			"I approach debugging systematically: reproduce the issue, check logs, use debugger, and write tests.",
			"I'm excited about learning WebAssembly and exploring its potential for high-performance web applications.",
			"My development process includes requirements analysis, design review, TDD, code review, and deployment.",
		}

		// Send technical responses
		for i, response := range responses {
			msgResp := SendMessage(t, session.ID, response)
			if msgResp.AIResponse == nil {
				t.Errorf("No AI response for technical message %d", i)
			}
			time.Sleep(50 * time.Millisecond)
		}

		// End session and get evaluation
		evaluation := EndChatSession(t, session.ID)

		// Verify evaluation for technical interview
		if evaluation.InterviewID != interview.ID {
			t.Errorf("Expected interview ID %s, got %s", interview.ID, evaluation.InterviewID)
		}
		if evaluation.Score <= 0 || evaluation.Score > 1 {
			t.Errorf("Score should be between 0 and 1, got %f", evaluation.Score)
		}
		if len(evaluation.Feedback) < 10 {
			t.Errorf("Feedback too short: expected at least 10 characters, got %d", len(evaluation.Feedback))
		}
	})

	t.Run("CompleteEvaluationWorkflow_BehavioralInterview", func(t *testing.T) {
		// Create behavioral interview
		interview := CreateTestInterviewWithType(t, "Behavioral Candidate", GetSampleBehavioralQuestions(), "behavioral")
		session := StartChatSession(t, interview.ID)

		// Verify the interview type
		if interview.InterviewType != "behavioral" {
			t.Errorf("Expected interview type 'behavioral', got '%s'", interview.InterviewType)
		}

		// Behavioral interview responses using STAR method
		responses := []string{
			"When facing a tight deadline, I prioritized tasks, communicated with stakeholders, and worked extra hours to deliver on time.",
			"I resolved a team conflict by listening to both sides, finding common ground, and proposing a compromise that worked for everyone.",
			"I led a team migration project by setting clear goals, delegating tasks effectively, and providing regular progress updates.",
			"I failed to meet a project deadline once. I learned to better estimate time requirements and communicate risks early.",
			"I handle feedback by listening actively, asking clarifying questions, and implementing improvements systematically.",
		}

		// Send behavioral responses
		for i, response := range responses {
			msgResp := SendMessage(t, session.ID, response)
			if msgResp.AIResponse == nil {
				t.Errorf("No AI response for behavioral message %d", i)
			}
			time.Sleep(50 * time.Millisecond)
		}

		// End session and get evaluation
		evaluation := EndChatSession(t, session.ID)

		// Verify evaluation for behavioral interview
		if evaluation.InterviewID != interview.ID {
			t.Errorf("Expected interview ID %s, got %s", interview.ID, evaluation.InterviewID)
		}
		if evaluation.Score <= 0 || evaluation.Score > 1 {
			t.Errorf("Score should be between 0 and 1, got %f", evaluation.Score)
		}
		if len(evaluation.Feedback) < 10 {
			t.Errorf("Feedback too short: expected at least 10 characters, got %d", len(evaluation.Feedback))
		}
	})

	t.Run("EvaluationComparison_DifferentTypes", func(t *testing.T) {
		// Test that different interview types can be evaluated independently
		generalInterview := CreateTestInterview(t, "General Candidate", GetSampleQuestions())
		technicalInterview := CreateTestInterviewWithType(t, "Tech Candidate", GetSampleTechnicalQuestions(), "technical")
		behavioralInterview := CreateTestInterviewWithType(t, "Behavioral Candidate", GetSampleBehavioralQuestions(), "behavioral")

		// Verify all interviews have correct types
		if generalInterview.InterviewType != "general" {
			t.Errorf("Expected general interview type, got %s", generalInterview.InterviewType)
		}
		if technicalInterview.InterviewType != "technical" {
			t.Errorf("Expected technical interview type, got %s", technicalInterview.InterviewType)
		}
		if behavioralInterview.InterviewType != "behavioral" {
			t.Errorf("Expected behavioral interview type, got %s", behavioralInterview.InterviewType)
		}

		// All interviews should be valid and have unique IDs
		interviews := []InterviewResponseDTO{generalInterview, technicalInterview, behavioralInterview}
		idMap := make(map[string]bool)
		for i, interview := range interviews {
			if interview.ID == "" {
				t.Errorf("Interview %d has empty ID", i)
			}
			if idMap[interview.ID] {
				t.Errorf("Duplicate interview ID: %s", interview.ID)
			}
			idMap[interview.ID] = true
		}
	})
}
