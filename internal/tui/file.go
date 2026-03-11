package tui

import (
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fsnotify/fsnotify"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/ansi"
	"github.com/aporicho/markcli/internal/markdown"
)

// fileLoadedMsg is sent when file content has been loaded and rendered.
type fileLoadedMsg struct {
	renderedLines []string
	strippedLines []string
	lineLengths   []int
	annotations   []annotation.Annotation
}

// fileChangedMsg is sent when the watched file is modified on disk.
type fileChangedMsg struct{ path string }

// loadFileCmd reads the file, renders Markdown, strips ANSI, and loads annotations.
func loadFileCmd(path string, width int) tea.Cmd {
	return func() tea.Msg {
		content, err := os.ReadFile(path)
		if err != nil {
			return fileLoadedMsg{}
		}

		renderedLines, err := markdown.RenderToLines(string(content), width)
		if err != nil {
			return fileLoadedMsg{}
		}

		strippedLines := make([]string, len(renderedLines))
		lineLengths := make([]int, len(renderedLines))
		for i, line := range renderedLines {
			stripped := ansi.StripAnsi(line)
			strippedLines[i] = stripped
			lineLengths[i] = len([]rune(stripped))
		}

		af, _ := annotation.Load(path)

		return fileLoadedMsg{
			renderedLines: renderedLines,
			strippedLines: strippedLines,
			lineLengths:   lineLengths,
			annotations:   af.Annotations,
		}
	}
}

// watchFileCmd watches for Write events on path, debounces 100ms, and returns fileChangedMsg.
// Call it again from Update to re-arm after each event.
func watchFileCmd(path string) tea.Cmd {
	return func() tea.Msg {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return nil
		}
		defer w.Close()

		dir := filepath.Dir(path)
		base := filepath.Base(path)
		if err := w.Add(dir); err != nil {
			return nil
		}

		for {
			select {
			case event, ok := <-w.Events:
				if !ok {
					return nil
				}
				if filepath.Base(event.Name) == base && event.Op&fsnotify.Write != 0 {
					// Debounce: drain additional events for 100ms
					timer := time.NewTimer(100 * time.Millisecond)
				draining:
					for {
						select {
						case <-w.Events:
							if !timer.Stop() {
								<-timer.C
							}
							timer.Reset(100 * time.Millisecond)
						case <-timer.C:
							break draining
						}
					}
					return fileChangedMsg{path: path}
				}
			case _, ok := <-w.Errors:
				if !ok {
					return nil
				}
			}
		}
	}
}
