package shdoc

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

// pipeNormalizer normalizes whitespace around pipe separators between option forms.
var pipeNormalizer = regexp.MustCompile(`[ \t]+\|[ \t]+`)

// shortFormRegex parses a single short option token: -x or -x<val> or -x <val>
var shortFormRegex = regexp.MustCompile(`^(-[a-zA-Z0-9])([ \t]*)(?:<([^>]+)>)?$`)

// longFormRegex parses a single long option token: --xxx or --xxx=<val> or --xxx <val>
var longFormRegex = regexp.MustCompile(`^(--[a-zA-Z0-9][a-zA-Z0-9-]*)(=|[ \t]+)?(?:<([^>]+)>)?$`)

// parseOptionForms splits a normalized term string into individual OptionForms.
func parseOptionForms(term string) []OptionForm {
	var forms []OptionForm
	for _, tok := range strings.Split(term, " | ") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		var form OptionForm
		if m := longFormRegex.FindStringSubmatch(tok); m != nil {
			form.Name = m[1]
			if m[3] != "" {
				sep := m[2]
				if sep != "=" {
					sep = " "
				}
				form.Value = m[3]
				form.ValueSep = sep
			}
		} else if m := shortFormRegex.FindStringSubmatch(tok); m != nil {
			form.Name = m[1]
			if m[3] != "" {
				form.Value = m[3]
				if m[2] != "" {
					form.ValueSep = " "
				}
				// empty m[2] means adjacent (value_sep omitted)
			}
		}
		forms = append(forms, form)
	}
	return forms
}

// processAtOption validates and parses an @option entry.
// Returns (forms, definition, valid).
func processAtOption(text string) ([]OptionForm, string, bool) {
	text = normalizeOptionSeparators(text)
	m := optionRegex.FindStringSubmatch(text)
	if m == nil {
		return nil, "", false
	}

	term := strings.TrimSpace(m[1])
	definition := strings.TrimSpace(m[8])

	term = pipeNormalizer.ReplaceAllString(term, " | ")

	return parseOptionForms(term), definition, true
}

