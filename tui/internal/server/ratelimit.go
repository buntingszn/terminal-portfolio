//go:build !js

package server

import (
	"sync"
	"time"
)

// ipState tracks rate limit state for a single IP address.
type ipState struct {
	count    int
	lastSeen time.Time
	active   int
}

// RateLimiter provides per-IP rate limiting with both request rate
// and concurrent connection tracking. It is safe for concurrent use.
type RateLimiter struct {
	mu         sync.Mutex
	ips        map[string]*ipState
	maxPerIP   int
	windowSize time.Duration
}

// NewRateLimiter creates a rate limiter that allows at most maxPerIP
// requests per window duration from a single IP address.
func NewRateLimiter(maxPerIP int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		ips:        make(map[string]*ipState),
		maxPerIP:   maxPerIP,
		windowSize: window,
	}
}

// Allow checks whether a request from the given IP should be allowed.
// If allowed, it increments both the request count and active connection
// count for the IP. The caller must call Release when the connection ends.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	state, ok := rl.ips[ip]
	if !ok {
		rl.ips[ip] = &ipState{
			count:    1,
			lastSeen: now,
			active:   1,
		}
		return true
	}

	// Reset count if the window has elapsed since last seen.
	if now.Sub(state.lastSeen) >= rl.windowSize {
		state.count = 0
	}

	// Reject if at the per-IP request limit within the window.
	if state.count >= rl.maxPerIP {
		return false
	}

	state.count++
	state.active++
	state.lastSeen = now
	return true
}

// Release decrements the active connection count for an IP.
// It should be called when a connection from that IP ends.
func (rl *RateLimiter) Release(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if state, ok := rl.ips[ip]; ok {
		state.active--
		if state.active < 0 {
			state.active = 0
		}
	}
}

// Cleanup removes entries for IPs that have not been seen within
// twice the window duration and have no active connections.
// It should be called periodically to prevent memory leaks.
func (rl *RateLimiter) Cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	cutoff := time.Now().Add(-2 * rl.windowSize)
	for ip, state := range rl.ips {
		if state.active <= 0 && state.lastSeen.Before(cutoff) {
			delete(rl.ips, ip)
		}
	}
}
