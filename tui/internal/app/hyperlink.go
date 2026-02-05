package app

import (
	"encoding/base64"
	"fmt"
)

// RenderHyperlink wraps displayText in an OSC 8 hyperlink escape sequence.
// Terminals that support OSC 8 render it as a clickable link; others show
// displayText unchanged.
func RenderHyperlink(url, displayText string) string {
	return fmt.Sprintf("\x1b]8;;%s\a%s\x1b]8;;\a", url, displayText)
}

// OSC52Sequence returns an OSC 52 escape sequence that sets the system
// clipboard to the given text. Embed the returned string in View output
// to write to the clipboard on the next render.
func OSC52Sequence(text string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return fmt.Sprintf("\x1b]52;c;%s\a", encoded)
}
