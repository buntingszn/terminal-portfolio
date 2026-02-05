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

// SetSize updates the viewport dimensions and clamps the scroll offset.
func (v *Viewport) SetSize(width, height int) {
	v.width = width
	v.height = height
	v.clampOffset()
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
// muted color and a thumb character (█) in the accent color. The thumb height
// is proportional to the visible/total content ratio (minimum 1 character).
// If all content fits in the viewport, the scrollbar is hidden and plain
// View() output is returned.
func (v *Viewport) ViewWithScrollbar(theme Theme) string {
	totalLines := v.TotalLines()
	visibleHeight := v.height

	if totalLines <= visibleHeight {
		return v.View()
	}

	thumbHeight, thumbStart := v.scrollbarMetrics()

	trackStyle := lipgloss.NewStyle().Foreground(theme.Colors.Muted)
	thumbStyle := lipgloss.NewStyle().Foreground(theme.Colors.Accent)

	indicator := make([]string, visibleHeight)
	for i := range visibleHeight {
		if i >= thumbStart && i < thumbStart+thumbHeight {
			indicator[i] = thumbStyle.Render(scrollThumbChar)
		} else {
			indicator[i] = trackStyle.Render(scrollTrackChar)
		}
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
		rendered := lipgloss.NewStyle().Width(contentWidth).MaxWidth(contentWidth).Render(line)
		b.WriteString(rendered)
		b.WriteString(indicator[i])
		if i < visibleHeight-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
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
