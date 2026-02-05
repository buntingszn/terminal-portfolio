package app

import (
	"strings"
	"unicode/utf8"

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

	// Build top border: ┌─ Title ───...───┐
	styledTitle := accentStyle.Render(title)
	titleLen := utf8.RuneCountInString(title)
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
	lines := wrapText(content, innerWidth)
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

// wrapText wraps text to fit within the given width, breaking at word
// boundaries. Lines containing a single word longer than width are kept
// intact (not split mid-word).
func wrapText(text string, width int) []string {
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
			wordLen := utf8.RuneCountInString(word)

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

// padRight pads a string with spaces on the right to reach the desired width.
func padRight(s string, width int) string {
	sLen := utf8.RuneCountInString(s)
	if sLen >= width {
		return s
	}
	return s + strings.Repeat(" ", width-sLen)
}
