package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/ipc"
	"github.com/aporicho/markcli/internal/tui/ui"
)

// ipcRequestMsg wraps an IPC request as a bubbletea message.
type ipcRequestMsg struct {
	req ipc.Request
}

// waitIpcCmd blocks on the IPC channel and returns the next request as a tea.Msg.
// Returns nil if ch is nil (no IPC server running).
func waitIpcCmd(ch <-chan ipc.Request) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		req, ok := <-ch
		if !ok {
			return nil
		}
		return ipcRequestMsg{req: req}
	}
}

// handleIpc processes an IPC request and returns the updated model + command.
func handleIpc(m Model, req ipc.Request) (Model, tea.Cmd) {
	switch req.Method {
	case "get_status":
		return ipcGetStatus(m, req)
	case "get_selection":
		return ipcGetSelection(m, req)
	case "list_annotations":
		return ipcListAnnotations(m, req)
	case "open_file":
		return ipcOpenFile(m, req)
	case "refresh_file":
		return ipcRefreshFile(m, req)
	case "add_annotation":
		return ipcAddAnnotation(m, req)
	case "update_annotation":
		return ipcUpdateAnnotation(m, req)
	case "remove_annotation":
		return ipcRemoveAnnotation(m, req)
	case "resolve_annotation":
		return ipcResolveAnnotation(m, req)
	case "clear_annotations":
		return ipcClearAnnotations(m, req)
	case "jump_to_annotation":
		return ipcJumpToAnnotation(m, req)
	default:
		req.Reply(ipc.Response{
			Type:    "error",
			Message: fmt.Sprintf("unknown method: %s", req.Method),
		})
		return m, waitIpcCmd(m.ipcCh)
	}
}

// --- Read-only handlers ---

func ipcGetStatus(m Model, req ipc.Request) (Model, tea.Cmd) {
	data, err := json.Marshal(map[string]any{
		"file":            filepath.Base(m.file.FilePath),
		"mode":            string(m.mode),
		"annotationCount": len(m.annotations),
	})
	if err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("marshal failed: %s", err)})
		return m, waitIpcCmd(m.ipcCh)
	}
	req.Reply(ipc.Response{Type: "status", Data: data})
	return m, waitIpcCmd(m.ipcCh)
}

func ipcGetSelection(m Model, req ipc.Request) (Model, tea.Cmd) {
	var selectedText *string
	if m.selection.Active && (m.mode == ui.ModeSelecting || m.mode == ui.ModeAnnotating) {
		s, e := ui.NormalizePos(m.selection.Start, m.selection.End)
		text := getSelectedText(m.file.StrippedLines, s, e)
		if text != "" {
			selectedText = &text
		}
	}
	data, err := json.Marshal(map[string]any{
		"file":         filepath.Base(m.file.FilePath),
		"selectedText": selectedText,
	})
	if err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("marshal failed: %s", err)})
		return m, waitIpcCmd(m.ipcCh)
	}
	req.Reply(ipc.Response{Type: "selection", Data: data})
	return m, waitIpcCmd(m.ipcCh)
}

func ipcListAnnotations(m Model, req ipc.Request) (Model, tea.Cmd) {
	data, err := json.Marshal(map[string]any{
		"file":        filepath.Base(m.file.FilePath),
		"annotations": m.resolved,
	})
	if err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("marshal failed: %s", err)})
		return m, waitIpcCmd(m.ipcCh)
	}
	req.Reply(ipc.Response{Type: "annotations", Data: data})
	return m, waitIpcCmd(m.ipcCh)
}

// --- Write handlers (need save + reload) ---

// ipcParams is a helper to unmarshal params from the raw request.
type ipcParams map[string]any

func parseParams(req ipc.Request) ipcParams {
	var p ipcParams
	_ = json.Unmarshal(req.Params, &p)
	if p == nil {
		p = make(ipcParams)
	}
	return p
}

func (p ipcParams) getString(key string) string {
	v, _ := p[key].(string)
	return v
}

func ipcOpenFile(m Model, req ipc.Request) (Model, tea.Cmd) {
	params := parseParams(req)
	path := params.getString("path")
	if path == "" {
		req.Reply(ipc.Response{Type: "error", Message: "missing required param: path"})
		return m, waitIpcCmd(m.ipcCh)
	}

	// Resolve to absolute path, follow symlinks
	absPath, err := filepath.Abs(path)
	if err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("invalid path: %s", err)})
		return m, waitIpcCmd(m.ipcCh)
	}
	absPath, err = filepath.EvalSymlinks(absPath)
	if err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("path not found: %s", err)})
		return m, waitIpcCmd(m.ipcCh)
	}

	// Check file exists and is not a directory
	info, err := os.Stat(absPath)
	if err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("file not found: %s", absPath)})
		return m, waitIpcCmd(m.ipcCh)
	}
	if info.IsDir() {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("path is a directory: %s", absPath)})
		return m, waitIpcCmd(m.ipcCh)
	}

	m.file.FilePath = absPath
	m.mode = ui.ModeReading
	m.selection = selectionState{}
	m.input = inputState{}
	m.editingID = ""
	m.overview = overviewState{}
	m.viewport.ScrollOffset = 0

	req.Reply(ipc.Response{Type: "ok", Message: fmt.Sprintf("Opened %s", filepath.Base(absPath))})
	return m, tea.Batch(loadFileCmd(absPath, m.viewport.Width, loadIPC), waitIpcCmd(m.ipcCh))
}

func ipcRefreshFile(m Model, req ipc.Request) (Model, tea.Cmd) {
	req.Reply(ipc.Response{Type: "ok", Message: "File refreshed"})
	return m, tea.Batch(loadFileCmd(m.file.FilePath, m.viewport.Width, loadIPC), waitIpcCmd(m.ipcCh))
}

func ipcAddAnnotation(m Model, req ipc.Request) (Model, tea.Cmd) {
	params := parseParams(req)
	selectedText := params.getString("selectedText")
	comment := params.getString("comment")

	if selectedText == "" {
		req.Reply(ipc.Response{Type: "error", Message: "missing required param: selectedText"})
		return m, waitIpcCmd(m.ipcCh)
	}
	if comment == "" {
		req.Reply(ipc.Response{Type: "error", Message: "missing required param: comment"})
		return m, waitIpcCmd(m.ipcCh)
	}

	fullText := strings.Join(m.file.StrippedLines, "\n")

	// Try exact match first
	matchStart := strings.Index(fullText, selectedText)
	var matchEnd int

	if matchStart == -1 {
		// Try fuzzy match via RelocateAnchor
		anchor := annotation.TextAnchor{
			Quote:  selectedText,
			Prefix: "",
			Suffix: "",
		}
		r := annotation.RelocateAnchor(fullText, anchor)
		if r == nil {
			req.Reply(ipc.Response{
				Type:    "error",
				Message: fmt.Sprintf("Text not found: \"%s\"", truncate(selectedText, 50)),
			})
			return m, waitIpcCmd(m.ipcCh)
		}
		matchStart = r.Start
		matchEnd = r.End
	} else {
		// Convert byte offset to rune offset
		matchStart = len([]rune(fullText[:matchStart]))
		matchEnd = matchStart + len([]rune(selectedText))
	}
	startLine, startCol := annotation.OffsetToLineCol(m.file.LineLengths, matchStart)
	endLine, endCol := annotation.OffsetToLineCol(m.file.LineLengths, matchEnd)
	anchor := annotation.ExtractAnchor(fullText, matchStart, matchEnd)

	ann := annotation.Annotation{
		ID:           generateID(),
		StartLine:    startLine,
		EndLine:      endLine,
		StartCol:     &startCol,
		EndCol:       &endCol,
		SelectedText: selectedText,
		Comment:      comment,
		CreatedAt:    time.Now().Format(time.RFC3339),
		Quote:        anchor.Quote,
		Prefix:       anchor.Prefix,
		Suffix:       anchor.Suffix,
	}
	m.annotations = append(m.annotations, ann)

	af := annotation.AnnotationFile{
		File:        filepath.Base(m.file.FilePath),
		Annotations: m.annotations,
	}
	if err := annotation.Save(m.file.FilePath, af); err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("save failed: %s", err)})
		return m, waitIpcCmd(m.ipcCh)
	}

	req.Reply(ipc.Response{Type: "ok", Message: "Annotation added"})
	return m, tea.Batch(loadFileCmd(m.file.FilePath, m.viewport.Width, loadIPC), waitIpcCmd(m.ipcCh))
}

func ipcUpdateAnnotation(m Model, req ipc.Request) (Model, tea.Cmd) {
	params := parseParams(req)
	id := params.getString("id")
	comment := params.getString("comment")

	if id == "" {
		req.Reply(ipc.Response{Type: "error", Message: "missing required param: id"})
		return m, waitIpcCmd(m.ipcCh)
	}
	if comment == "" {
		req.Reply(ipc.Response{Type: "error", Message: "missing required param: comment"})
		return m, waitIpcCmd(m.ipcCh)
	}

	found := false
	for i, ann := range m.annotations {
		if ann.ID == id {
			m.annotations[i].Comment = comment
			found = true
			break
		}
	}
	if !found {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("Annotation %s not found", id)})
		return m, waitIpcCmd(m.ipcCh)
	}

	af := annotation.AnnotationFile{
		File:        filepath.Base(m.file.FilePath),
		Annotations: m.annotations,
	}
	if err := annotation.Save(m.file.FilePath, af); err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("save failed: %s", err)})
		return m, waitIpcCmd(m.ipcCh)
	}

	req.Reply(ipc.Response{Type: "ok", Message: fmt.Sprintf("Annotation %s updated", id)})
	return m, tea.Batch(loadFileCmd(m.file.FilePath, m.viewport.Width, loadIPC), waitIpcCmd(m.ipcCh))
}

func ipcRemoveAnnotation(m Model, req ipc.Request) (Model, tea.Cmd) {
	params := parseParams(req)
	id := params.getString("id")

	if id == "" {
		req.Reply(ipc.Response{Type: "error", Message: "missing required param: id"})
		return m, waitIpcCmd(m.ipcCh)
	}

	newAnns := make([]annotation.Annotation, 0, len(m.annotations))
	found := false
	for _, ann := range m.annotations {
		if ann.ID == id {
			found = true
		} else {
			newAnns = append(newAnns, ann)
		}
	}
	if !found {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("Annotation %s not found", id)})
		return m, waitIpcCmd(m.ipcCh)
	}

	m.annotations = newAnns
	af := annotation.AnnotationFile{
		File:        filepath.Base(m.file.FilePath),
		Annotations: m.annotations,
	}
	if err := annotation.Save(m.file.FilePath, af); err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("save failed: %s", err)})
		return m, waitIpcCmd(m.ipcCh)
	}

	req.Reply(ipc.Response{Type: "ok", Message: fmt.Sprintf("Annotation %s removed", id)})
	return m, tea.Batch(loadFileCmd(m.file.FilePath, m.viewport.Width, loadIPC), waitIpcCmd(m.ipcCh))
}

func ipcResolveAnnotation(m Model, req ipc.Request) (Model, tea.Cmd) {
	params := parseParams(req)
	id := params.getString("id")

	if id == "" {
		req.Reply(ipc.Response{Type: "error", Message: "missing required param: id"})
		return m, waitIpcCmd(m.ipcCh)
	}

	found := false
	for i, ann := range m.annotations {
		if ann.ID == id {
			resolved := true
			m.annotations[i].Resolved = &resolved
			found = true
			break
		}
	}
	if !found {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("Annotation %s not found", id)})
		return m, waitIpcCmd(m.ipcCh)
	}

	af := annotation.AnnotationFile{
		File:        filepath.Base(m.file.FilePath),
		Annotations: m.annotations,
	}
	if err := annotation.Save(m.file.FilePath, af); err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("save failed: %s", err)})
		return m, waitIpcCmd(m.ipcCh)
	}

	req.Reply(ipc.Response{Type: "ok", Message: fmt.Sprintf("Annotation %s resolved", id)})
	return m, tea.Batch(loadFileCmd(m.file.FilePath, m.viewport.Width, loadIPC), waitIpcCmd(m.ipcCh))
}

func ipcClearAnnotations(m Model, req ipc.Request) (Model, tea.Cmd) {
	m.annotations = nil
	if err := annotation.Clear(m.file.FilePath); err != nil {
		req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("clear failed: %s", err)})
		return m, waitIpcCmd(m.ipcCh)
	}

	req.Reply(ipc.Response{Type: "ok", Message: "All annotations cleared"})
	return m, tea.Batch(loadFileCmd(m.file.FilePath, m.viewport.Width, loadIPC), waitIpcCmd(m.ipcCh))
}

func ipcJumpToAnnotation(m Model, req ipc.Request) (Model, tea.Cmd) {
	params := parseParams(req)
	id := params.getString("id")

	if id == "" {
		req.Reply(ipc.Response{Type: "error", Message: "missing required param: id"})
		return m, waitIpcCmd(m.ipcCh)
	}

	for _, ann := range m.resolved {
		if ann.ID == id {
			autoScrollToLine(&m, ann.StartLine)
			req.Reply(ipc.Response{Type: "ok", Message: fmt.Sprintf("Jumped to annotation %s", id)})
			return m, waitIpcCmd(m.ipcCh)
		}
	}

	req.Reply(ipc.Response{Type: "error", Message: fmt.Sprintf("Annotation %s not found", id)})
	return m, waitIpcCmd(m.ipcCh)
}

// truncate returns the first n runes of s, appending "..." if truncated.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}
