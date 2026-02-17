package main

import (
	"regexp"
	"strings"
)

// slug generates a GitHub-compatible anchor from text.
// Lowercase, remove chars except alphanumeric/space/underscore/dash,
// replace spaces with dashes. Underscores are preserved (GitHub keeps
// them in heading anchors).
func slug(text string) string {
	s := strings.ToLower(text)

	// Keep only alphanumeric, space, underscore, and dash
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' || r == '_' || r == '-' {
			b.WriteRune(r)
		}
	}
	s = b.String()

	// Replace spaces with dashes
	s = strings.ReplaceAll(s, " ", "-")

	return s
}

// urlPattern matches URLs not already inside markdown links.
// Matches: a non-] character followed by a non-lowercase-letter-or-paren (or a paren), then a URL scheme.
var urlDetectPattern = regexp.MustCompile(`[^\]]([^a-z(]|\()[a-z]+://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]`)

// urlReplacePattern is used to wrap bare URLs in markdown link syntax.
var urlReplacePattern = regexp.MustCompile(`([^\]]([^a-z(]|\())([a-z]+://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|])`)

// markdownLinkPattern detects if text already contains a markdown link.
var markdownLinkPattern = regexp.MustCompile(`\[[^\]]*\]\([^)]*\)`)

// renderTocLink converts text to a markdown link for TOC and @see entries.
// Detection order matches the awk implementation.
func renderTocLink(text string) string {
	// 1. Relative path (starts with ./ or ../)
	if strings.HasPrefix(text, "./") || strings.HasPrefix(text, "../") {
		return "[" + text + "](" + text + ")"
	}

	// 2. Absolute path (starts with /)
	if strings.HasPrefix(text, "/") {
		return "[" + text + "](" + text + ")"
	}

	// 3. Contains URLs not in markdown links - pad with spaces like awk does
	padded := "  " + text + " "
	if urlDetectPattern.MatchString(padded) {
		// Wrap bare URLs in markdown links
		padded = urlReplacePattern.ReplaceAllString(padded, "${1}[${3}](${3})")
		// Trim the padding spaces
		padded = strings.TrimLeft(padded, " ")
		padded = strings.TrimRight(padded, " ")
		return padded
	}

	// 4. Already a markdown link - pass through
	if markdownLinkPattern.MatchString(text) {
		return text
	}

	// 5. Plain text → anchor link
	return "[" + text + "](#" + slug(text) + ")"
}

// renderTocItem renders a TOC entry for a function name.
func renderTocItem(title string) string {
	return "* " + renderTocLink(title)
}
