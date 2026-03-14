package theme

import (
	"testing"
)

func TestDetect_ReturnsDarkOrLight(t *testing.T) {
	th := Detect()
	// Can't assert Dark value since it depends on terminal, just ensure no panic
	_ = th.Dark
}

func TestTheme_DarkMode_ContrastFg(t *testing.T) {
	th := Theme{Dark: true}
	if th.contrastFg() != "0" {
		t.Errorf("dark mode contrastFg should be \"0\", got %q", th.contrastFg())
	}
}

func TestTheme_LightMode_ContrastFg(t *testing.T) {
	th := Theme{Dark: false}
	if th.contrastFg() != "15" {
		t.Errorf("light mode contrastFg should be \"15\", got %q", th.contrastFg())
	}
}

func TestTheme_DarkMode_Methods(t *testing.T) {
	th := Theme{Dark: true}

	tests := []struct {
		name string
		got  string
		want string
	}{
		{"SelectionBg", th.SelectionBg(), "4"},
		{"SelectionFg", th.SelectionFg(), "0"},
		{"AnnotationBg", th.AnnotationBg(), "3"},
		{"AnnotationFg", th.AnnotationFg(), "0"},
		{"AnnotationAltBg", th.AnnotationAltBg(), "6"},
		{"AnnotationAltFg", th.AnnotationAltFg(), "0"},
		{"AnnotationResolvedBg", th.AnnotationResolvedBg(), "2"},
		{"AnnotationResolvedFg", th.AnnotationResolvedFg(), "0"},
		{"PanelBorder", th.PanelBorder(), "4"},
		{"PanelAccent", th.PanelAccent(), "6"},
		{"PanelBg", th.PanelBg(), "0"},
		{"ModeBrowsingBg", th.ModeBrowsingBg(), "2"},
		{"ModeReadingBg", th.ModeReadingBg(), "4"},
		{"ModeSelectingBg", th.ModeSelectingBg(), "3"},
		{"ModeAnnotatingBg", th.ModeAnnotatingBg(), "5"},
		{"ModeOverviewBg", th.ModeOverviewBg(), "6"},
		{"ModeFg", th.ModeFg(), "0"},
		{"StatusFg", th.StatusFg(), "7"},
		{"StatusBg", th.StatusBg(), "0"},
		{"StatusHintFg", th.StatusHintFg(), "8"},
		{"ErrorBg", th.ErrorBg(), "1"},
		{"ErrorFg", th.ErrorFg(), "15"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}

func TestTheme_LightMode_Methods(t *testing.T) {
	th := Theme{Dark: false}

	// Light mode differs in contrastFg and PanelBg/StatusBg
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"SelectionFg", th.SelectionFg(), "15"},
		{"AnnotationFg", th.AnnotationFg(), "15"},
		{"PanelBg", th.PanelBg(), "15"},
		{"ModeFg", th.ModeFg(), "15"},
		{"StatusBg", th.StatusBg(), "15"},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %q, want %q", tt.name, tt.got, tt.want)
		}
	}
}
