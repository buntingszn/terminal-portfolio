package app

import (
	"strings"
	"testing"
)

func TestShimmerNewDefaults(t *testing.T) {
	s := NewShimmer("test", DarkTheme())
	if s.id != "test" {
		t.Errorf("id = %q, want %q", s.id, "test")
	}
	if s.active {
		t.Error("new shimmer should not be active")
	}
	if s.baseL <= 0 {
		t.Error("baseL should be positive")
	}
	if s.peakL <= 0 {
		t.Error("peakL should be positive")
	}
}

func TestShimmerStartStop(t *testing.T) {
	s := NewShimmer("test", DarkTheme())
	cmd := s.Start()
	if !s.active {
		t.Error("expected active after Start")
	}
	if cmd == nil {
		t.Error("expected non-nil cmd from Start")
	}

	s.Stop()
	if s.active {
		t.Error("expected inactive after Stop")
	}
}

func TestShimmerUpdateAdvancesFrame(t *testing.T) {
	s := NewShimmer("test", DarkTheme())
	s.Start()

	s, cmd := s.Update(shimmerTickMsg{id: "test"})
	if s.frame != 1 {
		t.Errorf("frame = %d, want 1", s.frame)
	}
	if cmd == nil {
		t.Error("expected non-nil cmd for next tick")
	}
}

func TestShimmerUpdateIgnoresWrongID(t *testing.T) {
	s := NewShimmer("test", DarkTheme())
	s.Start()

	s, cmd := s.Update(shimmerTickMsg{id: "other"})
	if s.frame != 0 {
		t.Error("frame should not change on wrong ID")
	}
	if cmd != nil {
		t.Error("expected nil cmd for wrong ID")
	}
}

func TestShimmerUpdateIgnoresWhenInactive(t *testing.T) {
	s := NewShimmer("test", DarkTheme())
	s, cmd := s.Update(shimmerTickMsg{id: "test"})
	if cmd != nil {
		t.Error("expected nil cmd when inactive")
	}
	if s.active {
		t.Error("should remain inactive")
	}
}

func TestShimmerRenderProducesOutput(t *testing.T) {
	s := NewShimmer("test", DarkTheme())
	s.Start()
	// Advance a few frames so waves are in range.
	for range 10 {
		s, _ = s.Update(shimmerTickMsg{id: "test"})
	}

	text := "⣿⣿⣿⢿⣿\n⣿⡟⡼⢠⣈"
	result := s.Render(text, 5)
	if result == "" {
		t.Error("render should produce non-empty output")
	}
	if !strings.Contains(result, "\n") {
		t.Error("render should preserve newlines")
	}
}

func TestShimmerRenderEmptyBraille(t *testing.T) {
	s := NewShimmer("test", DarkTheme())
	s.Start()

	text := "⠀⠀⠀"
	result := s.Render(text, 3)
	if !strings.Contains(result, "⠀") {
		t.Error("empty Braille should be preserved in output")
	}
}

func TestShimmerRenderZeroWidth(t *testing.T) {
	s := NewShimmer("test", DarkTheme())
	s.Start()

	result := s.Render("text", 0)
	if result != "text" {
		t.Errorf("zero width render should return original text, got %q", result)
	}
}

func TestShimmerBrightnessAtVariation(t *testing.T) {
	s := NewShimmer("test", DarkTheme())
	s.Start()
	// Advance frames so primary wave is centered around col 5.
	for range 42 { // 42 * 0.12 ≈ phase 5.0
		s, _ = s.Update(shimmerTickMsg{id: "test"})
	}

	// Cells near the wave should have non-zero brightness.
	near := s.brightnessAt(0, 5, 20)
	// Cells far from the wave should be zero or very low.
	far := s.brightnessAt(0, 15, 20)

	if near <= far {
		t.Errorf("near-wave brightness (%f) should exceed far-wave (%f)", near, far)
	}
}

func TestShimmerOutputIsAchromatic(t *testing.T) {
	s := NewShimmer("test", DarkTheme())
	s.Start()
	for range 20 {
		s, _ = s.Update(shimmerTickMsg{id: "test"})
	}

	// Render and check that all hex colors in the output are grey (r == g == b).
	text := "⣿⣿⣿⢿⣿"
	result := s.Render(text, 5)

	// Extract hex colors: #rrggbb patterns.
	for i := 0; i < len(result)-6; i++ {
		if result[i] == '#' {
			hex := result[i : i+7]
			if len(hex) == 7 && isHexDigits(hex[1:]) {
				r := hex[1:3]
				g := hex[3:5]
				b := hex[5:7]
				if r != g || g != b {
					t.Errorf("non-grey color found in shimmer output: %s", hex)
				}
			}
		}
	}
}

func isHexDigits(s string) bool {
	for _, c := range s {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
