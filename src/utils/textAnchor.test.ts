import { describe, expect, it } from "vitest";
import {
	extractAnchor,
	lineColToOffset,
	offsetToLineCol,
	relocateAnchor,
} from "./textAnchor.js";

// ---- lineColToOffset ----

describe("lineColToOffset", () => {
	// lineLengths represents lengths of each line (without \n)
	// "abc\ndef\nghi" => lineLengths = [3, 3, 3]

	it("first line first col → 0", () => {
		expect(lineColToOffset([3, 3, 3], 1, 0)).toBe(0);
	});

	it("first line with col offset", () => {
		expect(lineColToOffset([3, 3, 3], 1, 2)).toBe(2);
	});

	it("second line accounts for \\n", () => {
		// line 2, col 0 → after "abc\n" = offset 4
		expect(lineColToOffset([3, 3, 3], 2, 0)).toBe(4);
	});

	it("third line", () => {
		// line 3, col 1 → after "abc\ndef\n" + 1 = 8 + 1 = 9
		expect(lineColToOffset([3, 3, 3], 3, 1)).toBe(9);
	});

	it("single line", () => {
		expect(lineColToOffset([5], 1, 3)).toBe(3);
	});
});

// ---- offsetToLineCol ----

describe("offsetToLineCol", () => {
	it("offset 0 → line 1, col 0", () => {
		expect(offsetToLineCol([3, 3, 3], 0)).toEqual({ line: 1, col: 0 });
	});

	it("offset at second line start", () => {
		expect(offsetToLineCol([3, 3, 3], 4)).toEqual({ line: 2, col: 0 });
	});

	it("offset in middle of last line", () => {
		expect(offsetToLineCol([3, 3, 3], 9)).toEqual({ line: 3, col: 1 });
	});

	it("clamps col to line length on last line", () => {
		// offset way beyond end — last line length 3, so col clamps to 3
		expect(offsetToLineCol([3, 3, 3], 100)).toEqual({ line: 3, col: 3 });
	});

	it("roundtrips with lineColToOffset", () => {
		const lengths = [5, 10, 3];
		for (let line = 1; line <= 3; line++) {
			for (let col = 0; col <= lengths[line - 1]!; col++) {
				const offset = lineColToOffset(lengths, line, col);
				expect(offsetToLineCol(lengths, offset)).toEqual({ line, col });
			}
		}
	});
});

// ---- extractAnchor ----

describe("extractAnchor", () => {
	const text = "The quick brown fox jumps over the lazy dog.";

	it("extracts quote, prefix, and suffix", () => {
		const anchor = extractAnchor(text, 10, 19); // "brown fox"
		expect(anchor.quote).toBe("brown fox");
		expect(anchor.prefix).toBe("The quick ");
		expect(anchor.suffix).toBe(" jumps over the lazy dog.");
	});

	it("truncates prefix when near start", () => {
		const anchor = extractAnchor(text, 0, 3); // "The"
		expect(anchor.quote).toBe("The");
		expect(anchor.prefix).toBe("");
	});

	it("truncates suffix when near end", () => {
		const anchor = extractAnchor(text, 40, 44); // "dog."
		expect(anchor.quote).toBe("dog.");
		expect(anchor.suffix).toBe("");
	});
});

// ---- relocateAnchor ----

describe("relocateAnchor", () => {
	it("returns null for empty quote", () => {
		expect(
			relocateAnchor("some text", { quote: "", prefix: "", suffix: "" }),
		).toBeNull();
	});

	it("unique exact match", () => {
		const text = "The quick brown fox jumps.";
		const result = relocateAnchor(text, {
			quote: "brown fox",
			prefix: "quick ",
			suffix: " jumps",
		});
		expect(result).toEqual({ start: 10, end: 19 });
	});

	it("multiple exact matches — disambiguates with prefix/suffix", () => {
		const text = "abc foo abc foo abc";
		// "foo" appears at index 4 and 12
		const result = relocateAnchor(text, {
			quote: "foo",
			prefix: "abc foo abc ", // suffix of this matches position 12 better
			suffix: " abc",
		});
		expect(result).toEqual({ start: 12, end: 15 });
	});

	it("fuzzy match with small edits", () => {
		// Use a longer quote so 20% threshold allows enough errors
		const modified = "The quick brownn foxx jumps over the lazy dog.";
		const anchor = {
			quote: "brown fox jumps over",
			prefix: "quick ",
			suffix: " the",
		};
		const result = relocateAnchor(modified, anchor);
		expect(result).not.toBeNull();
		expect(result!.start).toBeGreaterThanOrEqual(8);
	});

	it("returns null when text is completely different", () => {
		const result = relocateAnchor("completely unrelated text here", {
			quote: "xyzxyzxyzxyzxyzxyz",
			prefix: "aaa",
			suffix: "bbb",
		});
		expect(result).toBeNull();
	});

	it("handles text shifted by insertion", () => {
		const text = "INSERTED The quick brown fox.";
		const result = relocateAnchor(text, {
			quote: "brown fox",
			prefix: "quick ",
			suffix: ".",
		});
		expect(result).toEqual({ start: 19, end: 28 });
	});
});
