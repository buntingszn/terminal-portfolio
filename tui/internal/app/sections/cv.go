package sections

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/buntingszn/terminal-portfolio/tui/internal/app"
	"github.com/buntingszn/terminal-portfolio/tui/internal/content"
)

// CVSection implements app.SectionModel to render CV data in a single-column
// text layout: accent name header, contact info, summary, experience with
// reverse-video dividers, skills, and education.
type CVSection struct {
	content *content.Content
	theme   app.Theme
	viewport app.Viewport
	width   int
	height  int
	focused bool
}

// NewCVSection creates a new CVSection with the given content and theme.
func NewCVSection(c *content.Content, theme app.Theme) *CVSection {
	return &CVSection{
		content: c,
		theme:   theme,
	}
}

// Init implements app.SectionModel.
func (s *CVSection) Init() tea.Cmd {
	return nil
}

// Update implements app.SectionModel.
func (s *CVSection) Update(msg tea.Msg) (app.SectionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.width = msg.Width
		s.height = msg.Height
		s.viewport.SetSize(s.width, s.height)
		s.viewport.SetContentPreserveScroll(s.renderContent())

	case tea.KeyMsg:
		if !s.focused {
			break
		}
		switch msg.String() {
		case "j", "down":
			s.viewport.ScrollDown(1)
		case "k", "up":
			s.viewport.ScrollUp(1)
		case "g", "home":
			s.viewport.ScrollToTop()
		case "G", "end":
			s.viewport.ScrollToBottom()
		case "pgup":
			s.viewport.ScrollUp(s.viewport.VisibleLines())
		case "pgdown":
			s.viewport.ScrollDown(s.viewport.VisibleLines())
		case "ctrl+u":
			s.viewport.ScrollUp(s.viewport.VisibleLines() / 2)
		case "ctrl+d":
			s.viewport.ScrollDown(s.viewport.VisibleLines() / 2)
		}

	case tea.MouseMsg:
		if !s.focused {
			break
		}
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			s.viewport.ScrollUp(3)
		case tea.MouseButtonWheelDown:
			s.viewport.ScrollDown(3)
		}

	case app.FocusMsg:
		s.focused = true
		s.viewport.ScrollToTop()
		return s, nil

	case app.BlurMsg:
		s.focused = false
	}
	return s, nil
}

// View implements app.SectionModel.
func (s *CVSection) View() string {
	return s.viewport.ViewWithScrollbar(s.theme)
}

// ScrollInfo implements app.ScrollReporter for the status bar scroll indicator.
func (s *CVSection) ScrollInfo() app.ScrollInfo {
	return s.viewport.GetScrollInfo()
}

// KeyHints implements app.KeyHinter.
func (s *CVSection) KeyHints() string {
	return "j/k scroll " + app.BorderVertical + " pgup/dn page " + app.BorderVertical + " ^u/^d half " + app.BorderVertical + " 1-4 nav " + app.BorderVertical + " ? help"
}

// sectionDivider renders a reverse-video section heading: accent background, bg foreground.
func (s *CVSection) sectionDivider(title string) string {
	style := lipgloss.NewStyle().
		Background(s.theme.Colors.Accent).
		Foreground(s.theme.Colors.Bg).
		Bold(true)
	return style.Render(" " + title + " ")
}

// renderContent builds the full single-column text layout.
func (s *CVSection) renderContent() string {
	cv := s.content.CV
	meta := s.content.Meta
	bodyStyle := s.theme.Body
	mutedStyle := s.theme.Muted

	contentWidth := s.viewport.ContentWidth()
	if contentWidth < 10 {
		contentWidth = 10
	}

	density := app.DensityForHeight(s.height)
	sep := app.SectionSeparator(density)

	var sections []string

	// Header: name in accent+bold.
	nameStyle := lipgloss.NewStyle().Foreground(s.theme.Colors.Accent).Bold(true)
	sections = append(sections, nameStyle.Render(meta.Name))

	// Contact line: email · location in muted.
	var contactParts []string
	if cv.Contact.Email != "" {
		emailLink := app.RenderHyperlink("mailto:"+cv.Contact.Email, mutedStyle.Render(cv.Contact.Email))
		contactParts = append(contactParts, emailLink)
	}
	if cv.Contact.Location != "" {
		contactParts = append(contactParts, mutedStyle.Render(cv.Contact.Location))
	}
	if len(contactParts) > 0 {
		sections = append(sections, strings.Join(contactParts, mutedStyle.Render(" · ")))
	}

	// Summary.
	if cv.Summary != "" {
		dividerWidth := contentWidth - 2
		if dividerWidth < 10 {
			dividerWidth = 10
		}
		wrapped := app.WrapText(cv.Summary, dividerWidth)
		sections = append(sections, bodyStyle.Render(strings.Join(wrapped, "\n")))
	}

	sections = append(sections, s.renderExperience(contentWidth))
	sections = append(sections, s.renderSkills(contentWidth))
	sections = append(sections, s.renderEducation())

	return app.PadLinesToWidth("\n"+strings.Join(sections, sep), contentWidth)
}

// renderExperience builds the experience block with reverse-video divider.
func (s *CVSection) renderExperience(contentWidth int) string {
	accentStyle := lipgloss.NewStyle().Foreground(s.theme.Colors.Accent).Bold(true)
	bodyStyle := s.theme.Body
	mutedStyle := s.theme.Muted

	var b strings.Builder
	b.WriteByte('\n')
	b.WriteString(s.sectionDivider("EXPERIENCE"))
	b.WriteString("\n\n")

	for i, exp := range s.content.CV.Experience {
		dateRange := exp.Start
		if exp.End != "" {
			dateRange += " - " + exp.End
		}

		// Role @ Company  date
		rolePart := accentStyle.Render(exp.Role)
		companyPart := mutedStyle.Render(" @ " + exp.Company)
		datePart := accentStyle.Render(dateRange)

		// Right-align date.
		leftContent := "  " + rolePart + companyPart
		leftWidth := lipgloss.Width(leftContent)
		dateWidth := lipgloss.Width(datePart)
		gap := contentWidth - leftWidth - dateWidth
		if gap < 2 {
			gap = 2
		}
		b.WriteString(leftContent + strings.Repeat(" ", gap) + datePart)
		b.WriteByte('\n')

		for _, bullet := range exp.Bullets {
			wrapped := app.WrapText(bullet, contentWidth-6)
			for j, line := range wrapped {
				if j == 0 {
					b.WriteString("    " + bodyStyle.Render("- "+line))
				} else {
					b.WriteString("      " + bodyStyle.Render(line))
				}
				b.WriteByte('\n')
			}
		}
		if i < len(s.content.CV.Experience)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// renderSkills builds the skills block with aligned categories.
func (s *CVSection) renderSkills(contentWidth int) string {
	accentStyle := s.theme.Accent
	bodyStyle := s.theme.Body

	var b strings.Builder
	b.WriteString(s.sectionDivider("SKILLS"))
	b.WriteString("\n\n")

	maxCatLen := 0
	for _, sk := range s.content.CV.Skills {
		if len(sk.Category) > maxCatLen {
			maxCatLen = len(sk.Category)
		}
	}

	for _, sk := range s.content.CV.Skills {
		padded := fmt.Sprintf("%-*s", maxCatLen, sk.Category)
		skillsStr := strings.Join(sk.Items, ", ")
		availWidth := contentWidth - maxCatLen - 4
		if len(skillsStr) > availWidth && availWidth > 10 {
			wrapped := app.WrapText(skillsStr, availWidth)
			for j, line := range wrapped {
				if j == 0 {
					b.WriteString("  " + accentStyle.Render(padded) + bodyStyle.Render("  "+line))
				} else {
					b.WriteString(strings.Repeat(" ", maxCatLen+4) + bodyStyle.Render(line))
				}
				b.WriteByte('\n')
			}
		} else {
			b.WriteString("  " + accentStyle.Render(padded) + bodyStyle.Render("  "+skillsStr))
			b.WriteByte('\n')
		}
	}
	return b.String()
}

// renderEducation builds the education block.
func (s *CVSection) renderEducation() string {
	education := s.content.CV.Education
	if len(education) == 0 {
		return ""
	}

	accentStyle := lipgloss.NewStyle().Foreground(s.theme.Colors.Accent).Bold(true)
	mutedStyle := s.theme.Muted

	var b strings.Builder
	b.WriteString(s.sectionDivider("EDUCATION"))
	b.WriteString("\n\n")

	for i, edu := range education {
		// Degree @ Institution  year
		degreePart := accentStyle.Render(edu.Degree)
		instPart := mutedStyle.Render(" @ " + edu.Institution)
		yearPart := mutedStyle.Render(edu.Year)
		b.WriteString("  " + degreePart + instPart + "  " + yearPart)
		if i < len(education)-1 {
			b.WriteByte('\n')
		}
	}
	b.WriteByte('\n')

	return b.String()
}
