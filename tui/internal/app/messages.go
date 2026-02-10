package app

// Section identifies a navigable section of the TUI.
type Section int

const (
	SectionHome  Section = 0
	SectionWork  Section = 1
	SectionCV    Section = 2
	SectionLinks Section = 3
)

// SectionCount is the total number of navigable sections.
const SectionCount = 4

// SectionName returns the display name for a section.
func SectionName(s Section) string {
	switch s {
	case SectionHome:
		return "home"
	case SectionWork:
		return "work"
	case SectionCV:
		return "cv"
	case SectionLinks:
		return "links"
	default:
		return "unknown"
	}
}

// NavigateMsg requests navigation to a specific section.
type NavigateMsg struct {
	Section Section
}

// FocusMsg is sent to a section when it becomes the active section.
type FocusMsg struct{}

// BlurMsg is sent to a section when it loses focus.
type BlurMsg struct{}
