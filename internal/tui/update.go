package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/markcli/internal/tui/ui"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		prevWidth := m.viewport.Width
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height
		m.viewport.ViewportHeight = msg.Height - 1
		m.viewport.ScrollOffset = clampScroll(m.viewport.ScrollOffset, len(m.file.RenderedLines), m.viewport.ViewportHeight)
		if prevWidth != msg.Width && m.file.FilePath != "" {
			return m, loadFileCmd(m.file.FilePath, msg.Width)
		}
		return m, nil

	case tea.KeyMsg:
		m.errText = ""
		return handleKey(m, msg)

	case tea.MouseMsg:
		m.errText = ""
		return handleMouse(m, msg)

	case errMsg:
		m.errText = msg.err.Error()
		return m, nil

	case fileLoadedMsg:
		m.file.RenderedLines = msg.renderedLines
		m.file.StrippedLines = msg.strippedLines
		m.file.LineLengths = msg.lineLengths
		m.annotations = msg.annotations
		m.resolved = resolveAnnotations(m.file.StrippedLines, m.file.LineLengths, m.annotations)
		m.viewport.ScrollOffset = clampScroll(m.viewport.ScrollOffset, len(m.file.RenderedLines), m.viewport.ViewportHeight)
		// Reset selection, input, editingID — file content changed, old positions invalid
		m.selection = selectionState{}
		m.input = inputState{}
		m.editingID = ""
		if m.mode == ui.ModeOverview {
			// Stay in overview, clamp cursor to new annotation count
			if m.overview.Cursor >= len(m.resolved) && m.overview.Cursor > 0 {
				m.overview.Cursor = len(m.resolved) - 1
			}
			if len(m.resolved) == 0 {
				m.overview = overviewState{}
				m.mode = ui.ModeReading
			}
		} else {
			m.overview = overviewState{}
			m.mode = ui.ModeReading
		}
		return m, nil

	case fileChangedMsg:
		return m, tea.Batch(
			loadFileCmd(m.file.FilePath, m.viewport.Width),
			watchFileCmd(m.file.FilePath),
		)

	case ipcRequestMsg:
		return handleIpc(m, msg.req)

	}

	return m, nil
}
