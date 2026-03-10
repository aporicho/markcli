package annotation

import "testing"

// ---- LineColToOffset ----

func TestLineColToOffset(t *testing.T) {
	lengths := []int{3, 3, 3} // "abc\ndef\nghi"

	tests := []struct {
		line, col, want int
	}{
		{1, 0, 0},
		{1, 2, 2},
		{2, 0, 4}, // after "abc\n"
		{3, 1, 9}, // after "abc\ndef\n" + 1
	}
	for _, tt := range tests {
		got := LineColToOffset(lengths, tt.line, tt.col)
		if got != tt.want {
			t.Errorf("LineColToOffset(line=%d, col=%d) = %d, want %d", tt.line, tt.col, got, tt.want)
		}
	}
}

func TestLineColToOffset_singleLine(t *testing.T) {
	if got := LineColToOffset([]int{5}, 1, 3); got != 3 {
		t.Errorf("got %d, want 3", got)
	}
}

// ---- OffsetToLineCol ----

func TestOffsetToLineCol(t *testing.T) {
	lengths := []int{3, 3, 3}

	tests := []struct {
		offset   int
		wantLine int
		wantCol  int
	}{
		{0, 1, 0},
		{4, 2, 0},
		{9, 3, 1},
		{100, 3, 3}, // clamps to line length
	}
	for _, tt := range tests {
		line, col := OffsetToLineCol(lengths, tt.offset)
		if line != tt.wantLine || col != tt.wantCol {
			t.Errorf("OffsetToLineCol(%d) = (%d,%d), want (%d,%d)", tt.offset, line, col, tt.wantLine, tt.wantCol)
		}
	}
}

func TestOffsetToLineCol_roundtrip(t *testing.T) {
	lengths := []int{5, 10, 3}
	for line := 1; line <= 3; line++ {
		for col := 0; col <= lengths[line-1]; col++ {
			offset := LineColToOffset(lengths, line, col)
			gotLine, gotCol := OffsetToLineCol(lengths, offset)
			if gotLine != line || gotCol != col {
				t.Errorf("roundtrip (line=%d, col=%d): got (%d,%d)", line, col, gotLine, gotCol)
			}
		}
	}
}

// ---- ExtractAnchor ----

func TestExtractAnchor(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog."

	t.Run("extracts quote prefix suffix", func(t *testing.T) {
		a := ExtractAnchor(text, 10, 19) // "brown fox"
		if a.Quote != "brown fox" {
			t.Errorf("Quote = %q", a.Quote)
		}
		if a.Prefix != "The quick " {
			t.Errorf("Prefix = %q", a.Prefix)
		}
		if a.Suffix != " jumps over the lazy dog." {
			t.Errorf("Suffix = %q", a.Suffix)
		}
	})

	t.Run("truncates prefix near start", func(t *testing.T) {
		a := ExtractAnchor(text, 0, 3)
		if a.Quote != "The" {
			t.Errorf("Quote = %q", a.Quote)
		}
		if a.Prefix != "" {
			t.Errorf("Prefix = %q, want empty", a.Prefix)
		}
	})

	t.Run("truncates suffix near end", func(t *testing.T) {
		runes := []rune(text)
		n := len(runes)
		a := ExtractAnchor(text, n-4, n) // "dog."
		if a.Quote != "dog." {
			t.Errorf("Quote = %q", a.Quote)
		}
		if a.Suffix != "" {
			t.Errorf("Suffix = %q, want empty", a.Suffix)
		}
	})
}

// ---- RelocateAnchor ----

func TestRelocateAnchor_emptyQuote(t *testing.T) {
	if RelocateAnchor("some text", TextAnchor{}) != nil {
		t.Error("expected nil for empty quote")
	}
}

func TestRelocateAnchor_uniqueExact(t *testing.T) {
	text := "The quick brown fox jumps."
	got := RelocateAnchor(text, TextAnchor{Quote: "brown fox", Prefix: "quick ", Suffix: " jumps"})
	if got == nil {
		t.Fatal("expected match")
	}
	if got.Start != 10 || got.End != 19 {
		t.Errorf("got {%d,%d}, want {10,19}", got.Start, got.End)
	}
}

func TestRelocateAnchor_multipleExact_disambiguate(t *testing.T) {
	text := "abc foo abc foo abc"
	// "foo" at rune index 4 and 12
	got := RelocateAnchor(text, TextAnchor{
		Quote:  "foo",
		Prefix: "abc foo abc ",
		Suffix: " abc",
	})
	if got == nil {
		t.Fatal("expected match")
	}
	if got.Start != 12 || got.End != 15 {
		t.Errorf("got {%d,%d}, want {12,15}", got.Start, got.End)
	}
}

func TestRelocateAnchor_fuzzy(t *testing.T) {
	modified := "The quick brownn foxx jumps over the lazy dog."
	got := RelocateAnchor(modified, TextAnchor{
		Quote:  "brown fox jumps over",
		Prefix: "quick ",
		Suffix: " the",
	})
	if got == nil {
		t.Fatal("expected fuzzy match")
	}
	if got.Start < 8 {
		t.Errorf("Start = %d, want >= 8", got.Start)
	}
}

func TestRelocateAnchor_noMatch(t *testing.T) {
	got := RelocateAnchor("completely unrelated text here", TextAnchor{
		Quote:  "xyzxyzxyzxyzxyzxyz",
		Prefix: "aaa",
		Suffix: "bbb",
	})
	if got != nil {
		t.Errorf("expected nil, got %+v", got)
	}
}

func TestRelocateAnchor_shiftedByInsertion(t *testing.T) {
	text := "INSERTED The quick brown fox."
	got := RelocateAnchor(text, TextAnchor{Quote: "brown fox", Prefix: "quick ", Suffix: "."})
	if got == nil {
		t.Fatal("expected match")
	}
	if got.Start != 19 || got.End != 28 {
		t.Errorf("got {%d,%d}, want {19,28}", got.Start, got.End)
	}
}
