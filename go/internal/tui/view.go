package tui

import "github.com/aporicho/markcli/internal/tui/ui"

func (m Model) View() string {
	viewer := ui.RenderViewer(
		m.file.RenderedLines,
		m.file.StrippedLines,
		m.file.LineLengths,
		m.resolved,
		m.viewport.ScrollOffset,
		m.viewport.ViewportHeight,
		m.theme,
		nil, // selRange: no selection in reading mode
	)
	bar := ui.RenderStatusbar(
		m.mode,
		m.viewport.ScrollOffset,
		len(m.file.RenderedLines),
		len(m.annotations),
		m.viewport.Width,
		m.theme,
	)
	return viewer + bar
}
