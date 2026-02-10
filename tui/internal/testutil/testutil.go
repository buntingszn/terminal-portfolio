// Package testutil provides test helpers and fixtures for the terminal-portfolio TUI.
package testutil

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/buntingszn/terminal-portfolio/tui/internal/app"
	"github.com/buntingszn/terminal-portfolio/tui/internal/content"
)

// fixtureDataDir returns the absolute path to testdata/ relative to this source file.
func fixtureDataDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "testdata")
}

// FixtureContent returns a fully-populated Content struct loaded from
// testdata/content/*.json. Panics on load failure so tests fail fast.
func FixtureContent() *content.Content {
	c, err := content.LoadAll(fixtureDataDir())
	if err != nil {
		panic("testutil: failed to load fixture content: " + err.Error())
	}
	return c
}

// FixtureTheme returns the default dark theme for testing.
func FixtureTheme() app.Theme {
	return app.DarkTheme()
}

// RequireContains fails the test if s does not contain substr.
func RequireContains(t *testing.T, s, substr string) {
	t.Helper()
	if len(substr) == 0 {
		t.Fatal("RequireContains: substr must not be empty")
	}
	if !contains(s, substr) {
		t.Errorf("expected string to contain %q, got %q", substr, s)
	}
}

// RequireNotEmpty fails the test if s is empty.
func RequireNotEmpty(t *testing.T, s string) {
	t.Helper()
	if len(s) == 0 {
		t.Error("expected non-empty string, got empty")
	}
}

// contains checks whether s contains substr without importing strings.
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
