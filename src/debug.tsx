#!/usr/bin/env node
import React, { useState } from "react";
import { render, Box, Text, useInput, useStdout } from "ink";
import meow from "meow";
import fs from "node:fs";
import path from "node:path";
import { TextInput } from "./components/TextInput.js";
import { AnnotationInput } from "./components/AnnotationInput.js";
import { MarkdownViewer } from "./components/MarkdownViewer.js";
import { StatusBar } from "./components/StatusBar.js";
import { renderMarkdownWrapped } from "./utils/markdown.js";
import { disableMouseTracking, cleanupMouseOnExit } from "./utils/mouse.js";

// ============================================================
// 1. TextInput 单独测试
// ============================================================
function DebugTextInput() {
  const [value, setValue] = useState("");
  const [submitted, setSubmitted] = useState("");

  useInput((_input, key) => {
    if (key.escape) process.exit(0);
  });

  return (
    <Box flexDirection="column" padding={1}>
      <Text bold color="cyan">[ TextInput 组件测试 ]</Text>
      <Text dimColor>打字测试，Enter 提交，Esc 退出</Text>
      <Text> </Text>
      <Box>
        <Text>输入: </Text>
        <TextInput
          value={value}
          onChange={setValue}
          onSubmit={(val) => { setSubmitted(val); setValue(""); }}
          placeholder="输入点什么..."
        />
      </Box>
      <Text> </Text>
      <Text dimColor>value: &quot;{value}&quot; | len: {value.length}</Text>
      {submitted ? <Text color="green">已提交: {submitted}</Text> : null}
    </Box>
  );
}

// ============================================================
// 2. AnnotationInput 单独测试
// ============================================================
function DebugAnnotationInput() {
  const [result, setResult] = useState<string | null>(null);
  const [show, setShow] = useState(true);

  React.useEffect(() => {
    disableMouseTracking();
  }, []);

  if (!show) {
    return (
      <Box flexDirection="column" padding={1}>
        <Text bold color="cyan">[ AnnotationInput 组件测试 ]</Text>
        <Text> </Text>
        {result === null
          ? <Text color="yellow">已取消（Esc）</Text>
          : <Text color="green">已提交: {result}</Text>}
        <Text> </Text>
        <Text dimColor>按任意键重新打开，Ctrl+C 退出</Text>
      </Box>
    );
  }

  return (
    <Box flexDirection="column" padding={1}>
      <Text bold color="cyan">[ AnnotationInput 组件测试 ]</Text>
      <Text dimColor>模拟选中行 5-8，输入批注后 Enter 提交 / Esc 取消</Text>
      <Text> </Text>
      <AnnotationInput
        selectedText="这是一段测试选中文本"
        onSubmit={(comment) => { setResult(comment); setShow(false); }}
        onCancel={() => { setResult(null); setShow(false); }}
      />
    </Box>
  );
}

// ============================================================
// 3. MarkdownViewer 单独测试
// ============================================================
function DebugMarkdownViewer({ filePath }: { filePath: string }) {
  const { stdout } = useStdout();
  const viewportHeight = (stdout?.rows ?? 24) - 3;
  const content = fs.readFileSync(filePath, "utf-8");
  const lines = renderMarkdownWrapped(content, stdout?.columns ?? 80);
  const [lastEvent, setLastEvent] = useState("(等待操作)");

  return (
    <Box flexDirection="column" height={stdout?.rows ?? 24}>
      <MarkdownViewer
        lines={lines}
        viewportHeight={viewportHeight}
        annotations={[]}
        active={true}
        onSelect={(s, e, sc, ec) => setLastEvent(`onSelect(${s}:${sc}, ${e}:${ec})`)}
        onQuit={() => { process.exit(0); }}
      />
      <Text backgroundColor="gray" color="white">
        {" "}MarkdownViewer 测试 | 最后事件: {lastEvent} | q:退出 v:选中 a:确认选中{" "}
      </Text>
    </Box>
  );
}

// ============================================================
// 4. StatusBar 单独测试
// ============================================================
function DebugStatusBar() {
  const modes = ["reading", "selecting", "annotating"] as const;
  const [idx, setIdx] = useState(0);

  useInput((input, key) => {
    if (input === "q" || key.escape) process.exit(0);
    if (key.return || input === " ") setIdx((i) => (i + 1) % modes.length);
  });

  return (
    <Box flexDirection="column" padding={1}>
      <Text bold color="cyan">[ StatusBar 组件测试 ]</Text>
      <Text dimColor>Enter/空格 切换模式，q 退出</Text>
      <Text> </Text>
      <StatusBar
        mode={modes[idx]!}
        currentLine={15}
        totalLines={42}
        selectionStart={modes[idx] !== "reading" ? 10 : null}
        selectionEnd={modes[idx] !== "reading" ? 12 : null}
        annotationCount={3}
      />
    </Box>
  );
}

// ============================================================
// CLI
// ============================================================
const cli = meow(
  `
  用法
    $ markcli debug <组件名> [文件]

  组件名
    textinput       TextInput 文字输入组件
    annotation      AnnotationInput 批注输入面板
    viewer          MarkdownViewer 阅读器（需要文件参数）
    statusbar       StatusBar 状态栏

  示例
    $ markcli debug textinput
    $ markcli debug viewer test.md
    $ markcli debug annotation
    $ markcli debug statusbar
`,
  {
    importMeta: import.meta,
    flags: {},
  }
);

const [component, filePath] = cli.input;

// 所有 debug 模式启动前先关闭残留鼠标追踪
disableMouseTracking();
// 退出时也兜底关闭
cleanupMouseOnExit();

if (!component) {
  cli.showHelp();
  process.exit(0);
}

switch (component) {
  case "textinput":
    render(<DebugTextInput />);
    break;
  case "annotation":
    render(<DebugAnnotationInput />);
    break;
  case "viewer": {
    const resolved = path.resolve(filePath || "test.md");
    if (!fs.existsSync(resolved)) {
      console.error(`文件不存在: ${resolved}`);
      process.exit(1);
    }
    render(<DebugMarkdownViewer filePath={resolved} />);
    break;
  }
  case "statusbar":
    render(<DebugStatusBar />);
    break;
  default:
    console.error(`未知组件: ${component}`);
    console.error("可选: textinput, annotation, viewer, statusbar");
    process.exit(1);
}
