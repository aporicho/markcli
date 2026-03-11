package tui

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/tui/ui"
)

// submitAnnotation saves or updates an annotation and reloads the file.
// If m.editingID is set, updates the existing annotation; otherwise creates a new one.
func submitAnnotation(m Model) (Model, tea.Cmd) {
	comment := strings.TrimSpace(string(m.input.Value))
	if comment == "" {
		return m, nil
	}

	if m.editingID != "" {
		// Edit existing annotation
		for i, ann := range m.annotations {
			if ann.ID == m.editingID {
				m.annotations[i].Comment = comment
				break
			}
		}
		m.editingID = ""
	} else {
		// Create new annotation
		s, e := ui.NormalizePos(m.selection.Start, m.selection.End)
		selectedText := getSelectedText(m.file.StrippedLines, s, e)

		fullText := strings.Join(m.file.StrippedLines, "\n")
		startOffset := annotation.LineColToOffset(m.file.LineLengths, s.Line, s.Col)
		endOffset := annotation.LineColToOffset(m.file.LineLengths, e.Line, e.Col)

		anchor := annotation.ExtractAnchor(fullText, startOffset, endOffset)

		startCol := s.Col
		endCol := e.Col
		ann := annotation.Annotation{
			ID:           generateID(),
			StartLine:    s.Line,
			EndLine:      e.Line,
			StartCol:     &startCol,
			EndCol:       &endCol,
			SelectedText: selectedText,
			Comment:      comment,
			CreatedAt:    time.Now().Format(time.RFC3339),
			Quote:        anchor.Quote,
			Prefix:       anchor.Prefix,
			Suffix:       anchor.Suffix,
		}
		m.annotations = append(m.annotations, ann)
	}

	af := annotation.AnnotationFile{
		File:        filepath.Base(m.file.FilePath),
		Annotations: m.annotations,
	}
	if err := annotation.Save(m.file.FilePath, af); err != nil {
		return m, func() tea.Msg { return errMsg{err: fmt.Errorf("save failed: %w", err)} }
	}

	m.input = inputState{}
	m.selection = selectionState{}
	m.mode = ui.ModeReading

	return m, loadFileCmd(m.file.FilePath, m.viewport.Width, loadIPC)
}

// deleteAnnotation removes the annotation at overview.Cursor and persists.
func deleteAnnotation(m Model) (Model, tea.Cmd) {
	if len(m.resolved) == 0 {
		return m, nil
	}
	cursor := m.overview.Cursor
	if cursor < 0 || cursor >= len(m.resolved) {
		return m, nil
	}

	targetID := m.resolved[cursor].ID

	// Remove from m.annotations by ID
	newAnns := make([]annotation.Annotation, 0, len(m.annotations))
	for _, ann := range m.annotations {
		if ann.ID != targetID {
			newAnns = append(newAnns, ann)
		}
	}
	m.annotations = newAnns

	af := annotation.AnnotationFile{
		File:        filepath.Base(m.file.FilePath),
		Annotations: m.annotations,
	}
	if err := annotation.Save(m.file.FilePath, af); err != nil {
		return m, func() tea.Msg { return errMsg{err: fmt.Errorf("save failed: %w", err)} }
	}

	// Adjust cursor
	if m.overview.Cursor >= len(m.annotations) && m.overview.Cursor > 0 {
		m.overview.Cursor--
	}

	// If all deleted, return to reading
	if len(m.annotations) == 0 {
		m.mode = ui.ModeReading
		m.overview = overviewState{}
	}

	return m, loadFileCmd(m.file.FilePath, m.viewport.Width, loadIPC)
}

// enterEditFromOverview sets up annotating mode to edit the annotation at overview cursor.
func enterEditFromOverview(m Model) Model {
	if len(m.resolved) == 0 {
		return m
	}
	cursor := m.overview.Cursor
	if cursor < 0 || cursor >= len(m.resolved) {
		return m
	}

	ann := m.resolved[cursor]
	m.editingID = ann.ID

	// Set selection from annotation position
	startCol := 0
	if ann.StartCol != nil {
		startCol = *ann.StartCol
	}
	endCol := 0
	if ann.EndCol != nil {
		endCol = *ann.EndCol
	}
	m.selection = selectionState{
		Active: true,
		Start:  annotation.SelectionPos{Line: ann.StartLine, Col: startCol},
		End:    annotation.SelectionPos{Line: ann.EndLine, Col: endCol},
	}

	// Pre-fill input with existing comment
	runes := []rune(ann.Comment)
	m.input = inputState{
		Value:  runes,
		Cursor: len(runes),
	}
	m.mode = ui.ModeAnnotating

	// Scroll to annotation
	autoScrollToLine(&m, ann.StartLine)

	return m
}

// getSelectedText extracts text from strippedLines between normalized start and end positions.
func getSelectedText(strippedLines []string, start, end annotation.SelectionPos) string {
	if len(strippedLines) == 0 {
		return ""
	}

	if start.Line == end.Line {
		idx := start.Line - 1
		if idx < 0 || idx >= len(strippedLines) {
			return ""
		}
		runes := []rune(strippedLines[idx])
		s := clamp(start.Col, 0, len(runes))
		e := clamp(end.Col, 0, len(runes))
		if e <= s {
			return ""
		}
		return string(runes[s:e])
	}

	var sb strings.Builder

	// First line
	startIdx := start.Line - 1
	if startIdx >= 0 && startIdx < len(strippedLines) {
		runes := []rune(strippedLines[startIdx])
		s := clamp(start.Col, 0, len(runes))
		sb.WriteString(string(runes[s:]))
	}

	// Middle lines
	for lineIdx := start.Line; lineIdx < end.Line-1; lineIdx++ {
		if lineIdx >= 0 && lineIdx < len(strippedLines) {
			sb.WriteString("\n")
			sb.WriteString(strippedLines[lineIdx])
		}
	}

	// Last line
	endIdx := end.Line - 1
	if endIdx >= 0 && endIdx < len(strippedLines) && end.Line != start.Line {
		sb.WriteString("\n")
		runes := []rune(strippedLines[endIdx])
		e := clamp(end.Col, 0, len(runes))
		sb.WriteString(string(runes[:e]))
	}

	return sb.String()
}

// generateID returns a 6-character random hex ID.
func generateID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
