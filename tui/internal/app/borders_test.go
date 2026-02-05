package app

import (
	"strings"
	"testing"
)

func testTheme() Theme {
	return DarkTheme()
}

// ---------------------------------------------------------------------------
// RenderCard
// ---------------------------------------------------------------------------

func TestRenderCardContainsBorderCharacters(t *testing.T) {
	out := RenderCard(testTheme(), "Title", "Hello world", 40)

	for _, ch := range []string{"┌", "┐", "└", "┘", "│", "─"} {
		if !strings.Contains(out, ch) {
			t.Errorf("RenderCard output missing border character %q", ch)
		}
	}
}

func TestRenderCardContainsTitle(t *testing.T) {
	out := RenderCard(testTheme(), "Projects", "content here", 40)
	if !strings.Contains(out, "Projects") {
		t.Error("RenderCard output should contain the title text")
	}
}

func TestRenderCardContainsContent(t *testing.T) {
	out := RenderCard(testTheme(), "Title", "Hello world", 40)
	if !strings.Contains(out, "Hello world") {
		t.Error("RenderCard output should contain the content text")
	}
}

func TestRenderCardRespectsWidth(t *testing.T) {
	widths := []int{20, 40, 60, 80}
	for _, w := range widths {
		out := RenderCard(testTheme(), "Title", "Some content for the card", w)
		lines := strings.Split(out, "\n")
		for i, line := range lines {
			plain := stripANSI(line)
			if plain == "" {
				continue
			}
			runeCount := len([]rune(plain))
			if runeCount > w {
				t.Errorf("width=%d line %d has %d runes (exceeds width): %q", w, i, runeCount, plain)
			}
		}
	}
}

func TestRenderCardEmptyContent(t *testing.T) {
	out := RenderCard(testTheme(), "Empty", "", 30)
	if out == "" {
		t.Error("RenderCard with empty content should still produce bordered output")
	}
	if !strings.Contains(out, "┌") {
		t.Error("RenderCard with empty content should still have top-left corner")
	}
	if !strings.Contains(out, "└") {
		t.Error("RenderCard with empty content should still have bottom-left corner")
	}
}

func TestRenderCardNarrowWidthFallback(t *testing.T) {
	content := "fallback"
	out := RenderCard(testTheme(), "T", content, 5)
	if out != content {
		t.Errorf("RenderCard with width < 10 should return raw content, got %q", out)
	}
}

func TestRenderCardWidthExactlyTen(t *testing.T) {
	out := RenderCard(testTheme(), "T", "ok", 10)
	if !strings.Contains(out, "┌") {
		t.Error("RenderCard with width=10 should produce bordered output")
	}
}

func TestRenderCardMultilineContent(t *testing.T) {
	out := RenderCard(testTheme(), "Title", "line one\nline two\nline three", 40)
	if !strings.Contains(out, "line one") {
		t.Error("RenderCard should contain first line")
	}
	if !strings.Contains(out, "line two") {
		t.Error("RenderCard should contain second line")
	}
	if !strings.Contains(out, "line three") {
		t.Error("RenderCard should contain third line")
	}
}

func TestRenderCardLightTheme(t *testing.T) {
	theme := LightTheme()
	out := RenderCard(theme, "Title", "Content", 30)
	if !strings.Contains(out, "Title") {
		t.Error("RenderCard with light theme should contain title")
	}
	if !strings.Contains(out, "Content") {
		t.Error("RenderCard with light theme should contain content")
	}
}

// ---------------------------------------------------------------------------
// RenderDivider
// ---------------------------------------------------------------------------

func TestRenderDividerLength(t *testing.T) {
	out := RenderDivider(testTheme(), 30)
	plain := stripANSI(out)
	if len([]rune(plain)) != 30 {
		t.Errorf("RenderDivider(30) plain length = %d, want 30", len([]rune(plain)))
	}
}

func TestRenderDividerUsesHorizontalChar(t *testing.T) {
	out := RenderDivider(testTheme(), 10)
	if !strings.Contains(out, "─") {
		t.Error("RenderDivider should use ─ character")
	}
}

func TestRenderDividerZeroWidth(t *testing.T) {
	out := RenderDivider(testTheme(), 0)
	if out != "" {
		t.Errorf("RenderDivider(0) should return empty string, got %q", out)
	}
}

func TestRenderDividerNegativeWidth(t *testing.T) {
	out := RenderDivider(testTheme(), -5)
	if out != "" {
		t.Errorf("RenderDivider(-5) should return empty string, got %q", out)
	}
}

// ---------------------------------------------------------------------------
// wrapText
// ---------------------------------------------------------------------------

func TestWrapTextBreaksAtWordBoundaries(t *testing.T) {
	text := "the quick brown fox jumps over the lazy dog"
	lines := wrapText(text, 20)

	for _, line := range lines {
		if len([]rune(line)) > 20 {
			t.Errorf("wrapped line exceeds width 20: %q (%d runes)", line, len([]rune(line)))
		}
	}

	joined := strings.Join(lines, " ")
	if joined != text {
		t.Errorf("wrapText lost words:\n  got:  %q\n  want: %q", joined, text)
	}
}

func TestWrapTextPreservesNewlines(t *testing.T) {
	text := "line one\nline two"
	lines := wrapText(text, 80)
	if len(lines) != 2 {
		t.Errorf("wrapText should produce 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestWrapTextEmptyString(t *testing.T) {
	lines := wrapText("", 40)
	if len(lines) != 1 || lines[0] != "" {
		t.Errorf("wrapText(\"\") should return [\"\"], got %v", lines)
	}
}

func TestWrapTextSingleLongWord(t *testing.T) {
	word := "supercalifragilisticexpialidocious"
	lines := wrapText(word, 10)
	if len(lines) != 1 || lines[0] != word {
		t.Errorf("wrapText should keep single long word intact, got %v", lines)
	}
}

func TestWrapTextZeroWidth(t *testing.T) {
	lines := wrapText("hello world", 0)
	if len(lines) != 1 {
		t.Errorf("wrapText with width=0 should return original text as single element, got %v", lines)
	}
}

// ---------------------------------------------------------------------------
// padRight
// ---------------------------------------------------------------------------

func TestPadRightShortString(t *testing.T) {
	result := padRight("hi", 10)
	if len(result) != 10 {
		t.Errorf("padRight(\"hi\", 10) len = %d, want 10", len(result))
	}
	if result != "hi        " {
		t.Errorf("padRight(\"hi\", 10) = %q", result)
	}
}

func TestPadRightExactWidth(t *testing.T) {
	result := padRight("hello", 5)
	if result != "hello" {
		t.Errorf("padRight(\"hello\", 5) = %q, want %q", result, "hello")
	}
}

func TestPadRightLongString(t *testing.T) {
	result := padRight("hello world", 5)
	if result != "hello world" {
		t.Errorf("padRight should not truncate, got %q", result)
	}
}

// ---------------------------------------------------------------------------
// Exported constants
// ---------------------------------------------------------------------------

func TestExportedBorderConstants(t *testing.T) {
	if BorderTopLeft != "┌" {
		t.Errorf("BorderTopLeft = %q, want %q", BorderTopLeft, "┌")
	}
	if BorderTopRight != "┐" {
		t.Errorf("BorderTopRight = %q, want %q", BorderTopRight, "┐")
	}
	if BorderBottomLeft != "└" {
		t.Errorf("BorderBottomLeft = %q, want %q", BorderBottomLeft, "└")
	}
	if BorderBottomRight != "┘" {
		t.Errorf("BorderBottomRight = %q, want %q", BorderBottomRight, "┘")
	}
	if BorderHorizontal != "─" {
		t.Errorf("BorderHorizontal = %q, want %q", BorderHorizontal, "─")
	}
	if BorderVertical != "│" {
		t.Errorf("BorderVertical = %q, want %q", BorderVertical, "│")
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// stripANSI removes ANSI escape sequences from a string for width measurement.
func stripANSI(s string) string {
	var result strings.Builder
	inEscape := false
	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}
	return result.String()
}
