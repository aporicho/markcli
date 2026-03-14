package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/theme"
)

// RenderOverviewPanel renders a floating panel listing all annotations.
func RenderOverviewPanel(
	annotations []annotation.Annotation,
	cursor int,
	panelWidth int,
	maxHeight int,
	t theme.Theme,
) string {
	if panelWidth < 10 {
		panelWidth = 10
	}
	innerWidth := panelWidth - 4 // border (2) + padding (2)
	if innerWidth < 6 {
		innerWidth = 6
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.PanelBorder())).
		Background(lipgloss.Color(t.PanelBg())).
		Width(innerWidth).
		PaddingLeft(1).
		PaddingRight(1)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.PanelAccent())).
		Background(lipgloss.Color(t.PanelBg())).
		Bold(true)

	title := titleStyle.Render("批注总览")

	if len(annotations) == 0 {
		dimStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(t.PanelBorder())).
			Background(lipgloss.Color(t.PanelBg())).
			Faint(true)
		body := title + "\n" + dimStyle.Render("暂无批注")
		return borderStyle.Render(body)
	}

	// Calculate visible range for scrolling
	// maxHeight includes border (2 top/bottom) + title (1), so list area = maxHeight - 3
	listHeight := maxHeight - 3
	if listHeight < 1 {
		listHeight = 1
	}

	scrollOffset := 0
	if len(annotations) > listHeight {
		// Keep cursor centered
		scrollOffset = cursor - listHeight/2
		if scrollOffset < 0 {
			scrollOffset = 0
		}
		if scrollOffset > len(annotations)-listHeight {
			scrollOffset = len(annotations) - listHeight
		}
	}

	visibleEnd := scrollOffset + listHeight
	if visibleEnd > len(annotations) {
		visibleEnd = len(annotations)
	}

	normalStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(t.PanelBg()))

	accentStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.PanelAccent())).
		Background(lipgloss.Color(t.PanelBg()))

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(t.PanelBorder())).
		Background(lipgloss.Color(t.PanelBg())).
		Faint(true)

	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(t.PanelAccent())).
		Foreground(lipgloss.Color(t.PanelBg()))

	var lines []string
	for i := scrollOffset; i < visibleEnd; i++ {
		ann := annotations[i]
		isCursor := i == cursor
		isResolved := annotation.IsResolved(ann)

		// Prefix
		prefix := "  "
		if isCursor {
			prefix = "▸ "
		} else if isResolved {
			prefix = "✓ "
		}

		// Line badge: "L23" or "L23-25"
		var badge string
		if ann.StartLine == ann.EndLine {
			badge = fmt.Sprintf("L%d", ann.StartLine)
		} else {
			badge = fmt.Sprintf("L%d-%d", ann.StartLine, ann.EndLine)
		}
		// Pad to 9 chars
		for len(badge) < 9 {
			badge = " " + badge
		}

		// Text preview: first 25 chars of selectedText, replace \n with ↵
		textPreview := truncatePreview(ann.SelectedText, 25)
		// Comment preview
		commentPreview := truncatePreview(ann.Comment, 25)

		rest := " " + textPreview + " → " + commentPreview

		// Apply styles
		if isCursor {
			lineContent := prefix + badge + rest
			lines = append(lines, cursorStyle.Render(padOrTruncate(lineContent, innerWidth)))
		} else if isResolved {
			lineContent := prefix + badge + rest
			lines = append(lines, dimStyle.Render(padOrTruncate(lineContent, innerWidth)))
		} else {
			// Badge in accent color, rest in normal
			restPadded := padOrTruncate(rest, innerWidth-runewidth.StringWidth(prefix)-runewidth.StringWidth(badge))
			lines = append(lines, normalStyle.Render(prefix)+accentStyle.Render(badge)+normalStyle.Render(restPadded))
		}
	}

	body := title + "\n" + strings.Join(lines, "\n")
	return borderStyle.Render(body)
}

// truncatePreview returns the first maxChars characters of s,
// replacing newlines with ↵ and appending … if truncated.
func truncatePreview(s string, maxChars int) string {
	s = strings.ReplaceAll(s, "\n", "↵")
	runes := []rune(s)
	if len(runes) > maxChars {
		return string(runes[:maxChars]) + "…"
	}
	return string(runes)
}

// padOrTruncate ensures s is exactly width display columns (CJK-aware).
func padOrTruncate(s string, width int) string {
	sw := runewidth.StringWidth(s)
	if sw >= width {
		return runewidth.Truncate(s, width, "")
	}
	return s + strings.Repeat(" ", width-sw)
}
