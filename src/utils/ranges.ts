import stringWidth from "string-width";
import type { Annotation, SelectionPos } from "../types.js";

// ---- ANSI 处理 ----

export function stripAnsi(str: string): string {
	return str.replace(/\x1b\[[0-9;]*m/g, "");
}

/** 终端列号（0-based）→ 字符索引，正确处理双宽字符 */
export function termColToCharIndex(text: string, termCol: number): number {
	let colAcc = 0;
	for (let i = 0; i < text.length; i++) {
		const charW = stringWidth(text[i]!);
		if (colAcc + charW > termCol) return i;
		colAcc += charW;
	}
	return text.length;
}

/** 字符索引 → 终端列宽（用于 lineLength） */
export function displayWidth(text: string): number {
	return stringWidth(text);
}

// ---- 位置规范化 ----

/** 比较两个 SelectionPos，返回 [start, end] 使 start <= end */
export function normalizePos(
	a: SelectionPos,
	b: SelectionPos,
): [SelectionPos, SelectionPos] {
	if (a.line < b.line || (a.line === b.line && a.col <= b.col)) {
		return [a, b];
	}
	return [b, a];
}

// ---- 区间计算 ----

/** 带批注索引的区间，用于交替配色 */
export interface AnnotatedRange {
	start: number;
	end: number;
	index: number;
}

/** 计算某行上的未 resolved 批注区间（保留批注身份，不合并） */
export function getAnnotatedRangesForLine(
	annotations: Annotation[],
	lineNum: number,
	lineLength: number,
): AnnotatedRange[] {
	const ranges: AnnotatedRange[] = [];
	for (let i = 0; i < annotations.length; i++) {
		const ann = annotations[i]!;
		if (ann.resolved) continue;
		const range = annotationRangeForLine(ann, lineNum, lineLength);
		if (range) ranges.push({ start: range[0], end: range[1], index: i });
	}
	return ranges;
}

/** 计算某行上的 resolved 批注区间（保留批注身份，不合并） */
export function getResolvedRangesForLine(
	annotations: Annotation[],
	lineNum: number,
	lineLength: number,
): AnnotatedRange[] {
	const ranges: AnnotatedRange[] = [];
	for (let i = 0; i < annotations.length; i++) {
		const ann = annotations[i]!;
		if (!ann.resolved) continue;
		const range = annotationRangeForLine(ann, lineNum, lineLength);
		if (range) ranges.push({ start: range[0], end: range[1], index: i });
	}
	return ranges;
}

/** 计算单个批注在某行上的区间 */
function annotationRangeForLine(
	ann: Annotation,
	lineNum: number,
	lineLength: number,
): [number, number] | null {
	if (lineNum < ann.startLine || lineNum > ann.endLine) return null;

	let start: number;
	let end: number;
	const hasCol = ann.startCol !== undefined && ann.endCol !== undefined;

	if (ann.startLine === ann.endLine) {
		start = hasCol ? ann.startCol! : 0;
		end = hasCol ? ann.endCol! : lineLength;
	} else if (lineNum === ann.startLine) {
		start = hasCol ? ann.startCol! : 0;
		end = lineLength;
	} else if (lineNum === ann.endLine) {
		start = 0;
		end = hasCol ? ann.endCol! : lineLength;
	} else {
		start = 0;
		end = lineLength;
	}

	start = Math.max(0, start);
	end = Math.min(end, lineLength);
	return end > start ? [start, end] : null;
}

/** 计算某行在选择范围内的列区间 */
export function getSelectionRangeForLine(
	lineNum: number,
	lineLength: number,
	normStart: SelectionPos,
	normEnd: SelectionPos,
): [number, number] | null {
	if (lineNum < normStart.line || lineNum > normEnd.line) return null;

	const isSingleLine = normStart.line === normEnd.line;
	let s: number, e: number;

	if (isSingleLine) {
		s = normStart.col;
		e = normEnd.col;
	} else if (lineNum === normStart.line) {
		s = normStart.col;
		e = lineLength;
	} else if (lineNum === normEnd.line) {
		s = 0;
		e = normEnd.col;
	} else {
		s = 0;
		e = lineLength;
	}

	s = Math.max(0, Math.min(s, lineLength));
	e = Math.max(s, Math.min(e, lineLength));
	return e > s ? [s, e] : null;
}

// ---- 分段渲染 ----

export interface Segment {
	text: string;
	selected: boolean;
	annotationIndex: number | null;
	resolvedIndex: number | null;
}

/** 将一行文本按选择和批注区间切分为样式段 */
export function buildSegments(
	stripped: string,
	selRange: [number, number] | null,
	annRanges: AnnotatedRange[],
	resolvedRanges: AnnotatedRange[] = [],
): Segment[] {
	if (stripped.length === 0)
		return [
			{
				text: " ",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
		];

	const cuts = new Set<number>();
	cuts.add(0);
	cuts.add(stripped.length);

	if (selRange) {
		cuts.add(selRange[0]);
		cuts.add(selRange[1]);
	}
	for (const r of annRanges) {
		cuts.add(r.start);
		cuts.add(r.end);
	}
	for (const r of resolvedRanges) {
		cuts.add(r.start);
		cuts.add(r.end);
	}

	const sortedCuts = [...cuts].sort((a, b) => a - b);
	const segments: Segment[] = [];

	for (let i = 0; i < sortedCuts.length - 1; i++) {
		const start = sortedCuts[i]!;
		const end = sortedCuts[i + 1]!;
		if (start >= end) continue;

		const text = stripped.slice(start, end);
		const mid = start;

		const selected =
			selRange !== null && mid >= selRange[0] && mid < selRange[1];

		const annHit = annRanges.find((r) => mid >= r.start && mid < r.end);
		const resHit = resolvedRanges.find((r) => mid >= r.start && mid < r.end);

		segments.push({
			text,
			selected,
			annotationIndex: annHit ? annHit.index : null,
			resolvedIndex: resHit ? resHit.index : null,
		});
	}

	return segments.length > 0
		? segments
		: [
				{
					text: " ",
					selected: false,
					annotationIndex: null,
					resolvedIndex: null,
				},
			];
}
