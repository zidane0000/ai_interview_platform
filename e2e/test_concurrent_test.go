package e2e

import (
	"sync"
	"testing"
	"time"
)

// TestConcurrency tests concurrent operations to ensure thread safety
func TestConcurrency(t *testing.T) {
	t.Run("ConcurrentInterviewCreation", func(t *testing.T) {
		// Test creating multiple interviews concurrently
		const numConcurrent = 10
		var wg sync.WaitGroup
		interviews := make([]InterviewResponseDTO, numConcurrent)

		for i := 0; i < numConcurrent; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Panic in goroutine %d: %v", index, r)
					}
				}()

				candidateName := "Concurrent Test " + string(rune('A'+index))
				interview := CreateTestInterview(t, candidateName, GetSampleQuestions())
				interviews[index] = interview
			}(i)
		}

		wg.Wait()

		// Verify all interviews were created successfully
		idMap := make(map[string]bool)
		for i, interview := range interviews {
			if interview.ID == "" {
				t.Errorf("Interview %d has empty ID", i)
				continue
			}
			if idMap[interview.ID] {
				t.Errorf("Duplicate interview ID: %s", interview.ID)
			}
			idMap[interview.ID] = true
		}

		if len(idMap) != numConcurrent {
			t.Errorf("Expected %d unique interview IDs, got %d", numConcurrent, len(idMap))
		}
	})

	t.Run("ConcurrentChatSessions", func(t *testing.T) {
		// Create one interview, then start multiple chat sessions concurrently
		interview := CreateTestInterview(t, "Concurrent Chat Test", GetSampleQuestions())

		const numSessions = 5
		var wg sync.WaitGroup
		sessions := make([]ChatInterviewSessionDTO, numSessions)

		for i := 0; i < numSessions; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Panic in chat session goroutine %d: %v", index, r)
					}
				}()

				session := StartChatSession(t, interview.ID)
				sessions[index] = session
			}(i)
		}

		wg.Wait()

		// Verify all sessions were created with unique IDs
		sessionIDMap := make(map[string]bool)
		for i, session := range sessions {
			if session.ID == "" {
				t.Errorf("Session %d has empty ID", i)
				continue
			}
			if sessionIDMap[session.ID] {
				t.Errorf("Duplicate session ID: %s", session.ID)
			}
			sessionIDMap[session.ID] = true

			if session.InterviewID != interview.ID {
				t.Errorf("Session %d has wrong interview ID", i)
			}
		}

		if len(sessionIDMap) != numSessions {
			t.Errorf("Expected %d unique session IDs, got %d", numSessions, len(sessionIDMap))
		}
	})

	t.Run("ConcurrentMessaging", func(t *testing.T) {
		// Create interview and chat session
		interview := CreateTestInterview(t, "Concurrent Messaging Test", GetSampleQuestions())
		session := StartChatSession(t, interview.ID)

		// Send multiple messages concurrently to the same session
		const numMessages = 5
		var wg sync.WaitGroup
		responses := make([]SendMessageResponseDTO, numMessages)

		for i := 0; i < numMessages; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Panic in messaging goroutine %d: %v", index, r)
					}
				}()

				message := "Concurrent message " + string(rune('1'+index))
				response := SendMessage(t, session.ID, message)
				responses[index] = response
			}(i)
		}

		wg.Wait()

		// Verify all messages were processed
		messageIDMap := make(map[string]bool)
		for i, response := range responses {
			if response.Message.ID == "" {
				t.Errorf("Message %d has empty ID", i)
				continue
			}
			if messageIDMap[response.Message.ID] {
				t.Errorf("Duplicate message ID: %s", response.Message.ID)
			}
			messageIDMap[response.Message.ID] = true

			if response.AIResponse == nil {
				t.Errorf("Message %d missing AI response", i)
			}
		}

		// Get final session state to verify all messages are stored
		finalSession := GetChatSession(t, session.ID)

		// Should have initial AI greeting + user messages + AI responses
		expectedMinMessages := 1 + (numMessages * 2) // greeting + (user + ai) * numMessages
		if len(finalSession.Messages) < expectedMinMessages {
			t.Errorf("Expected at least %d messages in final session, got %d", expectedMinMessages, len(finalSession.Messages))
		}
	})

	t.Run("ConcurrentSessionOperations", func(t *testing.T) {
		// Test different operations on the same session concurrently
		interview := CreateTestInterview(t, "Concurrent Ops Test", GetSampleQuestions())
		session := StartChatSession(t, interview.ID)

		var wg sync.WaitGroup
		var getSessionError, sendMessageError error

		// Concurrent GET session state
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					getSessionError = r.(error)
				}
			}()
			GetChatSession(t, session.ID)
		}()

		// Concurrent send message
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					sendMessageError = r.(error)
				}
			}()
			SendMessage(t, session.ID, "Concurrent operation test")
		}()

		wg.Wait()

		if getSessionError != nil {
			t.Errorf("Error in concurrent get session: %v", getSessionError)
		}
		if sendMessageError != nil {
			t.Errorf("Error in concurrent send message: %v", sendMessageError)
		}
	})

	t.Run("MemoryStoreConsistency", func(t *testing.T) {
		// Test that memory store maintains consistency under concurrent access
		const numOperations = 20
		var wg sync.WaitGroup

		// Mix of different operations
		for i := 0; i < numOperations; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Panic in operation %d: %v", index, r)
					}
				}()

				switch index % 3 {
				case 0:
					// Create interview
					CreateTestInterview(t, "Memory Test", GetSampleQuestions())
				case 1:
					// Create interview and start chat
					interview := CreateTestInterview(t, "Memory Chat Test", GetSampleQuestions())
					StartChatSession(t, interview.ID)
				case 2:
					// Full workflow
					interview := CreateTestInterview(t, "Memory Full Test", GetSampleQuestions())
					session := StartChatSession(t, interview.ID)
					SendMessage(t, session.ID, "Test message")
				}
			}(i)
		}

		wg.Wait()
		// If we reach here without panics, memory store handled concurrent access
	})

	t.Run("LoadTesting", func(t *testing.T) {
		// Simple load test with many concurrent users
		if testing.Short() {
			t.Skip("Skipping load test in short mode")
		}

		const numUsers = 50
		const messagesPerUser = 3

		var wg sync.WaitGroup
		start := time.Now()

		for i := 0; i < numUsers; i++ {
			wg.Add(1)
			go func(userIndex int) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Panic for user %d: %v", userIndex, r)
					}
				}()

				// Each user creates interview, chats, and gets evaluation
				candidateName := "Load Test User " + string(rune('0'+userIndex))
				interview := CreateTestInterview(t, candidateName, GetSampleQuestions())
				session := StartChatSession(t, interview.ID)

				for j := 0; j < messagesPerUser; j++ {
					message := "Load test message " + string(rune('1'+j))
					SendMessage(t, session.ID, message)
				}

				EndChatSession(t, session.ID)
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)

		totalOperations := numUsers * (1 + 1 + messagesPerUser + 1) // create + start + messages + end
		operationsPerSecond := float64(totalOperations) / duration.Seconds()

		t.Logf("Load test completed: %d users, %d total operations in %v (%.2f ops/sec)",
			numUsers, totalOperations, duration, operationsPerSecond)

		// Basic performance assertion - should handle at least 10 ops/sec
		if operationsPerSecond < 10 {
			t.Errorf("Performance too slow: %.2f ops/sec", operationsPerSecond)
		}
	})
}
