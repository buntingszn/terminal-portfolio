//go:build !js

package config

import (
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// Set env vars to empty strings so Load() falls through to defaults.
	// t.Setenv restores original values after the test.
	t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "")
	t.Setenv("TERMINAL_PORTFOLIO_DATA_DIR", "")
	t.Setenv("TERMINAL_PORTFOLIO_MAX_SESSIONS", "")
	t.Setenv("TERMINAL_PORTFOLIO_IDLE_TIMEOUT", "")
	t.Setenv("TERMINAL_PORTFOLIO_DEBUG", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SSHPort != 2222 {
		t.Errorf("SSHPort = %d, want 2222", cfg.SSHPort)
	}
	if cfg.DataDir != "../data" {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, "../data")
	}
	if cfg.MaxSessions != 100 {
		t.Errorf("MaxSessions = %d, want 100", cfg.MaxSessions)
	}
	if cfg.IdleTimeout != 30*time.Minute {
		t.Errorf("IdleTimeout = %v, want 30m0s", cfg.IdleTimeout)
	}
	if cfg.Debug {
		t.Error("Debug should be false by default")
	}
}

func TestLoadOverrides(t *testing.T) {
	t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "3333")
	t.Setenv("TERMINAL_PORTFOLIO_DATA_DIR", "/custom/data")
	t.Setenv("TERMINAL_PORTFOLIO_MAX_SESSIONS", "50")
	t.Setenv("TERMINAL_PORTFOLIO_IDLE_TIMEOUT", "1h")
	t.Setenv("TERMINAL_PORTFOLIO_DEBUG", "true")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.SSHPort != 3333 {
		t.Errorf("SSHPort = %d, want 3333", cfg.SSHPort)
	}
	if cfg.DataDir != "/custom/data" {
		t.Errorf("DataDir = %q, want %q", cfg.DataDir, "/custom/data")
	}
	if cfg.MaxSessions != 50 {
		t.Errorf("MaxSessions = %d, want 50", cfg.MaxSessions)
	}
	if cfg.IdleTimeout != time.Hour {
		t.Errorf("IdleTimeout = %v, want 1h0m0s", cfg.IdleTimeout)
	}
	if !cfg.Debug {
		t.Error("Debug should be true")
	}
}

func TestLoadDebugVariants(t *testing.T) {
	tests := []struct {
		value string
		want  bool
	}{
		{"true", true},
		{"1", true},
		{"false", false},
		{"0", false},
		{"yes", false},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Setenv("TERMINAL_PORTFOLIO_DEBUG", tt.value)
			// Reset other vars to defaults.
			t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "2222")
			t.Setenv("TERMINAL_PORTFOLIO_MAX_SESSIONS", "100")

			cfg, err := Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.Debug != tt.want {
				t.Errorf("Debug = %v for %q, want %v", cfg.Debug, tt.value, tt.want)
			}
		})
	}
}

func TestValidationPortTooLow(t *testing.T) {
	t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "0")

	_, err := Load()
	if err == nil {
		t.Error("expected error for port 0")
	}
}

func TestValidationPortTooHigh(t *testing.T) {
	t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "99999")

	_, err := Load()
	if err == nil {
		t.Error("expected error for port 99999")
	}
}

func TestValidationPortNotNumeric(t *testing.T) {
	t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "abc")

	_, err := Load()
	if err == nil {
		t.Error("expected error for non-numeric port")
	}
}

func TestValidationMaxSessionsZero(t *testing.T) {
	t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "2222")
	t.Setenv("TERMINAL_PORTFOLIO_MAX_SESSIONS", "0")

	_, err := Load()
	if err == nil {
		t.Error("expected error for max sessions 0")
	}
}

func TestValidationMaxSessionsNegative(t *testing.T) {
	t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "2222")
	t.Setenv("TERMINAL_PORTFOLIO_MAX_SESSIONS", "-5")

	_, err := Load()
	if err == nil {
		t.Error("expected error for negative max sessions")
	}
}

func TestValidationMaxSessionsNotNumeric(t *testing.T) {
	t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "2222")
	t.Setenv("TERMINAL_PORTFOLIO_MAX_SESSIONS", "abc")

	_, err := Load()
	if err == nil {
		t.Error("expected error for non-numeric max sessions")
	}
}

func TestValidationInvalidTimeout(t *testing.T) {
	t.Setenv("TERMINAL_PORTFOLIO_IDLE_TIMEOUT", "notaduration")

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid timeout")
	}
}

func TestValidationEmptyDataDir(t *testing.T) {
	// DataDir can only be empty if explicitly set via env var,
	// but the env override only triggers on non-empty string.
	// So we test via the validate method directly.
	cfg := &Config{
		SSHPort:     2222,
		DataDir:     "",
		MaxSessions: 100,
		IdleTimeout: 30 * time.Minute,
	}
	if err := cfg.validate(); err == nil {
		t.Error("expected error for empty data dir")
	}
}

func TestValidationBoundaryPorts(t *testing.T) {
	// Port 1 should be valid.
	t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "1")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error for port 1: %v", err)
	}
	if cfg.SSHPort != 1 {
		t.Errorf("SSHPort = %d, want 1", cfg.SSHPort)
	}

	// Port 65535 should be valid.
	t.Setenv("TERMINAL_PORTFOLIO_SSH_PORT", "65535")
	cfg, err = Load()
	if err != nil {
		t.Fatalf("unexpected error for port 65535: %v", err)
	}
	if cfg.SSHPort != 65535 {
		t.Errorf("SSHPort = %d, want 65535", cfg.SSHPort)
	}
}
