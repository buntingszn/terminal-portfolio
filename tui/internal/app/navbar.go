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

// navLabelFormat determines how section labels are rendered based on width.
type navLabelFormat int

const (
	navLabelFull    navLabelFormat = iota // "1:home"
	navLabelShort                         // "1:hm"
	navLabelNumOnly                       // "1"
)

// navShortName returns the abbreviated name for a section.
func navShortName(s Section) string {
	switch s {
	case SectionHome:
		return "hm"
	case SectionWork:
		return "wk"
	case SectionCV:
		return "cv"
	case SectionLinks:
		return "lk"
	default:
		return "?"
	}
}

// navLabelForWidth returns the label format appropriate for the given width.
func navLabelForWidth(width int) navLabelFormat {
	if width >= 40 {
		return navLabelFull
	}
	if width >= 25 {
		return navLabelShort
	}
	return navLabelNumOnly
}

// navTabLabel returns the tab label string for a section at a given format.
func navTabLabel(s Section, format navLabelFormat) string {
	num := int(s) + 1
	switch format {
	case navLabelFull:
		return fmt.Sprintf("%d:%s", num, SectionName(s))
	case navLabelShort:
		return fmt.Sprintf("%d:%s", num, navShortName(s))
	default:
		return fmt.Sprintf("%d", num)
	}
}

// View renders the navigation bar.
// Layout adapts to width:
//   - >= 40: ┌[1:home]─[2:work]─[3:cv]─[4:links]─────┐
//   - 25-39: ┌[1:hm]─[2:wk]─[3:cv]─[4:lk]────────────┐
//   - < 25:  ┌[1]─[2]─[3]─[4]─────────────────────────┐
func (n NavBar) View() string {
	accentStyle := lipgloss.NewStyle().Foreground(n.theme.Colors.Accent).Bold(true)
	mutedStyle := lipgloss.NewStyle().Foreground(n.theme.Colors.Muted)
	borderStyle := lipgloss.NewStyle().Foreground(n.theme.Colors.Border)

	format := navLabelForWidth(n.width)

	var tabs strings.Builder
	tabsLen := 0

	for i := range SectionCount {
		s := Section(i)
		label := navTabLabel(s, format)

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
