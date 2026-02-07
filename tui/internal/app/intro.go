package app

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// introTickInterval is the delay between each boot message appearing.
const introTickInterval = 80 * time.Millisecond

// introFinalDelay is the longer delay before the final boot message appears.
const introFinalDelay = 150 * time.Millisecond

// introPauseDuration is the pause after all messages are revealed before
// transitioning out of the intro.
const introPauseDuration = 200 * time.Millisecond

// bootMessageType identifies the color category for a boot message.
type bootMessageType string

const (
	bootSystem  bootMessageType = "system"
	bootInfo    bootMessageType = "info"
	bootSuccess bootMessageType = "success"
	bootAccent  bootMessageType = "accent"
)

// bootMessage is a single line in the boot sequence.
type bootMessage struct {
	Text string
	Type bootMessageType
}

// bootMessages is the embedded boot sequence, matching boot-messages.json.
var bootMessages = []bootMessage{
	{Text: "POST: System initialization...", Type: bootSystem},
	{Text: "BIOS v1.0.0 — terminal-portfolio", Type: bootSystem},
	{Text: "Memory test: 128GB OK", Type: bootInfo},
	{Text: "Detecting hardware... AMD Ryzen AI MAX+ 395", Type: bootInfo},
	{Text: "GPU: Radeon 8060S (gfx1151) — 124GB VRAM allocated", Type: bootInfo},
	{Text: "Loading content modules...", Type: bootSystem},
	{Text: "  [OK] about.json", Type: bootSuccess},
	{Text: "  [OK] work.json", Type: bootSuccess},
	{Text: "  [OK] cv.json", Type: bootSuccess},
	{Text: "  [OK] links.json", Type: bootSuccess},
	{Text: "  [OK] meta.json", Type: bootSuccess},
	{Text: "Initializing theme engine... warm-minimalist loaded", Type: bootInfo},
	{Text: "Starting SSH listener on :2222...", Type: bootSystem},
	{Text: "All systems nominal. Welcome.", Type: bootAccent},
}

// introTickMsg advances the boot sequence by one message.
type introTickMsg struct{}

// introPauseMsg signals that the post-reveal pause has elapsed.
type introPauseMsg struct{}

// IntroDoneMsg signals that the boot sequence has completed.
type IntroDoneMsg struct{}

// IntroModel manages the BIOS/POST boot sequence animation.
type IntroModel struct {
	messages []bootMessage
	revealed int // number of messages currently visible
	done     bool
	paused   bool // true after all messages revealed, waiting before IntroDoneMsg
	cursor   Cursor
	theme    Theme
	width    int
	height   int
}

// NewIntroModel creates an IntroModel ready to animate the boot sequence.
func NewIntroModel(theme Theme) IntroModel {
	return IntroModel{
		messages: bootMessages,
		theme:    theme,
		cursor:   NewCursor("intro-cursor", theme),
	}
}

// Init returns the first tick command to start the boot sequence.
func (m IntroModel) Init() tea.Cmd {
	return tea.Tick(introTickInterval, func(_ time.Time) tea.Msg {
		return introTickMsg{}
	})
}

// Update handles tick messages and key presses (skip).
func (m IntroModel) Update(msg tea.Msg) (IntroModel, tea.Cmd) {
	if m.done {
		return m, nil
	}

	switch msg.(type) {
	case tea.KeyMsg:
		// Any key skips the intro (works during both reveal and pause phases).
		m.revealed = len(m.messages)
		m.done = true
		m.paused = false
		return m, func() tea.Msg { return IntroDoneMsg{} }

	case introTickMsg:
		m.revealed++
		if m.revealed >= len(m.messages) {
			// All messages revealed: enter the pause phase with a blinking cursor.
			m.revealed = len(m.messages)
			m.paused = true
			return m, tea.Batch(
				tea.Tick(introPauseDuration, func(_ time.Time) tea.Msg {
					return introPauseMsg{}
				}),
				m.cursor.Tick(),
			)
		}
		// Use a longer delay before revealing the final message.
		delay := introTickInterval
		if m.revealed == len(m.messages)-1 {
			delay = introFinalDelay
		}
		return m, tea.Tick(delay, func(_ time.Time) tea.Msg {
			return introTickMsg{}
		})

	case introPauseMsg:
		// Pause elapsed: complete the intro.
		m.done = true
		m.paused = false
		return m, func() tea.Msg { return IntroDoneMsg{} }

	case cursorBlinkMsg:
		// Delegate cursor blink messages during the pause phase.
		if m.paused {
			var cmd tea.Cmd
			m.cursor, cmd = m.cursor.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

// View renders the currently revealed boot messages.
func (m IntroModel) View() string {
	if m.revealed == 0 {
		return ""
	}

	endIdx := m.revealed
	if endIdx > len(m.messages) {
		endIdx = len(m.messages)
	}

	// Determine visible window: show only the most recent N messages
	// when terminal height is limited.
	startIdx := 0
	maxVisible := m.height
	if maxVisible <= 0 {
		maxVisible = endIdx // no limit if height unknown
	}
	if endIdx-startIdx > maxVisible {
		startIdx = endIdx - maxVisible
	}

	var b strings.Builder
	for i := startIdx; i < endIdx; i++ {
		msg := m.messages[i]
		text := truncateBootMsg(msg.Text, m.width)
		truncated := bootMessage{Text: text, Type: msg.Type}
		styled := m.styleMessage(truncated)
		b.WriteString(styled)
		// Append blinking cursor after the final message during the pause.
		if m.paused && i == endIdx-1 {
			b.WriteString(m.cursor.View())
		}
		if i < endIdx-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// SetSize updates the intro model's known terminal dimensions.
func (m *IntroModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// SetTheme updates the intro model's theme.
func (m *IntroModel) SetTheme(theme Theme) {
	m.theme = theme
	m.cursor.SetTheme(theme)
}

// truncateBootMsg truncates text to fit within maxWidth, adding an ellipsis
// when truncation occurs.
func truncateBootMsg(text string, maxWidth int) string {
	if maxWidth <= 0 || len(text) <= maxWidth {
		return text
	}
	if maxWidth <= 3 {
		return text[:maxWidth]
	}
	return text[:maxWidth-3] + "..."
}

// styleMessage returns the styled text for a single boot message.
func (m IntroModel) styleMessage(msg bootMessage) string {
	var style lipgloss.Style
	switch msg.Type {
	case bootSystem:
		style = lipgloss.NewStyle().Foreground(m.theme.Colors.Fg)
	case bootInfo:
		style = lipgloss.NewStyle().Foreground(m.theme.Colors.Muted)
	case bootSuccess:
		style = lipgloss.NewStyle().Foreground(m.theme.Colors.Accent)
	case bootAccent:
		style = lipgloss.NewStyle().Foreground(m.theme.Colors.Accent).Bold(true)
	default:
		style = lipgloss.NewStyle().Foreground(m.theme.Colors.Fg)
	}
	return style.Render(msg.Text)
}
