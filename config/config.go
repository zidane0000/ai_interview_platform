// Configuration loading from environment variables and .env files
package config

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/zidane0000/ai-interview-platform/utils"
)

// Config holds all application configuration
type Config struct {
	// Server configuration
	Port            string
	ShutdownTimeout time.Duration

	// Database configuration
	DatabaseURL string

	// AI service configuration
	GeminiAPIKey string
	OpenAIAPIKey string

	// TODO: Add more AI providers
	// TODO: Add file upload configuration
	// TODO: Add security configuration
	// TODO: Add logging configuration
	// TODO: Add internationalization configuration
	// TODO: Add email/notification configuration
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file (ignore error if file doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		Port:            utils.GetEnvString("PORT", "8080"),
		GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
		OpenAIAPIKey:    os.Getenv("OPENAI_API_KEY"),
		ShutdownTimeout: utils.GetEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),
	}

	// TODO: Load file upload configuration(cfg.UploadPath, cfg.MaxFileSize)
	// TODO: Load security configuration(cfg.JWTSecret, cfg.CORSOrigins)
	// TODO: Validate file paths and create directories if needed
	// TODO: Validate email configuration if notifications are enabled
	// TODO: Load configuration from config files (YAML, JSON, TOML)
	// TODO: Add configuration hot-reloading capability
	// TODO: Add configuration validation with detailed error messages

	return cfg, nil
}

// TODO: Add configuration for different environments (dev, staging, prod)
// TODO: Add configuration documentation and examples
// TODO: Add configuration schema validation
// TODO: Add sensitive data masking in logs
