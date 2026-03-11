package tui

import (
	"unicode"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/theme"
	"github.com/aporicho/markcli/internal/tui/ui"
)

func handleKey(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// ctrl+c: always quit
	if msg.String() == "ctrl+c" {
		return m, tea.Quit
	}

	// Annotating mode handles its own keys — no global q/ctrl+t
	if m.mode == ui.ModeAnnotating {
		return handleAnnotatingKey(m, msg)
	}

	// Overview mode has its own q behavior (return to reading, not quit)
	if m.mode == ui.ModeOverview {
		return handleOverviewKey(m, msg)
	}

	// Global keys for non-annotating, non-overview modes
	switch msg.String() {
	case "ctrl+t":
		names := theme.Names()
		m.themeIndex = (m.themeIndex + 1) % len(names)
		m.theme = theme.Get(names[m.themeIndex])
		return m, nil
	case "q":
		return m, tea.Quit
	}

	switch m.mode {
	case ui.ModeReading:
		return handleReadingKey(m, msg)
	case ui.ModeSelecting:
		return handleSelectingKey(m, msg)
	}
	return m, nil
}

func handleReadingKey(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalLines := len(m.file.RenderedLines)
	maxOffset := max(0, totalLines-m.viewport.ViewportHeight)

	switch msg.String() {
	case "up", "k":
		m.viewport.ScrollOffset = clamp(m.viewport.ScrollOffset-1, 0, maxOffset)
	case "down", "j":
		m.viewport.ScrollOffset = clamp(m.viewport.ScrollOffset+1, 0, maxOffset)
	case "pgup", "ctrl+u":
		half := m.viewport.ViewportHeight / 2
		m.viewport.ScrollOffset = clamp(m.viewport.ScrollOffset-half, 0, maxOffset)
	case "pgdown", "ctrl+d":
		half := m.viewport.ViewportHeight / 2
		m.viewport.ScrollOffset = clamp(m.viewport.ScrollOffset+half, 0, maxOffset)
	case "home", "g":
		m.viewport.ScrollOffset = 0
	case "end", "G":
		m.viewport.ScrollOffset = maxOffset
	case "v":
		contentLines := len(m.file.StrippedLines)
		if contentLines == 0 {
			return m, nil
		}
		// Start selection at center of viewport
		centerLineIdx := m.viewport.ScrollOffset + m.viewport.ViewportHeight/2
		centerLineIdx = clamp(centerLineIdx, 0, contentLines-1)
		centerLine := centerLineIdx + 1 // 1-based
		m.selection = selectionState{
			Active: true,
			Start:  annotation.SelectionPos{Line: centerLine, Col: 0},
			End:    annotation.SelectionPos{Line: centerLine, Col: lineLength(m, centerLine)},
		}
		m.mode = ui.ModeSelecting
	case "d":
		if len(m.annotations) > 0 {
			m.overview = overviewState{}
			m.mode = ui.ModeOverview
		}
	}

	return m, nil
}

func handleSelectingKey(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalLines := len(m.file.StrippedLines)
	maxOffset := max(0, len(m.file.RenderedLines)-m.viewport.ViewportHeight)

	switch msg.String() {
	case "up", "k":
		newLine := clamp(m.selection.End.Line-1, 1, totalLines)
		m.selection.End.Line = newLine
		m.selection.End.Col = clamp(m.selection.End.Col, 0, lineLength(m, newLine))
		autoScrollToLine(&m, newLine)
	case "down", "j":
		newLine := clamp(m.selection.End.Line+1, 1, totalLines)
		m.selection.End.Line = newLine
		m.selection.End.Col = clamp(m.selection.End.Col, 0, lineLength(m, newLine))
		autoScrollToLine(&m, newLine)
	case "left", "h":
		newCol := clamp(m.selection.End.Col-1, 0, lineLength(m, m.selection.End.Line))
		m.selection.End.Col = newCol
	case "right", "l":
		newCol := clamp(m.selection.End.Col+1, 0, lineLength(m, m.selection.End.Line))
		m.selection.End.Col = newCol
	case "pgup", "ctrl+u":
		half := m.viewport.ViewportHeight / 2
		m.viewport.ScrollOffset = clamp(m.viewport.ScrollOffset-half, 0, maxOffset)
	case "pgdown", "ctrl+d":
		half := m.viewport.ViewportHeight / 2
		m.viewport.ScrollOffset = clamp(m.viewport.ScrollOffset+half, 0, maxOffset)
	case "enter", "a":
		m.mode = ui.ModeAnnotating
		m.input = inputState{}
	case "esc":
		m.selection = selectionState{}
		m.mode = ui.ModeReading
	}

	return m, nil
}

func handleOverviewKey(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "ctrl+t":
		names := theme.Names()
		m.themeIndex = (m.themeIndex + 1) % len(names)
		m.theme = theme.Get(names[m.themeIndex])
		return m, nil
	case "up", "k":
		if m.overview.Cursor > 0 {
			m.overview.Cursor--
		}
		// Auto-scroll viewport to annotation position
		if m.overview.Cursor < len(m.resolved) {
			autoScrollToLine(&m, m.resolved[m.overview.Cursor].StartLine)
		}
	case "down", "j":
		if m.overview.Cursor < len(m.resolved)-1 {
			m.overview.Cursor++
		}
		if m.overview.Cursor < len(m.resolved) {
			autoScrollToLine(&m, m.resolved[m.overview.Cursor].StartLine)
		}
	case "enter", "e":
		m = enterEditFromOverview(m)
		return m, nil
	case "backspace", "delete", "x":
		return deleteAnnotation(m)
	case "esc", "q":
		m.overview = overviewState{}
		m.mode = ui.ModeReading
	}
	return m, nil
}

func handleAnnotatingKey(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(m.input.Value) > 0 {
			return submitAnnotation(m)
		}
		return m, nil

	case "ctrl+j":
		// Insert newline
		m.input.Value = insertRune(m.input.Value, m.input.Cursor, '\n')
		m.input.Cursor++

	case "esc":
		// If editing from overview, go back to overview; otherwise to reading
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

	case "backspace":
		if m.input.Cursor > 0 {
			m.input.Value = deleteRune(m.input.Value, m.input.Cursor-1)
			m.input.Cursor--
		}

	case "delete":
		if m.input.Cursor < len(m.input.Value) {
			m.input.Value = deleteRune(m.input.Value, m.input.Cursor)
		}

	case "left":
		if m.input.Cursor > 0 {
			m.input.Cursor--
		}

	case "right":
		if m.input.Cursor < len(m.input.Value) {
			m.input.Cursor++
		}

	case "home", "ctrl+a":
		m.input.Cursor = 0

	case "end", "ctrl+e":
		m.input.Cursor = len(m.input.Value)

	default:
		// Insert printable runes
		if msg.Type == tea.KeyRunes {
			for _, r := range msg.Runes {
				if unicode.IsPrint(r) {
					m.input.Value = insertRune(m.input.Value, m.input.Cursor, r)
					m.input.Cursor++
				}
			}
		}
	}

	return m, nil
}

// insertRune inserts r at position pos in the rune slice.
func insertRune(runes []rune, pos int, r rune) []rune {
	if pos < 0 {
		pos = 0
	}
	if pos > len(runes) {
		pos = len(runes)
	}
	result := make([]rune, len(runes)+1)
	copy(result, runes[:pos])
	result[pos] = r
	copy(result[pos+1:], runes[pos:])
	return result
}

// deleteRune removes the rune at position pos from the slice.
func deleteRune(runes []rune, pos int) []rune {
	if pos < 0 || pos >= len(runes) {
		return runes
	}
	result := make([]rune, len(runes)-1)
	copy(result, runes[:pos])
	copy(result[pos:], runes[pos+1:])
	return result
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func clampScroll(offset, totalLines, viewportHeight int) int {
	return clamp(offset, 0, max(0, totalLines-viewportHeight))
}
