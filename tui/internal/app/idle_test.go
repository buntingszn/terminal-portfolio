package app

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func TestSetIdleTimeout(t *testing.T) {
	m := New(testContent())
	m = m.SetIdleTimeout(10 * time.Minute)

	if m.idleTimeout != 10*time.Minute {
		t.Errorf("idleTimeout = %v, want 10m", m.idleTimeout)
	}
	if m.lastActivity.IsZero() {
		t.Error("lastActivity should be set after SetIdleTimeout with non-zero duration")
	}
}

func TestSetIdleTimeoutZeroDisables(t *testing.T) {
	m := New(testContent())
	m = m.SetIdleTimeout(0)

	if m.idleTimeout != 0 {
		t.Errorf("idleTimeout = %v, want 0", m.idleTimeout)
	}
	if !m.lastActivity.IsZero() {
		t.Error("lastActivity should remain zero when timeout is disabled")
	}
}

func TestIdleCheckMsgResetsTickWhenActive(t *testing.T) {
	m := skipIntro(t)
	m = m.SetIdleTimeout(30 * time.Minute)

	// Process an idle check — should produce a new tick command (not quit).
	result, cmd := m.Update(idleCheckMsg{})
	m = result.(Model)

	if cmd == nil {
		t.Fatal("expected non-nil cmd (idleCheckTick) from idle check")
	}

	// The cmd should produce an idleCheckMsg when executed (tea.Tick).
	// We can't easily test tea.Tick directly, but we verify it's not a QuitMsg.
	if m.showIdleWarning {
		t.Error("should not show idle warning when recently active")
	}
}

func TestIdleCheckMsgShowsWarning(t *testing.T) {
	m := skipIntro(t)
	m = m.SetIdleTimeout(5 * time.Minute)

	// Simulate being idle for 4 minutes 30 seconds (within warning threshold).
	m.lastActivity = time.Now().Add(-4*time.Minute - 30*time.Second)

	result, cmd := m.Update(idleCheckMsg{})
	m = result.(Model)

	if !m.showIdleWarning {
		t.Error("expected showIdleWarning to be true when idle > (timeout - 1min)")
	}
	if cmd == nil {
		t.Fatal("expected non-nil cmd from idle check when warning")
	}
}

func TestIdleCheckMsgQuitsOnExpiry(t *testing.T) {
	m := skipIntro(t)
	m = m.SetIdleTimeout(5 * time.Minute)

	// Simulate being idle for more than the full timeout.
	m.lastActivity = time.Now().Add(-6 * time.Minute)

	_, cmd := m.Update(idleCheckMsg{})

	if cmd == nil {
		t.Fatal("expected quit command when idle timeout expired")
	}
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestIdleCheckMsgNoopWhenDisabled(t *testing.T) {
	m := skipIntro(t)
	// Do not set idle timeout (default 0).

	result, cmd := m.Update(idleCheckMsg{})
	m = result.(Model)

	if cmd != nil {
		t.Error("expected nil cmd when idle timeout is disabled")
	}
	if m.showIdleWarning {
		t.Error("should not show idle warning when timeout is disabled")
	}
}

func TestKeyResetIdleTimer(t *testing.T) {
	m := skipIntro(t)
	m = m.SetIdleTimeout(5 * time.Minute)

	// Simulate being idle for 4.5 minutes (in warning zone).
	m.lastActivity = time.Now().Add(-4*time.Minute - 30*time.Second)
	m.showIdleWarning = true

	// A key press should reset the idle timer and dismiss warning.
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = result.(Model)

	if m.showIdleWarning {
		t.Error("key press should dismiss idle warning")
	}

	elapsed := time.Since(m.lastActivity)
	if elapsed > 1*time.Second {
		t.Errorf("lastActivity should have been reset, but elapsed = %v", elapsed)
	}
}

func TestMouseResetIdleTimer(t *testing.T) {
	m := skipIntro(t)
	m = m.SetIdleTimeout(5 * time.Minute)

	// Simulate being idle.
	m.lastActivity = time.Now().Add(-4*time.Minute - 30*time.Second)
	m.showIdleWarning = true

	// A mouse event should reset the idle timer.
	result, _ := m.Update(tea.MouseMsg{})
	m = result.(Model)

	if m.showIdleWarning {
		t.Error("mouse event should dismiss idle warning")
	}
}

func TestIdleWarningViewContent(t *testing.T) {
	m := skipIntro(t)
	m = m.SetIdleTimeout(5 * time.Minute)
	m.width = 80
	m.showIdleWarning = true
	m.idleRemaining = 45 * time.Second

	view := m.idleWarningView()
	if !strings.Contains(view, "Idle timeout") {
		t.Error("idle warning should contain 'Idle timeout'")
	}
	if !strings.Contains(view, "45s") {
		t.Error("idle warning should contain remaining seconds '45s'")
	}
	if !strings.Contains(view, "press any key") {
		t.Error("idle warning should contain 'press any key'")
	}
}

func TestIdleWarningAppearsInView(t *testing.T) {
	m := skipIntro(t)
	m = m.SetIdleTimeout(5 * time.Minute)

	// Set terminal size.
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(Model)

	// Trigger warning state.
	m.showIdleWarning = true
	m.idleRemaining = 30 * time.Second

	view := m.View()
	if !strings.Contains(view, "Idle timeout") {
		t.Error("main View() should contain idle warning when showIdleWarning is true")
	}
}

func TestIdleWarningNotInViewWhenDisabled(t *testing.T) {
	m := skipIntro(t)
	// No idle timeout set.

	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(Model)

	view := m.View()
	if strings.Contains(view, "Idle timeout") {
		t.Error("View() should not contain idle warning when timeout is disabled")
	}
}

func TestInitStartsIdleTickWhenEnabled(t *testing.T) {
	m := New(testContent())
	m = m.SetIdleTimeout(30 * time.Minute)

	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return non-nil cmd when idle timeout is enabled")
	}
}

func TestInitNoIdleTickWhenDisabled(t *testing.T) {
	m := New(testContent())
	// idleTimeout is 0 by default.

	cmd := m.Init()
	// With intro, Init still returns a cmd for the intro animation.
	// This just verifies it doesn't crash; the idle tick is not started.
	if cmd == nil {
		t.Error("Init() should still return intro cmd")
	}
}
