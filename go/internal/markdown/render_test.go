package markdown

import (
	"strings"
	"testing"

	"github.com/aporicho/markcli/internal/ansi"
)

func TestRenderToLines_NonEmpty(t *testing.T) {
	lines, err := RenderToLines("# Hello\n\nSome content here.", 80)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Error("expected non-empty lines")
	}
}

func TestRenderToLines_TitlePresent(t *testing.T) {
	lines, err := RenderToLines("# Title\n\nBody text.", 80)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, l := range lines {
		if strings.Contains(l, "Title") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Title' in rendered output, got: %v", lines)
	}
}

func TestRenderToLines_CJK(t *testing.T) {
	content := "# 标题\n\n这是中文内容，包含 CJK 字符。"
	lines, err := RenderToLines(content, 80)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Error("expected non-empty lines for CJK content")
	}
	// verify DisplayWidth doesn't panic on CJK ANSI output
	for _, l := range lines {
		_ = ansi.DisplayWidth(ansi.StripAnsi(l))
	}
}

func TestRenderToLines_WordWrap(t *testing.T) {
	content := strings.Repeat("word ", 20)
	lines, err := RenderToLines(content, 40)
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) == 0 {
		t.Error("expected lines")
	}
	// glamour adds ~4 chars of margin on each side; 40+8=48, use 60 as safe upper bound
	for _, l := range lines {
		w := ansi.DisplayWidth(ansi.StripAnsi(l))
		if w > 60 {
			t.Errorf("line too wide (%d > 60): %q", w, l)
		}
	}
}
