package config

import "time"

// Config holds the application configuration.
type Config struct {
	SSHPort     int
	DataDir     string
	MaxSessions int
	// IdleTimeout controls how long a session can remain idle before being
	// disconnected. A value of 0 disables idle timeout entirely.
	IdleTimeout time.Duration
	// RateLimitPerIP is the maximum number of connections per IP within the
	// rate limit window. Default: 10.
	RateLimitPerIP int
	// RateLimitWindow is the time window for rate limiting. Default: 1m.
	RateLimitWindow time.Duration
	Debug           bool
}
