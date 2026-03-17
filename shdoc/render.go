package shdoc

import (
	"strings"
)

// unindent removes common leading whitespace from text lines.
// Matches the awk implementation precisely.
func unindent(text string) string {
	lines := strings.Split(text, "\n")

	// Find first non-empty line and minimum indent in a single pass.
	start := -1
	indent := -1 // -1 = not yet set
	for i, line := range lines {
		if line == "" {
			continue
		}
		if start == -1 {
			start = i
		}
		spaces := 0
		for _, ch := range line {
			if ch == ' ' {
				spaces++
			} else {
				break
			}
		}
		if indent == -1 || spaces < indent {
			indent = spaces
		}
	}

	// If no non-empty lines found, return empty
	if start == -1 {
		return ""
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

