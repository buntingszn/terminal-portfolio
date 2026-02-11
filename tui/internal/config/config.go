package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds the application configuration.
type Config struct {
	SSHHost     string
	SSHPort     int
	DataDir     string
	MaxSessions int
	// IdleTimeout controls how long a session can remain idle before being
	// disconnected. A value of 0 disables idle timeout entirely.
	IdleTimeout time.Duration
	// AnalyticsFile is the path to the JSONL analytics log file.
	// An empty string disables analytics logging.
	AnalyticsFile string
	Debug         bool
}

// Load reads configuration from TERMINAL_PORTFOLIO_ environment variables
// with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		SSHHost:       "127.0.0.1",
		SSHPort:       2222,
		DataDir:       "../data",
		MaxSessions:   100,
		IdleTimeout:   30 * time.Minute,
		AnalyticsFile: "analytics.jsonl",
		Debug:         false,
	}

	if v := os.Getenv("TERMINAL_PORTFOLIO_SSH_HOST"); v != "" {
		cfg.SSHHost = v
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

	if v, ok := os.LookupEnv("TERMINAL_PORTFOLIO_ANALYTICS_FILE"); ok {
		cfg.AnalyticsFile = v
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
