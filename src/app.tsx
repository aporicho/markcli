import fs from "node:fs";
import path from "node:path";
import { Box, useApp, useStdout } from "ink";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { AnnotationInput } from "./components/AnnotationInput.js";
import { getSelectedText } from "./components/LineSelector.js";
import {
	MarkdownViewer,
	type ViewerStatus,
} from "./components/MarkdownViewer.js";
import { OverviewPanel } from "./components/OverviewPanel.js";
import { StatusBar } from "./components/StatusBar.js";
import { useAnnotations } from "./hooks/useAnnotations.js";
import { useFileWatcher } from "./hooks/useFileWatcher.js";
import { useIpcServer } from "./hooks/useIpcServer.js";
import type { Annotation, AppMode } from "./types.js";
import { renderMarkdownWrapped } from "./utils/markdown.js";
import { cleanupMouseOnExit, disableMouseTracking } from "./utils/mouse.js";
import { stripAnsi } from "./utils/ranges.js";
import {
	extractAnchor,
	lineColToOffset,
	offsetToLineCol,
	relocateAnchor,
} from "./utils/textAnchor.js";

interface AppProps {
	filePath: string;
	content: string;
}

export function App({
	filePath: initialFilePath,
	content: initialContent,
}: AppProps) {
	const [currentFilePath, setCurrentFilePath] = useState(initialFilePath);
	const [currentContent, setCurrentContent] = useState(initialContent);
	const { exit } = useApp();
	const { stdout } = useStdout();
	const viewportHeight = (stdout?.rows ?? 24) - 1; // -1 for StatusBar

	// open_file handler: MCP 通过 IPC 要求打开新文件
	const handleOpenFile = useCallback((newPath: string) => {
		const resolved = path.resolve(newPath);
		if (!fs.existsSync(resolved)) return;
		const newContent = fs.readFileSync(resolved, "utf-8");
		setCurrentFilePath(resolved);
		setCurrentContent(newContent);
	}, []);

	// 文件监听：Claude 编辑文件后自动刷新
	useFileWatcher(currentFilePath, setCurrentContent);

	const termWidth = stdout?.columns ?? 80;
	const renderedLines = useMemo(
		() => renderMarkdownWrapped(currentContent, termWidth),
		[currentContent, termWidth],
	);

	// stripped 行 + 行长度数组（用于偏移量转换）
	const strippedLines = useMemo(
		() => renderedLines.map((l) => stripAnsi(l)),
		[renderedLines],
	);
	const lineLengths = useMemo(
		() => strippedLines.map((l) => l.length),
		[strippedLines],
	);
	const fullStrippedText = useMemo(
		() => strippedLines.join("\n"),
		[strippedLines],
	);

	const [mode, setMode] = useState<AppMode>("reading");
	const [viewerStatus, setViewerStatus] = useState<ViewerStatus>({
		scrollOffset: 0,
		selecting: false,
	});
	const [pendingSelection, setPendingSelection] = useState<{
		startLine: number;
		endLine: number;
		startCol: number;
		endCol: number;
	} | null>(null);

	const { annotations, addAnnotation, updateAnnotation, removeAnnotation } =
		useAnnotations(currentFilePath);

	// IPC server：暴露状态给 MCP
	const ipcState = useMemo(
		() => ({
			filePath: currentFilePath,
			mode,
			annotations,
			strippedLines,
		}),
		[currentFilePath, mode, annotations, strippedLines],
	);
	useIpcServer(ipcState, {
		onOpenFile: handleOpenFile,
		onAddAnnotation: addAnnotation,
		onUpdateAnnotation: updateAnnotation,
		onRemoveAnnotation: removeAnnotation,
	});

	const [overviewScrollTarget, setOverviewScrollTarget] = useState<{
		offset: number;
		rev: number;
	} | null>(null);
	const overviewRevRef = useRef(0);
	const [editingAnnotation, setEditingAnnotation] = useState<Annotation | null>(
		null,
	);
	const [reselectingComment, setReselectingComment] = useState<string | null>(
		null,
	);

	// ---- 将存储的批注解析到当前渲染位置 ----
	const resolvedAnnotations = useMemo(() => {
		return annotations.map((ann): Annotation => {
			// 有文本锚定 → 重新定位
			if (ann.quote) {
				const range = relocateAnchor(fullStrippedText, {
					quote: ann.quote,
					prefix: ann.prefix ?? "",
					suffix: ann.suffix ?? "",
				});
				if (range) {
					const start = offsetToLineCol(lineLengths, range.start);
					const end = offsetToLineCol(lineLengths, range.end);
					return {
						...ann,
						startLine: start.line,
						endLine: end.line,
						startCol: start.col,
						endCol: end.col,
					};
				}
			}
			// 无锚定或定位失败 → 使用存储的 line/col（旧批注兼容）
			return ann;
		});
	}, [annotations, fullStrippedText, lineLengths]);

	// 进程退出兜底清理鼠标
	useEffect(() => {
		return cleanupMouseOnExit();
	}, []);

	// MarkdownViewer 选中完成 → 进入批注模式
	const handleSelect = useCallback(
		(startLine: number, endLine: number, startCol: number, endCol: number) => {
			// 不再 disableMouseTracking()，由 AnnotationInput 接管鼠标以支持点击外部取消
			setPendingSelection({ startLine, endLine, startCol, endCol });
			setMode("annotating");
		},
		[],
	);

	// Ctrl+R 调整选区
	const handleReselect = useCallback((comment: string) => {
		setReselectingComment(comment);
		setPendingSelection(null);
		setMode("reading");
	}, []);

	// Ctrl+D 删除编辑中的批注
	const handleDeleteEditing = useCallback(() => {
		if (editingAnnotation) {
			removeAnnotation(editingAnnotation.id);
		}
		setEditingAnnotation(null);
		setReselectingComment(null);
		setPendingSelection(null);
		setMode("reading");
	}, [editingAnnotation, removeAnnotation]);

	// 双击批注 → 编辑
	const handleEditAnnotation = useCallback((ann: Annotation) => {
		// 不再 disableMouseTracking()，由 AnnotationInput 接管鼠标
		setEditingAnnotation(ann);
		setPendingSelection({
			startLine: ann.startLine,
			endLine: ann.endLine,
			startCol: ann.startCol ?? 0,
			endCol: ann.endCol ?? 0,
		});
		setMode("annotating");
	}, []);

	// 退出
	const handleQuit = useCallback(() => {
		disableMouseTracking();
		exit();
	}, [exit]);

	// 批注提交
	const handleAnnotationSubmit = useCallback(
		(comment: string) => {
			if (!pendingSelection) return;

			if (editingAnnotation) {
				if (!comment) {
					// 空内容 → 删除批注
					removeAnnotation(editingAnnotation.id);
				} else {
					// 检查选区是否变了（经过 Ctrl+R 重选）
					const selChanged =
						reselectingComment !== null &&
						(pendingSelection.startLine !== editingAnnotation.startLine ||
							pendingSelection.endLine !== editingAnnotation.endLine ||
							pendingSelection.startCol !== (editingAnnotation.startCol ?? 0) ||
							pendingSelection.endCol !== (editingAnnotation.endCol ?? 0));

					if (selChanged) {
						// 选区变了 → 删除旧批注 + 新增
						removeAnnotation(editingAnnotation.id);
						const { startLine, endLine, startCol, endCol } = pendingSelection;
						const selectedText = getSelectedText(
							strippedLines,
							startLine,
							endLine,
							startCol,
							endCol,
						);
						const offsetStart = lineColToOffset(
							lineLengths,
							startLine,
							startCol,
						);
						const offsetEnd = lineColToOffset(lineLengths, endLine, endCol);
						const anchor = extractAnchor(
							fullStrippedText,
							offsetStart,
							offsetEnd,
						);
						addAnnotation({
							startLine,
							endLine,
							startCol,
							endCol,
							selectedText,
							comment,
							quote: anchor.quote,
							prefix: anchor.prefix,
							suffix: anchor.suffix,
						});
					} else {
						updateAnnotation(editingAnnotation.id, comment);
					}
				}
				setEditingAnnotation(null);
			} else {
				const { startLine, endLine, startCol, endCol } = pendingSelection;
				const selectedText = getSelectedText(
					strippedLines,
					startLine,
					endLine,
					startCol,
					endCol,
				);

				// 提取文本锚定
				const offsetStart = lineColToOffset(lineLengths, startLine, startCol);
				const offsetEnd = lineColToOffset(lineLengths, endLine, endCol);
				const anchor = extractAnchor(fullStrippedText, offsetStart, offsetEnd);

				addAnnotation({
					startLine,
					endLine,
					startCol,
					endCol,
					selectedText,
					comment,
					quote: anchor.quote,
					prefix: anchor.prefix,
					suffix: anchor.suffix,
				});
			}
			setPendingSelection(null);
			setReselectingComment(null);
			setMode("reading");
		},
		[
			pendingSelection,
			editingAnnotation,
			reselectingComment,
			strippedLines,
			lineLengths,
			fullStrippedText,
			addAnnotation,
			updateAnnotation,
			removeAnnotation,
		],
	);

	// 批注取消
	const handleAnnotationCancel = useCallback(() => {
		setPendingSelection(null);
		setEditingAnnotation(null);
		setReselectingComment(null);
		setMode("reading");
	}, []);

	// 总览面板高度为视口 40%，垂直居中
	const overviewPanelHeight = Math.min(
		Math.max(5, Math.round(viewportHeight * 0.4)),
		viewportHeight - 2,
	);
	const overviewFloatTop = Math.max(
		0,
		Math.floor((viewportHeight - overviewPanelHeight) / 2),
	);

	// 总览模式
	const handleEnterOverview = useCallback(() => {
		setMode("overview");
	}, []);

	const handleOverviewDelete = useCallback(
		(id: string) => {
			removeAnnotation(id);
			// 留在 overview 模式，批注清空时 OverviewPanel 会自动退出
		},
		[removeAnnotation],
	);

	const handleOverviewCancel = useCallback(() => {
		setOverviewScrollTarget(null);
		setMode("reading");
	}, []);

	const handleOverviewEdit = useCallback(
		(ann: Annotation) => {
			setOverviewScrollTarget(null);
			handleEditAnnotation(ann);
		},
		[handleEditAnnotation],
	);

	const handleOverviewHighlight = useCallback(
		(ann: Annotation | null) => {
			if (ann) {
				overviewRevRef.current += 1;
				const offset = Math.max(
					0,
					ann.endLine - 1 - Math.max(0, overviewFloatTop - 1),
				);
				setOverviewScrollTarget({ offset, rev: overviewRevRef.current });
			}
		},
		[overviewFloatTop],
	);

	const isViewing = mode === "reading" || mode === "selecting";

	// 批注模式时，将 pendingSelection 作为临时批注合并，使选中文字立即高亮
	const displayAnnotations = useMemo(() => {
		if (mode !== "annotating" || !pendingSelection) return resolvedAnnotations;
		const temp: Annotation = {
			id: "__pending__",
			startLine: pendingSelection.startLine,
			endLine: pendingSelection.endLine,
			startCol: pendingSelection.startCol,
			endCol: pendingSelection.endCol,
			selectedText: "",
			comment: "",
			createdAt: "",
		};
		return [...resolvedAnnotations, temp];
	}, [mode, pendingSelection, resolvedAnnotations]);

	// StatusBar 显示的模式：优先用 viewer 的 selecting 状态
	const displayMode: AppMode =
		mode === "annotating"
			? "annotating"
			: mode === "overview"
				? "overview"
				: viewerStatus.selecting
					? "selecting"
					: "reading";

	// StatusBar 显示的选区行号
	const selStart =
		mode === "annotating"
			? (pendingSelection?.startLine ?? null)
			: (viewerStatus.selStartLine ?? null);
	const selEnd =
		mode === "annotating"
			? (pendingSelection?.endLine ?? null)
			: (viewerStatus.selEndLine ?? null);

	// 浮动窗口公共参数
	const floatLeft = 2;
	const floatWidth = termWidth - floatLeft * 2 + 1;

	// 浮动批注窗口位置计算
	const floatTop = useMemo(() => {
		if (!pendingSelection) return 0;
		const selEndInViewport =
			pendingSelection.endLine - viewerStatus.scrollOffset - 1;
		const selStartInViewport =
			pendingSelection.startLine - viewerStatus.scrollOffset - 1;
		// 弹窗高度约 4 行（边框 + 预览 + 输入）
		if (selEndInViewport + 6 < viewportHeight) {
			return selEndInViewport + 1;
		}
		return Math.max(0, selStartInViewport - 5);
	}, [pendingSelection, viewerStatus.scrollOffset, viewportHeight]);

	return (
		<Box flexDirection="column" height={stdout?.rows ?? 24}>
			<Box height={viewportHeight}>
				<MarkdownViewer
					lines={renderedLines}
					viewportHeight={viewportHeight}
					annotations={displayAnnotations}
					active={isViewing}
					onSelect={handleSelect}
					onQuit={handleQuit}
					onStatusChange={setViewerStatus}
					onOverviewMode={handleEnterOverview}
					onEditAnnotation={handleEditAnnotation}
					scrollToOffset={overviewScrollTarget}
					extraScrollPadding={
						mode === "overview" ? viewportHeight - overviewFloatTop : 0
					}
				/>
				{mode === "annotating" && pendingSelection && (
					<AnnotationInput
						onSubmit={handleAnnotationSubmit}
						onCancel={handleAnnotationCancel}
						onReselect={editingAnnotation ? handleReselect : undefined}
						onDelete={editingAnnotation ? handleDeleteEditing : undefined}
						isEditing={!!editingAnnotation}
						top={floatTop}
						left={floatLeft}
						width={floatWidth}
						initialValue={reselectingComment ?? editingAnnotation?.comment}
					/>
				)}
				{mode === "overview" && (
					<OverviewPanel
						annotations={resolvedAnnotations}
						onEdit={handleOverviewEdit}
						onDelete={handleOverviewDelete}
						onCancel={handleOverviewCancel}
						onHighlightChange={handleOverviewHighlight}
						top={overviewFloatTop}
						left={floatLeft}
						width={floatWidth}
						maxHeight={overviewPanelHeight}
					/>
				)}
			</Box>
			<StatusBar
				mode={displayMode}
				currentLine={viewerStatus.scrollOffset + 1}
				totalLines={renderedLines.length}
				selectionStart={selStart}
				selectionEnd={selEnd}
				annotationCount={resolvedAnnotations.length}
				isEditing={mode === "annotating" && !!editingAnnotation}
			/>
		</Box>
	);
}
