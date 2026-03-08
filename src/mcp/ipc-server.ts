// TUI 进程侧：监听 Unix socket，暴露状态给 MCP server

import fs from "node:fs";
import net from "node:net";
import type { Annotation } from "../types.js";
import type { IpcRequest, IpcResponse } from "./protocol.js";
import { getSocketPath } from "./socket-path.js";

export interface TuiState {
	filePath: string;
	mode: string;
	annotations: Annotation[];
}

type OpenFileHandler = (filePath: string) => void;

export class IpcServer {
	private server: net.Server | null = null;
	private state: TuiState;
	private onOpenFile: OpenFileHandler | null = null;

	constructor(initialState: TuiState) {
		this.state = initialState;
	}

	updateState(state: Partial<TuiState>): void {
		Object.assign(this.state, state);
	}

	setOpenFileHandler(handler: OpenFileHandler): void {
		this.onOpenFile = handler;
	}

	start(): void {
		const socketPath = getSocketPath();

		// 清理残留 socket 文件
		try {
			fs.unlinkSync(socketPath);
		} catch {
			// 忽略
		}

		this.server = net.createServer((conn) => {
			let buffer = "";
			conn.on("data", (chunk) => {
				buffer += chunk.toString();
				// NDJSON: 按换行分割
				const lines = buffer.split("\n");
				buffer = lines.pop() ?? "";
				for (const line of lines) {
					if (!line.trim()) continue;
					try {
						const req = JSON.parse(line) as IpcRequest;
						const res = this.handleRequest(req);
						conn.write(`${JSON.stringify(res)}\n`);
					} catch {
						const err: IpcResponse = {
							type: "error",
							message: "Invalid JSON",
						};
						conn.write(`${JSON.stringify(err)}\n`);
					}
				}
			});
		});

		this.server.listen(socketPath);
	}

	private handleRequest(req: IpcRequest): IpcResponse {
		switch (req.type) {
			case "get_annotations": {
				return {
					type: "annotations",
					data: {
						file: this.state.filePath,
						annotations: this.state.annotations.map((a) => ({
							selectedText: a.selectedText,
							comment: a.comment,
							startLine: a.startLine,
							endLine: a.endLine,
						})),
					},
				};
			}
			case "get_status": {
				return {
					type: "status",
					data: {
						file: this.state.filePath,
						mode: this.state.mode,
						annotationCount: this.state.annotations.length,
					},
				};
			}
			case "open_file": {
				if (this.onOpenFile) {
					this.onOpenFile(req.path);
					return { type: "ok", message: `Opened ${req.path}` };
				}
				return { type: "error", message: "open_file handler not set" };
			}
		}
	}

	stop(): void {
		if (this.server) {
			this.server.close();
			this.server = null;
		}
		try {
			fs.unlinkSync(getSocketPath());
		} catch {
			// 忽略
		}
	}
}
