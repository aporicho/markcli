import { describe, expect, it } from "vitest";
import { getSelectedText } from "./LineSelector.js";

describe("getSelectedText", () => {
	const lines = ["first line", "second line", "third line", "fourth line"];

	it("single line without col returns whole line", () => {
		expect(getSelectedText(lines, 1, 1)).toBe("first line");
	});

	it("single line with col returns slice", () => {
		expect(getSelectedText(lines, 1, 1, 2, 7)).toBe("rst l");
	});

	it("multi-line without col joins lines", () => {
		expect(getSelectedText(lines, 1, 3)).toBe(
			"first line\nsecond line\nthird line",
		);
	});

	it("multi-line with col slices first and last", () => {
		expect(getSelectedText(lines, 1, 3, 6, 5)).toBe("line\nsecond line\nthird");
	});

	it("out-of-bounds line returns empty string", () => {
		expect(getSelectedText(lines, 10, 10)).toBe("");
	});

	it("out-of-bounds in multi-line uses empty for missing lines", () => {
		// lines 3-6, lines 5 and 6 don't exist
		expect(getSelectedText(lines, 3, 6)).toBe("third line\nfourth line\n\n");
	});
});
