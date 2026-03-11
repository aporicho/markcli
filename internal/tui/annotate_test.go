package tui

import (
	"testing"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/tui/ui"
)

func TestGetSelectedText_SingleLine(t *testing.T) {
	lines := []string{"hello world", "second line"}
	start := annotation.SelectionPos{Line: 1, Col: 0}
	end := annotation.SelectionPos{Line: 1, Col: 5}

	got := getSelectedText(lines, start, end)
	if got != "hello" {
		t.Errorf("expected 'hello', got %q", got)
	}
}

func TestGetSelectedText_SingleLineMiddle(t *testing.T) {
	lines := []string{"hello world"}
	start := annotation.SelectionPos{Line: 1, Col: 6}
	end := annotation.SelectionPos{Line: 1, Col: 11}

	got := getSelectedText(lines, start, end)
	if got != "world" {
		t.Errorf("expected 'world', got %q", got)
	}
}

func TestGetSelectedText_MultiLine(t *testing.T) {
	lines := []string{"first line", "second line", "third line"}
	start := annotation.SelectionPos{Line: 1, Col: 6}
	end := annotation.SelectionPos{Line: 3, Col: 5}

	got := getSelectedText(lines, start, end)
	want := "line\nsecond line\nthird"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestGetSelectedText_EmptyLines(t *testing.T) {
	lines := []string{}
	start := annotation.SelectionPos{Line: 1, Col: 0}
	end := annotation.SelectionPos{Line: 1, Col: 5}

	got := getSelectedText(lines, start, end)
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestGetSelectedText_EmptySelection(t *testing.T) {
	lines := []string{"hello"}
	start := annotation.SelectionPos{Line: 1, Col: 3}
	end := annotation.SelectionPos{Line: 1, Col: 3}

	got := getSelectedText(lines, start, end)
	if got != "" {
		t.Errorf("expected empty string for zero-width selection, got %q", got)
	}
}

func TestGenerateID_Length(t *testing.T) {
	id := generateID()
	if len(id) != 6 {
		t.Errorf("expected 6 chars, got %d: %q", len(id), id)
	}
}

func TestGenerateID_Unique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := generateID()
		if ids[id] {
			t.Errorf("duplicate ID generated: %q", id)
		}
		ids[id] = true
	}
}

func TestEnterEditFromOverview(t *testing.T) {
	startCol := 2
	endCol := 7
	m := Model{
		resolved: []annotation.Annotation{
			{
				ID:        "abc",
				StartLine: 3,
				EndLine:   3,
				StartCol:  &startCol,
				EndCol:    &endCol,
				Comment:   "fix me",
			},
		},
		overview: overviewState{Cursor: 0},
		file: fileState{
			RenderedLines: make([]string, 20),
		},
		viewport: viewportState{ViewportHeight: 20},
	}

	m = enterEditFromOverview(m)

	if m.editingID != "abc" {
		t.Errorf("expected editingID='abc', got %q", m.editingID)
	}
	if m.mode != ui.ModeAnnotating {
		t.Errorf("expected ModeAnnotating, got %v", m.mode)
	}
	if !m.selection.Active {
		t.Error("expected selection to be active")
	}
	if m.selection.Start.Line != 3 || m.selection.Start.Col != 2 {
		t.Errorf("expected start (3,2), got (%d,%d)", m.selection.Start.Line, m.selection.Start.Col)
	}
	if string(m.input.Value) != "fix me" {
		t.Errorf("expected input 'fix me', got %q", string(m.input.Value))
	}
	if m.input.Cursor != 6 {
		t.Errorf("expected cursor at 6, got %d", m.input.Cursor)
	}
}

func TestEnterEditFromOverview_Empty(t *testing.T) {
	m := Model{
		resolved: nil,
		overview: overviewState{Cursor: 0},
	}

	m = enterEditFromOverview(m)
	if m.editingID != "" {
		t.Error("expected no editingID for empty list")
	}
}

func TestDeleteAnnotation(t *testing.T) {
	m := Model{
		annotations: []annotation.Annotation{
			{ID: "a1", Comment: "first"},
			{ID: "a2", Comment: "second"},
			{ID: "a3", Comment: "third"},
		},
		resolved: []annotation.Annotation{
			{ID: "a1", Comment: "first"},
			{ID: "a2", Comment: "second"},
			{ID: "a3", Comment: "third"},
		},
		overview: overviewState{Cursor: 1}, // delete "a2"
		file:     fileState{FilePath: "/tmp/test-delete-ann.md"},
		viewport: viewportState{Width: 80},
	}

	m, _ = deleteAnnotation(m)

	if len(m.annotations) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(m.annotations))
	}
	for _, a := range m.annotations {
		if a.ID == "a2" {
			t.Error("expected 'a2' to be deleted")
		}
	}
}

func TestDeleteAnnotation_LastItem(t *testing.T) {
	m := Model{
		annotations: []annotation.Annotation{
			{ID: "a1", Comment: "first"},
			{ID: "a2", Comment: "second"},
		},
		resolved: []annotation.Annotation{
			{ID: "a1", Comment: "first"},
			{ID: "a2", Comment: "second"},
		},
		overview: overviewState{Cursor: 1}, // delete last item
		file:     fileState{FilePath: "/tmp/test-delete-last.md"},
		viewport: viewportState{Width: 80},
	}

	m, _ = deleteAnnotation(m)

	if m.overview.Cursor != 0 {
		t.Errorf("expected cursor to move to 0, got %d", m.overview.Cursor)
	}
}

func TestDeleteAnnotation_AllDeleted(t *testing.T) {
	m := Model{
		annotations: []annotation.Annotation{
			{ID: "a1", Comment: "only"},
		},
		resolved: []annotation.Annotation{
			{ID: "a1", Comment: "only"},
		},
		overview: overviewState{Cursor: 0},
		file:     fileState{FilePath: "/tmp/test-delete-all.md"},
		viewport: viewportState{Width: 80},
	}

	m, _ = deleteAnnotation(m)

	if m.mode != ui.ModeReading {
		t.Errorf("expected ModeReading after all deleted, got %v", m.mode)
	}
	if len(m.annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(m.annotations))
	}
}
