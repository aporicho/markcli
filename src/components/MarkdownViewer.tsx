import { Box, Text } from "ink";
import { useCallback, useEffect, useRef, useState } from "react";
import { useKeyboard } from "../hooks/useKeyboard.js";
import { useMouse } from "../hooks/useMouse.js";
import { useSelection } from "../hooks/useSelection.js";
import type { Annotation } from "../types.js";
import {
	buildSegments,
	getAnnotationRangesForLine,
	getSelectionRangeForLine,
	stripAnsi,
} from "../utils/ranges.js";
import { getSelectedText } from "./LineSelector.js";

export interface ViewerStatus {
	scrollOffset: number;
	selecting: boolean;
	selStartLine?: number;
	selEndLine?: number;
	selectedText?: string;
}

interface MarkdownViewerProps {
	lines: string[];
	viewportHeight: number;
	annotations: Annotation[];
	active: boolean;
	onSelect: (
		startLine: number,
		endLine: number,
		startCol: number,
		endCol: number,
	) => void;
	onQuit: () => void;
	onStatusChange?: (status: ViewerStatus) => void;
	onOverviewMode?: () => void;
	onEditAnnotation?: (annotation: Annotation) => void;
	scrollToOffset?: { offset: number; rev: number } | null;
	extraScrollPadding?: number;
}

export function MarkdownViewer({
	lines,
	viewportHeight,
	annotations,
	active,
	onSelect,
	onQuit,
	onStatusChange,
	onOverviewMode,
	onEditAnnotation,
	scrollToOffset,
	extraScrollPadding = 0,
}: MarkdownViewerProps) {
	const [scrollOffset, setScrollOffset] = useState(0);
	const maxOffset = Math.max(
		0,
		lines.length - viewportHeight + extraScrollPadding,
	);

	// resize 时修正 scrollOffset
	const prevLinesLenRef = useRef(lines.length);
	useEffect(() => {
		if (prevLinesLenRef.current !== lines.length) {
			prevLinesLenRef.current = lines.length;
			setScrollOffset((prev) => Math.min(prev, maxOffset));
		}
	}, [lines.length, maxOffset]);

	// ---- 外部滚动控制 ----
	useEffect(() => {
		if (scrollToOffset == null) return;
		setScrollOffset(Math.max(0, Math.min(scrollToOffset.offset, maxOffset)));
	}, [scrollToOffset, maxOffset]);

	// ---- 滚动 ----
	const scrollUp = useCallback(
		(n = 1) => setScrollOffset((p) => Math.max(0, p - n)),
		[],
	);
	const scrollDown = useCallback(
		(n = 1) => setScrollOffset((p) => Math.min(maxOffset, p + n)),
		[maxOffset],
	);

	// ---- 选择 ----
	const selection = useSelection({
		lines,
		viewportHeight,
		scrollOffset,
		onConfirm: onSelect,
	});

	// ---- 鼠标 ----
	// 单击只记录起点，拖拽时才真正进入选中模式
	const pendingClickRef = useRef<{ line: number; col: number } | null>(null);

	const handleMouseEvent = useCallback(
		(event: import("../hooks/useMouse.js").MouseEvent) => {
			if (event.type === "scroll") {
				if (event.direction === "up") scrollUp(3);
				else scrollDown(3);
			} else if (event.type === "doubleclick") {
				pendingClickRef.current = null;
				const hit = annotations.find((a) => {
					if (event.lineNum < a.startLine || event.lineNum > a.endLine)
						return false;
					if (
						event.lineNum === a.startLine &&
						event.textCol < (a.startCol ?? 0)
					)
						return false;
					if (
						event.lineNum === a.endLine &&
						event.textCol > (a.endCol ?? Infinity)
					)
						return false;
					return true;
				});
				if (hit && onEditAnnotation) {
					onEditAnnotation(hit);
				}
			} else if (event.type === "click") {
				// 单击只暂存起点，不进入选中模式
				pendingClickRef.current = { line: event.lineNum, col: event.textCol };
				selection.cancel();
			} else if (event.type === "drag") {
				// 首次拖拽：用暂存的起点开始选区
				if (pendingClickRef.current) {
					selection.startSelection(pendingClickRef.current);
					pendingClickRef.current = null;
				}
				selection.updateEnd({ line: event.lineNum, col: event.textCol });
			} else if (event.type === "release") {
				pendingClickRef.current = null;
			}
		},
		[scrollUp, scrollDown, selection, annotations, onEditAnnotation],
	);

	useMouse({ active, lines, scrollOffset, onEvent: handleMouseEvent });

	// ---- 键盘 ----
	useKeyboard({
		active,
		selecting: selection.selecting,
		viewportHeight,
		onScrollUp: scrollUp,
		onScrollDown: scrollDown,
		onQuit,
		onEnterSelection: selection.enterWithKeyboard,
		onCancelSelection: selection.cancel,
		onConfirmSelection: selection.confirm,
		onMoveLineBy: (delta) => {
			const newLine = selection.moveLineBy(delta);
			if (newLine !== null) {
				if (newLine <= scrollOffset) scrollUp();
				if (newLine > scrollOffset + viewportHeight) scrollDown();
			}
		},
		onMoveColBy: selection.moveColBy,
		onOverviewMode,
	});

	// ---- 状态上报 ----
	useEffect(() => {
		let selectedText: string | undefined;
		if (selection.selecting && selection.normStart && selection.normEnd) {
			selectedText = getSelectedText(
				lines.map((l) => stripAnsi(l)),
				selection.normStart.line,
				selection.normEnd.line,
				selection.normStart.col,
				selection.normEnd.col,
			);
		}
		onStatusChange?.({
			scrollOffset,
			selecting: selection.selecting,
			selStartLine: selection.normStart?.line,
			selEndLine: selection.normEnd?.line,
			selectedText,
		});
	}, [
		scrollOffset,
		lines,
		selection.selecting,
		selection.normStart,
		selection.normEnd,
		onStatusChange,
	]);

	// ---- 渲染 ----
	const sliceEnd = Math.min(lines.length, scrollOffset + viewportHeight);
	const visibleLines = lines.slice(scrollOffset, sliceEnd);
	// 当 scrollOffset 超出实际行数（extraScrollPadding 区域），用空行填充
	const padCount = scrollOffset + viewportHeight - sliceEnd;
	const { normStart, normEnd } = selection;

	return (
		<Box
			flexDirection="column"
			flexGrow={1}
			flexShrink={1}
			flexBasis={0}
			overflow="hidden"
		>
			{visibleLines.map((line, idx) => {
				const lineNum = scrollOffset + idx + 1;
				const stripped = stripAnsi(line) || " ";

				// 计算高亮区间
				const selRange =
					normStart && normEnd
						? getSelectionRangeForLine(
								lineNum,
								stripped.length,
								normStart,
								normEnd,
							)
						: null;
				const annRanges = getAnnotationRangesForLine(
					annotations,
					lineNum,
					stripped.length,
				);

				// 无高亮 → 直接渲染原始行（保留 ANSI 颜色）
				if (!selRange && annRanges.length === 0) {
					return (
						<Box key={lineNum} flexDirection="row">
							<Text>{line || " "}</Text>
						</Box>
					);
				}

				// 有高亮 → 分段渲染
				const segments = buildSegments(stripped, selRange, annRanges);

				return (
					<Box key={lineNum} flexDirection="row">
						{segments.map((seg, i) => {
							if (seg.selected) {
								return (
									<Text key={i} backgroundColor="blue" color="white">
										{seg.text}
									</Text>
								);
							}
							if (seg.annotated) {
								return (
									<Text key={i} backgroundColor="yellow" color="black">
										{seg.text}
									</Text>
								);
							}
							return <Text key={i}>{seg.text}</Text>;
						})}
					</Box>
				);
			})}
			{padCount > 0 &&
				Array.from({ length: padCount }, (_, i) => (
					<Box key={`pad-${i}`} flexDirection="row">
						<Text> </Text>
					</Box>
				))}
		</Box>
	);
}
