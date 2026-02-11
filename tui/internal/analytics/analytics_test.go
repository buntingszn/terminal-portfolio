package analytics

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewLoggerDisabled(t *testing.T) {
	l, err := NewLogger("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if l != nil {
		t.Error("expected nil logger when path is empty")
	}
}

func TestNilLoggerSafe(t *testing.T) {
	var l *Logger
	// Should not panic.
	l.Log(Event{Type: EventSessionStart})
	if err := l.Close(); err != nil {
		t.Errorf("Close on nil logger: %v", err)
	}
}

func TestLogWritesJSONL(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.jsonl")

	l, err := NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}

	now := time.Now()
	l.Log(Event{
		Timestamp: now,
		SessionID: "abc123",
		Type:      EventSessionStart,
		IP:        "1.2.3.4",
	})
	l.Log(Event{
		Timestamp: now,
		SessionID: "abc123",
		Type:      EventSectionView,
		Section:   "home",
		DurationMs: 5000,
	})
	if err := l.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	var events []Event
	for scanner.Scan() {
		var e Event
		if err := json.Unmarshal(scanner.Bytes(), &e); err != nil {
			t.Fatalf("unmarshal line: %v", err)
		}
		events = append(events, e)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Type != EventSessionStart {
		t.Errorf("event[0].Type = %q, want %q", events[0].Type, EventSessionStart)
	}
	if events[0].SessionID != "abc123" {
		t.Errorf("event[0].SessionID = %q, want %q", events[0].SessionID, "abc123")
	}
	if events[0].IP != "1.2.3.4" {
		t.Errorf("event[0].IP = %q, want %q", events[0].IP, "1.2.3.4")
	}
	if events[1].Type != EventSectionView {
		t.Errorf("event[1].Type = %q, want %q", events[1].Type, EventSectionView)
	}
	if events[1].Section != "home" {
		t.Errorf("event[1].Section = %q, want %q", events[1].Section, "home")
	}
	if events[1].DurationMs != 5000 {
		t.Errorf("event[1].DurationMs = %d, want 5000", events[1].DurationMs)
	}
}

func TestLogAppendsToExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "append.jsonl")

	// Write one event, close.
	l, err := NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	l.Log(Event{SessionID: "s1", Type: EventSessionStart})
	if err := l.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Open again, write another.
	l2, err := NewLogger(path)
	if err != nil {
		t.Fatalf("NewLogger (second): %v", err)
	}
	l2.Log(Event{SessionID: "s2", Type: EventSessionStart})
	if err := l2.Close(); err != nil {
		t.Fatalf("Close (second): %v", err)
	}

	// Should have 2 lines.
	f, _ := os.Open(path)
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	count := 0
	for scanner.Scan() {
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 lines after append, got %d", count)
	}
}

func TestNewLoggerInvalidPath(t *testing.T) {
	_, err := NewLogger("/nonexistent/dir/file.jsonl")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}
