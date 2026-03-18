package shdoc

import (
	_ "embed"
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/alecthomas/chroma/v2"
	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

//go:embed templates/markdown.tmpl
var DefaultMarkdownTemplate string

//go:embed templates/html.tmpl
var DefaultHTMLTemplate string

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
	return linkifyBareURLs(s)
}

// md2html converts a Markdown string to an HTML string.
var mdRenderer = goldmark.New(
	goldmark.WithExtensions(extension.Linkify),
)

func md2html(s string) string {
	var buf bytes.Buffer
	if err := mdRenderer.Convert([]byte(s), &buf); err != nil {
		return s
	}
	return buf.String()
}

// md2inline converts a Markdown string to HTML and strips the wrapping <p>
// tags, suitable for use inside table cells or other inline contexts.
func md2inline(s string) string {
	result := strings.TrimSpace(md2html(s))
	result = strings.TrimPrefix(result, "<p>")
	result = strings.TrimSuffix(result, "</p>")
	return strings.TrimSpace(result)
}

// highlightCode syntax-highlights code using chroma. The language parameter
// selects the lexer (e.g. "bash", "json"). Falls back to plain text if the
// language is unknown. Returns HTML using CSS classes (chroma token classes
// like .k, .s, .nv) so that theme-specific color rules can be applied.
func highlightCode(lang, code string) string {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	formatter := chromahtml.New(chromahtml.WithClasses(true), chromahtml.PreventSurroundingPre(true))

	iterator, err := lexer.Tokenise(nil, code)
	if err != nil {
		return template.HTMLEscapeString(code)
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, styles.Fallback, iterator); err != nil {
		return template.HTMLEscapeString(code)
	}
	return buf.String()
}

// chromaThemeCSS generates CSS rules for chroma token classes scoped under
// [data-theme="name"] .code-block code selectors. Iterates chroma's token
// types directly to build properly scoped CSS without string hacks.
// Background colors are omitted — the code block background is managed
// by --theme-bg-code in the main CSS.
func chromaThemeCSS() string {
	themes := []struct {
		dataTheme  string
		chromaName string
	}{
		{"mocha", "catppuccin-mocha"},
		{"macchiato", "catppuccin-macchiato"},
		{"frappe", "catppuccin-frappe"},
		{"latte", "catppuccin-latte"},
	}

	var buf bytes.Buffer
	for _, t := range themes {
		style := styles.Get(t.chromaName)
		if style == nil {
			continue
		}
		prefix := fmt.Sprintf("[data-theme=%q] .code-block code", t.dataTheme)

		for _, ttype := range style.Types() {
			entry := style.Get(ttype)
			if entry.IsZero() {
				continue
			}
			cls := chroma.StandardTypes[ttype]
			if cls == "" {
				continue
			}

			var props []string
			if entry.Colour.IsSet() {
				props = append(props, fmt.Sprintf("color: %s", entry.Colour.String()))
			}
			if entry.Bold == chroma.Yes {
				props = append(props, "font-weight: bold")
			}
			if entry.Italic == chroma.Yes {
				props = append(props, "font-style: italic")
			}
			if entry.Underline == chroma.Yes {
				props = append(props, "text-decoration: underline")
			}
			if len(props) > 0 {
				fmt.Fprintf(&buf, "%s .%s { %s }\n", prefix, cls, strings.Join(props, "; "))
			}
		}
		buf.WriteByte('\n')
	}

	return buf.String()
}

// funcMap is the template function map used for rendering.
var funcMap = template.FuncMap{
	"slug":          slug,
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
	"md2html":       md2html,
	"md2inline":     md2inline,
	"highlightCode":  highlightCode,
	"chromaThemeCSS": chromaThemeCSS,
}

// renderWithTemplate renders a Document using the given template text.
func RenderWithTemplate(doc *Document, tmplText string) (string, error) {
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
