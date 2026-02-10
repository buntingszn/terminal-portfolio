package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// KeyHinter is an optional interface that SectionModels can implement to
// provide contextual key hints displayed in the center of the status bar.
type KeyHinter interface {
	KeyHints() string
}

// ScrollInfo holds viewport scroll state for the status bar.
// When Fits is true, the content does not require scrolling and no scroll
// indicator is shown. Otherwise AtTop/AtBottom determine whether "TOP"/"BOT"
// labels appear, and Percent holds the formatted percentage (e.g., "45%").
type ScrollInfo struct {
	AtTop    bool
	AtBottom bool
	Percent  string // e.g., " 45%"; empty if content fits
	Fits     bool   // true if all content fits without scrolling
}

// ScrollReporter is an optional interface that SectionModels can implement
// to provide scroll position information displayed in the status bar.
type ScrollReporter interface {
	ScrollInfo() ScrollInfo
}

// staticHints is the fixed center text shown in the status bar.
const staticHints = "\u2190/\u2192 nav \u00b7 ? help"

// StatusBar renders a centered status bar with static hints.
type StatusBar struct {
	theme Theme
	width int
}

// NewStatusBar creates a StatusBar with the given theme and terminal width.
func NewStatusBar(theme Theme, width int) StatusBar {
	return StatusBar{
		theme: theme,
		width: width,
	}
}

// SetWidth updates the status bar's width.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// truncateRuneSafe truncates a string to fit within maxWidth visual columns,
// cutting at rune boundaries to avoid splitting multi-byte UTF-8 characters.
func truncateRuneSafe(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	runes := []rune(s)
	for i := len(runes); i > 0; i-- {
		candidate := string(runes[:i])
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}
	return ""
}

// Render returns the styled status bar string with centered static hints.
func (s StatusBar) Render(section Section, hints string, scroll ScrollInfo) string {
	content := staticHints

	hintsW := lipgloss.Width(content)

	// Ultra-narrow: truncate if needed.
	if hintsW > s.width {
		content = truncateRuneSafe(content, s.width)
		hintsW = lipgloss.Width(content)
	}

	// Center the content.
	totalPad := s.width - hintsW
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad
	if leftPad < 0 {
		leftPad = 0
	}
	if rightPad < 0 {
		rightPad = 0
	}

	bar := strings.Repeat(" ", leftPad) + content + strings.Repeat(" ", rightPad)
	return s.theme.StatusBar.Render(bar)
}
