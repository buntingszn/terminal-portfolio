package app

import "github.com/charmbracelet/lipgloss"

// Colors holds the 5-color palette.
type Colors struct {
	Bg     lipgloss.Color
	Fg     lipgloss.Color
	Accent lipgloss.Color
	Muted  lipgloss.Color
	Border lipgloss.Color
}

// Theme holds colors and pre-built styles.
type Theme struct {
	Colors Colors

	// Pre-built styles
	Title       lipgloss.Style
	Body        lipgloss.Style
	Accent      lipgloss.Style
	Muted       lipgloss.Style
	Border      lipgloss.Style
	StatusBar   lipgloss.Style
	NavActive   lipgloss.Style
	NavInactive lipgloss.Style
}

var darkColors = Colors{
	Bg:     lipgloss.Color("#0d0d0d"),
	Fg:     lipgloss.Color("#c8c0b8"),
	Accent: lipgloss.Color("#e8536d"),
	Muted:  lipgloss.Color("#555250"),
	Border: lipgloss.Color("#2a2826"),
}

func newTheme(colors Colors) Theme {
	return Theme{
		Colors:      colors,
		Title:       lipgloss.NewStyle().Foreground(colors.Accent).Bold(true),
		Body:        lipgloss.NewStyle().Foreground(colors.Fg),
		Accent:      lipgloss.NewStyle().Foreground(colors.Accent),
		Muted:       lipgloss.NewStyle().Foreground(colors.Muted),
		Border:      lipgloss.NewStyle().Foreground(colors.Border),
		StatusBar:   lipgloss.NewStyle().Background(colors.Border).Foreground(colors.Muted),
		NavActive:   lipgloss.NewStyle().Foreground(colors.Accent).Bold(true),
		NavInactive: lipgloss.NewStyle().Foreground(colors.Muted),
	}
}

// DarkTheme returns the dark theme.
func DarkTheme() Theme {
	return newTheme(darkColors)
}
