package ansi

import "testing"

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"", ""},
		{"hello world", "hello world"},
		{"\x1b[31mred\x1b[0m", "red"},
		{"\x1b[1m\x1b[31mbold red\x1b[0m normal", "bold red normal"},
		{"\x1b[32m你好\x1b[0m世界", "你好世界"},
	}
	for _, tt := range tests {
		got := StripAnsi(tt.input)
		if got != tt.want {
			t.Errorf("StripAnsi(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestTermColToCharIndex(t *testing.T) {
	tests := []struct {
		text    string
		termCol int
		want    int
	}{
		{"abc", 0, 0},
		{"abc", 1, 1},
		{"abc", 2, 2},
		// CJK: "你好" — 你 occupies cols 0-1, 好 occupies cols 2-3
		{"你好", 0, 0},
		{"你好", 1, 0}, // mid of 你
		{"你好", 2, 1},
		// beyond end
		{"ab", 5, 2},
		// empty
		{"", 0, 0},
	}
	for _, tt := range tests {
		got := TermColToCharIndex(tt.text, tt.termCol)
		if got != tt.want {
			t.Errorf("TermColToCharIndex(%q, %d) = %d, want %d", tt.text, tt.termCol, got, tt.want)
		}
	}
}

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		text string
		want int
	}{
		{"", 0},
		{"abc", 3},
		{"你好", 4},
		{"a你b", 4},
	}
	for _, tt := range tests {
		got := DisplayWidth(tt.text)
		if got != tt.want {
			t.Errorf("DisplayWidth(%q) = %d, want %d", tt.text, got, tt.want)
		}
	}
}
