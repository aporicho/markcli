package ui

import (
	"strings"
	"testing"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/theme"
)

func TestRenderOverviewPanel_Empty(t *testing.T) {
	th := theme.Theme{Dark: true}
	result := RenderOverviewPanel(nil, 0, 40, 10, th)

	if result == "" {
		t.Fatal("expected non-empty panel")
	}
	if !strings.Contains(result, "批注总览") {
		t.Error("expected title '批注总览'")
	}
	if !strings.Contains(result, "暂无批注") {
		t.Error("expected '暂无批注' for empty list")
	}
}

func TestRenderOverviewPanel_SingleAnnotation(t *testing.T) {
	th := theme.Theme{Dark: true}
	anns := []annotation.Annotation{
		{
			ID:           "abc123",
			StartLine:    5,
			EndLine:      5,
			SelectedText: "hello world",
			Comment:      "fix this",
		},
	}

	result := RenderOverviewPanel(anns, 0, 40, 10, th)

	if result == "" {
		t.Fatal("expected non-empty panel")
	}
	if !strings.Contains(result, "批注总览") {
		t.Error("expected title")
	}
	if !strings.Contains(result, "L5") {
		t.Error("expected line badge 'L5'")
	}
	if !strings.Contains(result, "hello world") {
		t.Error("expected selected text preview")
	}
	if !strings.Contains(result, "fix this") {
		t.Error("expected comment preview")
	}
}

func TestRenderOverviewPanel_MultiLineAnnotation(t *testing.T) {
	th := theme.Theme{Dark: true}
	anns := []annotation.Annotation{
		{
			ID:           "abc123",
			StartLine:    10,
			EndLine:      15,
			SelectedText: "multi line text",
			Comment:      "spans multiple lines",
		},
	}

	result := RenderOverviewPanel(anns, 0, 40, 10, th)

	if !strings.Contains(result, "L10-15") {
		t.Error("expected multi-line badge 'L10-15'")
	}
}

func TestRenderOverviewPanel_CursorHighlight(t *testing.T) {
	th := theme.Theme{Dark: true}
	anns := []annotation.Annotation{
		{ID: "a1", StartLine: 1, EndLine: 1, SelectedText: "first", Comment: "c1"},
		{ID: "a2", StartLine: 2, EndLine: 2, SelectedText: "second", Comment: "c2"},
		{ID: "a3", StartLine: 3, EndLine: 3, SelectedText: "third", Comment: "c3"},
	}

	// Cursor at index 1 — "second" item should have ▸ prefix
	result := RenderOverviewPanel(anns, 1, 40, 10, th)
	if !strings.Contains(result, "▸") {
		t.Error("expected cursor indicator ▸ in result")
	}
}

func TestRenderOverviewPanel_ResolvedItem(t *testing.T) {
	th := theme.Theme{Dark: true}
	resolved := true
	anns := []annotation.Annotation{
		{ID: "a1", StartLine: 1, EndLine: 1, SelectedText: "text", Comment: "c", Resolved: &resolved},
	}

	result := RenderOverviewPanel(anns, -1, 40, 10, th) // cursor -1 = no cursor on this item
	if !strings.Contains(result, "✓") {
		t.Error("expected resolved indicator ✓")
	}
}

func TestRenderOverviewPanel_LongListScrolls(t *testing.T) {
	th := theme.Theme{Dark: true}
	anns := make([]annotation.Annotation, 20)
	for i := range anns {
		anns[i] = annotation.Annotation{
			ID:           generateTestID(i),
			StartLine:    i + 1,
			EndLine:      i + 1,
			SelectedText: "text",
			Comment:      "comment",
		}
	}

	// maxHeight=8 → listHeight=5, cursor at 15 should scroll
	result := RenderOverviewPanel(anns, 15, 40, 8, th)
	if result == "" {
		t.Fatal("expected non-empty panel for scrolled list")
	}
	// Should contain the cursor item's line badge
	if !strings.Contains(result, "L16") {
		t.Error("expected L16 (cursor at index 15, 1-based line 16) to be visible")
	}
}

func TestTruncatePreview(t *testing.T) {
	// Short text: no truncation
	if got := truncatePreview("hello", 10); got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}

	// Newlines replaced
	if got := truncatePreview("a\nb", 10); got != "a↵b" {
		t.Errorf("expected 'a↵b', got %q", got)
	}

	// Truncation with ellipsis
	got := truncatePreview("abcdefghijklmnop", 5)
	if got != "abcde…" {
		t.Errorf("expected 'abcde…', got %q", got)
	}
}

func generateTestID(i int) string {
	return strings.Repeat("a", 5) + string(rune('0'+i%10))
}
