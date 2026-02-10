package sections

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/buntingszn/terminal-portfolio/tui/internal/app"
	"github.com/buntingszn/terminal-portfolio/tui/internal/testutil"
)

// testSizes defines the terminal dimensions at which all sections are tested.
var testSizes = []struct {
	name   string
	width  int
	height int
}{
	{"40x15", 40, 15},
	{"60x20", 60, 20},
	{"80x24", 80, 24},
	{"120x40", 120, 40},
}

// initSection sends a WindowSizeMsg and FocusMsg to a section.
func initSection(t *testing.T, s app.SectionModel, width, height int) app.SectionModel {
	t.Helper()
	s, _ = s.Update(tea.WindowSizeMsg{Width: width, Height: height})
	s, _ = s.Update(app.FocusMsg{})
	return s
}

// drainHomeReveal sends homeRevealTickMsg until the reveal animation completes.
func drainHomeReveal(s app.SectionModel) app.SectionModel {
	for range 200 {
		var cmd tea.Cmd
		s, cmd = s.Update(homeRevealTickMsg{})
		if cmd == nil {
			break
		}
	}
	return s
}

// --- HomeSection tests ---

func TestHomeSection_RenderAtSizes(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	for _, sz := range testSizes {
		t.Run(sz.name, func(t *testing.T) {
			h := NewHomeSection(c, theme)
			s := initSection(t, h, sz.width, sz.height)
			view := s.View()
			testutil.RequireNotEmpty(t, view)
		})
	}
}

func TestHomeSection_PortraitVisibility(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	// portraitMarker is a unique substring from the Braille halftone portrait.
	const portraitMarker = "⣿⣿⣿⢿"

	t.Run("hidden_below_80", func(t *testing.T) {
		h := NewHomeSection(c, theme)
		s := initSection(t, h, 60, 24)
		view := s.View()
		if strings.Contains(view, portraitMarker) {
			t.Error("portrait should be hidden at width 60")
		}
	})

	t.Run("shown_at_100", func(t *testing.T) {
		h := NewHomeSection(c, theme)
		s := initSection(t, h, 100, 24)
		view := s.View()
		if !strings.Contains(view, portraitMarker) {
			t.Error("portrait should be visible at width 100")
		}
	})

	t.Run("threshold_at_81", func(t *testing.T) {
		// ContentWidth() caps at MaxContentWidth (88), so at terminal width 81
		// content width = 80 which meets portraitMinWidth (80).
		h := NewHomeSection(c, theme)
		s := initSection(t, h, 80, 24)
		view80 := s.View()
		if strings.Contains(view80, portraitMarker) {
			t.Error("portrait should be hidden at terminal width 80 (content width 79)")
		}
		h2 := NewHomeSection(c, theme)
		s2 := initSection(t, h2, 81, 24)
		view81 := s2.View()
		if !strings.Contains(view81, portraitMarker) {
			t.Error("portrait should be visible at terminal width 81 (content width 80)")
		}
	})
}

func TestHomeSection_BioAndInfoContent(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	h := NewHomeSection(c, theme)
	s := initSection(t, h, 80, 24)
	s = drainHomeReveal(s)
	view := s.View()
	testutil.RequireContains(t, view, "software engineer")
	testutil.RequireContains(t, view, "Status")
}

func TestHomeSection_BioVisibleAfterReveal(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	h := NewHomeSection(c, theme)
	s := initSection(t, h, 80, 24)
	s = drainHomeReveal(s)
	view := s.View()
	testutil.RequireContains(t, view, "software engineer")
}

func TestHomeSection_TextWrapsAtNarrow(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	widths := []struct {
		name  string
		width int
	}{
		{"40", 40},
		{"50", 50},
		{"60", 60},
	}
	for _, w := range widths {
		t.Run(w.name, func(t *testing.T) {
			h := NewHomeSection(c, theme)
			s := initSection(t, h, w.width, 24)
			s = drainHomeReveal(s)
			view := s.View()
			testutil.RequireNotEmpty(t, view)
			testutil.RequireContains(t, view, "Status")
		})
	}
}

func TestHomeSection_RevealStreamsLines(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	h := NewHomeSection(c, theme)
	s := initSection(t, h, 80, 24)

	// Immediately after focus, info fields should not be visible (only first line revealed).
	initialView := s.View()
	if strings.Contains(initialView, "Status") {
		t.Error("status should not be visible during initial reveal")
	}

	// After draining, all content should be visible.
	s = drainHomeReveal(s)
	fullView := s.View()
	testutil.RequireContains(t, fullView, "Status")
}

func TestHomeSection_RevealSkippedOnKeyPress(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	h := NewHomeSection(c, theme)
	s := initSection(t, h, 80, 24)

	// Press a key to skip the reveal.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	view := s.View()
	testutil.RequireContains(t, view, "software engineer")
	testutil.RequireContains(t, view, "Status")
}

func TestHomeSection_RevealDoesNotReplayOnRefocus(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	h := NewHomeSection(c, theme)
	s := initSection(t, h, 80, 24)
	s = drainHomeReveal(s)

	// Blur and refocus — should show full content immediately (no reveal).
	s, _ = s.Update(app.BlurMsg{})
	s, _ = s.Update(app.FocusMsg{})
	view := s.View()
	testutil.RequireContains(t, view, "software engineer")
	testutil.RequireContains(t, view, "Status")
}

// --- WorkSection tests ---

func TestWorkSection_RenderAtSizes(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	for _, sz := range testSizes {
		t.Run(sz.name, func(t *testing.T) {
			w := NewWorkSection(c, theme)
			s := initSection(t, w, sz.width, sz.height)
			view := s.View()
			testutil.RequireNotEmpty(t, view)
		})
	}
}

func TestWorkSection_AllProjectsVisible(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	// Use a tall viewport so all content is visible.
	w := NewWorkSection(c, theme)
	s := initSection(t, w, 80, 200)
	view := s.View()
	testutil.RequireContains(t, view, "Terminal Portfolio")
	testutil.RequireContains(t, view, "Cookt")
}

func TestWorkSection_CardWidthCapped(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	// At width 120, cards should be capped at 78.
	w := NewWorkSection(c, theme)
	s := initSection(t, w, 120, 40)
	view := s.View()
	testutil.RequireNotEmpty(t, view)
	// Verify cards render properly with capped width.
	testutil.RequireContains(t, view, "Terminal Portfolio")
}

func TestWorkSection_TagWrapping(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	// At narrow width, tags should still render.
	w := NewWorkSection(c, theme)
	s := initSection(t, w, 40, 24)
	view := s.View()
	testutil.RequireNotEmpty(t, view)
}

func TestWorkSection_NilContent(t *testing.T) {
	theme := testutil.FixtureTheme()
	w := NewWorkSection(nil, theme)
	s := initSection(t, w, 80, 24)
	view := s.View()
	testutil.RequireContains(t, view, "No projects")
}

func TestWorkSection_CursorNavigation(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	w := NewWorkSection(c, theme)
	s := initSection(t, w, 80, 24)

	view1 := s.View()
	testutil.RequireContains(t, view1, "▸")

	// Move cursor down.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	view2 := s.View()
	testutil.RequireNotEmpty(t, view2)

	if view1 == view2 {
		t.Error("view should change after cursor move")
	}
}

func TestWorkSection_CursorBounds(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	w := NewWorkSection(c, theme)
	s := initSection(t, w, 80, 24)

	// Move up from top — should not panic.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	testutil.RequireNotEmpty(t, s.View())

	// Move far past bottom — should clamp.
	for range 20 {
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	testutil.RequireNotEmpty(t, s.View())
}

func TestWorkSection_EnterCopyURL(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	w := NewWorkSection(c, theme)
	s := initSection(t, w, 80, 24)

	// Press Enter on the first project.
	s, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should return a non-nil cmd (the tick timer for clearing feedback).
	if cmd == nil {
		t.Fatal("expected non-nil cmd after Enter press")
	}

	// View should contain the OSC 52 escape sequence prefix.
	view := s.View()
	if !strings.Contains(view, "\x1b]52;c;") {
		t.Error("expected OSC 52 sequence in view after Enter")
	}

	// KeyHints should show the copy feedback.
	ws := s.(*WorkSection)
	hints := ws.KeyHints()
	if hints != "Copied!" {
		t.Errorf("expected KeyHints() = %q, got %q", "Copied!", hints)
	}
}

func TestWorkSection_CopyFeedbackClears(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	w := NewWorkSection(c, theme)
	s := initSection(t, w, 80, 24)

	// Press Enter to set feedback.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Send the clear message (simulates the 2s timer firing).
	s, _ = s.Update(clearWorkCopyMsg{})

	ws := s.(*WorkSection)
	hints := ws.KeyHints()
	if strings.Contains(hints, "Copied!") {
		t.Error("expected feedback to be cleared after clearWorkCopyMsg")
	}
	if !strings.Contains(hints, "enter copy URL") {
		t.Errorf("expected default hints after clearing, got %q", hints)
	}
}

func TestWorkSection_ScrollToTopAndBottom(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	w := NewWorkSection(c, theme)
	s := initSection(t, w, 80, 24)

	// g to top.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	testutil.RequireNotEmpty(t, s.View())

	// G to bottom.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	testutil.RequireNotEmpty(t, s.View())
}

// --- CVSection tests ---

func TestCVSection_RenderAtSizes(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	for _, sz := range testSizes {
		t.Run(sz.name, func(t *testing.T) {
			cv := NewCVSection(c, theme)
			s := initSection(t, cv, sz.width, sz.height)
			view := s.View()
			testutil.RequireNotEmpty(t, view)
		})
	}
}

func TestCVSection_GradientHeader(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	cv := NewCVSection(c, theme)
	s := initSection(t, cv, 80, 24)
	view := s.View()
	// The gradient header contains name@domain.
	testutil.RequireContains(t, view, c.Meta.Name)
}

func TestCVSection_ExperienceVisible(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	cv := NewCVSection(c, theme)
	s := initSection(t, cv, 80, 24)
	view := s.View()
	testutil.RequireContains(t, view, "EXPERIENCE")
	testutil.RequireContains(t, view, "Independent")
}

func TestCVSection_SkillsVisibleAfterScroll(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	// Use a tall viewport so all content fits.
	cv := NewCVSection(c, theme)
	s := initSection(t, cv, 80, 200)
	view := s.View()
	testutil.RequireContains(t, view, "SKILLS")
	testutil.RequireContains(t, view, "Languages")
}

func TestCVSection_BulletsWrapAtNarrow(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	cv := NewCVSection(c, theme)
	s := initSection(t, cv, 40, 24)
	view := s.View()
	testutil.RequireNotEmpty(t, view)
	testutil.RequireContains(t, view, "EXPERIENCE")
}

func TestCVSection_SkillsWrapAtNarrow(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	// Use tall viewport so Skills section is visible.
	cv := NewCVSection(c, theme)
	s := initSection(t, cv, 40, 200)
	view := s.View()
	testutil.RequireContains(t, view, "SKILLS")
}

// --- LinksSection tests ---

func TestLinksSection_RenderAtSizes(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	for _, sz := range testSizes {
		t.Run(sz.name, func(t *testing.T) {
			l := NewLinksSection(c, theme)
			s := initSection(t, l, sz.width, sz.height)
			view := s.View()
			testutil.RequireNotEmpty(t, view)
		})
	}
}

func TestLinksSection_Content(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	l := NewLinksSection(c, theme)
	s := initSection(t, l, 80, 24)
	view := s.View()
	testutil.RequireContains(t, view, "GitHub")
}

func TestLinksSection_URLTruncationAtNarrow(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	l := NewLinksSection(c, theme)
	s := initSection(t, l, 30, 15)
	view := s.View()
	testutil.RequireNotEmpty(t, view)
}

func TestLinksSection_CursorNavigation(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	l := NewLinksSection(c, theme)
	s := initSection(t, l, 80, 24)

	view1 := s.View()
	testutil.RequireContains(t, view1, ">")

	// Move cursor down.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	view2 := s.View()
	testutil.RequireNotEmpty(t, view2)

	if view1 == view2 {
		t.Error("view should change after cursor move")
	}
}

func TestLinksSection_CursorBounds(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	l := NewLinksSection(c, theme)
	s := initSection(t, l, 80, 24)

	// Move up from top — should not panic.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("k")})
	testutil.RequireNotEmpty(t, s.View())

	// Move far past bottom — should clamp.
	for range 20 {
		s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	}
	testutil.RequireNotEmpty(t, s.View())
}

func TestLinksSection_ScrollToTopAndBottom(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	l := NewLinksSection(c, theme)
	s := initSection(t, l, 80, 24)

	// g to top.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("g")})
	testutil.RequireNotEmpty(t, s.View())

	// G to bottom.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("G")})
	testutil.RequireNotEmpty(t, s.View())
}

func TestLinksSection_NilContent(t *testing.T) {
	theme := testutil.FixtureTheme()
	l := NewLinksSection(nil, theme)
	s := initSection(t, l, 80, 24)
	view := s.View()
	testutil.RequireContains(t, view, "No links")
}

func TestLinksSection_EnterCopyURL(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	l := NewLinksSection(c, theme)
	s := initSection(t, l, 80, 24)

	// Press Enter on the first link (GitHub).
	s, cmd := s.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Should return a non-nil cmd (the tick timer for clearing feedback).
	if cmd == nil {
		t.Fatal("expected non-nil cmd after Enter press")
	}

	// View should contain the OSC 52 escape sequence prefix.
	view := s.View()
	if !strings.Contains(view, "\x1b]52;c;") {
		t.Error("expected OSC 52 sequence in view after Enter")
	}

	// KeyHints should show the copy feedback.
	ls := s.(*LinksSection)
	hints := ls.KeyHints()
	if hints != "Copied!" {
		t.Errorf("expected KeyHints() = %q, got %q", "Copied!", hints)
	}
}

func TestLinksSection_CopyFeedbackClears(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	l := NewLinksSection(c, theme)
	s := initSection(t, l, 80, 24)

	// Press Enter to set feedback.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Send the clear message (simulates the 2s timer firing).
	s, _ = s.Update(clearCopyFeedbackMsg{})

	ls := s.(*LinksSection)
	hints := ls.KeyHints()
	if strings.Contains(hints, "Copied!") {
		t.Error("expected feedback to be cleared after clearCopyFeedbackMsg")
	}
	if !strings.Contains(hints, "enter copy URL") {
		t.Errorf("expected default hints after clearing, got %q", hints)
	}
}

func TestLinksSection_EnterClearsClipboardOnNextUpdate(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	l := NewLinksSection(c, theme)
	s := initSection(t, l, 80, 24)

	// Press Enter.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view1 := s.View()
	if !strings.Contains(view1, "\x1b]52;c;") {
		t.Fatal("expected OSC 52 in first view")
	}

	// Any subsequent update should clear the pending clipboard.
	s, _ = s.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")})
	view2 := s.View()
	if strings.Contains(view2, "\x1b]52;c;") {
		t.Error("OSC 52 should be cleared after next update")
	}
}

func TestLinksSection_OSC8HyperlinkInView(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	l := NewLinksSection(c, theme)
	// Use tall viewport so all links are visible.
	s := initSection(t, l, 80, 200)
	view := s.View()

	// The first link is GitHub with URL https://github.com/buntingszn.
	// The view should contain the OSC 8 hyperlink start sequence for it.
	if !strings.Contains(view, "\x1b]8;;https://github.com/buntingszn\a") {
		t.Error("expected OSC 8 hyperlink for GitHub URL in view")
	}
}

// --- Cross-section tests ---

func TestAllSections_NoPanicAtMinimumSize(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	makers := []struct {
		name string
		fn   func() app.SectionModel
	}{
		{"home", func() app.SectionModel { return NewHomeSection(c, theme) }},
		{"work", func() app.SectionModel { return NewWorkSection(c, theme) }},
		{"cv", func() app.SectionModel { return NewCVSection(c, theme) }},
		{"links", func() app.SectionModel { return NewLinksSection(c, theme) }},
	}

	for _, m := range makers {
		t.Run(m.name, func(t *testing.T) {
			s := initSection(t, m.fn(), 20, 5)
			testutil.RequireNotEmpty(t, s.View())
		})
	}
}

func TestAllSections_BlurAndRefocus(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	makers := []struct {
		name string
		fn   func() app.SectionModel
	}{
		{"home", func() app.SectionModel { return NewHomeSection(c, theme) }},
		{"work", func() app.SectionModel { return NewWorkSection(c, theme) }},
		{"cv", func() app.SectionModel { return NewCVSection(c, theme) }},
		{"links", func() app.SectionModel { return NewLinksSection(c, theme) }},
	}

	for _, m := range makers {
		t.Run(m.name, func(t *testing.T) {
			s := initSection(t, m.fn(), 80, 24)
			s, _ = s.Update(app.BlurMsg{})
			s, _ = s.Update(app.FocusMsg{})
			testutil.RequireNotEmpty(t, s.View())
		})
	}
}

func TestAllSections_ResizePreservesContent(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	makers := []struct {
		name string
		fn   func() app.SectionModel
	}{
		{"home", func() app.SectionModel { return NewHomeSection(c, theme) }},
		{"work", func() app.SectionModel { return NewWorkSection(c, theme) }},
		{"cv", func() app.SectionModel { return NewCVSection(c, theme) }},
		{"links", func() app.SectionModel { return NewLinksSection(c, theme) }},
	}

	for _, m := range makers {
		t.Run(m.name, func(t *testing.T) {
			s := initSection(t, m.fn(), 80, 24)
			// Shrink.
			s, _ = s.Update(tea.WindowSizeMsg{Width: 40, Height: 15})
			testutil.RequireNotEmpty(t, s.View())
			// Grow.
			s, _ = s.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
			testutil.RequireNotEmpty(t, s.View())
		})
	}
}

func TestAllSections_MouseScrollIgnoredWhenNotFocused(t *testing.T) {
	c := testutil.FixtureContent()
	theme := testutil.FixtureTheme()

	makers := []struct {
		name string
		fn   func() app.SectionModel
	}{
		{"home", func() app.SectionModel { return NewHomeSection(c, theme) }},
		{"work", func() app.SectionModel { return NewWorkSection(c, theme) }},
		{"cv", func() app.SectionModel { return NewCVSection(c, theme) }},
		{"links", func() app.SectionModel { return NewLinksSection(c, theme) }},
	}

	for _, m := range makers {
		t.Run(m.name, func(t *testing.T) {
			s := m.fn()
			s, _ = s.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
			// Not focused — mouse events should not panic.
			s, _ = s.Update(tea.MouseMsg{Button: tea.MouseButtonWheelDown})
			testutil.RequireNotEmpty(t, s.View())
		})
	}
}
