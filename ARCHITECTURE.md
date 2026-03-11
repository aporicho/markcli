# Go 重写架构设计

## 技术栈

| 功能 | 包 |
|------|---|
| TUI 框架 | charmbracelet/bubbletea |
| 样式/颜色 | charmbracelet/lipgloss |
| Markdown 渲染 | charmbracelet/glamour |
| 字符宽度 | mattn/go-runewidth |
| 文件监听 | fsnotify/fsnotify |
| CLI 解析 | spf13/cobra |
| JSON | 标准库 encoding/json |

## 目录结构

```
go/
├── cmd/
│   ├── mark/main.go          # TUI 入口
│   └── mark-mcp/main.go      # MCP server 入口
│
├── internal/
│   ├── tui/
│   │   ├── model.go          # Model 定义（含子结构体）
│   │   ├── init.go           # Init()：启动文件监听、IPC 读取 cmd
│   │   ├── update.go         # Update() 顶层分派
│   │   ├── view.go           # View() 组合 ui/ 各组件
│   │   ├── keys.go           # handleKey()（模式感知）
│   │   ├── mouse.go          # SGR 解析 + handleMouse()
│   │   ├── ipc.go            # waitIpcCmd() + handleIpc()
│   │   ├── file.go           # watchFile() + handleFileChanged()
│   │   └── ui/
│   │       ├── viewer.go     # RenderViewer(m Model) string
│   │       ├── input.go      # RenderInput(m Model) string
│   │       ├── overview.go   # RenderOverview(m Model) string
│   │       ├── statusbar.go  # RenderStatusbar(m Model) string
│   │       └── segments.go   # buildSegments 等纯函数
│   │
│   ├── annotation/
│   │   ├── annotation.go     # 类型定义（兼容 .markcli.json）
│   │   ├── store.go          # 读写 .markcli.json
│   │   └── anchor.go         # 文本锚定（textAnchor 移植）
│   │
│   ├── markdown/
│   │   └── render.go         # glamour 渲染 + 行分割
│   │
│   ├── ansi/
│   │   └── ansi.go           # stripAnsi, termColToCharIndex, displayWidth
│   │
│   ├── ipc/
│   │   ├── protocol.go       # Request（含 reply chan）/ Response 类型
│   │   └── server.go         # Unix socket → chan Request
│   │
│   ├── mcp/
│   │   └── server.go         # JSON-RPC stdio + ipc.Client + annotation 降级
│   │
│   ├── theme/
│   │   └── theme.go          # Theme 结构体 + 四套预置主题
│   │
│   └── config/
│       └── config.go         # 读取 ~/.config/markcli/config.json
│
├── go.mod
└── go.sum
```

## Model 设计

```go
type Model struct {
    file     fileState      // FilePath, RawContent, RenderedLines, StrippedLines, TermWidth
    viewport viewportState  // ScrollOffset, Width, Height, ViewportHeight
    select_  selectState    // Mode, SelectionStart/End, PendingMouseClick
    input    inputState     // InputValue, Cursor, EditingID, Pending *PendingSelection
    overview overviewState  // Cursor

    annotations []annotation.Annotation  // 从 .markcli.json 加载
    resolved    []annotation.Annotation  // 锚定重定位后的位置（缓存）

    theme theme.Theme
    ipcCh <-chan ipc.Request  // nil = 无 IPC
}
```

## 依赖图（无环）

```
cmd/mark      → tui → annotation, markdown, ansi, theme, config, ipc
cmd/mark-mcp  → mcp → ipc, annotation
ipc           → 标准库
annotation    → 标准库
ansi          → mattn/go-runewidth
markdown      → glamour
tui/ui        → ansi, annotation, theme, lipgloss
```

## IPC 核心模式：Request 自带 reply channel

```go
// ipc/protocol.go
type Request struct {
    ID     string
    Method string
    Params json.RawMessage
    reply  chan Response  // 不导出，由 server 设置
}

func (r *Request) Reply(resp Response) { r.reply <- resp }
```

流程：IPC server goroutine 创建 Request → 写入 chan → TUI Update 处理 → req.Reply() → server goroutine 写回 conn。
ipc 包不依赖 tui，无循环依赖。

## Update 按输入来源拆分

```
update.go  → 顶层分派
keys.go    → tea.KeyMsg（模式感知）
mouse.go   → SGR 鼠标解析 + 事件处理
ipc.go     → ipc.Request 处理
file.go    → 文件变更处理
```

## .markcli.json 兼容性

与现有 TypeScript 版本完全兼容，字段名不变。
startCol/endCol/resolved 用指针 + omitempty 处理旧数据缺字段的情况。

---

# 实施阶段

## Phase 1：基础数据层
**目标**：无依赖模块，可独立测试
- `go.mod` 初始化
- `internal/annotation/annotation.go` — 类型定义（兼容现有 JSON）
- `internal/annotation/store.go` — 读写 .markcli.json
- `internal/annotation/anchor.go` — 文本锚定（移植 textAnchor.ts）
- `internal/ansi/ansi.go` — stripAnsi, termColToCharIndex, displayWidth
- `internal/tui/ui/segments.go` — buildSegments 等纯区间算法（移植 ranges.ts）
- 每个模块配套测试，对照 TS 测试用例验证行为

**验收**：`go test ./internal/annotation/... ./internal/ansi/... ./internal/tui/ui/...` 全绿

## Phase 2：Markdown 渲染 + 主题
**目标**：能把 md 文件渲染成 ANSI 行数组
- `internal/markdown/render.go` — glamour 渲染 + 宽度换行 + 行分割
- `internal/theme/theme.go` — 四套主题
- `internal/config/config.go` — 配置读取

**验收**：单测验证渲染行数、CJK 字符宽度计算正确

## Phase 3：TUI 骨架（reading 模式）
**目标**：能打开文件、滚动阅读、显示已有批注高亮
- `internal/ipc/protocol.go` + `ipc/server.go`
- `internal/tui/model.go` — Model + 子结构体
- `internal/tui/init.go` — Init()
- `internal/tui/update.go` + `file.go` — 文件变更处理
- `internal/tui/ui/viewer.go` + `ui/statusbar.go` — 渲染
- `internal/tui/keys.go` — 仅 reading 模式按键（滚动/退出）
- `cmd/mark/main.go` — open 子命令

**验收**：`go run ./cmd/mark README.md` 能滚动阅读，批注高亮显示

## Phase 4：鼠标 + 选区（selecting 模式）
**目标**：鼠标拖拽/键盘选区，高亮显示
- `internal/tui/mouse.go` — SGR 1006 解析 + handleMouse()
- `internal/tui/keys.go` — 补充 selecting 模式按键（v/方向键/ESC）
- `internal/tui/ui/viewer.go` — 选区高亮渲染

**验收**：鼠标拖拽/v 键选区，选中文本高亮，ESC 取消

## Phase 5：批注输入（annotating 模式）
**目标**：选区确认后输入批注，保存到 .markcli.json
- `internal/tui/keys.go` — annotating 模式按键（Enter 确认/ESC 取消）
- `internal/tui/ui/input.go` — 浮动输入框渲染（lipgloss overlay）
- 批注保存、加载、锚定重定位

**验收**：完整流程 selecting → annotating → 保存 → 重新加载显示

## Phase 6：总览面板 + 删除模式
**目标**：d 键打开总览，支持编辑/删除批注
- `internal/tui/ui/overview.go` — 总览面板渲染
- `internal/tui/keys.go` — overview 模式按键
- 编辑模式（双击/回车进入）

**验收**：d 键打开总览，e 编辑，x 删除

## Phase 7：IPC + MCP server
**目标**：MCP 工具链完整可用
- `internal/tui/ipc.go` — waitIpcCmd() + 全部 IPC handler
- `internal/mcp/server.go` — JSON-RPC stdio（10 个工具）
- `cmd/mark-mcp/main.go`
- list/show/clear 子命令（cmd/mark）

**验收**：`claude mcp add mark-go -- mark-go mcp` 配置后，list_annotations/add_annotation 等工具可用

## Phase 8：构建 + 发布
**目标**：替换现有 TypeScript 版本
- CI workflow 更新（go build 替换 bun compile）
- install.sh 更新（二进制名不变）
- 删除 TypeScript 源码
- 验证二进制大小 < 20MB

**验收**：`go build` 产物 < 20MB，install.sh 安装后功能完整
