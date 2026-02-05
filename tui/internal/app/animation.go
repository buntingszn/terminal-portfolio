package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// animationTickInterval is the frame rate for transition animations.
const animationTickInterval = 16 * time.Millisecond // ~60fps

// AnimationTickMsg advances a running animation by one frame.
type AnimationTickMsg struct {
	ID string
}

// AnimationFrame holds the current state of an animation.
type AnimationFrame struct {
	Progress float64 // 0.0 to 1.0
	Done     bool
}

// animationTick returns a tea.Cmd that fires an AnimationTickMsg after one frame interval.
func animationTick(id string) tea.Cmd {
	return tea.Tick(animationTickInterval, func(_ time.Time) tea.Msg {
		return AnimationTickMsg{ID: id}
	})
}

// easeInOut applies a smooth ease-in-out curve (cubic).
func easeInOut(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - (-2*t+2)*(-2*t+2)*(-2*t+2)/2
}
