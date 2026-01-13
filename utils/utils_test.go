package utils_test

import (
	"os"
	"testing"
	"time"

	"github.com/zidane0000/ai-interview-platform/utils"
)

func TestGetEnvString(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue string
		expected     string
	}{
		{"with value", "test-value", "default", "test-value"},
		{"empty value", "", "default", "default"},
		{"whitespace value", "  ", "default", "  "}, // whitespace is considered a value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_STRING"
			os.Unsetenv(key)
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			}
			defer os.Unsetenv(key)

			result := utils.GetEnvString(key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue int
		expected     int
	}{
		{"valid integer", "42", 10, 42},
		{"empty string", "", 10, 10},
		{"invalid integer", "abc", 10, 10},
		{"negative integer", "-5", 10, -5},
		{"zero", "0", 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_INT"
			os.Unsetenv(key)
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			}
			defer os.Unsetenv(key)

			result := utils.GetEnvInt(key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue bool
		expected     bool
	}{
		{"true", "true", false, true},
		{"false", "false", true, false},
		{"1", "1", false, true},
		{"0", "0", true, false},
		{"empty string", "", true, true},
		{"invalid boolean", "maybe", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_BOOL"
			os.Unsetenv(key)
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			}
			defer os.Unsetenv(key)

			result := utils.GetEnvBool(key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetEnvFloat64(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue float64
		expected     float64
	}{
		{"valid float", "3.14", 1.0, 3.14},
		{"integer as float", "42", 1.0, 42.0},
		{"empty string", "", 1.5, 1.5},
		{"invalid float", "not-a-number", 2.0, 2.0},
		{"zero", "0", 1.0, 0.0},
		{"negative float", "-2.5", 1.0, -2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_FLOAT"
			os.Unsetenv(key)
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			}
			defer os.Unsetenv(key)

			result := utils.GetEnvFloat64(key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestGetEnvDuration(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		defaultValue time.Duration
		expected     time.Duration
	}{
		{"valid duration seconds", "30s", 10 * time.Second, 30 * time.Second},
		{"valid duration minutes", "5m", 1 * time.Minute, 5 * time.Minute},
		{"valid duration hours", "2h", 1 * time.Hour, 2 * time.Hour},
		{"empty string", "", 15 * time.Second, 15 * time.Second},
		{"invalid duration", "invalid", 20 * time.Second, 20 * time.Second},
		{"zero duration", "0s", 10 * time.Second, 0 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_DURATION"
			os.Unsetenv(key)
			if tt.envValue != "" {
				os.Setenv(key, tt.envValue)
			}
			defer os.Unsetenv(key)

			result := utils.GetEnvDuration(key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
