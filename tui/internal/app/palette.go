package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PaletteAction describes the result of a command palette invocation.
type PaletteAction int

const (
	// PaletteNone means no action (palette dismissed with no command).
	PaletteNone PaletteAction = iota
	// PaletteNavigate means navigate to the section in PaletteResultMsg.Section.
	PaletteNavigate
	// PaletteTheme means toggle the theme.
	PaletteTheme
	// PaletteQuit means quit the application.
	PaletteQuit
	// PaletteHelp means show the help overlay.
	PaletteHelp
)

// PaletteResultMsg is sent when the command palette resolves a command.
type PaletteResultMsg struct {
	Action  PaletteAction
	Section Section
}

// PaletteModel implements the command palette overlay.
type PaletteModel struct {
	visible bool
	input   string
	err     string
	theme   Theme
	width   int
}

// NewPaletteModel creates a PaletteModel with the given theme.
func NewPaletteModel(theme Theme) PaletteModel {
	return PaletteModel{
		theme: theme,
	}
}

// Open makes the palette visible and clears any previous state.
func (p *PaletteModel) Open() {
	p.visible = true
	p.input = ""
	p.err = ""
}

// Close hides the palette.
func (p *PaletteModel) Close() {
	p.visible = false
	p.input = ""
	p.err = ""
}

// Visible returns whether the palette is currently shown.
func (p *PaletteModel) Visible() bool {
	return p.visible
}

// SetTheme updates the palette's theme.
func (p *PaletteModel) SetTheme(theme Theme) {
	p.theme = theme
}

// SetWidth updates the palette's rendering width.
func (p *PaletteModel) SetWidth(width int) {
	p.width = width
}

// Update handles key input for the command palette.
func (p PaletteModel) Update(msg tea.Msg) (PaletteModel, tea.Cmd) {
	if !p.visible {
		return p, nil
	}

	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return p, nil
	}

	switch keyMsg.Type {
	case tea.KeyEscape:
		p.visible = false
		return p, func() tea.Msg {
			return PaletteResultMsg{Action: PaletteNone}
		}

	case tea.KeyEnter:
		if p.input == "" {
			// Empty enter dismisses.
			p.visible = false
			return p, func() tea.Msg {
				return PaletteResultMsg{Action: PaletteNone}
			}
		}
		return p.execute()

	case tea.KeyBackspace:
		if len(p.input) > 0 {
			p.input = p.input[:len(p.input)-1]
			p.err = ""
		}
		return p, nil

	default:
		// Append typed characters.
		s := keyMsg.String()
		if len(s) == 1 {
			p.input += s
			p.err = ""
		}
		return p, nil
	}
}

// execute resolves the current input to an action.
func (p PaletteModel) execute() (PaletteModel, tea.Cmd) {
	cmd := strings.TrimSpace(p.input)

	type commandDef struct {
		action  PaletteAction
		section Section
	}

	commands := map[string]commandDef{
		"home":  {action: PaletteNavigate, section: SectionHome},
		"work":  {action: PaletteNavigate, section: SectionWork},
		"cv":    {action: PaletteNavigate, section: SectionCV},
		"links": {action: PaletteNavigate, section: SectionLinks},
		"theme": {action: PaletteTheme},
		"quit":  {action: PaletteQuit},
		"q":     {action: PaletteQuit},
		"help":  {action: PaletteHelp},
	}

	if def, ok := commands[cmd]; ok {
		p.visible = false
		result := PaletteResultMsg{
			Action:  def.action,
			Section: def.section,
		}
		return p, func() tea.Msg { return result }
	}

	// Unknown command.
	p.err = "unknown: " + cmd
	p.input = ""
	return p, nil
}

// View renders the command palette overlay.
func (p PaletteModel) View() string {
	if !p.visible {
		return ""
	}

	fgStyle := lipgloss.NewStyle().Foreground(p.theme.Colors.Fg)
	accentStyle := lipgloss.NewStyle().Foreground(p.theme.Colors.Accent)

	prompt := accentStyle.Render(":") + fgStyle.Render(p.input) + accentStyle.Render("█")

	width := p.width
	if width < 1 {
		width = 1
	}

	// For very narrow terminals, render a simple single-line palette without box border.
	if width < 20 {
		return prompt
	}

	borderStyle := lipgloss.NewStyle().Foreground(p.theme.Colors.Border)
	mutedStyle := lipgloss.NewStyle().Foreground(p.theme.Colors.Muted)

	innerWidth := width - 4
	if innerWidth < 1 {
		innerWidth = 1
	}

	// Top border.
	topFill := innerWidth + 2
	if topFill < 0 {
		topFill = 0
	}
	top := borderStyle.Render(borderTopLeft + strings.Repeat(borderHorizontal, topFill) + borderTopRight)

	// Prompt line. Use lipgloss.Width for correct rune-aware measurement.
	promptVisualWidth := lipgloss.Width(":" + p.input + "█")
	promptPad := innerWidth - promptVisualWidth + 1
	if promptPad < 0 {
		promptPad = 0
	}
	middle := borderStyle.Render(borderVertical) + " " + prompt +
		strings.Repeat(" ", promptPad) +
		borderStyle.Render(borderVertical)

	// Bottom border.
	bottomFill := innerWidth + 2
	if bottomFill < 0 {
		bottomFill = 0
	}
	bottom := borderStyle.Render(borderBottomLeft + strings.Repeat(borderHorizontal, bottomFill) + borderBottomRight)

	// For narrow terminals (< 40), skip the hint line to save space.
	if width < 40 {
		return top + "\n" + middle + "\n" + bottom
	}

	// Error or hints line.
	var infoLine string
	if p.err != "" {
		infoLine = accentStyle.Render(p.err)
	} else {
		infoLine = mutedStyle.Render("home work cv links theme quit help")
	}
	infoPad := innerWidth - lipgloss.Width(infoLine) + 1
	if infoPad < 0 {
		infoPad = 0
	}
	info := borderStyle.Render(borderVertical) + " " + infoLine +
		strings.Repeat(" ", infoPad) +
		borderStyle.Render(borderVertical)

	return top + "\n" + middle + "\n" + info + "\n" + bottom
}
