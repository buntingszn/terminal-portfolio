package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// cursorChar is the block character rendered when the cursor is visible.
const cursorChar = "█"

// DefaultBlinkInterval is the default cursor blink interval.
const DefaultBlinkInterval = 530 * time.Millisecond

// cursorBlinkMsg is sent on each tick to toggle cursor visibility.
// The id field supports multiple independent cursor instances.
type cursorBlinkMsg struct {
	id string
}

// Cursor is a blinking block cursor component that integrates with Bubbletea's
// tick system. It can be embedded in any section's View() output.
type Cursor struct {
	visible  bool
	interval time.Duration
	style    lipgloss.Style
	id       string
}

// NewCursor creates a new Cursor with a 530ms blink interval styled with the
// theme's accent color.
func NewCursor(id string, theme Theme) Cursor {
	return Cursor{
		visible:  true,
		interval: DefaultBlinkInterval,
		style:    lipgloss.NewStyle().Foreground(theme.Colors.Accent),
		id:       id,
	}
}

// WithInterval returns a copy of the Cursor with the given blink interval.
func (c Cursor) WithInterval(d time.Duration) Cursor {
	c.interval = d
	return c
}

// Update handles cursorBlinkMsg to toggle visibility and schedule the next tick.
func (c Cursor) Update(msg tea.Msg) (Cursor, tea.Cmd) {
	if blink, ok := msg.(cursorBlinkMsg); ok && blink.id == c.id {
		c.visible = !c.visible
		return c, c.tick()
	}
	return c, nil
}

// View returns the cursor character in accent color when visible, or an
// equal-width space when hidden.
func (c Cursor) View() string {
	if c.visible {
		return c.style.Render(cursorChar)
	}
	return " "
}

// Tick returns the initial tick command that starts the blink loop.
func (c Cursor) Tick() tea.Cmd {
	return c.tick()
}

// tick returns a tea.Cmd that fires a cursorBlinkMsg after the configured interval.
func (c Cursor) tick() tea.Cmd {
	id := c.id
	return tea.Tick(c.interval, func(_ time.Time) tea.Msg {
		return cursorBlinkMsg{id: id}
	})
}
