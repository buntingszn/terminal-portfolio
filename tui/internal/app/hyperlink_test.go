package app

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestRenderHyperlink(t *testing.T) {
	url := "https://example.com"
	text := "Example"
	got := RenderHyperlink(url, text)
	want := "\x1b]8;;https://example.com\aExample\x1b]8;;\a"
	if got != want {
		t.Errorf("RenderHyperlink(%q, %q) = %q, want %q", url, text, got, want)
	}
}

func TestRenderHyperlink_EmptyURL(t *testing.T) {
	got := RenderHyperlink("", "text")
	want := "\x1b]8;;\atext\x1b]8;;\a"
	if got != want {
		t.Errorf("RenderHyperlink(\"\", \"text\") = %q, want %q", got, want)
	}
}

func TestRenderHyperlink_SanitizesInjection(t *testing.T) {
	// A URL containing ESC and BEL should have those bytes stripped,
	// preventing early termination of the OSC 8 sequence.
	malicious := "https://evil.com\x1b]8;;\aINJECTED"
	got := RenderHyperlink(malicious, "click")
	// After stripping \x1b and \a, the "]8;;" remnants are harmless text.
	want := "\x1b]8;;https://evil.com]8;;INJECTED\aclick\x1b]8;;\a"
	if got != want {
		t.Errorf("RenderHyperlink did not sanitize URL:\ngot:  %q\nwant: %q", got, want)
	}
	// Extract the href portion (between the opening \x1b]8;; and closing \a).
	href := got[len("\x1b]8;;") : strings.Index(got, "\a")]
	if strings.Contains(href, "\a") {
		t.Errorf("sanitized href still contains BEL: %q", href)
	}
	if strings.Contains(href, "\x1b") {
		t.Errorf("sanitized href still contains ESC: %q", href)
	}
}

func TestOSC52Sequence(t *testing.T) {
	text := "https://example.com"
	got := OSC52Sequence(text)
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	want := "\x1b]52;c;" + encoded + "\a"
	if got != want {
		t.Errorf("OSC52Sequence(%q) = %q, want %q", text, got, want)
	}
}
