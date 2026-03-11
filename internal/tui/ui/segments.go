package ui

import (
	"sort"

	"github.com/aporicho/markcli/internal/annotation"
)

type AnnotatedRange struct {
	Start int
	End   int
	Index int
}

type Segment struct {
	Text            string
	Selected        bool
	AnnotationIndex *int
	ResolvedIndex   *int
}

func NormalizePos(a, b annotation.SelectionPos) (annotation.SelectionPos, annotation.SelectionPos) {
	if a.Line < b.Line || (a.Line == b.Line && a.Col <= b.Col) {
		return a, b
	}
	return b, a
}

func GetAnnotatedRangesForLine(annotations []annotation.Annotation, lineNum, lineLength int) []AnnotatedRange {
	var ranges []AnnotatedRange
	for i, ann := range annotations {
		if annotation.IsResolved(ann) {
			continue
		}
		if r := annotationRangeForLine(ann, lineNum, lineLength); r != nil {
			ranges = append(ranges, AnnotatedRange{Start: r[0], End: r[1], Index: i})
		}
	}
	return ranges
}

func GetResolvedRangesForLine(annotations []annotation.Annotation, lineNum, lineLength int) []AnnotatedRange {
	var ranges []AnnotatedRange
	for i, ann := range annotations {
		if !annotation.IsResolved(ann) {
			continue
		}
		if r := annotationRangeForLine(ann, lineNum, lineLength); r != nil {
			ranges = append(ranges, AnnotatedRange{Start: r[0], End: r[1], Index: i})
		}
	}
	return ranges
}

func annotationRangeForLine(ann annotation.Annotation, lineNum, lineLength int) *[2]int {
	if lineNum < ann.StartLine || lineNum > ann.EndLine {
		return nil
	}

	hasCol := ann.StartCol != nil && ann.EndCol != nil
	var start, end int

	if ann.StartLine == ann.EndLine {
		if hasCol {
			start, end = *ann.StartCol, *ann.EndCol
		} else {
			start, end = 0, lineLength
		}
	} else if lineNum == ann.StartLine {
		if hasCol {
			start = *ann.StartCol
		}
		end = lineLength
	} else if lineNum == ann.EndLine {
		start = 0
		if hasCol {
			end = *ann.EndCol
		} else {
			end = lineLength
		}
	} else {
		start, end = 0, lineLength
	}

	if start < 0 {
		start = 0
	}
	if end > lineLength {
		end = lineLength
	}
	if end <= start {
		return nil
	}
	return &[2]int{start, end}
}

func GetSelectionRangeForLine(lineNum, lineLength int, normStart, normEnd annotation.SelectionPos) (int, int, bool) {
	if lineNum < normStart.Line || lineNum > normEnd.Line {
		return 0, 0, false
	}

	isSingleLine := normStart.Line == normEnd.Line
	var s, e int

	if isSingleLine {
		s, e = normStart.Col, normEnd.Col
	} else if lineNum == normStart.Line {
		s, e = normStart.Col, lineLength
	} else if lineNum == normEnd.Line {
		s, e = 0, normEnd.Col
	} else {
		s, e = 0, lineLength
	}

	if s < 0 {
		s = 0
	}
	if s > lineLength {
		s = lineLength
	}
	if e < s {
		e = s
	}
	if e > lineLength {
		e = lineLength
	}
	if e <= s {
		return 0, 0, false
	}
	return s, e, true
}

func BuildSegments(stripped string, selRange *[2]int, annRanges, resolvedRanges []AnnotatedRange) []Segment {
	runes := []rune(stripped)
	if len(runes) == 0 {
		// Return a single space so the UI always has content to render
		// (prevents line height collapse for empty lines).
		return []Segment{{Text: " "}}
	}

	cuts := map[int]struct{}{
		0:          {},
		len(runes): {},
	}
	if selRange != nil {
		cuts[selRange[0]] = struct{}{}
		cuts[selRange[1]] = struct{}{}
	}
	for _, r := range annRanges {
		cuts[r.Start] = struct{}{}
		cuts[r.End] = struct{}{}
	}
	for _, r := range resolvedRanges {
		cuts[r.Start] = struct{}{}
		cuts[r.End] = struct{}{}
	}

	sortedCuts := make([]int, 0, len(cuts))
	for c := range cuts {
		sortedCuts = append(sortedCuts, c)
	}
	sort.Ints(sortedCuts)

	var segments []Segment
	for i := 0; i < len(sortedCuts)-1; i++ {
		start := sortedCuts[i]
		end := sortedCuts[i+1]
		if start >= end {
			continue
		}

		text := string(runes[start:end])
		mid := start

		selected := selRange != nil && mid >= selRange[0] && mid < selRange[1]

		var annIdx *int
		for _, r := range annRanges {
			if mid >= r.Start && mid < r.End {
				idx := r.Index
				annIdx = &idx
				break
			}
		}

		var resIdx *int
		for _, r := range resolvedRanges {
			if mid >= r.Start && mid < r.End {
				idx := r.Index
				resIdx = &idx
				break
			}
		}

		segments = append(segments, Segment{
			Text:            text,
			Selected:        selected,
			AnnotationIndex: annIdx,
			ResolvedIndex:   resIdx,
		})
	}

	if len(segments) == 0 {
		// Same as above: ensure at least one renderable segment.
		return []Segment{{Text: " "}}
	}
	return segments
}
