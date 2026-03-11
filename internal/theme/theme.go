package theme

// Color holds foreground and background hex color values.
type Color struct{ Fg, Bg string }

// PanelStyle holds colors for floating panels.
type PanelStyle struct{ Bg, Border, Accent string }

// StatusBarStyle holds colors for the status bar.
type StatusBarStyle struct {
	ModeBrowsing   string
	ModeReading    string
	ModeSelecting  string
	ModeOverview   string
	ModeAnnotating string
	ModeFg         string
	Fg             string
	Bg             string
	DimBg          string
	AccentBg       string
}

// Theme is the full color theme for the TUI.
type Theme struct {
	Name               string
	Selection          Color
	Annotation         Color
	AnnotationAlt      Color
	AnnotationResolved Color
	Panel              PanelStyle
	StatusBar          StatusBarStyle
}

var themes = map[string]Theme{
	"tokyonight-night": {
		Name:               "tokyonight-night",
		Selection:          Color{Fg: "#c0caf5", Bg: "#283457"},
		Annotation:         Color{Fg: "#1a1b26", Bg: "#e0af68"},
		AnnotationAlt:      Color{Fg: "#1a1b26", Bg: "#7dcfff"},
		AnnotationResolved: Color{Fg: "#1a1b26", Bg: "#9ece6a"},
		Panel:              PanelStyle{Bg: "#1a1b26", Border: "#7aa2f7", Accent: "#7dcfff"},
		StatusBar: StatusBarStyle{
			ModeBrowsing:   "#73daca",
			ModeReading:    "#7aa2f7",
			ModeSelecting:  "#ff9e64",
			ModeOverview:   "#73daca",
			ModeAnnotating: "#bb9af7",
			ModeFg:         "#1a1b26",
			Fg:             "#c0caf5",
			Bg:             "#1a1b26",
			DimBg:          "#414868",
			AccentBg:       "#7aa2f7",
		},
	},
	"tokyonight-storm": {
		Name:               "tokyonight-storm",
		Selection:          Color{Fg: "#c0caf5", Bg: "#2e3c64"},
		Annotation:         Color{Fg: "#24283b", Bg: "#e0af68"},
		AnnotationAlt:      Color{Fg: "#24283b", Bg: "#7dcfff"},
		AnnotationResolved: Color{Fg: "#24283b", Bg: "#9ece6a"},
		Panel:              PanelStyle{Bg: "#24283b", Border: "#7aa2f7", Accent: "#7dcfff"},
		StatusBar: StatusBarStyle{
			ModeBrowsing:   "#73daca",
			ModeReading:    "#7aa2f7",
			ModeSelecting:  "#ff9e64",
			ModeOverview:   "#73daca",
			ModeAnnotating: "#bb9af7",
			ModeFg:         "#24283b",
			Fg:             "#c0caf5",
			Bg:             "#24283b",
			DimBg:          "#414868",
			AccentBg:       "#7aa2f7",
		},
	},
	"tokyonight-moon": {
		Name:               "tokyonight-moon",
		Selection:          Color{Fg: "#c8d3f5", Bg: "#2d3f76"},
		Annotation:         Color{Fg: "#222436", Bg: "#ffc777"},
		AnnotationAlt:      Color{Fg: "#222436", Bg: "#86e1fc"},
		AnnotationResolved: Color{Fg: "#222436", Bg: "#c3e88d"},
		Panel:              PanelStyle{Bg: "#222436", Border: "#82aaff", Accent: "#86e1fc"},
		StatusBar: StatusBarStyle{
			ModeBrowsing:   "#4fd6be",
			ModeReading:    "#82aaff",
			ModeSelecting:  "#ff966c",
			ModeOverview:   "#4fd6be",
			ModeAnnotating: "#fca7ea",
			ModeFg:         "#222436",
			Fg:             "#c8d3f5",
			Bg:             "#222436",
			DimBg:          "#636da6",
			AccentBg:       "#82aaff",
		},
	},
	"tokyonight-day": {
		Name:               "tokyonight-day",
		Selection:          Color{Fg: "#3760bf", Bg: "#b6bfe2"},
		Annotation:         Color{Fg: "#e1e2e7", Bg: "#8c6c3e"},
		AnnotationAlt:      Color{Fg: "#e1e2e7", Bg: "#007197"},
		AnnotationResolved: Color{Fg: "#e1e2e7", Bg: "#587539"},
		Panel:              PanelStyle{Bg: "#e1e2e7", Border: "#2e7de9", Accent: "#007197"},
		StatusBar: StatusBarStyle{
			ModeBrowsing:   "#118c74",
			ModeReading:    "#2e7de9",
			ModeSelecting:  "#b15c00",
			ModeOverview:   "#118c74",
			ModeAnnotating: "#7847bd",
			ModeFg:         "#e1e2e7",
			Fg:             "#3760bf",
			Bg:             "#d5d6db",
			DimBg:          "#848cb5",
			AccentBg:       "#2e7de9",
		},
	},
}

// Names returns all available theme names.
func Names() []string {
	return []string{
		"tokyonight-night",
		"tokyonight-storm",
		"tokyonight-moon",
		"tokyonight-day",
	}
}

// Default returns the default theme (tokyonight-night).
func Default() Theme {
	return themes["tokyonight-night"]
}

// Get returns the named theme, falling back to Default if not found.
func Get(name string) Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return Default()
}
