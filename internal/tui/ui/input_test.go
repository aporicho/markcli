package ui

import (
	"strings"
	"testing"

	"github.com/aporicho/markcli/internal/theme"
)

func TestRenderInputPanel_EmptyValue(t *testing.T) {
	th := theme.Theme{Dark: true}
	result := RenderInputPanel("", 0, 40, th)

	if result == "" {
		t.Fatal("expected non-empty panel")
	}
}

func TestRenderInputPanel_WithValue(t *testing.T) {
	th := theme.Theme{Dark: true}
	result := RenderInputPanel("hello", 5, 40, th)

	if result == "" {
		t.Fatal("expected non-empty panel")
	}
	if !strings.Contains(result, "hello") {
		t.Error("expected panel to contain input value")
	}
}

func TestRenderInputPanel_NarrowWidth(t *testing.T) {
	th := theme.Theme{Dark: true}
	result := RenderInputPanel("test", 0, 4, th)

	if result == "" {
		t.Fatal("expected non-empty panel even with narrow width")
	}
}

func TestOverlayAt_BasicOverlay(t *testing.T) {
	base := "aaaa\nbbbb\ncccc\ndddd"
	overlay := "XX\nYY"

	result := OverlayAt(base, overlay, 1, 1)
	lines := strings.Split(result, "\n")

	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}
	// Line 0 should be unchanged
	if lines[0] != "aaaa" {
		t.Errorf("line 0: expected 'aaaa', got %q", lines[0])
	}
}

func TestOverlayAt_OutOfBounds(t *testing.T) {
	base := "aa\nbb"
	overlay := "XX\nYY\nZZ"

	// overlay starts at row 1, only 1 row fits (row 1), row 2 is out of bounds
	result := OverlayAt(base, overlay, 1, 0)
	lines := strings.Split(result, "\n")

	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "aa" {
		t.Errorf("line 0 should be unchanged, got %q", lines[0])
	}
}

func TestOverlayAt_TopZero(t *testing.T) {
	base := "aaaa\nbbbb"
	overlay := "XX"

	result := OverlayAt(base, overlay, 0, 0)
	lines := strings.Split(result, "\n")

	if lines[1] != "bbbb" {
		t.Errorf("line 1 should be unchanged, got %q", lines[1])
	}
}

func TestOverlayLine_PlainText(t *testing.T) {
	result := overlayLine("abcdefgh", "XY", 2)
	// "ab" + "XY" + skip 'cd' (2 cols) + "efgh"
	if result != "ab\x1b[0mXYefgh" {
		t.Errorf("expected 'ab\\x1b[0mXYefgh', got %q", result)
	}
}

func TestOverlayLine_BaseWithANSI(t *testing.T) {
	// base: "a" + red-styled "bc" + "defgh"
	base := "a\x1b[31mbc\x1b[0mdefgh"
	result := overlayLine(base, "XY", 1)
	// At col 1: 'a' written (col=1), then ANSI \x1b[31m passed through
	// Actually at col 1 we reached left, so ANSI before col 1 is written,
	// then overlay "XY" replaces 'bc' (2 cols)
	// Just verify overlay content is present and base isn't corrupted
	if !strings.Contains(result, "XY") {
		t.Errorf("expected overlay 'XY' in result, got %q", result)
	}
	if !strings.Contains(result, "a") {
		t.Errorf("expected base prefix 'a' preserved, got %q", result)
	}
}

func TestConsumeAnsi(t *testing.T) {
	tests := []struct {
		input []rune
		want  int
	}{
		{[]rune("\x1b[31m"), 5},           // basic SGR
		{[]rune("\x1b[38;5;196m"), 11},    // 256-color
		{[]rune("\x1b[0m"), 4},            // reset
		{[]rune("hello"), 0},              // not ANSI
		{[]rune("\x1b"), 0},               // incomplete
		{[]rune("\x1b["), 0},              // incomplete
		{[]rune("\x1b[1;2;3H"), 8},        // cursor position (CSI)
	}
	for _, tt := range tests {
		got := consumeAnsi(tt.input)
		if got != tt.want {
			t.Errorf("consumeAnsi(%q) = %d, want %d", string(tt.input), got, tt.want)
		}
	}
}
