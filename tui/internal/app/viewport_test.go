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
	plain := vp.View()

	if withScrollbar != plain {
		t.Error("ViewWithScrollbar should return View() when content fits in viewport")
	}

	if strings.Contains(withScrollbar, scrollTrackChar) {
		t.Errorf("ViewWithScrollbar should not contain track char %q when content fits", scrollTrackChar)
	}
	if strings.Contains(withScrollbar, scrollThumbChar) {
		t.Errorf("ViewWithScrollbar should not contain thumb char %q when content fits", scrollThumbChar)
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

func TestScrollTrackAndThumbConstants(t *testing.T) {
	if scrollTrackChar != "░" {
		t.Errorf("scrollTrackChar = %q, want %q", scrollTrackChar, "░")
	}
	if scrollThumbChar != "█" {
		t.Errorf("scrollThumbChar = %q, want %q", scrollThumbChar, "█")
	}
}
