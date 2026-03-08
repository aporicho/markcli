import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AnnotationFile } from "../types.js";
import {
	clearAnnotations,
	loadAnnotations,
	saveAnnotations,
} from "./storage.js";

vi.mock("node:fs", () => ({
	default: {
		existsSync: vi.fn(),
		readFileSync: vi.fn(),
		writeFileSync: vi.fn(),
		unlinkSync: vi.fn(),
	},
}));

// Import the mocked fs after mock declaration
import fs from "node:fs";

const mockedFs = vi.mocked(fs);

beforeEach(() => {
	vi.clearAllMocks();
});

describe("loadAnnotations", () => {
	it("returns parsed JSON when file exists", () => {
		const data: AnnotationFile = { file: "test.md", annotations: [] };
		mockedFs.existsSync.mockReturnValue(true);
		mockedFs.readFileSync.mockReturnValue(JSON.stringify(data));

		const result = loadAnnotations("/a/b/test.md");
		expect(result).toEqual(data);
		expect(mockedFs.existsSync).toHaveBeenCalledWith(
			"/a/b/test.md.markcli.json",
		);
	});

	it("returns default empty object when file does not exist", () => {
		mockedFs.existsSync.mockReturnValue(false);

		const result = loadAnnotations("/a/b/test.md");
		expect(result).toEqual({ file: "test.md", annotations: [] });
	});
});

describe("saveAnnotations", () => {
	it("writes JSON to correct path", () => {
		const data: AnnotationFile = { file: "test.md", annotations: [] };
		saveAnnotations("/a/b/test.md", data);

		expect(mockedFs.writeFileSync).toHaveBeenCalledWith(
			"/a/b/test.md.markcli.json",
			JSON.stringify(data, null, 2),
			"utf-8",
		);
	});
});

describe("clearAnnotations", () => {
	it("deletes file when it exists", () => {
		mockedFs.existsSync.mockReturnValue(true);
		clearAnnotations("/a/b/test.md");
		expect(mockedFs.unlinkSync).toHaveBeenCalledWith(
			"/a/b/test.md.markcli.json",
		);
	});

	it("does nothing when file does not exist", () => {
		mockedFs.existsSync.mockReturnValue(false);
		clearAnnotations("/a/b/test.md");
		expect(mockedFs.unlinkSync).not.toHaveBeenCalled();
	});
});
