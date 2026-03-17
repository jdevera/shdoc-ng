package shdoc

import "testing"

func TestParseSeeRef(t *testing.T) {
	tests := []struct {
		input string
		want  SeeRef
	}{
		// Pure URL
		{
			input: "https://github.com/reconquest/shdoc",
			want:  SeeRef{Kind: "url", Href: "https://github.com/reconquest/shdoc"},
		},
		// URL with trailing text should be "text", not "url"
		{
			input: "https://example.com see also other stuff",
			want:  SeeRef{Kind: "text", Text: "https://example.com see also other stuff"},
		},
		// Relative path
		{
			input: "./some/relative/path",
			want:  SeeRef{Kind: "path", Href: "./some/relative/path"},
		},
		// Absolute path
		{
			input: "/some/absolute/path",
			want:  SeeRef{Kind: "path", Href: "/some/absolute/path"},
		},
		// Markdown link
		{
			input: "[shdoc](https://github.com/reconquest/shdoc)",
			want:  SeeRef{Kind: "link", Text: "shdoc", Href: "https://github.com/reconquest/shdoc"},
		},
		// Plain anchor reference
		{
			input: "some-function",
			want:  SeeRef{Kind: "ref", Text: "some-function"},
		},
		// Text with embedded URL
		{
			input: "shdoc: https://github.com/reconquest/shdoc",
			want:  SeeRef{Kind: "text", Text: "shdoc: https://github.com/reconquest/shdoc"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSeeRef(tt.input)
			if got != tt.want {
				t.Errorf("parseSeeRef(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
