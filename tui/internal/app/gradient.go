package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderGradientText applies a static color gradient across each character
// of the text, interpolating in Lab color space from start to end color.
// The result is bold.
func RenderGradientText(text string, startColor, endColor lipgloss.Color) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	c1, err1 := HexToColorful(startColor)
	c2, err2 := HexToColorful(endColor)
	if err1 != nil || err2 != nil {
		// Fallback: render without gradient.
		return lipgloss.NewStyle().Foreground(startColor).Bold(true).Render(text)
	}

	var b strings.Builder
	b.Grow(len(text) * 20) // ANSI escape codes expand each character

	last := len(runes) - 1
	if last == 0 {
		last = 1
	}

	for i, r := range runes {
		t := float64(i) / float64(last)
		blended := c1.BlendLab(c2, t)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(blended.Hex())).Bold(true)
		b.WriteString(style.Render(string(r)))
	}

	return b.String()
}
