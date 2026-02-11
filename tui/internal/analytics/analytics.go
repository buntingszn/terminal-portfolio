package analytics

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// EventType identifies the kind of analytics event.
type EventType string

const (
	EventSessionStart EventType = "session_start"
	EventSessionEnd   EventType = "session_end"
	EventSectionView  EventType = "section_view"
)

// Event is a single analytics record written as JSON Lines.
type Event struct {
	Timestamp  time.Time `json:"ts"`
	SessionID  string    `json:"sid"`
	Type       EventType `json:"type"`
	IP         string    `json:"ip,omitempty"`
	Section    string    `json:"section,omitempty"`
	DurationMs int64     `json:"duration_ms,omitempty"`
}

// Logger writes analytics events as JSON Lines to a file.
// A nil Logger is safe to use; all methods are no-ops.
type Logger struct {
	mu   sync.Mutex
	file *os.File
}

// NewLogger opens (or creates) the analytics file in append mode.
// If path is empty, analytics are disabled and nil is returned.
func NewLogger(path string) (*Logger, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &Logger{file: f}, nil
}

// Log writes a single event as a JSON line. No-op on nil Logger.
func (l *Logger) Log(e Event) {
	if l == nil {
		return
	}
	data, err := json.Marshal(e)
	if err != nil {
		return
	}
	data = append(data, '\n')
	l.mu.Lock()
	defer l.mu.Unlock()
	_, _ = l.file.Write(data)
}

// Close closes the underlying file. No-op on nil Logger.
func (l *Logger) Close() error {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.file.Close()
}
