package ui

import (
	"testing"

	"github.com/aporicho/markcli/internal/annotation"
)

func intPtr(i int) *int { return &i }
func boolPtr(b bool) *bool { return &b }

func makeAnn(startLine, endLine int, startCol, endCol *int, resolved *bool) annotation.Annotation {
	return annotation.Annotation{
		ID:        "test",
		StartLine: startLine,
		EndLine:   endLine,
		StartCol:  startCol,
		EndCol:    endCol,
		Resolved:  resolved,
	}
}

// ---- NormalizePos ----

func TestNormalizePos(t *testing.T) {
	a := annotation.SelectionPos{Line: 1, Col: 0}
	b := annotation.SelectionPos{Line: 2, Col: 5}
	s, e := NormalizePos(a, b)
	if s != a || e != b {
		t.Error("already sorted should be unchanged")
	}

	// reversed
	c := annotation.SelectionPos{Line: 3, Col: 5}
	d := annotation.SelectionPos{Line: 1, Col: 2}
	s, e = NormalizePos(c, d)
	if s != d || e != c {
		t.Error("reversed positions should be swapped")
	}

	// same line, col order
	e1 := annotation.SelectionPos{Line: 1, Col: 10}
	e2 := annotation.SelectionPos{Line: 1, Col: 3}
	s, e = NormalizePos(e1, e2)
	if s != e2 || e != e1 {
		t.Error("same line should sort by col")
	}
}

// ---- GetAnnotatedRangesForLine ----

func TestGetAnnotatedRangesForLine_empty(t *testing.T) {
	got := GetAnnotatedRangesForLine(nil, 1, 10)
	if len(got) != 0 {
		t.Errorf("expected empty, got %v", got)
	}
}

func TestGetAnnotatedRangesForLine_outsideRange(t *testing.T) {
	anns := []annotation.Annotation{makeAnn(2, 3, nil, nil, nil)}
	if len(GetAnnotatedRangesForLine(anns, 1, 10)) != 0 {
		t.Error("line 1 outside ann(2,3)")
	}
	if len(GetAnnotatedRangesForLine(anns, 4, 10)) != 0 {
		t.Error("line 4 outside ann(2,3)")
	}
}

func TestGetAnnotatedRangesForLine_singleLineNoCol(t *testing.T) {
	anns := []annotation.Annotation{makeAnn(2, 2, nil, nil, nil)}
	got := GetAnnotatedRangesForLine(anns, 2, 10)
	if len(got) != 1 || got[0].Start != 0 || got[0].End != 10 || got[0].Index != 0 {
		t.Errorf("unexpected: %v", got)
	}
}

func TestGetAnnotatedRangesForLine_singleLineWithCol(t *testing.T) {
	anns := []annotation.Annotation{makeAnn(2, 2, intPtr(3), intPtr(7), nil)}
	got := GetAnnotatedRangesForLine(anns, 2, 10)
	if len(got) != 1 || got[0].Start != 3 || got[0].End != 7 {
		t.Errorf("unexpected: %v", got)
	}
}

func TestGetAnnotatedRangesForLine_multiLine(t *testing.T) {
	anns := []annotation.Annotation{makeAnn(2, 5, intPtr(4), intPtr(8), nil)}

	// start line
	got := GetAnnotatedRangesForLine(anns, 2, 10)
	if len(got) != 1 || got[0].Start != 4 || got[0].End != 10 {
		t.Errorf("start line: %v", got)
	}

	// middle line
	got = GetAnnotatedRangesForLine(anns, 3, 10)
	if len(got) != 1 || got[0].Start != 0 || got[0].End != 10 {
		t.Errorf("middle line: %v", got)
	}

	// end line
	got = GetAnnotatedRangesForLine(anns, 5, 10)
	if len(got) != 1 || got[0].Start != 0 || got[0].End != 8 {
		t.Errorf("end line: %v", got)
	}
}

func TestGetAnnotatedRangesForLine_skipsResolved(t *testing.T) {
	anns := []annotation.Annotation{
		makeAnn(1, 1, intPtr(0), intPtr(5), boolPtr(true)),
		makeAnn(1, 1, intPtr(5), intPtr(10), nil),
	}
	got := GetAnnotatedRangesForLine(anns, 1, 10)
	if len(got) != 1 || got[0].Index != 1 {
		t.Errorf("should skip resolved, got %v", got)
	}
}

func TestGetAnnotatedRangesForLine_overlapping(t *testing.T) {
	anns := []annotation.Annotation{
		makeAnn(1, 1, intPtr(0), intPtr(6), nil),
		makeAnn(1, 1, intPtr(4), intPtr(10), nil),
	}
	got := GetAnnotatedRangesForLine(anns, 1, 10)
	if len(got) != 2 {
		t.Fatalf("expected 2 ranges, got %d", len(got))
	}
	if got[0].Start != 0 || got[0].End != 6 || got[0].Index != 0 {
		t.Errorf("range 0: %v", got[0])
	}
	if got[1].Start != 4 || got[1].End != 10 || got[1].Index != 1 {
		t.Errorf("range 1: %v", got[1])
	}
}

// ---- GetSelectionRangeForLine ----

func TestGetSelectionRangeForLine_outside(t *testing.T) {
	_, _, ok := GetSelectionRangeForLine(1, 10,
		annotation.SelectionPos{Line: 2, Col: 0},
		annotation.SelectionPos{Line: 3, Col: 5})
	if ok {
		t.Error("line 1 outside selection [2,3]")
	}
}

func TestGetSelectionRangeForLine_singleLine(t *testing.T) {
	s, e, ok := GetSelectionRangeForLine(2, 10,
		annotation.SelectionPos{Line: 2, Col: 3},
		annotation.SelectionPos{Line: 2, Col: 7})
	if !ok || s != 3 || e != 7 {
		t.Errorf("got (%d,%d,%v)", s, e, ok)
	}
}

func TestGetSelectionRangeForLine_multiLine(t *testing.T) {
	start := annotation.SelectionPos{Line: 2, Col: 4}
	end := annotation.SelectionPos{Line: 5, Col: 6}

	s, e, ok := GetSelectionRangeForLine(2, 10, start, end)
	if !ok || s != 4 || e != 10 {
		t.Errorf("start line: (%d,%d,%v)", s, e, ok)
	}

	s, e, ok = GetSelectionRangeForLine(3, 10, start, end)
	if !ok || s != 0 || e != 10 {
		t.Errorf("mid line: (%d,%d,%v)", s, e, ok)
	}

	s, e, ok = GetSelectionRangeForLine(5, 10, start, end)
	if !ok || s != 0 || e != 6 {
		t.Errorf("end line: (%d,%d,%v)", s, e, ok)
	}
}

func TestGetSelectionRangeForLine_zeroWidth(t *testing.T) {
	_, _, ok := GetSelectionRangeForLine(1, 10,
		annotation.SelectionPos{Line: 1, Col: 5},
		annotation.SelectionPos{Line: 1, Col: 5})
	if ok {
		t.Error("zero-width selection should return false")
	}
}

// ---- BuildSegments ----

func segEq(a, b Segment) bool {
	if a.Text != b.Text || a.Selected != b.Selected {
		return false
	}
	if (a.AnnotationIndex == nil) != (b.AnnotationIndex == nil) {
		return false
	}
	if a.AnnotationIndex != nil && *a.AnnotationIndex != *b.AnnotationIndex {
		return false
	}
	if (a.ResolvedIndex == nil) != (b.ResolvedIndex == nil) {
		return false
	}
	if a.ResolvedIndex != nil && *a.ResolvedIndex != *b.ResolvedIndex {
		return false
	}
	return true
}

func TestBuildSegments_empty(t *testing.T) {
	got := BuildSegments("", nil, nil, nil)
	if len(got) != 1 || got[0].Text != " " || got[0].Selected || got[0].AnnotationIndex != nil {
		t.Errorf("empty string: %v", got)
	}
}

func TestBuildSegments_noHighlights(t *testing.T) {
	got := BuildSegments("hello", nil, nil, nil)
	want := []Segment{{Text: "hello"}}
	if len(got) != 1 || !segEq(got[0], want[0]) {
		t.Errorf("got %v", got)
	}
}

func TestBuildSegments_selectionRange(t *testing.T) {
	got := BuildSegments("abcde", &[2]int{1, 3}, nil, nil)
	if len(got) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(got))
	}
	if got[0].Text != "a" || got[0].Selected {
		t.Errorf("seg0: %v", got[0])
	}
	if got[1].Text != "bc" || !got[1].Selected {
		t.Errorf("seg1: %v", got[1])
	}
	if got[2].Text != "de" || got[2].Selected {
		t.Errorf("seg2: %v", got[2])
	}
}

func TestBuildSegments_annotationRange(t *testing.T) {
	got := BuildSegments("abcde", nil, []AnnotatedRange{{Start: 2, End: 4, Index: 0}}, nil)
	if len(got) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(got))
	}
	if got[0].Text != "ab" || got[0].AnnotationIndex != nil {
		t.Errorf("seg0: %v", got[0])
	}
	if got[1].Text != "cd" || got[1].AnnotationIndex == nil || *got[1].AnnotationIndex != 0 {
		t.Errorf("seg1: %v", got[1])
	}
	if got[2].Text != "e" || got[2].AnnotationIndex != nil {
		t.Errorf("seg2: %v", got[2])
	}
}

func TestBuildSegments_overlappingSelAndAnn(t *testing.T) {
	// sel:[1,4], ann:[3,6] on "abcdef"
	got := BuildSegments("abcdef", &[2]int{1, 4}, []AnnotatedRange{{Start: 3, End: 6, Index: 0}}, nil)
	if len(got) != 4 {
		t.Fatalf("expected 4 segments, got %d: %v", len(got), got)
	}
	// "a" not sel, not ann
	if got[0].Text != "a" || got[0].Selected || got[0].AnnotationIndex != nil {
		t.Errorf("seg0: %v", got[0])
	}
	// "bc" sel, not ann
	if got[1].Text != "bc" || !got[1].Selected || got[1].AnnotationIndex != nil {
		t.Errorf("seg1: %v", got[1])
	}
	// "d" sel and ann
	if got[2].Text != "d" || !got[2].Selected || got[2].AnnotationIndex == nil || *got[2].AnnotationIndex != 0 {
		t.Errorf("seg2: %v", got[2])
	}
	// "ef" not sel, ann
	if got[3].Text != "ef" || got[3].Selected || got[3].AnnotationIndex == nil || *got[3].AnnotationIndex != 0 {
		t.Errorf("seg3: %v", got[3])
	}
}

func TestBuildSegments_multipleAnnotations(t *testing.T) {
	got := BuildSegments("abcdefgh", nil, []AnnotatedRange{
		{Start: 1, End: 3, Index: 0},
		{Start: 5, End: 7, Index: 1},
	}, nil)
	if len(got) != 5 {
		t.Fatalf("expected 5 segments, got %d: %v", len(got), got)
	}
	texts := []string{"a", "bc", "de", "fg", "h"}
	for i, tt := range texts {
		if got[i].Text != tt {
			t.Errorf("seg%d text = %q, want %q", i, got[i].Text, tt)
		}
	}
	if got[1].AnnotationIndex == nil || *got[1].AnnotationIndex != 0 {
		t.Errorf("seg1 ann index wrong")
	}
	if got[3].AnnotationIndex == nil || *got[3].AnnotationIndex != 1 {
		t.Errorf("seg3 ann index wrong")
	}
}

func TestBuildSegments_alternatingIndices(t *testing.T) {
	got := BuildSegments("abcdefghi", nil, []AnnotatedRange{
		{Start: 0, End: 3, Index: 0},
		{Start: 3, End: 6, Index: 1},
		{Start: 6, End: 9, Index: 2},
	}, nil)
	if len(got) != 3 {
		t.Fatalf("expected 3 segments, got %d", len(got))
	}
	for i, want := range []int{0, 1, 2} {
		if got[i].AnnotationIndex == nil || *got[i].AnnotationIndex != want {
			t.Errorf("seg%d index = %v, want %d", i, got[i].AnnotationIndex, want)
		}
	}
	// even/odd alternation
	if *got[0].AnnotationIndex%2 != 0 || *got[1].AnnotationIndex%2 != 1 || *got[2].AnnotationIndex%2 != 0 {
		t.Error("alternating indices mismatch")
	}
}
