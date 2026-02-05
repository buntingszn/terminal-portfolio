package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DefaultKeyHints are shown when the active section does not implement KeyHinter.
const DefaultKeyHints = "j/k scroll \u00b7 1-4 nav \u00b7 ? help"

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

// StatusBar renders a 3-zone bottom bar: left path, center hints, right section name.
// Progressive degradation for narrow terminals:
//   - width >= 30: all three zones (left path, center hints, right section)
//   - 15 <= width < 30: left path + right section (no center hints)
//   - width < 15: right section only
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

// Render returns the styled status bar string for the given section and hints.
// Left zone: ~/section-name (terminal path style)
// Center zone: key hints
// Right zone: SECTION NAME in caps, with optional scroll position
//
// The bar progressively degrades at narrow widths:
//   - width < 15: right section only
//   - width < 30: left path + right section (no center hints)
//   - width >= 30: all three zones
func (s StatusBar) Render(section Section, hints string, scroll ScrollInfo) string {
	left := " ~/" + SectionName(section)
	right := s.buildRight(section, scroll)

	if hints == "" {
		hints = DefaultKeyHints
	}

	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)

	// Ultra-narrow: right section only.
	if s.width < 15 {
		if rightW >= s.width {
			right = truncateRuneSafe(right, s.width)
			rightW = lipgloss.Width(right)
		}
		pad := s.width - rightW
		if pad < 0 {
			pad = 0
		}
		bar := strings.Repeat(" ", pad) + right
		return s.theme.StatusBar.Render(bar)
	}

	// Narrow: left + right, no center hints.
	if s.width < 30 {
		usedByEnds := leftW + rightW
		gap := s.width - usedByEnds
		if gap < 0 {
			gap = 0
		}
		bar := left + strings.Repeat(" ", gap) + right
		return s.theme.StatusBar.Render(bar)
	}

	// Normal: all three zones.
	usedByEnds := leftW + rightW
	remaining := s.width - usedByEnds

	if remaining <= 0 {
		// No room for center hints; just render left + right.
		gap := s.width - usedByEnds
		if gap < 0 {
			gap = 0
		}
		bar := left + strings.Repeat(" ", gap) + right
		return s.theme.StatusBar.Render(bar)
	}

	// Truncate hints at rune boundaries if wider than available space.
	hintsW := lipgloss.Width(hints)
	if hintsW > remaining {
		hints = truncateRuneSafe(hints, remaining)
		hintsW = lipgloss.Width(hints)
	}

	totalPad := remaining - hintsW
	leftPad := totalPad / 2
	rightPad := totalPad - leftPad

	bar := left +
		strings.Repeat(" ", leftPad) +
		hints +
		strings.Repeat(" ", rightPad) +
		right

	return s.theme.StatusBar.Render(bar)
}

// buildRight constructs the right zone string. When the content is scrollable,
// a scroll indicator is appended after the section name:
//   - "TOP" when at the top
//   - "BOT" when at the bottom
//   - percentage (e.g., "45%") when in the middle
//
// When content fits without scrolling, only the section name is shown.
func (s StatusBar) buildRight(section Section, scroll ScrollInfo) string {
	name := strings.ToUpper(SectionName(section))
	if scroll.Fits {
		return name + " "
	}
	var indicator string
	switch {
	case scroll.AtTop:
		indicator = "TOP"
	case scroll.AtBottom:
		indicator = "BOT"
	default:
		indicator = strings.TrimSpace(scroll.Percent)
	}
	return name + " " + indicator + " "
}
