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
		if !m.loaded {
			m.loaded = true
			return m, loadFileCmd(m.file.FilePath, msg.Width, loadInitial)
		}
		if prevWidth != msg.Width && m.file.FilePath != "" {
			return m, loadFileCmd(m.file.FilePath, msg.Width, loadResize)
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

		switch msg.source {
		case loadInitial:
			// First load: reset everything
			m.selection = selectionState{}
			m.input = inputState{}
			m.editingID = ""
			m.overview = overviewState{}
			m.mode = ui.ModeReading
		case loadResize, loadIPC:
			// Only update content, preserve user state
		case loadFileChange:
			// File changed on disk: reset if reading/selecting, preserve if annotating/overview
			if m.mode == ui.ModeReading || m.mode == ui.ModeSelecting {
				m.selection = selectionState{}
				m.input = inputState{}
				m.editingID = ""
				m.overview = overviewState{}
				m.mode = ui.ModeReading
			} else if m.mode == ui.ModeOverview {
				// Stay in overview, clamp cursor
				if m.overview.Cursor >= len(m.resolved) && m.overview.Cursor > 0 {
					m.overview.Cursor = len(m.resolved) - 1
				}
				if len(m.resolved) == 0 {
					m.overview = overviewState{}
					m.mode = ui.ModeReading
				}
			}
			// ModeAnnotating: keep input, selection, and mode as-is
		}
		return m, nil

	case fileChangedMsg:
		return m, tea.Batch(
			loadFileCmd(m.file.FilePath, m.viewport.Width, loadFileChange),
			watchFileCmd(m.file.FilePath),
		)

	case ipcRequestMsg:
		return handleIpc(m, msg.req)

	}

	return m, nil
}
