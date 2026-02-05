package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	// transitionID identifies transition animation ticks.
	transitionID = "section-transition"

	// baseTransitionSteps is the step count for adjacent section transitions.
	// Each step takes ~16ms (one animation frame), so 6 steps ≈ 96ms.
	baseTransitionSteps = 6

	// extraStepsPerDistance adds steps for each additional section distance,
	// making distant transitions slightly longer than adjacent ones.
	extraStepsPerDistance = 2
)

// TransitionDirection indicates the visual direction of the transition.
type TransitionDirection int

const (
	// TransitionRight slides content to the right (navigating forward).
	TransitionRight TransitionDirection = 1
	// TransitionLeft slides content to the left (navigating backward).
	TransitionLeft TransitionDirection = -1
)

// TransitionDoneMsg signals that the transition animation has completed.
type TransitionDoneMsg struct{}

// TransitionManager handles animated transitions between sections.
type TransitionManager struct {
	active    bool
	from      Section
	to        Section
	direction TransitionDirection
	step      int
	steps     int
}

// NewTransitionManager creates a TransitionManager with default settings.
func NewTransitionManager() TransitionManager {
	return TransitionManager{}
}

// Start begins a transition from one section to another.
// The step count varies by section distance: adjacent sections use fewer
// steps (~96ms) while distant sections use more (~160ms).
// Returns a tea.Cmd to start the animation tick loop.
func (t *TransitionManager) Start(from, to Section) tea.Cmd {
	t.active = true
	t.from = from
	t.to = to
	t.step = 0

	distance := int(to) - int(from)
	if distance < 0 {
		distance = -distance
	}
	t.steps = baseTransitionSteps + (distance-1)*extraStepsPerDistance

	if to > from {
		t.direction = TransitionRight
	} else {
		t.direction = TransitionLeft
	}

	return animationTick(transitionID)
}

// Active returns whether a transition is currently running.
func (t *TransitionManager) Active() bool {
	return t.active
}

// Update handles AnimationTickMsg to advance the transition.
func (t *TransitionManager) Update(msg tea.Msg) tea.Cmd {
	if !t.active {
		return nil
	}

	tick, ok := msg.(AnimationTickMsg)
	if !ok || tick.ID != transitionID {
		return nil
	}

	t.step++
	if t.step >= t.steps {
		t.active = false
		return func() tea.Msg { return TransitionDoneMsg{} }
	}

	return animationTick(transitionID)
}

// View renders the mid-transition view by blending fromView and toView.
// Uses a simple horizontal offset effect. Falls back to toView for very
// small terminals (width < 20).
func (t *TransitionManager) View(fromView, toView string, width int) string {
	if width < 20 || t.steps <= 0 {
		return toView
	}

	progress := float64(t.step) / float64(t.steps)
	eased := easeInOut(progress)

	// Calculate horizontal offset for the "old" view sliding out.
	offset := int(eased * float64(width) / 3)

	fromLines := strings.Split(fromView, "\n")
	toLines := strings.Split(toView, "\n")

	// Use the longer of the two line counts.
	maxLines := len(fromLines)
	if len(toLines) > maxLines {
		maxLines = len(toLines)
	}

	var b strings.Builder
	for i := range maxLines {
		if i > 0 {
			b.WriteByte('\n')
		}

		if eased < 0.5 {
			// First half: old view sliding out.
			line := ""
			if i < len(fromLines) {
				line = fromLines[i]
			}
			shifted := shiftLine(line, offset, int(t.direction), width)
			b.WriteString(shifted)
		} else {
			// Second half: new view sliding in.
			line := ""
			if i < len(toLines) {
				line = toLines[i]
			}
			reverseOffset := int((1.0 - eased) * float64(width) / 3)
			shifted := shiftLine(line, reverseOffset, -int(t.direction), width)
			b.WriteString(shifted)
		}
	}

	return b.String()
}

// shiftLine shifts a line by offset visual columns in the given direction.
// Positive direction shifts right, negative shifts left. Output is
// clamped to width columns. Uses lipgloss for ANSI-safe text manipulation
// rather than operating on raw runes, which would break escape sequences.
func shiftLine(line string, offset, direction, width int) string {
	if offset <= 0 || width <= 0 {
		return line
	}

	clamp := lipgloss.NewStyle().Width(width).MaxWidth(width)

	if direction > 0 {
		// Shift right: prepend spaces to push content rightward,
		// then clamp to width (lipgloss handles ANSI truncation).
		return clamp.Render(strings.Repeat(" ", offset) + line)
	}

	// Shift left: content slides off the left edge.
	// Truncate from the right to (width - offset) visible columns,
	// then right-pad with spaces to fill the width. This produces a
	// smooth shrink-away effect during the ~80 ms half-transition.
	remaining := width - offset
	if remaining <= 0 {
		return strings.Repeat(" ", width)
	}

	truncated := lipgloss.NewStyle().Width(remaining).MaxWidth(remaining).Render(line)
	return truncated + strings.Repeat(" ", offset)
}
