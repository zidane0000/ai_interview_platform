// API route definitions and HTTP server setup
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zidane0000/ai-interview-platform/ai"
	"github.com/zidane0000/ai-interview-platform/config"
	"github.com/zidane0000/ai-interview-platform/utils"
)

// SetupRouter initializes the HTTP routes for the API using chi
// Config is injected from main.go to avoid loading configuration multiple times
func SetupRouter(cfg *config.Config) http.Handler {
	// Create AI client with configuration (simplified - no factory pattern)
	aiConfig := ai.NewDefaultAIConfig()
	aiClient, err := ai.NewAIClient(aiConfig)
	if err != nil {
		utils.Errorf("Failed to create AI client: %v", err)
		// Fall back to mock provider if client creation fails
		aiConfig.DefaultProvider = ai.ProviderMock
		aiClient, _ = ai.NewAIClient(aiConfig)
	}

	// Create handler dependencies
	deps := NewHandlerDependencies(aiClient)

	r := chi.NewRouter()

	r.Use(CORSMiddleware)
	r.Use(LoggingMiddleware)

	// TODO: Add rate limiting middleware for production
	// TODO: Add authentication middleware if user accounts are implemented
	// TODO: Add request validation middleware
	// TODO: Add API versioning support (e.g., /api/v1/)

	// Custom NotFound for trailing slash
	r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/interviews/" {
			http.Error(w, ErrMsgMissingInterviewID, ErrCodeBadRequest)
			return
		}
		if r.URL.Path == "/evaluation/" {
			http.Error(w, ErrMsgMissingEvaluationID, ErrCodeBadRequest)
			return
		}
		// TODO: Add custom 404 response for chat endpoints
		http.NotFound(w, r)
	}))

	// Interview routes
	r.Route("/interviews", func(r chi.Router) {
		r.Post("/", CreateInterviewHandler)
		r.Get("/", ListInterviewsHandler)
		r.Get("/{id}", GetInterviewHandler)

		// TODO: Implement chat session routes for conversational interviews
		// These routes are expected by the frontend for chat-based interviews
		r.Post("/{id}/chat/start", deps.StartChatSessionHandler)
		// TODO: Add PUT /{id} for updating interviews
		// TODO: Add DELETE /{id} for removing interviews
	})

	// Evaluation routes
	r.Route("/evaluation", func(r chi.Router) {
		r.Post("/", deps.SubmitEvaluationHandler)
		r.Get("/{id}", GetEvaluationHandler)
		// TODO: Add GET / for listing evaluations
		// TODO: Add PUT /{id} for updating evaluations
		// TODO: Add DELETE /{id} for removing evaluations
	})
	// TODO: Implement chat routes for real-time interview conversations
	// These routes are required by the frontend chat functionality
	r.Route("/chat", func(r chi.Router) {
		r.Post("/{sessionId}/message", deps.SendMessageHandler)
		r.Get("/{sessionId}", GetChatSessionHandler)
		r.Post("/{sessionId}/end", deps.EndChatSessionHandler)
		// TODO: Add WebSocket support for real-time messaging
		// TODO: Add DELETE /{sessionId} for cleaning up sessions
	})
	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok","service":"ai_interview_backend"}`)); err != nil {
			utils.Errorf("Failed to write health check response: %v", err)
		}
	})

	// TODO: Add metrics endpoint for monitoring
	// TODO: Add file upload endpoints for resume handling
	// TODO: Add internationalization endpoints for multi-language support

	return r
}
