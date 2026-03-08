import { describe, expect, it } from "vitest";
import type { Annotation, SelectionPos } from "../types.js";
import {
	buildSegments,
	displayWidth,
	getAnnotationRangesForLine,
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

// ---- getAnnotationRangesForLine ----

function makeAnn(
	startLine: number,
	endLine: number,
	startCol?: number,
	endCol?: number,
): Annotation {
	return {
		id: "test",
		startLine,
		endLine,
		startCol,
		endCol,
		selectedText: "",
		comment: "",
		createdAt: "",
	};
}

describe("getAnnotationRangesForLine", () => {
	it("returns empty for no annotations", () => {
		expect(getAnnotationRangesForLine([], 1, 10)).toEqual([]);
	});

	it("returns empty when line is outside annotation", () => {
		expect(getAnnotationRangesForLine([makeAnn(2, 3)], 1, 10)).toEqual([]);
		expect(getAnnotationRangesForLine([makeAnn(2, 3)], 4, 10)).toEqual([]);
	});

	it("single-line annotation without col covers full line", () => {
		expect(getAnnotationRangesForLine([makeAnn(2, 2)], 2, 10)).toEqual([
			[0, 10],
		]);
	});

	it("single-line annotation with col uses col range", () => {
		expect(getAnnotationRangesForLine([makeAnn(2, 2, 3, 7)], 2, 10)).toEqual([
			[3, 7],
		]);
	});

	it("multi-line annotation — start line", () => {
		expect(getAnnotationRangesForLine([makeAnn(2, 5, 4, 8)], 2, 10)).toEqual([
			[4, 10],
		]);
	});

	it("multi-line annotation — middle line covers full", () => {
		expect(getAnnotationRangesForLine([makeAnn(2, 5, 4, 8)], 3, 10)).toEqual([
			[0, 10],
		]);
	});

	it("multi-line annotation — end line", () => {
		expect(getAnnotationRangesForLine([makeAnn(2, 5, 4, 8)], 5, 10)).toEqual([
			[0, 8],
		]);
	});

	it("merges overlapping ranges", () => {
		const anns = [makeAnn(1, 1, 0, 6), makeAnn(1, 1, 4, 10)];
		expect(getAnnotationRangesForLine(anns, 1, 10)).toEqual([[0, 10]]);
	});

	it("keeps non-overlapping ranges separate", () => {
		const anns = [makeAnn(1, 1, 0, 3), makeAnn(1, 1, 5, 8)];
		expect(getAnnotationRangesForLine(anns, 1, 10)).toEqual([
			[0, 3],
			[5, 8],
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
			{ text: " ", selected: false, annotated: false },
		]);
	});

	it("returns full text when no highlights", () => {
		expect(buildSegments("hello", null, [])).toEqual([
			{ text: "hello", selected: false, annotated: false },
		]);
	});

	it("splits on selection range", () => {
		const result = buildSegments("abcde", [1, 3], []);
		expect(result).toEqual([
			{ text: "a", selected: false, annotated: false },
			{ text: "bc", selected: true, annotated: false },
			{ text: "de", selected: false, annotated: false },
		]);
	});

	it("splits on annotation range", () => {
		const result = buildSegments("abcde", null, [[2, 4]]);
		expect(result).toEqual([
			{ text: "ab", selected: false, annotated: false },
			{ text: "cd", selected: false, annotated: true },
			{ text: "e", selected: false, annotated: false },
		]);
	});

	it("handles overlapping selection and annotation", () => {
		// text: "abcdef", sel: [1,4], ann: [3,6]
		const result = buildSegments("abcdef", [1, 4], [[3, 6]]);
		expect(result).toEqual([
			{ text: "a", selected: false, annotated: false },
			{ text: "bc", selected: true, annotated: false },
			{ text: "d", selected: true, annotated: true },
			{ text: "ef", selected: false, annotated: true },
		]);
	});

	it("handles multiple annotation ranges", () => {
		const result = buildSegments("abcdefgh", null, [
			[1, 3],
			[5, 7],
		]);
		expect(result).toEqual([
			{ text: "a", selected: false, annotated: false },
			{ text: "bc", selected: false, annotated: true },
			{ text: "de", selected: false, annotated: false },
			{ text: "fg", selected: false, annotated: true },
			{ text: "h", selected: false, annotated: false },
		]);
	});
});
