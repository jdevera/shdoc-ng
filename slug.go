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

// urlPattern matches URLs not already inside markdown links.
// Matches: a non-] character followed by a non-lowercase-letter-or-paren (or a paren), then a URL scheme.
var urlDetectPattern = regexp.MustCompile(`[^\]]([^a-z(]|\()[a-z]+://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|]`)

// urlReplacePattern is used to wrap bare URLs in markdown link syntax.
var urlReplacePattern = regexp.MustCompile(`([^\]]([^a-z(]|\())([a-z]+://[-a-zA-Z0-9+&@#/%?=~_|!:,.;]*[-a-zA-Z0-9+&@#/%=~_|])`)

// markdownLinkPattern detects if text contains a markdown link anywhere.
var markdownLinkPattern = regexp.MustCompile(`\[[^\]]*\]\([^)]*\)`)

// markdownLinkParseRegex matches text that is entirely a single markdown link.
var markdownLinkParseRegex = regexp.MustCompile(`^\[([^\]]*)\]\(([^)]*)\)$`)

// bareURLRegex matches text that is entirely a URL.
var bareURLRegex = regexp.MustCompile(`^[a-z]+://\S`)

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
	padded := "  " + text + " "
	if urlDetectPattern.MatchString(padded) || markdownLinkPattern.MatchString(text) {
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
		// Mixed content: wrap bare URLs but leave existing markdown links alone
		padded := "  " + ref.Text + " "
		padded = urlReplacePattern.ReplaceAllString(padded, "${1}[${3}](${3})")
		padded = strings.TrimLeft(padded, " ")
		padded = strings.TrimRight(padded, " ")
		return padded
	}
	return ref.Text
}

