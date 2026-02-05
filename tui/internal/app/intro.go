package app

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// introTickInterval is the delay between each boot message appearing.
const introTickInterval = 150 * time.Millisecond

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

// IntroDoneMsg signals that the boot sequence has completed.
type IntroDoneMsg struct{}

// IntroModel manages the BIOS/POST boot sequence animation.
type IntroModel struct {
	messages []bootMessage
	revealed int // number of messages currently visible
	done     bool
	theme    Theme
}

// NewIntroModel creates an IntroModel ready to animate the boot sequence.
func NewIntroModel(theme Theme) IntroModel {
	return IntroModel{
		messages: bootMessages,
		theme:    theme,
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
		// Any key skips the intro.
		m.revealed = len(m.messages)
		m.done = true
		return m, func() tea.Msg { return IntroDoneMsg{} }

	case introTickMsg:
		m.revealed++
		if m.revealed >= len(m.messages) {
			m.revealed = len(m.messages)
			m.done = true
			return m, func() tea.Msg { return IntroDoneMsg{} }
		}
		return m, tea.Tick(introTickInterval, func(_ time.Time) tea.Msg {
			return introTickMsg{}
		})
	}

	return m, nil
}

// View renders the currently revealed boot messages.
func (m IntroModel) View() string {
	if m.revealed == 0 {
		return ""
	}

	var b strings.Builder
	for i := range m.revealed {
		if i >= len(m.messages) {
			break
		}
		msg := m.messages[i]
		styled := m.styleMessage(msg)
		b.WriteString(styled)
		if i < m.revealed-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// SetTheme updates the intro model's theme.
func (m *IntroModel) SetTheme(theme Theme) {
	m.theme = theme
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
