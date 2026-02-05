package app

import (
	"strings"
)

// DefaultKeyHints are shown when the active section does not implement KeyHinter.
const DefaultKeyHints = "j/k scroll \u00b7 1-4 nav \u00b7 ? help"

// KeyHinter is an optional interface that SectionModels can implement to
// provide contextual key hints displayed in the center of the status bar.
type KeyHinter interface {
	KeyHints() string
}

// StatusBar renders a 3-zone bottom bar: left path, center hints, right section name.
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

// SetTheme updates the status bar's theme.
func (s *StatusBar) SetTheme(theme Theme) {
	s.theme = theme
}

// SetWidth updates the status bar's width.
func (s *StatusBar) SetWidth(width int) {
	s.width = width
}

// Render returns the styled status bar string for the given section and hints.
// Left zone: ~/section-name (terminal path style)
// Center zone: key hints
// Right zone: SECTION NAME in caps
func (s StatusBar) Render(section Section, hints string) string {
	left := " ~/" + SectionName(section)
	right := strings.ToUpper(SectionName(section)) + " "

	if hints == "" {
		hints = DefaultKeyHints
	}

	// Calculate available space for center text and gaps.
	usedByEnds := len(left) + len(right)
	remaining := s.width - usedByEnds

	if remaining <= 0 {
		// Terminal too narrow for center text; just render left + right.
		gap := ""
		if s.width > usedByEnds {
			gap = strings.Repeat(" ", s.width-usedByEnds)
		}
		return s.theme.StatusBar.Render(left + gap + right)
	}

	// Center the hints text within the remaining space.
	centerLen := len(hints)
	if centerLen > remaining {
		// Truncate hints if wider than available space.
		hints = hints[:remaining]
		centerLen = remaining
	}

	totalPad := remaining - centerLen
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad

	bar := left +
		strings.Repeat(" ", leftPad) +
		hints +
		strings.Repeat(" ", rightPad) +
		right

	return s.theme.StatusBar.Render(bar)
}
