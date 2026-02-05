package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// defaultTickDuration is the base interval between typewriter ticks.
const defaultTickDuration = 50 * time.Millisecond

// TypewriterDoneMsg is sent when a typewriter finishes revealing all text.
type TypewriterDoneMsg struct {
	ID string
}

// typewriterTickMsg is an internal tick message scoped to a specific typewriter instance.
type typewriterTickMsg struct {
	id string
}

// Typewriter reveals text character-by-character using Bubbletea's tick system.
// Multiple instances can coexist by using distinct IDs.
type Typewriter struct {
	id          string
	text        []rune
	pos         int
	charsPerTick int
	done        bool
}

// NewTypewriter creates a Typewriter that reveals the given text at the specified
// speed (characters per tick). The id distinguishes this instance's tick messages
// from those of other Typewriter instances.
func NewTypewriter(id, text string, charsPerTick int) Typewriter {
	if charsPerTick < 1 {
		charsPerTick = 1
	}
	runes := []rune(text)
	return Typewriter{
		id:           id,
		text:         runes,
		pos:          0,
		charsPerTick: charsPerTick,
		done:         len(runes) == 0,
	}
}

// Update handles typewriterTickMsg to advance the revealed text position.
// It returns the updated Typewriter and any follow-up command (next tick or done message).
func (tw Typewriter) Update(msg tea.Msg) (Typewriter, tea.Cmd) {
	if tw.done {
		return tw, nil
	}

	tick, ok := msg.(typewriterTickMsg)
	if !ok || tick.id != tw.id {
		return tw, nil
	}

	tw.pos += tw.charsPerTick
	if tw.pos >= len(tw.text) {
		tw.pos = len(tw.text)
		tw.done = true
		return tw, func() tea.Msg {
			return TypewriterDoneMsg{ID: tw.id}
		}
	}

	return tw, tw.Tick()
}

// View returns the currently revealed portion of the text.
func (tw Typewriter) View() string {
	return string(tw.text[:tw.pos])
}

// Tick returns a tea.Cmd that schedules the next typewriter tick.
func (tw Typewriter) Tick() tea.Cmd {
	id := tw.id
	return tea.Tick(defaultTickDuration, func(_ time.Time) tea.Msg {
		return typewriterTickMsg{id: id}
	})
}

// Done reports whether the typewriter has finished revealing all text.
func (tw Typewriter) Done() bool {
	return tw.done
}

// Skip immediately reveals all remaining text, marking the typewriter as done.
func (tw *Typewriter) Skip() {
	tw.pos = len(tw.text)
	tw.done = true
}
