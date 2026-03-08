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

const cli = meow(
	`
  用法
    $ mark                  选择当前目录的 .md 文件打开
    $ mark <文件>           直接打开文件
    $ mark <命令> <文件>

  命令
    open <文件>    打开文件进行阅读和批注
    list <文件>    查看文件的所有批注（JSON 格式）
    show <文件>    输出格式化的批注摘要
    clear <文件>   清除所有批注
    mcp            启动 MCP server（供 Claude Code 调用）
    update         检查并更新到最新版本
    doctor         检查环境配置是否正常

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
		console.error(`不支持的 shell: ${shell}（支持 zsh, bash, fish）`);
		process.exit(1);
	}
	process.exit(0);
}

let [command, filePath] = cli.input;

// 无参数时：列出当前目录的 .md 文件让用户选择
if (!command) {
	const mdFiles = fs
		.readdirSync(".")
		.filter((f) => f.endsWith(".md") && fs.statSync(f).isFile())
		.sort();
	if (mdFiles.length === 0) {
		console.error("当前目录没有 Markdown 文件");
		process.exit(1);
	}
	if (mdFiles.length === 1) {
		command = "open";
		filePath = mdFiles[0];
		console.log(`打开 ${filePath}`);
	} else {
		const readline = await import("node:readline");
		const rl = readline.createInterface({
			input: process.stdin,
			output: process.stdout,
		});
		console.log("选择要打开的文件：\n");
		for (let i = 0; i < mdFiles.length; i++) {
			console.log(`  ${i + 1}) ${mdFiles[i]}`);
		}
		const answer = await new Promise<string>((resolve) => {
			rl.question("\n输入编号: ", resolve);
		});
		rl.close();
		const idx = Number.parseInt(answer, 10) - 1;
		if (idx < 0 || idx >= mdFiles.length || Number.isNaN(idx)) {
			console.error("无效选择");
			process.exit(1);
		}
		command = "open";
		filePath = mdFiles[idx];
	}
}

// `mark doctor` 检查环境配置
if (command === "doctor") {
	const { execSync } = await import("node:child_process");
	let ok = true;

	// 版本
	console.log(`mark ${version}`);
	console.log();

	// PATH
	const binPath = process.execPath;
	const binDir = path.dirname(binPath);
	const inPath = process.env.PATH?.split(":").includes(binDir);
	console.log(
		`${inPath ? "✅" : "❌"} PATH: ${binDir} ${inPath ? "已在 PATH 中" : "不在 PATH 中"}`,
	);
	if (!inPath) ok = false;

	// Claude Code
	let hasClaude = false;
	try {
		execSync("which claude", { stdio: "pipe" });
		hasClaude = true;
	} catch {}
	console.log(
		`${hasClaude ? "✅" : "⚠️"} Claude Code: ${hasClaude ? "已安装" : "未安装"}`,
	);

	// MCP 配置
	if (hasClaude) {
		const homeDir = (await import("node:os")).homedir();
		const claudeConfig = path.join(homeDir, ".claude.json");
		let mcpOk = false;
		try {
			const config = JSON.parse(fs.readFileSync(claudeConfig, "utf-8"));
			const mcpServers =
				config?.projects?.mcpServers ?? config?.mcpServers ?? {};
			// 遍历所有项目配置查找 mark
			const allServers = { ...mcpServers };
			if (config?.projects) {
				for (const proj of Object.values(config.projects) as Record<
					string,
					unknown
				>[]) {
					if (proj?.mcpServers) Object.assign(allServers, proj.mcpServers);
				}
			}
			mcpOk = "mark" in allServers;
		} catch {}
		console.log(
			`${mcpOk ? "✅" : "❌"} MCP 集成: ${mcpOk ? "已配置" : "未配置（运行 claude mcp add mark -- mark mcp）"}`,
		);
		if (!mcpOk) ok = false;
	}

	// macOS 签名
	if (process.platform === "darwin") {
		let signOk = false;
		try {
			execSync(`codesign -v "${binPath}" 2>&1`, { stdio: "pipe" });
			signOk = true;
		} catch {}
		console.log(
			`${signOk ? "✅" : "❌"} 代码签名: ${signOk ? "有效" : `无效（运行 codesign --force --sign - ${binPath}）`}`,
		);
		if (!signOk) ok = false;
	}

	console.log();
	console.log(ok ? "一切正常 👍" : "存在问题，请按提示修复");
	process.exit(ok ? 0 : 1);
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

// `mark update` 自更新
if (command === "update") {
	const { execSync } = await import("node:child_process");
	const os = await import("node:os");

	const platform = os.platform() === "darwin" ? "darwin" : "linux";
	const arch = os.arch() === "arm64" ? "arm64" : "x64";
	const asset = `mark-${platform}-${arch}`;
	const repo = "aporicho/markcli";

	console.log(`当前版本: ${version}`);
	console.log("检查最新版本...");

	try {
		const res = execSync(
			`curl -fsSL "https://api.github.com/repos/${repo}/releases/latest"`,
			{ encoding: "utf-8" },
		);
		const latest = JSON.parse(res).tag_name as string;
		const latestVersion = latest.replace(/^v/, "");

		if (version === latestVersion) {
			console.log(`已是最新版本 (${version})`);
			process.exit(0);
		}

		console.log(`发现新版本: ${latest}`);
		const binPath = process.execPath;
		const url = `https://github.com/${repo}/releases/download/${latest}/${asset}`;

		console.log("下载中...");
		execSync(`curl -fsSL -o "${binPath}" "${url}"`, { stdio: "inherit" });
		execSync(`chmod +x "${binPath}"`);

		// macOS: 清除隔离属性并重新签名
		if (platform === "darwin") {
			execSync(
				`xattr -dr com.apple.quarantine "${binPath}" 2>/dev/null; xattr -dr com.apple.provenance "${binPath}" 2>/dev/null; codesign --force --sign - "${binPath}" 2>/dev/null`,
				{ shell: "/bin/bash" },
			);
		}

		console.log(`✅ 已更新到 ${latest}`);
	} catch (e) {
		console.error("更新失败:", (e as Error).message);
		process.exit(1);
	}
	process.exit(0);
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
		// 仅在打开 TUI 时初始化鼠标追踪
		disableMouseTracking();
		cleanupMouseOnExit();
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
