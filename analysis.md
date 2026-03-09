## MarkCLI 全面分析报告

经过对全部源码的深度审查，以下按 **架构 → 健壮性 → 性能 → 体验 → 简化** 五个维度，整理出可行的改进方向。

---

### 一、架构：app.tsx 状态管理失控

**现状**：`app.tsx` 449 行，9 个 `useState`、9 个 `useMemo`、多个 `useCallback`，是全项目最大的技术债。

| 问题 | 影响 |
|------|------|
| 9 个独立状态变量高度耦合 | 改一个常需同步改多个，极易产生不一致 |
| 7 条模式转换路径分散在不同 handler 中 | 无法一眼看清状态机全貌 |
| `handleAnnotationSubmit` 80+ 行、5 层嵌套 | 编辑 vs 新增 vs 重选，分支爆炸 |
| 数据转换链 annotations → relocate → resolve → display | 每步一个 useMemo，数据流难追踪 |

**建议**：用 `useReducer` 合并状态，定义显式的 action/transition 表：

```typescript
type Action =
  | { type: "START_SELECT"; pos: SelectionPos }
  | { type: "CONFIRM_SELECT" }
  | { type: "SUBMIT_ANNOTATION"; comment: string }
  | { type: "ENTER_OVERVIEW" }
  | { type: "EDIT_ANNOTATION"; id: string }
  // ...
```

将 `handleAnnotationSubmit` 拆为 `submitNew` / `submitEdit` / `submitReselect` 三个函数。

---

### 二、健壮性：错误处理几乎为零

整个项目的错误处理是最薄弱的环节：

| 位置 | 风险 | 影响 |
|------|------|------|
| `useAnnotations` — save 失败不回滚 | 状态与文件不一致 | 批注丢失 |
| `ipc-server.ts` — 连接无 `end`/`error` 监听 | 僵尸连接堆积 | 内存泄漏 |
| `ipc-server.ts` — socket 文件权限默认 0o644 | 其他用户可读写 | 安全隐患 |
| `socket-path.ts` — 单一 socket 路径 | 多实例互相覆盖 | 第二个 TUI 无法通信 |
| `ipc-client.ts` — NDJSON 按 `\n` split | 批注含换行时消息截断 | IPC 通信失败 |
| `useFileWatcher` — catch 块为空 | 权限错误被吞掉 | 文件变更不触发刷新 |
| `storage.ts` — 直接 writeFileSync | 写入中断导致文件损坏 | 批注全丢 |

**优先修复建议**：

1. **IPC 连接管理** — 添加 `conn.on("end"/"error")` + socket 权限 `0o600`
2. **原子写入** — 写临时文件再 `rename`，防止断电损坏
3. **多实例支持** — socket 路径加入文件 hash 或 PID：`mark-<uid>-<hash>.sock`
4. **save 失败回滚** — `saveAnnotations` 包裹 try-catch，失败时还原 state

---

### 三、性能：关键渲染路径有瓶颈

| 瓶颈 | 位置 | 复杂度 | 改进 |
|------|------|--------|------|
| `stripAnsi` 重复调用 | MarkdownViewer 每行调一次，useSelection 的 `clampCol` 又调一次 | 浪费 | 缓存 strippedLines，传入各处复用 |
| `buildSegments` 对每段遍历全部 annRanges | ranges.ts | O(segments × annotations) | 已排序可二分查找 |
| `termColToCharIndex` 逐字符计算宽度 | ranges.ts | O(行长) | 预计算宽度映射表 |
| `resolvedAnnotations` 每次全量 relocate | app.tsx | O(批注数 × 文本长度) | 仅对变更的批注重定位 |
| visibleLines 无 memo | MarkdownViewer 每帧重新创建行组件 | 无谓重渲染 | React.memo 包装行组件 |
| MCP server markdown 拼接用 `+=` | mcp/server.ts | O(n²) 字符串分配 | 改为数组 push + join |

其中 `stripAnsi` 重复调用影响最广，一次修复受益多处。

---

### 四、使用体验提升

#### 4.1 交互改进

| 改进点 | 现状 | 建议 |
|------|------|------|
| 空批注列表提示 | "暂无批注" | 改为"选中文本后按 a 添加批注" |
| 批注提交反馈 | 静默回到 reading | 短暂显示"批注已保存"提示 |
| 错误反馈 | 异常被吞、无提示 | StatusBar 显示最近一条错误 |
| 文件变更通知 | 静默刷新 | StatusBar 闪烁提示"文件已更新" |
| 长按方向键 | 逐行移动，无加速 | 长按加速或支持数字前缀（vim 风格 `5j`） |

#### 4.2 MCP 集成优化

| 改进点 | 现状 | 建议 |
|------|------|------|
| IPC 双重超时 | isConnected(1s) + send(3s) = 最多 4s | 合并为一次 send(3s)，省去冗余 ping |
| 多实例 | 新实例覆盖旧 socket | 支持多文件多 socket，MCP 按文件路由 |
| 批注格式 | 纯 markdown 文本 | 同时返回结构化 JSON，让 Claude 更精准定位 |

#### 4.3 键位可配置化

当前所有快捷键硬编码在 `useKeyboard.ts`（`q`/`v`/`d`/`a`/`h`/`l`）。可以提取为配置对象，支持用户自定义。

---

### 五、代码简化

| 简化项 | 当前 | 改进 |
|------|------|------|
| SGR 鼠标解析重复 | `useMouse.ts` 和 `AnnotationInput.tsx` 各写一遍 | 提取 `utils/sgrMouse.ts` 共用 |
| MCP 降级逻辑重复 | `get_annotations` 的 IPC 路径和 fallback 路径几乎相同 | 提取 `formatAnnotationsMarkdown()` 公用 |
| 魔数散落各处 | 400ms(双击)、50ms(delay)、100ms(debounce)、25/60(截断) | 集中到 `constants.ts` |
| StatusBar 配置分散 | `MODE_COLORS`、`MODE_LABELS`、`SHORTCUTS` 三个独立 Record | 合并为 `MODE_CONFIG: Record<AppMode, {color, label, shortcuts}>` |
| `mouse.ts` ANSI 序列 | 硬编码 `\x1b[?1000h...` | 定义命名常量 + 注释说明用途 |

---

### 优先级排序

| 优先级 | 方向 | 收益 | 工作量 |
|--------|------|------|--------|
| **P0** | IPC 连接泄漏 + socket 权限 | 防止崩溃和安全漏洞 | 小 |
| **P0** | storage 原子写入 | 防止数据丢失 | 小 |
| **P1** | app.tsx → useReducer 重构 | 大幅降低维护成本 | 中 |
| **P1** | stripAnsi 缓存 + 复用 | 全局性能提升 | 小 |
| **P2** | SGR 解析提取复用 | 消除重复代码 | 小 |
| **P2** | 常量集中化 | 可维护性提升 | 小 |
| **P2** | 错误反馈 UI | 用户体验提升 | 中 |
| **P3** | 多实例支持 | 高级使用场景 | 中 |
| **P3** | 键位配置化 | 用户自定义 | 中 |
