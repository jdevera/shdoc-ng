package lsp

import (
	"strings"
	"testing"
)

func TestInSingleQuotes(t *testing.T) {
	tests := []struct {
		line string
		pos  int
		want bool
	}{
		{"echo hello", 5, false},
		{"echo 'hello'", 6, true},
		{"echo 'hello'", 12, false},
		{"echo 'he'llo", 10, false},
		{"echo ''", 6, true},
		{"'quoted' unquoted", 3, true},
		{"'quoted' unquoted", 10, false},
	}
	for _, tc := range tests {
		got := inSingleQuotes(tc.line, tc.pos)
		if got != tc.want {
			t.Errorf("inSingleQuotes(%q, %d) = %v, want %v", tc.line, tc.pos, got, tc.want)
		}
	}
}

func TestCursorOnTag(t *testing.T) {
	tests := []struct {
		raw       string
		cursorCol int
		want      string
	}{
		{"# @description Hello", 2, "description"},
		{"# @description Hello", 3, "description"},
		{"# @description Hello", 14, ""}, // past end of tag keyword
		{"# @description Hello", 15, ""},
		{"# @desc Hello", 3, "description"}, // shorthand normalized
		{"# not a tag", 3, ""},
		{"echo hello", 0, ""},
		{"# @arg $1 string Name", 2, "arg"},
		{"# @see foo()", 2, "see"},
	}
	for _, tc := range tests {
		got := cursorOnTag(tc.raw, tc.cursorCol)
		if got != tc.want {
			t.Errorf("cursorOnTag(%q, %d) = %q, want %q", tc.raw, tc.cursorCol, got, tc.want)
		}
	}
}

func TestDocBlockSkeleton(t *testing.T) {
	got := docBlockSkeleton("my_func")
	if got == "" {
		t.Fatal("docBlockSkeleton returned empty string")
	}
	// Should mention the function name.
	if !strings.Contains(got, "my_func") {
		t.Errorf("skeleton does not contain function name: %q", got)
	}
	// Should contain key tags.
	for _, tag := range []string{"@description", "@arg", "@exitcode"} {
		if !strings.Contains(got, tag) {
			t.Errorf("skeleton missing %s: %q", tag, got)
		}
	}
}

func TestCountBraces(t *testing.T) {
	tests := []struct {
		name  string
		line  string
		start int
		want  int
	}{
		{"open brace", "func() {", 0, 1},
		{"close brace", "}", 0, -1},
		{"balanced", "{ foo; }", 0, 0},
		{"in single quotes", "echo '{'", 0, 0},
		{"in double quotes", `echo "{"`, 0, 0},
		{"after comment", "echo hello # {", 0, 0},
		{"escaped in dquote", `echo "\{"`, 0, 0},
		{"mixed", `{ echo '}' "}" # }`, 0, 1},
		{"nested unbalanced", "{ { inner; }", 0, 1},
		{"accumulate", "{ {", 0, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			depth := tc.start
			countBraces(tc.line, &depth)
			if depth != tc.want {
				t.Errorf("countBraces(%q) depth = %d, want %d", tc.line, depth, tc.want)
			}
		})
	}
}
