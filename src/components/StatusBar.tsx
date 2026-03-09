import { Box, Text } from "ink";
import type React from "react";
import type { Theme } from "../themes.js";
import type { AppMode } from "../types.js";

interface StatusBarProps {
	mode: AppMode;
	currentLine: number;
	totalLines: number;
	selectionStart: number | null;
	selectionEnd: number | null;
	annotationCount: number;
	isEditing?: boolean;
	theme: Theme;
}

interface Segment {
	text: string;
	fg: string;
	bg: string;
}

const ARROW = "\ue0b0";

const MODE_LABELS: Record<AppMode, string> = {
	reading: " 阅读 ",
	selecting: " 选中 ",
	overview: " 总览 ",
	annotating: " 批注 ",
};

const SHORTCUTS: Record<AppMode, string> = {
	reading: " v:选中 d:总览 ^T:主题 q:退出 ↑↓:滚动 ",
	selecting: " a:批注 Esc:取消 ↑↓:扩展 ",
	overview: " Enter:编辑 ⌫:删除 ↑↓:选择 Esc:返回 ",
	annotating: " Enter:确认 ^J:换行 Esc:取消 ",
};

const SHORTCUTS_EDITING = " Enter:提交 ^J:换行 ^R:调整选区 ^D:删除 Esc:取消 ";

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
	theme,
}: StatusBarProps) {
	const sb = theme.statusBar;
	const leftSegments: Segment[] = [
		{ text: MODE_LABELS[mode], fg: sb.modeFg, bg: sb.mode[mode] },
		{ text: ` ${annotationCount} 批注 `, fg: sb.fg, bg: sb.dimBg },
	];

	if (selectionStart !== null) {
		const range =
			selectionEnd !== null && selectionEnd !== selectionStart
				? `L${selectionStart}-${selectionEnd}`
				: `L${selectionStart}`;
		leftSegments.push({ text: ` ${range} `, fg: sb.fg, bg: sb.accentBg });
	}

	const shortcuts =
		mode === "annotating" && isEditing ? SHORTCUTS_EDITING : SHORTCUTS[mode];

	const rightSegments: Segment[] = [
		{ text: ` 行 ${currentLine}/${totalLines} `, fg: sb.fg, bg: sb.dimBg },
		{ text: shortcuts, fg: sb.fg, bg: sb.bg },
	];

	return (
		<Box flexDirection="row" width="100%">
			<Box>{renderSegments(leftSegments)}</Box>
			<Box flexGrow={1} />
			<Box>{renderSegments(rightSegments)}</Box>
		</Box>
	);
}
