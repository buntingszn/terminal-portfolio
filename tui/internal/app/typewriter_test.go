package app

import (
	"testing"
)

func TestNewTypewriter(t *testing.T) {
	tw := NewTypewriter("test", "hello", 1)
	if tw.Done() {
		t.Error("new typewriter should not be done")
	}
	if tw.View() != "" {
		t.Errorf("View() = %q, want empty string", tw.View())
	}
}

func TestNewTypewriterEmptyText(t *testing.T) {
	tw := NewTypewriter("empty", "", 1)
	if !tw.Done() {
		t.Error("typewriter with empty text should be done immediately")
	}
	if tw.View() != "" {
		t.Errorf("View() = %q, want empty string", tw.View())
	}
}

func TestNewTypewriterClampsSpeed(t *testing.T) {
	tw := NewTypewriter("clamp", "hi", 0)
	if tw.charsPerTick != 1 {
		t.Errorf("charsPerTick = %d, want 1 (clamped from 0)", tw.charsPerTick)
	}

	tw = NewTypewriter("neg", "hi", -5)
	if tw.charsPerTick != 1 {
		t.Errorf("charsPerTick = %d, want 1 (clamped from -5)", tw.charsPerTick)
	}
}

func TestTypewriterAdvancesOneCharPerTick(t *testing.T) {
	tw := NewTypewriter("t1", "abc", 1)

	// First tick: reveal 'a'.
	tw, cmd := tw.Update(typewriterTickMsg{id: "t1"})
	if tw.View() != "a" {
		t.Errorf("after tick 1: View() = %q, want %q", tw.View(), "a")
	}
	if tw.Done() {
		t.Error("should not be done after 1 of 3 chars")
	}
	if cmd == nil {
		t.Error("expected a tick command, got nil")
	}

	// Second tick: reveal 'ab'.
	tw, cmd = tw.Update(typewriterTickMsg{id: "t1"})
	if tw.View() != "ab" {
		t.Errorf("after tick 2: View() = %q, want %q", tw.View(), "ab")
	}
	if tw.Done() {
		t.Error("should not be done after 2 of 3 chars")
	}
	if cmd == nil {
		t.Error("expected a tick command, got nil")
	}

	// Third tick: reveal 'abc' and complete.
	tw, cmd = tw.Update(typewriterTickMsg{id: "t1"})
	if tw.View() != "abc" {
		t.Errorf("after tick 3: View() = %q, want %q", tw.View(), "abc")
	}
	if !tw.Done() {
		t.Error("should be done after revealing all chars")
	}
	// cmd should produce TypewriterDoneMsg.
	if cmd == nil {
		t.Fatal("expected done command, got nil")
	}
	msg := cmd()
	done, ok := msg.(TypewriterDoneMsg)
	if !ok {
		t.Fatalf("expected TypewriterDoneMsg, got %T", msg)
	}
	if done.ID != "t1" {
		t.Errorf("done ID = %q, want %q", done.ID, "t1")
	}
}

func TestTypewriterMultipleCharsPerTick(t *testing.T) {
	tw := NewTypewriter("fast", "hello", 3)

	// First tick: reveal 'hel'.
	tw, _ = tw.Update(typewriterTickMsg{id: "fast"})
	if tw.View() != "hel" {
		t.Errorf("after tick 1: View() = %q, want %q", tw.View(), "hel")
	}
	if tw.Done() {
		t.Error("should not be done after 3 of 5 chars")
	}

	// Second tick: reveal all 'hello' (3 more would overshoot, capped at len).
	tw, cmd := tw.Update(typewriterTickMsg{id: "fast"})
	if tw.View() != "hello" {
		t.Errorf("after tick 2: View() = %q, want %q", tw.View(), "hello")
	}
	if !tw.Done() {
		t.Error("should be done after position exceeds text length")
	}
	if cmd == nil {
		t.Fatal("expected done command")
	}
	msg := cmd()
	if _, ok := msg.(TypewriterDoneMsg); !ok {
		t.Fatalf("expected TypewriterDoneMsg, got %T", msg)
	}
}

func TestTypewriterIgnoresWrongID(t *testing.T) {
	tw := NewTypewriter("mine", "abc", 1)

	tw, cmd := tw.Update(typewriterTickMsg{id: "other"})
	if tw.View() != "" {
		t.Errorf("View() = %q, want empty (wrong ID should be ignored)", tw.View())
	}
	if cmd != nil {
		t.Error("expected nil cmd when tick ID does not match")
	}
}

func TestTypewriterIgnoresUnrelatedMsg(t *testing.T) {
	tw := NewTypewriter("tw", "abc", 1)

	tw, cmd := tw.Update("some string message")
	if tw.View() != "" {
		t.Errorf("View() = %q, want empty (unrelated msg should be ignored)", tw.View())
	}
	if cmd != nil {
		t.Error("expected nil cmd for unrelated message")
	}
}

func TestTypewriterSkip(t *testing.T) {
	tw := NewTypewriter("skip", "hello world", 1)

	// Advance one tick so position is not zero.
	tw, _ = tw.Update(typewriterTickMsg{id: "skip"})
	if tw.Done() {
		t.Error("should not be done yet")
	}

	tw.Skip()
	if !tw.Done() {
		t.Error("should be done after Skip()")
	}
	if tw.View() != "hello world" {
		t.Errorf("View() = %q, want %q", tw.View(), "hello world")
	}
}

func TestTypewriterNoOpAfterDone(t *testing.T) {
	tw := NewTypewriter("done", "ab", 2)

	// Complete in one tick.
	tw, _ = tw.Update(typewriterTickMsg{id: "done"})
	if !tw.Done() {
		t.Fatal("should be done")
	}

	// Another tick should be a no-op.
	tw, cmd := tw.Update(typewriterTickMsg{id: "done"})
	if cmd != nil {
		t.Error("expected nil cmd after typewriter is done")
	}
	if tw.View() != "ab" {
		t.Errorf("View() = %q, want %q", tw.View(), "ab")
	}
}

func TestTypewriterTickReturnsCmd(t *testing.T) {
	tw := NewTypewriter("tick", "abc", 1)
	cmd := tw.Tick()
	if cmd == nil {
		t.Error("Tick() should return a non-nil command")
	}
}

func TestTypewriterUnicodeText(t *testing.T) {
	tw := NewTypewriter("uni", "cafe\u0301", 1) // "café" with combining accent

	// Should handle runes, not bytes.
	tw, _ = tw.Update(typewriterTickMsg{id: "uni"})
	if tw.View() != "c" {
		t.Errorf("after tick 1: View() = %q, want %q", tw.View(), "c")
	}

	tw, _ = tw.Update(typewriterTickMsg{id: "uni"})
	tw, _ = tw.Update(typewriterTickMsg{id: "uni"})
	tw, _ = tw.Update(typewriterTickMsg{id: "uni"})
	tw, _ = tw.Update(typewriterTickMsg{id: "uni"})
	if !tw.Done() {
		t.Error("should be done after all runes revealed")
	}
	if tw.View() != "cafe\u0301" {
		t.Errorf("View() = %q, want %q", tw.View(), "cafe\u0301")
	}
}
