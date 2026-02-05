package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// NavBar renders a horizontal tab navigation bar using box-drawing characters.
// Active tab is styled with accent color; inactive tabs use muted color.
type NavBar struct {
	theme  Theme
	width  int
	active Section
}

// NewNavBar creates a NavBar with the given theme and terminal width.
func NewNavBar(theme Theme, width int) NavBar {
	return NavBar{
		theme: theme,
		width: width,
	}
}

// SetTheme updates the NavBar's theme.
func (n *NavBar) SetTheme(theme Theme) {
	n.theme = theme
}

// SetWidth updates the NavBar's width.
func (n *NavBar) SetWidth(width int) {
	n.width = width
}

// SetActive sets which section tab is highlighted.
func (n *NavBar) SetActive(s Section) {
	n.active = s
}

// View renders the navigation bar.
// Layout: ┌[home]─[work]─[cv]─[links]─────────────┐
func (n NavBar) View() string {
	accentStyle := lipgloss.NewStyle().Foreground(n.theme.Colors.Accent).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(n.theme.Colors.Muted)
	borderStyle := lipgloss.NewStyle().Foreground(n.theme.Colors.Border)

	var tabs strings.Builder
	tabsLen := 0

	for i := range SectionCount {
		s := Section(i)
		label := fmt.Sprintf("%d:%s", i+1, SectionName(s))

		if i > 0 {
			tabs.WriteString(borderStyle.Render(borderHorizontal))
			tabsLen++
		}

		if s == n.active {
			tabs.WriteString(accentStyle.Render("[" + label + "]"))
		} else {
			tabs.WriteString(mutedStyle.Render("[" + label + "]"))
		}
		// +2 for the brackets
		tabsLen += len(label) + 2
	}

	// Build the full bar: ┌ + tabs + fill + ┐
	// ┌ = 1, ┐ = 1
	fillLen := n.width - tabsLen - 2
	if fillLen < 0 {
		fillLen = 0
	}

	fill := strings.Repeat(borderHorizontal, fillLen)

	return borderStyle.Render(borderTopLeft) +
		tabs.String() +
		borderStyle.Render(fill+borderTopRight)
}
