package tui

import (
	"testing"
	"time"
)

func TestMouseToTextPos(t *testing.T) {
	m := Model{}
	m.file.StrippedLines = []string{"hello", "world", "foo"}
	m.file.LineLengths = []int{5, 5, 3}
	m.viewport.ScrollOffset = 0

	// Basic: first line
	line, col := mouseToTextPos(m, 0, 0)
	if line != 1 || col != 0 {
		t.Errorf("expected (1,0), got (%d,%d)", line, col)
	}

	// Second line, col=3
	line, col = mouseToTextPos(m, 3, 1)
	if line != 2 || col != 3 {
		t.Errorf("expected (2,3), got (%d,%d)", line, col)
	}

	// With scroll offset
	m.viewport.ScrollOffset = 1
	line, col = mouseToTextPos(m, 0, 0)
	if line != 2 || col != 0 {
		t.Errorf("expected (2,0) with scroll, got (%d,%d)", line, col)
	}

	// Out of bounds: y too large
	m.viewport.ScrollOffset = 0
	line, _ = mouseToTextPos(m, 0, 10)
	if line != 0 {
		t.Errorf("expected line=0 for OOB, got %d", line)
	}

	// Out of bounds: negative lineIdx (scroll=0, y=0 is fine, but y<0 is impossible from terminal)
	m.viewport.ScrollOffset = 0
	line, col = mouseToTextPos(m, 100, 0)
	// col should clamp to end of line
	if line != 1 || col != 5 {
		t.Errorf("expected (1,5) for past-end col, got (%d,%d)", line, col)
	}
}

func TestIsDoubleClick(t *testing.T) {
	now := time.Now()

	// nil last click
	if isDoubleClick(nil, 1, 0) {
		t.Error("expected false for nil last click")
	}

	// Same line, within window
	last := &clickRecord{Time: now.Add(-200 * time.Millisecond), Line: 1, Col: 5}
	if !isDoubleClick(last, 1, 5) {
		t.Error("expected true for same-line within window")
	}

	// Same line, different col but still same line (should be true)
	if !isDoubleClick(last, 1, 10) {
		t.Error("expected true for same-line different col")
	}

	// Different line
	if isDoubleClick(last, 2, 5) {
		t.Error("expected false for different line")
	}

	// Same line, expired window
	expired := &clickRecord{Time: now.Add(-500 * time.Millisecond), Line: 1, Col: 5}
	if isDoubleClick(expired, 1, 5) {
		t.Error("expected false for expired window")
	}
}

func TestAutoScrollToLine(t *testing.T) {
	m := Model{}
	m.file.RenderedLines = make([]string, 100)
	m.viewport.ViewportHeight = 20

	// Line within viewport — no change
	m.viewport.ScrollOffset = 10
	autoScrollToLine(&m, 15) // lineIdx=14, within [10..29]
	if m.viewport.ScrollOffset != 10 {
		t.Errorf("expected no scroll change, got offset=%d", m.viewport.ScrollOffset)
	}

	// Line above viewport
	m.viewport.ScrollOffset = 20
	autoScrollToLine(&m, 5) // lineIdx=4, above [20..39]
	if m.viewport.ScrollOffset != 4 {
		t.Errorf("expected scroll to 4, got %d", m.viewport.ScrollOffset)
	}

	// Line below viewport
	m.viewport.ScrollOffset = 0
	autoScrollToLine(&m, 25) // lineIdx=24, below [0..19]
	if m.viewport.ScrollOffset != 5 {
		t.Errorf("expected scroll to 5, got %d", m.viewport.ScrollOffset)
	}
}

func TestLineLength(t *testing.T) {
	m := Model{}
	m.file.LineLengths = []int{5, 10, 3}

	// Normal
	if l := lineLength(m, 1); l != 5 {
		t.Errorf("expected 5, got %d", l)
	}
	if l := lineLength(m, 2); l != 10 {
		t.Errorf("expected 10, got %d", l)
	}

	// Out of bounds
	if l := lineLength(m, 0); l != 0 {
		t.Errorf("expected 0 for line 0, got %d", l)
	}
	if l := lineLength(m, 99); l != 0 {
		t.Errorf("expected 0 for line 99, got %d", l)
	}
}
