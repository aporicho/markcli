package ansi

import (
	"regexp"

	"github.com/mattn/go-runewidth"
)

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func StripAnsi(s string) string {
	return ansiRe.ReplaceAllString(s, "")
}

func DisplayWidth(text string) int {
	return runewidth.StringWidth(text)
}

// TermColToCharIndex maps a terminal column (0-based) to a rune index in text.
// Accounts for double-width characters (CJK etc.).
func TermColToCharIndex(text string, termCol int) int {
	colAcc := 0
	i := 0
	for _, r := range text {
		w := runewidth.RuneWidth(r)
		if colAcc+w > termCol {
			return i
		}
		colAcc += w
		i++
	}
	return i
}
