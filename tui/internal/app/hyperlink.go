package app

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// RenderHyperlink wraps displayText in an OSC 8 hyperlink escape sequence.
// Terminals that support OSC 8 render it as a clickable link; others show
// displayText unchanged. The url is sanitized to prevent escape sequence
// injection (ESC, BEL, and control characters are stripped).
func RenderHyperlink(url, displayText string) string {
	return fmt.Sprintf("\x1b]8;;%s\a%s\x1b]8;;\a", sanitizeOSCParam(url), displayText)
}

// OSC52Sequence returns an OSC 52 escape sequence that sets the system
// clipboard to the given text. The payload is base64-encoded, so injection
// is not possible.
//
// Embed the returned string in View output to write to the clipboard on
// the next render. Bubbletea v1 has no tea.Raw; prepending to View works
// in alt screen mode because the terminal parses OSC sequences from the
// byte stream regardless of interleaved cursor-positioning escapes.
// Tested with iTerm2, Ghostty, and WezTerm. Unsupported terminals
// silently ignore the sequence.
func OSC52Sequence(text string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	return fmt.Sprintf("\x1b]52;c;%s\a", encoded)
}

// sanitizeOSCParam strips characters that could terminate or break out of
// an OSC escape sequence: ESC (0x1B), BEL (0x07), and CR/LF.
func sanitizeOSCParam(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\x1b', '\a', '\n', '\r':
			return -1
		}
		return r
	}, s)
}
