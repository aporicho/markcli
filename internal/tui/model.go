package tui

import (
	"strings"
	"time"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/ipc"
	"github.com/aporicho/markcli/internal/theme"
	"github.com/aporicho/markcli/internal/tui/ui"
)

type fileState struct {
	FilePath      string
	RenderedLines []string // ANSI, from markdown.RenderToLines
	StrippedLines []string // plain text lines (ANSI stripped)
	LineLengths   []int    // rune count per line
}

type clickPos struct {
	Line int // 1-based
	Col  int // 0-based
}

type clickRecord struct {
	Time time.Time
	Line int
	Col  int
}

type selectionState struct {
	Active       bool
	Start        annotation.SelectionPos // anchor point
	End          annotation.SelectionPos // moving endpoint
	PendingClick *clickPos               // press not yet dragged
	LastClick    *clickRecord            // double-click detection (400ms window)
}

type viewportState struct {
	ScrollOffset   int
	Width          int
	Height         int // terminal total rows
	ViewportHeight int // Height - 1 (minus status bar)
}

type inputState struct {
	Value  []rune // 输入内容
	Cursor int    // 光标位置（rune 索引）
}

type overviewState struct {
	Cursor int // 当前选中的批注索引
}

// Model is the bubbletea model for the TUI.
type Model struct {
	file        fileState
	viewport    viewportState
	mode        ui.AppMode
	selection   selectionState
	input       inputState
	overview    overviewState
	editingID   string                   // 编辑现有批注时的 ID
	annotations []annotation.Annotation // raw from .markcli.json
	resolved    []annotation.Annotation // positions relocated by anchor
	theme       theme.Theme
	themeIndex  int              // for Ctrl+T cycling
	ipcCh       <-chan ipc.Request // nil = no IPC (Phase 7)
	loaded      bool             // true after first WindowSizeMsg triggers file load
	errText     string           // shown in statusbar, cleared on next key/mouse
}

// errMsg is a bubbletea message carrying an error to display in the statusbar.
type errMsg struct{ err error }

// New returns an initial Model. file/viewport are zero-valued until Init() fills them.
func New(filePath string, t theme.Theme, themeIndex int, ipcCh <-chan ipc.Request) Model {
	return Model{
		file:       fileState{FilePath: filePath},
		mode:       ui.ModeReading,
		theme:      t,
		themeIndex: themeIndex,
		ipcCh:      ipcCh,
	}
}

// resolveAnnotations relocates annotation positions using TextAnchor fuzzy matching.
func resolveAnnotations(strippedLines []string, lineLengths []int, annotations []annotation.Annotation) []annotation.Annotation {
	fullText := strings.Join(strippedLines, "\n")
	resolved := make([]annotation.Annotation, 0, len(annotations))

	for _, ann := range annotations {
		if ann.Quote == "" {
			resolved = append(resolved, ann)
			continue
		}
		anchor := annotation.TextAnchor{
			Quote:  ann.Quote,
			Prefix: ann.Prefix,
			Suffix: ann.Suffix,
		}
		r := annotation.RelocateAnchor(fullText, anchor)
		if r != nil {
			line, col := annotation.OffsetToLineCol(lineLengths, r.Start)
			endLine, endCol := annotation.OffsetToLineCol(lineLengths, r.End)
			startCol := col
			ann.StartLine = line
			ann.EndLine = endLine
			ann.StartCol = &startCol
			ann.EndCol = &endCol
		}
		resolved = append(resolved, ann)
	}

	return resolved
}
