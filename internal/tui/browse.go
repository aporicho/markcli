package tui

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/markcli/internal/tui/ui"
)

// filesScannedMsg is sent when directory scanning completes.
type filesScannedMsg struct {
	Entries []fileEntry
	Dir     string
}

// scanFilesCmd scans dir for directories and .md files.
func scanFilesCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		dirEntries, err := os.ReadDir(dir)
		if err != nil {
			return filesScannedMsg{Dir: dir}
		}

		var dirs, files []fileEntry
		for _, e := range dirEntries {
			name := e.Name()
			// Skip hidden files/dirs
			if strings.HasPrefix(name, ".") {
				continue
			}

			info, err := e.Info()
			if err != nil {
				continue
			}

			entry := fileEntry{
				Name:    name,
				Path:    filepath.Join(dir, name),
				Size:    info.Size(),
				ModTime: info.ModTime(),
				IsDir:   e.IsDir(),
			}

			if e.IsDir() {
				dirs = append(dirs, entry)
			} else if strings.HasSuffix(strings.ToLower(name), ".md") {
				files = append(files, entry)
			}
		}

		sort.Slice(dirs, func(i, j int) bool {
			return dirs[i].Name < dirs[j].Name
		})
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name < files[j].Name
		})

		entries := append(dirs, files...)
		return filesScannedMsg{Entries: entries, Dir: dir}
	}
}

// handleBrowsingKey handles keyboard input in browsing mode.
func handleBrowsingKey(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		if m.file.FilePath != "" {
			m.mode = ui.ModeReading
			return m, nil
		}
		return m, tea.Quit
	case "q":
		return m, tea.Quit

	case "up", "k":
		if m.browsing.Cursor > 0 {
			m.browsing.Cursor--
		}
		return m, nil

	case "down", "j":
		if m.browsing.Cursor < len(m.browsing.Entries)-1 {
			m.browsing.Cursor++
		}
		return m, nil

	case "enter", "l", "right":
		if len(m.browsing.Entries) == 0 {
			return m, nil
		}
		entry := m.browsing.Entries[m.browsing.Cursor]
		if entry.IsDir {
			m.browsing.Dir = entry.Path
			m.browsing.Cursor = 0
			m.browsing.Entries = nil
			return m, scanFilesCmd(entry.Path)
		}
		return openFileFromBrowser(m, entry.Path)

	case "backspace", "h", "left":
		parent := filepath.Dir(m.browsing.Dir)
		if parent == m.browsing.Dir {
			// Already at root
			return m, nil
		}
		m.browsing.Dir = parent
		m.browsing.Cursor = 0
		m.browsing.Entries = nil
		return m, scanFilesCmd(parent)
	}

	return m, nil
}

// enterBrowsing switches from reading to browsing mode, starting from the current file's directory.
func enterBrowsing(m Model) (Model, tea.Cmd) {
	dir := filepath.Dir(m.file.FilePath)
	m.mode = ui.ModeBrowsing
	m.browsing = browsingState{Dir: dir}
	return m, scanFilesCmd(dir)
}

// openFileFromBrowser transitions from browsing to reading mode for the given file.
func openFileFromBrowser(m Model, path string) (Model, tea.Cmd) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		m.errText = err.Error()
		return m, nil
	}

	m.file.FilePath = absPath
	m.mode = ui.ModeReading
	m.selection = selectionState{}
	m.input = inputState{}
	m.editingID = ""
	m.overview = overviewState{}
	m.viewport.ScrollOffset = 0

	return m, tea.Batch(
		loadFileCmd(absPath, m.viewport.Width, loadInitial),
		watchFileCmd(absPath),
		waitIpcCmd(m.ipcCh),
	)
}

// formatSize formats a file size in human-readable form.
func formatSize(size int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
	)
	switch {
	case size >= mb:
		return formatFloat(float64(size)/float64(mb)) + " MB"
	case size >= kb:
		return formatFloat(float64(size)/float64(kb)) + " KB"
	default:
		return formatInt(size) + " B"
	}
}

func formatFloat(f float64) string {
	if f >= 100 {
		return formatInt(int64(f))
	}
	// One decimal place
	i := int64(f * 10)
	whole := i / 10
	frac := i % 10
	if frac == 0 {
		return formatInt(whole)
	}
	return formatInt(whole) + "." + formatInt(frac)
}

func formatInt(n int64) string {
	if n < 0 {
		return "-" + formatInt(-n)
	}
	s := ""
	for n >= 10 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return string(rune('0'+n)) + s
}

// formatDate formats a time as "M月D日".
func formatDate(t time.Time) string {
	return formatInt(int64(t.Month())) + "月" + formatInt(int64(t.Day())) + "日"
}
