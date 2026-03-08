package shdoc

import "testing"

func TestCommentBlockEndNum(t *testing.T) {
	src := `# line one
# line two
# line three
foo() {
    :
}`
	lines := LexLines(src)
	blocks := SegmentBlocks(lines)

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Comments.StartNum != 1 {
		t.Errorf("StartNum = %d, want 1", b.Comments.StartNum)
	}
	if b.Comments.EndNum != 3 {
		t.Errorf("EndNum = %d, want 3", b.Comments.EndNum)
	}
}

func TestCommentBlockEndNumSingleLine(t *testing.T) {
	src := `# @description Hello.
greet() { :; }`
	lines := LexLines(src)
	blocks := SegmentBlocks(lines)

	if len(blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(blocks))
	}
	b := blocks[0]
	if b.Comments.StartNum != 1 {
		t.Errorf("StartNum = %d, want 1", b.Comments.StartNum)
	}
	if b.Comments.EndNum != 1 {
		t.Errorf("EndNum = %d, want 1 (same as start for single-line block)", b.Comments.EndNum)
	}
}

func TestIsFuncDecl(t *testing.T) {
	tests := []struct {
		line string
		want bool
	}{
		{"foo() {", true},
		{"function bar {", true},
		{"function baz() {", true},
		{"  indented() {", true},
		{"# a comment", false},
		{"echo hello", false},
		{"", false},
	}
	for _, tc := range tests {
		got := IsFuncDecl(tc.line)
		if got != tc.want {
			t.Errorf("IsFuncDecl(%q) = %v, want %v", tc.line, got, tc.want)
		}
	}
}
