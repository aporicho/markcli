// MCP Server 主逻辑：注册 tools，通过 IPC 连接 TUI

import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { z } from "zod";
import { loadAnnotations } from "../utils/storage.js";
import { IpcClient } from "./ipc-client.js";

export function createMcpServer(): McpServer {
	const server = new McpServer({
		name: "mark",
		version: "1.0.0",
	});

	const ipc = new IpcClient();

	// get_annotations: 读取当前批注
	server.tool(
		"get_annotations",
		"读取 Mark 中的批注。返回用户对文件内容的划线批注，包括选中的文本和批注内容。",
		{
			file: z
				.string()
				.optional()
				.describe("文件路径（可选，默认为当前打开的文件）"),
		},
		async ({ file }) => {
			// 先尝试通过 IPC 从运行中的 TUI 获取
			const connected = await ipc.isConnected();
			if (connected) {
				try {
					const res = await ipc.send({
						type: "get_annotations",
						file,
					});
					if (res.type === "annotations") {
						const { data } = res;
						if (data.annotations.length === 0) {
							return {
								content: [
									{
										type: "text",
										text: `## Annotations for ${data.file}\n\n（暂无批注）`,
									},
								],
							};
						}
						let md = `## Annotations for ${data.file}\n\n`;
						for (const ann of data.annotations) {
							md += `> "${ann.selectedText}"\n\n`;
							md += `批注: ${ann.comment}\n\n---\n\n`;
						}
						return { content: [{ type: "text", text: md.trim() }] };
					}
					if (res.type === "error") {
						return {
							content: [{ type: "text", text: res.message }],
							isError: true,
						};
					}
				} catch {
					// IPC 失败，尝试降级
				}
			}

			// 降级：直接读文件
			if (file) {
				try {
					const data = loadAnnotations(file);
					if (data.annotations.length === 0) {
						return {
							content: [
								{
									type: "text",
									text: `## Annotations for ${file}\n\n（暂无批注）`,
								},
							],
						};
					}
					let md = `## Annotations for ${file}\n\n`;
					for (const ann of data.annotations) {
						md += `> "${ann.selectedText}"\n\n`;
						md += `批注: ${ann.comment}\n\n---\n\n`;
					}
					return { content: [{ type: "text", text: md.trim() }] };
				} catch {
					return {
						content: [
							{
								type: "text",
								text: "Mark is not running and no file specified for fallback.",
							},
						],
						isError: true,
					};
				}
			}

			return {
				content: [
					{
						type: "text",
						text: "Mark is not running. Please start Mark with `mark <file>` first.",
					},
				],
				isError: true,
			};
		},
	);

	// get_file_status: 查询 Mark 状态
	server.tool(
		"get_file_status",
		"查询 Mark 的运行状态，包括当前打开的文件、批注数量和模式。",
		{},
		async () => {
			const connected = await ipc.isConnected();
			if (!connected) {
				return {
					content: [{ type: "text", text: "Mark is not running" }],
				};
			}

			try {
				const res = await ipc.send({ type: "get_status" });
				if (res.type === "status") {
					const { data } = res;
					const text = `File: ${data.file}\nAnnotations: ${data.annotationCount}\nMode: ${data.mode}`;
					return { content: [{ type: "text", text }] };
				}
				return {
					content: [
						{ type: "text", text: "Unexpected response from Mark" },
					],
					isError: true,
				};
			} catch {
				return {
					content: [
						{ type: "text", text: "Failed to connect to Mark" },
					],
					isError: true,
				};
			}
		},
	);

	// open_file: 让 Mark 打开指定文件
	server.tool(
		"open_file",
		"让 Mark 打开指定文件。Mark 必须已经在运行。",
		{
			path: z.string().describe("要打开的文件路径"),
		},
		async ({ path: filePath }) => {
			const connected = await ipc.isConnected();
			if (!connected) {
				return {
					content: [
						{
							type: "text",
							text: "Mark is not running. Please start Mark with `mark <file>` first.",
						},
					],
					isError: true,
				};
			}

			try {
				const res = await ipc.send({
					type: "open_file",
					path: filePath,
				});
				if (res.type === "ok") {
					return {
						content: [{ type: "text", text: res.message }],
					};
				}
				if (res.type === "error") {
					return {
						content: [{ type: "text", text: res.message }],
						isError: true,
					};
				}
				return {
					content: [
						{ type: "text", text: "Unexpected response from Mark" },
					],
					isError: true,
				};
			} catch {
				return {
					content: [
						{ type: "text", text: "Failed to connect to Mark" },
					],
					isError: true,
				};
			}
		},
	);

	return server;
}
