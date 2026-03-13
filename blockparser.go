package shdoc

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Warning is a parse warning collected during ParseDocument.
type Warning struct {
	Line    int
	Col     int // 0-based character offset of the offending token
	Message string
}

// Regexes compiled once for the block parser.
var (
	// Strips leading whitespace + # + all following whitespace.
	// Used for tag detection and description-line stripping.
	bpStripRe = regexp.MustCompile(`^\s*#\s*`)

	// Strips leading whitespace + # only (no following whitespace).
	// Used for example-line stripping to preserve indentation.
	bpStripHashRe = regexp.MustCompile(`^\s*#`)

	// Strips leading whitespace + # + one-or-more whitespace.
	// Used for stripping multi-line stdin/stdout/stderr continuations.
	bpStripContRe = regexp.MustCompile(`^\s*#\s+`)

	// Matches a tag line: @tagname optionally followed by whitespace + value.
	bpTagRe = regexp.MustCompile(`^@(\w+)(?:\s+(.*))?$`)

	// Detects start of a valid @arg value ($N or $@).
	bpArgStartRe = regexp.MustCompile(`^\$([0-9]+|@)\s`)

	// Parses a numbered @arg: $N type description
	bpArgNRe = regexp.MustCompile(`^\$([0-9]+)\s+(\S+)\s+(.*)$`)

	// Parses a rest @arg: $@ type description
	bpArgAtRe = regexp.MustCompile(`^\$@\s+(\S+)\s+(.*)$`)

	// Parses @set / @env: varname [type [description]]
	bpSetVarRe = regexp.MustCompile(`^(\S+)(?:\s+(\S+)(?:\s+(.*))?)?$`)

	// Parses @exitcode: code description
	bpExitCodeRe = regexp.MustCompile(`^([>!]?[0-9]{1,3}) (.*)$`)

	// Extracts the indentation prefix from a @stdin/@stdout/@stderr tag line.
	// The prefix is everything up to and including the # plus trailing whitespace.
	bpStdioIndentRe = regexp.MustCompile(`^([\t ]*#[\t ]+)`)

	// Matches a valid @example continuation line: any comment with >= 1 space after #.
	bpExampleContRe = regexp.MustCompile(`^[\s]*#[ ]+`)

	// Trims leading/trailing empty lines from a description.
	bpCleanLeadingRe  = regexp.MustCompile(`^[\s\n]*\n`)
	bpCleanTrailingRe = regexp.MustCompile(`[\s\n]*$`)
)

// ParseDocument parses src and returns the document and any warnings.
func ParseDocument(src string) (Document, []Warning) {
	lines := LexLines(src)
	blocks := SegmentBlocks(lines)
	return ParseBlocks(blocks)
}

// ParseBlocks parses pre-segmented blocks and returns the document and any warnings.
func ParseBlocks(blocks []ParsedBlock) (Document, []Warning) {
	bp := &blockParser{}
	for _, block := range blocks {
		if block.Kind == MetaBlockKind {
			bp.parseMetaBlock(block)
		} else {
			bp.parseFuncBlock(block)
		}
	}

	// Mark the first function in each section.
	seenSections := map[string]bool{}
	for i := range bp.doc.Functions {
		f := &bp.doc.Functions[i]
		if f.Section != "" && !seenSections[f.Section] {
			f.IsFirstInSection = true
			seenSections[f.Section] = true
		}
	}

	return bp.doc, bp.warns
}

type blockParser struct {
	doc     Document
	warns   []Warning
	section pendingSection
	// Track which file-level singleton tags have been set, to warn on duplicates.
	seenFileTags map[string]int // tag -> first line number
	// Cache compiled continuation regexes for @stdin/@stdout/@stderr across blocks.
	contReCache map[string]*regexp.Regexp
}

// fileLevelTags are tags that only make sense at the file level (meta blocks).
// They should be warned about and ignored when they appear in function blocks.
var fileLevelTags = map[string]bool{
	"name": true, "file": true, "brief": true,
	"author": true, "license": true, "version": true,
}

// fileSingletonTags are file-level tags that should only appear once.
// @author is excluded because multiple authors are valid (it appends).
var fileSingletonTags = map[string]bool{
	"name": true, "file": true, "brief": true,
	"license": true, "version": true,
}

type pendingSection struct {
	name string
	desc string
}

func (bp *blockParser) warn(lineNum int, col int, msg string) {
	bp.warns = append(bp.warns, Warning{Line: lineNum, Col: col, Message: msg})
}

// tagCol returns the 0-based column offset of the '@' in a raw comment line,
// or 0 if not found.
func tagCol(raw string) int {
	return strings.Index(raw, "@")
}

// isTagLine returns true if a raw comment line contains a @tag.
// Scans manually instead of using regex to avoid allocating a new string
// on every call — this is a hot path during block collection.
func isTagLine(raw string) bool {
	i := 0
	for i < len(raw) && (raw[i] == ' ' || raw[i] == '\t') {
		i++
	}
	if i >= len(raw) || raw[i] != '#' {
		return false
	}
	i++
	for i < len(raw) && (raw[i] == ' ' || raw[i] == '\t') {
		i++
	}
	return i < len(raw) && raw[i] == '@'
}

// collectUntilNextTag collects non-tag lines from lines[start:], stopping
// before the first tag line or the end of the slice.
// Returns the collected lines and the index of the first tag line (or len(lines)).
func collectUntilNextTag(lines []LexedLine, start int) ([]LexedLine, int) {
	i := start
	for i < len(lines) && !isTagLine(lines[i].Raw) {
		i++
	}
	return lines[start:i], i
}

// collectExampleLines collects @example continuation lines. A line continues
// the example if it has at least one space after the # character, matching the
// original awk/Go behaviour: /^[[:space:]]*#[ ]+/.
func collectExampleLines(lines []LexedLine, start int) ([]LexedLine, int) {
	i := start
	for i < len(lines) && bpExampleContRe.MatchString(lines[i].Raw) {
		// Stop if the continuation line contains a @tag.
		stripped := bpStripRe.ReplaceAllString(lines[i].Raw, "")
		if tag, _ := ParseTag(stripped); tag != "" {
			break
		}
		i++
	}
	return lines[start:i], i
}

// StripCommentPrefix strips the leading `# ` prefix from a comment line,
// removing all whitespace after the `#`.
func StripCommentPrefix(raw string) string {
	return bpStripRe.ReplaceAllString(raw, "")
}

// stripDesc is an internal alias for StripCommentPrefix.
func stripDesc(raw string) string {
	return StripCommentPrefix(raw)
}

// stripExample strips just the leading `#` from an example line,
// preserving any whitespace that follows (for the unindent template function).
func stripExample(raw string) string {
	return bpStripHashRe.ReplaceAllString(raw, "")
}

// tagShorthands maps shorthand tag names to their canonical form.
var tagShorthands = map[string]string{
	"desc": "description",
	"exit": "exitcode",
	"opt":  "option",
	"warn": "warning",
}

// parseTag extracts a tag name and value from a stripped line.
// Returns ("", "") if the line does not start with @.
// Shorthand tag names are normalized to their full form.
func ParseTag(stripped string) (name, value string) {
	m := bpTagRe.FindStringSubmatch(stripped)
	if m == nil {
		return "", ""
	}
	tag := m[1]
	if full, ok := tagShorthands[tag]; ok {
		tag = full
	}
	return tag, strings.TrimSpace(m[2])
}

// cleanDescription trims leading and trailing empty lines.
func cleanDescription(s string) string {
	s = bpCleanLeadingRe.ReplaceAllString(s, "")
	s = bpCleanTrailingRe.ReplaceAllString(s, "")
	return s
}

// routeDescription routes a description from a meta block to the appropriate
// document field: section description (if a section is pending) or file
// description.
func (bp *blockParser) routeDescription(desc string) {
	if desc == "" {
		return
	}
	if bp.section.name != "" && bp.section.desc == "" {
		bp.section.desc = desc
		return
	}
	if bp.doc.FileDescription == "" {
		bp.doc.FileDescription = desc
	}
}

// parseMetaBlock processes a MetaBlock (comment block not followed by a
// function declaration).
func (bp *blockParser) parseMetaBlock(block ParsedBlock) {
	lines := block.Comments.Lines
	for i := 0; i < len(lines); {
		raw := lines[i].Raw
		stripped := stripDesc(raw)
		tag, value := ParseTag(stripped)
		if tag == "" {
			i++
			continue
		}
		lineNum := lines[i].Num
		// Warn on duplicate singleton tags.
		if fileSingletonTags[tag] {
			canonicalTag := tag
			if tag == "file" {
				canonicalTag = "name"
			}
			if prev, ok := bp.seenFileTags[canonicalTag]; ok {
				bp.warn(lineNum, tagCol(raw), fmt.Sprintf("Duplicate @%s (first seen at line %d), overwriting previous value", tag, prev))
			}
			if bp.seenFileTags == nil {
				bp.seenFileTags = make(map[string]int)
			}
			bp.seenFileTags[canonicalTag] = lineNum
		}

		switch tag {
		case "name", "file":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @"+tag+" requires a name")
			}
			bp.doc.FileTitle = value
			i++
		case "brief":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @brief requires a description")
			}
			bp.doc.FileBrief = value
			i++
		case "author":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @author requires a name")
			}
			bp.doc.Authors = append(bp.doc.Authors, value)
			i++
		case "license":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @license requires a value")
			}
			bp.doc.License = value
			i++
		case "version":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @version requires a value")
			}
			bp.doc.Version = value
			i++
		case "section":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @section requires a name")
			}
			bp.section = pendingSection{name: value}
			i++
		case "description":
			var parts []string
			if value != "" {
				parts = append(parts, value)
			}
			cont, next := collectUntilNextTag(lines, i+1)
			for _, l := range cont {
				parts = append(parts, stripDesc(l.Raw))
			}
			i = next
			desc := cleanDescription(strings.Join(parts, "\n"))
			if desc == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @description requires content")
			}
			bp.routeDescription(desc)
		default:
			i++
		}
	}
}

// parseFuncBlock processes a FuncDocBlock (comment block immediately preceding
// a function declaration).
func (bp *blockParser) parseFuncBlock(block ParsedBlock) {
	lines := block.Comments.Lines
	var docblock FuncDoc
	tempArgs := make(map[string]Arg)
	isInternal := false

	// pendingDesc holds the most-recently-seen @description content. When a
	// second @description is encountered, the pending one is routed via the
	// normal meta routing (file description / section description). The last
	// one becomes the function description.
	var pendingDesc string

	for i := 0; i < len(lines); {
		raw := lines[i].Raw
		lineNum := lines[i].Num
		stripped := stripDesc(raw)
		tag, value := ParseTag(stripped)
		if tag == "" {
			i++
			continue
		}

		// File-level tags are not valid in function blocks — warn and skip.
		if fileLevelTags[tag] {
			bp.warn(lineNum, tagCol(raw), fmt.Sprintf("@%s is a file-level tag and will be ignored inside a function block", tag))
			i++
			continue
		}

		switch tag {
		case "section":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @section requires a name")
			}
			bp.section = pendingSection{name: value}
			i++

		case "internal":
			isInternal = true
			i++

		case "deprecated":
			docblock.IsDeprecated = true
			docblock.DeprecatedMessage = value
			i++

		case "warning":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @warning requires a message")
			}
			docblock.Warnings = append(docblock.Warnings, value)
			i++

		case "description":
			// Only the last @description becomes the function description;
			// earlier ones are discarded. Collect the new one.
			var parts []string
			if value != "" {
				parts = append(parts, value)
			}
			cont, next := collectUntilNextTag(lines, i+1)
			for _, l := range cont {
				parts = append(parts, stripDesc(l.Raw))
			}
			i = next
			collected := strings.Join(parts, "\n")
			if cleanDescription(collected) == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @description requires content")
			}
			pendingDesc = collected

		case "example":
			cont, next := collectExampleLines(lines, i+1)
			i = next
			var exParts []string
			for _, l := range cont {
				exParts = append(exParts, stripExample(l.Raw))
			}
			docblock.Example = strings.Join(exParts, "\n")
			if strings.TrimSpace(docblock.Example) == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @example requires content on following lines")
			}

		case "option":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @option requires a flag definition")
			} else {
				forms, def, valid := processAtOption(value)
				if valid {
					docblock.Options = append(docblock.Options, OptionEntry{Forms: forms, Definition: def})
				} else {
					bp.warn(lineNum, tagCol(raw), "Invalid format: @option "+value)
					docblock.BadOptions = append(docblock.BadOptions, value)
				}
			}
			i++

		case "arg":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @arg requires $N type description")
			} else if argMatch := bpArgStartRe.FindStringSubmatch(value); argMatch != nil {
				argNumber := argMatch[1]
				sortKey := argNumber
				if sortKey != "@" {
					sortKey = fmt.Sprintf("%3s", sortKey)
				}
				var arg Arg
				if m := bpArgNRe.FindStringSubmatch(value); m != nil {
					if strings.TrimSpace(m[3]) == "" {
						bp.warn(lineNum, tagCol(raw), "Empty value: @arg requires '$N type description'")
					}
					arg = Arg{Name: "$" + m[1], Type: m[2], Description: m[3]}
				} else if m := bpArgAtRe.FindStringSubmatch(value); m != nil {
					if strings.TrimSpace(m[2]) == "" {
						bp.warn(lineNum, tagCol(raw), "Empty value: @arg requires '$@ type description'")
					}
					arg = Arg{Name: "$@", Type: m[1], Description: m[2]}
				}
				tempArgs[sortKey] = arg
			} else {
				// Invalid @arg format — fall through to @option processing,
				// preserving the original awk behaviour.
				bp.warn(lineNum, tagCol(raw), "Invalid format, processed as @option: @arg "+value)
				forms, def, valid := processAtOption(value)
				if valid {
					docblock.Options = append(docblock.Options, OptionEntry{Forms: forms, Definition: def})
				} else {
					bp.warn(lineNum, tagCol(raw), "Invalid format: @option "+value)
					docblock.BadOptions = append(docblock.BadOptions, value)
				}
			}
			i++

		case "noargs":
			docblock.IsNoArgs = true
			i++

		case "set":
			if m := bpSetVarRe.FindStringSubmatch(value); m != nil {
				docblock.Sets = append(docblock.Sets, SetVar{
					Name: m[1], Type: m[2], Description: strings.TrimSpace(m[3]),
				})
			} else {
				bp.warn(lineNum, tagCol(raw), "Invalid format: @set requires at least a variable name")
			}
			i++

		case "env":
			if m := bpSetVarRe.FindStringSubmatch(value); m != nil {
				docblock.Env = append(docblock.Env, SetVar{
					Name: m[1], Type: m[2], Description: strings.TrimSpace(m[3]),
				})
			} else {
				bp.warn(lineNum, tagCol(raw), "Invalid format: @env requires at least a variable name")
			}
			i++

		case "exitcode":
			if m := bpExitCodeRe.FindStringSubmatch(value); m != nil {
				if strings.TrimSpace(m[2]) == "" {
					bp.warn(lineNum, tagCol(raw), "Empty value: @exitcode requires 'code description'")
				}
				docblock.ExitCodes = append(docblock.ExitCodes, ExitCode{Code: m[1], Description: m[2]})
			} else {
				bp.warn(lineNum, tagCol(raw), "Invalid format: @exitcode requires 'code description'")
			}
			i++

		case "stdin", "stdout", "stderr":
			// Extract the indentation prefix from the raw tag line to build
			// the continuation regex (same logic as the original multiLineIndentRegex).
			var indent string
			if m := bpStdioIndentRe.FindStringSubmatch(raw); m != nil {
				indent = m[1]
			}
			entry := strings.TrimRight(value, " \t")
			j := i + 1
			if indent != "" {
				if bp.contReCache == nil {
					bp.contReCache = make(map[string]*regexp.Regexp)
				}
				key := regexp.QuoteMeta(indent)
				contRe, ok := bp.contReCache[key]
				if !ok {
					contRe = regexp.MustCompile(`^` + key + `\s+\S.*$`)
					bp.contReCache[key] = contRe
				}
				for j < len(lines) && contRe.MatchString(lines[j].Raw) {
					cont := bpStripContRe.ReplaceAllString(lines[j].Raw, "")
					cont = strings.TrimRight(cont, " \t")
					entry += "\n" + cont
					j++
				}
			}
			if strings.TrimSpace(entry) == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @"+tag+" requires a description")
			}
			switch tag {
			case "stdin":
				docblock.Stdin = append(docblock.Stdin, entry)
			case "stdout":
				docblock.Stdout = append(docblock.Stdout, entry)
			case "stderr":
				docblock.Stderr = append(docblock.Stderr, entry)
			}
			i = j

		case "see":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @see requires a reference")
			}
			docblock.See = append(docblock.See, parseSeeRef(value))
			i++

		default:
			i++
		}
	}

	if isInternal {
		return
	}

	finalDesc := cleanDescription(pendingDesc)
	docblock.Description = finalDesc

	if !docblock.hasDocumentation() && docblock.Description == "" {
		return
	}

	// Sort args by key (zero-padded numeric, "@" last).
	sortKeys := make([]string, 0, len(tempArgs))
	for k := range tempArgs {
		sortKeys = append(sortKeys, k)
	}
	sort.Strings(sortKeys)
	for _, k := range sortKeys {
		docblock.Args = append(docblock.Args, tempArgs[k])
	}

	docblock.Name = block.FuncName
	docblock.Section = bp.section.name
	docblock.SectionDesc = bp.section.desc

	bp.doc.Functions = append(bp.doc.Functions, docblock)
}
