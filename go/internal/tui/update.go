package tui

import tea "github.com/charmbracelet/bubbletea"

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
		return handleKey(m, msg)

	case fileLoadedMsg:
		m.file.RenderedLines = msg.renderedLines
		m.file.StrippedLines = msg.strippedLines
		m.file.LineLengths = msg.lineLengths
		m.annotations = msg.annotations
		m.resolved = resolveAnnotations(m.file.StrippedLines, m.file.LineLengths, m.annotations)
		m.viewport.ScrollOffset = clampScroll(m.viewport.ScrollOffset, len(m.file.RenderedLines), m.viewport.ViewportHeight)
		return m, nil

	case fileChangedMsg:
		return m, tea.Batch(
			loadFileCmd(m.file.FilePath, m.viewport.Width),
			watchFileCmd(m.file.FilePath),
		)
	}

	return m, nil
}
