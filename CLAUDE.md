# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
npm run build          # TypeScript → dist/ (tsc)
npm run dev -- open <file>   # 开发模式运行 (tsx)
npm start -- open <file>     # 运行编译产物
```

CLI 命令: `open <file>` | `list <file>` | `show <file>` | `clear <file>`

## Architecture

终端 Markdown 批注工具，使用 **React + Ink** 构建终端 UI。ESM 模块 (`"type": "module"`)，TSX 编译目标 `react-jsx`。

### 三种应用模式

`reading` → `selecting` → `annotating`，由 `app.tsx` 中的 `mode` 状态驱动。鼠标拖拽可直接从 reading 跳到 annotating。

### 核心数据流

1. **Markdown 渲染**: `cli-markdown` 将 Markdown 转为 ANSI 终端文本 → 按行分割 → `MarkdownViewer` 渲染
2. **选择**: `useMouse`(SGR 鼠标协议) 和 `useKeyboard` 产生选择事件 → `useSelection` 管理选区状态 → `ranges.ts` 的 `buildSegments()` 将行切分为高亮/普通片段
3. **批注持久化**: 批注存为 `<file>.markcli.json`（同目录），使用 W3C Text Anchor（quote + prefix + suffix）定位，`textAnchor.ts` 的 `relocateAnchor()` 在���件变更后通过精确搜索 → 模糊匹配（approx-string-match）重新定位

### 关键复杂度

- **ANSI 与字符宽度**: `ranges.ts` 处理 ANSI 转义序列剥离和 CJK 双宽字符的终端列 → 字符索引映射
- **分段渲染**: 同一行可能同时有选择高亮和批注高亮，`buildSegments()` 合并多个区间后分段着色
- **鼠标协议**: `useMouse.ts` 启用 SGR 1006 模式解析终端鼠标事件（点击/拖拽/滚轮），需在退出时正确关闭追踪

## Conventions

- 所有内部导入使用 `.js` 扩展名（ESM + NodeNext 要求）
- 中文用户界面文案
- Hook 命名: `use<Feature>.ts`，组件: `<Name>.tsx`
