package main

import (
	"regexp"
	"strings"
)

// optionSepNormalizer matches | or / separators (with optional surrounding
// whitespace) followed by a dash, and normalizes them to " | -".
// This allows users to write -s/--long, -s|--long, or -s / --long and have
// them all treated equivalently to -s | --long.
var optionSepNormalizer = regexp.MustCompile(`[ \t]*[|/][ \t]*(-)`)

// normalizeOptionSeparators replaces / or | separators between flag-like
// tokens with the canonical " | " form expected by the option regex.
func normalizeOptionSeparators(text string) string {
	return optionSepNormalizer.ReplaceAllString(text, " | $1")
}

// optionRegex matches valid @option formats.
// Port of the awk regex:
// ^(((-[[:alnum:]]([[:blank:]]*<[^>]+>)?|--[[:alnum:]][[:alnum:]-]*((=|[[:blank:]]+)<[^>]+>)?)([[:blank:]]*\|?[[:blank:]]+))+)([^[:blank:]|<-].*)?$
var optionRegex = regexp.MustCompile(
	`^(((-[a-zA-Z0-9]([ \t]*<[^>]+>)?|--[a-zA-Z0-9][a-zA-Z0-9-]*((=|[ \t]+)<[^>]+>)?)([ \t]*\|?[ \t]+))+)([^ \t|<-].*)?$`,
)

// processAtOption validates and parses an @option entry.
// Returns (term, definition, valid).
func processAtOption(text string) (string, string, bool) {
	text = normalizeOptionSeparators(text)
	m := optionRegex.FindStringSubmatch(text)
	if m == nil {
		return "", "", false
	}

	term := strings.TrimSpace(m[1])
	definition := strings.TrimSpace(m[8])

	// Normalize spaces around pipes
	pipeRe := regexp.MustCompile(`[ \t]+\|[ \t]+`)
	term = pipeRe.ReplaceAllString(term, " | ")

	return term, definition, true
}

// renderOptionTerm renders the term portion of an option for markdown output.
// Wraps in bold, splits around pipes, escapes < and >.
func renderOptionTerm(term string) string {
	// Wrap in bold
	result := "**" + term + "**"

	// Split bold around pipes: " | " within bold becomes "** | **"
	result = strings.ReplaceAll(result, " | ", "** | **")

	// Escape < and >
	result = strings.ReplaceAll(result, "<", `\<`)
	result = strings.ReplaceAll(result, ">", `\>`)

	return result
}
