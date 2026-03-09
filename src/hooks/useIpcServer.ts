import { useEffect, useRef } from "react";
import { IpcServer, type TuiState } from "../mcp/ipc-server.js";

interface IpcHandlers {
	onOpenFile: (filePath: string) => void;
	onRefresh: () => void;
	onAddAnnotation: (params: {
		selectedText: string;
		comment: string;
		startLine: number;
		endLine: number;
		startCol: number;
		endCol: number;
		quote: string;
		prefix: string;
		suffix: string;
	}) => void;
	onUpdateAnnotation: (id: string, comment: string) => void;
	onRemoveAnnotation: (id: string) => void;
	onResolveAnnotation: (id: string) => void;
	onClearAnnotations: () => void;
	onJumpToAnnotation: (id: string) => void;
}

/**
 * React hook：TUI 启动时创建 IPC socket server，暴露状态给 MCP。
 * 通过 useRef 保持对最新 state 的引用。
 */
export function useIpcServer(state: TuiState, handlers: IpcHandlers): void {
	const serverRef = useRef<IpcServer | null>(null);

	// 初始化 server（仅挂载时启动一次，状态和 handler 通过后续 effect 同步）
	// biome-ignore lint/correctness/useExhaustiveDependencies: intentionally run once on mount
	useEffect(() => {
		const server = new IpcServer(state);
		server.setOpenFileHandler(handlers.onOpenFile);
		server.setRefreshHandler(handlers.onRefresh);
		server.setAddAnnotationHandler(handlers.onAddAnnotation);
		server.setUpdateAnnotationHandler(handlers.onUpdateAnnotation);
		server.setRemoveAnnotationHandler(handlers.onRemoveAnnotation);
		server.setResolveAnnotationHandler(handlers.onResolveAnnotation);
		server.setClearAnnotationsHandler(handlers.onClearAnnotations);
		server.setJumpToAnnotationHandler(handlers.onJumpToAnnotation);
		server.start();
		serverRef.current = server;
		return () => {
			server.stop();
			serverRef.current = null;
		};
	}, []);

	// 同步最新状态到 server
	useEffect(() => {
		serverRef.current?.updateState(state);
	}, [state]);

	// 同步 handlers
	useEffect(() => {
		serverRef.current?.setOpenFileHandler(handlers.onOpenFile);
		serverRef.current?.setRefreshHandler(handlers.onRefresh);
		serverRef.current?.setAddAnnotationHandler(handlers.onAddAnnotation);
		serverRef.current?.setUpdateAnnotationHandler(handlers.onUpdateAnnotation);
		serverRef.current?.setRemoveAnnotationHandler(handlers.onRemoveAnnotation);
		serverRef.current?.setResolveAnnotationHandler(
			handlers.onResolveAnnotation,
		);
		serverRef.current?.setClearAnnotationsHandler(handlers.onClearAnnotations);
		serverRef.current?.setJumpToAnnotationHandler(handlers.onJumpToAnnotation);
	}, [
		handlers.onOpenFile,
		handlers.onRefresh,
		handlers.onAddAnnotation,
		handlers.onUpdateAnnotation,
		handlers.onRemoveAnnotation,
		handlers.onResolveAnnotation,
		handlers.onClearAnnotations,
		handlers.onJumpToAnnotation,
	]);
}
