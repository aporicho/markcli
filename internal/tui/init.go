package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/markcli/internal/tui/ui"
)

func (m Model) Init() tea.Cmd {
	if m.mode == ui.ModeBrowsing {
		return scanFilesCmd(m.browsing.Dir)
	}
	return tea.Batch(
		watchFileCmd(m.file.FilePath),
		waitIpcCmd(m.ipcCh),
	)
}
