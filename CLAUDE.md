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
npm run build          # TypeScript → dist/ (tsc)
npm run dev -- open <file>   # 开发模式运行 (tsx)
npm start -- open <file>     # 运行编译产物
npm run dev -- debug <component> [file]  # 单组件调试 (textinput|annotation|viewer|statusbar)
```

CLI 命令: `open <file>` | `list <file>` | `show <file>` | `clear <file>`

MCP server: `mark-mcp`（独立进程，Claude Code 后台自动运行）

## Test & Lint

```bash
npm test               # vitest run（运行全部测试）
npm run test:watch     # vitest 监听模式
npx vitest run src/utils/ranges.test.ts  # 运行单个测试文件
npm run lint           # biome check src/
npm run lint:fix       # biome check src/ --write（自动修复）
```

测试文件与源文件同目录，命名 `<name>.test.ts`。

## Build & Release

```bash
npm run build:binary             # esbuild 打包 + bun compile 当前平台二进制
npm run build:binary:all         # 构建全部平台 (darwin-arm64/x64, linux-x64/arm64)
node scripts/release.mjs patch   # 发版: patch|minor|major|x.y.z
```

## Architecture

使用 **React + Ink** 构建终端 UI。ESM 模块 (`"type": "module"`)，TSX 编译目标 `react-jsx`。

### 四种应用模式

`reading` → `selecting` → `annotating`，加上 `deleting` 模式，由 `app.tsx` 中的 `mode` 状态驱动。鼠标拖拽可直接从 reading 跳到 annotating；双击批注可进入编辑模式。

### 核心数据流

1. **Markdown 渲染**: `cli-markdown` 将 Markdown 转为 ANSI 终端文本 → 按行分割 → `MarkdownViewer` 渲染
2. **选择**: `useMouse`(SGR 鼠标协议) 和 `useKeyboard` 产生选择事件 → `useSelection` 管理选区状态 → `ranges.ts` 的 `buildSegments()` 将行切分为高亮/普通片段
3. **批注输出**: 退出时将所有批注以引用+指令格式输出到 stdout，供 Claude Code 消费

### MCP 集成（双进程架构）

```
Claude Code  <--stdio JSON-RPC-->  mark-mcp (MCP server 进程)
                                       |
                                   Unix socket (/tmp/mark-<uid>.sock)
                                       |
                                   mark (TUI 进程)
```

- **MCP server** (`src/mcp-cli.ts` → `src/mcp/server.ts`): 独立进程，通过 stdio 与 Claude Code 通信，提供 `list_annotations`、`get_status`、`get_selection`、`open_file`、`refresh_file`、`add_annotation`、`update_annotation`、`remove_annotation`、`resolve_annotation`、`clear_annotations` 等 tools
- **IPC 通信**: NDJSON over Unix socket (`src/mcp/ipc-server.ts` + `ipc-client.ts`)，TUI 侧通过 `useIpcServer` hook 暴露状态
- **文件监听**: `useFileWatcher` hook 监听 md 文件变化（100ms debounce），Claude 改文件后 Mark 自动刷新
- **降级模式**: TUI 未运行时，`list_annotations` 直接读 `.markcli.json` 文件

配置: `claude mcp add mark -- mark-mcp`

### 关键复杂度

- **ANSI 与字符宽度**: `ranges.ts` 处理 ANSI 转义序列剥离和 CJK 双宽字符的终端列 → 字符索引映射
- **分段渲染**: 同一行可能同时有选择高亮和批注高亮，`buildSegments()` 合并多个区间后分段着色
- **鼠标协议**: `useMouse.ts` 启用 SGR 1006 模式解析终端鼠标事件（点击/拖拽/滚轮），需在退出时正确关闭追踪

## Conventions

- 所有内部导入使用 `.js` 扩展名（ESM + NodeNext 要求）
- 中文用户界面文案
- Hook 命名: `use<Feature>.ts`，组件: `<Name>.tsx`
- 格式化: Biome，tab 缩进，双引号
- 批注存储文件: `.markcli.json`（与被批注的 md 文件同目录）

### MCP 工具命名规范

格式: `{verb}_{resource}`，下划线分隔。操作单条用单数，操作集合用复数。

**标准动词（封闭集）**:
- 查询: `get`（单个/当前）、`list`（多条）
- 写入: `add`、`update`、`remove`（单条）、`clear`（全部）
- 文件: `open`、`refresh`
- 状态: `resolve`

新增工具必须从上述动词中选取，资源名与已有工具保持一致。
