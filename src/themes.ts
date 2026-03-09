import type { AppMode } from "./types.js";

export interface Theme {
	name: string;
	selection: { fg: string; bg: string };
	annotation: { fg: string; bg: string };
	annotationAlt: { fg: string; bg: string };
	annotationResolved: { fg: string; bg: string };
	panel: { bg: string; border: string; accent: string };
	statusBar: {
		mode: Record<AppMode, string>;
		modeFg: string;
		fg: string;
		bg: string;
		dimBg: string;
		accentBg: string;
	};
}

const night: Theme = {
	name: "tokyonight-night",
	selection: { fg: "#c0caf5", bg: "#283457" },
	annotation: { fg: "#1a1b26", bg: "#e0af68" },
	annotationAlt: { fg: "#1a1b26", bg: "#7dcfff" },
	annotationResolved: { fg: "#1a1b26", bg: "#9ece6a" },
	panel: { bg: "#1a1b26", border: "#7aa2f7", accent: "#7dcfff" },
	statusBar: {
		mode: {
			reading: "#7aa2f7",
			selecting: "#ff9e64",
			overview: "#73daca",
			annotating: "#bb9af7",
		},
		modeFg: "#1a1b26",
		fg: "#c0caf5",
		bg: "#1a1b26",
		dimBg: "#414868",
		accentBg: "#7aa2f7",
	},
};

const storm: Theme = {
	name: "tokyonight-storm",
	selection: { fg: "#c0caf5", bg: "#2e3c64" },
	annotation: { fg: "#24283b", bg: "#e0af68" },
	annotationAlt: { fg: "#24283b", bg: "#7dcfff" },
	annotationResolved: { fg: "#24283b", bg: "#9ece6a" },
	panel: { bg: "#24283b", border: "#7aa2f7", accent: "#7dcfff" },
	statusBar: {
		mode: {
			reading: "#7aa2f7",
			selecting: "#ff9e64",
			overview: "#73daca",
			annotating: "#bb9af7",
		},
		modeFg: "#24283b",
		fg: "#c0caf5",
		bg: "#24283b",
		dimBg: "#414868",
		accentBg: "#7aa2f7",
	},
};

const moon: Theme = {
	name: "tokyonight-moon",
	selection: { fg: "#c8d3f5", bg: "#2d3f76" },
	annotation: { fg: "#222436", bg: "#ffc777" },
	annotationAlt: { fg: "#222436", bg: "#86e1fc" },
	annotationResolved: { fg: "#222436", bg: "#c3e88d" },
	panel: { bg: "#222436", border: "#82aaff", accent: "#86e1fc" },
	statusBar: {
		mode: {
			reading: "#82aaff",
			selecting: "#ff966c",
			overview: "#4fd6be",
			annotating: "#fca7ea",
		},
		modeFg: "#222436",
		fg: "#c8d3f5",
		bg: "#222436",
		dimBg: "#636da6",
		accentBg: "#82aaff",
	},
};

const day: Theme = {
	name: "tokyonight-day",
	selection: { fg: "#3760bf", bg: "#b6bfe2" },
	annotation: { fg: "#e1e2e7", bg: "#8c6c3e" },
	annotationAlt: { fg: "#e1e2e7", bg: "#007197" },
	annotationResolved: { fg: "#e1e2e7", bg: "#587539" },
	panel: { bg: "#e1e2e7", border: "#2e7de9", accent: "#007197" },
	statusBar: {
		mode: {
			reading: "#2e7de9",
			selecting: "#b15c00",
			overview: "#118c74",
			annotating: "#7847bd",
		},
		modeFg: "#e1e2e7",
		fg: "#3760bf",
		bg: "#d5d6db",
		dimBg: "#848cb5",
		accentBg: "#2e7de9",
	},
};

export const THEMES: Record<string, Theme> = {
	"tokyonight-night": night,
	"tokyonight-storm": storm,
	"tokyonight-moon": moon,
	"tokyonight-day": day,
};

export const THEME_NAMES = Object.keys(THEMES);

export const DEFAULT_THEME = "tokyonight-night";

export function getTheme(name: string): Theme {
	return THEMES[name] ?? THEMES[DEFAULT_THEME]!;
}

export function isValidTheme(name: string): boolean {
	return name in THEMES;
}
