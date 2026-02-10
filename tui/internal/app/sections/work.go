package sections

import (
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/buntingszn/terminal-portfolio/tui/internal/app"
	"github.com/buntingszn/terminal-portfolio/tui/internal/content"
)

// clearWorkCopyMsg is sent after a delay to clear the copy feedback text.
type clearWorkCopyMsg struct{}

// WorkSection displays the projects list sorted featured-first.
type WorkSection struct {
	content          *content.Content
	theme            app.Theme
	viewport         app.Viewport
	width            int
	height           int
	focused          bool
	cursor           int
	copyFeedback     string
	pendingClipboard string
	projectOffsets   []int    // line offset for each project in rendered content
	projectURLs      []string // URL for each project (URL or Repo)
}

// NewWorkSection creates a new work section from the loaded content.
func NewWorkSection(c *content.Content, theme app.Theme) *WorkSection {
	return &WorkSection{
		content: c,
		theme:   theme,
	}
}

// Init implements app.SectionModel.
func (w *WorkSection) Init() tea.Cmd {
	return nil
}

// Update implements app.SectionModel.
func (w *WorkSection) Update(msg tea.Msg) (app.SectionModel, tea.Cmd) {
	// Clear pending clipboard after each render cycle so the OSC 52
	// sequence is emitted exactly once.
	w.pendingClipboard = ""

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w.width = msg.Width
		w.height = msg.Height
		w.viewport.SetSize(w.width, w.height)
		w.viewport.SetContentPreserveScroll(w.renderContent())
		return w, nil

	case tea.KeyMsg:
		if !w.focused {
			return w, nil
		}
		switch msg.String() {
		case "j", "down":
			w.moveCursor(1)
			return w, nil
		case "k", "up":
			w.moveCursor(-1)
			return w, nil
		case "g", "home":
			w.cursor = 0
			w.viewport.SetContent(w.renderContent())
			w.viewport.ScrollToTop()
			return w, nil
		case "G", "end":
			if len(w.projectURLs) > 0 {
				w.cursor = len(w.projectURLs) - 1
			}
			w.viewport.SetContent(w.renderContent())
			w.viewport.ScrollToBottom()
			return w, nil
		case "enter":
			if len(w.projectURLs) > 0 && w.cursor < len(w.projectURLs) {
				url := w.projectURLs[w.cursor]
				if url == "" {
					break
				}
				w.pendingClipboard = app.OSC52Sequence(url)
				w.copyFeedback = "Copied!"
				w.viewport.SetContent(w.renderContent())
				return w, tea.Tick(2*time.Second, func(time.Time) tea.Msg {
					return clearWorkCopyMsg{}
				})
			}
		case "pgup":
			w.viewport.ScrollUp(w.viewport.VisibleLines())
			return w, nil
		case "pgdown":
			w.viewport.ScrollDown(w.viewport.VisibleLines())
			return w, nil
		case "ctrl+u":
			w.viewport.ScrollUp(w.viewport.VisibleLines() / 2)
			return w, nil
		case "ctrl+d":
			w.viewport.ScrollDown(w.viewport.VisibleLines() / 2)
			return w, nil
		}

	case clearWorkCopyMsg:
		w.copyFeedback = ""
		w.viewport.SetContent(w.renderContent())
		return w, nil

	case tea.MouseMsg:
		if !w.focused {
			return w, nil
		}
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			w.moveCursor(-1)
		case tea.MouseButtonWheelDown:
			w.moveCursor(1)
		}
		return w, nil

	case app.FocusMsg:
		w.focused = true
		w.cursor = 0
		w.viewport.SetContent(w.renderContent())
		w.viewport.ScrollToTop()
		return w, nil

	case app.BlurMsg:
		w.focused = false
		return w, nil
	}

	return w, nil
}

// View implements app.SectionModel.
func (w *WorkSection) View() string {
	return w.pendingClipboard + w.viewport.ViewWithScrollbar(w.theme)
}

// ScrollInfo implements app.ScrollReporter for the status bar scroll indicator.
func (w *WorkSection) ScrollInfo() app.ScrollInfo {
	return w.viewport.GetScrollInfo()
}

// KeyHints implements app.KeyHinter for contextual status bar hints.
func (w *WorkSection) KeyHints() string {
	if w.copyFeedback != "" {
		return w.copyFeedback
	}
	return "j/k navigate " + app.BorderVertical + " enter copy URL " + app.BorderVertical + " 1-4 nav " + app.BorderVertical + " ? help"
}

// moveCursor moves the selection cursor by delta and re-renders.
func (w *WorkSection) moveCursor(delta int) {
	if len(w.projectURLs) == 0 {
		return
	}

	w.cursor += delta
	if w.cursor < 0 {
		w.cursor = 0
	}
	if w.cursor >= len(w.projectURLs) {
		w.cursor = len(w.projectURLs) - 1
	}

	w.viewport.SetContent(w.renderContent())

	// Scroll the viewport so the selected project stays visible.
	if w.cursor < len(w.projectOffsets) {
		targetLine := w.projectOffsets[w.cursor]
		totalLines := w.viewport.TotalLines()
		visibleLines := w.viewport.VisibleLines()

		if visibleLines > 0 && totalLines > visibleLines {
			w.viewport.ScrollToTop()
			if targetLine > 0 {
				w.viewport.ScrollDown(targetLine)
			}
		}
	}
}

// sortedProjects returns a copy of projects sorted featured-first (stable).
func sortedProjects(projects []content.WorkProject) []content.WorkProject {
	sorted := make([]content.WorkProject, len(projects))
	copy(sorted, projects)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Featured == sorted[j].Featured {
			return false
		}
		return sorted[i].Featured
	})
	return sorted
}

// renderContent builds the full rendered text for the viewport.
func (w *WorkSection) renderContent() string {
	if w.content == nil {
		return w.theme.Muted.Render("No projects loaded.")
	}

	projects := sortedProjects(w.content.Work.Projects)
	if len(projects) == 0 {
		return w.theme.Muted.Render("No projects to display.")
	}

	contentWidth := w.viewport.ContentWidth()
	if contentWidth > 78 {
		contentWidth = 78
	}
	if contentWidth < 10 {
		contentWidth = 10
	}

	var b strings.Builder

	// Reset tracking slices.
	w.projectOffsets = nil
	w.projectURLs = nil
	lineCount := 0

	countLines := func(s string) int {
		if s == "" {
			return 0
		}
		return strings.Count(s, "\n") + 1
	}

	// Top padding.
	b.WriteByte('\n')
	lineCount++

	for i, p := range projects {
		w.projectOffsets = append(w.projectOffsets, lineCount)
		url := p.URL
		if url == "" {
			url = p.Repo
		}
		w.projectURLs = append(w.projectURLs, url)

		selected := i == w.cursor
		rendered := w.renderProjectInline(p, contentWidth, selected)
		b.WriteString(rendered)
		lineCount += countLines(rendered)

		if i < len(projects)-1 {
			b.WriteString("\n\n")
			lineCount += 2
		}
	}

	return app.PadLinesToWidth(b.String(), contentWidth)
}

// renderProjectInline formats a single project: title → description → tags.
func (w *WorkSection) renderProjectInline(p content.WorkProject, width int, selected bool) string {
	accentStyle := w.theme.Accent
	bodyStyle := w.theme.Body
	mutedStyle := w.theme.Muted

	var lines []string

	// Selection prefix.
	prefix := "  "
	if selected {
		prefix = accentStyle.Render("▸") + " "
	}
	title := prefix + accentStyle.Render(p.Title)
	lines = append(lines, title)

	// Indent for sub-lines (description, tags, URL).
	indent := "    "

	// Description: word-wrapped with indent.
	if p.Description != "" {
		descWidth := width - len(indent)
		if descWidth < 10 {
			descWidth = 10
		}
		wrapped := app.WrapText(p.Description, descWidth)
		for _, wl := range wrapped {
			lines = append(lines, indent+bodyStyle.Render(wl))
		}
	}

	// Tags: rendered below description in muted style.
	if len(p.Tags) > 0 {
		tagStr := mutedStyle.Render(strings.Join(p.Tags, " · "))
		lines = append(lines, indent+tagStr)
	}

	// URL: indented, OSC 8 hyperlink, muted.
	if p.URL != "" {
		url := app.TruncateWithEllipsis(p.URL, width-len(indent))
		lines = append(lines, indent+app.RenderHyperlink(p.URL, mutedStyle.Render(url)))
	}

	// Repo: indented, OSC 8 hyperlink, muted (only if different from URL).
	if p.Repo != "" && p.Repo != p.URL {
		repo := app.TruncateWithEllipsis(p.Repo, width-len(indent))
		lines = append(lines, indent+app.RenderHyperlink(p.Repo, mutedStyle.Render(repo)))
	}

	return strings.Join(lines, "\n")
}
