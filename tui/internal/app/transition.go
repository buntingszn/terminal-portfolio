package app

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	// transitionID identifies transition animation ticks.
	transitionID = "section-transition"

	// transitionDuration is the total duration for a section transition.
	// Kept under 200ms for snappy feel.
	transitionSteps = 10 // ~160ms at 16ms/frame
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
	return TransitionManager{
		steps: transitionSteps,
	}
}

// Start begins a transition from one section to another.
// Returns a tea.Cmd to start the animation tick loop.
func (t *TransitionManager) Start(from, to Section) tea.Cmd {
	t.active = true
	t.from = from
	t.to = to
	t.step = 0

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

// shiftLine shifts a line by offset characters in the given direction.
// Positive direction shifts right, negative shifts left. Output is
// clamped to width characters.
func shiftLine(line string, offset, direction, width int) string {
	if offset <= 0 || width <= 0 {
		return line
	}

	runes := []rune(line)
	result := make([]rune, width)

	// Fill with spaces.
	for i := range result {
		result[i] = ' '
	}

	if direction > 0 {
		// Shift right: content moves right.
		for i, r := range runes {
			dest := i + offset
			if dest >= 0 && dest < width {
				result[dest] = r
			}
		}
	} else {
		// Shift left: content moves left.
		for i, r := range runes {
			dest := i - offset
			if dest >= 0 && dest < width {
				result[dest] = r
			}
		}
	}

	return string(result)
}
