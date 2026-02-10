package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/buntingszn/terminal-portfolio/tui/internal/content"
)

// testContent returns minimal content for testing.
func testContent() *content.Content {
	return &content.Content{
		Meta: content.Meta{
			Name:    "Test User",
			Title:   "Developer",
			Version: "1.0.0",
		},
		About: content.About{
			Bio:   "A bio",
			Email: "test@example.com",
		},
	}
}

// skipIntro creates a model and skips the intro sequence so tests can
// exercise normal navigation without dealing with boot animation.
func skipIntro(t *testing.T) Model {
	t.Helper()
	m := New(testContent())
	// Set a reasonable default terminal size so View() doesn't hit the
	// minimum-size guard (width < 20 || height < 8).
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(Model)
	// Send a key to skip the intro.
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = result.(Model)
	// Process the IntroDoneMsg.
	result, _ = m.Update(IntroDoneMsg{})
	m = result.(Model)
	if m.showIntro {
		t.Fatal("expected intro to be skipped")
	}
	return m
}

// drainTransition runs all animation ticks until the transition completes.
func drainTransition(t *testing.T, m Model) Model {
	t.Helper()
	for m.transition.Active() {
		result, _ := m.Update(AnimationTickMsg{ID: transitionID})
		m = result.(Model)
	}
	// Process the TransitionDoneMsg.
	result, _ := m.Update(TransitionDoneMsg{})
	m = result.(Model)
	return m
}

func TestNewModel(t *testing.T) {
	m := New(testContent())

	if m.activeSection != SectionHome {
		t.Errorf("activeSection = %d, want %d (home)", m.activeSection, SectionHome)
	}
	if m.content == nil {
		t.Error("content should not be nil")
	}
	if !m.showIntro {
		t.Error("expected showIntro to be true by default")
	}
}

func TestInitReturnsCmd(t *testing.T) {
	m := New(testContent())
	cmd := m.Init()
	// With showIntro=true, Init returns the intro tick command.
	if cmd == nil {
		t.Error("expected non-nil cmd from Init for intro")
	}
}

func TestIntroSkipByKey(t *testing.T) {
	m := New(testContent())
	// Any key during intro should skip it.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	// The skip produces an IntroDoneMsg command.
	if cmd == nil {
		t.Fatal("expected cmd after skipping intro")
	}
	msg := cmd()
	if _, ok := msg.(IntroDoneMsg); !ok {
		t.Errorf("expected IntroDoneMsg, got %T", msg)
	}
}

func TestIntroDoneSwitchesToNormal(t *testing.T) {
	m := New(testContent())
	result, _ := m.Update(IntroDoneMsg{})
	m = result.(Model)
	if m.showIntro {
		t.Error("expected showIntro to be false after IntroDoneMsg")
	}
}

func TestViewContainsNav(t *testing.T) {
	m := skipIntro(t)
	// Set a width large enough for full labels.
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(Model)
	view := m.View()
	if !strings.Contains(view, "home") {
		t.Error("view should contain 'home'")
	}
	if !strings.Contains(view, "work") {
		t.Error("view should contain 'work'")
	}
	if !strings.Contains(view, "cv") {
		t.Error("view should contain 'cv'")
	}
	if !strings.Contains(view, "links") {
		t.Error("view should contain 'links'")
	}
}

func TestNavigateByKeyMsg(t *testing.T) {
	m := skipIntro(t)

	tests := []struct {
		key  string
		want Section
	}{
		{"2", SectionWork},
		{"3", SectionCV},
		{"4", SectionLinks},
		{"1", SectionHome},
	}

	for _, tt := range tests {
		result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(tt.key)})
		m = result.(Model)
		if m.activeSection != tt.want {
			t.Errorf("after pressing %q: activeSection = %d, want %d", tt.key, m.activeSection, tt.want)
		}
		// Drain the transition so the next navigation key is not buffered.
		m = drainTransition(t, m)
	}
}

func TestNavigateToSameSection(t *testing.T) {
	m := skipIntro(t)
	// Already on home, pressing 1 should be a no-op.
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("1")})
	m = result.(Model)
	if m.activeSection != SectionHome {
		t.Errorf("activeSection = %d, want %d", m.activeSection, SectionHome)
	}
	if cmd != nil {
		t.Error("expected nil cmd for navigating to same section")
	}
}

func TestNavigateMsg(t *testing.T) {
	m := skipIntro(t)
	result, _ := m.Update(NavigateMsg{Section: SectionCV})
	m = result.(Model)
	if m.activeSection != SectionCV {
		t.Errorf("activeSection = %d, want %d (cv)", m.activeSection, SectionCV)
	}
	m = drainTransition(t, m)
}

func TestHelpToggle(t *testing.T) {
	m := skipIntro(t)
	// Set a valid terminal size so View() does not show the "too small" guard.
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(Model)

	// Press ? to show help.
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = result.(Model)
	if !m.showHelp {
		t.Error("expected showHelp to be true after pressing ?")
	}

	view := m.View()
	if !strings.Contains(view, "Keyboard Shortcuts") {
		t.Error("help view should contain 'Keyboard Shortcuts'")
	}

	// Any key dismisses help.
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")})
	m = result.(Model)
	if m.showHelp {
		t.Error("expected showHelp to be false after pressing any key")
	}
}

func TestQuitKey(t *testing.T) {
	m := skipIntro(t)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("q")})
	if cmd == nil {
		t.Fatal("expected quit command")
	}
	// Execute the cmd to verify it produces a quit message.
	msg := cmd()
	if _, ok := msg.(tea.QuitMsg); !ok {
		t.Errorf("expected tea.QuitMsg, got %T", msg)
	}
}

func TestWindowSizeMsg(t *testing.T) {
	m := skipIntro(t)
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = result.(Model)
	if m.width != 120 {
		t.Errorf("width = %d, want 120", m.width)
	}
	if m.height != 40 {
		t.Errorf("height = %d, want 40", m.height)
	}
}

func TestWindowSizeMsgAdjustsForChrome(t *testing.T) {
	// Use a spy section to capture the WindowSizeMsg it receives.
	spy := &spySection{}
	m := New(testContent(), spy)
	// Skip intro.
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = result.(Model)
	result, _ = m.Update(IntroDoneMsg{})
	m = result.(Model)

	result, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_ = result.(Model)

	wantHeight := 24 - ChromeHeight
	if spy.lastHeight != wantHeight {
		t.Errorf("section received height %d, want %d (terminal %d - chrome %d)",
			spy.lastHeight, wantHeight, 24, ChromeHeight)
	}
	if spy.lastWidth != 80 {
		t.Errorf("section received width %d, want 80", spy.lastWidth)
	}
}

func TestWindowSizeMsgMinimumHeight(t *testing.T) {
	spy := &spySection{}
	m := New(testContent(), spy)
	// Skip intro.
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = result.(Model)
	result, _ = m.Update(IntroDoneMsg{})
	m = result.(Model)

	// Terminal height = 2, which is less than ChromeHeight. Section should get minimum 1.
	result, _ = m.Update(tea.WindowSizeMsg{Width: 40, Height: 2})
	_ = result.(Model)

	if spy.lastHeight != 1 {
		t.Errorf("section received height %d, want 1 (minimum)", spy.lastHeight)
	}
}

// spySection captures WindowSizeMsg dimensions for testing.
type spySection struct {
	lastWidth  int
	lastHeight int
}

func (s *spySection) Init() tea.Cmd { return nil }

func (s *spySection) Update(msg tea.Msg) (SectionModel, tea.Cmd) {
	if wsm, ok := msg.(tea.WindowSizeMsg); ok {
		s.lastWidth = wsm.Width
		s.lastHeight = wsm.Height
	}
	return s, nil
}

func (s *spySection) View() string { return "" }

func TestSectionName(t *testing.T) {
	tests := []struct {
		section Section
		want    string
	}{
		{SectionHome, "home"},
		{SectionWork, "work"},
		{SectionCV, "cv"},
		{SectionLinks, "links"},
		{Section(99), "unknown"},
	}
	for _, tt := range tests {
		got := SectionName(tt.section)
		if got != tt.want {
			t.Errorf("SectionName(%d) = %q, want %q", tt.section, got, tt.want)
		}
	}
}

func TestStatusViewContainsHints(t *testing.T) {
	m := skipIntro(t)
	m.width = 80
	m.statusBar.SetWidth(80)
	view := m.statusView()
	if !strings.Contains(view, "? help") {
		t.Error("status bar should contain '? help'")
	}
	if !strings.Contains(view, "nav") {
		t.Error("status bar should contain 'nav'")
	}
}

func TestPlaceholderSectionView(t *testing.T) {
	theme := DarkTheme()
	p := newPlaceholderSection("test", theme)
	view := p.View()
	if !strings.Contains(view, "test section") {
		t.Errorf("placeholder view = %q, should contain 'test section'", view)
	}
}

func TestCommandPaletteOpen(t *testing.T) {
	m := skipIntro(t)

	// Press : to open palette.
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
	m = result.(Model)
	if !m.showPalette {
		t.Error("expected showPalette to be true after pressing :")
	}
}

func TestCommandPaletteEscape(t *testing.T) {
	m := skipIntro(t)

	// Open palette.
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
	m = result.(Model)

	// Press Escape to close.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEscape})

	if cmd == nil {
		t.Fatal("expected PaletteResultMsg cmd")
	}
	msg := cmd()
	pr, ok := msg.(PaletteResultMsg)
	if !ok {
		t.Fatalf("expected PaletteResultMsg, got %T", msg)
	}
	if pr.Action != PaletteNone {
		t.Errorf("expected PaletteNone, got %d", pr.Action)
	}
}

func TestCommandPaletteNavigate(t *testing.T) {
	m := skipIntro(t)

	// Open palette.
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(":")})
	m = result.(Model)

	// Type "work" and press Enter.
	for _, c := range "work" {
		result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{c}})
		m = result.(Model)
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})

	if cmd == nil {
		t.Fatal("expected cmd after palette enter")
	}
	msg := cmd()
	pr, ok := msg.(PaletteResultMsg)
	if !ok {
		t.Fatalf("expected PaletteResultMsg, got %T", msg)
	}
	if pr.Action != PaletteNavigate {
		t.Errorf("expected PaletteNavigate, got %d", pr.Action)
	}
	if pr.Section != SectionWork {
		t.Errorf("expected SectionWork, got %d", pr.Section)
	}
}

func TestIntroViewShowsMessages(t *testing.T) {
	m := New(testContent())
	// Set terminal size so View() doesn't hit the minimum-size guard.
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(Model)
	// Process a few ticks.
	result, _ = m.Update(introTickMsg{})
	m = result.(Model)
	result, _ = m.Update(introTickMsg{})
	m = result.(Model)

	view := m.View()
	if !strings.Contains(view, "POST") {
		t.Error("intro view should contain 'POST' after ticks")
	}
}

func TestNavBarView(t *testing.T) {
	theme := DarkTheme()
	nb := NewNavBar(theme, 60)
	nb.SetActive(SectionWork)
	view := nb.View()

	if !strings.Contains(view, "home") {
		t.Error("navbar should contain 'home'")
	}
	if !strings.Contains(view, "work") {
		t.Error("navbar should contain 'work'")
	}
}

func TestNavBarViewFullLabelsWide(t *testing.T) {
	theme := DarkTheme()
	nb := NewNavBar(theme, 80)
	nb.SetActive(SectionHome)
	view := nb.View()

	for _, name := range []string{"home", "work", "cv", "links"} {
		if !strings.Contains(view, name) {
			t.Errorf("navbar at width 80 should contain %q", name)
		}
	}
}

func TestNavBarViewShortLabels(t *testing.T) {
	theme := DarkTheme()
	nb := NewNavBar(theme, 30)
	nb.SetActive(SectionHome)
	view := nb.View()

	if !strings.Contains(view, "hm") {
		t.Error("navbar at width 30 should contain short label 'hm'")
	}
	if !strings.Contains(view, "wk") {
		t.Error("navbar at width 30 should contain short label 'wk'")
	}
	if !strings.Contains(view, "lk") {
		t.Error("navbar at width 30 should contain short label 'lk'")
	}
	if strings.Contains(view, "home") {
		t.Error("navbar at width 30 should NOT contain full label 'home'")
	}
	if strings.Contains(view, "links") {
		t.Error("navbar at width 30 should NOT contain full label 'links'")
	}
}

func TestNavBarViewNumberOnly(t *testing.T) {
	theme := DarkTheme()
	nb := NewNavBar(theme, 20)
	nb.SetActive(SectionHome)
	view := nb.View()

	if !strings.Contains(view, "1") {
		t.Error("navbar at width 20 should contain '1'")
	}
	if !strings.Contains(view, "2") {
		t.Error("navbar at width 20 should contain '2'")
	}
	if strings.Contains(view, "home") {
		t.Error("navbar at width 20 should NOT contain 'home'")
	}
	if strings.Contains(view, "hm") {
		t.Error("navbar at width 20 should NOT contain 'hm'")
	}
}

func TestNavBarViewNarrowNoPanic(t *testing.T) {
	theme := DarkTheme()
	nb := NewNavBar(theme, 10)
	nb.SetActive(SectionCV)
	// Should not panic at very narrow widths.
	view := nb.View()
	if len(view) == 0 {
		t.Error("navbar at width 10 should produce non-empty output")
	}
}

func TestNavLabelForWidth(t *testing.T) {
	tests := []struct {
		width int
		want  navLabelFormat
	}{
		{80, navLabelFull},
		{40, navLabelFull},
		{39, navLabelShort},
		{25, navLabelShort},
		{24, navLabelNumOnly},
		{10, navLabelNumOnly},
		{0, navLabelNumOnly},
	}
	for _, tt := range tests {
		got := navLabelForWidth(tt.width)
		if got != tt.want {
			t.Errorf("navLabelForWidth(%d) = %d, want %d", tt.width, got, tt.want)
		}
	}
}

func TestTransitionManagerStartAndComplete(t *testing.T) {
	tm := NewTransitionManager()
	cmd := tm.Start(SectionHome, SectionWork)
	if cmd == nil {
		t.Fatal("expected cmd from Start")
	}
	if !tm.Active() {
		t.Error("expected transition to be active")
	}

	// Run through all steps.
	for tm.Active() {
		cmd = tm.Update(AnimationTickMsg{ID: transitionID})
	}
	if tm.Active() {
		t.Error("expected transition to be inactive after all steps")
	}

	// Last cmd should produce TransitionDoneMsg.
	if cmd == nil {
		t.Fatal("expected final cmd")
	}
	msg := cmd()
	if _, ok := msg.(TransitionDoneMsg); !ok {
		t.Errorf("expected TransitionDoneMsg, got %T", msg)
	}
}

func TestKeysBufferedDuringTransition(t *testing.T) {
	m := skipIntro(t)

	// Navigate to work (starts transition).
	result, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	m = result.(Model)

	if !m.transition.Active() {
		t.Fatal("expected transition to be active")
	}

	// Press 3 during transition - should be ignored.
	result, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("3")})
	m = result.(Model)

	if m.activeSection != SectionWork {
		t.Errorf("activeSection should still be work during transition, got %d", m.activeSection)
	}
	if cmd != nil {
		t.Error("expected nil cmd for buffered key during transition")
	}
}

func TestMinTerminalSizeGuard(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		wantSmall bool
	}{
		{"narrow width", 15, 20, true},
		{"short height", 80, 5, true},
		{"one below min width", 19, 8, true},
		{"exact minimum size", 20, 8, false},
		{"large terminal", 80, 24, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := skipIntro(t)
			result, _ := m.Update(tea.WindowSizeMsg{Width: tt.width, Height: tt.height})
			m = result.(Model)
			view := m.View()
			hasSmall := strings.Contains(view, "too small")
			if hasSmall != tt.wantSmall {
				t.Errorf("View() contains 'too small' = %v, want %v (width=%d, height=%d)",
					hasSmall, tt.wantSmall, tt.width, tt.height)
			}
		})
	}
}

func TestHelpViewContainsCommandPalette(t *testing.T) {
	m := skipIntro(t)
	// Set a valid terminal size so View() does not show the "too small" guard.
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(Model)

	// Show help.
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = result.(Model)

	view := m.View()
	if !strings.Contains(view, "Command palette") {
		t.Error("help view should contain 'Command palette'")
	}
}

func TestHelpOverlayContainsBorder(t *testing.T) {
	m := skipIntro(t)
	// Set a reasonable terminal size so the card renders with borders.
	result, _ := m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m = result.(Model)

	// Show help.
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("?")})
	m = result.(Model)

	view := m.View()
	if !strings.Contains(view, "┌") && !strings.Contains(view, "─") {
		t.Error("help overlay should contain border characters (┌ or ─)")
	}
}

func TestPaletteViewNarrowWidth(t *testing.T) {
	p := NewPaletteModel(DarkTheme())
	p.Open()
	p.SetWidth(15)

	// Should not panic and should render something.
	view := p.View()
	if len(view) == 0 {
		t.Error("palette View() at width 15 should produce non-empty output")
	}

	// At width < 20, should render simple single-line (no box border).
	if strings.Contains(view, "┌") {
		t.Error("palette View() at width 15 should not contain box border")
	}
}

func TestPaletteViewWideContainsHints(t *testing.T) {
	p := NewPaletteModel(DarkTheme())
	p.Open()
	p.SetWidth(80)

	view := p.View()
	if !strings.Contains(view, "home work") {
		t.Error("palette View() at width 80 should contain 'home work' hints")
	}
}

func TestStatusBarCenteredHints(t *testing.T) {
	theme := DarkTheme()

	t.Run("width80_centered_hints", func(t *testing.T) {
		sb := NewStatusBar(theme, 80)
		out := sb.Render(SectionHome, "", ScrollInfo{Fits: true})
		if !strings.Contains(out, "? help") {
			t.Error("width 80: expected '? help' in centered hints")
		}
		if !strings.Contains(out, "nav") {
			t.Error("width 80: expected 'nav' in centered hints")
		}
	})

	t.Run("width25_still_shows_hints", func(t *testing.T) {
		sb := NewStatusBar(theme, 25)
		out := sb.Render(SectionHome, "", ScrollInfo{Fits: true})
		if out == "" {
			t.Error("width 25: expected non-empty output")
		}
	})

	t.Run("width5_no_panic", func(t *testing.T) {
		sb := NewStatusBar(theme, 5)
		out := sb.Render(SectionHome, "", ScrollInfo{Fits: true})
		if out == "" {
			t.Error("width 5: expected non-empty output")
		}
	})
}

func TestStatusBarRuneSafeTruncation(t *testing.T) {
	theme := DarkTheme()
	// Use a very narrow width where hints must be truncated; verify no broken UTF-8.
	// The static hints contain multi-byte arrow and middle-dot characters.
	sb := NewStatusBar(theme, 10)
	out := sb.Render(SectionHome, "", ScrollInfo{Fits: true})
	// Verify the output contains no replacement character (broken UTF-8).
	if strings.Contains(out, "\ufffd") {
		t.Error("output contains replacement character, indicating broken UTF-8")
	}
}

func TestStatusBarStaticContent(t *testing.T) {
	theme := DarkTheme()

	// The status bar now shows only static centered hints regardless of scroll state.
	t.Run("always_shows_static_hints", func(t *testing.T) {
		sb := NewStatusBar(theme, 80)
		out := sb.Render(SectionHome, "", ScrollInfo{Fits: true})
		if !strings.Contains(out, "? help") {
			t.Error("expected '? help' in status bar")
		}
	})

	t.Run("scroll_state_ignored", func(t *testing.T) {
		sb := NewStatusBar(theme, 80)
		scroll := ScrollInfo{AtTop: true, AtBottom: false, Percent: "  0%"}
		out := sb.Render(SectionHome, "", scroll)
		// Should NOT contain scroll indicators since status bar is now static.
		if strings.Contains(out, "TOP") {
			t.Error("should not contain TOP in simplified status bar")
		}
	})
}

func TestViewportGetScrollInfo(t *testing.T) {
	t.Run("content_fits", func(t *testing.T) {
		vp := NewViewport(80, 20)
		vp.SetContent("short\ncontent")
		info := vp.GetScrollInfo()
		if !info.Fits {
			t.Error("expected Fits=true when content fits in viewport")
		}
		if !info.AtTop {
			t.Error("expected AtTop=true when content fits")
		}
		if !info.AtBottom {
			t.Error("expected AtBottom=true when content fits")
		}
	})

	t.Run("scrollable_at_top", func(t *testing.T) {
		vp := NewViewport(80, 5)
		lines := make([]string, 20)
		for i := range lines {
			lines[i] = "line"
		}
		vp.SetContent(strings.Join(lines, "\n"))
		info := vp.GetScrollInfo()
		if info.Fits {
			t.Error("expected Fits=false when content overflows")
		}
		if !info.AtTop {
			t.Error("expected AtTop=true at initial position")
		}
		if info.AtBottom {
			t.Error("expected AtBottom=false at top of long content")
		}
	})

	t.Run("scrollable_at_bottom", func(t *testing.T) {
		vp := NewViewport(80, 5)
		lines := make([]string, 20)
		for i := range lines {
			lines[i] = "line"
		}
		vp.SetContent(strings.Join(lines, "\n"))
		vp.ScrollToBottom()
		info := vp.GetScrollInfo()
		if info.Fits {
			t.Error("expected Fits=false when content overflows")
		}
		if info.AtTop {
			t.Error("expected AtTop=false at bottom")
		}
		if !info.AtBottom {
			t.Error("expected AtBottom=true after ScrollToBottom")
		}
	})

	t.Run("scrollable_in_middle", func(t *testing.T) {
		vp := NewViewport(80, 5)
		lines := make([]string, 20)
		for i := range lines {
			lines[i] = "line"
		}
		vp.SetContent(strings.Join(lines, "\n"))
		vp.ScrollDown(5)
		info := vp.GetScrollInfo()
		if info.Fits {
			t.Error("expected Fits=false")
		}
		if info.AtTop {
			t.Error("expected AtTop=false in middle")
		}
		if info.AtBottom {
			t.Error("expected AtBottom=false in middle")
		}
		if info.Percent == "" {
			t.Error("expected non-empty Percent in middle")
		}
	})
}

func TestTransitionStepsVaryByDistance(t *testing.T) {
	tests := []struct {
		from, to  Section
		wantSteps int
	}{
		{SectionHome, SectionWork, baseTransitionSteps},                                   // distance 1
		{SectionHome, SectionCV, baseTransitionSteps + extraStepsPerDistance},              // distance 2
		{SectionHome, SectionLinks, baseTransitionSteps + 2*extraStepsPerDistance},         // distance 3
		{SectionLinks, SectionCV, baseTransitionSteps},                                    // distance 1 backward
		{SectionLinks, SectionHome, baseTransitionSteps + 2*extraStepsPerDistance},         // distance 3 backward
	}
	for _, tt := range tests {
		tm := NewTransitionManager()
		tm.Start(tt.from, tt.to)
		if tm.steps != tt.wantSteps {
			t.Errorf("Start(%d→%d): steps = %d, want %d", tt.from, tt.to, tm.steps, tt.wantSteps)
		}
	}
}

func TestNavigateDuringTransitionIsNoop(t *testing.T) {
	m := skipIntro(t)

	// Navigate to work (starts transition).
	result, _ := m.Update(NavigateMsg{Section: SectionWork})
	m = result.(Model)

	if !m.transition.Active() {
		t.Fatal("expected transition to be active")
	}

	// Try to navigate via NavigateMsg during active transition.
	result, cmd := m.Update(NavigateMsg{Section: SectionLinks})
	m = result.(Model)

	if m.activeSection != SectionWork {
		t.Errorf("activeSection should still be work, got %d", m.activeSection)
	}
	if cmd != nil {
		t.Error("expected nil cmd for NavigateMsg during active transition")
	}
}

func TestFocusDeferredToTransitionDone(t *testing.T) {
	// Use a spy to track Focus/Blur messages.
	spy := &focusSpy{}
	secs := [SectionCount]SectionModel{}
	for i := range secs {
		secs[i] = newPlaceholderSection(SectionName(Section(i)), DarkTheme())
	}
	secs[SectionWork] = spy

	m := New(testContent(), secs[0], secs[1], secs[2], secs[3])
	// Skip intro.
	result, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = result.(Model)
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(" ")})
	m = result.(Model)
	result, _ = m.Update(IntroDoneMsg{})
	m = result.(Model)

	// Navigate to work (spy is at index 1).
	result, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	m = result.(Model)

	// During transition, spy should NOT have received FocusMsg yet.
	if spy.focusCount > 0 {
		t.Error("expected no FocusMsg during transition")
	}

	// Drain transition.
	_ = drainTransition(t, m)

	// Now spy should have received exactly one FocusMsg.
	if spy.focusCount != 1 {
		t.Errorf("expected 1 FocusMsg after transition, got %d", spy.focusCount)
	}
}

// focusSpy tracks Focus and Blur messages for testing deferred focus.
type focusSpy struct {
	focusCount int
	blurCount  int
}

func (s *focusSpy) Init() tea.Cmd { return nil }

func (s *focusSpy) Update(msg tea.Msg) (SectionModel, tea.Cmd) {
	switch msg.(type) {
	case FocusMsg:
		s.focusCount++
	case BlurMsg:
		s.blurCount++
	}
	return s, nil
}

func (s *focusSpy) View() string { return "" }
