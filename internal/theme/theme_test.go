package theme

import (
	"testing"
)

func TestGet_KnownTheme(t *testing.T) {
	th := Get("tokyonight-night")
	if th.Name != "tokyonight-night" {
		t.Errorf("expected tokyonight-night, got %q", th.Name)
	}
	if th.Selection.Bg != "#283457" {
		t.Errorf("wrong selection bg: %q", th.Selection.Bg)
	}
	if th.Annotation.Fg != "#1a1b26" {
		t.Errorf("wrong annotation fg: %q", th.Annotation.Fg)
	}
	if th.Panel.Border != "#7aa2f7" {
		t.Errorf("wrong panel border: %q", th.Panel.Border)
	}
	if th.StatusBar.ModeReading != "#7aa2f7" {
		t.Errorf("wrong status bar mode reading: %q", th.StatusBar.ModeReading)
	}
}

func TestGet_UnknownFallsToDefault(t *testing.T) {
	th := Get("unknown-theme")
	def := Default()
	if th.Name != def.Name {
		t.Errorf("expected fallback to %q, got %q", def.Name, th.Name)
	}
}

func TestGet_AllThemes(t *testing.T) {
	expected := map[string]string{
		"tokyonight-storm": "#2e3c64",
		"tokyonight-moon":  "#2d3f76",
		"tokyonight-day":   "#b6bfe2",
	}
	for name, selBg := range expected {
		th := Get(name)
		if th.Name != name {
			t.Errorf("%s: wrong name %q", name, th.Name)
		}
		if th.Selection.Bg != selBg {
			t.Errorf("%s: wrong selection bg %q, want %q", name, th.Selection.Bg, selBg)
		}
	}
}

func TestNames_ReturnsFour(t *testing.T) {
	names := Names()
	if len(names) != 4 {
		t.Errorf("expected 4 names, got %d", len(names))
	}
	seen := map[string]bool{}
	for _, n := range names {
		seen[n] = true
	}
	for _, want := range []string{"tokyonight-night", "tokyonight-storm", "tokyonight-moon", "tokyonight-day"} {
		if !seen[want] {
			t.Errorf("missing theme name %q", want)
		}
	}
}

func TestDefault_IsNight(t *testing.T) {
	def := Default()
	if def.Name != "tokyonight-night" {
		t.Errorf("expected tokyonight-night, got %q", def.Name)
	}
}

func TestNames_ConsistentWithThemesMap(t *testing.T) {
	names := Names()
	// every name in Names() must exist in themes map
	for _, n := range names {
		if _, ok := themes[n]; !ok {
			t.Errorf("Names() contains %q but it's missing from themes map", n)
		}
	}
	// every key in themes map must be in Names()
	nameSet := map[string]bool{}
	for _, n := range names {
		nameSet[n] = true
	}
	for k := range themes {
		if !nameSet[k] {
			t.Errorf("themes map has %q but it's missing from Names()", k)
		}
	}
}
