package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration.
type Config struct {
	SSHPort     int
	DataDir     string
	MaxSessions int
	IdleTimeout time.Duration
	Debug       bool
}

// Load reads configuration from TERMINAL_PORTFOLIO_ environment variables
// with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		SSHPort:     2222,
		DataDir:     "../data",
		MaxSessions: 100,
		IdleTimeout: 30 * time.Minute,
		Debug:       false,
	}

	if v := os.Getenv("TERMINAL_PORTFOLIO_SSH_PORT"); v != "" {
		port, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid SSH port: %w", err)
		}
		cfg.SSHPort = port
	}

	if v := os.Getenv("TERMINAL_PORTFOLIO_DATA_DIR"); v != "" {
		cfg.DataDir = v
	}

	if v := os.Getenv("TERMINAL_PORTFOLIO_MAX_SESSIONS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid max sessions: %w", err)
		}
		cfg.MaxSessions = n
	}

	if v := os.Getenv("TERMINAL_PORTFOLIO_IDLE_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("invalid idle timeout: %w", err)
		}
		cfg.IdleTimeout = d
	}

	if v := os.Getenv("TERMINAL_PORTFOLIO_DEBUG"); v != "" {
		cfg.Debug = v == "true" || v == "1"
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.SSHPort < 1 || c.SSHPort > 65535 {
		return fmt.Errorf("SSH port must be between 1 and 65535, got %d", c.SSHPort)
	}
	if c.DataDir == "" {
		return fmt.Errorf("data directory must not be empty")
	}
	if c.MaxSessions < 1 {
		return fmt.Errorf("max sessions must be positive, got %d", c.MaxSessions)
	}
	return nil
}
