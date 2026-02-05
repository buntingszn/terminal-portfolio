package app

import (
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// tabGlowTickInterval is the frame rate for the glow pulse animation.
const tabGlowTickInterval = 16 * time.Millisecond // ~60fps

// tabGlowDuration is how long the full pulse lasts.
const tabGlowDuration = 600 * time.Millisecond

// tabGlowTickMsg advances the tab glow animation by one frame.
type tabGlowTickMsg struct{}

// TabGlow animates a brief brightness pulse on a color. It ramps the
// accent color's lightness up then back down over tabGlowDuration.
type TabGlow struct {
	active bool
	step   int
	steps  int // total steps = duration / tickInterval
	theme  Theme
}

// NewTabGlow creates a TabGlow ready to be triggered.
func NewTabGlow(theme Theme) TabGlow {
	return TabGlow{
		steps: int(tabGlowDuration / tabGlowTickInterval),
		theme: theme,
	}
}

// Start begins the glow pulse and returns the first tick command.
func (g *TabGlow) Start() tea.Cmd {
	g.active = true
	g.step = 0
	return g.tick()
}

// Active returns whether the glow animation is running.
func (g TabGlow) Active() bool {
	return g.active
}

// Update advances the glow by one step on a tabGlowTickMsg.
func (g TabGlow) Update(msg tea.Msg) (TabGlow, tea.Cmd) {
	if _, ok := msg.(tabGlowTickMsg); !ok || !g.active {
		return g, nil
	}

	g.step++
	if g.step >= g.steps {
		g.active = false
		return g, nil
	}
	return g, g.tick()
}

// BrightenedAccent returns the theme accent color with lightness boosted
// according to the current pulse progress.
func (g TabGlow) BrightenedAccent() lipgloss.Color {
	if !g.active || g.steps <= 0 {
		return g.theme.Colors.Accent
	}

	progress := float64(g.step) / float64(g.steps)
	// Ease in-out: 0 → 1 → 0, peaking at progress=0.5.
	pulse := math.Sin(progress * math.Pi)
	// Lighten accent by up to 40%.
	maxBoost := 0.4

	c, err := HexToColorful(g.theme.Colors.Accent)
	if err != nil {
		return g.theme.Colors.Accent
	}

	h, s, l := c.Hsl()
	l += maxBoost * pulse
	if l > 1.0 {
		l = 1.0
	}

	brightened := colorful.Hsl(h, s, l)
	return lipgloss.Color(brightened.Hex())
}

// SetTheme updates the glow's theme reference.
func (g *TabGlow) SetTheme(theme Theme) {
	g.theme = theme
}

// tick returns a tea.Cmd that fires a tabGlowTickMsg after one frame.
func (g TabGlow) tick() tea.Cmd {
	return tea.Tick(tabGlowTickInterval, func(_ time.Time) tea.Msg {
		return tabGlowTickMsg{}
	})
}
