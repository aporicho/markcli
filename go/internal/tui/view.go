package tui

import (
	"strings"

	"github.com/aporicho/markcli/internal/tui/ui"
)

func (m Model) View() string {
	var selRange *[4]int
	if m.selection.Active {
		s, e := ui.NormalizePos(m.selection.Start, m.selection.End)
		selRange = &[4]int{s.Line, s.Col, e.Line, e.Col}
	}

	viewer := ui.RenderViewer(
		m.file.RenderedLines,
		m.file.StrippedLines,
		m.file.LineLengths,
		m.resolved,
		m.viewport.ScrollOffset,
		m.viewport.ViewportHeight,
		m.theme,
		selRange,
	)

	if m.mode == ui.ModeAnnotating && m.selection.Active {
		panelTop := calcInputPanelTop(m)
		panelWidth := m.viewport.Width - 4
		if panelWidth < 10 {
			panelWidth = 10
		}
		panel := ui.RenderInputPanel(string(m.input.Value), m.input.Cursor, panelWidth, m.theme)
		viewer = ui.OverlayAt(viewer, panel, panelTop, 2)
	}

	if m.mode == ui.ModeOverview {
		panelHeight := min(len(m.resolved)+3, m.viewport.ViewportHeight-2) // +3 for border+title
		if panelHeight < 5 {
			panelHeight = 5
		}
		panelWidth := m.viewport.Width - 4
		if panelWidth < 10 {
			panelWidth = 10
		}
		panelTop := max(0, (m.viewport.ViewportHeight-panelHeight)/2)
		panel := ui.RenderOverviewPanel(m.resolved, m.overview.Cursor, panelWidth, panelHeight, m.theme)
		viewer = ui.OverlayAt(viewer, panel, panelTop, 2)
	}

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

// calcInputPanelTop determines the vertical position for the input panel.
// Places it below the selection end if space allows, otherwise above the selection start.
func calcInputPanelTop(m Model) int {
	s, e := ui.NormalizePos(m.selection.Start, m.selection.End)

	selEndRow := e.Line - 1 - m.viewport.ScrollOffset // 0-based viewport row
	selStartRow := s.Line - 1 - m.viewport.ScrollOffset

	inputLines := max(1, strings.Count(string(m.input.Value), "\n")+1)
	panelHeight := 2 + 1 + inputLines // border(2) + title(1) + content lines

	// Try below selection
	if selEndRow+1+panelHeight <= m.viewport.ViewportHeight {
		return selEndRow + 1
	}

	// Try above selection
	if selStartRow-panelHeight >= 0 {
		return selStartRow - panelHeight
	}

	// Fallback: just below selection end, clamp to viewport
	top := selEndRow + 1
	if top+panelHeight > m.viewport.ViewportHeight {
		top = m.viewport.ViewportHeight - panelHeight
	}
	if top < 0 {
		top = 0
	}
	return top
}
