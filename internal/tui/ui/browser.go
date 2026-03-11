package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/aporicho/markcli/internal/theme"
)

// BrowserEntry holds display data for one file/directory in the browser.
type BrowserEntry struct {
	Name    string
	IsDir   bool
	SizeStr string // pre-formatted, e.g. "12.3 KB"
	DateStr string // pre-formatted, e.g. "3月11日"
}

// RenderBrowser renders a full-screen file browser.
func RenderBrowser(
	dir string,
	entries []BrowserEntry,
	cursor int,
	viewportWidth int,
	viewportHeight int,
	t theme.Theme,
) string {
	var sb strings.Builder

	// Header: directory path
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.Panel.Accent)).
		Bold(true)
	header := headerStyle.Render(" \U000F094B " + dir)
	sb.WriteString(header)
	sb.WriteString("\n\n")

	linesUsed := 2 // header + blank line
	listHeight := viewportHeight - linesUsed
	if listHeight < 1 {
		listHeight = 1
	}

	// Calculate scroll offset to keep cursor visible
	scrollOffset := 0
	if len(entries) > listHeight {
		scrollOffset = cursor - listHeight/2
		if scrollOffset < 0 {
			scrollOffset = 0
		}
		if scrollOffset > len(entries)-listHeight {
			scrollOffset = len(entries) - listHeight
		}
	}

	visibleEnd := scrollOffset + listHeight
	if visibleEnd > len(entries) {
		visibleEnd = len(entries)
	}

	normalStyle := lipgloss.NewStyle()
	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(t.Panel.Accent)).
		Foreground(lipgloss.Color(t.Panel.Bg))
	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.StatusBar.DimBg)).
		Faint(true)

	// Calculate max name width for alignment
	metaWidth := 22 // " 12.3 KB  3月11日" approx
	nameWidth := viewportWidth - 5 - metaWidth // 5 = prefix(3) + icon(1) + space(1)
	if nameWidth < 10 {
		nameWidth = 10
	}

	linesRendered := 0
	for i := scrollOffset; i < visibleEnd; i++ {
		e := entries[i]
		isCursor := i == cursor

		prefix := "   "
		if isCursor {
			prefix = " \u25B8 "
		}

		icon := "\U000F0354 "
		if e.IsDir {
			icon = "\U000F0956 "
		}

		name := e.Name
		if e.IsDir {
			name += "/"
		}

		// Truncate name if needed
		nameW := runewidth.StringWidth(name)
		if nameW > nameWidth {
			name = runewidth.Truncate(name, nameWidth-1, "\u2026")
			nameW = runewidth.StringWidth(name)
		}
		namePad := strings.Repeat(" ", max(0, nameWidth-nameW))

		var meta string
		if e.IsDir {
			meta = ""
		} else {
			sizePad := strings.Repeat(" ", max(0, 10-runewidth.StringWidth(e.SizeStr)))
			meta = sizePad + e.SizeStr + "  " + e.DateStr
		}

		line := prefix + icon + name + namePad + meta

		if isCursor {
			sb.WriteString(cursorStyle.Render(padOrTruncate(line, viewportWidth)))
		} else if e.IsDir {
			sb.WriteString(normalStyle.Render(line))
		} else {
			// Name in normal, meta in dim
			namepart := prefix + icon + name + namePad
			sb.WriteString(normalStyle.Render(namepart))
			sb.WriteString(dimStyle.Render(meta))
		}
		sb.WriteString("\n")
		linesRendered++
	}

	// If no entries, show empty message
	if len(entries) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.StatusBar.DimBg)).
			Faint(true)
		sb.WriteString(emptyStyle.Render("   （无 .md 文件）"))
		sb.WriteString("\n")
		linesRendered++
	}

	// Pad remaining lines
	for linesRendered < listHeight {
		sb.WriteString("\n")
		linesRendered++
	}

	return sb.String()
}
