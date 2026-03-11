package tui

import tea "github.com/charmbracelet/bubbletea"

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadFileCmd(m.file.FilePath, m.viewport.Width),
		watchFileCmd(m.file.FilePath),
		waitIpcCmd(m.ipcCh),
	)
}
