package app

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
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

func TestRenderCardLongTitle(t *testing.T) {
	// Title is 54 chars. With width=20, innerWidth=16.
	// Title (54) > innerWidth (16), so truncate to 13 chars + "..." = 16.
	longTitle := "This Is A Very Long Title That Exceeds The Card Width"
	out := RenderCard(testTheme(), longTitle, "body", 20)
	plain := stripANSI(out)

	// The full title should NOT appear.
	if strings.Contains(plain, longTitle) {
		t.Error("RenderCard should truncate a title longer than inner width")
	}

	// The truncated title should end with "...".
	if !strings.Contains(plain, "...") {
		t.Error("Truncated title should contain ellipsis '...'")
	}

	// Verify all lines respect the width.
	for i, line := range strings.Split(out, "\n") {
		p := stripANSI(line)
		if p == "" {
			continue
		}
		if len([]rune(p)) > 20 {
			t.Errorf("line %d exceeds width 20: %d runes: %q", i, len([]rune(p)), p)
		}
	}
}

func TestRenderCardMinimalWidth(t *testing.T) {
	// Test widths 10 through 15 to verify bordered output without panics.
	for w := 10; w <= 15; w++ {
		out := RenderCard(testTheme(), "T", "ok", w)
		if !strings.Contains(out, "┌") {
			t.Errorf("width=%d: missing top-left corner", w)
		}
		if !strings.Contains(out, "└") {
			t.Errorf("width=%d: missing bottom-left corner", w)
		}
		if !strings.Contains(out, "┐") {
			t.Errorf("width=%d: missing top-right corner", w)
		}
		if !strings.Contains(out, "┘") {
			t.Errorf("width=%d: missing bottom-right corner", w)
		}

		// Verify no line exceeds the target width.
		for i, line := range strings.Split(out, "\n") {
			p := stripANSI(line)
			if p == "" {
				continue
			}
			if len([]rune(p)) > w {
				t.Errorf("width=%d line %d has %d runes (exceeds): %q", w, i, len([]rune(p)), p)
			}
		}
	}
}

func TestRenderCardEmptyContentRendersBox(t *testing.T) {
	out := RenderCard(testTheme(), "Header", "", 30)

	// Should produce a bordered box with an empty body line.
	if !strings.Contains(out, "Header") {
		t.Error("Card with empty content should still contain the title")
	}

	// Count body lines (lines between top and bottom border).
	lines := strings.Split(out, "\n")
	// Expected: top border, at least one body line, bottom border.
	if len(lines) < 3 {
		t.Errorf("Card with empty content should have at least 3 lines (top, body, bottom), got %d", len(lines))
	}

	// The body line should contain vertical border chars.
	bodyFound := false
	for _, line := range lines[1 : len(lines)-1] {
		plain := stripANSI(line)
		if strings.Contains(plain, "│") {
			bodyFound = true
		}
	}
	if !bodyFound {
		t.Error("Card with empty content should have at least one body line with vertical borders")
	}
}

func TestRenderCardWidthBelowMinimum(t *testing.T) {
	// Width < 10 should return content without borders.
	for _, w := range []int{0, 1, 5, 9} {
		content := "raw content"
		out := RenderCard(testTheme(), "Title", content, w)
		if out != content {
			t.Errorf("width=%d: expected raw content %q, got %q", w, content, out)
		}
	}

	// Negative width.
	out := RenderCard(testTheme(), "Title", "text", -1)
	if out != "text" {
		t.Errorf("width=-1: expected raw content, got %q", out)
	}

	// Width 0 with empty content.
	out = RenderCard(testTheme(), "Title", "", 0)
	if out != "" {
		t.Errorf("width=0, empty content: expected empty string, got %q", out)
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

func TestPadRightWithAnsi(t *testing.T) {
	// ANSI-styled text has more bytes/runes than visual width.
	// lipgloss.Width should measure visual width correctly.
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("#ff0000"))
	styled := style.Render("hi")

	// Visual width of "hi" is 2; pad to 10.
	result := padRight(styled, 10)
	visualWidth := lipgloss.Width(result)
	if visualWidth != 10 {
		t.Errorf("padRight with ANSI: visual width = %d, want 10", visualWidth)
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
