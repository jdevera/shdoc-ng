package shdoc

import (
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
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
	// Start with one unnamed section for functions before any @section tag.
	bp.doc.Sections = []Section{{}}
	bp.currentSection = 0

	for _, block := range blocks {
		if block.Kind == MetaBlockKind {
			bp.parseMetaBlock(block)
		} else {
			bp.parseFuncBlock(block)
		}
	}

	// Prune sections with no functions.
	var pruned []Section
	for _, s := range bp.doc.Sections {
		if len(s.Functions) == 0 {
			continue
		}
		pruned = append(pruned, s)
	}
	bp.doc.Sections = pruned

	return bp.doc, bp.warns
}

type blockParser struct {
	doc            Document
	warns          []Warning
	currentSection int // index into doc.Sections
	// Track which file-level singleton tags have been set, to warn on duplicates.
	seenFileTags map[string]int // tag -> first line number
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

func (bp *blockParser) warn(lineNum int, col int, msg string) {
	bp.warns = append(bp.warns, Warning{Line: lineNum, Col: col, Message: msg})
}

// tagCol returns the absolute 0-based column of the '@' in a raw comment line,
// or 0 if not found.
func tagCol(raw string) int {
	return strings.Index(raw, "@")
}

// commentCol returns the absolute 0-based column of the '#' in a raw comment line,
// or 0 if not found.
func commentCol(raw string) int {
	return strings.Index(raw, "#")
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
// Used by @description which flows at the same indent level as the tag.
// Returns the collected lines and the index of the first tag line (or len(lines)).
func collectUntilNextTag(lines []LexedLine, start int) ([]LexedLine, int) {
	i := start
	for i < len(lines) && !isTagLine(lines[i].Raw) {
		i++
	}
	return lines[start:i], i
}

// collectContinuation collects indented continuation lines from lines[start:].
//
// tagAbsCol is the absolute 0-based column of the '@' in the tag line.
// tagHashCol is the absolute 0-based column of the '#' in the tag line.
//
// A continuation line must:
//   - have its '#' at the same column as the tag line's '#'
//   - have text content starting at an absolute column greater than tagAbsCol
//
// The indentation of the first continuation line (spaces after '#') sets the
// baseline. That baseline is stripped from all continuation lines, preserving
// any additional indentation.
//
// Collection stops at: a @tag line, mismatched '#' column, insufficient
// indentation, bare '#' (no content), or end of the block.
//
// Returns the collected text lines (with baseline stripped) and the next index.
func collectContinuation(lines []LexedLine, start int, tagAbsCol int, tagHashCol int) ([]string, int) {
	i := start
	baseline := -1 // spaces after '#' of the first continuation line
	var result []string

	for i < len(lines) {
		raw := lines[i].Raw

		// Find the '#' (absolute column) and get everything after it.
		lineHashCol := strings.Index(raw, "#")
		if lineHashCol < 0 {
			break
		}
		after := raw[lineHashCol+1:]

		// The '#' must be at the same column as the tag line's '#'.
		if lineHashCol != tagHashCol {
			break
		}

		// A bare '#' with nothing after it (or only whitespace) is not a continuation.
		if strings.TrimSpace(after) == "" {
			break
		}

		// Check if this is a @tag line — if so, stop.
		if isTagLine(raw) {
			break
		}

		// Count leading spaces/tabs after '#' to find the absolute content column.
		spacesAfterHash := 0
		for _, ch := range after {
			if ch == ' ' || ch == '\t' {
				spacesAfterHash++
			} else {
				break
			}
		}
		contentCol := lineHashCol + 1 + spacesAfterHash

		// Content must start past the tag column.
		if contentCol <= tagAbsCol {
			break
		}

		// First continuation line sets the baseline indent (spaces after '#').
		if baseline < 0 {
			baseline = spacesAfterHash
		}

		// Strip exactly the baseline number of leading whitespace characters from after-'#' text.
		stripped := after
		toStrip := baseline
		for toStrip > 0 && len(stripped) > 0 && (stripped[0] == ' ' || stripped[0] == '\t') {
			stripped = stripped[1:]
			toStrip--
		}
		stripped = strings.TrimRight(stripped, " \t")

		result = append(result, stripped)
		i++
	}

	return result, i
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
// document field: current section description (if named and empty) or file
// description.
func (bp *blockParser) routeDescription(desc string) {
	if desc == "" {
		return
	}
	sec := &bp.doc.Sections[bp.currentSection]
	if sec.Name != "" && sec.Description == "" {
		sec.Description = desc
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
			bp.doc.Sections = append(bp.doc.Sections, Section{Name: value})
			bp.currentSection = len(bp.doc.Sections) - 1
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
	tempArgs := make(map[int]Arg)
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
			bp.doc.Sections = append(bp.doc.Sections, Section{Name: value})
			bp.currentSection = len(bp.doc.Sections) - 1
			i++

		case "internal":
			isInternal = true
			i++

		case "deprecated":
			docblock.IsDeprecated = true
			parts := []string{}
			if value != "" {
				parts = append(parts, value)
			}
			cont, next := collectContinuation(lines, i+1, tagCol(raw), commentCol(raw))
			parts = append(parts, cont...)
			docblock.DeprecatedMessage = strings.Join(parts, "\n")
			i = next

		case "warning":
			parts := []string{}
			if value != "" {
				parts = append(parts, value)
			}
			cont, next := collectContinuation(lines, i+1, tagCol(raw), commentCol(raw))
			parts = append(parts, cont...)
			msg := strings.Join(parts, "\n")
			if msg == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @warning requires a message")
			} else {
				docblock.Warnings = append(docblock.Warnings, msg)
			}
			i = next

		case "label":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @label requires at least one label")
			} else {
				for _, l := range strings.Split(value, ",") {
					l = strings.TrimSpace(l)
					if l != "" {
						docblock.Labels = append(docblock.Labels, l)
					}
				}
			}
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
				i++
			} else {
				forms, def, valid := processAtOption(value)
				if valid {
					cont, next := collectContinuation(lines, i+1, tagCol(raw), commentCol(raw))
					if len(cont) > 0 {
						def = def + "\n" + strings.Join(cont, "\n")
					}
					docblock.Options = append(docblock.Options, OptionEntry{Forms: forms, Definition: def})
					i = next
				} else {
					bp.warn(lineNum, tagCol(raw), "Invalid format: @option "+value)
					docblock.BadOptions = append(docblock.BadOptions, value)
					i++
				}
			}

		case "arg":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @arg requires $N type description")
				i++
			} else if argMatch := bpArgStartRe.FindStringSubmatch(value); argMatch != nil {
				argNumber := argMatch[1]
				sortKey := math.MaxInt
				if argNumber != "@" {
					var err error
					sortKey, err = strconv.Atoi(argNumber)
					if err != nil {
						sortKey = math.MaxInt - 1 // overflow fallback, keep near end
					}
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
				cont, next := collectContinuation(lines, i+1, tagCol(raw), commentCol(raw))
				if len(cont) > 0 {
					arg.Description = arg.Description + "\n" + strings.Join(cont, "\n")
				}
				tempArgs[sortKey] = arg
				i = next
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
				i++
			}

		case "noargs":
			docblock.IsNoArgs = true
			i++

		case "set":
			if m := bpSetVarRe.FindStringSubmatch(value); m != nil {
				desc := strings.TrimSpace(m[3])
				cont, next := collectContinuation(lines, i+1, tagCol(raw), commentCol(raw))
				if len(cont) > 0 {
					desc = desc + "\n" + strings.Join(cont, "\n")
				}
				docblock.Sets = append(docblock.Sets, SetVar{
					Name: m[1], Type: m[2], Description: desc,
				})
				i = next
			} else {
				bp.warn(lineNum, tagCol(raw), "Invalid format: @set requires at least a variable name")
				i++
			}

		case "env":
			if m := bpSetVarRe.FindStringSubmatch(value); m != nil {
				desc := strings.TrimSpace(m[3])
				cont, next := collectContinuation(lines, i+1, tagCol(raw), commentCol(raw))
				if len(cont) > 0 {
					desc = desc + "\n" + strings.Join(cont, "\n")
				}
				docblock.Env = append(docblock.Env, SetVar{
					Name: m[1], Type: m[2], Description: desc,
				})
				i = next
			} else {
				bp.warn(lineNum, tagCol(raw), "Invalid format: @env requires at least a variable name")
				i++
			}

		case "exitcode":
			if m := bpExitCodeRe.FindStringSubmatch(value); m != nil {
				if strings.TrimSpace(m[2]) == "" {
					bp.warn(lineNum, tagCol(raw), "Empty value: @exitcode requires 'code description'")
				}
				desc := m[2]
				cont, next := collectContinuation(lines, i+1, tagCol(raw), commentCol(raw))
				if len(cont) > 0 {
					desc = desc + "\n" + strings.Join(cont, "\n")
				}
				docblock.ExitCodes = append(docblock.ExitCodes, ExitCode{Code: m[1], Description: desc})
				i = next
			} else {
				bp.warn(lineNum, tagCol(raw), "Invalid format: @exitcode requires 'code description'")
				i++
			}

		case "stdin", "stdout", "stderr":
			parts := []string{}
			if value != "" {
				parts = append(parts, strings.TrimRight(value, " \t"))
			}
			cont, next := collectContinuation(lines, i+1, tagCol(raw), commentCol(raw))
			parts = append(parts, cont...)
			entry := strings.Join(parts, "\n")
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
			i = next

		case "see":
			if value == "" {
				bp.warn(lineNum, tagCol(raw), "Empty value: @see requires a reference")
			}
			cont, next := collectContinuation(lines, i+1, tagCol(raw), commentCol(raw))
			fullValue := value
			if len(cont) > 0 {
				fullValue = fullValue + "\n" + strings.Join(cont, "\n")
			}
			docblock.See = append(docblock.See, parseSeeRef(fullValue))
			i = next

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

	// Sort args numerically ($1, $2, …), with $@ last.
	sortKeys := make([]int, 0, len(tempArgs))
	for k := range tempArgs {
		sortKeys = append(sortKeys, k)
	}
	sort.Ints(sortKeys)
	for _, k := range sortKeys {
		docblock.Args = append(docblock.Args, tempArgs[k])
	}

	docblock.Name = block.FuncName

	sec := &bp.doc.Sections[bp.currentSection]
	sec.Functions = append(sec.Functions, docblock)
}
