# MarkCLI

终端 Markdown 批注工具 —— 在命令行中阅读 Markdown 并为文本添加批注。

![TypeScript](https://img.shields.io/badge/TypeScript-3178C6?logo=typescript&logoColor=white)
![React](https://img.shields.io/badge/React-61DAFB?logo=react&logoColor=black)
![Node.js](https://img.shields.io/badge/Node.js-339933?logo=node.js&logoColor=white)

## 特性

- **终端原生** — 基于 [Ink](https://github.com/vadimdemedes/ink) 构建，无需离开终端
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

自动检测系统和架构，下载对应二进制到 `~/.local/bin`，无需 Node.js。

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
```

### 从源码运行

```bash
git clone https://github.com/aporicho/markcli.git
cd markcli && npm install
npm run dev -- open README.md
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
| `Esc` | 取消批注 |
| `Ctrl+R` | 重新选择文本范围（编辑时） |
| `Ctrl+D` | 删除当前批注（编辑时） |

### 总览模式

| 操作 | 功能 |
|------|------|
| `↑` / `↓` | 选择批注（Markdown 视图自动滚动） |
| `Enter` | 编辑选中批注 |
| `⌫` / `d` | 删除选中批注 |
| `Esc` | 返回阅读模式 |

## Claude Code 集成（MCP）

Mark 内置 MCP server，让 Claude Code 可以读取你的批注。使用一键安装脚本会自动完成配置；如需手动配置：

```bash
claude mcp add mark -- mark mcp
```

### 工作流

1. 在一个终端运行 `mark README.md`，划线、写批注
2. 在另一个终端的 Claude Code 中告诉 Claude 去看你的批注，Claude 会调用 `get_annotations` 获取内容
3. Claude 根据批注精准修改代码
4. Mark 自动检测文件变化并刷新显示

你也可以让 Claude 控制 Mark 打开其他文件，或者查询 Mark 当前的运行状态。

### MCP Tools

| Tool | 功能 | 参数 |
|------|------|------|
| `get_annotations` | 读取批注（选中文本 + 批注内容） | `file`（可选，默认当前文件） |
| `get_file_status` | 查询 Mark 运行状态、当前文件、批注数 | 无 |
| `open_file` | 让 Mark 切换到指定文件 | `path`（文件路径） |

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
src/
├── cli.tsx                 # CLI 入口，命令解析
├── mcp-cli.ts              # MCP server 入口（独立进程）
├── app.tsx                 # 主应用，状态与模式管理
├── debug.tsx               # 单组件调试入口
├── types.ts                # 类型定义
├── components/
│   ├── MarkdownViewer.tsx  # Markdown 渲染 + 滚动 + 高亮
│   ├── AnnotationInput.tsx # 批注输入面板
│   ├── StatusBar.tsx       # 底部状态栏
│   ├── OverviewPanel.tsx   # 总览模式浮动面板
│   ├── LineSelector.tsx    # 选择区间工具
│   └── TextInput.tsx       # 文字输入组件
├── hooks/
│   ├── useAnnotations.ts   # 批注 CRUD + 持久化
│   ├── useFileWatcher.ts   # 文件变化监听（自动刷新）
│   ├── useIpcServer.ts     # IPC socket server（暴露状态给 MCP）
│   ├── useKeyboard.ts      # 键盘事件处理
│   ├── useSelection.ts     # 选择状态管理
│   └── useMouse.ts         # 终端鼠标追踪 (SGR 协议)
├── mcp/
│   ├── protocol.ts         # IPC 消息类型定义
│   ├── socket-path.ts      # Unix socket 路径计算
│   ├── ipc-server.ts       # TUI 侧 socket server
│   ├── ipc-client.ts       # MCP 侧 socket client
│   └── server.ts           # MCP server（注册 tools）
└── utils/
    ├── storage.ts          # 文件存储 (.markcli.json)
    ├── textAnchor.ts       # 文本锚定与重定位
    ├── markdown.ts         # Markdown → 终端渲染（含预换行）
    ├── mouse.ts            # 鼠标协议开关工具
    └── ranges.ts           # ANSI 处理 + 分段高亮
```

## 技术栈

| 技术 | 用途 |
|------|------|
| [Ink](https://github.com/vadimdemedes/ink) | 终端 React UI 框架 |
| [TypeScript](https://www.typescriptlang.org/) | 类型安全 |
| [cli-markdown](https://github.com/pmuens/cli-markdown) | Markdown 终端渲染 |
| [wrap-ansi](https://github.com/chalk/wrap-ansi) | ANSI 感知的文本换行 |
| [approx-string-match](https://github.com/nickstenning/approx-string-match) | 模糊文本匹配 |
| [meow](https://github.com/sindresorhus/meow) | CLI 参数解析 |
| [@modelcontextprotocol/sdk](https://github.com/modelcontextprotocol/typescript-sdk) | MCP server 实现 |

## License

ISC
