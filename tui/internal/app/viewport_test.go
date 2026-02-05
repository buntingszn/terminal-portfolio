package app

import (
	"strings"
	"testing"
)

func TestNewViewport(t *testing.T) {
	vp := NewViewport(80, 24)
	if vp.width != 80 {
		t.Errorf("NewViewport width = %d, want 80", vp.width)
	}
	if vp.height != 24 {
		t.Errorf("NewViewport height = %d, want 24", vp.height)
	}
}

func TestSetContentAndView(t *testing.T) {
	vp := NewViewport(40, 5)
	content := "line 1\nline 2\nline 3"
	vp.SetContent(content)

	view := vp.View()
	if !strings.Contains(view, "line 1") {
		t.Error("View() should contain 'line 1'")
	}
	if !strings.Contains(view, "line 2") {
		t.Error("View() should contain 'line 2'")
	}
	if !strings.Contains(view, "line 3") {
		t.Error("View() should contain 'line 3'")
	}
}

func TestSetContentResetsOffset(t *testing.T) {
	vp := NewViewport(40, 3)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollToBottom()

	// Setting new content should reset scroll to top.
	vp.SetContent("new content")
	if vp.yOffset != 0 {
		t.Errorf("SetContent should reset yOffset to 0, got %d", vp.yOffset)
	}
}

func TestSetSize(t *testing.T) {
	vp := NewViewport(40, 10)
	vp.SetSize(100, 50)
	if vp.width != 100 {
		t.Errorf("SetSize width = %d, want 100", vp.width)
	}
	if vp.height != 50 {
		t.Errorf("SetSize height = %d, want 50", vp.height)
	}
}

func TestSetSizeClampsOffset(t *testing.T) {
	vp := NewViewport(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollToBottom()

	// Shrink the content area — offset should be clamped.
	vp.SetSize(40, 18)
	max := vp.maxOffset()
	if vp.yOffset > max {
		t.Errorf("yOffset %d exceeds max %d after SetSize", vp.yOffset, max)
	}
}

func TestScrollUpAndDown(t *testing.T) {
	vp := NewViewport(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	vp.ScrollDown(3)
	if vp.yOffset != 3 {
		t.Errorf("after ScrollDown(3), yOffset = %d, want 3", vp.yOffset)
	}

	vp.ScrollUp(1)
	if vp.yOffset != 2 {
		t.Errorf("after ScrollUp(1), yOffset = %d, want 2", vp.yOffset)
	}
}

func TestScrollUpClampsAtZero(t *testing.T) {
	vp := NewViewport(40, 5)
	vp.SetContent("one\ntwo\nthree")
	vp.ScrollUp(10)
	if vp.yOffset != 0 {
		t.Errorf("ScrollUp beyond top: yOffset = %d, want 0", vp.yOffset)
	}
}

func TestScrollDownClampsAtMax(t *testing.T) {
	vp := NewViewport(40, 5)
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollDown(100)

	max := vp.maxOffset()
	if vp.yOffset != max {
		t.Errorf("ScrollDown beyond bottom: yOffset = %d, want %d", vp.yOffset, max)
	}
}

func TestScrollToTopAndBottom(t *testing.T) {
	vp := NewViewport(40, 3)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	vp.ScrollToBottom()
	if !vp.AtBottom() {
		t.Error("ScrollToBottom should put viewport at bottom")
	}

	vp.ScrollToTop()
	if !vp.AtTop() {
		t.Error("ScrollToTop should put viewport at top")
	}
}

func TestAtTopAndAtBottom(t *testing.T) {
	vp := NewViewport(40, 5)
	vp.SetContent("line 1\nline 2\nline 3")

	if !vp.AtTop() {
		t.Error("AtTop() should be true when at the beginning")
	}
	if !vp.AtBottom() {
		t.Error("AtBottom() should be true when content fits in viewport")
	}
}

func TestAtTopAndAtBottomWithLongContent(t *testing.T) {
	vp := NewViewport(40, 3)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	if !vp.AtTop() {
		t.Error("AtTop() should be true initially")
	}
	if vp.AtBottom() {
		t.Error("AtBottom() should be false when more content below")
	}

	vp.ScrollToBottom()
	if vp.AtTop() {
		t.Error("AtTop() should be false after ScrollToBottom()")
	}
	if !vp.AtBottom() {
		t.Error("AtBottom() should be true after ScrollToBottom()")
	}

	vp.ScrollToTop()
	if !vp.AtTop() {
		t.Error("AtTop() should be true after ScrollToTop()")
	}
}

func TestTotalLines(t *testing.T) {
	vp := NewViewport(40, 5)
	content := "line 1\nline 2\nline 3\nline 4\nline 5\nline 6\nline 7"
	vp.SetContent(content)

	total := vp.TotalLines()
	if total != 7 {
		t.Errorf("TotalLines() = %d, want 7", total)
	}
}

func TestVisibleLines(t *testing.T) {
	vp := NewViewport(40, 10)
	visible := vp.VisibleLines()
	if visible != 10 {
		t.Errorf("VisibleLines() = %d, want 10", visible)
	}

	vp.SetSize(40, 25)
	visible = vp.VisibleLines()
	if visible != 25 {
		t.Errorf("VisibleLines() after SetSize = %d, want 25", visible)
	}
}

func TestScrollPercent(t *testing.T) {
	vp := NewViewport(40, 5)
	vp.SetContent("line 1\nline 2\nline 3")

	pct := vp.ScrollPercent()
	if !strings.HasSuffix(pct, "%") {
		t.Errorf("ScrollPercent() = %q, want suffix '%%'", pct)
	}
}

func TestScrollPercentAtTopIsZero(t *testing.T) {
	vp := NewViewport(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	pct := vp.ScrollPercent()
	if pct != "  0%" {
		t.Errorf("ScrollPercent at top = %q, want \"  0%%\"", pct)
	}
}

func TestScrollPercentAtBottomIs100(t *testing.T) {
	vp := NewViewport(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollToBottom()

	pct := vp.ScrollPercent()
	if pct != "100%" {
		t.Errorf("ScrollPercent at bottom = %q, want \"100%%\"", pct)
	}
}

func TestRawScrollPercent(t *testing.T) {
	vp := NewViewport(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	if vp.RawScrollPercent() != 0.0 {
		t.Errorf("RawScrollPercent at top = %f, want 0.0", vp.RawScrollPercent())
	}

	vp.ScrollToBottom()
	if vp.RawScrollPercent() != 1.0 {
		t.Errorf("RawScrollPercent at bottom = %f, want 1.0", vp.RawScrollPercent())
	}
}

func TestViewHeightZero(t *testing.T) {
	vp := NewViewport(40, 0)
	vp.SetContent("some content")
	if vp.View() != "" {
		t.Errorf("View() with height=0 should return empty string, got %q", vp.View())
	}
}

func TestViewShowsCorrectSlice(t *testing.T) {
	vp := NewViewport(40, 3)
	vp.SetContent("a\nb\nc\nd\ne")

	// At top, should see a, b, c.
	view := vp.View()
	if view != "a\nb\nc" {
		t.Errorf("View at top = %q, want %q", view, "a\nb\nc")
	}

	// After scrolling down 2, should see c, d, e.
	vp.ScrollDown(2)
	view = vp.View()
	if view != "c\nd\ne" {
		t.Errorf("View after scroll = %q, want %q", view, "c\nd\ne")
	}
}

func TestViewWithScrollbarHiddenWhenContentFits(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 10)
	vp.SetContent("short content\nonly two lines")

	withScrollbar := vp.ViewWithScrollbar(theme)

	// When content fits, the output is centered (not raw View()).
	// It should still contain the content and no scrollbar chars.
	if !strings.Contains(withScrollbar, "short content") {
		t.Error("ViewWithScrollbar should contain content text when content fits")
	}
	if !strings.Contains(withScrollbar, "only two lines") {
		t.Error("ViewWithScrollbar should contain content text when content fits")
	}

	if strings.Contains(withScrollbar, scrollTrackChar) {
		t.Errorf("ViewWithScrollbar should not contain track char %q when content fits", scrollTrackChar)
	}
	if strings.Contains(withScrollbar, scrollThumbChar) {
		t.Errorf("ViewWithScrollbar should not contain thumb char %q when content fits", scrollThumbChar)
	}

	// Centered output should have exactly viewport height lines.
	resultLines := strings.Split(withScrollbar, "\n")
	if len(resultLines) != 10 {
		t.Errorf("Centered output should have %d lines, got %d", 10, len(resultLines))
	}
}

func TestViewWithScrollbarShownWhenContentOverflows(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 5)

	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "content line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	result := vp.ViewWithScrollbar(theme)
	if !strings.Contains(result, scrollThumbChar) {
		t.Errorf("ViewWithScrollbar should contain thumb char %q when content overflows", scrollThumbChar)
	}
	if !strings.Contains(result, scrollTrackChar) {
		t.Errorf("ViewWithScrollbar should contain track char %q when content overflows", scrollTrackChar)
	}
}

func TestViewWithScrollbarLineCount(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 5)

	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "content"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	result := vp.ViewWithScrollbar(theme)
	resultLines := strings.Split(result, "\n")
	if len(resultLines) != 5 {
		t.Errorf("ViewWithScrollbar line count = %d, want 5", len(resultLines))
	}
}

func TestViewWithScrollbarWorksWithBothThemes(t *testing.T) {
	themes := []struct {
		name  string
		theme Theme
	}{
		{"dark", DarkTheme()},
		{"light", LightTheme()},
	}

	for _, tt := range themes {
		t.Run(tt.name, func(t *testing.T) {
			vp := NewViewport(40, 5)
			lines := make([]string, 20)
			for i := range lines {
				lines[i] = "content"
			}
			vp.SetContent(strings.Join(lines, "\n"))

			result := vp.ViewWithScrollbar(tt.theme)
			if result == "" {
				t.Errorf("ViewWithScrollbar with %s theme returned empty", tt.name)
			}
			if !strings.Contains(result, scrollThumbChar) {
				t.Errorf("ViewWithScrollbar with %s theme missing thumb char", tt.name)
			}
		})
	}
}

func TestScrollbarMetrics(t *testing.T) {
	tests := []struct {
		name           string
		height         int
		totalLines     int
		wantThumbMin   int
		wantThumbMax   int
		wantStartAtTop int
	}{
		{
			name:           "content fits",
			height:         10,
			totalLines:     5,
			wantThumbMin:   10,
			wantThumbMax:   10,
			wantStartAtTop: 0,
		},
		{
			name:           "double content",
			height:         10,
			totalLines:     20,
			wantThumbMin:   5,
			wantThumbMax:   5,
			wantStartAtTop: 0,
		},
		{
			name:           "large content small viewport",
			height:         5,
			totalLines:     100,
			wantThumbMin:   1,
			wantThumbMax:   1,
			wantStartAtTop: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := NewViewport(40, tt.height)
			lines := make([]string, tt.totalLines)
			for i := range lines {
				lines[i] = "line"
			}
			vp.SetContent(strings.Join(lines, "\n"))

			thumbH, thumbS := vp.scrollbarMetrics()
			if thumbH < tt.wantThumbMin || thumbH > tt.wantThumbMax {
				t.Errorf("thumbHeight = %d, want [%d, %d]", thumbH, tt.wantThumbMin, tt.wantThumbMax)
			}
			if thumbS != tt.wantStartAtTop {
				t.Errorf("thumbStart = %d, want %d at top", thumbS, tt.wantStartAtTop)
			}
		})
	}
}

func TestScrollbarMetricsAfterScroll(t *testing.T) {
	vp := NewViewport(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	_, startAtTop := vp.scrollbarMetrics()
	if startAtTop != 0 {
		t.Errorf("thumbStart at top = %d, want 0", startAtTop)
	}

	vp.ScrollToBottom()
	thumbH, startAtBottom := vp.scrollbarMetrics()
	expectedEnd := vp.VisibleLines() - thumbH
	if startAtBottom != expectedEnd {
		t.Errorf("thumbStart at bottom = %d, want %d", startAtBottom, expectedEnd)
	}
}

func TestViewWithArrowsNoArrowsWhenContentFits(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 10)
	vp.SetContent("short")

	result := vp.ViewWithArrows(theme)
	if strings.Contains(result, scrollUpArrow) {
		t.Error("ViewWithArrows should not show up arrow when content fits")
	}
	if strings.Contains(result, scrollDownArrow) {
		t.Error("ViewWithArrows should not show down arrow when content fits")
	}
}

func TestViewWithArrowsShowsDownAtTop(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 3)
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	result := vp.ViewWithArrows(theme)
	if strings.Contains(result, scrollUpArrow) {
		t.Error("ViewWithArrows should not show up arrow at top")
	}
	if !strings.Contains(result, scrollDownArrow) {
		t.Error("ViewWithArrows should show down arrow when more content below")
	}
}

func TestViewWithArrowsShowsUpAtBottom(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 3)
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollToBottom()

	result := vp.ViewWithArrows(theme)
	if !strings.Contains(result, scrollUpArrow) {
		t.Error("ViewWithArrows should show up arrow when more content above")
	}
	if strings.Contains(result, scrollDownArrow) {
		t.Error("ViewWithArrows should not show down arrow at bottom")
	}
}

func TestViewWithArrowsShowsBothInMiddle(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 3)
	lines := make([]string, 10)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollDown(3)

	result := vp.ViewWithArrows(theme)
	if !strings.Contains(result, scrollUpArrow) {
		t.Error("ViewWithArrows should show up arrow in middle")
	}
	if !strings.Contains(result, scrollDownArrow) {
		t.Error("ViewWithArrows should show down arrow in middle")
	}
}

func TestSetContentPreserveScroll_AtTop(t *testing.T) {
	vp := NewViewport(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	// Viewport is at top initially.
	if !vp.AtTop() {
		t.Fatal("expected viewport at top before SetContentPreserveScroll")
	}

	// Replace content — should stay at top.
	newLines := make([]string, 30)
	for i := range newLines {
		newLines[i] = "new line"
	}
	vp.SetContentPreserveScroll(strings.Join(newLines, "\n"))

	if !vp.AtTop() {
		t.Errorf("expected viewport at top after SetContentPreserveScroll, yOffset = %d", vp.yOffset)
	}
	if vp.yOffset != 0 {
		t.Errorf("yOffset = %d, want 0", vp.yOffset)
	}
}

func TestSetContentPreserveScroll_AtBottom(t *testing.T) {
	vp := NewViewport(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollToBottom()

	if !vp.AtBottom() {
		t.Fatal("expected viewport at bottom before SetContentPreserveScroll")
	}

	// Replace with longer content — should stay at bottom.
	newLines := make([]string, 40)
	for i := range newLines {
		newLines[i] = "new line"
	}
	vp.SetContentPreserveScroll(strings.Join(newLines, "\n"))

	if !vp.AtBottom() {
		t.Errorf("expected viewport at bottom after SetContentPreserveScroll, yOffset = %d, maxOffset = %d",
			vp.yOffset, vp.maxOffset())
	}
}

func TestSetContentPreserveScroll_Proportional(t *testing.T) {
	vp := NewViewport(40, 10)
	// 110 lines, maxOffset = 100, so scrolling to 50 gives 50%.
	lines := make([]string, 110)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollDown(50)

	pctBefore := vp.RawScrollPercent()
	if pctBefore < 0.45 || pctBefore > 0.55 {
		t.Fatalf("expected ~50%% scroll before, got %f", pctBefore)
	}

	// Replace with different length content (210 lines, maxOffset = 200).
	newLines := make([]string, 210)
	for i := range newLines {
		newLines[i] = "new line"
	}
	vp.SetContentPreserveScroll(strings.Join(newLines, "\n"))

	pctAfter := vp.RawScrollPercent()
	if pctAfter < 0.40 || pctAfter > 0.60 {
		t.Errorf("expected roughly 50%% scroll after, got %f (yOffset=%d, maxOffset=%d)",
			pctAfter, vp.yOffset, vp.maxOffset())
	}
}

func TestSetSizePreservesScroll(t *testing.T) {
	vp := NewViewport(40, 10)
	lines := make([]string, 30)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	// Scroll down to offset 10 (maxOffset = 20).
	vp.ScrollDown(10)
	if vp.yOffset != 10 {
		t.Fatalf("yOffset = %d, want 10", vp.yOffset)
	}

	// Shrink the viewport — offset should be clamped but not reset to 0.
	vp.SetSize(40, 15)
	// New maxOffset = 30 - 15 = 15, so offset 10 is still valid.
	if vp.yOffset == 0 {
		t.Error("SetSize should not reset yOffset to 0")
	}
	if vp.yOffset != 10 {
		t.Errorf("yOffset = %d, want 10 (still valid after resize)", vp.yOffset)
	}

	// Shrink further so maxOffset < current offset.
	vp.SetSize(40, 25)
	// New maxOffset = 30 - 25 = 5, so offset 10 should be clamped to 5.
	if vp.yOffset > vp.maxOffset() {
		t.Errorf("yOffset %d exceeds maxOffset %d after aggressive resize", vp.yOffset, vp.maxOffset())
	}
	if vp.yOffset != 5 {
		t.Errorf("yOffset = %d, want 5 (clamped to maxOffset)", vp.yOffset)
	}
}

func TestScrollTrackAndThumbConstants(t *testing.T) {
	if scrollTrackChar != "░" {
		t.Errorf("scrollTrackChar = %q, want %q", scrollTrackChar, "░")
	}
	if scrollThumbChar != "█" {
		t.Errorf("scrollThumbChar = %q, want %q", scrollThumbChar, "█")
	}
}

func TestViewWithScrollbarDownArrowAtTop(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "content"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	// At top: should show ▼ (more below), no ▲ (already at top).
	result := vp.ViewWithScrollbar(theme)
	if !strings.Contains(result, scrollDownArrow) {
		t.Error("ViewWithScrollbar at top should contain down arrow for more content below")
	}
	if strings.Contains(result, scrollUpArrow) {
		t.Error("ViewWithScrollbar at top should not contain up arrow")
	}
}

func TestViewWithScrollbarUpArrowAtBottom(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 5)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "content"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollToBottom()

	// At bottom: should show ▲ (more above), no ▼ (already at bottom).
	result := vp.ViewWithScrollbar(theme)
	if !strings.Contains(result, scrollUpArrow) {
		t.Error("ViewWithScrollbar at bottom should contain up arrow for more content above")
	}
	if strings.Contains(result, scrollDownArrow) {
		t.Error("ViewWithScrollbar at bottom should not contain down arrow")
	}
}

func TestViewWithScrollbarBothArrowsInMiddle(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 10)
	lines := make([]string, 100)
	for i := range lines {
		lines[i] = "content"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollDown(40) // Middle of content.

	result := vp.ViewWithScrollbar(theme)
	if !strings.Contains(result, scrollUpArrow) {
		t.Error("ViewWithScrollbar in middle should contain up arrow")
	}
	if !strings.Contains(result, scrollDownArrow) {
		t.Error("ViewWithScrollbar in middle should contain down arrow")
	}
}

func TestViewWithScrollbarNoArrowsWhenContentFits(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 10)
	vp.SetContent("short\ncontent")

	result := vp.ViewWithScrollbar(theme)
	if strings.Contains(result, scrollUpArrow) {
		t.Error("ViewWithScrollbar should not show up arrow when content fits")
	}
	if strings.Contains(result, scrollDownArrow) {
		t.Error("ViewWithScrollbar should not show down arrow when content fits")
	}
}

func TestViewWithScrollbarSmallViewport(t *testing.T) {
	theme := DarkTheme()
	// Test with minimum reasonable viewport height (5 lines).
	vp := NewViewport(40, 5)
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "content"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollDown(20)

	result := vp.ViewWithScrollbar(theme)
	resultLines := strings.Split(result, "\n")
	if len(resultLines) != 5 {
		t.Errorf("ViewWithScrollbar at height 5 produced %d lines, want 5", len(resultLines))
	}
}

func TestViewWithScrollbarLargeViewport(t *testing.T) {
	theme := DarkTheme()
	// Test with large viewport height (60 lines).
	vp := NewViewport(80, 60)
	lines := make([]string, 200)
	for i := range lines {
		lines[i] = "content line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollDown(70)

	result := vp.ViewWithScrollbar(theme)
	resultLines := strings.Split(result, "\n")
	if len(resultLines) != 60 {
		t.Errorf("ViewWithScrollbar at height 60 produced %d lines, want 60", len(resultLines))
	}
	if !strings.Contains(result, scrollUpArrow) {
		t.Error("ViewWithScrollbar large viewport in middle should contain up arrow")
	}
	if !strings.Contains(result, scrollDownArrow) {
		t.Error("ViewWithScrollbar large viewport in middle should contain down arrow")
	}
}

// --- US-087: Viewport interaction tests ---

func TestPageUpDown(t *testing.T) {
	vp := NewViewport(40, 10)
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	// Page down scrolls by viewport height (10).
	vp.ScrollDown(vp.VisibleLines())
	if vp.yOffset != 10 {
		t.Errorf("after page down, yOffset = %d, want 10", vp.yOffset)
	}

	// Page up scrolls back.
	vp.ScrollUp(vp.VisibleLines())
	if vp.yOffset != 0 {
		t.Errorf("after page up, yOffset = %d, want 0", vp.yOffset)
	}

	// Multiple page downs.
	vp.ScrollDown(vp.VisibleLines())
	vp.ScrollDown(vp.VisibleLines())
	if vp.yOffset != 20 {
		t.Errorf("after 2 page downs, yOffset = %d, want 20", vp.yOffset)
	}
}

func TestHalfPageScroll(t *testing.T) {
	vp := NewViewport(40, 10)
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	// Half-page down (Ctrl+d equivalent) scrolls by height/2 = 5.
	vp.ScrollDown(vp.VisibleLines() / 2)
	if vp.yOffset != 5 {
		t.Errorf("after half page down, yOffset = %d, want 5", vp.yOffset)
	}

	// Half-page up (Ctrl+u equivalent).
	vp.ScrollUp(vp.VisibleLines() / 2)
	if vp.yOffset != 0 {
		t.Errorf("after half page up, yOffset = %d, want 0", vp.yOffset)
	}
}

func TestMouseWheelScroll(t *testing.T) {
	vp := NewViewport(40, 10)
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	// Mouse wheel down scrolls by 3 lines.
	vp.ScrollDown(3)
	if vp.yOffset != 3 {
		t.Errorf("after mouse wheel down, yOffset = %d, want 3", vp.yOffset)
	}

	// Mouse wheel up scrolls by 3 lines.
	vp.ScrollUp(3)
	if vp.yOffset != 0 {
		t.Errorf("after mouse wheel up, yOffset = %d, want 0", vp.yOffset)
	}
}

func TestResizePreservesScrollAtTop(t *testing.T) {
	vp := NewViewport(40, 10)
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	// At top, resize should keep at top.
	if !vp.AtTop() {
		t.Fatal("expected at top")
	}
	vp.SetSize(60, 20)
	if !vp.AtTop() {
		t.Errorf("expected at top after resize, yOffset = %d", vp.yOffset)
	}
}

func TestResizePreservesScrollAtBottom(t *testing.T) {
	vp := NewViewport(40, 10)
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollToBottom()

	if !vp.AtBottom() {
		t.Fatal("expected at bottom")
	}

	// Resize — offset should be clamped to new max.
	vp.SetSize(40, 20)
	max := vp.maxOffset()
	if vp.yOffset > max {
		t.Errorf("yOffset %d exceeds maxOffset %d after resize", vp.yOffset, max)
	}
}

func TestResizeClampsOffsetInMiddle(t *testing.T) {
	vp := NewViewport(40, 10)
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollDown(20)

	// Grow viewport so maxOffset shrinks. Offset should be clamped.
	vp.SetSize(40, 45)
	max := vp.maxOffset() // 50 - 45 = 5
	if vp.yOffset > max {
		t.Errorf("yOffset %d should be clamped to maxOffset %d", vp.yOffset, max)
	}
	if vp.yOffset != max {
		t.Errorf("yOffset = %d, want %d (clamped)", vp.yOffset, max)
	}
}

func TestScrollWhenContentShorterThanViewport(t *testing.T) {
	vp := NewViewport(40, 20)
	vp.SetContent("just\nthree\nlines")

	// Scrolling should be a no-op when content fits.
	vp.ScrollDown(10)
	if vp.yOffset != 0 {
		t.Errorf("yOffset should be 0 when content fits, got %d", vp.yOffset)
	}
	vp.ScrollUp(10)
	if vp.yOffset != 0 {
		t.Errorf("yOffset should be 0 after ScrollUp when content fits, got %d", vp.yOffset)
	}
}

func TestScrollAtBoundariesNoPanic(t *testing.T) {
	vp := NewViewport(40, 10)
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	// Scroll far past bottom.
	vp.ScrollDown(1000)
	if vp.yOffset != vp.maxOffset() {
		t.Errorf("yOffset %d should equal maxOffset %d", vp.yOffset, vp.maxOffset())
	}

	// Scroll far past top.
	vp.ScrollUp(1000)
	if vp.yOffset != 0 {
		t.Errorf("yOffset should be 0, got %d", vp.yOffset)
	}
}

func TestSetContentPreserveScroll_ContentShrinksBelowPosition(t *testing.T) {
	vp := NewViewport(40, 10)
	lines := make([]string, 50)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))
	vp.ScrollDown(30) // maxOffset = 40, so 30 is valid

	// Replace with shorter content.
	shortLines := make([]string, 12)
	for i := range shortLines {
		shortLines[i] = "new"
	}
	vp.SetContentPreserveScroll(strings.Join(shortLines, "\n"))

	// New maxOffset = 12 - 10 = 2. Offset should be clamped.
	max := vp.maxOffset()
	if vp.yOffset > max {
		t.Errorf("yOffset %d exceeds maxOffset %d after content shrink", vp.yOffset, max)
	}
}

func TestPageScrollClampsAtBoundaries(t *testing.T) {
	vp := NewViewport(40, 10)
	lines := make([]string, 15)
	for i := range lines {
		lines[i] = "line"
	}
	vp.SetContent(strings.Join(lines, "\n"))

	// Page down twice: first goes to 10, second should clamp at maxOffset (5).
	vp.ScrollDown(vp.VisibleLines()) // 10
	vp.ScrollDown(vp.VisibleLines()) // would be 20, clamped to 5
	if vp.yOffset != vp.maxOffset() {
		t.Errorf("yOffset %d should be clamped at maxOffset %d", vp.yOffset, vp.maxOffset())
	}

	// Page up twice from bottom: should reach 0.
	vp.ScrollUp(vp.VisibleLines())
	vp.ScrollUp(vp.VisibleLines())
	if vp.yOffset != 0 {
		t.Errorf("yOffset should be 0 after page up from near top, got %d", vp.yOffset)
	}
}

func TestEmptyContentScrollSafe(t *testing.T) {
	vp := NewViewport(40, 10)
	vp.SetContent("")

	// All scroll operations should be safe on empty content.
	vp.ScrollDown(10)
	vp.ScrollUp(10)
	vp.ScrollToBottom()
	vp.ScrollToTop()

	if vp.yOffset != 0 {
		t.Errorf("yOffset should be 0 with empty content, got %d", vp.yOffset)
	}

	view := vp.View()
	if view != "" {
		t.Errorf("View with empty content should be empty, got %q", view)
	}
}

func TestViewportZeroHeight(t *testing.T) {
	vp := NewViewport(40, 0)
	vp.SetContent("some\ncontent\nhere")

	// All operations should be safe.
	vp.ScrollDown(5)
	vp.ScrollUp(5)
	view := vp.View()
	if view != "" {
		t.Errorf("View with height 0 should be empty, got %q", view)
	}
}

func TestContentWidth(t *testing.T) {
	tests := []struct {
		name      string
		vpWidth   int
		wantWidth int
	}{
		{"narrow", 40, 39},
		{"standard", 80, 79},
		{"wide_capped", 120, MaxContentWidth},
		{"exact_cap", MaxContentWidth + 1, MaxContentWidth},
		{"below_cap", MaxContentWidth, MaxContentWidth - 1},
		{"zero", 0, 0},
		{"one", 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vp := NewViewport(tt.vpWidth, 10)
			got := vp.ContentWidth()
			if got != tt.wantWidth {
				t.Errorf("ContentWidth() = %d, want %d (vpWidth=%d)", got, tt.wantWidth, tt.vpWidth)
			}
		})
	}
}

func TestVerticalCentering(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 10)
	vp.SetContent("hello")

	result := vp.ViewWithScrollbar(theme)
	lines := strings.Split(result, "\n")
	if len(lines) != 10 {
		t.Fatalf("expected 10 lines, got %d", len(lines))
	}

	// Content (1 line) in 10-line viewport: topPad = (10-1)/2 = 4.
	// Lines 0-3 should be blank, line 4 should contain "hello", lines 5-9 blank.
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 4 {
			if trimmed != "hello" {
				t.Errorf("line %d = %q, want 'hello'", i, trimmed)
			}
		} else {
			if trimmed != "" {
				t.Errorf("line %d should be blank, got %q", i, trimmed)
			}
		}
	}
}

func TestHorizontalCentering(t *testing.T) {
	theme := DarkTheme()
	vp := NewViewport(40, 5)
	vp.SetContent("hi")

	result := vp.ViewWithScrollbar(theme)
	lines := strings.Split(result, "\n")

	// Find the content line (vertically centered).
	contentLine := lines[(5-1)/2]
	// "hi" should be centered in 40 cols, so it should have leading spaces.
	if !strings.Contains(contentLine, "hi") {
		t.Error("centered line should contain 'hi'")
	}
	leading := len(contentLine) - len(strings.TrimLeft(contentLine, " "))
	if leading == 0 {
		t.Error("content should have leading spaces when centered in wide viewport")
	}
}
