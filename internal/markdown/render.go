package markdown

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
)

var (
	rendererMu   sync.Mutex
	lastWidth    = -1
	lastRenderer *glamour.TermRenderer
)

func getRenderer(width int) (*glamour.TermRenderer, error) {
	rendererMu.Lock()
	defer rendererMu.Unlock()
	if lastRenderer != nil && lastWidth == width {
		return lastRenderer, nil
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStylePath("dark"),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}
	lastWidth = width
	lastRenderer = r
	return r, nil
}

// Render renders Markdown content to an ANSI string.
// width <= 0 means no word-wrap limit.
func Render(content string, width int) (string, error) {
	if width < 0 {
		width = 0
	}
	r, err := getRenderer(width)
	if err != nil {
		return "", err
	}
	return r.Render(content)
}

// RenderToLines renders Markdown and splits into lines, trimming trailing empty lines.
func RenderToLines(content string, width int) ([]string, error) {
	rendered, err := Render(content, width)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(rendered, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return lines, nil
}
