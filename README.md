# MarkCLI

终端 Markdown 批注工具 —— 在命令行中阅读 Markdown 并为文本添加批注。

![Go](https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white)
![Bubbletea](https://img.shields.io/badge/Bubbletea-FF75B5?logo=go&logoColor=white)

## 特性

- **终端原生** — 基于 [Bubbletea](https://github.com/charmbracelet/bubbletea) 构建，单二进制分发，无需运行时依赖
- **鼠标选择** — 支持鼠标拖拽选中文本，也可用键盘操作
- **字符级精度** — 正确处理 ANSI 转义序列和双宽字符（中日韩文字），预换行确保长行渲染一致
- **文本锚定** — 采用 W3C Web Annotation 标准，文件修改后批注自动重定位
- **JSON 持久化** — 批注保存在 `.markcli.json` 文件中，可版本控制
- **MCP 集成** — 内置 MCP server，Claude Code 可实时读取批注、查询状态、控制打开文件
- **自动刷新** — Claude 编辑文件后，Mark 自动检测变化并刷新内容

## 安装

```bash
curl -fsSL https://raw.githubusercontent.com/aporicho/markcli/main/install.sh | bash
```

自动检测系统和架构，下载对应二进制到 `~/.local/bin`，无需任何运行时依赖。

支持平台：macOS (Apple Silicon / Intel)、Linux (x64 / ARM64)。

## 使用

```bash
# 直接打开文件
mark README.md

# 也可以显式指定命令
mark open README.md

# 查看文件的所有批注（JSON）
mark list README.md

# 输出格式化的批注摘要
mark show README.md

# 清除所有批注
mark clear README.md

# 检查环境配置
mark doctor

# 自动更新
mark update
```

### 从源码运行

```bash
git clone https://github.com/aporicho/markcli.git
cd markcli
go run ./cmd/mark open README.md
```

## 快捷键

### 阅读模式

| 操作 | 功能 |
|------|------|
| `↑` / `↓` | 滚动 |
| `PgUp` / `PgDn` | 翻页 |
| 鼠标滚轮 | 滚动 |
| `v` | 进入选择模式 |
| 鼠标拖拽 | 直接选择文本 |
| 双击批注 | 编辑已有批注 |
| `d` | 进入总览模式 |
| `q` | 退出 |

### 选择模式

| 操作 | 功能 |
|------|------|
| `↑` / `↓` | 移动选择端点（行） |
| `←` / `→` 或 `h` / `l` | 移动选择端点（列） |
| `a` / `Enter` | 确认选择，进入批注 |
| `Esc` | 取消选择 |

### 批注模式

| 操作 | 功能 |
|------|------|
| 输入文字 | 编写批注内容 |
| `Enter` | 提交批注 |
| `Ctrl+J` | 换行 |
| `Esc` | 取消批注 |

### 总览模式

| 操作 | 功能 |
|------|------|
| `↑` / `↓` | 选择批注（Markdown 视图自动滚动） |
| `Enter` / `e` | 编辑选中批注 |
| `⌫` / `x` | 删除选中批注 |
| `Esc` | 返回阅读模式 |

## Claude Code 集成（MCP）

Mark 内置 MCP server，让 Claude Code 可以读取你的批注。使用一键安装脚本会自动完成配置；如需手动配置：

```bash
claude mcp add mark -- mark mcp
```

### 工作流

1. 在一个终端运行 `mark README.md`，划线、写批注
2. 在另一个终端的 Claude Code 中告诉 Claude 去看你的批注，Claude 会调用 MCP 工具获取内容
3. Claude 根据批注精准修改代码
4. Mark 自动检测文件变化并刷新显示

### MCP Tools

| Tool | 功能 | 参数 |
|------|------|------|
| `get_status` | 查询 Mark 运行状态、当前文件、批注数 | 无 |
| `get_selection` | 获取用户当前选中的文本 | 无 |
| `list_annotations` | 读取所有批注 | `file`（可选） |
| `add_annotation` | 添加批注 | `selectedText`, `comment`, `file`（可选） |
| `update_annotation` | 更新批注内容 | `id`, `comment`, `file`（可选） |
| `remove_annotation` | 删除批注 | `id`, `file`（可选） |
| `resolve_annotation` | 标记批注为已处理 | `id`, `file`（可选） |
| `clear_annotations` | 清除所有批注 | `file`（可选） |
| `open_file` | 让 Mark 打开指定文件 | `path` |
| `refresh_file` | 刷新当前文件 | 无 |
| `jump_to_annotation` | 滚动到指定批注位置 | `id` |

## 文本锚定

传统批注工具以行号定位，文件编辑后批注就会错位。MarkCLI 采用 W3C Web Annotation 文本锚定方案：

```json
{
  "quote": "被选中的原文",
  "prefix": "选中文本之前的上下文",
  "suffix": "选中文本之后的上下文"
}
```

通过 quote + prefix + suffix 三重上下文匹配（含模糊匹配回退），即使文件内容发生变化，批注仍能自动定位到正确位置。

## 项目结构

```
cmd/mark/main.go              # CLI 入口，所有子命令 (cobra)
internal/
├── tui/                       # TUI 核心 (bubbletea)
│   ├── model.go               # Model 定义
│   ├── update.go              # Update() 顶层分派
│   ├── view.go                # View() 组合渲染
│   ├── keys.go                # 键盘处理（模式感知）
│   ├── mouse.go               # SGR 1006 鼠标处理
│   ├── annotate.go            # 批注提交/删除/编辑
│   ├── ipc.go                 # IPC 请求处理
│   ├── file.go                # 文件监听
│   └── ui/                    # 纯渲染函数
│       ├── viewer.go          # Markdown 视图
│       ├── input.go           # 批注输入面板
│       ├── overview.go        # 总览面板
│       ├── statusbar.go       # 状态栏
│       └── segments.go        # 分段高亮算法
├── annotation/                # 批注数据、存储、文本锚定
├── markdown/                  # Markdown 渲染 (glamour)
├── ansi/                      # ANSI 转义序列处理
├── ipc/                       # Unix socket NDJSON 协议
├── mcp/                       # MCP server (JSON-RPC stdio)
├── theme/                     # 四套预置主题
└── config/                    # 用户配置
```

## 技术栈

| 技术 | 用途 |
|------|------|
| [Bubbletea](https://github.com/charmbracelet/bubbletea) | 终端 TUI 框架 |
| [Lipgloss](https://github.com/charmbracelet/lipgloss) | 终端样式 |
| [Glamour](https://github.com/charmbracelet/glamour) | Markdown 终端渲染 |
| [Cobra](https://github.com/spf13/cobra) | CLI 命令解析 |
| [mcp-go](https://github.com/mark3labs/mcp-go) | MCP server 实现 |
| [fsnotify](https://github.com/fsnotify/fsnotify) | 文件变化监听 |
| [go-runewidth](https://github.com/mattn/go-runewidth) | ���符宽度计算 |

## License

ISC
