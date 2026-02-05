package app

import (
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lucasb-eyer/go-colorful"
)

// shimmerTickInterval is the frame rate for the shimmer animation.
const shimmerTickInterval = 16 * time.Millisecond // ~60fps

// shimmerTickMsg advances the shimmer animation by one frame.
type shimmerTickMsg struct {
	id string
}

// Shimmer animates an organic brightness wave across text content. Pure
// achromatic greys only — no hue. Multiple overlapping wave frequencies,
// smooth 2D value noise for spatial variation, and global breathing create
// a natural, varied effect with irregular blob sizes.
type Shimmer struct {
	id     string
	active bool
	frame  int // monotonic frame counter

	// Base and peak lightness (CIE L*) for pure grey output.
	baseL float64
	peakL float64
}

// greyFromL returns a pure achromatic grey lipgloss.Color for a CIE L* value.
func greyFromL(l float64) lipgloss.Color {
	c := colorful.Lab(l, 0, 0)
	r, g, b := c.Clamped().RGB255()
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

// shimmerLightness extracts CIE L* from a theme color.
func shimmerLightness(c lipgloss.Color) float64 {
	col, err := colorful.Hex(string(c))
	if err != nil {
		return 0.5
	}
	l, _, _ := col.Lab()
	return l
}

// NewShimmer creates a Shimmer with default parameters.
func NewShimmer(id string, theme Theme) Shimmer {
	return Shimmer{
		id:    id,
		baseL: shimmerLightness(theme.Colors.Muted),
		peakL: shimmerLightness(theme.Colors.Fg),
	}
}

// Start begins the shimmer animation and returns the first tick command.
func (s *Shimmer) Start() tea.Cmd {
	s.active = true
	s.frame = 0
	return s.tick()
}

// Stop halts the shimmer animation.
func (s *Shimmer) Stop() {
	s.active = false
}

// Active returns whether the shimmer is currently animating.
func (s Shimmer) Active() bool {
	return s.active
}

// Update advances the shimmer by one frame on a matching tick message.
func (s Shimmer) Update(msg tea.Msg) (Shimmer, tea.Cmd) {
	if tick, ok := msg.(shimmerTickMsg); ok && tick.id == s.id && s.active {
		s.frame++
		return s, s.tick()
	}
	return s, nil
}

// SetTheme updates the shimmer lightness values from the theme.
func (s *Shimmer) SetTheme(theme Theme) {
	s.baseL = shimmerLightness(theme.Colors.Muted)
	s.peakL = shimmerLightness(theme.Colors.Fg)
}

// Render applies the shimmer to text, returning per-character styled output.
// textWidth is the number of columns in the widest line.
func (s Shimmer) Render(text string, textWidth int) string {
	if textWidth <= 0 {
		return text
	}

	lines := strings.Split(text, "\n")
	var b strings.Builder
	b.Grow(len(text) * 3)

	baseColor := greyFromL(s.baseL)
	baseStyle := lipgloss.NewStyle().Foreground(baseColor)

	for li, line := range lines {
		if li > 0 {
			b.WriteByte('\n')
		}
		col := 0
		for i := 0; i < len(line); {
			r, size := utf8.DecodeRuneInString(line[i:])
			i += size

			// Skip empty Braille (U+2800) — no visual dots to highlight.
			if r == '\u2800' {
				b.WriteRune(r)
				col++
				continue
			}

			brightness := s.brightnessAt(li, col, textWidth)
			if brightness > 0.005 {
				l := s.baseL + (s.peakL-s.baseL)*brightness
				style := lipgloss.NewStyle().Foreground(greyFromL(l))
				b.WriteString(style.Render(string(r)))
			} else {
				b.WriteString(baseStyle.Render(string(r)))
			}
			col++
		}
	}

	return b.String()
}

// brightnessAt computes the combined brightness boost (0..1) for a cell.
// Instead of uniform wave sweeps, brightness is sampled directly from a
// time-evolving 3D noise field. Per-row drift uses competing sinusoids at
// incommensurate frequencies, producing natural slowdowns, pauses, and
// occasional direction reversals. Multiple noise layers at different spatial
// scales create varied blob shapes and sizes.
func (s Shimmer) brightnessAt(row, col, textWidth int) float64 {
	t := float64(s.frame)
	if textWidth <= 0 {
		return 0
	}
	r := float64(row)
	c := float64(col)

	// Per-row horizontal drift: sum of sinusoids at incommensurate frequencies.
	// When they align, the row sweeps smoothly; when they oppose, the row
	// slows, pauses, or reverses. Each row has a different phase offset.
	drift := 3.0*math.Sin(t*0.006+r*0.41) +
		2.0*math.Sin(t*0.011+r*0.67) +
		1.5*math.Sin(t*0.003+r*0.23) +
		1.0*math.Sin(t*0.017+r*1.1)

	// Noise-sampled coordinates: column shifts with drift for the sweep,
	// row provides vertical variation, and time evolves the field.
	nx := (c + drift) * 0.14
	ny := r * 0.22
	nz := t * 0.004

	// Layer 1: primary — medium-scale blobs.
	n1 := fbmNoise(nx, ny, nz)
	b1 := smoothThreshold(n1, 0.52, 0.18)

	// Layer 2: large slow wash — broad ambient glow at a different drift rate.
	drift2 := 2.0*math.Sin(t*0.004+r*0.3) +
		1.5*math.Sin(t*0.009+r*0.55)
	nx2 := (c + drift2) * 0.07
	ny2 := r * 0.1
	n2 := fbmNoise(nx2, ny2, t*0.002+80)
	b2 := smoothThreshold(n2, 0.5, 0.25) * 0.35

	// Layer 3: fine detail — small bright speckles drifting independently.
	drift3 := 2.5*math.Sin(t*0.014+r*0.8) +
		1.0*math.Sin(t*0.008+r*0.35)
	nx3 := (c + drift3) * 0.25
	ny3 := r * 0.35
	n3 := fbmNoise(nx3, ny3, t*0.006+160)
	b3 := smoothThreshold(n3, 0.58, 0.12) * 0.3

	combined := b1 + b2 + b3

	// Global breathing: slow oscillation of overall intensity.
	breath := 0.7 + 0.3*math.Sin(t*0.010)
	combined *= breath

	if combined > 1 {
		combined = 1
	}
	return combined
}

// smoothThreshold maps a noise value (0..1) through a soft step centered at
// 'center' with the given radius. Returns 0 below center-radius, 1 above
// center+radius, and a smooth ramp between. This creates distinct bright
// regions with soft edges and varied shapes.
func smoothThreshold(value, center, radius float64) float64 {
	low := center - radius
	high := center + radius
	if value <= low {
		return 0
	}
	if value >= high {
		return 1
	}
	t := (value - low) / (high - low)
	// Smoothstep.
	return t * t * (3 - 2*t)
}

// --- Smooth 2D value noise ---

// fbmNoise returns fractal Brownian motion noise in [0, 1] at the given
// coordinates. Three octaves of smooth value noise at increasing frequency
// and decreasing amplitude produce natural, multi-scale variation.
func fbmNoise(x, y, z float64) float64 {
	v := 0.0
	amp := 0.5
	freq := 1.0
	for range 3 {
		v += amp * smoothNoise3D(x*freq, y*freq, z*freq)
		freq *= 2.0
		amp *= 0.5
	}
	// Normalize from roughly [-0.5, 0.5] to [0, 1].
	return v + 0.5
}

// smoothNoise3D returns interpolated value noise in roughly [-0.5, 0.5].
func smoothNoise3D(x, y, z float64) float64 {
	ix := int(math.Floor(x))
	iy := int(math.Floor(y))
	iz := int(math.Floor(z))
	fx := x - math.Floor(x)
	fy := y - math.Floor(y)
	fz := z - math.Floor(z)

	// Smoothstep for organic interpolation.
	fx = fx * fx * (3 - 2*fx)
	fy = fy * fy * (3 - 2*fy)
	fz = fz * fz * (3 - 2*fz)

	// Trilinear interpolation of hashed lattice values.
	c000 := latticeHash(ix, iy, iz)
	c100 := latticeHash(ix+1, iy, iz)
	c010 := latticeHash(ix, iy+1, iz)
	c110 := latticeHash(ix+1, iy+1, iz)
	c001 := latticeHash(ix, iy, iz+1)
	c101 := latticeHash(ix+1, iy, iz+1)
	c011 := latticeHash(ix, iy+1, iz+1)
	c111 := latticeHash(ix+1, iy+1, iz+1)

	x0 := lerp(c000, c100, fx)
	x1 := lerp(c010, c110, fx)
	x2 := lerp(c001, c101, fx)
	x3 := lerp(c011, c111, fx)

	y0 := lerp(x0, x1, fy)
	y1 := lerp(x2, x3, fy)

	return lerp(y0, y1, fz)
}

// latticeHash returns a deterministic pseudo-random value in [-0.5, 0.5)
// for an integer lattice point.
func latticeHash(x, y, z int) float64 {
	h := uint32(x*374761393+y*668265263+z*1440670441) ^ 0x27d4eb2d
	h = (h ^ (h >> 13)) * 1274126177
	h = h ^ (h >> 16)
	return float64(h&0x7fffffff)/float64(0x80000000) - 0.5
}

// lerp linearly interpolates between a and b.
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// tick returns a tea.Cmd that fires a shimmerTickMsg after one frame interval.
func (s Shimmer) tick() tea.Cmd {
	id := s.id
	return tea.Tick(shimmerTickInterval, func(_ time.Time) tea.Msg {
		return shimmerTickMsg{id: id}
	})
}
