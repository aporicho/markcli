#!/usr/bin/env node
import React from "react";
import { render } from "ink";
import meow from "meow";
import fs from "node:fs";
import path from "node:path";
import { App } from "./app.js";
import { loadAnnotations, clearAnnotations } from "./utils/storage.js";

import { disableMouseTracking, cleanupMouseOnExit } from "./utils/mouse.js";

// 启动时清理残留鼠标追踪，退出时兜底关闭
disableMouseTracking();
cleanupMouseOnExit();

const cli = meow(
  `
  用法
    $ markcli <命令> <文件>

  命令
    open <文件>    打开文件进行阅读和批注
    list <文件>    查看文件的所有批注（JSON 格式）
    show <文件>    输出格式化的批注摘要
    clear <文件>   清除所有批注

  示例
    $ markcli open README.md
    $ markcli show README.md
`,
  {
    importMeta: import.meta,
    flags: {},
  }
);

const [command, filePath] = cli.input;

if (!command) {
  cli.showHelp();
  process.exit(0);
}

if (!filePath) {
  console.error("错误：请指定文件路径");
  process.exit(1);
}

const resolvedPath = path.resolve(filePath);

if (!fs.existsSync(resolvedPath)) {
  console.error(`错误：文件不存在 - ${resolvedPath}`);
  process.exit(1);
}

switch (command) {
  case "open": {
    const content = fs.readFileSync(resolvedPath, "utf-8");
    const { waitUntilExit } = render(<App filePath={resolvedPath} content={content} />);
    waitUntilExit().then(() => {
      process.exit(0);
    });
    break;
  }

  case "list": {
    const data = loadAnnotations(resolvedPath);
    console.log(JSON.stringify(data, null, 2));
    break;
  }

  case "show": {
    const data = loadAnnotations(resolvedPath);
    if (data.annotations.length === 0) {
      console.log(`📄 File: ${path.basename(resolvedPath)}`);
      console.log("---");
      console.log("（暂无批注）");
      console.log("---");
      break;
    }
    console.log(`📄 File: ${path.basename(resolvedPath)}`);
    console.log("---");
    for (const anno of data.annotations) {
      const range =
        anno.startLine === anno.endLine
          ? `Line ${anno.startLine}`
          : `Line ${anno.startLine}-${anno.endLine}`;
      const textPreview =
        anno.selectedText.length > 60
          ? anno.selectedText.substring(0, 60) + "..."
          : anno.selectedText;
      console.log(`[${range}] "${textPreview}"`);
      console.log(`💬 批注: ${anno.comment}`);
      console.log();
    }
    console.log("---");
    break;
  }

  case "clear": {
    clearAnnotations(resolvedPath);
    console.log(`已清除 ${path.basename(resolvedPath)} 的所有批注`);
    break;
  }

  default:
    console.error(`未知命令: ${command}`);
    cli.showHelp();
    process.exit(1);
}
