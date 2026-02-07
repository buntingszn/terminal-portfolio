package app

// DensityLevel controls vertical spacing between sections based on terminal height.
type DensityLevel int

const (
	DensityCompact     DensityLevel = iota // < 30 rows
	DensityComfortable                     // 30-49 rows
	DensitySpacious                        // >= 50 rows
)

// DensityForHeight returns the appropriate density level for the given terminal height.
func DensityForHeight(height int) DensityLevel {
	switch {
	case height < 30:
		return DensityCompact
	case height < 50:
		return DensityComfortable
	default:
		return DensitySpacious
	}
}

// SectionSeparator returns the vertical separator string between content sections
// for the given density level.
func SectionSeparator(d DensityLevel) string {
	switch d {
	case DensityCompact:
		return "\n"
	case DensityComfortable:
		return "\n\n"
	default:
		return "\n\n\n"
	}
}
