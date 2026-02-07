//go:build !js

package server

import (
	"sync"
	"testing"
	"time"
)

func TestAllow_UnderLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)

	if !rl.Allow("10.0.0.1") {
		t.Error("first request from new IP should be allowed")
	}
}

func TestAllow_OverLimit(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)

	for i := range 3 {
		if !rl.Allow("10.0.0.1") {
			t.Errorf("request %d should be allowed", i+1)
		}
		rl.Release("10.0.0.1")
	}

	// 4th request within the same window should be rejected.
	if rl.Allow("10.0.0.1") {
		t.Error("request exceeding maxPerIP should be rejected")
	}
}

func TestRelease(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)

	if !rl.Allow("10.0.0.1") {
		t.Fatal("first request should be allowed")
	}

	rl.Release("10.0.0.1")

	rl.mu.Lock()
	state := rl.ips["10.0.0.1"]
	active := state.active
	rl.mu.Unlock()

	if active != 0 {
		t.Errorf("active count after release = %d, want 0", active)
	}
}

func TestRelease_NeverNegative(t *testing.T) {
	rl := NewRateLimiter(10, time.Minute)

	if !rl.Allow("10.0.0.1") {
		t.Fatal("first request should be allowed")
	}

	// Release more times than allowed.
	rl.Release("10.0.0.1")
	rl.Release("10.0.0.1")

	rl.mu.Lock()
	state := rl.ips["10.0.0.1"]
	active := state.active
	rl.mu.Unlock()

	if active < 0 {
		t.Errorf("active count should never go negative, got %d", active)
	}
}

func TestCleanup_RemovesStale(t *testing.T) {
	window := 50 * time.Millisecond
	rl := NewRateLimiter(10, window)

	if !rl.Allow("10.0.0.1") {
		t.Fatal("first request should be allowed")
	}
	rl.Release("10.0.0.1")

	// Wait longer than 2x window so the entry becomes stale.
	time.Sleep(3 * window)

	rl.Cleanup()

	rl.mu.Lock()
	_, exists := rl.ips["10.0.0.1"]
	rl.mu.Unlock()

	if exists {
		t.Error("stale IP entry should have been removed by cleanup")
	}
}

func TestCleanup_KeepsRecent(t *testing.T) {
	window := time.Minute
	rl := NewRateLimiter(10, window)

	if !rl.Allow("10.0.0.1") {
		t.Fatal("first request should be allowed")
	}
	rl.Release("10.0.0.1")

	// Cleanup immediately; entry was just created so it's recent.
	rl.Cleanup()

	rl.mu.Lock()
	_, exists := rl.ips["10.0.0.1"]
	rl.mu.Unlock()

	if !exists {
		t.Error("recent IP entry should be kept after cleanup")
	}
}

func TestCleanup_KeepsActive(t *testing.T) {
	window := 50 * time.Millisecond
	rl := NewRateLimiter(10, window)

	if !rl.Allow("10.0.0.1") {
		t.Fatal("first request should be allowed")
	}

	// Wait longer than 2x window but do NOT release.
	time.Sleep(3 * window)

	rl.Cleanup()

	rl.mu.Lock()
	_, exists := rl.ips["10.0.0.1"]
	rl.mu.Unlock()

	if !exists {
		t.Error("IP with active connections should not be removed even if stale")
	}
}

func TestAllow_DifferentIPs(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)

	if !rl.Allow("10.0.0.1") {
		t.Error("first IP should be allowed")
	}

	if !rl.Allow("10.0.0.2") {
		t.Error("second IP should be allowed (independent limit)")
	}
}

func TestAllow_WindowReset(t *testing.T) {
	window := 50 * time.Millisecond
	rl := NewRateLimiter(1, window)

	if !rl.Allow("10.0.0.1") {
		t.Fatal("first request should be allowed")
	}
	rl.Release("10.0.0.1")

	if rl.Allow("10.0.0.1") {
		t.Error("second request within window should be rejected")
	}

	// Wait for the window to elapse.
	time.Sleep(2 * window)

	if !rl.Allow("10.0.0.1") {
		t.Error("request after window reset should be allowed")
	}
}

func TestConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(1000, time.Minute)

	var wg sync.WaitGroup
	goroutines := 100

	for range goroutines {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 10 {
				if rl.Allow("10.0.0.1") {
					rl.Release("10.0.0.1")
				}
			}
		}()
	}

	wg.Wait()

	rl.mu.Lock()
	state := rl.ips["10.0.0.1"]
	active := state.active
	rl.mu.Unlock()

	if active != 0 {
		t.Errorf("after all goroutines complete, active = %d, want 0", active)
	}
}
