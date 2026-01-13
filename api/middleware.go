// Middleware for logging, authentication, etc.
package api

import (
	"net/http"
	"time"

	"github.com/zidane0000/ai-interview-platform/utils"
)

// LoggingMiddleware logs the HTTP method, path, status, and duration for each request.
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		utils.Infof("%s %s %d %s", r.Method, r.URL.Path, lrw.statusCode, duration)
	})
}

// loggingResponseWriter wraps http.ResponseWriter to capture status code.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// CORSMiddleware adds CORS headers to allow cross-origin requests from browsers
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Development: Allow localhost origins
		// TODO: In production, replace with specific allowed origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
		}

		// Check if origin is allowed
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}

		if isAllowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else if origin == "" {
			// Allow same-origin requests (no Origin header)
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-OpenAI-Key, X-Gemini-Key, X-OpenAI-Base-URL")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// TODO: Implement additional middleware for production readiness:

// TODO: RequestIDMiddleware - Essential for distributed tracing
// - Assigns a unique request ID to each request
// - Adds request ID to context and response headers
// - Enables request tracing across microservices
// - Helps with debugging and monitoring

// TODO: AuthMiddleware - Required for user authentication
// - Validates JWT tokens or API keys
// - Extracts user information from tokens
// - Handles token refresh logic
// - Protects secured endpoints

// TODO: RecoveryMiddleware - Critical for application stability
// - Catches panics and prevents server crashes
// - Logs detailed error information
// - Returns appropriate 500 error responses
// - Maintains service availability

// TODO: RateLimitMiddleware - Essential for API protection
// - Prevents abuse and DOS attacks
// - Configurable limits per endpoint/user
// - Implements sliding window or token bucket algorithms
// - Returns 429 Too Many Requests with retry-after headers

// TODO: MetricsMiddleware - Required for monitoring
// - Collects request count, duration, and status codes
// - Exposes Prometheus-compatible metrics
// - Tracks error rates and response times
// - Enables alerting and SLA monitoring

// TODO: ValidationMiddleware - Improves security and data quality
// - Validates request content types and sizes
// - Sanitizes input data for XSS prevention
// - Validates required headers and parameters
// - Returns detailed validation error messages

// TODO: CompressionMiddleware - Optimizes performance
// - Compresses responses with gzip/deflate
// - Reduces bandwidth usage and improves speed
// - Configurable compression levels
// - Handles Accept-Encoding headers properly

// TODO: SecurityMiddleware - Hardens application security
// - Adds security headers (CSP, HSTS, X-Frame-Options)
// - Prevents common web vulnerabilities
// - Implements IP whitelisting/blacklisting
// - Handles security-related response headers

// TODO: CacheMiddleware - Improves performance for read operations
// - Implements ETag and conditional requests
// - Caches frequently accessed data
// - Handles cache invalidation properly
// - Reduces database load for static content

// TODO: Enhanced CORS middleware improvements:
// - Load allowed origins from configuration
// - Support wildcard subdomain matching
// - Add more granular header controls
// - Implement CORS preflight caching
