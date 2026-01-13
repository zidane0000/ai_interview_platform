package config_test

import (
	"os"
	"testing"

	"github.com/zidane0000/ai-interview-platform/config"
)

func TestLoadConfig(t *testing.T) {
	testCases := []struct {
		name        string
		dbURL       string
		port        string
		expectError bool
		expectPort  string
	}{
		{
			name:        "default port",
			dbURL:       "postgres://user:pass@localhost:5432/db",
			port:        "",
			expectError: false,
			expectPort:  "8080",
		},
		{
			name:        "custom port",
			dbURL:       "postgres://user:pass@localhost:5432/db",
			port:        "1234",
			expectError: false,
			expectPort:  "1234",
		}, {
			name:        "missing db url (uses memory backend)",
			dbURL:       "",
			port:        "",
			expectError: false,
			expectPort:  "8080",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			os.Setenv("DATABASE_URL", tc.dbURL)
			os.Setenv("PORT", tc.port)

			cfg, err := config.LoadConfig()
			if tc.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if cfg.Port != tc.expectPort {
					t.Errorf("expected port %s, got %s", tc.expectPort, cfg.Port)
				}
			}
		})
	}
}
