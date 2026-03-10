package lsp

import (
	"regexp"
	"strings"
	"sync"

	shdoc "github.com/jdevera/shdoc-ng"

	"github.com/tliron/commonlog"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"github.com/tliron/glsp/server"
)

// positionalParamRe matches shell positional parameters: $1-$9, $@, $*, $#, ${N}
var positionalParamRe = regexp.MustCompile(`\$([0-9]+|[@*#]|\{[0-9]+\})`)

// inSingleQuotes reports whether the byte offset pos falls inside single quotes on the line.
func inSingleQuotes(line string, pos int) bool {
	inSQ := false
	for i := 0; i < len(line) && i < pos; i++ {
		if line[i] == '\'' {
			inSQ = !inSQ
		}
	}
	return inSQ
}

const serverName = "shdoc-lsp"

// hoverTemplate renders a single FuncDoc without TOC, section headers, or file metadata.
const hoverTemplate = `{{- range .Functions -}}
### {{.Name}}
{{- if .IsDeprecated}}
{{if eq .DeprecatedMessage ""}}
**DEPRECATED.**
{{- else}}
**DEPRECATED:** {{.DeprecatedMessage}}
{{- end}}{{end}}
{{- if .Description}}

{{.Description}}
{{- end}}
{{- if .Warnings}}

#### Warnings

{{range $i, $w := .Warnings}}{{if $i}}
{{end}}* {{$w}}{{end}}
{{- end}}
{{- if .Example}}

#### Example

` + "```bash" + `
{{unindent .Example}}
` + "```" + `
{{- end}}
{{- if or .Options .BadOptions}}

#### Options

{{range $i, $o := .Options}}{{if $i}}

{{end}}* {{range $j, $f := $o.Forms}}{{if $j}} | {{end}}{{mdBold (mdEscape (optionFormStr $f))}}{{end}}

  {{$o.Definition}}{{end}}{{if and .Options .BadOptions}}

{{end}}{{range $i, $b := .BadOptions}}{{if $i}}
{{end}}* {{$b}}{{end}}
{{- end}}
{{- if .Args}}

#### Arguments

{{range $i, $a := .Args}}{{if $i}}
{{end}}{{if eq $a.Name "$@"}}* **...** ({{$a.Type}}): {{$a.Description}}{{else}}* **{{$a.Name}}** ({{$a.Type}}): {{$a.Description}}{{end}}{{end}}
{{- end}}
{{- if .IsNoArgs}}

_Function has no arguments._
{{- end}}
{{- if .Sets}}

#### Variables set

{{range $i, $s := .Sets}}{{if $i}}
{{end}}* **{{$s.Name}}** ({{$s.Type}}): {{$s.Description}}{{end}}
{{- end}}
{{- if .Env}}

#### Environment variables

{{range $i, $e := .Env}}{{if $i}}
{{end}}* **{{$e.Name}}** ({{$e.Type}}): {{$e.Description}}{{end}}
{{- end}}
{{- if .ExitCodes}}

#### Exit codes

{{range $i, $e := .ExitCodes}}{{if $i}}
{{end}}* **{{$e.Code}}**: {{$e.Description}}{{end}}
{{- end}}
{{- if .Stdin}}

#### Input on stdin

{{range $i, $s := .Stdin}}{{if $i}}
{{end}}* {{replaceAll $s "\n" "\n  "}}{{end}}
{{- end}}
{{- if .Stdout}}

#### Output on stdout

{{range $i, $s := .Stdout}}{{if $i}}
{{end}}* {{replaceAll $s "\n" "\n  "}}{{end}}
{{- end}}
{{- if .Stderr}}

#### Output on stderr

{{range $i, $s := .Stderr}}{{if $i}}
{{end}}* {{replaceAll $s "\n" "\n  "}}{{end}}
{{- end}}
{{- if .See}}

#### See also

{{range $i, $s := .See}}{{if $i}}
{{end}}* {{renderSeeRef $s}}{{end}}
{{- end}}
{{end}}`

// fileHoverTemplate renders the file-level metadata for hover on meta block tags.
const fileHoverTemplate = `{{- if .FileTitle}}# {{.FileTitle}}{{end}}
{{- if .FileBrief}}

{{.FileBrief}}
{{- end}}
{{- if .FileDescription}}

{{.FileDescription}}
{{- end}}
{{- if .Authors}}

**Authors:** {{range $i, $a := .Authors}}{{if $i}}, {{end}}{{$a}}{{end}}
{{- end}}
{{- if .License}}

**License:** {{.License}}
{{- end}}
{{- if .Version}}

**Version:** {{.Version}}
{{- end}}
`

var handler protocol.Handler

// docState holds the parsed state for a single open document.
type docState struct {
	source string
	lines  []shdoc.LexedLine
	blocks []shdoc.ParsedBlock
	doc    shdoc.Document
	warns  []shdoc.Warning
}

var (
	mu    sync.RWMutex
	store = map[string]*docState{}
)

// Run starts the LSP server on stdio.
func Run() {
	commonlog.Configure(1, nil)
	handler = protocol.Handler{
		Initialize:                 initialize,
		Initialized:               func(*glsp.Context, *protocol.InitializedParams) error { return nil },
		Shutdown:                   func(*glsp.Context) error { return nil },
		SetTrace:                   func(_ *glsp.Context, p *protocol.SetTraceParams) error { return nil },
		TextDocumentDidOpen:        didOpen,
		TextDocumentDidChange:      didChange,
		TextDocumentDidClose:       didClose,
		TextDocumentHover:          hover,
		TextDocumentCompletion:     completion,
		TextDocumentDocumentSymbol: documentSymbol,
		TextDocumentDefinition:     definition,
		TextDocumentFoldingRange:   foldingRange,
		TextDocumentCodeAction:     codeAction,
	}
	_ = server.NewServer(&handler, serverName, false).RunStdio()
}

func initialize(_ *glsp.Context, _ *protocol.InitializeParams) (any, error) {
	caps := handler.CreateServerCapabilities()
	caps.TextDocumentSync = protocol.TextDocumentSyncKindFull
	caps.HoverProvider = true
	caps.CompletionProvider = &protocol.CompletionOptions{
		TriggerCharacters: []string{"@"},
	}
	caps.DocumentSymbolProvider = true
	caps.DefinitionProvider = true
	caps.FoldingRangeProvider = true
	caps.CodeActionProvider = true
	return protocol.InitializeResult{
		Capabilities: caps,
		ServerInfo:   &protocol.InitializeResultServerInfo{Name: serverName},
	}, nil
}

// parse re-parses a document and updates the store, then publishes diagnostics.
func parse(ctx *glsp.Context, uri, src string) {
	lines := shdoc.LexLines(src)
	blocks := shdoc.SegmentBlocks(lines)
	doc, warns := shdoc.ParseDocument(src)

	st := &docState{source: src, lines: lines, blocks: blocks, doc: doc, warns: warns}
	mu.Lock()
	store[uri] = st
	mu.Unlock()

	publishDiagnostics(ctx, uri, st)
}

func publishDiagnostics(ctx *glsp.Context, uri string, state *docState) {
	var diags []protocol.Diagnostic
	src := serverName

	for _, w := range state.warns {
		line := uint32(w.Line - 1)
		col := uint32(w.Col)
		sev := protocol.DiagnosticSeverityWarning
		diags = append(diags, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{Line: line, Character: col},
				End:   protocol.Position{Line: line, Character: col + 1},
			},
			Severity: &sev,
			Message:  w.Message,
			Source:   &src,
		})
	}

	// Add deprecated strikethrough on function declaration lines.
	for _, b := range state.blocks {
		if b.Kind != shdoc.FuncDocBlockKind {
			continue
		}
		for _, f := range state.doc.Functions {
			if f.Name == b.FuncName && f.IsDeprecated {
				funcLine := uint32(b.Comments.EndNum) // EndNum is 1-based, func decl is next line, so 0-based = EndNum
				sev := protocol.DiagnosticSeverityHint
				tag := protocol.DiagnosticTagDeprecated
				msg := "deprecated"
				if f.DeprecatedMessage != "" {
					msg = "deprecated: " + f.DeprecatedMessage
				}
				diags = append(diags, protocol.Diagnostic{
					Range: protocol.Range{
						Start: protocol.Position{Line: funcLine, Character: 0},
						End:   protocol.Position{Line: funcLine, Character: 1000},
					},
					Severity: &sev,
					Message:  msg,
					Source:   &src,
					Tags:     []protocol.DiagnosticTag{tag},
				})
				break
			}
		}
	}

	// Warn when @noargs function uses positional parameters in its body.
	for _, b := range state.blocks {
		if b.Kind != shdoc.FuncDocBlockKind {
			continue
		}
		for _, f := range state.doc.Functions {
			if f.Name == b.FuncName && f.IsNoArgs {
				// Find function body: starts at decl line (EndNum is 1-based, so 0-based index = EndNum)
				funcStart := b.Comments.EndNum // 0-based index of func decl line
				// Scan forward tracking brace depth to find function end.
				depth := 0
				for li := funcStart; li < len(state.lines); li++ {
					raw := state.lines[li].Raw
					for _, ch := range raw {
						switch ch {
						case '{':
							depth++
						case '}':
							depth--
						}
					}
					// Skip the declaration line itself for param scanning.
					if li > funcStart {
						locs := positionalParamRe.FindAllStringIndex(raw, -1)
						for _, loc := range locs {
							if inSingleQuotes(raw, loc[0]) {
								continue
							}
							lineNum := uint32(li)
							col := uint32(loc[0])
							sev := protocol.DiagnosticSeverityWarning
							param := raw[loc[0]:loc[1]]
							diags = append(diags, protocol.Diagnostic{
								Range: protocol.Range{
									Start: protocol.Position{Line: lineNum, Character: col},
									End:   protocol.Position{Line: lineNum, Character: uint32(loc[1])},
								},
								Severity: &sev,
								Message:  "function " + f.Name + " is marked @noargs but uses " + param,
								Source:   &src,
							})
						}
					}
					if depth == 0 && li > funcStart {
						break
					}
				}
				break
			}
		}
	}

	ctx.Notify(protocol.ServerTextDocumentPublishDiagnostics,
		protocol.PublishDiagnosticsParams{URI: uri, Diagnostics: diags})
}

func didOpen(ctx *glsp.Context, p *protocol.DidOpenTextDocumentParams) error {
	parse(ctx, string(p.TextDocument.URI), p.TextDocument.Text)
	return nil
}

func didChange(ctx *glsp.Context, p *protocol.DidChangeTextDocumentParams) error {
	if len(p.ContentChanges) == 0 {
		return nil
	}
	// With TextDocumentSyncKindFull the first change has the entire file.
	// glsp may deserialize as the concrete type or as a map; handle both.
	var text string
	switch c := p.ContentChanges[0].(type) {
	case protocol.TextDocumentContentChangeEventWhole:
		text = c.Text
	case map[string]any:
		if t, ok := c["text"].(string); ok {
			text = t
		}
	}
	if text != "" {
		parse(ctx, string(p.TextDocument.URI), text)
	}
	return nil
}

func didClose(_ *glsp.Context, p *protocol.DidCloseTextDocumentParams) error {
	mu.Lock()
	delete(store, string(p.TextDocument.URI))
	mu.Unlock()
	return nil
}

// documentSymbol returns a hierarchy: file name > sections > functions.
func documentSymbol(_ *glsp.Context, p *protocol.DocumentSymbolParams) (any, error) {
	mu.RLock()
	state := store[string(p.TextDocument.URI)]
	mu.RUnlock()
	if state == nil {
		return nil, nil
	}

	// Build a map from function name to its block for line info.
	blockByFunc := map[string]*shdoc.ParsedBlock{}
	for i := range state.blocks {
		b := &state.blocks[i]
		if b.Kind == shdoc.FuncDocBlockKind {
			blockByFunc[b.FuncName] = b
		}
	}

	// Build function symbols grouped by section.
	type sectionInfo struct {
		name     string
		children []protocol.DocumentSymbol
		// Track range: first and last line across all children.
		startLine uint32
		endLine   uint32
	}
	sections := map[string]*sectionInfo{}
	var sectionOrder []string
	var ungrouped []protocol.DocumentSymbol

	funcKind := protocol.SymbolKindString

	for _, f := range state.doc.Functions {
		b := blockByFunc[f.Name]
		if b == nil {
			continue
		}
		start := uint32(b.Comments.StartNum - 1)
		end := uint32(b.Comments.EndNum - 1)
		sym := protocol.DocumentSymbol{
			Name: f.Name,
			Kind: funcKind,
			Range: protocol.Range{
				Start: protocol.Position{Line: start},
				End:   protocol.Position{Line: end},
			},
			SelectionRange: protocol.Range{
				Start: protocol.Position{Line: start},
				End:   protocol.Position{Line: start},
			},
		}

		if f.Section == "" {
			ungrouped = append(ungrouped, sym)
		} else {
			sec, ok := sections[f.Section]
			if !ok {
				sec = &sectionInfo{name: f.Section, startLine: start, endLine: end}
				sections[f.Section] = sec
				sectionOrder = append(sectionOrder, f.Section)
			}
			sec.children = append(sec.children, sym)
			if start < sec.startLine {
				sec.startLine = start
			}
			if end > sec.endLine {
				sec.endLine = end
			}
		}
	}

	// Assemble top-level symbols.
	var topLevel []protocol.DocumentSymbol

	// File-level meta symbol.
	if state.doc.FileTitle != "" {
		// Find the meta block for range info.
		var metaStart, metaEnd uint32
		for _, b := range state.blocks {
			if b.Kind == shdoc.MetaBlockKind {
				metaStart = uint32(b.Comments.StartNum - 1)
				metaEnd = uint32(b.Comments.EndNum - 1)
				break
			}
		}
		metaKind := protocol.SymbolKindModule
		detail := state.doc.FileBrief
		topLevel = append(topLevel, protocol.DocumentSymbol{
			Name:   state.doc.FileTitle,
			Detail: &detail,
			Kind:   metaKind,
			Range: protocol.Range{
				Start: protocol.Position{Line: metaStart},
				End:   protocol.Position{Line: metaEnd},
			},
			SelectionRange: protocol.Range{
				Start: protocol.Position{Line: metaStart},
				End:   protocol.Position{Line: metaStart},
			},
		})
	}

	// Section symbols with function children.
	sectionKind := protocol.SymbolKindNamespace
	for _, name := range sectionOrder {
		sec := sections[name]
		topLevel = append(topLevel, protocol.DocumentSymbol{
			Name:     sec.name,
			Kind:     sectionKind,
			Children: sec.children,
			Range: protocol.Range{
				Start: protocol.Position{Line: sec.startLine},
				End:   protocol.Position{Line: sec.endLine},
			},
			SelectionRange: protocol.Range{
				Start: protocol.Position{Line: sec.startLine},
				End:   protocol.Position{Line: sec.startLine},
			},
		})
	}

	// Ungrouped functions at top level.
	topLevel = append(topLevel, ungrouped...)

	return topLevel, nil
}

// foldingRange returns one folding range per comment block.
func foldingRange(_ *glsp.Context, p *protocol.FoldingRangeParams) ([]protocol.FoldingRange, error) {
	mu.RLock()
	state := store[string(p.TextDocument.URI)]
	mu.RUnlock()
	if state == nil {
		return nil, nil
	}

	var ranges []protocol.FoldingRange
	kind := string(protocol.FoldingRangeKindComment)
	for _, b := range state.blocks {
		start := uint32(b.Comments.StartNum - 1)
		end := uint32(b.Comments.EndNum - 1)
		if end <= start {
			continue
		}
		ranges = append(ranges, protocol.FoldingRange{
			StartLine: start,
			EndLine:   end,
			Kind:      &kind,
		})
	}
	return ranges, nil
}

// cursorOnTag checks whether the cursor (0-based column) is positioned on the
// @tag keyword in a raw comment line. Returns the parsed tag name or "" if the
// cursor is not on the tag.
func cursorOnTag(raw string, cursorCol int) string {
	atIdx := strings.Index(raw, "@")
	if atIdx < 0 || cursorCol < atIdx {
		return ""
	}
	stripped := shdoc.StripCommentPrefix(raw)
	tag, _ := shdoc.ParseTag(stripped)
	if tag == "" {
		return ""
	}
	// The tag spans from @ to end of tag name in the raw line.
	// Find end: atIdx + 1 + len(original tag text before normalization).
	// We need the original tag text, so re-match from the raw line.
	after := raw[atIdx+1:]
	tagEnd := atIdx + 1
	for i, ch := range after {
		if ch == ' ' || ch == '\t' {
			tagEnd = atIdx + 1 + i
			break
		}
		tagEnd = atIdx + 1 + i + 1
	}
	if cursorCol >= atIdx && cursorCol < tagEnd {
		return tag
	}
	return ""
}

// hover shows a rendered preview of the function doc block under the cursor.
func hover(_ *glsp.Context, p *protocol.HoverParams) (*protocol.Hover, error) {
	mu.RLock()
	state := store[string(p.TextDocument.URI)]
	mu.RUnlock()
	if state == nil {
		return nil, nil
	}

	cursorLine := int(p.Position.Line) + 1
	lineIdx := int(p.Position.Line)
	cursorCol := int(p.Position.Character)

	var lineRaw string
	if lineIdx < len(state.lines) {
		lineRaw = state.lines[lineIdx].Raw
	}

	// Check if cursor is on a @tag keyword.
	hoveredTag := cursorOnTag(lineRaw, cursorCol)

	// If cursor is on @see, show the referenced function's docs.
	if hoveredTag == "see" {
		stripped := shdoc.StripCommentPrefix(lineRaw)
		_, value := shdoc.ParseTag(stripped)
		target := strings.TrimRight(strings.TrimSuffix(value, "()"), " ")
		for i := range state.doc.Functions {
			if state.doc.Functions[i].Name == target {
				fd := state.doc.Functions[i]
				md, err := shdoc.RenderWithTemplate(
					&shdoc.Document{Functions: []shdoc.FuncDoc{fd}},
					hoverTemplate,
				)
				if err != nil {
					return nil, err
				}
				return &protocol.Hover{
					Contents: protocol.MarkupContent{Kind: protocol.MarkupKindMarkdown, Value: md},
				}, nil
			}
		}
	}

	// Hover on @tag in meta blocks shows file-level metadata.
	if hoveredTag != "" {
		for _, b := range state.blocks {
			if b.Kind == shdoc.MetaBlockKind &&
				cursorLine >= b.Comments.StartNum && cursorLine <= b.Comments.EndNum {
				md, err := shdoc.RenderWithTemplate(&state.doc, fileHoverTemplate)
				if err != nil {
					return nil, err
				}
				if strings.TrimSpace(md) != "" {
					return &protocol.Hover{
						Contents: protocol.MarkupContent{Kind: protocol.MarkupKindMarkdown, Value: md},
					}, nil
				}
			}
		}
	}

	for _, b := range state.blocks {
		if b.Kind != shdoc.FuncDocBlockKind {
			continue
		}

		// Hover on the function declaration line or on @tag keyword
		// within the comment block.
		onDeclLine := cursorLine == b.Comments.EndNum+1
		onTagLine := hoveredTag != "" &&
			cursorLine >= b.Comments.StartNum && cursorLine <= b.Comments.EndNum
		if !onDeclLine && !onTagLine {
			continue
		}

		var fd *shdoc.FuncDoc
		for i := range state.doc.Functions {
			if state.doc.Functions[i].Name == b.FuncName {
				fd = &state.doc.Functions[i]
				break
			}
		}
		if fd == nil {
			return nil, nil
		}

		md, err := shdoc.RenderWithTemplate(
			&shdoc.Document{Functions: []shdoc.FuncDoc{*fd}},
			hoverTemplate,
		)
		if err != nil {
			return nil, err
		}
		return &protocol.Hover{
			Contents: protocol.MarkupContent{Kind: protocol.MarkupKindMarkdown, Value: md},
		}, nil
	}
	return nil, nil
}

// Tag lists for completion.
var metaTags = []string{"name", "file", "brief", "description", "desc", "author",
	"license", "version", "section"}
var funcTags = []string{"description", "desc", "internal", "deprecated", "warning", "warn",
	"example", "option", "opt", "arg", "noargs", "set", "env", "exitcode", "exit",
	"stdin", "stdout", "stderr", "see"}

// completion offers @tag completions inside comment blocks.
func completion(_ *glsp.Context, p *protocol.CompletionParams) (any, error) {
	mu.RLock()
	state := store[string(p.TextDocument.URI)]
	mu.RUnlock()
	if state == nil {
		return nil, nil
	}

	lineIdx := int(p.Position.Line)
	if lineIdx >= len(state.lines) {
		return nil, nil
	}

	lineText := state.lines[lineIdx].Raw
	stripped := shdoc.StripCommentPrefix(lineText)

	// Only offer completions on comment lines.
	if lineIdx < len(state.lines) && state.lines[lineIdx].Kind != shdoc.LineComment {
		return nil, nil
	}

	// Offer completions when cursor is after @ or on an empty comment line.
	// Always offer all tags — the parser handles context gracefully, and
	// meta tags like @author can legitimately appear in func blocks.
	if strings.HasPrefix(stripped, "@") || stripped == "" {
		var tags []string
		seen := map[string]bool{}
		for _, t := range metaTags {
			if !seen[t] {
				tags = append(tags, t)
				seen[t] = true
			}
		}
		for _, t := range funcTags {
			if !seen[t] {
				tags = append(tags, t)
				seen[t] = true
			}
		}

		// Find the column of '@' in the raw line to build the replacement range.
		atCol := strings.Index(lineText, "@")
		if atCol < 0 {
			// No @ yet (empty comment line) — insert at cursor.
			atCol = int(p.Position.Character)
		}

		kind := protocol.CompletionItemKindKeyword
		items := make([]protocol.CompletionItem, len(tags))
		for i, t := range tags {
			label := "@" + t
			items[i] = protocol.CompletionItem{
				Label: label,
				Kind:  &kind,
				TextEdit: protocol.TextEdit{
					Range: protocol.Range{
						Start: protocol.Position{Line: p.Position.Line, Character: uint32(atCol)},
						End:   p.Position,
					},
					NewText: label,
				},
			}
		}
		return items, nil
	}
	return nil, nil
}

// definition resolves @see funcName() to the target function's comment block.
func definition(_ *glsp.Context, p *protocol.DefinitionParams) (any, error) {
	mu.RLock()
	state := store[string(p.TextDocument.URI)]
	mu.RUnlock()
	if state == nil {
		return nil, nil
	}

	lineIdx := int(p.Position.Line)
	if lineIdx >= len(state.lines) {
		return nil, nil
	}

	lineText := state.lines[lineIdx].Raw
	stripped := shdoc.StripCommentPrefix(lineText)
	tag, value := shdoc.ParseTag(stripped)
	if tag != "see" {
		return nil, nil
	}

	target := strings.TrimRight(strings.TrimSuffix(value, "()"), " ")
	for _, b := range state.blocks {
		if b.Kind == shdoc.FuncDocBlockKind && b.FuncName == target {
			line := uint32(b.Comments.StartNum - 1)
			return protocol.Location{
				URI: p.TextDocument.URI,
				Range: protocol.Range{
					Start: protocol.Position{Line: line},
					End:   protocol.Position{Line: line},
				},
			}, nil
		}
	}
	return nil, nil
}

// docBlockSkeleton returns a doc block comment to insert above a function.
func docBlockSkeleton(funcName string) string {
	return "# @description TODO: document " + funcName + ".\n" +
		"#\n" +
		"# @arg $1 string Description.\n" +
		"#\n" +
		"# @exitcode 0 Success.\n"
}

// codeAction offers to insert a doc block skeleton above undocumented functions.
func codeAction(_ *glsp.Context, p *protocol.CodeActionParams) (any, error) {
	mu.RLock()
	state := store[string(p.TextDocument.URI)]
	mu.RUnlock()
	if state == nil {
		return nil, nil
	}

	var actions []protocol.CodeAction
	startLine := int(p.Range.Start.Line)
	endLine := int(p.Range.End.Line)

	for lineIdx := startLine; lineIdx <= endLine && lineIdx < len(state.lines); lineIdx++ {
		line := state.lines[lineIdx]
		if line.Kind != shdoc.LineCode {
			continue
		}
		if !shdoc.IsFuncDecl(line.Raw) {
			continue
		}
		funcName := shdoc.ExtractFuncName(line.Raw)
		if funcName == "" {
			continue
		}

		documented := false
		for _, b := range state.blocks {
			if b.Kind == shdoc.FuncDocBlockKind && b.FuncName == funcName {
				documented = true
				break
			}
		}
		if documented {
			continue
		}

		skeleton := docBlockSkeleton(funcName)
		insertLine := uint32(lineIdx)
		title := "Insert shdoc-ng doc block for " + funcName
		kind := protocol.CodeActionKindQuickFix
		edit := protocol.WorkspaceEdit{
			Changes: map[protocol.DocumentUri][]protocol.TextEdit{
				p.TextDocument.URI: {
					{
						Range: protocol.Range{
							Start: protocol.Position{Line: insertLine, Character: 0},
							End:   protocol.Position{Line: insertLine, Character: 0},
						},
						NewText: skeleton,
					},
				},
			},
		}
		actions = append(actions, protocol.CodeAction{
			Title: title,
			Kind:  &kind,
			Edit:  &edit,
		})
	}

	return actions, nil
}
