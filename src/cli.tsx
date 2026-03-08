#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { render } from "ink";
import meow from "meow";
import { App } from "./app.js";
import { cleanupMouseOnExit, disableMouseTracking } from "./utils/mouse.js";
import { clearAnnotations, loadAnnotations } from "./utils/storage.js";

// 读取版本号
const __dirname = path.dirname(fileURLToPath(import.meta.url));
const pkgPath = path.resolve(__dirname, "..", "package.json");
const version = fs.existsSync(pkgPath)
	? JSON.parse(fs.readFileSync(pkgPath, "utf-8")).version
	: "unknown";

// 启动时清理残留鼠标追踪，退出时兜底关闭
disableMouseTracking();
cleanupMouseOnExit();

const cli = meow(
	`
  用法
    $ mark <文件>           直接打开文件
    $ mark <命令> <文件>

  命令
    open <文件>    打开文件进行阅读和批注
    list <文件>    查看文件的所有批注（JSON 格式）
    show <文件>    输出格式化的批注摘要
    clear <文件>   清除所有批注
    mcp            启动 MCP server（供 Claude Code 调用）

  选项
    --version, -V  显示版本号
    --completions  输出 shell 补全脚本（支持 zsh/bash/fish）

  示例
    $ mark README.md
    $ mark open README.md
    $ mark show README.md
`,
	{
		importMeta: import.meta,
		flags: {
			version: { type: "boolean", shortFlag: "V" },
			completions: { type: "string" },
		},
		autoVersion: false,
	},
);

// --version / -V
if (cli.flags.version) {
	console.log(`mark ${version}`);
	process.exit(0);
}

// --completions <shell>
if (cli.flags.completions !== undefined) {
	const shell = cli.flags.completions || "zsh";
	const completionsDir = path.resolve(__dirname, "..", "completions");
	const file = path.join(completionsDir, `mark.${shell}`);
	if (fs.existsSync(file)) {
		console.log(fs.readFileSync(file, "utf-8"));
	} else {
		console.error(
			`不支持的 shell: ${shell}（支持 zsh, bash, fish）`,
		);
		process.exit(1);
	}
	process.exit(0);
}

let [command, filePath] = cli.input;

if (!command) {
	cli.showHelp();
	process.exit(0);
}

// `mark mcp` 启动 MCP server（供 Claude Code 调用）
if (command === "mcp") {
	const { StdioServerTransport } = await import(
		"@modelcontextprotocol/sdk/server/stdio.js"
	);
	const { createMcpServer } = await import("./mcp/server.js");
	const server = createMcpServer();
	const transport = new StdioServerTransport();
	await server.connect(transport);
	// MCP server 通过 stdio 持续运行，永不 resolve
	await new Promise<void>(() => {});
}

// 支持 `mark README.md` 直接打开（省略 open）
if (!["open", "list", "show", "clear"].includes(command)) {
	filePath = command;
	command = "open";
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
		const { waitUntilExit } = render(
			<App filePath={resolvedPath} content={content} />,
		);
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
					? `${anno.selectedText.substring(0, 60)}...`
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
