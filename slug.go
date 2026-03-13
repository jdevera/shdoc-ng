package shdoc

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

// urlReplacePattern matches bare URLs not already inside markdown links
// and is used both for detection and for wrapping them in markdown link syntax.
// The regex requires leading context characters, so callers must pad the input;
// use hasBareURL for detection and linkifyBareURLs for replacement.
var urlReplacePattern = regexp.MustCompile(`([^\]]([^a-z(]|\())([a-z]+://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|])`)

// hasBareURL reports whether text contains a bare URL that is not already
// inside a markdown link.
func hasBareURL(text string) bool {
	return urlReplacePattern.MatchString("  " + text + " ")
}

// linkifyBareURLs wraps bare URLs in markdown link syntax, leaving existing
// markdown links alone.
func linkifyBareURLs(text string) string {
	padded := "  " + text + " "
	padded = urlReplacePattern.ReplaceAllString(padded, "${1}[${3}](${3})")
	padded = strings.TrimLeft(padded, " ")
	padded = strings.TrimRight(padded, " ")
	return padded
}

// markdownLinkPattern detects if text contains a markdown link anywhere.
var markdownLinkPattern = regexp.MustCompile(`\[[^\]]*\]\([^)]*\)`)

// markdownLinkParseRegex matches text that is entirely a single markdown link.
var markdownLinkParseRegex = regexp.MustCompile(`^\[([^\]]*)\]\(([^)]*)\)$`)

// bareURLRegex matches text that is entirely a URL (no trailing text).
var bareURLRegex = regexp.MustCompile(`^[a-z]+://\S+$`)

// parseSeeRef classifies a raw @see value into a structured SeeRef.
// Classification order matches the awk implementation.
func parseSeeRef(text string) SeeRef {
	// 1. Relative path
	if strings.HasPrefix(text, "./") || strings.HasPrefix(text, "../") {
		return SeeRef{Kind: "path", Href: text}
	}

	// 2. Absolute path
	if strings.HasPrefix(text, "/") {
		return SeeRef{Kind: "path", Href: text}
	}

	// 3. Pure markdown link [text](url) — entire text is the link
	if m := markdownLinkParseRegex.FindStringSubmatch(text); m != nil {
		return SeeRef{Kind: "link", Text: m[1], Href: m[2]}
	}

	// 4. Bare URL — entire text is a URL
	if bareURLRegex.MatchString(text) {
		return SeeRef{Kind: "url", Href: text}
	}

	// 5. Mixed content — text containing embedded URLs or markdown links
	if hasBareURL(text) || markdownLinkPattern.MatchString(text) {
		return SeeRef{Kind: "text", Text: text}
	}

	// 6. Plain anchor reference
	return SeeRef{Kind: "ref", Text: text}
}

// renderSeeRef renders a SeeRef to markdown.
func renderSeeRef(ref SeeRef) string {
	switch ref.Kind {
	case "path":
		return "[" + ref.Href + "](" + ref.Href + ")"
	case "url":
		return "[" + ref.Href + "](" + ref.Href + ")"
	case "link":
		return "[" + ref.Text + "](" + ref.Href + ")"
	case "ref":
		return "[" + ref.Text + "](#" + slug(ref.Text) + ")"
	case "text":
		return linkifyBareURLs(ref.Text)
	}
	return ref.Text
}

