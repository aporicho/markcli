import { Box, Text } from "ink";
import type React from "react";
import type { AppMode } from "../types.js";

interface StatusBarProps {
	mode: AppMode;
	currentLine: number;
	totalLines: number;
	selectionStart: number | null;
	selectionEnd: number | null;
	annotationCount: number;
	isEditing?: boolean;
}

interface Segment {
	text: string;
	fg: string;
	bg: string;
}

const ARROW = "\ue0b0";

const MODE_COLORS: Record<AppMode, string> = {
	reading: "green",
	selecting: "yellow",
	overview: "blue",
	annotating: "cyan",
};

const MODE_LABELS: Record<AppMode, string> = {
	reading: " 阅读 ",
	selecting: " 选中 ",
	overview: " 总览 ",
	annotating: " 批注 ",
};

const SHORTCUTS: Record<AppMode, string> = {
	reading: " v:选中 d:总览 q:退出 ↑↓:滚动 ",
	selecting: " a:批注 Esc:取消 ↑↓:扩展 ",
	overview: " Enter:编辑 ⌫:删除 ↑↓:选择 Esc:返回 ",
	annotating: " Enter:确认 Esc:取消 ",
};

const SHORTCUTS_EDITING = " Enter:提交 ^R:调整选区 ^D:删除 Esc:取消 ";

function renderSegments(segments: Segment[], trailingBg?: string) {
	const nodes: React.ReactNode[] = [];
	for (let i = 0; i < segments.length; i++) {
		const seg = segments[i]!;
		// Segment content
		nodes.push(
			<Text key={`s${i}`} color={seg.fg} backgroundColor={seg.bg}>
				{seg.text}
			</Text>,
		);
		// Arrow separator
		const nextBg = i < segments.length - 1 ? segments[i + 1]!.bg : trailingBg;
		nodes.push(
			<Text key={`a${i}`} color={seg.bg} backgroundColor={nextBg}>
				{ARROW}
			</Text>,
		);
	}
	return nodes;
}

export function StatusBar({
	mode,
	currentLine,
	totalLines,
	selectionStart,
	selectionEnd,
	annotationCount,
	isEditing,
}: StatusBarProps) {
	const leftSegments: Segment[] = [
		{ text: MODE_LABELS[mode], fg: "black", bg: MODE_COLORS[mode] },
		{ text: ` ${annotationCount} 批注 `, fg: "white", bg: "gray" },
	];

	if (selectionStart !== null) {
		const range =
			selectionEnd !== null && selectionEnd !== selectionStart
				? `L${selectionStart}-${selectionEnd}`
				: `L${selectionStart}`;
		leftSegments.push({ text: ` ${range} `, fg: "white", bg: "blue" });
	}

	const shortcuts =
		mode === "annotating" && isEditing ? SHORTCUTS_EDITING : SHORTCUTS[mode];

	const rightSegments: Segment[] = [
		{ text: ` 行 ${currentLine}/${totalLines} `, fg: "white", bg: "gray" },
		{ text: shortcuts, fg: "white", bg: "black" },
	];

	return (
		<Box flexDirection="row" width="100%">
			<Box>{renderSegments(leftSegments)}</Box>
			<Box flexGrow={1} />
			<Box>{renderSegments(rightSegments)}</Box>
		</Box>
	);
}
