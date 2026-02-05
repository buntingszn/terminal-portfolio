package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	// scrollTrackChar is rendered for the non-thumb portion of the scroll indicator.
	scrollTrackChar = "░"
	// scrollThumbChar is rendered for the thumb (current position) of the scroll indicator.
	scrollThumbChar = "█"
	// scrollUpArrow indicates more content above.
	scrollUpArrow = "▲"
	// scrollDownArrow indicates more content below.
	scrollDownArrow = "▼"

	// MaxContentWidth is the maximum width for section content. On wide
	// terminals the content is capped at this width and centered horizontally.
	// 88 = comfortable reading width (80 content + card borders), leaving
	// ~16 cols margin per side on a 120-col terminal.
	MaxContentWidth = 88
)

// Viewport is a scrollable content viewer. It slices pre-rendered text into a
// visible window and provides scroll position indicators. It is a pure
// rendering utility — it does not implement tea.Model and has no bubbletea
// dependency.
type Viewport struct {
	content string
	lines   []string
	width   int
	height  int
	yOffset int
}

// NewViewport creates a Viewport with the given dimensions.
func NewViewport(width, height int) Viewport {
	return Viewport{
		width:  width,
		height: height,
	}
}

// SetContent loads rendered text into the viewport and resets the scroll
// position to the top.
func (v *Viewport) SetContent(content string) {
	v.content = content
	v.lines = strings.Split(content, "\n")
	v.yOffset = 0
}

// SetContentPreserveScroll updates content without resetting the scroll
// position. If the user was at the bottom, they stay at the bottom. If at the
// top, they stay at the top. Otherwise the proportional scroll position is
// restored. This is intended for resize-triggered re-renders where the user's
// reading position should be preserved.
func (v *Viewport) SetContentPreserveScroll(content string) {
	wasAtBottom := v.AtBottom()
	wasAtTop := v.AtTop()
	oldPercent := v.RawScrollPercent()

	v.content = content
	v.lines = strings.Split(content, "\n")

	if wasAtTop {
		v.yOffset = 0
	} else if wasAtBottom {
		v.yOffset = v.maxOffset()
	} else {
		// Restore proportional position.
		v.yOffset = int(oldPercent * float64(v.maxOffset()))
	}
	v.clampOffset()
}

// SetSize updates the viewport dimensions and clamps the scroll offset.
func (v *Viewport) SetSize(width, height int) {
	v.width = width
	v.height = height
	v.clampOffset()
}

// ContentWidth returns the usable width for section content, accounting for the
// scrollbar column and capping at MaxContentWidth for readability on wide terminals.
func (v *Viewport) ContentWidth() int {
	w := v.width - 1 // scrollbar
	if w > MaxContentWidth {
		w = MaxContentWidth
	}
	if w < 0 {
		w = 0
	}
	return w
}

// ScrollUp scrolls up by n lines.
func (v *Viewport) ScrollUp(n int) {
	v.yOffset -= n
	v.clampOffset()
}

// ScrollDown scrolls down by n lines.
func (v *Viewport) ScrollDown(n int) {
	v.yOffset += n
	v.clampOffset()
}

// ScrollToTop scrolls to the very top.
func (v *Viewport) ScrollToTop() {
	v.yOffset = 0
}

// ScrollToBottom scrolls to the very bottom.
func (v *Viewport) ScrollToBottom() {
	v.yOffset = v.maxOffset()
}

// AtTop returns true when the viewport is scrolled to the top.
func (v *Viewport) AtTop() bool {
	return v.yOffset <= 0
}

// AtBottom returns true when the viewport is scrolled to the bottom or when
// all content fits without scrolling.
func (v *Viewport) AtBottom() bool {
	return v.yOffset >= v.maxOffset()
}

// TotalLines returns the total number of lines in the content.
func (v *Viewport) TotalLines() int {
	return len(v.lines)
}

// VisibleLines returns the viewport height.
func (v *Viewport) VisibleLines() int {
	return v.height
}

// ScrollPercent returns the scroll position as a formatted percentage string.
func (v *Viewport) ScrollPercent() string {
	if v.maxOffset() <= 0 {
		return "100%"
	}
	pct := float64(v.yOffset) / float64(v.maxOffset()) * 100
	return fmt.Sprintf("%3.f%%", pct)
}

// RawScrollPercent returns the scroll position as a float between 0.0 and 1.0.
func (v *Viewport) RawScrollPercent() float64 {
	if v.maxOffset() <= 0 {
		return 1.0
	}
	return float64(v.yOffset) / float64(v.maxOffset())
}

// View renders the visible portion of the content. Lines are not styled or
// padded; they are returned exactly as they appear in the content.
func (v *Viewport) View() string {
	if v.height <= 0 {
		return ""
	}
	visible := v.visibleSlice()
	return strings.Join(visible, "\n")
}

// ViewWithScrollbar renders the viewport content with a vertical scrollbar on
// the right edge. The scrollbar uses a track character (░) in the theme's
// border color and a thumb character (█) in the muted color. The thumb height
// is proportional to the visible/total content ratio (minimum 1 character).
// When more content exists above or below the visible area, ▲/▼ arrows in the
// accent color replace the first/last track character. If all content fits in
// the viewport, the scrollbar is hidden and plain View() output is returned.
func (v *Viewport) ViewWithScrollbar(theme Theme) string {
	totalLines := v.TotalLines()
	visibleHeight := v.height

	if totalLines <= visibleHeight {
		return v.viewCentered()
	}

	thumbHeight, thumbStart := v.scrollbarMetrics()

	trackStyle := lipgloss.NewStyle().Foreground(theme.Colors.Border)
	thumbStyle := lipgloss.NewStyle().Foreground(theme.Colors.Muted)
	arrowStyle := lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	indicator := make([]string, visibleHeight)
	for i := range visibleHeight {
		if i >= thumbStart && i < thumbStart+thumbHeight {
			indicator[i] = thumbStyle.Render(scrollThumbChar)
		} else {
			indicator[i] = trackStyle.Render(scrollTrackChar)
		}
	}

	// Replace first/last track character with directional arrows when there
	// is more content above or below, respectively. Arrows only replace
	// track characters, never the thumb.
	if !v.AtTop() && (0 < thumbStart || 0 >= thumbStart+thumbHeight) {
		indicator[0] = arrowStyle.Render(scrollUpArrow)
	}
	if !v.AtBottom() && (visibleHeight-1 < thumbStart || visibleHeight-1 >= thumbStart+thumbHeight) {
		indicator[visibleHeight-1] = arrowStyle.Render(scrollDownArrow)
	}

	visible := v.visibleSlice()

	// Pad or trim to match viewport height.
	for len(visible) < visibleHeight {
		visible = append(visible, "")
	}
	if len(visible) > visibleHeight {
		visible = visible[:visibleHeight]
	}

	contentWidth := v.width - 1
	if contentWidth < 0 {
		contentWidth = 0
	}

	var b strings.Builder
	for i := range visibleHeight {
		line := visible[i]
		// Center content horizontally within the available width.
		centered := lipgloss.PlaceHorizontal(contentWidth, lipgloss.Center, line)
		rendered := lipgloss.NewStyle().Width(contentWidth).MaxWidth(contentWidth).Render(centered)
		b.WriteString(rendered)
		b.WriteString(indicator[i])
		if i < visibleHeight-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}

// viewCentered renders content centered both vertically and horizontally
// when all content fits within the viewport (no scrollbar needed).
func (v *Viewport) viewCentered() string {
	if v.height <= 0 {
		return ""
	}

	visible := v.visibleSlice()
	totalLines := len(visible)
	fullWidth := v.width

	// Vertical padding: center content within viewport height.
	topPad := (v.height - totalLines) / 2
	if topPad < 0 {
		topPad = 0
	}

	output := make([]string, v.height)

	for i := range v.height {
		contentIdx := i - topPad
		var line string
		if contentIdx >= 0 && contentIdx < totalLines {
			line = visible[contentIdx]
		}
		// Center each line horizontally across the full width.
		output[i] = lipgloss.PlaceHorizontal(fullWidth, lipgloss.Center, line)
	}

	return strings.Join(output, "\n")
}

// ViewWithArrows renders the viewport with ▲/▼ arrow indicators in the theme
// accent color when there is more content above or below the visible area.
// The arrows are centered at the top/bottom of the viewport.
func (v *Viewport) ViewWithArrows(theme Theme) string {
	visible := v.visibleSlice()
	if v.height <= 0 {
		return ""
	}

	var b strings.Builder

	arrowStyle := lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	if !v.AtTop() {
		arrow := arrowStyle.Render(scrollUpArrow)
		b.WriteString(lipgloss.PlaceHorizontal(v.width, lipgloss.Center, arrow))
		b.WriteByte('\n')
	}

	b.WriteString(strings.Join(visible, "\n"))

	if !v.AtBottom() {
		b.WriteByte('\n')
		arrow := arrowStyle.Render(scrollDownArrow)
		b.WriteString(lipgloss.PlaceHorizontal(v.width, lipgloss.Center, arrow))
	}

	return b.String()
}

// scrollbarMetrics returns the thumb height and start position for the
// scrollbar indicator.
func (v *Viewport) scrollbarMetrics() (thumbHeight, thumbStart int) {
	totalLines := v.TotalLines()
	visibleHeight := v.height

	if totalLines <= visibleHeight || visibleHeight <= 0 {
		return visibleHeight, 0
	}

	thumbHeight = visibleHeight * visibleHeight / totalLines
	if thumbHeight < 1 {
		thumbHeight = 1
	}

	maxOff := v.maxOffset()
	yOff := v.yOffset
	if yOff > maxOff {
		yOff = maxOff
	}
	if yOff < 0 {
		yOff = 0
	}

	trackSpace := visibleHeight - thumbHeight
	if maxOff > 0 && trackSpace > 0 {
		thumbStart = yOff * trackSpace / maxOff
	}

	return thumbHeight, thumbStart
}

// maxOffset returns the maximum valid yOffset value.
func (v *Viewport) maxOffset() int {
	max := len(v.lines) - v.height
	if max < 0 {
		return 0
	}
	return max
}

// clampOffset ensures yOffset stays within [0, maxOffset].
func (v *Viewport) clampOffset() {
	if v.yOffset < 0 {
		v.yOffset = 0
	}
	if m := v.maxOffset(); v.yOffset > m {
		v.yOffset = m
	}
}

// GetScrollInfo returns the current scroll state suitable for display in the
// status bar. If all content fits within the viewport, Fits is true and no
// scroll indicator is needed.
func (v *Viewport) GetScrollInfo() ScrollInfo {
	if v.TotalLines() <= v.height {
		return ScrollInfo{Fits: true, AtTop: true, AtBottom: true}
	}
	return ScrollInfo{
		AtTop:   v.AtTop(),
		AtBottom: v.AtBottom(),
		Percent:  v.ScrollPercent(),
	}
}

// visibleSlice returns the slice of lines currently visible.
func (v *Viewport) visibleSlice() []string {
	if len(v.lines) == 0 {
		return nil
	}
	start := v.yOffset
	end := start + v.height
	if end > len(v.lines) {
		end = len(v.lines)
	}
	if start >= end {
		return nil
	}
	return v.lines[start:end]
}
