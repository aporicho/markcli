package mcp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/ipc"
)

// NewServer creates and configures an MCP server with all 11 tools.
func NewServer(ipcClient *ipc.Client) *server.MCPServer {
	s := server.NewMCPServer("mark", "0.1.0")

	s.AddTool(
		mcp.NewTool("list_annotations",
			mcp.WithDescription("读取 Mark 中的批注。返回用户对文件内容的划线批注，包括选中的文本和批注内容。"),
			mcp.WithString("file", mcp.Description("文件路径（可选，默认为当前打开的文件）")),
		),
		handleListAnnotations(ipcClient),
	)

	s.AddTool(
		mcp.NewTool("get_status",
			mcp.WithDescription("查询 Mark 的运行状态，包括当前打开的文件、批注数量和模式。"),
		),
		handleGetStatus(ipcClient),
	)

	s.AddTool(
		mcp.NewTool("get_selection",
			mcp.WithDescription("获取用户当前在 Mark 中选中的文本。"),
		),
		handleGetSelection(ipcClient),
	)

	s.AddTool(
		mcp.NewTool("open_file",
			mcp.WithDescription("让 Mark 打开指定文件。Mark 必须已经在运行。"),
			mcp.WithString("path", mcp.Required(), mcp.Description("要打开的文件路径")),
		),
		handleOpenFile(ipcClient),
	)

	s.AddTool(
		mcp.NewTool("refresh_file",
			mcp.WithDescription("通知 Mark 重新读取当前打开的文件。当你修改了文件内容后调用此工具，Mark 会立即刷新显示。Mark 必须已经在运行。"),
		),
		handleRefreshFile(ipcClient),
	)

	s.AddTool(
		mcp.NewTool("add_annotation",
			mcp.WithDescription("在 Mark 中添加批注。只需指定要选中的文本和批注内容，自动定位文本位置。TUI 运行时通过 IPC 通信，未运行时直接操作 .markcli.json 文件（需指定 file 参数）。"),
			mcp.WithString("file", mcp.Description("文件路径（可选，默认为当前打开的文件）")),
			mcp.WithString("selectedText", mcp.Required(), mcp.Description("选中的原文")),
			mcp.WithString("comment", mcp.Required(), mcp.Description("批注内容")),
		),
		handleAddAnnotation(ipcClient),
	)

	s.AddTool(
		mcp.NewTool("update_annotation",
			mcp.WithDescription("更新 Mark 中已有批注的内容。需要提供批注 ID 和新的批注内容。TUI 运行时通过 IPC 通信，未运行时直接操作 .markcli.json 文件（需指定 file 参数）。"),
			mcp.WithString("id", mcp.Required(), mcp.Description("批注 ID")),
			mcp.WithString("comment", mcp.Required(), mcp.Description("新的批注内容")),
			mcp.WithString("file", mcp.Description("文件路径（可选，TUI 未运行时必须指定）")),
		),
		handleUpdateAnnotation(ipcClient),
	)

	s.AddTool(
		mcp.NewTool("remove_annotation",
			mcp.WithDescription("删除 Mark 中的一条批注。需要提供批注 ID。TUI 运行时通过 IPC 通信，未运行时直接操作 .markcli.json 文件（需指定 file 参数）。"),
			mcp.WithString("id", mcp.Required(), mcp.Description("批注 ID")),
			mcp.WithString("file", mcp.Description("文件路径（可选，TUI 未运行时必须指定）")),
		),
		handleRemoveAnnotation(ipcClient),
	)

	s.AddTool(
		mcp.NewTool("resolve_annotation",
			mcp.WithDescription("将 Mark 中的一条批注标记为已处理（resolved）。批注不会被删除，而是以不同颜色显示。需要提供批注 ID。TUI 运行时通过 IPC 通信，未运行时直接操作 .markcli.json 文件（需指定 file 参数）。"),
			mcp.WithString("id", mcp.Required(), mcp.Description("批注 ID")),
			mcp.WithString("file", mcp.Description("文件路径（可选，TUI 未运行时必须指定）")),
		),
		handleResolveAnnotation(ipcClient),
	)

	s.AddTool(
		mcp.NewTool("clear_annotations",
			mcp.WithDescription("清空指定文件的全部批注。一次性删除所有批注及 .markcli.json 文件。"),
			mcp.WithString("file", mcp.Description("文件路径（可选，默认为当前打开的文件）")),
		),
		handleClearAnnotations(ipcClient),
	)

	s.AddTool(
		mcp.NewTool("jump_to_annotation",
			mcp.WithDescription("让 Mark TUI 滚动到指定批注的位置，用户可以实时看到当前正在处理哪条批注。Mark 必须已经在运行。"),
			mcp.WithString("id", mcp.Required(), mcp.Description("批注 ID")),
		),
		handleJumpToAnnotation(ipcClient),
	)

	return s
}

// --- Tool Handlers ---

func handleListAnnotations(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file := req.GetString("file", "")

		if c.IsConnected() {
			resp, err := c.Send(map[string]any{"type": "list_annotations", "file": file})
			if err == nil && resp.Type == "annotations" {
				return mcp.NewToolResultText(string(resp.Data)), nil
			}
		}

		// Fallback: read .markcli.json directly
		if file != "" {
			af, err := annotation.Load(file)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to load annotations: %s", err)), nil
			}
			data, err := json.Marshal(af)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal annotations: %s", err)), nil
			}
			return mcp.NewToolResultText(string(data)), nil
		}

		return mcp.NewToolResultError("Mark is not running. Provide a file path for offline access."), nil
	}
}

func handleGetStatus(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if !c.IsConnected() {
			return mcp.NewToolResultText("Mark is not running"), nil
		}
		resp, err := c.Send(map[string]any{"type": "get_status"})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to connect to Mark: %s", err)), nil
		}
		if resp.Type == "status" {
			return mcp.NewToolResultText(string(resp.Data)), nil
		}
		return mcp.NewToolResultError("Unexpected response from Mark"), nil
	}
}

func handleGetSelection(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if !c.IsConnected() {
			return mcp.NewToolResultText("Mark is not running"), nil
		}
		resp, err := c.Send(map[string]any{"type": "get_selection"})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to connect to Mark: %s", err)), nil
		}
		if resp.Type == "selection" {
			return mcp.NewToolResultText(string(resp.Data)), nil
		}
		return mcp.NewToolResultError("Unexpected response from Mark"), nil
	}
}

func handleOpenFile(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := req.GetString("path", "")
		if !c.IsConnected() {
			return mcp.NewToolResultError("Mark is not running. Please start Mark with `mark <file>` first."), nil
		}
		resp, err := c.Send(map[string]any{"type": "open_file", "path": path})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to connect to Mark: %s", err)), nil
		}
		if resp.Type == "ok" {
			return mcp.NewToolResultText(resp.Message), nil
		}
		return mcp.NewToolResultError(resp.Message), nil
	}
}

func handleRefreshFile(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		if !c.IsConnected() {
			return mcp.NewToolResultError("Mark is not running. Please start Mark with `mark <file>` first."), nil
		}
		resp, err := c.Send(map[string]any{"type": "refresh_file"})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to connect to Mark: %s", err)), nil
		}
		if resp.Type == "ok" {
			return mcp.NewToolResultText(resp.Message), nil
		}
		return mcp.NewToolResultError(resp.Message), nil
	}
}

func handleAddAnnotation(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file := req.GetString("file", "")
		selectedText := req.GetString("selectedText", "")
		comment := req.GetString("comment", "")

		if c.IsConnected() {
			resp, err := c.Send(map[string]any{
				"type":         "add_annotation",
				"file":         file,
				"selectedText": selectedText,
				"comment":      comment,
			})
			if err == nil {
				if resp.Type == "ok" {
					return mcp.NewToolResultText(resp.Message), nil
				}
				return mcp.NewToolResultError(resp.Message), nil
			}
			// IPC failed, try fallback
		}

		// Fallback: operate on file directly
		if file != "" {
			return addAnnotationOffline(file, selectedText, comment)
		}

		return mcp.NewToolResultError("Mark is not running. Please specify a file path or start Mark first."), nil
	}
}

func handleUpdateAnnotation(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetString("id", "")
		comment := req.GetString("comment", "")
		file := req.GetString("file", "")

		if c.IsConnected() {
			resp, err := c.Send(map[string]any{"type": "update_annotation", "id": id, "comment": comment})
			if err == nil {
				if resp.Type == "ok" {
					return mcp.NewToolResultText(resp.Message), nil
				}
				return mcp.NewToolResultError(resp.Message), nil
			}
		}

		// Fallback
		if file != "" {
			af, err := annotation.Load(file)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to load annotations: %s", err)), nil
			}
			found := false
			for i, ann := range af.Annotations {
				if ann.ID == id {
					af.Annotations[i].Comment = comment
					found = true
					break
				}
			}
			if !found {
				return mcp.NewToolResultError(fmt.Sprintf("Annotation %s not found", id)), nil
			}
			if err := annotation.Save(file, af); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to save: %s", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Annotation %s updated", id)), nil
		}

		return mcp.NewToolResultError("Mark is not running. Please specify a file path or start Mark first."), nil
	}
}

func handleRemoveAnnotation(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetString("id", "")
		file := req.GetString("file", "")

		if c.IsConnected() {
			resp, err := c.Send(map[string]any{"type": "remove_annotation", "id": id})
			if err == nil {
				if resp.Type == "ok" {
					return mcp.NewToolResultText(resp.Message), nil
				}
				return mcp.NewToolResultError(resp.Message), nil
			}
		}

		// Fallback
		if file != "" {
			af, err := annotation.Load(file)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to load annotations: %s", err)), nil
			}
			newAnns := make([]annotation.Annotation, 0, len(af.Annotations))
			found := false
			for _, ann := range af.Annotations {
				if ann.ID == id {
					found = true
				} else {
					newAnns = append(newAnns, ann)
				}
			}
			if !found {
				return mcp.NewToolResultError(fmt.Sprintf("Annotation %s not found", id)), nil
			}
			af.Annotations = newAnns
			if err := annotation.Save(file, af); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to save: %s", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Annotation %s removed", id)), nil
		}

		return mcp.NewToolResultError("Mark is not running. Please specify a file path or start Mark first."), nil
	}
}

func handleResolveAnnotation(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetString("id", "")
		file := req.GetString("file", "")

		if c.IsConnected() {
			resp, err := c.Send(map[string]any{"type": "resolve_annotation", "id": id})
			if err == nil {
				if resp.Type == "ok" {
					return mcp.NewToolResultText(resp.Message), nil
				}
				return mcp.NewToolResultError(resp.Message), nil
			}
		}

		// Fallback
		if file != "" {
			af, err := annotation.Load(file)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to load annotations: %s", err)), nil
			}
			found := false
			for i, ann := range af.Annotations {
				if ann.ID == id {
					resolved := true
					af.Annotations[i].Resolved = &resolved
					found = true
					break
				}
			}
			if !found {
				return mcp.NewToolResultError(fmt.Sprintf("Annotation %s not found", id)), nil
			}
			if err := annotation.Save(file, af); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to save: %s", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Annotation %s resolved", id)), nil
		}

		return mcp.NewToolResultError("Mark is not running. Please specify a file path or start Mark first."), nil
	}
}

func handleClearAnnotations(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		file := req.GetString("file", "")

		if c.IsConnected() {
			resp, err := c.Send(map[string]any{"type": "clear_annotations", "file": file})
			if err == nil {
				if resp.Type == "ok" {
					return mcp.NewToolResultText(resp.Message), nil
				}
				return mcp.NewToolResultError(resp.Message), nil
			}
		}

		// Fallback
		if file != "" {
			if err := annotation.Clear(file); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to clear annotations: %s", err)), nil
			}
			return mcp.NewToolResultText("All annotations cleared"), nil
		}

		return mcp.NewToolResultError("Mark is not running. Please specify a file path or start Mark first."), nil
	}
}

func handleJumpToAnnotation(c *ipc.Client) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id := req.GetString("id", "")
		if !c.IsConnected() {
			return mcp.NewToolResultError("Mark is not running. Please start Mark with `mark <file>` first."), nil
		}
		resp, err := c.Send(map[string]any{"type": "jump_to_annotation", "id": id})
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to connect to Mark: %s", err)), nil
		}
		if resp.Type == "ok" {
			return mcp.NewToolResultText(resp.Message), nil
		}
		return mcp.NewToolResultError(resp.Message), nil
	}
}

// --- Offline helpers ---

// addAnnotationOffline creates an annotation by reading the markdown file directly.
func addAnnotationOffline(file, selectedText, comment string) (*mcp.CallToolResult, error) {
	content, err := os.ReadFile(file)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to read file: %s", err)), nil
	}

	fullText := string(content)
	lines := strings.Split(fullText, "\n")
	lineLengths := make([]int, len(lines))
	for i, line := range lines {
		lineLengths[i] = len([]rune(line))
	}

	// Try exact match first
	matchStart := strings.Index(fullText, selectedText)
	var matchEnd int

	if matchStart == -1 {
		// Try fuzzy match
		anchor := annotation.TextAnchor{Quote: selectedText}
		r := annotation.RelocateAnchor(fullText, anchor)
		if r == nil {
			preview := selectedText
			if len([]rune(preview)) > 50 {
				preview = string([]rune(preview)[:50]) + "..."
			}
			return mcp.NewToolResultError(fmt.Sprintf("Text not found: \"%s\"", preview)), nil
		}
		matchStart = r.Start
		matchEnd = r.End
	} else {
		// Convert byte offset to rune offset
		matchStart = len([]rune(fullText[:matchStart]))
		matchEnd = matchStart + len([]rune(selectedText))
	}
	startLine, startCol := annotation.OffsetToLineCol(lineLengths, matchStart)
	endLine, endCol := annotation.OffsetToLineCol(lineLengths, matchEnd)
	anchor := annotation.ExtractAnchor(fullText, matchStart, matchEnd)

	ann := annotation.Annotation{
		ID:           generateOfflineID(),
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

	af, err := annotation.Load(file)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to load annotations: %s", err)), nil
	}
	af.Annotations = append(af.Annotations, ann)
	af.File = filepath.Base(file)
	if err := annotation.Save(file, af); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save: %s", err)), nil
	}
	return mcp.NewToolResultText("Annotation added"), nil
}

// generateOfflineID returns a 6-char random hex ID.
func generateOfflineID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
