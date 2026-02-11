package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/buntingszn/terminal-portfolio/tui/internal/analytics"
	"github.com/buntingszn/terminal-portfolio/tui/internal/content"
)

// ChromeHeight is the number of terminal lines consumed by the root model's
// chrome (navbar + blank line + statusbar). Sections receive a WindowSizeMsg
// with Height already reduced by this value.
const ChromeHeight = 3

// MinWidth and MinHeight define the minimum terminal dimensions required to
// render the UI. Below these thresholds, View() displays a resize message.
const (
	MinWidth  = 20
	MinHeight = 8
)

// SectionModel defines the interface that each navigable section must implement.
// It mirrors tea.Model so sections can be used as standalone Bubbletea models,
// but returns SectionModel from Update to preserve the concrete type.
type SectionModel interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (SectionModel, tea.Cmd)
	View() string
}

// Model is the root Bubbletea model that manages section routing,
// global key bindings, and theme state.
type Model struct {
	activeSection Section
	sections      [SectionCount]SectionModel
	theme         Theme
	content       *content.Content
	statusBar     StatusBar
	navBar        NavBar
	intro         IntroModel
	showIntro     bool
	transition    TransitionManager
	palette       PaletteModel
	showPalette   bool
	width         int
	height        int
	showHelp      bool

	// Idle timeout fields. When idleTimeout > 0, the model tracks user
	// activity and shows a warning before disconnecting idle sessions.
	// A value of 0 disables idle tracking entirely.
	idleTimeout     time.Duration
	lastActivity    time.Time
	showIdleWarning bool
	idleRemaining   time.Duration

	// Analytics fields. When analyticsLog is non-nil, the model emits
	// session_start, section_view, and session_end events to the JSONL log.
	analyticsLog  *analytics.Logger
	sessionID     string
	sessionIP     string
	sessionStart  time.Time
	sectionStart  time.Time
}

// New creates a new root Model with the given content data.
// It initializes all sections with placeholder implementations and starts
// with the dark theme, home section active, and the intro boot sequence.
func New(c *content.Content, secs ...SectionModel) Model {
	theme := DarkTheme()
	var sections [SectionCount]SectionModel
	for i := range SectionCount {
		if i < len(secs) {
			sections[i] = secs[i]
		} else {
			sections[i] = newPlaceholderSection(SectionName(Section(i)), theme)
		}
	}
	return Model{
		activeSection: SectionHome,
		sections:      sections,
		theme:      theme,
		content:    c,
		statusBar:  NewStatusBar(theme, 0),
		navBar:     NewNavBar(theme, 0),
		intro:      NewIntroModel(theme),
		showIntro:  true,
		transition: NewTransitionManager(),
		palette:    NewPaletteModel(theme),
	}
}

// SetIdleTimeout configures the idle timeout duration for the model.
// A value of 0 disables idle tracking. This should be called before Init().
func (m Model) SetIdleTimeout(d time.Duration) Model {
	m.idleTimeout = d
	if d > 0 {
		m.lastActivity = time.Now()
	}
	return m
}

// SetAnalytics configures analytics logging for the model.
// A nil logger disables analytics. This should be called before Init().
func (m Model) SetAnalytics(l *analytics.Logger, sid, ip string) Model {
	m.analyticsLog = l
	m.sessionID = sid
	m.sessionIP = ip
	m.sessionStart = time.Now()
	m.sectionStart = m.sessionStart
	return m
}

// logSectionView emits a section_view event for the current section and
// returns the current time for use as the next sectionStart.
func (m *Model) logSectionView() time.Time {
	now := time.Now()
	if m.analyticsLog == nil {
		return now
	}
	m.analyticsLog.Log(analytics.Event{
		Timestamp:  now,
		SessionID:  m.sessionID,
		Type:       analytics.EventSectionView,
		Section:    SectionName(m.activeSection),
		DurationMs: now.Sub(m.sectionStart).Milliseconds(),
	})
	return now
}

// logSessionEnd emits the final section_view and session_end events.
func (m *Model) logSessionEnd() {
	if m.analyticsLog == nil {
		return
	}
	m.logSectionView()
	m.analyticsLog.Log(analytics.Event{
		Timestamp:  time.Now(),
		SessionID:  m.sessionID,
		Type:       analytics.EventSessionEnd,
		DurationMs: time.Since(m.sessionStart).Milliseconds(),
	})
}

// Init implements tea.Model. It starts the intro boot sequence and, if
// idle timeout is configured, begins the periodic idle check.
func (m Model) Init() tea.Cmd {
	var cmds []tea.Cmd
	cmds = append(cmds, tea.SetWindowTitle(m.content.Meta.Name+" — "+m.content.Meta.Title))
	if m.showIntro {
		cmds = append(cmds, m.intro.Init())
	} else {
		cmds = append(cmds, m.sections[m.activeSection].Init())
	}
	if m.idleTimeout > 0 {
		cmds = append(cmds, idleCheckTick())
	}
	return tea.Batch(cmds...)
}

// Update implements tea.Model. It handles global keys before delegating to sections.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case idleCheckMsg:
		return m.handleIdleCheck()
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)
	case IntroDoneMsg:
		return m.handleIntroDone()
	case TransitionDoneMsg:
		return m.handleTransitionDone()
	case AnimationTickMsg:
		if m.transition.Active() {
			return m, m.transition.Update(msg)
		}
		return m, nil
	case PaletteResultMsg:
		return m.handlePaletteResult(msg)
	case NavigateMsg:
		return m.navigateTo(msg.Section)
	case tea.MouseMsg:
		return m.handleMouse(msg)
	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// During intro, delegate non-key messages to intro.
	if m.showIntro {
		var cmd tea.Cmd
		m.intro, cmd = m.intro.Update(msg)
		return m, cmd
	}

	// Delegate to active section.
	var cmd tea.Cmd
	m.sections[m.activeSection], cmd = m.sections[m.activeSection].Update(msg)
	return m, cmd
}

// handleWindowSize propagates resize events to all chrome and sections.
func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	m.statusBar.SetWidth(msg.Width)
	m.navBar.SetWidth(msg.Width)
	m.palette.SetWidth(msg.Width)
	m.intro.SetSize(msg.Width, msg.Height)

	sectionHeight := msg.Height - ChromeHeight
	if sectionHeight < 1 {
		sectionHeight = 1
	}
	sectionMsg := tea.WindowSizeMsg{Width: msg.Width, Height: sectionHeight}
	var cmds []tea.Cmd
	for i := range m.sections {
		var cmd tea.Cmd
		m.sections[i], cmd = m.sections[i].Update(sectionMsg)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return m, tea.Batch(cmds...)
}

// handleIntroDone transitions from the boot sequence to the active section.
func (m Model) handleIntroDone() (tea.Model, tea.Cmd) {
	m.showIntro = false
	m.sectionStart = time.Now()
	m.navBar.SetActive(m.activeSection)
	initCmd := m.sections[m.activeSection].Init()
	var focusCmd tea.Cmd
	m.sections[m.activeSection], focusCmd = m.sections[m.activeSection].Update(FocusMsg{})
	return m, tea.Batch(initCmd, focusCmd)
}

// handleTransitionDone sends FocusMsg to the now-active section.
func (m Model) handleTransitionDone() (tea.Model, tea.Cmd) {
	var focusCmd tea.Cmd
	m.sections[m.activeSection], focusCmd = m.sections[m.activeSection].Update(FocusMsg{})
	return m, focusCmd
}

// handlePaletteResult processes the selected command palette action.
func (m Model) handlePaletteResult(msg PaletteResultMsg) (tea.Model, tea.Cmd) {
	m.showPalette = false
	m.palette.Close()
	switch msg.Action {
	case PaletteNavigate:
		return m.navigateTo(msg.Section)
	case PaletteQuit:
		m.logSessionEnd()
		return m, tea.Quit
	case PaletteHelp:
		m.showHelp = true
		return m, nil
	default:
		return m, nil
	}
}

// handleMouse delegates mouse events to the active section for scroll handling.
func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	m.resetIdleTimer()
	if m.showIntro || m.transition.Active() || m.showPalette || m.showHelp {
		return m, nil
	}
	var cmd tea.Cmd
	m.sections[m.activeSection], cmd = m.sections[m.activeSection].Update(msg)
	return m, cmd
}

// handleKey processes global key bindings and delegates to overlays or sections.
func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.resetIdleTimer()

	if m.showIntro {
		var cmd tea.Cmd
		m.intro, cmd = m.intro.Update(msg)
		return m, cmd
	}
	if m.transition.Active() {
		return m, nil
	}
	if m.showPalette {
		var cmd tea.Cmd
		m.palette, cmd = m.palette.Update(msg)
		return m, cmd
	}
	if m.showHelp {
		m.showHelp = false
		return m, nil
	}

	switch msg.String() {
	case "q", "ctrl+c":
		m.logSessionEnd()
		return m, tea.Quit
	case "?":
		m.showHelp = true
		return m, nil
	case ":":
		m.showPalette = true
		m.palette.Open()
		return m, nil
	case "tab", "right":
		next := Section((int(m.activeSection) + 1) % SectionCount)
		return m.navigateTo(next)
	case "shift+tab", "left":
		prev := Section((int(m.activeSection) - 1 + SectionCount) % SectionCount)
		return m.navigateTo(prev)
	case "1":
		return m.navigateTo(SectionHome)
	case "2":
		return m.navigateTo(SectionWork)
	case "3":
		return m.navigateTo(SectionCV)
	case "4":
		return m.navigateTo(SectionLinks)
	}

	// Delegate unmatched keys to the active section (j/k/g/G/pgup/etc).
	var cmd tea.Cmd
	m.sections[m.activeSection], cmd = m.sections[m.activeSection].Update(msg)
	return m, cmd
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width < MinWidth || m.height < MinHeight {
		title := m.theme.Accent.Render("Terminal too small")
		body := m.theme.Body.Render(fmt.Sprintf("Please resize to at least %d\u00d7%d", MinWidth, MinHeight))
		msg := title + "\n" + body
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
	}

	if m.showIntro {
		return m.intro.View()
	}

	if m.showHelp {
		return m.helpView()
	}

	var b strings.Builder
	b.WriteString(m.navBar.View())
	b.WriteString("\n\n")

	if m.transition.Active() {
		fromView := m.sections[m.transition.from].View()
		toView := m.sections[m.transition.to].View()
		b.WriteString(m.transition.View(fromView, toView, m.width))
	} else {
		b.WriteString(m.sections[m.activeSection].View())
	}

	b.WriteString("\n")
	b.WriteString(m.statusView())

	if m.showPalette {
		b.WriteString("\n")
		b.WriteString(m.palette.View())
	}

	if m.showIdleWarning {
		b.WriteString("\n")
		b.WriteString(m.idleWarningView())
	}

	return b.String()
}

// navigateTo switches to the target section with a transition animation.
// FocusMsg is deferred until the transition completes (TransitionDoneMsg).
// Navigating to the already-active section is a no-op, and navigation
// during an active transition is ignored to prevent duplicate processing.
func (m Model) navigateTo(target Section) (tea.Model, tea.Cmd) {
	if target == m.activeSection {
		return m, nil
	}
	if m.transition.Active() {
		return m, nil
	}

	// Log the departing section view before switching.
	m.sectionStart = m.logSectionView()

	var cmds []tea.Cmd

	// Blur the current section.
	var blurCmd tea.Cmd
	m.sections[m.activeSection], blurCmd = m.sections[m.activeSection].Update(BlurMsg{})
	if blurCmd != nil {
		cmds = append(cmds, blurCmd)
	}

	// Start transition animation (step count varies by section distance).
	from := m.activeSection
	transCmd := m.transition.Start(from, target)
	if transCmd != nil {
		cmds = append(cmds, transCmd)
	}

	// Switch active section and update navbar.
	// FocusMsg is sent later when TransitionDoneMsg fires.
	m.activeSection = target
	m.navBar.SetActive(target)

	return m, tea.Batch(cmds...)
}

// statusView renders the bottom status bar.
func (m Model) statusView() string {
	var hints string
	if kh, ok := m.sections[m.activeSection].(KeyHinter); ok {
		hints = kh.KeyHints()
	}
	var scroll ScrollInfo
	if sr, ok := m.sections[m.activeSection].(ScrollReporter); ok {
		scroll = sr.ScrollInfo()
	} else {
		scroll = ScrollInfo{Fits: true}
	}
	return m.statusBar.Render(m.activeSection, hints, scroll)
}

// helpShortcut defines a single key-description pair for the help overlay.
type helpShortcut struct {
	key  string
	desc string
}

// helpShortcuts returns the full list of keyboard shortcuts displayed in the
// help overlay. The key column width is chosen so that the longest key label
// fits comfortably with trailing padding.
func helpShortcuts() []helpShortcut {
	return []helpShortcut{
		{"\u2190 / \u2192", "Previous / next section"},
		{"1-4", "Jump to section"},
		{"j / k", "Scroll down / up"},
		{"g / G", "Jump to top / bottom"},
		{"PgUp", "Page up"},
		{"PgDn", "Page down"},
		{"^u / ^d", "Half-page up / down"},
		{":", "Command palette"},
		{"q", "Quit"},
		{"?", "Toggle help"},
	}
}

// helpView renders the help overlay.
func (m Model) helpView() string {
	shortcuts := helpShortcuts()

	// Build two-column aligned help text. Key column is right-padded to a
	// fixed width so descriptions line up neatly.
	const keyColWidth = 10
	var lines []string
	for _, sc := range shortcuts {
		keyStr := fmt.Sprintf("%-*s", keyColWidth, sc.key)
		line := m.theme.Accent.Render(keyStr) + m.theme.Body.Render(sc.desc)
		lines = append(lines, line)
	}
	lines = append(lines, "")
	lines = append(lines, m.theme.Muted.Render("Press any key to dismiss"))

	helpLines := strings.Join(lines, "\n")

	// Determine card width: cap at 50, but don't exceed terminal width.
	cardWidth := 50
	if m.width > 0 && m.width < cardWidth {
		cardWidth = m.width
	}

	// If terminal is too small for a card, render plain text without centering.
	if cardWidth < 10 || m.width < 10 || m.height < 10 {
		title := m.theme.Title.Render("Keyboard Shortcuts")
		return title + "\n\n" + helpLines
	}

	card := RenderCard(m.theme, "Keyboard Shortcuts", helpLines, cardWidth)
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		card,
		lipgloss.WithWhitespaceChars("·"),
		lipgloss.WithWhitespaceForeground(m.theme.Colors.Border),
	)
}

// --- Placeholder section (replaced by real sections in later stories) ---

// placeholderSection is a minimal SectionModel used until real sections are built.
type placeholderSection struct {
	name  string
	theme Theme
}

func newPlaceholderSection(name string, theme Theme) *placeholderSection {
	return &placeholderSection{name: name, theme: theme}
}

func (p *placeholderSection) Init() tea.Cmd {
	return nil
}

func (p *placeholderSection) Update(_ tea.Msg) (SectionModel, tea.Cmd) {
	return p, nil
}

func (p *placeholderSection) View() string {
	return p.theme.Body.Render(fmt.Sprintf("[ %s section ]", p.name))
}
