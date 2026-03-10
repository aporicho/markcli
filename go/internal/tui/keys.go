package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/markcli/internal/theme"
	"github.com/aporicho/markcli/internal/tui/ui"
)

func handleKey(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ui.ModeReading:
		return handleReadingKey(m, msg)
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
	case "ctrl+t":
		names := theme.Names()
		m.themeIndex = (m.themeIndex + 1) % len(names)
		m.theme = theme.Get(names[m.themeIndex])
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
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
