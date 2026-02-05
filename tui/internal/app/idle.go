package app

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Idle timeout constants.
const (
	// idleCheckInterval is how often we check the idle timer.
	idleCheckInterval = 10 * time.Second

	// idleWarningBefore is how long before timeout we show a warning.
	idleWarningBefore = 1 * time.Minute
)

// idleCheckMsg is sent periodically to check idle time.
type idleCheckMsg struct{}

// idleCheckTick returns a tea.Cmd that fires idleCheckMsg after idleCheckInterval.
func idleCheckTick() tea.Cmd {
	return tea.Tick(idleCheckInterval, func(_ time.Time) tea.Msg {
		return idleCheckMsg{}
	})
}

// resetIdleTimer marks the current time as the last user activity,
// dismisses any idle warning, and returns true if idle tracking is active.
func (m *Model) resetIdleTimer() {
	if m.idleTimeout > 0 {
		m.lastActivity = time.Now()
		m.showIdleWarning = false
	}
}

// handleIdleCheck processes an idleCheckMsg: checks elapsed idle time,
// shows a warning when approaching timeout, or quits on expiry.
// Returns the updated model and any commands.
func (m Model) handleIdleCheck() (Model, tea.Cmd) {
	if m.idleTimeout <= 0 {
		return m, nil
	}

	elapsed := time.Since(m.lastActivity)

	// Timeout expired: quit the session.
	if elapsed >= m.idleTimeout {
		return m, tea.Quit
	}

	// Approaching timeout: show warning.
	remaining := m.idleTimeout - elapsed
	if remaining <= idleWarningBefore {
		m.showIdleWarning = true
		m.idleRemaining = remaining
	}

	return m, idleCheckTick()
}

// idleWarningView renders the idle timeout warning banner.
func (m Model) idleWarningView() string {
	secs := int(m.idleRemaining.Seconds())
	if secs < 0 {
		secs = 0
	}

	msg := fmt.Sprintf("Idle timeout in %ds — press any key to stay connected", secs)

	style := lipgloss.NewStyle().
		Foreground(m.theme.Colors.Bg).
		Background(m.theme.Colors.Accent).
		Bold(true).
		Padding(0, 1)

	rendered := style.Render(msg)

	// Center the warning horizontally.
	if m.width > 0 {
		return lipgloss.PlaceHorizontal(m.width, lipgloss.Center, rendered)
	}
	return rendered
}
