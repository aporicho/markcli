import { describe, expect, it } from "vitest";
import type { Annotation, SelectionPos } from "../types.js";
import {
	buildSegments,
	displayWidth,
	getAnnotatedRangesForLine,
	getSelectionRangeForLine,
	normalizePos,
	stripAnsi,
	termColToCharIndex,
} from "./ranges.js";

// ---- stripAnsi ----

describe("stripAnsi", () => {
	it("returns empty string unchanged", () => {
		expect(stripAnsi("")).toBe("");
	});

	it("returns plain text unchanged", () => {
		expect(stripAnsi("hello world")).toBe("hello world");
	});

	it("strips a single SGR color code", () => {
		expect(stripAnsi("\x1b[31mred\x1b[0m")).toBe("red");
	});

	it("strips nested/multiple SGR codes", () => {
		expect(stripAnsi("\x1b[1m\x1b[31mbold red\x1b[0m normal")).toBe(
			"bold red normal",
		);
	});

	it("handles CJK text mixed with ANSI", () => {
		expect(stripAnsi("\x1b[32m你好\x1b[0m世界")).toBe("你好世界");
	});
});

// ---- termColToCharIndex ----

describe("termColToCharIndex", () => {
	it("maps ASCII cols 1:1", () => {
		expect(termColToCharIndex("abc", 0)).toBe(0);
		expect(termColToCharIndex("abc", 1)).toBe(1);
		expect(termColToCharIndex("abc", 2)).toBe(2);
	});

	it("handles CJK double-width characters", () => {
		// "你好" => col 0-1 = 你(index 0), col 2-3 = 好(index 1)
		expect(termColToCharIndex("你好", 0)).toBe(0);
		expect(termColToCharIndex("你好", 1)).toBe(0); // middle of 你
		expect(termColToCharIndex("你好", 2)).toBe(1);
	});

	it("returns text.length when col exceeds width", () => {
		expect(termColToCharIndex("ab", 5)).toBe(2);
	});

	it("returns 0 for empty string", () => {
		expect(termColToCharIndex("", 0)).toBe(0);
	});
});

// ---- displayWidth ----

describe("displayWidth", () => {
	it("returns 0 for empty string", () => {
		expect(displayWidth("")).toBe(0);
	});

	it("counts ASCII characters as width 1", () => {
		expect(displayWidth("abc")).toBe(3);
	});

	it("counts CJK characters as width 2", () => {
		expect(displayWidth("你好")).toBe(4);
	});

	it("handles mixed ASCII and CJK", () => {
		expect(displayWidth("a你b")).toBe(4); // 1 + 2 + 1
	});
});

// ---- normalizePos ----

describe("normalizePos", () => {
	it("returns already-sorted positions unchanged", () => {
		const a: SelectionPos = { line: 1, col: 0 };
		const b: SelectionPos = { line: 2, col: 5 };
		expect(normalizePos(a, b)).toEqual([a, b]);
	});

	it("swaps reversed positions", () => {
		const a: SelectionPos = { line: 3, col: 5 };
		const b: SelectionPos = { line: 1, col: 2 };
		expect(normalizePos(a, b)).toEqual([b, a]);
	});

	it("sorts same-line positions by col", () => {
		const a: SelectionPos = { line: 1, col: 10 };
		const b: SelectionPos = { line: 1, col: 3 };
		expect(normalizePos(a, b)).toEqual([b, a]);
	});

	it("keeps equal positions", () => {
		const a: SelectionPos = { line: 2, col: 4 };
		const b: SelectionPos = { line: 2, col: 4 };
		expect(normalizePos(a, b)).toEqual([a, b]);
	});
});

// ---- getAnnotatedRangesForLine ----

function makeAnn(
	startLine: number,
	endLine: number,
	startCol?: number,
	endCol?: number,
	opts?: { id?: string; resolved?: boolean },
): Annotation {
	return {
		id: opts?.id ?? "test",
		startLine,
		endLine,
		startCol,
		endCol,
		selectedText: "",
		comment: "",
		createdAt: "",
		resolved: opts?.resolved,
	};
}

describe("getAnnotatedRangesForLine", () => {
	it("returns empty for no annotations", () => {
		expect(getAnnotatedRangesForLine([], 1, 10)).toEqual([]);
	});

	it("returns empty when line is outside annotation", () => {
		expect(getAnnotatedRangesForLine([makeAnn(2, 3)], 1, 10)).toEqual([]);
		expect(getAnnotatedRangesForLine([makeAnn(2, 3)], 4, 10)).toEqual([]);
	});

	it("single-line annotation without col covers full line", () => {
		expect(getAnnotatedRangesForLine([makeAnn(2, 2)], 2, 10)).toEqual([
			{ start: 0, end: 10, index: 0 },
		]);
	});

	it("single-line annotation with col uses col range", () => {
		expect(getAnnotatedRangesForLine([makeAnn(2, 2, 3, 7)], 2, 10)).toEqual([
			{ start: 3, end: 7, index: 0 },
		]);
	});

	it("multi-line annotation — start line", () => {
		expect(getAnnotatedRangesForLine([makeAnn(2, 5, 4, 8)], 2, 10)).toEqual([
			{ start: 4, end: 10, index: 0 },
		]);
	});

	it("multi-line annotation — middle line covers full", () => {
		expect(getAnnotatedRangesForLine([makeAnn(2, 5, 4, 8)], 3, 10)).toEqual([
			{ start: 0, end: 10, index: 0 },
		]);
	});

	it("multi-line annotation — end line", () => {
		expect(getAnnotatedRangesForLine([makeAnn(2, 5, 4, 8)], 5, 10)).toEqual([
			{ start: 0, end: 8, index: 0 },
		]);
	});

	it("preserves overlapping ranges with separate indices", () => {
		const anns = [makeAnn(1, 1, 0, 6), makeAnn(1, 1, 4, 10)];
		expect(getAnnotatedRangesForLine(anns, 1, 10)).toEqual([
			{ start: 0, end: 6, index: 0 },
			{ start: 4, end: 10, index: 1 },
		]);
	});

	it("keeps non-overlapping ranges with separate indices", () => {
		const anns = [makeAnn(1, 1, 0, 3), makeAnn(1, 1, 5, 8)];
		expect(getAnnotatedRangesForLine(anns, 1, 10)).toEqual([
			{ start: 0, end: 3, index: 0 },
			{ start: 5, end: 8, index: 1 },
		]);
	});

	it("skips resolved annotations", () => {
		const anns = [
			makeAnn(1, 1, 0, 5, { resolved: true }),
			makeAnn(1, 1, 5, 10),
		];
		expect(getAnnotatedRangesForLine(anns, 1, 10)).toEqual([
			{ start: 5, end: 10, index: 1 },
		]);
	});
});

// ---- getSelectionRangeForLine ----

describe("getSelectionRangeForLine", () => {
	it("returns null when line is outside selection", () => {
		expect(
			getSelectionRangeForLine(1, 10, { line: 2, col: 0 }, { line: 3, col: 5 }),
		).toBeNull();
		expect(
			getSelectionRangeForLine(5, 10, { line: 2, col: 0 }, { line: 3, col: 5 }),
		).toBeNull();
	});

	it("single-line selection", () => {
		expect(
			getSelectionRangeForLine(2, 10, { line: 2, col: 3 }, { line: 2, col: 7 }),
		).toEqual([3, 7]);
	});

	it("multi-line — start line", () => {
		expect(
			getSelectionRangeForLine(2, 10, { line: 2, col: 4 }, { line: 5, col: 6 }),
		).toEqual([4, 10]);
	});

	it("multi-line — middle line", () => {
		expect(
			getSelectionRangeForLine(3, 10, { line: 2, col: 4 }, { line: 5, col: 6 }),
		).toEqual([0, 10]);
	});

	it("multi-line — end line", () => {
		expect(
			getSelectionRangeForLine(5, 10, { line: 2, col: 4 }, { line: 5, col: 6 }),
		).toEqual([0, 6]);
	});

	it("returns null for zero-width selection", () => {
		expect(
			getSelectionRangeForLine(1, 10, { line: 1, col: 5 }, { line: 1, col: 5 }),
		).toBeNull();
	});
});

// ---- buildSegments ----

describe("buildSegments", () => {
	it("returns a space segment for empty string", () => {
		expect(buildSegments("", null, [])).toEqual([
			{
				text: " ",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
		]);
	});

	it("returns full text when no highlights", () => {
		expect(buildSegments("hello", null, [])).toEqual([
			{
				text: "hello",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
		]);
	});

	it("splits on selection range", () => {
		const result = buildSegments("abcde", [1, 3], []);
		expect(result).toEqual([
			{
				text: "a",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
			{
				text: "bc",
				selected: true,
				annotationIndex: null,
				resolvedIndex: null,
			},
			{
				text: "de",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
		]);
	});

	it("splits on annotation range", () => {
		const result = buildSegments("abcde", null, [
			{ start: 2, end: 4, index: 0 },
		]);
		expect(result).toEqual([
			{
				text: "ab",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
			{
				text: "cd",
				selected: false,
				annotationIndex: 0,
				resolvedIndex: null,
			},
			{
				text: "e",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
		]);
	});

	it("handles overlapping selection and annotation", () => {
		// text: "abcdef", sel: [1,4], ann: [3,6]
		const result = buildSegments(
			"abcdef",
			[1, 4],
			[{ start: 3, end: 6, index: 0 }],
		);
		expect(result).toEqual([
			{
				text: "a",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
			{
				text: "bc",
				selected: true,
				annotationIndex: null,
				resolvedIndex: null,
			},
			{
				text: "d",
				selected: true,
				annotationIndex: 0,
				resolvedIndex: null,
			},
			{
				text: "ef",
				selected: false,
				annotationIndex: 0,
				resolvedIndex: null,
			},
		]);
	});

	it("handles multiple annotation ranges with different indices", () => {
		const result = buildSegments("abcdefgh", null, [
			{ start: 1, end: 3, index: 0 },
			{ start: 5, end: 7, index: 1 },
		]);
		expect(result).toEqual([
			{
				text: "a",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
			{
				text: "bc",
				selected: false,
				annotationIndex: 0,
				resolvedIndex: null,
			},
			{
				text: "de",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
			{
				text: "fg",
				selected: false,
				annotationIndex: 1,
				resolvedIndex: null,
			},
			{
				text: "h",
				selected: false,
				annotationIndex: null,
				resolvedIndex: null,
			},
		]);
	});

	it("alternating indices enable color differentiation", () => {
		// 3 adjacent annotations → indices 0, 1, 2 → colors alternate by % 2
		const result = buildSegments("abcdefghi", null, [
			{ start: 0, end: 3, index: 0 },
			{ start: 3, end: 6, index: 1 },
			{ start: 6, end: 9, index: 2 },
		]);
		expect(result).toEqual([
			{
				text: "abc",
				selected: false,
				annotationIndex: 0,
				resolvedIndex: null,
			},
			{
				text: "def",
				selected: false,
				annotationIndex: 1,
				resolvedIndex: null,
			},
			{
				text: "ghi",
				selected: false,
				annotationIndex: 2,
				resolvedIndex: null,
			},
		]);
		// even indices (0, 2) → one color, odd (1) → another
		expect(result[0]!.annotationIndex! % 2).toBe(0);
		expect(result[1]!.annotationIndex! % 2).toBe(1);
		expect(result[2]!.annotationIndex! % 2).toBe(0);
	});
});
