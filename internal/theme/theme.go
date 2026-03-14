package theme

import "github.com/charmbracelet/lipgloss"

// Theme uses ANSI 0-15 color numbers so the terminal's palette drives all colors.
// Dark is detected via lipgloss.HasDarkBackground().
type Theme struct {
	Dark bool
}

// Detect returns a Theme with Dark set based on the terminal background.
func Detect() Theme {
	return Theme{Dark: lipgloss.HasDarkBackground()}
}

// contrastFg returns a foreground suitable for colored backgrounds:
// "0" (black) on dark terminals, "15" (bright white) on light.
func (t Theme) contrastFg() string {
	if t.Dark {
		return "0"
	}
	return "15"
}

// --- Selection ---

func (t Theme) SelectionBg() string { return "4" }
func (t Theme) SelectionFg() string { return t.contrastFg() }

// --- Annotation ---

func (t Theme) AnnotationBg() string  { return "3" }
func (t Theme) AnnotationFg() string  { return t.contrastFg() }
func (t Theme) AnnotationAltBg() string { return "6" }
func (t Theme) AnnotationAltFg() string { return t.contrastFg() }
func (t Theme) AnnotationResolvedBg() string { return "2" }
func (t Theme) AnnotationResolvedFg() string { return t.contrastFg() }

// --- Panel ---

func (t Theme) PanelBorder() string { return "4" }
func (t Theme) PanelAccent() string { return "6" }

func (t Theme) PanelBg() string {
	if t.Dark {
		return "0"
	}
	return "15"
}

// --- StatusBar mode labels ---

func (t Theme) ModeBrowsingBg() string   { return "2" }
func (t Theme) ModeReadingBg() string    { return "4" }
func (t Theme) ModeSelectingBg() string  { return "3" }
func (t Theme) ModeAnnotatingBg() string { return "5" }
func (t Theme) ModeOverviewBg() string   { return "6" }
func (t Theme) ModeFg() string           { return t.contrastFg() }

// --- StatusBar info area ---

func (t Theme) StatusFg() string { return "7" }

func (t Theme) StatusBg() string {
	if t.Dark {
		return "0"
	}
	return "15"
}

func (t Theme) StatusHintFg() string { return "8" }

// --- Error ---

func (t Theme) ErrorBg() string { return "1" }
func (t Theme) ErrorFg() string { return "15" }
