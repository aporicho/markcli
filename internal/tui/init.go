package tui

import (
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/markcli/internal/tui/ui"
)

// setPointerTextCmd sets the mouse pointer to an I-beam (text cursor) via OSC 22.
// This overrides the default arrow cursor that terminals show when mouse tracking is enabled.
func setPointerTextCmd() tea.Cmd {
	return func() tea.Msg {
		os.Stdout.WriteString("\x1b]22;text\x1b\\")
		return nil
	}
}

// ResetPointer resets the mouse pointer to the terminal default.
// Call this after the Bubbletea program exits.
func ResetPointer() {
	os.Stdout.WriteString("\x1b]22;\x1b\\")
}

func (m Model) Init() tea.Cmd {
	if m.mode == ui.ModeBrowsing {
		return tea.Batch(
			scanFilesCmd(m.browsing.Dir),
			setPointerTextCmd(),
		)
	}
	return tea.Batch(
		watchFileCmd(m.file.FilePath),
		waitIpcCmd(m.ipcCh),
		setPointerTextCmd(),
	)
}
