import { useCallback, useEffect, useRef, useState } from "react";
import type { SelectionPos } from "../types.js";
import { normalizePos, stripAnsi } from "../utils/ranges.js";

interface UseSelectionOptions {
	lines: string[];
	viewportHeight: number;
	scrollOffset: number;
	onConfirm: (
		startLine: number,
		endLine: number,
		startCol: number,
		endCol: number,
	) => void;
}

interface SelectionState {
	selecting: boolean;
	start: SelectionPos | null;
	end: SelectionPos | null;
}

const EMPTY: SelectionState = { selecting: false, start: null, end: null };

export function useSelection({
	lines,
	viewportHeight,
	scrollOffset,
	onConfirm,
}: UseSelectionOptions) {
	const [sel, setSel] = useState<SelectionState>(EMPTY);

	// lines 变化时重置选择（终端 resize）
	const prevLinesLenRef = useRef(lines.length);
	useEffect(() => {
		if (prevLinesLenRef.current !== lines.length) {
			prevLinesLenRef.current = lines.length;
			if (sel.selecting) {
				setSel(EMPTY);
			}
		}
	}, [lines.length, sel.selecting]);

	const clampCol = useCallback(
		(lineNum: number, col: number) => {
			const line = lines[lineNum - 1];
			if (!line) return 0;
			return Math.max(0, Math.min(col, stripAnsi(line).length));
		},
		[lines],
	);

	// ---- 对外操作 ----

	const startSelection = useCallback((pos: SelectionPos) => {
		setSel({ selecting: true, start: pos, end: pos });
	}, []);

	const updateEnd = useCallback((pos: SelectionPos) => {
		setSel((prev) => ({ ...prev, end: pos }));
	}, []);

	const cancel = useCallback(() => {
		setSel(EMPTY);
	}, []);

	const confirm = useCallback(() => {
		setSel((prev) => {
			if (prev.start !== null && prev.end !== null) {
				const [s, e] = normalizePos(prev.start, prev.end);
				onConfirm(s.line, e.line, s.col, e.col);
			}
			return EMPTY;
		});
	}, [onConfirm]);

	/** v 键进入选择模式 */
	const enterWithKeyboard = useCallback(() => {
		const center = scrollOffset + Math.floor(viewportHeight / 2);
		startSelection({ line: center, col: 0 });
	}, [scrollOffset, viewportHeight, startSelection]);

	/** 上下移动选择端点，返回新行号 */
	const moveLineBy = useCallback(
		(delta: number): number | null => {
			let newLine: number | null = null;
			setSel((prev) => {
				if (!prev.end) return prev;
				newLine = Math.max(1, Math.min(lines.length, prev.end.line + delta));
				const newCol = clampCol(newLine, prev.end.col);
				return { ...prev, end: { line: newLine, col: newCol } };
			});
			return newLine;
		},
		[lines.length, clampCol],
	);

	/** 左右移动列 */
	const moveColBy = useCallback(
		(delta: number) => {
			setSel((prev) => {
				if (!prev.end) return prev;
				const maxCol = clampCol(prev.end.line, Infinity);
				const newCol = Math.max(0, Math.min(maxCol, prev.end.col + delta));
				return { ...prev, end: { line: prev.end.line, col: newCol } };
			});
		},
		[clampCol],
	);

	// 规范化后的范围（渲染用）
	const [normStart, normEnd] =
		sel.start !== null && sel.end !== null
			? normalizePos(sel.start, sel.end)
			: [null, null];

	return {
		selecting: sel.selecting,
		selEnd: sel.end,
		normStart,
		normEnd,
		startSelection,
		updateEnd,
		cancel,
		confirm,
		enterWithKeyboard,
		moveLineBy,
		moveColBy,
	};
}
