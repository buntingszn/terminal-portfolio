package app

import (
	"encoding/base64"
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

func TestOSC52Sequence(t *testing.T) {
	text := "https://example.com"
	got := OSC52Sequence(text)
	encoded := base64.StdEncoding.EncodeToString([]byte(text))
	want := "\x1b]52;c;" + encoded + "\a"
	if got != want {
		t.Errorf("OSC52Sequence(%q) = %q, want %q", text, got, want)
	}
}
