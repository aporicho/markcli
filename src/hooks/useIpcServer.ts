import { useEffect, useRef } from "react";
import { IpcServer, type TuiState } from "../mcp/ipc-server.js";

/**
 * React hook：TUI 启动时创建 IPC socket server，暴露状态给 MCP。
 * 通过 useRef 保持对最新 state 的引用。
 */
export function useIpcServer(
	state: TuiState,
	onOpenFile: (filePath: string) => void,
): void {
	const serverRef = useRef<IpcServer | null>(null);

	// 初始化 server（仅一次）
	useEffect(() => {
		const server = new IpcServer(state);
		server.setOpenFileHandler(onOpenFile);
		server.start();
		serverRef.current = server;
		return () => {
			server.stop();
			serverRef.current = null;
		};
		// 仅在挂载时启动
		// eslint-disable-next-line react-hooks/exhaustive-deps
	}, []);

	// 同步最新状态到 server
	useEffect(() => {
		serverRef.current?.updateState(state);
	}, [state]);

	// 同步 onOpenFile handler
	useEffect(() => {
		serverRef.current?.setOpenFileHandler(onOpenFile);
	}, [onOpenFile]);
}
