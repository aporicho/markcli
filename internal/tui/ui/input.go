package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"

	"github.com/aporicho/markcli/internal/theme"
)

// RenderInputPanel renders a minimal floating input panel — just a bordered box with cursor.
func RenderInputPanel(value string, cursorPos int, panelWidth int, t theme.Theme) string {
	if panelWidth < 6 {
		panelWidth = 6
	}
	innerWidth := panelWidth - 4 // border (2) + padding (2)
	if innerWidth < 2 {
		innerWidth = 2
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(t.PanelBorder())).
		Background(lipgloss.Color(t.PanelBg())).
		Width(innerWidth).
		PaddingLeft(1).
		PaddingRight(1)

	var content string
	if value == "" {
		cursorStyle := lipgloss.NewStyle().
			Background(lipgloss.Color(t.PanelAccent())).
			Foreground(lipgloss.Color(t.PanelBg()))
		content = cursorStyle.Render(" ")
	} else {
		content = renderInputWithCursor(value, cursorPos, t)
	}

	return borderStyle.Render(content)
}

// renderInputWithCursor renders input text with a block cursor at cursorPos.
func renderInputWithCursor(value string, cursorPos int, t theme.Theme) string {
	runes := []rune(value)
	if cursorPos < 0 {
		cursorPos = 0
	}
	if cursorPos > len(runes) {
		cursorPos = len(runes)
	}

	cursorStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(t.PanelAccent())).
		Foreground(lipgloss.Color(t.PanelBg()))

	textStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(t.PanelBg()))

	var sb strings.Builder
	for i, r := range runes {
		if i == cursorPos {
			sb.WriteString(cursorStyle.Render(string(r)))
		} else {
			sb.WriteString(textStyle.Render(string(r)))
		}
	}
	// Cursor at end: render block space
	if cursorPos >= len(runes) {
		sb.WriteString(cursorStyle.Render(" "))
	}

	return sb.String()
}

// OverlayAt places overlay on top of base at the given (top, left) position.
// Both base and overlay are newline-separated strings.
func OverlayAt(base, overlay string, top, left int) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	for i, oLine := range overlayLines {
		row := top + i
		if row < 0 || row >= len(baseLines) {
			continue
		}
		baseLines[row] = overlayLine(baseLines[row], oLine, left)
	}

	return strings.Join(baseLines, "\n")
}

// overlayLine replaces part of a base line with an overlay line starting at column left.
// ANSI-aware: escape sequences in base are passed through without affecting column tracking.
func overlayLine(base, overlay string, left int) string {
	overlayWidth := lipgloss.Width(overlay)

	var sb strings.Builder
	col := 0
	i := 0
	baseRunes := []rune(base)

	// Write base until we reach the left column (ANSI-aware)
	for i < len(baseRunes) && col < left {
		if n := consumeAnsi(baseRunes[i:]); n > 0 {
			for j := 0; j < n; j++ {
				sb.WriteRune(baseRunes[i+j])
			}
			i += n
			continue
		}
		w := runewidth.RuneWidth(baseRunes[i])
		sb.WriteRune(baseRunes[i])
		col += w
		i++
	}

	// Pad if base is shorter than left
	for col < left {
		sb.WriteRune(' ')
		col++
	}

	// Reset any active base styling before overlay
	sb.WriteString("\x1b[0m")
	// Write overlay content
	sb.WriteString(overlay)

	// Skip base runes covered by overlay (ANSI-aware)
	skipCols := overlayWidth
	for i < len(baseRunes) && skipCols > 0 {
		if n := consumeAnsi(baseRunes[i:]); n > 0 {
			i += n // drop covered ANSI sequences
			continue
		}
		w := runewidth.RuneWidth(baseRunes[i])
		skipCols -= w
		i++
	}

	// Append remaining base runes (preserves trailing ANSI resets)
	for i < len(baseRunes) {
		sb.WriteRune(baseRunes[i])
		i++
	}

	return sb.String()
}

// consumeAnsi returns the length (in runes) of an ANSI SGR escape sequence
// starting at runes[0], or 0 if not an escape sequence.
// Matches the pattern: ESC '[' (digits/';')* letter
func consumeAnsi(runes []rune) int {
	if len(runes) < 3 || runes[0] != '\x1b' || runes[1] != '[' {
		return 0
	}
	i := 2
	for i < len(runes) && ((runes[i] >= '0' && runes[i] <= '9') || runes[i] == ';') {
		i++
	}
	if i < len(runes) && runes[i] >= '@' && runes[i] <= '~' {
		return i + 1 // include the terminating letter
	}
	return 0
}
