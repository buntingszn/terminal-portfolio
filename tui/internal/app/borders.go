package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Box-drawing characters for straight borders.
const (
	borderTopLeft     = "┌"
	borderTopRight    = "┐"
	borderBottomLeft  = "└"
	borderBottomRight = "┘"
	borderHorizontal  = "─"
	borderVertical    = "│"
)

// Exported box-drawing constants for use by section renderers.
const (
	BorderTopLeft     = borderTopLeft
	BorderTopRight    = borderTopRight
	BorderBottomLeft  = borderBottomLeft
	BorderBottomRight = borderBottomRight
	BorderHorizontal  = borderHorizontal
	BorderVertical    = borderVertical
)

// RenderCard renders content inside a bordered card with a title embedded in
// the top border. The border uses the theme's border color and the title is
// rendered in the accent color.
//
// Layout:
//
//	┌─ Title ──────────┐
//	│ content line 1    │
//	│ content line 2    │
//	└──────────────────┘
//
// If width < 10, returns content without any border decoration.
func RenderCard(theme Theme, title, content string, width int) string {
	if width < 10 {
		return content
	}

	borderStyle := lipgloss.NewStyle().Foreground(theme.Colors.Border)
	accentStyle := lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	// Inner width is total width minus two border columns and two padding spaces.
	innerWidth := width - 4

	// Truncate title if it would overflow the top border line.
	// Top border layout: "┌─ " (3) + title + " " (1) + "─" (min 1) + "┐" (1) = 6 overhead.
	maxTitleLen := width - 6
	if maxTitleLen < 0 {
		maxTitleLen = 0
	}
	displayTitle := title
	titleRunes := []rune(title)
	titleLen := len(titleRunes)
	if titleLen > maxTitleLen {
		if maxTitleLen > 3 {
			displayTitle = string(titleRunes[:maxTitleLen-3]) + "..."
			titleLen = maxTitleLen
		} else if maxTitleLen > 0 {
			displayTitle = string([]rune("...")[:maxTitleLen])
			titleLen = maxTitleLen
		} else {
			displayTitle = ""
			titleLen = 0
		}
	}

	// Build top border: ┌─ Title ───...───┐
	styledTitle := accentStyle.Render(displayTitle)
	// "┌─ " = 3 chars, then title, then " ─...─┐"
	topLineRemain := width - 3 - titleLen - 1 - 1 // 3 for "┌─ ", 1 for " ", 1 for "┐"
	if topLineRemain < 1 {
		topLineRemain = 1
	}
	topBorder := borderStyle.Render(borderTopLeft+borderHorizontal+" ") +
		styledTitle +
		borderStyle.Render(" "+strings.Repeat(borderHorizontal, topLineRemain)+borderTopRight)

	// Build bottom border: └───...───┘
	bottomLineWidth := width - 2 // subtract └ and ┘
	if bottomLineWidth < 0 {
		bottomLineWidth = 0
	}
	bottomBorder := borderStyle.Render(borderBottomLeft + strings.Repeat(borderHorizontal, bottomLineWidth) + borderBottomRight)

	// Wrap and render content lines.
	lines := WrapText(content, innerWidth)
	styledBorderV := borderStyle.Render(borderVertical)

	var body strings.Builder
	for _, line := range lines {
		padded := padRight(line, innerWidth)
		body.WriteString(styledBorderV + " " + padded + " " + styledBorderV + "\n")
	}

	return topBorder + "\n" + body.String() + bottomBorder
}

// RenderDivider renders a horizontal rule spanning the given width using the
// ─ character, styled in the theme's border color.
func RenderDivider(theme Theme, width int) string {
	if width <= 0 {
		return ""
	}
	borderStyle := lipgloss.NewStyle().Foreground(theme.Colors.Border)
	return borderStyle.Render(strings.Repeat(borderHorizontal, width))
}

// WrapText wraps text to fit within the given width, breaking at word
// boundaries. Lines containing a single word longer than width are kept
// intact (not split mid-word).
func WrapText(text string, width int) []string {
	if width <= 0 {
		return []string{text}
	}

	var result []string
	for _, paragraph := range strings.Split(text, "\n") {
		if paragraph == "" {
			result = append(result, "")
			continue
		}
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			result = append(result, "")
			continue
		}

		var line strings.Builder
		lineLen := 0

		for _, word := range words {
			wordLen := lipgloss.Width(word)

			if lineLen == 0 {
				line.WriteString(word)
				lineLen = wordLen
				continue
			}

			if lineLen+1+wordLen > width {
				result = append(result, line.String())
				line.Reset()
				line.WriteString(word)
				lineLen = wordLen
			} else {
				line.WriteString(" ")
				line.WriteString(word)
				lineLen += 1 + wordLen
			}
		}

		if lineLen > 0 {
			result = append(result, line.String())
		}
	}

	return result
}

// padRight pads a string with spaces on the right to reach the desired visual
// width. Uses lipgloss.Width for correct measurement of strings containing
// ANSI escape codes or wide Unicode characters.
func padRight(s string, width int) string {
	sLen := lipgloss.Width(s)
	if sLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-sLen)
}

// PadRight pads a string with trailing spaces to reach the desired visual width.
// Uses lipgloss.Width for ANSI-aware measurement.
func PadRight(s string, width int) string {
	return padRight(s, width)
}

// TruncateWithEllipsis truncates a string to fit within maxWidth visual columns,
// appending an ellipsis ("...") if truncation is needed. Uses lipgloss.Width for
// ANSI-aware measurement.
func TruncateWithEllipsis(s string, maxWidth int) string {
	if maxWidth <= 3 || lipgloss.Width(s) <= maxWidth {
		return s
	}
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		candidate := string(runes[:i]) + "..."
		if lipgloss.Width(candidate) <= maxWidth {
			return candidate
		}
	}
	return "..."
}

// PadLinesToWidth pads every line in content to targetWidth visual columns,
// ensuring consistent horizontal centering when the viewport centers per-line.
func PadLinesToWidth(content string, targetWidth int) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = padRight(line, targetWidth)
	}
	return strings.Join(lines, "\n")
}
