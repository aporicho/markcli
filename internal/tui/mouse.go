package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/markcli/internal/ansi"
	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/tui/ui"
)

const doubleClickWindow = 400 * time.Millisecond

func handleMouse(m Model, msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	totalLines := len(m.file.RenderedLines)
	maxOffset := max(0, totalLines-m.viewport.ViewportHeight)

	// Wheel events: scroll viewport in any mode (including annotating)
	switch msg.Button {
	case tea.MouseButtonWheelUp:
		m.viewport.ScrollOffset = clamp(m.viewport.ScrollOffset-3, 0, maxOffset)
		return m, nil
	case tea.MouseButtonWheelDown:
		m.viewport.ScrollOffset = clamp(m.viewport.ScrollOffset+3, 0, maxOffset)
		return m, nil
	}

	// Annotating mode: left click cancels input, ignore drag
	if m.mode == ui.ModeAnnotating {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			if m.editingID != "" {
				m.editingID = ""
				m.input = inputState{}
				m.selection = selectionState{}
				m.mode = ui.ModeOverview
			} else {
				m.input = inputState{}
				m.selection = selectionState{}
				m.mode = ui.ModeReading
			}
		}
		return m, nil
	}

	// Overview mode: left click cancels back to reading
	if m.mode == ui.ModeOverview {
		if msg.Button == tea.MouseButtonLeft && msg.Action == tea.MouseActionPress {
			m.overview = overviewState{}
			m.mode = ui.ModeReading
		}
		return m, nil
	}

	switch msg.Button {
	case tea.MouseButtonLeft:
		return handleLeftButton(m, msg)
	}

	// Left button motion/release comes with Button=MouseButtonNone in some terminals
	if msg.Button == tea.MouseButtonNone {
		switch msg.Action {
		case tea.MouseActionMotion:
			return handleLeftMotion(m, msg)
		case tea.MouseActionRelease:
			return handleLeftRelease(m)
		}
	}

	return m, nil
}

func handleLeftButton(m Model, msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	switch msg.Action {
	case tea.MouseActionPress:
		lineNum, charIdx := mouseToTextPos(m, msg.X, msg.Y)
		if lineNum == 0 {
			return m, nil
		}

		// Double-click detection: edit annotation if clicked on one
		if isDoubleClick(m.selection.LastClick, lineNum, charIdx) {
			m.selection.LastClick = nil
			// Check if position is within any annotation
			for i, ann := range m.resolved {
				if posInAnnotation(lineNum, charIdx, ann) {
					m.overview = overviewState{Cursor: i}
					m = enterEditFromOverview(m)
					return m, nil
				}
			}
			return m, nil
		}

		now := time.Now()
		m.selection.LastClick = &clickRecord{Time: now, Line: lineNum, Col: charIdx}
		m.selection.PendingClick = &clickPos{Line: lineNum, Col: charIdx}

		// Clear any active selection and return to reading
		m.selection.Active = false
		m.mode = ui.ModeReading
		return m, nil

	case tea.MouseActionMotion:
		return handleLeftMotion(m, msg)

	case tea.MouseActionRelease:
		return handleLeftRelease(m)
	}

	return m, nil
}

func handleLeftMotion(m Model, msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.selection.PendingClick == nil && !m.selection.Active {
		return m, nil
	}

	lineNum, charIdx := mouseToTextPos(m, msg.X, msg.Y)
	if lineNum == 0 {
		return m, nil
	}

	// First drag: initialize selection from pending click
	if m.selection.PendingClick != nil && !m.selection.Active {
		m.selection.Active = true
		m.selection.Start = annotation.SelectionPos{
			Line: m.selection.PendingClick.Line,
			Col:  m.selection.PendingClick.Col,
		}
		m.selection.PendingClick = nil
		m.mode = ui.ModeSelecting
	}

	// Update moving endpoint
	m.selection.End = annotation.SelectionPos{Line: lineNum, Col: charIdx}
	return m, nil
}

func handleLeftRelease(m Model) (tea.Model, tea.Cmd) {
	m.selection.PendingClick = nil
	// Selection stays active — user presses Enter to confirm or Esc to cancel
	return m, nil
}

// mouseToTextPos maps terminal coordinates (0-based) to text position.
// Returns lineNum (1-based) and charIdx (0-based rune index).
// Returns lineNum=0 if out of bounds.
func mouseToTextPos(m Model, x, y int) (lineNum, charIdx int) {
	lineIdx := m.viewport.ScrollOffset + y
	totalLines := len(m.file.StrippedLines)
	if lineIdx < 0 || lineIdx >= totalLines {
		return 0, 0
	}
	lineNum = lineIdx + 1 // 1-based
	charIdx = ansi.TermColToCharIndex(m.file.StrippedLines[lineIdx], x)
	return lineNum, charIdx
}

// isDoubleClick returns true if the current click is within 400ms of the last
// click on the same line and within ±2 columns.
func isDoubleClick(last *clickRecord, lineNum, col int) bool {
	if last == nil {
		return false
	}
	if last.Line != lineNum {
		return false
	}
	diff := col - last.Col
	if diff < -2 || diff > 2 {
		return false
	}
	return time.Since(last.Time) <= doubleClickWindow
}

// lineLength returns the rune length for the given 1-based line number.
// Returns 0 for out-of-bounds lines.
func lineLength(m Model, lineNum int) int {
	idx := lineNum - 1
	if idx < 0 || idx >= len(m.file.LineLengths) {
		return 0
	}
	return m.file.LineLengths[idx]
}

// posInAnnotation returns true if (lineNum, charIdx) falls within the annotation's range.
// lineNum is 1-based, charIdx is 0-based.
func posInAnnotation(lineNum, charIdx int, ann annotation.Annotation) bool {
	if lineNum < ann.StartLine || lineNum > ann.EndLine {
		return false
	}
	if lineNum == ann.StartLine && ann.StartCol != nil && charIdx < *ann.StartCol {
		return false
	}
	if lineNum == ann.EndLine && ann.EndCol != nil && charIdx >= *ann.EndCol {
		return false
	}
	return true
}

// autoScrollToLine adjusts the viewport so the given 1-based line is visible.
func autoScrollToLine(m *Model, lineNum int) {
	lineIdx := lineNum - 1 // 0-based
	totalLines := len(m.file.RenderedLines)
	maxOffset := max(0, totalLines-m.viewport.ViewportHeight)

	if lineIdx < m.viewport.ScrollOffset {
		// Line is above viewport
		m.viewport.ScrollOffset = clamp(lineIdx, 0, maxOffset)
	} else if lineIdx >= m.viewport.ScrollOffset+m.viewport.ViewportHeight {
		// Line is below viewport
		m.viewport.ScrollOffset = clamp(lineIdx-m.viewport.ViewportHeight+1, 0, maxOffset)
	}
}
