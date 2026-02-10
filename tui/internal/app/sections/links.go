package sections

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/buntingszn/terminal-portfolio/tui/internal/app"
	"github.com/buntingszn/terminal-portfolio/tui/internal/content"
)

// clearCopyFeedbackMsg is sent after a delay to clear the copy feedback text.
type clearCopyFeedbackMsg struct{}

// LinksSection implements app.SectionModel and renders a navigable links list.
type LinksSection struct {
	content          *content.Content
	theme            app.Theme
	viewport         app.Viewport
	width            int
	height           int
	cursor           int
	focused          bool
	copyFeedback     string
	pendingClipboard string
}

// NewLinksSection creates a new LinksSection with the given content and theme.
func NewLinksSection(c *content.Content, theme app.Theme) *LinksSection {
	return &LinksSection{
		content:  c,
		theme:    theme,
		viewport: app.NewViewport(0, 0),
	}
}

// Init implements app.SectionModel.
func (l *LinksSection) Init() tea.Cmd {
	return nil
}

// Update implements app.SectionModel.
func (l *LinksSection) Update(msg tea.Msg) (app.SectionModel, tea.Cmd) {
	// Clear pending clipboard after each render cycle so the OSC 52
	// sequence is emitted exactly once.
	l.pendingClipboard = ""

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		l.width = msg.Width
		l.height = msg.Height
		l.viewport.SetSize(l.width, l.height)
		l.viewport.SetContentPreserveScroll(l.renderContent())

	case tea.KeyMsg:
		if !l.focused {
			break
		}
		switch msg.String() {
		case "j", "down":
			l.moveCursor(1)
		case "k", "up":
			l.moveCursor(-1)
		case "g", "home":
			l.cursor = 0
			l.viewport.SetContent(l.renderContent())
			l.viewport.ScrollToTop()
		case "G", "end":
			if l.content != nil && len(l.content.Links.Links) > 0 {
				l.cursor = len(l.content.Links.Links) - 1
			}
			l.viewport.SetContent(l.renderContent())
			l.viewport.ScrollToBottom()
		case "enter":
			if l.content != nil && l.cursor < len(l.content.Links.Links) {
				url := l.content.Links.Links[l.cursor].URL
				if url == "" {
					break
				}
				l.pendingClipboard = app.OSC52Sequence(url)
				l.copyFeedback = "Copied!"
				l.viewport.SetContent(l.renderContent())
				return l, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
					return clearCopyFeedbackMsg{}
				})
			}
		case "pgup":
			l.viewport.ScrollUp(l.viewport.VisibleLines())
		case "pgdown":
			l.viewport.ScrollDown(l.viewport.VisibleLines())
		case "ctrl+u":
			l.viewport.ScrollUp(l.viewport.VisibleLines() / 2)
		case "ctrl+d":
			l.viewport.ScrollDown(l.viewport.VisibleLines() / 2)
		}

	case clearCopyFeedbackMsg:
		l.copyFeedback = ""
		l.viewport.SetContent(l.renderContent())

	case tea.MouseMsg:
		if !l.focused {
			break
		}
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			l.moveCursor(-1)
		case tea.MouseButtonWheelDown:
			l.moveCursor(1)
		}

	case app.FocusMsg:
		l.focused = true
		l.cursor = 0
		l.viewport.SetContent(l.renderContent())
		l.viewport.ScrollToTop()
		return l, nil

	case app.BlurMsg:
		l.focused = false
	}

	return l, nil
}

// View implements app.SectionModel.
func (l *LinksSection) View() string {
	return l.pendingClipboard + l.viewport.ViewWithScrollbar(l.theme)
}

// ScrollInfo implements app.ScrollReporter for the status bar scroll indicator.
func (l *LinksSection) ScrollInfo() app.ScrollInfo {
	return l.viewport.GetScrollInfo()
}

// KeyHints implements app.KeyHinter for contextual status bar hints.
func (l *LinksSection) KeyHints() string {
	if l.copyFeedback != "" {
		return l.copyFeedback
	}
	return "j/k navigate " + app.BorderVertical + " enter copy URL " + app.BorderVertical + " 1-4 nav " + app.BorderVertical + " ? help"
}

// linesPerLink is the number of rendered lines each link entry occupies
// (label line + blank separator line).
const linesPerLink = 2

// topPadLines is the number of lines consumed by the top padding.
const topPadLines = 1

// moveCursor moves the selection cursor by delta and re-renders.
func (l *LinksSection) moveCursor(delta int) {
	if l.content == nil || len(l.content.Links.Links) == 0 {
		return
	}

	count := len(l.content.Links.Links)
	l.cursor += delta

	if l.cursor < 0 {
		l.cursor = 0
	}
	if l.cursor >= count {
		l.cursor = count - 1
	}

	l.viewport.SetContent(l.renderContent())

	// Scroll the viewport so the selected link stays visible.
	targetLine := topPadLines + l.cursor*linesPerLink
	totalLines := l.viewport.TotalLines()
	visibleLines := l.viewport.VisibleLines()

	if visibleLines > 0 && totalLines > visibleLines {
		l.viewport.ScrollToTop()
		if targetLine > 0 {
			l.viewport.ScrollDown(targetLine)
		}
	}
}

// renderContent builds the full rendered text for the viewport.
func (l *LinksSection) renderContent() string {
	if l.content == nil {
		return l.theme.Muted.Render("No links loaded.")
	}

	links := l.content.Links.Links
	if len(links) == 0 {
		return l.theme.Muted.Render("No links to display.")
	}

	var b strings.Builder

	// Top padding.
	b.WriteByte('\n')

	// Maximum width for label/text.
	maxTextWidth := l.viewport.ContentWidth() - 2
	if maxTextWidth < 0 {
		maxTextWidth = 0
	}

	for i, link := range links {
		selected := i == l.cursor

		// Truncate label if terminal is very narrow.
		label := link.Label
		if maxTextWidth > 0 && lipgloss.Width(label) > maxTextWidth {
			label = app.TruncateWithEllipsis(label, maxTextWidth)
		}

		// Build the line: prefix + label + display text/URL.
		var line strings.Builder
		if selected {
			line.WriteString(l.theme.Accent.Render("> "))
			line.WriteString(l.theme.Accent.Render(label))
		} else {
			line.WriteString("  ")
			line.WriteString(l.theme.Body.Render(label))
		}

		// Show link.Text when available, otherwise URL.
		displayText := link.Text
		if displayText == "" {
			displayText = link.URL
		}
		if maxTextWidth > 0 {
			displayText = app.TruncateWithEllipsis(displayText, maxTextWidth)
		}

		// Append display text in muted style with spacing.
		line.WriteString("  ")
		line.WriteString(app.RenderHyperlink(link.URL, l.theme.Muted.Render(displayText)))

		b.WriteString(line.String())

		// Blank line separator between entries, but not after the last one.
		if i < len(links)-1 {
			b.WriteString("\n\n")
		}
	}

	return app.PadLinesToWidth(b.String(), l.viewport.ContentWidth())
}
