package main

import (
	"regexp"
	"strings"
)

// LineKind classifies each source line.
type LineKind int

const (
	LineComment LineKind = iota // starts with optional whitespace then #
	LineBlank                   // empty or whitespace only
	LineCode                    // everything else
)

// LexedLine is a source line with its classification.
type LexedLine struct {
	Raw  string
	Kind LineKind
	Num  int // 1-based line number
}

var (
	lexCommentRe = regexp.MustCompile(`^\s*#`)
	lexBlankRe   = regexp.MustCompile(`^\s*$`)
)

// lexLines classifies every line in src.
func lexLines(src string) []LexedLine {
	rawLines := strings.Split(src, "\n")
	result := make([]LexedLine, 0, len(rawLines))
	for i, raw := range rawLines {
		var kind LineKind
		switch {
		case lexCommentRe.MatchString(raw):
			kind = LineComment
		case lexBlankRe.MatchString(raw):
			kind = LineBlank
		default:
			kind = LineCode
		}
		result = append(result, LexedLine{Raw: raw, Kind: kind, Num: i + 1})
	}
	return result
}
