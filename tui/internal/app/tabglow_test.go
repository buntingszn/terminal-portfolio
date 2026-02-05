package app

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestTabGlowNewDefaults(t *testing.T) {
	g := NewTabGlow(DarkTheme())
	if g.active {
		t.Error("new TabGlow should not be active")
	}
	if g.steps <= 0 {
		t.Error("steps should be positive")
	}
}

func TestTabGlowStartAndActive(t *testing.T) {
	g := NewTabGlow(DarkTheme())
	cmd := g.Start()
	if !g.active {
		t.Error("expected active after Start")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd from Start")
	}
}

func TestTabGlowUpdateAdvances(t *testing.T) {
	g := NewTabGlow(DarkTheme())
	g.Start()

	g, cmd := g.Update(tabGlowTickMsg{})
	if g.step != 1 {
		t.Errorf("step = %d, want 1", g.step)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for next tick")
	}
}

func TestTabGlowCompletesAfterAllSteps(t *testing.T) {
	g := NewTabGlow(DarkTheme())
	g.Start()

	for g.Active() {
		var cmd tea.Cmd
		g, cmd = g.Update(tabGlowTickMsg{})
		if !g.active && cmd != nil {
			t.Error("expected nil cmd after completion")
		}
	}

	if g.active {
		t.Error("expected inactive after all steps")
	}
}

func TestTabGlowIgnoresWhenInactive(t *testing.T) {
	g := NewTabGlow(DarkTheme())
	g, cmd := g.Update(tabGlowTickMsg{})
	if cmd != nil {
		t.Error("expected nil cmd when inactive")
	}
	if g.step != 0 {
		t.Error("step should remain 0 when inactive")
	}
}

func TestTabGlowIgnoresWrongMsg(t *testing.T) {
	g := NewTabGlow(DarkTheme())
	g.Start()
	g, cmd := g.Update(shimmerTickMsg{id: "wrong"})
	if cmd != nil {
		t.Error("expected nil cmd for wrong msg type")
	}
	if g.step != 0 {
		t.Error("step should remain 0 for wrong msg")
	}
}

func TestTabGlowBrightenedAccentMidPulse(t *testing.T) {
	g := NewTabGlow(DarkTheme())
	g.Start()

	// Advance to about halfway.
	halfSteps := g.steps / 2
	for range halfSteps {
		g, _ = g.Update(tabGlowTickMsg{})
	}

	brightened := g.BrightenedAccent()
	original := g.theme.Colors.Accent
	if brightened == original {
		t.Error("brightened accent at mid-pulse should differ from original")
	}
}

func TestTabGlowBrightenedAccentWhenInactive(t *testing.T) {
	g := NewTabGlow(DarkTheme())
	// Not started — should return the unchanged accent.
	result := g.BrightenedAccent()
	if result != g.theme.Colors.Accent {
		t.Error("inactive glow should return unchanged accent color")
	}
}

func TestTabGlowSetTheme(t *testing.T) {
	g := NewTabGlow(DarkTheme())
	g.SetTheme(LightTheme())
	if g.theme.IsDark {
		t.Error("SetTheme should update theme")
	}
}
