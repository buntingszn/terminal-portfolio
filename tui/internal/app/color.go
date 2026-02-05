package app

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// HexToColorful converts a lipgloss.Color (hex string) to a go-colorful Color.
func HexToColorful(c lipgloss.Color) (colorful.Color, error) {
	return colorful.Hex(string(c))
}
