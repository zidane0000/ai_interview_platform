// API route definitions and HTTP server setup
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zidane0000/ai-interview-platform/config"
	"github.com/zidane0000/ai-interview-platform/utils"
)

// SetupRouter initializes the HTTP routes for the API using chi
// Config is injected from main.go to avoid loading configuration multiple times
// frontendHandler is optional - if provided, serves SPA at root
func SetupRouter(cfg *config.Config, frontendHandler http.Handler) http.Handler {
	// BYOK pattern: AI clients created per-request from user-provided keys
	// No shared client needed - see createClientFromRequest() in handlers.go

	// Create handler dependencies (currently empty - BYOK uses per-request clients)
	deps := NewHandlerDependencies()

	r := chi.NewRouter()

	r.Use(CORSMiddleware)
	r.Use(LoggingMiddleware)

	// Health check endpoint at root (for load balancers)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"status":"ok","service":"ai_interview_backend"}`)); err != nil {
			utils.Errorf("Failed to write health check response: %v", err)
		}
	})

	// All API routes under /api prefix
	r.Route("/api", func(r chi.Router) {
		// TODO: Add rate limiting middleware for production
		// TODO: Add authentication middleware if user accounts are implemented
		// TODO: Add request validation middleware
		// TODO: Add API versioning support (e.g., /v1/)

		// Custom NotFound for trailing slash
		r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/interviews/" {
				http.Error(w, ErrMsgMissingInterviewID, ErrCodeBadRequest)
				return
			}
			if r.URL.Path == "/api/evaluation/" {
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

			// Chat session routes for conversational interviews
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

		// Chat routes for real-time interview conversations
		r.Route("/chat", func(r chi.Router) {
			r.Post("/{sessionId}/message", deps.SendMessageHandler)
			r.Get("/{sessionId}", GetChatSessionHandler)
			r.Post("/{sessionId}/end", deps.EndChatSessionHandler)
			// TODO: Add WebSocket support for real-time messaging
			// TODO: Add DELETE /{sessionId} for cleaning up sessions
		})

		// TODO: Add metrics endpoint for monitoring
		// TODO: Add file upload endpoints for resume handling
		// TODO: Add internationalization endpoints for multi-language support
	})

	// Serve frontend SPA if handler provided (production mode)
	if frontendHandler != nil {
		r.Handle("/*", frontendHandler)
	}

	return r
}
