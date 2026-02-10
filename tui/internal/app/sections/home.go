package sections

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/buntingszn/terminal-portfolio/tui/internal/app"
	"github.com/buntingszn/terminal-portfolio/tui/internal/content"
)

// portraitMinWidth is the minimum terminal width needed to show the ASCII
// portrait next to the bio text.
const portraitMinWidth = 80

// scrollStep is how many lines to scroll per key press.
const scrollStep = 3

const (
	// revealLinesPerTick is how many content lines to reveal each tick.
	revealLinesPerTick = 1

	// revealTickInterval is the delay between line-reveal ticks, producing
	// a streaming/typewriter effect when the home section first appears.
	revealTickInterval = 120 * time.Millisecond
)

// homeRevealTickMsg advances the line-by-line reveal animation.
type homeRevealTickMsg struct{}

// homeRevealTick schedules the next reveal tick.
func homeRevealTick() tea.Cmd {
	return tea.Tick(revealTickInterval, func(_ time.Time) tea.Msg {
		return homeRevealTickMsg{}
	})
}

// portrait is a Braille halftone developer portrait shown beside the bio text.
// Generated from a headshot photo using scripts/img2braille.py with Atkinson
// dithering and CLAHE preprocessing for facial feature preservation.
const portrait = "" +
	"⣿⣿⣿⢿⣿⣿⣿⠿⠿⣟⡻⢿⣿⡿⢿⣿⣽⣻⣿⣻⢬⣹\n" +
	"⣿⡟⡼⢠⣈⡵⣉⠔⡒⠂⠡⢨⡉⣛⢫⡛⢷⣯⢿⡝⡧⢸\n" +
	"⣿⣎⡝⣻⠿⡖⠉⣄⣤⣤⣤⣄⣁⣙⠠⠽⢾⣯⣿⠗⡡⢻\n" +
	"⣿⣿⣞⡴⣾⠁⣰⣿⡿⣛⢿⡿⠟⣿⣷⡄⠺⣿⣯⡝⡆⢹\n" +
	"⣿⣿⣿⣿⣿⡆⣿⣻⣵⣼⣫⣾⡵⣌⣿⣷⣸⣿⣳⢿⡌⣹\n" +
	"⢸⣿⣿⣿⣿⠇⢛⡉⠙⡒⠿⠯⠙⢉⠉⣛⠿⢿⣿⣯⡓⢸\n" +
	"⢘⣿⣿⣿⣏⡀⣿⣭⠁⡘⣰⣆⠁⠉⣽⣿⠀⣿⡟⢦⠁⢸\n" +
	"⠈⣿⢷⣻⣿⠄⣮⣙⣉⢥⡟⣿⡮⡙⣛⣵⣀⣿⡿⢎⠛⠺\n" +
	"⢠⢻⡷⣎⡿⠷⣼⣿⡧⠜⠣⠞⢣⢈⣿⣿⠘⣿⠆⢤⣷⢒\n" +
	"⢰⣿⣿⠟⡡⠂⢸⣯⢀⣴⣤⣤⣦⠀⢻⡏⢀⣿⡧⠘⣿⣯\n" +
	"⢹⣿⢩⢾⠀⢀⡈⡷⢾⢿⣉⣩⣿⣥⠟⠀⣰⡏⢉⢉⠛⠿\n" +
	"⣺⡏⢱⣻⣀⠈⢧⣿⣮⡟⣁⠇⠛⣣⠆⡀⠹⣿⣦⣄⡉⠒\n" +
	"⡷⠌⠀⣹⡻⣦⡈⣿⢽⣿⢷⣖⣾⠟⠁⢀⣼⡿⢏⡛⠻⣷\n" +
	"⣴⣾⡟⠁⣷⣌⠻⠌⠛⢿⠟⠚⢋⣐⣿⠿⡏⣀⠘⢿⡳⢮"

// HomeSection implements app.SectionModel and renders the bio/about view.
type HomeSection struct {
	content        *content.Content
	theme          app.Theme
	viewport       app.Viewport
	portraitShimmer app.Shimmer
	width          int
	height         int
	focused        bool
	revealLines    int  // number of lines currently visible during reveal
	revealDone     bool // true when reveal animation is complete
	hasRevealed    bool // true after first reveal finishes (prevents replay)
}

// NewHomeSection creates a new HomeSection with the given content and theme.
func NewHomeSection(c *content.Content, theme app.Theme) *HomeSection {
	return &HomeSection{
		content:        c,
		theme:          theme,
		viewport:       app.NewViewport(0, 0),
		portraitShimmer: app.NewShimmer("portrait-shimmer", theme),
		revealDone:     true, // safe default until first FocusMsg
	}
}

// Init implements app.SectionModel.
func (h *HomeSection) Init() tea.Cmd {
	return nil
}

// Update implements app.SectionModel.
func (h *HomeSection) Update(msg tea.Msg) (app.SectionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h.width = msg.Width
		h.height = msg.Height
		h.viewport.SetSize(h.width, h.height)
		h.viewport.SetContentPreserveScroll(h.buildContent())

	case tea.KeyMsg:
		if !h.focused {
			break
		}
		h.completeReveal()
		switch msg.String() {
		case "j", "down":
			h.viewport.ScrollDown(scrollStep)
		case "k", "up":
			h.viewport.ScrollUp(scrollStep)
		case "g", "home":
			h.viewport.ScrollToTop()
		case "G", "end":
			h.viewport.ScrollToBottom()
		case "pgup":
			h.viewport.ScrollUp(h.viewport.VisibleLines())
		case "pgdown":
			h.viewport.ScrollDown(h.viewport.VisibleLines())
		case "ctrl+u":
			h.viewport.ScrollUp(h.viewport.VisibleLines() / 2)
		case "ctrl+d":
			h.viewport.ScrollDown(h.viewport.VisibleLines() / 2)
		}

	case tea.MouseMsg:
		if !h.focused {
			break
		}
		h.completeReveal()
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			h.viewport.ScrollUp(scrollStep)
		case tea.MouseButtonWheelDown:
			h.viewport.ScrollDown(scrollStep)
		}

	case app.FocusMsg:
		h.focused = true
		h.viewport.ScrollToTop()
		cmds := []tea.Cmd{h.portraitShimmer.Start()}
		if !h.hasRevealed {
			h.revealLines = 1
			h.revealDone = false
			h.viewport.SetContent(h.buildContent())
			cmds = append(cmds, homeRevealTick())
		}
		return h, tea.Batch(cmds...)

	case app.BlurMsg:
		h.focused = false
		h.portraitShimmer.Stop()
		h.completeReveal()

	case homeRevealTickMsg:
		if h.revealDone {
			return h, nil
		}
		h.revealLines += revealLinesPerTick
		full := h.buildFullContent()
		totalLines := strings.Count(full, "\n") + 1
		if h.revealLines >= totalLines {
			h.revealDone = true
			h.hasRevealed = true
			h.viewport.SetContentPreserveScroll(full)
			return h, nil
		}
		h.viewport.SetContentPreserveScroll(h.buildContent())
		return h, homeRevealTick()

	default:
		// Delegate shimmer ticks.
		var cmd tea.Cmd
		h.portraitShimmer, cmd = h.portraitShimmer.Update(msg)
		if cmd != nil {
			h.viewport.SetContentPreserveScroll(h.buildContent())
			return h, cmd
		}
	}

	return h, nil
}

// completeReveal finishes any running line-by-line reveal animation immediately.
func (h *HomeSection) completeReveal() {
	if h.revealDone {
		return
	}
	h.revealDone = true
	h.hasRevealed = true
	h.viewport.SetContentPreserveScroll(h.buildFullContent())
}

// View implements app.SectionModel.
func (h *HomeSection) View() string {
	return h.viewport.ViewWithScrollbar(h.theme)
}

// ScrollInfo implements app.ScrollReporter for the status bar scroll indicator.
func (h *HomeSection) ScrollInfo() app.ScrollInfo {
	return h.viewport.GetScrollInfo()
}

// KeyHints implements app.KeyHinter for contextual status bar hints.
func (h *HomeSection) KeyHints() string {
	return "j/k scroll " + app.BorderVertical + " pgup/dn page " + app.BorderVertical + " ^u/^d half " + app.BorderVertical + " 1-4 nav " + app.BorderVertical + " ? help"
}

// buildFullContent builds the complete section text regardless of reveal state.
func (h *HomeSection) buildFullContent() string {
	if h.content == nil {
		return ""
	}

	about := h.content.About
	contentWidth := h.viewport.ContentWidth()
	if contentWidth < 1 {
		contentWidth = 1
	}

	if contentWidth >= portraitMinWidth && portrait != "" {
		return h.renderNeofetch(about, contentWidth)
	}
	return h.renderStacked(about, contentWidth)
}

// buildContent returns the visible portion of the section content,
// accounting for the line-by-line reveal animation.
func (h *HomeSection) buildContent() string {
	full := h.buildFullContent()
	if h.revealDone {
		return full
	}
	lines := strings.Split(full, "\n")
	if h.revealLines >= len(lines) {
		return full
	}
	return strings.Join(lines[:h.revealLines], "\n")
}

// styledPortrait returns the portrait text with shimmer or muted styling.
func (h *HomeSection) styledPortrait() string {
	if h.portraitShimmer.Active() {
		firstLine := strings.SplitN(portrait, "\n", 2)[0]
		pw := lipgloss.Width(firstLine)
		return h.portraitShimmer.Render(portrait, pw)
	}
	return h.theme.Muted.Render(portrait)
}

// renderNeofetch renders the side-by-side neofetch-style layout.
func (h *HomeSection) renderNeofetch(about content.About, contentWidth int) string {
	styledP := h.styledPortrait()
	portraitWidth := lipgloss.Width(styledP)

	// Responsive gap.
	gap := 4
	remaining := contentWidth - portraitWidth - 30
	if remaining > 12 {
		gap = 6
	} else if remaining < 6 {
		gap = 2
	}

	rightColWidth := contentWidth - portraitWidth - gap
	if rightColWidth < 20 {
		rightColWidth = 20
	}

	var lines []string

	// Bio word-wrapped.
	if about.Bio != "" {
		bioWidth := rightColWidth
		if bioWidth < 10 {
			bioWidth = 10
		}
		wrapped := app.WrapText(about.Bio, bioWidth)
		for _, wl := range wrapped {
			lines = append(lines, h.theme.Body.Render(wl))
		}
	}

	// Blank line before info fields.
	lines = append(lines, "")

	infoBlock := h.renderInfo(about)
	if infoBlock != "" {
		lines = append(lines, infoBlock)
	}

	rightBlock := strings.Join(lines, "\n")
	gapStr := strings.Repeat(" ", gap)

	return lipgloss.JoinHorizontal(lipgloss.Center, styledP, gapStr, rightBlock)
}

// renderStacked renders the vertically-stacked layout for narrow terminals.
func (h *HomeSection) renderStacked(about content.About, contentWidth int) string {
	density := app.DensityForHeight(h.height)
	sep := app.SectionSeparator(density)

	var sections []string

	// Bio.
	if about.Bio != "" {
		wrapped := app.WrapText(about.Bio, contentWidth)
		sections = append(sections, h.theme.Body.Render(strings.Join(wrapped, "\n")))
	}

	// Info fields (status, email, CLI).
	infoBlock := h.renderInfo(about)
	if infoBlock != "" {
		sections = append(sections, infoBlock)
	}

	return strings.Join(sections, sep)
}

// renderInfo renders status, email, and CLI with accent-colored labels.
func (h *HomeSection) renderInfo(about content.About) string {
	var lines []string

	labelStyle := h.theme.Accent
	valueStyle := h.theme.Body

	if about.Status != "" {
		lines = append(lines, fmt.Sprintf(
			"%s %s",
			labelStyle.Render("Status"),
			valueStyle.Render(about.Status),
		))
	}
	if about.Email != "" {
		lines = append(lines, fmt.Sprintf(
			"%s %s",
			labelStyle.Render("Email"),
			valueStyle.Render(about.Email),
		))
	}
	if siteURL := h.content.Meta.SiteURL; siteURL != "" {
		display := strings.TrimPrefix(siteURL, "https://")
		lines = append(lines, fmt.Sprintf(
			"%s %s",
			labelStyle.Render("Web"),
			valueStyle.Render(display),
		))
	}

	return strings.Join(lines, "\n")
}
