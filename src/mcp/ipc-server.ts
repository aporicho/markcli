// TUI 进程侧：监听 Unix socket，暴露状态给 MCP server

import fs from "node:fs";
import net from "node:net";
import type { Annotation } from "../types.js";
import {
	extractAnchor,
	offsetToLineCol,
	relocateAnchor,
} from "../utils/textAnchor.js";
import type { IpcRequest, IpcResponse } from "./protocol.js";
import { getSocketPath } from "./socket-path.js";

export interface TuiState {
	filePath: string;
	mode: string;
	annotations: Annotation[];
	strippedLines: string[];
	selectedText?: string;
}

type OpenFileHandler = (filePath: string) => void;
type AddAnnotationHandler = (params: {
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
type UpdateAnnotationHandler = (id: string, comment: string) => void;
type RemoveAnnotationHandler = (id: string) => void;
type ResolveAnnotationHandler = (id: string) => void;
type RefreshHandler = () => void;
type ClearAnnotationsHandler = () => void;
type JumpToAnnotationHandler = (id: string) => void;

export class IpcServer {
	private server: net.Server | null = null;
	private state: TuiState;
	private onOpenFile: OpenFileHandler | null = null;
	private onRefresh: RefreshHandler | null = null;
	private onAddAnnotation: AddAnnotationHandler | null = null;
	private onUpdateAnnotation: UpdateAnnotationHandler | null = null;
	private onRemoveAnnotation: RemoveAnnotationHandler | null = null;
	private onResolveAnnotation: ResolveAnnotationHandler | null = null;
	private onClearAnnotations: ClearAnnotationsHandler | null = null;
	private onJumpToAnnotation: JumpToAnnotationHandler | null = null;

	constructor(initialState: TuiState) {
		this.state = initialState;
	}

	updateState(state: Partial<TuiState>): void {
		Object.assign(this.state, state);
	}

	setOpenFileHandler(handler: OpenFileHandler): void {
		this.onOpenFile = handler;
	}

	setRefreshHandler(handler: RefreshHandler): void {
		this.onRefresh = handler;
	}

	setAddAnnotationHandler(handler: AddAnnotationHandler): void {
		this.onAddAnnotation = handler;
	}

	setUpdateAnnotationHandler(handler: UpdateAnnotationHandler): void {
		this.onUpdateAnnotation = handler;
	}

	setRemoveAnnotationHandler(handler: RemoveAnnotationHandler): void {
		this.onRemoveAnnotation = handler;
	}

	setResolveAnnotationHandler(handler: ResolveAnnotationHandler): void {
		this.onResolveAnnotation = handler;
	}

	setClearAnnotationsHandler(handler: ClearAnnotationsHandler): void {
		this.onClearAnnotations = handler;
	}

	setJumpToAnnotationHandler(handler: JumpToAnnotationHandler): void {
		this.onJumpToAnnotation = handler;
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
			case "list_annotations": {
				return {
					type: "annotations",
					data: {
						file: this.state.filePath,
						annotations: this.state.annotations.map((a) => ({
							id: a.id,
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
			case "get_selection": {
				return {
					type: "selection",
					data: {
						file: this.state.filePath,
						selectedText: this.state.selectedText ?? null,
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
			case "refresh_file": {
				if (this.onRefresh) {
					this.onRefresh();
					return { type: "ok", message: "Refreshed" };
				}
				return { type: "error", message: "refresh_file handler not set" };
			}
			case "add_annotation": {
				if (!this.onAddAnnotation) {
					return { type: "error", message: "add_annotation handler not set" };
				}

				const { strippedLines } = this.state;
				const lineLengths = strippedLines.map((l) => l.length);
				const fullStrippedText = strippedLines.join("\n");

				// 精确匹配
				let matchStart = fullStrippedText.indexOf(req.selectedText);

				if (matchStart === -1) {
					// 模糊匹配兜底
					const range = relocateAnchor(fullStrippedText, {
						quote: req.selectedText,
						prefix: "",
						suffix: "",
					});
					if (!range) {
						return {
							type: "error",
							message: `Text not found: "${req.selectedText.slice(0, 50)}..."`,
						};
					}
					matchStart = range.start;
				}

				const matchEnd = matchStart + req.selectedText.length;
				const start = offsetToLineCol(lineLengths, matchStart);
				const end = offsetToLineCol(lineLengths, matchEnd);
				const anchor = extractAnchor(fullStrippedText, matchStart, matchEnd);

				this.onAddAnnotation({
					selectedText: req.selectedText,
					comment: req.comment,
					startLine: start.line,
					endLine: end.line,
					startCol: start.col,
					endCol: end.col,
					quote: anchor.quote,
					prefix: anchor.prefix,
					suffix: anchor.suffix,
				});
				return { type: "ok", message: "Annotation added" };
			}
			case "update_annotation": {
				if (this.onUpdateAnnotation) {
					this.onUpdateAnnotation(req.id, req.comment);
					return { type: "ok", message: `Annotation ${req.id} updated` };
				}
				return { type: "error", message: "update_annotation handler not set" };
			}
			case "remove_annotation": {
				if (this.onRemoveAnnotation) {
					this.onRemoveAnnotation(req.id);
					return { type: "ok", message: `Annotation ${req.id} removed` };
				}
				return { type: "error", message: "remove_annotation handler not set" };
			}
			case "resolve_annotation": {
				if (this.onResolveAnnotation) {
					this.onResolveAnnotation(req.id);
					return { type: "ok", message: `Annotation ${req.id} resolved` };
				}
				return { type: "error", message: "resolve_annotation handler not set" };
			}
			case "clear_annotations": {
				if (this.onClearAnnotations) {
					this.onClearAnnotations();
					return { type: "ok", message: "All annotations cleared" };
				}
				return { type: "error", message: "clear_annotations handler not set" };
			}
			case "jump_to_annotation": {
				const ann = this.state.annotations.find((a) => a.id === req.id);
				if (!ann) {
					return { type: "error", message: `Annotation ${req.id} not found` };
				}
				if (this.onJumpToAnnotation) {
					this.onJumpToAnnotation(req.id);
					return { type: "ok", message: `Jumped to annotation ${req.id}` };
				}
				return { type: "error", message: "jump_to_annotation handler not set" };
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
