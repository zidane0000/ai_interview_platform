// Utility functions for the AI Interview Backend
package utils

import (
	"os"
	"strconv"
	"time"
)

// Environment variable parsing utilities

// GetEnvString returns environment variable value or default string
func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt returns environment variable as int or default value
func GetEnvInt(key string, defaultValue int) int {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

// GetEnvBool returns environment variable as bool or default value
func GetEnvBool(key string, defaultValue bool) bool {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}

// GetEnvFloat64 returns environment variable as float64 or default value
func GetEnvFloat64(key string, defaultValue float64) float64 {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
			return value
		}
	}
	return defaultValue
}

// GetEnvDuration returns environment variable as time.Duration or default value
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if valueStr := os.Getenv(key); valueStr != "" {
		if value, err := time.ParseDuration(valueStr); err == nil {
			return value
		}
	}
	return defaultValue
}
