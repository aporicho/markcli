import { Box, Text, useInput } from "ink";
import { useEffect, useState } from "react";
import type { Theme } from "../themes.js";
import type { Annotation } from "../types.js";
import { disableMouseTracking } from "../utils/mouse.js";

interface OverviewPanelProps {
	annotations: Annotation[];
	onEdit: (annotation: Annotation) => void;
	onDelete: (id: string) => void;
	onCancel: () => void;
	onHighlightChange: (annotation: Annotation | null) => void;
	top?: number;
	left?: number;
	width?: number;
	maxHeight: number;
	theme?: Theme;
}

export function OverviewPanel({
	annotations,
	onEdit,
	onDelete,
	onCancel,
	onHighlightChange,
	top,
	left,
	width,
	maxHeight,
	theme,
}: OverviewPanelProps) {
	const [cursorIndex, setCursorIndex] = useState(0);

	useEffect(() => {
		disableMouseTracking();
	}, []);

	// 光标越界修正
	useEffect(() => {
		if (annotations.length === 0) {
			onHighlightChange(null);
			onCancel();
			return;
		}
		const idx = Math.min(cursorIndex, annotations.length - 1);
		if (idx !== cursorIndex) setCursorIndex(idx);
		onHighlightChange(annotations[idx]!);
	}, [annotations, cursorIndex, onHighlightChange, onCancel]);

	// 初始高亮（仅挂载时执行一次）
	// biome-ignore lint/correctness/useExhaustiveDependencies: intentionally run once on mount
	useEffect(() => {
		if (annotations.length > 0) {
			onHighlightChange(annotations[0]!);
		}
	}, []);

	useInput((input, key) => {
		if (key.escape) {
			onCancel();
			return;
		}
		if (key.upArrow) {
			setCursorIndex((i) => {
				const next = Math.max(0, i - 1);
				onHighlightChange(annotations[next]!);
				return next;
			});
			return;
		}
		if (key.downArrow) {
			setCursorIndex((i) => {
				const next = Math.min(annotations.length - 1, i + 1);
				onHighlightChange(annotations[next]!);
				return next;
			});
			return;
		}
		if (key.return) {
			if (annotations.length > 0) {
				onEdit(annotations[cursorIndex]!);
			}
			return;
		}
		if (key.backspace || key.delete || input === "d") {
			if (annotations.length > 0) {
				onDelete(annotations[cursorIndex]!.id);
			}
			return;
		}
	});

	// 固定面板高度，内部可见行数 = 总高度 - 边框 2 行
	const panelHeight = maxHeight;
	const innerMaxHeight = maxHeight - 2;

	// 列表内部滚动
	let scrollStart = 0;
	if (annotations.length > innerMaxHeight) {
		scrollStart = Math.max(
			0,
			Math.min(
				cursorIndex - Math.floor(innerMaxHeight / 2),
				annotations.length - innerMaxHeight,
			),
		);
	}
	const visibleAnnotations = annotations.slice(
		scrollStart,
		scrollStart + innerMaxHeight,
	);

	return (
		<Box
			flexDirection="column"
			borderStyle="round"
			borderColor={theme?.panel.border ?? "blue"}
			backgroundColor={theme?.panel.bg ?? "black"}
			paddingX={1}
			position="absolute"
			marginTop={top}
			marginLeft={left}
			width={width}
			height={panelHeight}
		>
			{annotations.length === 0 ? (
				<Text dimColor>暂无批注</Text>
			) : (
				visibleAnnotations.map((ann, i) => {
					const realIndex = scrollStart + i;
					const isCurrent = realIndex === cursorIndex;
					const range =
						ann.startLine === ann.endLine
							? `L${ann.startLine}`
							: `L${ann.startLine}-${ann.endLine}`;
					const textPreview =
						ann.selectedText.length > 25
							? `${ann.selectedText.slice(0, 22)}...`
							: ann.selectedText;
					const commentPreview =
						ann.comment.length > 25
							? `${ann.comment.slice(0, 22)}...`
							: ann.comment;
					const isResolved = ann.resolved === true;
					const prefix = isCurrent ? "▸" : isResolved ? "✓" : " ";
					return (
						<Text
							key={ann.id}
							inverse={isCurrent}
							dimColor={isResolved && !isCurrent}
						>
							{prefix}{" "}
							<Text
								color={
									isCurrent
										? undefined
										: isResolved
											? undefined
											: (theme?.panel.accent ?? "cyan")
								}
							>
								{range.padEnd(9)}
							</Text>
							<Text dimColor={!isCurrent}>
								"{textPreview.replace(/\n/g, "↵")}"
							</Text>
							<Text> → {commentPreview}</Text>
						</Text>
					);
				})
			)}
		</Box>
	);
}
