package shdoc

import (
	"strings"
)

// unindent removes common leading whitespace from text lines.
// Matches the awk implementation precisely.
func unindent(text string) string {
	lines := strings.Split(text, "\n")

	// Find first non-empty line and max indent
	// Use -1 as sentinel since Go arrays are 0-indexed (awk uses 1-indexed)
	start := -1
	maxIndent := 0
	for i := 0; i < len(lines); i++ {
		if lines[i] != "" && start == -1 {
			start = i
		}
		// Count leading spaces
		spaces := 0
		for _, ch := range lines[i] {
			if ch == ' ' {
				spaces++
			} else {
				break
			}
		}
		if spaces > maxIndent {
			maxIndent = spaces
		}
	}

	// If no non-empty lines found, return empty
	if start == -1 {
		return ""
	}

	// Find minimum indent from start
	indent := maxIndent
	for i := start; i < len(lines); i++ {
		spaces := 0
		for _, ch := range lines[i] {
			if ch == ' ' {
				spaces++
			} else {
				break
			}
		}
		if spaces < indent {
			indent = spaces
		}
	}

	// Remove indent and join from start
	var result strings.Builder
	for i := start; i < len(lines); i++ {
		if i > start {
			result.WriteString("\n")
		}
		if len(lines[i]) > indent {
			result.WriteString(lines[i][indent:])
		} else {
			result.WriteString(lines[i])
		}
	}

	return result.String()
}

