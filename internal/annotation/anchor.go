package annotation

import (
	"strings"
	"unicode/utf8"
)

const contextChars = 30

// TextAnchor holds quote + surrounding context for relocating text across edits.
type TextAnchor struct {
	Quote  string
	Prefix string
	Suffix string
}

// TextRange is a half-open [Start, End) interval in rune offsets.
type TextRange struct {
	Start int
	End   int
}

// LineColToOffset converts (line, col) to a rune offset in the concatenated text.
// line is 1-based, col is 0-based. lineLengths[i] = rune count of line i (without \n).
func LineColToOffset(lineLengths []int, line, col int) int {
	offset := 0
	for i := 0; i < line-1 && i < len(lineLengths); i++ {
		offset += lineLengths[i] + 1 // +1 for \n
	}
	return offset + col
}

// OffsetToLineCol is the inverse of LineColToOffset.
// Returns line (1-based) and col (0-based, clamped to line length).
func OffsetToLineCol(lineLengths []int, offset int) (line, col int) {
	remaining := offset
	for i := 0; i < len(lineLengths); i++ {
		length := lineLengths[i] + 1 // +1 for \n
		if remaining < length || i == len(lineLengths)-1 {
			col = remaining
			if col > lineLengths[i] {
				col = lineLengths[i]
			}
			return i + 1, col
		}
		remaining -= length
	}
	return 1, 0
}

// ExtractAnchor extracts a TextAnchor from fullText at rune offsets [start, end).
func ExtractAnchor(fullText string, start, end int) TextAnchor {
	runes := []rune(fullText)
	prefixStart := start - contextChars
	if prefixStart < 0 {
		prefixStart = 0
	}
	suffixEnd := end + contextChars
	if suffixEnd > len(runes) {
		suffixEnd = len(runes)
	}
	return TextAnchor{
		Quote:  string(runes[start:end]),
		Prefix: string(runes[prefixStart:start]),
		Suffix: string(runes[end:suffixEnd]),
	}
}

// RelocateAnchor finds the best match for anchor in fullText.
// Returns nil if no match found.
func RelocateAnchor(fullText string, anchor TextAnchor) *TextRange {
	if anchor.Quote == "" {
		return nil
	}

	// Strategy 1: exact search
	var exactMatches []int
	idx := 0
	for {
		pos := strings.Index(fullText[idx:], anchor.Quote)
		if pos == -1 {
			break
		}
		bytePos := idx + pos
		// convert byte pos to rune pos
		runePos := utf8.RuneCountInString(fullText[:bytePos])
		exactMatches = append(exactMatches, runePos)
		// advance by one byte to find next
		_, size := utf8.DecodeRuneInString(fullText[bytePos:])
		idx = bytePos + size
	}

	quoteRunes := []rune(anchor.Quote)
	quoteRuneLen := len(quoteRunes)

	if len(exactMatches) == 1 {
		return &TextRange{Start: exactMatches[0], End: exactMatches[0] + quoteRuneLen}
	}

	if len(exactMatches) > 1 {
		// Strategy 2: disambiguate with prefix/suffix
		bestIdx := exactMatches[0]
		bestScore := -1
		fullRunes := []rune(fullText)

		for _, matchRune := range exactMatches {
			score := 0
			if anchor.Prefix != "" {
				prefixRunes := []rune(anchor.Prefix)
				start := matchRune - len(prefixRunes)
				if start < 0 {
					start = 0
				}
				before := string(fullRunes[start:matchRune])
				score += commonSuffixLength(before, anchor.Prefix)
			}
			if anchor.Suffix != "" {
				suffixRunes := []rune(anchor.Suffix)
				end := matchRune + quoteRuneLen + len(suffixRunes)
				if end > len(fullRunes) {
					end = len(fullRunes)
				}
				after := string(fullRunes[matchRune+quoteRuneLen : end])
				score += commonPrefixLength(after, anchor.Suffix)
			}
			if score > bestScore {
				bestScore = score
				bestIdx = matchRune
			}
		}
		return &TextRange{Start: bestIdx, End: bestIdx + quoteRuneLen}
	}

	// Strategy 3: Bitap fuzzy match
	maxErrors := quoteRuneLen * 20 / 100
	if maxErrors < 1 {
		maxErrors = 1
	}
	result := bitapSearch([]rune(fullText), quoteRunes, maxErrors)
	if result == nil {
		return nil
	}
	return result
}

// bitapSearch performs approximate string matching using the Bitap algorithm.
// Returns the match with fewest errors (earliest on tie).
func bitapSearch(text, pattern []rune, maxErrors int) *TextRange {
	m := len(pattern)
	if m == 0 {
		return nil
	}

	// For long patterns, use first 30 runes to find candidates, then verify
	const maxPatternLen = 63
	if m > maxPatternLen {
		return bitapSearchLong(text, pattern, maxErrors)
	}

	// Build pattern bitmask: patMask[c] has bit i set if pattern[i] == c
	patMask := make(map[rune]uint64)
	for i, c := range pattern {
		patMask[c] |= 1 << uint(i)
	}

	type match struct {
		start  int
		end    int
		errors int
	}
	var best *match

	// R[k] = bitmask of states reachable with k errors
	R := make([]uint64, maxErrors+1)
	acceptBit := uint64(1) << uint(m-1)

	n := len(text)
	for j := 0; j < n; j++ {
		// Update R arrays from right to left to avoid using updated values
		charMask := patMask[text[j]]
		var prevR [64 + 1]uint64
		for k := 0; k <= maxErrors; k++ {
			prevR[k] = R[k]
		}

		R[0] = ((prevR[0] << 1) | 1) & charMask
		for k := 1; k <= maxErrors; k++ {
			R[k] = (((prevR[k] << 1) | 1) & charMask) | // match / new start
				(prevR[k-1] << 1) | // substitution or deletion in pattern
				prevR[k-1] // insertion in text
		}

		for k := 0; k <= maxErrors; k++ {
			if R[k]&acceptBit != 0 {
				end := j + 1
				start := end - m - k
				if start < 0 {
					start = 0
				}
				if best == nil || k < best.errors || (k == best.errors && start < best.start) {
					best = &match{start: start, end: end, errors: k}
				}
			}
		}
	}

	if best == nil {
		return nil
	}
	return &TextRange{Start: best.start, End: best.end}
}

// bitapSearchLong handles patterns longer than 63 runes by using first 30 runes
// to find candidates, then doing full comparison in candidate windows.
func bitapSearchLong(text, pattern []rune, maxErrors int) *TextRange {
	const anchorLen = 30
	anchor := pattern[:anchorLen]

	// Find approximate positions using anchor
	type match struct {
		start  int
		end    int
		errors int
	}
	var best *match

	// Search for anchor with tight tolerance
	anchorResult := bitapSearch(text, anchor, 2)
	if anchorResult == nil {
		return nil
	}

	// Search window around anchor match
	windowStart := anchorResult.Start - maxErrors
	if windowStart < 0 {
		windowStart = 0
	}
	windowEnd := anchorResult.Start + len(pattern) + maxErrors
	if windowEnd > len(text) {
		windowEnd = len(text)
	}

	window := text[windowStart:windowEnd]
	for i := 0; i <= len(window)-len(pattern); i++ {
		errors := editDistance(window[i:i+len(pattern)], pattern)
		if errors <= maxErrors {
			start := windowStart + i
			end := start + len(pattern)
			if best == nil || errors < best.errors {
				best = &match{start: start, end: end, errors: errors}
			}
		}
	}

	if best == nil {
		return nil
	}
	return &TextRange{Start: best.start, End: best.end}
}

// editDistance computes Levenshtein distance between two rune slices.
func editDistance(a, b []rune) int {
	m, n := len(a), len(b)
	dp := make([]int, n+1)
	for j := range dp {
		dp[j] = j
	}
	for i := 1; i <= m; i++ {
		prev := dp[0]
		dp[0] = i
		for j := 1; j <= n; j++ {
			temp := dp[j]
			if a[i-1] == b[j-1] {
				dp[j] = prev
			} else {
				dp[j] = 1 + min3(prev, dp[j], dp[j-1])
			}
			prev = temp
		}
	}
	return dp[n]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}

func commonSuffixLength(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	n := 0
	maxLen := len(ra)
	if len(rb) < maxLen {
		maxLen = len(rb)
	}
	for i := 0; i < maxLen; i++ {
		if ra[len(ra)-1-i] == rb[len(rb)-1-i] {
			n++
		} else {
			break
		}
	}
	return n
}

func commonPrefixLength(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	n := 0
	maxLen := len(ra)
	if len(rb) < maxLen {
		maxLen = len(rb)
	}
	for i := 0; i < maxLen; i++ {
		if ra[i] == rb[i] {
			n++
		} else {
			break
		}
	}
	return n
}
