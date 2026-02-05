package app

import (
	"strings"
	"testing"
)

func TestGradientAnimNewDefaults(t *testing.T) {
	g := NewGradientAnim("test", DarkTheme())
	if g.id != "test" {
		t.Errorf("id = %q, want %q", g.id, "test")
	}
	if g.active {
		t.Error("new GradientAnim should not be active")
	}
	if g.frame != 0 {
		t.Error("new GradientAnim frame should be 0")
	}
}

func TestGradientAnimStartReturnsCmd(t *testing.T) {
	g := NewGradientAnim("test", DarkTheme())
	cmd := g.Start()
	if !g.active {
		t.Error("expected active after Start")
	}
	if g.frame != 0 {
		t.Error("frame should be 0 after Start")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd from Start")
	}
}

func TestGradientAnimStop(t *testing.T) {
	g := NewGradientAnim("test", DarkTheme())
	g.Start()
	g.Stop()
	if g.active {
		t.Error("expected inactive after Stop")
	}
}

func TestGradientAnimUpdateWrongID(t *testing.T) {
	g := NewGradientAnim("test", DarkTheme())
	g.Start()

	g, cmd := g.Update(gradientAnimTickMsg{id: "other"})
	if g.frame != 0 {
		t.Error("frame should not change on wrong ID")
	}
	if cmd != nil {
		t.Error("expected nil cmd for wrong ID")
	}
}

func TestGradientAnimUpdateAdvancesFrame(t *testing.T) {
	g := NewGradientAnim("test", DarkTheme())
	g.Start()

	g, cmd := g.Update(gradientAnimTickMsg{id: "test"})
	if g.frame != 1 {
		t.Errorf("frame = %d, want 1", g.frame)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for next tick")
	}
}

func TestGradientAnimUpdateInactive(t *testing.T) {
	g := NewGradientAnim("test", DarkTheme())
	g, cmd := g.Update(gradientAnimTickMsg{id: "test"})
	if cmd != nil {
		t.Error("expected nil cmd when inactive")
	}
	if g.active {
		t.Error("should remain inactive")
	}
}

func TestGradientAnimRenderNonEmpty(t *testing.T) {
	g := NewGradientAnim("test", DarkTheme())
	g.Start()
	for range 5 {
		g, _ = g.Update(gradientAnimTickMsg{id: "test"})
	}

	result := g.Render("Hello World")
	if result == "" {
		t.Error("render should produce non-empty output")
	}
	// Result should contain the original text characters.
	if !strings.Contains(result, "H") {
		t.Error("render should contain original text characters")
	}
}

func TestGradientAnimRenderEmpty(t *testing.T) {
	g := NewGradientAnim("test", DarkTheme())
	g.Start()

	result := g.Render("")
	if result != "" {
		t.Errorf("render of empty string should return empty, got %q", result)
	}
}

func TestGradientAnimSetTheme(t *testing.T) {
	g := NewGradientAnim("test", DarkTheme())
	darkStart := g.startC

	g.SetTheme(LightTheme())
	// Dark accent and light accent hex values differ, so startC should change.
	if g.startC == darkStart {
		t.Error("SetTheme should update startC color")
	}
}

func TestGradientAnimActiveReflectsState(t *testing.T) {
	g := NewGradientAnim("test", DarkTheme())
	if g.Active() {
		t.Error("new GradientAnim should not be active")
	}

	g.Start()
	if !g.Active() {
		t.Error("should be active after Start")
	}

	g.Stop()
	if g.Active() {
		t.Error("should be inactive after Stop")
	}
}
