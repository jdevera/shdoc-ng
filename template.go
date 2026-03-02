package main

import (
	_ "embed"
	"strings"
	"text/template"
)

//go:embed templates/markdown.tmpl
var defaultMarkdownTemplate string

// optionFormStr reconstructs the raw display string for one OptionForm.
// e.g. OptionForm{Name: "--file", Value: "path", ValueSep: " "} → "--file <path>"
func optionFormStr(f OptionForm) string {
	if f.Value == "" {
		return f.Name
	}
	return f.Name + f.ValueSep + "<" + f.Value + ">"
}

// mdEscape replaces < and > with their escaped markdown equivalents.
func mdEscape(s string) string {
	s = strings.ReplaceAll(s, "<", `\<`)
	s = strings.ReplaceAll(s, ">", `\>`)
	return s
}

// mdBold wraps a string in markdown bold markers.
func mdBold(s string) string {
	return "**" + s + "**"
}

// mdLink produces a markdown link.
func mdLink(text, href string) string {
	return "[" + text + "](" + href + ")"
}

// mdAnchor produces a markdown anchor link to a heading on the same page.
func mdAnchor(text string) string {
	return "[" + text + "](#" + slug(text) + ")"
}

// mdLinkify wraps bare URLs in markdown link syntax, leaving existing links alone.
func mdLinkify(s string) string {
	padded := "  " + s + " "
	padded = urlReplacePattern.ReplaceAllString(padded, "${1}[${3}](${3})")
	padded = strings.TrimLeft(padded, " ")
	padded = strings.TrimRight(padded, " ")
	return padded
}

// funcMap is the template function map used for rendering.
var funcMap = template.FuncMap{
	"slug":         slug,
	"unindent":     unindent,
	"optionFormStr": optionFormStr,
	"mdEscape":     mdEscape,
	"mdBold":       mdBold,
	"mdLink":       mdLink,
	"mdAnchor":     mdAnchor,
	"mdLinkify":    mdLinkify,
	"renderSeeRef": renderSeeRef,
	"trimSpace":    strings.TrimSpace,
	"replaceAll":   strings.ReplaceAll,
}

// renderWithTemplate renders a Document using the given template text.
func renderWithTemplate(doc *Document, tmplText string) (string, error) {
	tmpl, err := template.New("doc").Funcs(funcMap).Parse(tmplText)
	if err != nil {
		return "", err
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, doc); err != nil {
		return "", err
	}
	return buf.String(), nil
}
