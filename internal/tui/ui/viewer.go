package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aporicho/markcli/internal/annotation"
	"github.com/aporicho/markcli/internal/theme"
)

// RenderViewer renders the visible slice of lines with selection/annotation highlighting.
// selRange is [startLine, startCol, endLine, endCol] (1-based lines, 0-based cols), nil = no selection.
func RenderViewer(
	renderedLines []string,
	strippedLines []string,
	lineLengths []int,
	annotations []annotation.Annotation,
	scrollOffset int,
	viewportHeight int,
	t theme.Theme,
	selRange *[4]int,
) string {
	var sb strings.Builder

	end := scrollOffset + viewportHeight
	if end > len(renderedLines) {
		end = len(renderedLines)
	}

	linesRendered := 0
	for lineIdx := scrollOffset; lineIdx < end; lineIdx++ {
		lineNum := lineIdx + 1 // 1-based

		lineLen := 0
		if lineIdx < len(lineLengths) {
			lineLen = lineLengths[lineIdx]
		}

		rendered := ""
		if lineIdx < len(renderedLines) {
			rendered = renderedLines[lineIdx]
		}

		stripped := ""
		if lineIdx < len(strippedLines) {
			stripped = strippedLines[lineIdx]
		}

		annRanges := GetAnnotatedRangesForLine(annotations, lineNum, lineLen)
		resolvedRanges := GetResolvedRangesForLine(annotations, lineNum, lineLen)

		var sel *[2]int
		if selRange != nil {
			normStart := annotation.SelectionPos{Line: selRange[0], Col: selRange[1]}
			normEnd := annotation.SelectionPos{Line: selRange[2], Col: selRange[3]}
			if s, e, ok := GetSelectionRangeForLine(lineNum, lineLen, normStart, normEnd); ok {
				sel = &[2]int{s, e}
			}
		}

		if len(annRanges) == 0 && len(resolvedRanges) == 0 && sel == nil {
			sb.WriteString(rendered)
			sb.WriteString("\n")
		} else {
			segments := BuildSegments(stripped, sel, annRanges, resolvedRanges)
			for _, seg := range segments {
				sb.WriteString(colorSegment(seg, t))
			}
			sb.WriteString("\n")
		}
		linesRendered++
	}

	// Pad remaining lines to fill viewport
	for linesRendered < viewportHeight {
		sb.WriteString("\n")
		linesRendered++
	}

	return sb.String()
}

func colorSegment(seg Segment, t theme.Theme) string {
	style := lipgloss.NewStyle()

	if seg.Selected {
		if t.Selection.Fg != "" {
			style = style.Foreground(lipgloss.Color(t.Selection.Fg))
		}
		if t.Selection.Bg != "" {
			style = style.Background(lipgloss.Color(t.Selection.Bg))
		}
	} else if seg.AnnotationIndex != nil {
		var c theme.Color
		if *seg.AnnotationIndex%2 == 0 {
			c = t.Annotation
		} else {
			c = t.AnnotationAlt
		}
		if c.Fg != "" {
			style = style.Foreground(lipgloss.Color(c.Fg))
		}
		if c.Bg != "" {
			style = style.Background(lipgloss.Color(c.Bg))
		}
	} else if seg.ResolvedIndex != nil {
		if t.AnnotationResolved.Fg != "" {
			style = style.Foreground(lipgloss.Color(t.AnnotationResolved.Fg))
		}
		if t.AnnotationResolved.Bg != "" {
			style = style.Background(lipgloss.Color(t.AnnotationResolved.Bg))
		}
	} else {
		return seg.Text
	}

	return style.Render(seg.Text)
}
