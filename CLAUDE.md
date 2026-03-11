# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 产品定义

Mark 是 Claude Code 的精准指令工具。用户在终端里阅读 Markdown 内容，对特定文本划线并写批注，批注结果自动回到 Claude Code 对话上下文，Claude 据此执行修改。

**核心价值**: 指哪打哪——比纯文字对话精确，消除"第三段改一下"这类模糊指令。

### 使用流程

1. Claude Code 输出方案/文档/代码说明
2. 用户运行 `mark file.md`，在终端中阅读
3. 用户对具体文本划线，写批注指令（如"展开讲"、"改用 SQLite"、"加错误处理"）
4. 用户退出，批注结果自动回到 Claude Code 上下文
5. Claude 根据批注精准修改
6. Claude 完成每条批注后调用 `resolve_annotation` 标记为已处理，全部完成后调用 `clear_annotations` 清理

### 核心场景

- **方案 review**: Claude 写了实现方案，用户划出不同意的部分并给出修改意见
- **精准指令**: 划一段代码说"加错误处理"，划一段架构描述说"改用单体"
- **文档批注**: 对生成的文档标注需要修改、展开、删除的部分

### 设计原则

- **输出即结果**: 退出时将批注输出到 stdout，格式让 Claude 能精准理解
- **零摩擦集成**: 作为 Claude Code 的 MCP 工具/slash command，批注自动回到对话

## Build & Run Commands

```bash
go build ./cmd/mark          # 编译二进制到当前目录
go run ./cmd/mark open <file>  # 直接运行
go run ./cmd/mark mcp          # 启动 MCP server
```

CLI 子命令: `open <file>` | `list <file>` | `show <file>` | `clear <file>` | `mcp` | `update` | `doctor` | `completion`

MCP server: `mark mcp`（独立进程，Claude Code 后台自动运行）

配置: `claude mcp add mark -- mark mcp`

## Test & Lint

```bash
go test ./...                          # 运行全部测试
go test ./internal/tui/...             # 运行单个包测试
go test -run TestBuildSegments ./internal/tui/ui/  # 运行单个测试
go vet ./...                           # 静态检查
```

测试文件与源文件同目录，命名 `<name>_test.go`。

## Build & Release

```bash
# 交叉编译（CI 自动执行）
GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=x.y.z" -o dist/mark-darwin-arm64 ./cmd/mark

# 发版
node scripts/release.mjs patch   # patch|minor|major|x.y.z（打 tag 触发 CI）
```

## Architecture

使用 **Bubbletea + Lipgloss + Glamour** 构建终端 UI。纯 Go 实现，单二进制分发。

### 目录结构

```
cmd/mark/main.go           # CLI 入口 (cobra)，所有子命令
internal/
├── tui/                   # TUI 核心
│   ├── model.go           # Model 定义（含子结构体）
│   ├── init.go            # Init()：启动文件监听、IPC
│   ├── update.go          # Update() 顶层分派
│   ├── view.go            # View() 组合 ui/ 各组件
│   ├── keys.go            # handleKey()（模式感知）
│   ├── mouse.go           # SGR 1006 解析 + handleMouse()
│   ├── annotate.go        # 批注提交/删除/编辑逻辑
│   ├── ipc.go             # waitIpcCmd() + handleIpc()
│   ├── file.go            # watchFile() + handleFileChanged()
│   └── ui/                # 纯渲染函数
│       ├── viewer.go      # RenderViewer()
│       ├── input.go       # RenderInputPanel()
│       ├── overview.go    # RenderOverviewPanel()
│       ├── statusbar.go   # RenderStatusbar()
│       └── segments.go    # buildSegments 等纯函数
├── annotation/            # 批注数据：类型、存储、文本锚定
├── markdown/              # glamour 渲染 + 行分割
├── ansi/                  # ANSI 转义序列处理
├── ipc/                   # Unix socket NDJSON 协议
├── mcp/                   # MCP server（JSON-RPC stdio）
├── theme/                 # 四套预置主题
└── config/                # ~/.config/markcli/config.json
```

### 四种应用模式

`reading` → `selecting` → `annotating`，加上 `overview` 模式，由 `model.go` 中的 `mode` 状态驱动。鼠标拖拽可直接从 reading 跳到 annotating；双击批注可进入编辑模式。

### 核心数据流

1. **Markdown 渲染**: `glamour` 将 Markdown 转为 ANSI 终端文本 → 按行分割 → `ui.RenderViewer()` 渲染
2. **选择**: `mouse.go`(SGR 1006 鼠标协议) 和 `keys.go` 产生选择事件 → `model.selectionState` 管理选区 → `segments.go` 的 `BuildSegments()` 将行切分为高亮/普通片段
3. **批注保存**: `annotate.go` 提取文本锚定 → 写入 `.markcli.json` → 触发文件重载

### MCP 集成（单二进制架构）

```
Claude Code  <--stdio JSON-RPC-->  mark mcp (MCP server，同一二进制)
                                       |
                                   Unix socket (/tmp/mark-<uid>.sock)
                                       |
                                   mark open (TUI 进程)
```

- **MCP server** (`internal/mcp/server.go`): `mark mcp` 子命令启动，通过 stdio 与 Claude Code 通信，提供 11 个工具
- **IPC 通信**: NDJSON over Unix socket (`internal/ipc/`)，TUI 侧通过 `ipc.go` 的 `waitIpcCmd` 接收请求
- **文件监听**: `fsnotify` 监听 md 文件变化（100ms debounce），Claude 改文件后 Mark 自动刷新
- **降级模式**: TUI 未运行时，MCP server 直接读写 `.markcli.json` 文件

### 关键复杂度

- **ANSI 与字符宽度**: `ansi/ansi.go` 处理 ANSI 转义序列剥离和 CJK 双宽字符的终端列 → 字符索引映射
- **分段渲染**: 同一行可能同时有选择高亮和批注高亮，`BuildSegments()` 合并多个区间后分段着色
- **鼠标协议**: `mouse.go` 处理 SGR 1006 模式鼠标事件（点击/拖拽/双击/滚轮）

## Conventions

- Go 标准代码风格（gofmt）
- 中文用户界面文案
- 测试文件: `<name>_test.go`，与源文件同目录
- 内部包全部在 `internal/` 下，对外仅暴露 `cmd/mark`
- 批注存储文件: `.markcli.json`（与被批注的 md 文件同目录）

### MCP 工具命名规范

格式: `{verb}_{resource}`，下划线分隔。操作单条用单数，操作集合用复数。

**标准动词（封闭集）**:
- 查询: `get`（单个/当前）、`list`（多条）
- 写入: `add`、`update`、`remove`（单条）、`clear`（全部）
- 文件: `open`、`refresh`
- 状态: `resolve`
- 导航: `jump_to`

新增工具必须从上述动词中选取，资源名与已有工具保持一致。
