package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aporicho/markcli/internal/theme"
)

// RenderStatusbar renders a single-line status bar.
func RenderStatusbar(
	mode AppMode,
	scrollOffset int,
	totalLines int,
	annotationCount int,
	termWidth int,
	t theme.Theme,
	errText string,
	isEditing bool,
) string {
	modeLabel, modeBg, hints := modeInfo(mode, t, isEditing)

	// Show error in statusbar if present
	if errText != "" {
		errStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#cc0000")).
			Foreground(lipgloss.Color("#ffffff")).
			Bold(true)
		errContent := fmt.Sprintf(" ✗ %s ", errText)
		errStr := errStyle.Render(errContent)
		padWidth := termWidth - lipgloss.Width(errStr)
		if padWidth < 0 {
			padWidth = 0
		}
		padStr := errStyle.Render(strings.Repeat(" ", padWidth))
		return errStr + padStr
	}

	modeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(modeBg)).
		Foreground(lipgloss.Color(t.StatusBar.ModeFg)).
		Bold(true)
	modeStr := modeStyle.Render(fmt.Sprintf(" %s ", modeLabel))

	var leftContent string
	if mode == ModeBrowsing {
		leftContent = fmt.Sprintf("  %d 项 ", totalLines)
	} else {
		currentLine := scrollOffset + 1
		leftContent = fmt.Sprintf("  %d 批注  %d/%d 行 ", annotationCount, currentLine, totalLines)
	}
	leftStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(t.StatusBar.Bg)).
		Foreground(lipgloss.Color(t.StatusBar.Fg))
	leftStr := leftStyle.Render(leftContent)

	rightContent := fmt.Sprintf(" %s ", hints)
	rightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(t.StatusBar.DimBg)).
		Foreground(lipgloss.Color(t.StatusBar.Fg))
	rightStr := rightStyle.Render(rightContent)

	// Pad the middle to fill terminal width
	usedWidth := lipgloss.Width(modeStr) + lipgloss.Width(leftStr) + lipgloss.Width(rightStr)
	padWidth := termWidth - usedWidth
	if padWidth < 0 {
		padWidth = 0
	}
	midStr := leftStyle.Render(strings.Repeat(" ", padWidth))

	return modeStr + leftStr + midStr + rightStr
}

func modeInfo(mode AppMode, t theme.Theme, isEditing bool) (label, bg, hints string) {
	switch mode {
	case ModeBrowsing:
		return "浏览", t.StatusBar.ModeBrowsing, "↑↓:选择  →/Enter:打开  ←:上级  Esc:返回  q:退出  ^T:主题"
	case ModeReading:
		return "阅读", t.StatusBar.ModeReading, "↑↓:滚动  b:浏览  d:总览  q:退出  ^T:主题"
	case ModeSelecting:
		return "选择", t.StatusBar.ModeSelecting, "↑↓:扩选  Enter:批注  Esc:取消"
	case ModeAnnotating:
		if isEditing {
			return "编辑", t.StatusBar.ModeAnnotating, "Enter:提交  ^J:换行  ^R:调整选区  ^D:删除  Esc:取消"
		}
		return "批注", t.StatusBar.ModeAnnotating, "Enter:提交  ^J:换行  ^R:调整选区  Esc:取消"
	case ModeOverview:
		return "概览", t.StatusBar.ModeOverview, "Enter:编辑  ⌫:删除  ↑↓:选择  Esc:返回  q:退出"
	default:
		return "阅读", t.StatusBar.ModeReading, "↑↓:滚动  q:退出  ^T:主题"
	}
}
