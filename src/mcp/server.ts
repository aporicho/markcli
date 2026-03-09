// MCP Server 主逻辑：注册 tools，通过 IPC 连接 TUI

import fs from "node:fs";
import { McpServer } from "@modelcontextprotocol/sdk/server/mcp.js";
import { nanoid } from "nanoid";
import { z } from "zod";
import {
	clearAnnotations,
	loadAnnotations,
	saveAnnotations,
} from "../utils/storage.js";
import {
	extractAnchor,
	offsetToLineCol,
	relocateAnchor,
} from "../utils/textAnchor.js";
import { IpcClient } from "./ipc-client.js";

export function createMcpServer(): McpServer {
	const server = new McpServer({
		name: "mark",
		version: "1.0.0",
	});

	const ipc = new IpcClient();

	// list_annotations: 列出当前批注
	server.tool(
		"list_annotations",
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
						type: "list_annotations",
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
							md += `**[${ann.id}]** > "${ann.selectedText}"\n\n`;
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

	// get_status: 查询 Mark 状态
	server.tool(
		"get_status",
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
					content: [{ type: "text", text: "Unexpected response from Mark" }],
					isError: true,
				};
			} catch {
				return {
					content: [{ type: "text", text: "Failed to connect to Mark" }],
					isError: true,
				};
			}
		},
	);

	// get_selection: 获取用户当前选中的文本
	server.tool(
		"get_selection",
		"获取用户当前在 Mark 中选中的文本。",
		{},
		async () => {
			const connected = await ipc.isConnected();
			if (!connected) {
				return {
					content: [{ type: "text", text: "Mark is not running" }],
				};
			}

			try {
				const res = await ipc.send({ type: "get_selection" });
				if (res.type === "selection") {
					const { data } = res;
					if (!data.selectedText) {
						return {
							content: [{ type: "text", text: "No text currently selected" }],
						};
					}
					return {
						content: [
							{
								type: "text",
								text: `File: ${data.file}\n\nSelected text:\n> ${data.selectedText}`,
							},
						],
					};
				}
				return {
					content: [{ type: "text", text: "Unexpected response from Mark" }],
					isError: true,
				};
			} catch {
				return {
					content: [{ type: "text", text: "Failed to connect to Mark" }],
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
					content: [{ type: "text", text: "Unexpected response from Mark" }],
					isError: true,
				};
			} catch {
				return {
					content: [{ type: "text", text: "Failed to connect to Mark" }],
					isError: true,
				};
			}
		},
	);

	// refresh_file: 通知 Mark 刷新当前文件
	server.tool(
		"refresh_file",
		"通知 Mark 重新读取当前打开的文件。当你修改了文件内容后调用此工具，Mark 会立即刷新显示。Mark 必须已经在运行。",
		{},
		async () => {
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
				const res = await ipc.send({ type: "refresh_file" });
				if (res.type === "ok") {
					return { content: [{ type: "text", text: res.message }] };
				}
				return {
					content: [
						{
							type: "text",
							text: res.type === "error" ? res.message : "Unexpected response",
						},
					],
					isError: true,
				};
			} catch {
				return {
					content: [{ type: "text", text: "Failed to connect to Mark" }],
					isError: true,
				};
			}
		},
	);

	// add_annotation: 添加批注
	server.tool(
		"add_annotation",
		"在 Mark 中添加批注。只需指定要选中的文本和批注内容，自动定位文本位置。TUI 运行时通过 IPC 通信，未运行时直接操作 .markcli.json 文件（需指定 file 参数）。",
		{
			file: z
				.string()
				.optional()
				.describe("文件路径（可选，默认为当前打开的文件）"),
			selectedText: z.string().describe("选中的原文"),
			comment: z.string().describe("批注内容"),
		},
		async ({ file, selectedText, comment }) => {
			const connected = await ipc.isConnected();
			if (connected) {
				try {
					const res = await ipc.send({
						type: "add_annotation",
						file,
						selectedText,
						comment,
					});
					if (res.type === "ok") {
						return { content: [{ type: "text", text: res.message }] };
					}
					return {
						content: [
							{
								type: "text",
								text:
									res.type === "error" ? res.message : "Unexpected response",
							},
						],
						isError: true,
					};
				} catch {
					// IPC 失败，尝试降级
				}
			}

			// 降级：直接操作文件
			if (file) {
				try {
					const mdContent = fs.readFileSync(file, "utf-8");
					const lines = mdContent.split("\n");
					const lineLengths = lines.map((l) => l.length);

					// 精确匹配
					let matchStart = mdContent.indexOf(selectedText);

					if (matchStart === -1) {
						// 模糊匹配兜底
						const range = relocateAnchor(mdContent, {
							quote: selectedText,
							prefix: "",
							suffix: "",
						});
						if (!range) {
							return {
								content: [
									{
										type: "text",
										text: `Text not found: "${selectedText.slice(0, 50)}..."`,
									},
								],
								isError: true,
							};
						}
						matchStart = range.start;
					}

					const matchEnd = matchStart + selectedText.length;
					const start = offsetToLineCol(lineLengths, matchStart);
					const end = offsetToLineCol(lineLengths, matchEnd);
					const anchor = extractAnchor(mdContent, matchStart, matchEnd);

					const annotation = {
						id: nanoid(6),
						startLine: start.line,
						endLine: end.line,
						startCol: start.col,
						endCol: end.col,
						selectedText,
						comment,
						createdAt: new Date().toISOString(),
						quote: anchor.quote,
						prefix: anchor.prefix,
						suffix: anchor.suffix,
					};

					const data = loadAnnotations(file);
					data.annotations.push(annotation);
					saveAnnotations(file, data);

					return {
						content: [{ type: "text", text: "Annotation added" }],
					};
				} catch (err) {
					return {
						content: [
							{
								type: "text",
								text: `Failed to add annotation: ${err instanceof Error ? err.message : String(err)}`,
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
						text: "Mark is not running. Please specify a file path or start Mark first.",
					},
				],
				isError: true,
			};
		},
	);

	// update_annotation: 更新批注内容
	server.tool(
		"update_annotation",
		"更新 Mark 中已有批注的内容。需要提供批注 ID 和新的批注内容。TUI 运行时通过 IPC 通信，未运行时直接操作 .markcli.json 文件（需指定 file 参数）。",
		{
			id: z.string().describe("批注 ID"),
			comment: z.string().describe("新的批注内容"),
			file: z
				.string()
				.optional()
				.describe("文件路径（可选，TUI 未运行时必须指定）"),
		},
		async ({ id, comment, file }) => {
			const connected = await ipc.isConnected();
			if (connected) {
				try {
					const res = await ipc.send({
						type: "update_annotation",
						id,
						comment,
					});
					if (res.type === "ok") {
						return { content: [{ type: "text", text: res.message }] };
					}
					return {
						content: [
							{
								type: "text",
								text:
									res.type === "error" ? res.message : "Unexpected response",
							},
						],
						isError: true,
					};
				} catch {
					// IPC 失败，尝试降级
				}
			}

			// 降级：直接操作文件
			if (file) {
				try {
					const data = loadAnnotations(file);
					const ann = data.annotations.find((a) => a.id === id);
					if (!ann) {
						return {
							content: [{ type: "text", text: `Annotation ${id} not found` }],
							isError: true,
						};
					}
					ann.comment = comment;
					saveAnnotations(file, data);
					return {
						content: [{ type: "text", text: `Annotation ${id} updated` }],
					};
				} catch (err) {
					return {
						content: [
							{
								type: "text",
								text: `Failed to update annotation: ${err instanceof Error ? err.message : String(err)}`,
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
						text: "Mark is not running. Please specify a file path or start Mark first.",
					},
				],
				isError: true,
			};
		},
	);

	// remove_annotation: 删除批注
	server.tool(
		"remove_annotation",
		"删除 Mark 中的一条批注。需要提供批注 ID。TUI 运行时通过 IPC 通信，未运行时直接操作 .markcli.json 文件（需指定 file 参数）。",
		{
			id: z.string().describe("批注 ID"),
			file: z
				.string()
				.optional()
				.describe("文件路径（可选，TUI 未运行时必须指定）"),
		},
		async ({ id, file }) => {
			const connected = await ipc.isConnected();
			if (connected) {
				try {
					const res = await ipc.send({
						type: "remove_annotation",
						id,
					});
					if (res.type === "ok") {
						return { content: [{ type: "text", text: res.message }] };
					}
					return {
						content: [
							{
								type: "text",
								text:
									res.type === "error" ? res.message : "Unexpected response",
							},
						],
						isError: true,
					};
				} catch {
					// IPC 失败，尝试降级
				}
			}

			// 降级：直接操作文件
			if (file) {
				try {
					const data = loadAnnotations(file);
					const before = data.annotations.length;
					data.annotations = data.annotations.filter((a) => a.id !== id);
					if (data.annotations.length === before) {
						return {
							content: [{ type: "text", text: `Annotation ${id} not found` }],
							isError: true,
						};
					}
					saveAnnotations(file, data);
					return {
						content: [{ type: "text", text: `Annotation ${id} removed` }],
					};
				} catch (err) {
					return {
						content: [
							{
								type: "text",
								text: `Failed to remove annotation: ${err instanceof Error ? err.message : String(err)}`,
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
						text: "Mark is not running. Please specify a file path or start Mark first.",
					},
				],
				isError: true,
			};
		},
	);

	// resolve_annotation: 标记批注为已处理
	server.tool(
		"resolve_annotation",
		"将 Mark 中的一条批注标记为已处理（resolved）。批注不会被删除，而是以不同颜色显示。需要提供批注 ID。TUI 运行时通过 IPC 通信，未运行时直接操作 .markcli.json 文件（需指定 file 参数）。",
		{
			id: z.string().describe("批注 ID"),
			file: z
				.string()
				.optional()
				.describe("文件路径（可选，TUI 未运行时必须指定）"),
		},
		async ({ id, file }) => {
			const connected = await ipc.isConnected();
			if (connected) {
				try {
					const res = await ipc.send({
						type: "resolve_annotation",
						id,
					});
					if (res.type === "ok") {
						return { content: [{ type: "text", text: res.message }] };
					}
					return {
						content: [
							{
								type: "text",
								text:
									res.type === "error" ? res.message : "Unexpected response",
							},
						],
						isError: true,
					};
				} catch {
					// IPC 失败，尝试降级
				}
			}

			// 降级：直接操作文件
			if (file) {
				try {
					const data = loadAnnotations(file);
					const ann = data.annotations.find((a) => a.id === id);
					if (!ann) {
						return {
							content: [{ type: "text", text: `Annotation ${id} not found` }],
							isError: true,
						};
					}
					ann.resolved = true;
					saveAnnotations(file, data);
					return {
						content: [{ type: "text", text: `Annotation ${id} resolved` }],
					};
				} catch (err) {
					return {
						content: [
							{
								type: "text",
								text: `Failed to resolve annotation: ${err instanceof Error ? err.message : String(err)}`,
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
						text: "Mark is not running. Please specify a file path or start Mark first.",
					},
				],
				isError: true,
			};
		},
	);

	// clear_annotations: 清空全部批注
	server.tool(
		"clear_annotations",
		"清空指定文件的全部批注。一次性删除所有批注及 .markcli.json 文件。",
		{
			file: z
				.string()
				.optional()
				.describe("文件路径（可选，默认为当前打开的文件）"),
		},
		async ({ file }) => {
			const connected = await ipc.isConnected();
			if (connected) {
				try {
					const res = await ipc.send({
						type: "clear_annotations",
						file,
					});
					if (res.type === "ok") {
						return { content: [{ type: "text", text: res.message }] };
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

			// 降级：直接删除文件
			if (file) {
				try {
					clearAnnotations(file);
					return {
						content: [{ type: "text", text: "All annotations cleared" }],
					};
				} catch {
					return {
						content: [
							{
								type: "text",
								text: "Failed to clear annotations",
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
						text: "Mark is not running. Please specify a file path or start Mark first.",
					},
				],
				isError: true,
			};
		},
	);

	// jump_to_annotation: 跳转到指定批注
	server.tool(
		"jump_to_annotation",
		"让 Mark TUI 滚动到指定批注的位置，用户可以实时看到当前正在处理哪条批注。Mark 必须已经在运行。",
		{
			id: z.string().describe("批注 ID"),
		},
		async ({ id }) => {
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
					type: "jump_to_annotation",
					id,
				});
				if (res.type === "ok") {
					return { content: [{ type: "text", text: res.message }] };
				}
				return {
					content: [
						{
							type: "text",
							text: res.type === "error" ? res.message : "Unexpected response",
						},
					],
					isError: true,
				};
			} catch {
				return {
					content: [{ type: "text", text: "Failed to connect to Mark" }],
					isError: true,
				};
			}
		},
	);

	return server;
}
