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
			Background(lipgloss.Color(t.ErrorBg())).
			Foreground(lipgloss.Color(t.ErrorFg())).
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
		Foreground(lipgloss.Color(t.ModeFg())).
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
		Background(lipgloss.Color(t.StatusBg())).
		Foreground(lipgloss.Color(t.StatusFg()))
	leftStr := leftStyle.Render(leftContent)

	rightContent := fmt.Sprintf(" %s ", hints)
	rightStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(t.StatusBg())).
		Foreground(lipgloss.Color(t.StatusHintFg()))
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
		return "浏览", t.ModeBrowsingBg(), "↑↓:选择  →/Enter:打开  ←:上级  Esc:返回  q:退出"
	case ModeReading:
		return "阅读", t.ModeReadingBg(), "↑↓:滚动  b:浏览  d:总览  q:退出"
	case ModeSelecting:
		return "选择", t.ModeSelectingBg(), "↑↓:扩选  Enter:批注  Esc:取消"
	case ModeAnnotating:
		if isEditing {
			return "编辑", t.ModeAnnotatingBg(), "Enter:提交  ^J:换行  ^R:调整选区  ^D:删除  Esc:取消"
		}
		return "批注", t.ModeAnnotatingBg(), "Enter:提交  ^J:换行  ^R:调整选区  Esc:取消"
	case ModeOverview:
		return "概览", t.ModeOverviewBg(), "Enter:编辑  ⌫:删除  ↑↓:选择  Esc:返回  q:退出"
	default:
		return "阅读", t.ModeReadingBg(), "↑↓:滚动  q:退出"
	}
}
