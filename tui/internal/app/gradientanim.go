package app

import (
	"math"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// gradientAnimTickInterval is the frame rate for gradient animation.
const gradientAnimTickInterval = 50 * time.Millisecond // ~20fps

// gradientAnimTickMsg advances the gradient animation by one frame.
type gradientAnimTickMsg struct {
	id string
}

// GradientAnim animates a color gradient sweep across text, interpolating
// in Lab color space with a time-varying offset driven by incommensurate
// sinusoids for organic movement.
type GradientAnim struct {
	id     string
	active bool
	frame  int

	startC colorful.Color // cached from theme accent
	endC   colorful.Color // cached from theme fg
}

// NewGradientAnim creates a GradientAnim with colors cached from the theme.
func NewGradientAnim(id string, theme Theme) GradientAnim {
	startC, _ := HexToColorful(theme.Colors.Accent)
	endC, _ := HexToColorful(theme.Colors.Fg)
	return GradientAnim{
		id:     id,
		startC: startC,
		endC:   endC,
	}
}

// Start begins the gradient animation and returns the first tick command.
func (g *GradientAnim) Start() tea.Cmd {
	g.active = true
	g.frame = 0
	return g.tick()
}

// Stop halts the gradient animation.
func (g *GradientAnim) Stop() {
	g.active = false
}

// Active returns whether the gradient animation is currently running.
func (g GradientAnim) Active() bool {
	return g.active
}

// Update advances the animation by one frame on a matching tick message.
func (g GradientAnim) Update(msg tea.Msg) (GradientAnim, tea.Cmd) {
	if tick, ok := msg.(gradientAnimTickMsg); ok && tick.id == g.id && g.active {
		g.frame++
		return g, g.tick()
	}
	return g, nil
}

// SetTheme re-caches colors from the theme.
func (g *GradientAnim) SetTheme(theme Theme) {
	startC, err1 := HexToColorful(theme.Colors.Accent)
	endC, err2 := HexToColorful(theme.Colors.Fg)
	if err1 == nil {
		g.startC = startC
	}
	if err2 == nil {
		g.endC = endC
	}
}

// Render applies the animated gradient to text, returning per-character
// styled output with bold.
func (g GradientAnim) Render(text string) string {
	runes := []rune(text)
	if len(runes) == 0 {
		return text
	}

	// Animated offset: 3 incommensurate sinusoids for organic movement.
	offset := 0.3*math.Sin(float64(g.frame)*0.012) +
		0.2*math.Sin(float64(g.frame)*0.007) +
		0.1*math.Sin(float64(g.frame)*0.019)

	var b strings.Builder
	b.Grow(len(text) * 20)

	last := len(runes) - 1
	if last == 0 {
		last = 1
	}

	for i, r := range runes {
		t := float64(i)/float64(last) + offset
		// Clamp to [0, 1].
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		blended := g.startC.BlendLab(g.endC, t)
		style := lipgloss.NewStyle().Foreground(lipgloss.Color(blended.Hex())).Bold(true)
		b.WriteString(style.Render(string(r)))
	}

	return b.String()
}

// tick returns a tea.Cmd that fires a gradientAnimTickMsg after one interval.
func (g GradientAnim) tick() tea.Cmd {
	id := g.id
	return tea.Tick(gradientAnimTickInterval, func(_ time.Time) tea.Msg {
		return gradientAnimTickMsg{id: id}
	})
}
