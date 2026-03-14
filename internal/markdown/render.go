package markdown

import (
	"strings"
	"sync"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/lipgloss"
)

var (
	rendererMu   sync.Mutex
	lastWidth    = -1
	lastDark     = true
	lastRenderer *glamour.TermRenderer
)

func boolPtr(b bool) *bool { return &b }
func strPtr(s string) *string { return &s }
func uintPtr(u uint) *uint { return &u }

// buildStyleConfig returns a glamour StyleConfig using ANSI 0-15 palette colors.
func buildStyleConfig() ansi.StyleConfig {
	return ansi.StyleConfig{
		Document: ansi.StyleBlock{
			Margin: uintPtr(2),
		},
		Heading: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Bold: boolPtr(true),
			},
		},
		H1: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				BackgroundColor: strPtr("4"),
				Bold:            boolPtr(true),
				Prefix:          " ",
				Suffix:          " ",
			},
		},
		H2: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr("4"),
				Bold:   boolPtr(true),
				Prefix: "## ",
			},
		},
		H3: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr("5"),
				Bold:   boolPtr(true),
				Prefix: "### ",
			},
		},
		H4: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr("2"),
				Bold:   boolPtr(true),
				Prefix: "#### ",
			},
		},
		H5: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr("8"),
				Bold:   boolPtr(true),
				Prefix: "##### ",
			},
		},
		H6: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color:  strPtr("8"),
				Prefix: "###### ",
			},
		},
		Paragraph: ansi.StyleBlock{},
		BlockQuote: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Italic: boolPtr(true),
			},
			Indent:      uintPtr(1),
			IndentToken: strPtr("│ "),
		},
		List: ansi.StyleList{
			LevelIndent: 2,
		},
		Link: ansi.StylePrimitive{
			Color:     strPtr("4"),
			Underline: boolPtr(true),
		},
		LinkText: ansi.StylePrimitive{
			Color: strPtr("4"),
			Bold:  boolPtr(true),
		},
		Image: ansi.StylePrimitive{
			Color:     strPtr("5"),
			Underline: boolPtr(true),
		},
		ImageText: ansi.StylePrimitive{
			Color: strPtr("5"),
		},
		Code: ansi.StyleBlock{
			StylePrimitive: ansi.StylePrimitive{
				Color: strPtr("2"),
			},
		},
		CodeBlock: ansi.StyleCodeBlock{
			StyleBlock: ansi.StyleBlock{
				Margin: uintPtr(2),
			},
		},
		HorizontalRule: ansi.StylePrimitive{
			Color:  strPtr("8"),
			Format: "─",
		},
		Emph: ansi.StylePrimitive{
			Italic: boolPtr(true),
		},
		Strong: ansi.StylePrimitive{
			Bold: boolPtr(true),
		},
		Strikethrough: ansi.StylePrimitive{
			CrossedOut: boolPtr(true),
		},
		Item: ansi.StylePrimitive{},
		Enumeration: ansi.StylePrimitive{},
		Table: ansi.StyleTable{
			StyleBlock: ansi.StyleBlock{
				StylePrimitive: ansi.StylePrimitive{},
			},
			CenterSeparator: strPtr("┼"),
			ColumnSeparator: strPtr("│"),
			RowSeparator:    strPtr("─"),
		},
		DefinitionTerm: ansi.StylePrimitive{
			Bold: boolPtr(true),
		},
		DefinitionDescription: ansi.StylePrimitive{
			BlockPrefix: "\n",
		},
	}
}

func getRenderer(width int) (*glamour.TermRenderer, error) {
	dark := lipgloss.HasDarkBackground()

	rendererMu.Lock()
	defer rendererMu.Unlock()
	if lastRenderer != nil && lastWidth == width && lastDark == dark {
		return lastRenderer, nil
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStyles(buildStyleConfig()),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return nil, err
	}
	lastWidth = width
	lastDark = dark
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
