// MCP 进程侧：连接 TUI 的 Unix socket

import net from "node:net";
import type { IpcRequest, IpcResponse } from "./protocol.js";
import { getSocketPath } from "./socket-path.js";

export class IpcClient {
	private socketPath: string;

	constructor() {
		this.socketPath = getSocketPath();
	}

	/**
	 * 发送请求并等待响应。超时 3 秒。
	 */
	send(request: IpcRequest): Promise<IpcResponse> {
		return new Promise((resolve, reject) => {
			const conn = net.createConnection(this.socketPath);
			let buffer = "";
			const timeout = setTimeout(() => {
				conn.destroy();
				reject(new Error("IPC timeout"));
			}, 3000);

			conn.on("connect", () => {
				conn.write(`${JSON.stringify(request)}\n`);
			});

			conn.on("data", (chunk) => {
				buffer += chunk.toString();
				const idx = buffer.indexOf("\n");
				if (idx !== -1) {
					clearTimeout(timeout);
					const line = buffer.slice(0, idx);
					conn.destroy();
					try {
						resolve(JSON.parse(line) as IpcResponse);
					} catch {
						reject(new Error("Invalid JSON response"));
					}
				}
			});

			conn.on("error", (err) => {
				clearTimeout(timeout);
				reject(err);
			});
		});
	}

	/**
	 * 检查 TUI 是否在运行（socket 是否可连接）
	 */
	isConnected(): Promise<boolean> {
		return new Promise((resolve) => {
			const conn = net.createConnection(this.socketPath);
			const timeout = setTimeout(() => {
				conn.destroy();
				resolve(false);
			}, 1000);

			conn.on("connect", () => {
				clearTimeout(timeout);
				conn.destroy();
				resolve(true);
			});

			conn.on("error", () => {
				clearTimeout(timeout);
				resolve(false);
			});
		});
	}
}
