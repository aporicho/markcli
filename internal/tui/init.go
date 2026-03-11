package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		watchFileCmd(m.file.FilePath),
		waitIpcCmd(m.ipcCh),
	)
}
