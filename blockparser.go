package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Warning is a parse warning collected during ParseDocument.
type Warning struct {
	Line    int
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

	// Parses @set / @env: varname type description
	bpSetVarRe = regexp.MustCompile(`^(\S+)\s+(\S+)\s*(.*)$`)

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
	lines := lexLines(src)
	blocks := segmentBlocks(lines)
	bp := &blockParser{}
	for _, block := range blocks {
		if block.Kind == MetaBlockKind {
			bp.parseMetaBlock(block)
		} else {
			bp.parseFuncBlock(block)
		}
	}
	return bp.doc, bp.warns
}

type blockParser struct {
	doc     Document
	warns   []Warning
	section pendingSection
}

type pendingSection struct {
	name string
	desc string
}

func (bp *blockParser) warn(lineNum int, msg string) {
	bp.warns = append(bp.warns, Warning{Line: lineNum, Message: msg})
}

// isTagLine returns true if a raw comment line contains a @tag.
func isTagLine(raw string) bool {
	return strings.HasPrefix(bpStripRe.ReplaceAllString(raw, ""), "@")
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
		i++
	}
	return lines[start:i], i
}

// stripDesc strips the leading `# ` prefix from a description line,
// removing all whitespace after the `#` (aggressive strip).
func stripDesc(raw string) string {
	return bpStripRe.ReplaceAllString(raw, "")
}

// stripExample strips just the leading `#` from an example line,
// preserving any whitespace that follows (for the unindent template function).
func stripExample(raw string) string {
	return bpStripHashRe.ReplaceAllString(raw, "")
}

// parseTag extracts a tag name and value from a stripped line.
// Returns ("", "") if the line does not start with @.
func parseTag(stripped string) (name, value string) {
	m := bpTagRe.FindStringSubmatch(stripped)
	if m == nil {
		return "", ""
	}
	return m[1], strings.TrimSpace(m[2])
}

// cleanDescription trims leading and trailing empty lines.
func cleanDescription(s string) string {
	s = bpCleanLeadingRe.ReplaceAllString(s, "")
	s = bpCleanTrailingRe.ReplaceAllString(s, "")
	return s
}

// routeDescription routes a cleaned description to the appropriate document
// field. This mirrors the original handleDescription() behaviour, including
// its double-routing quirk: when a section is waiting for its description,
// the text goes to section.desc AND also to FileDescription (if still empty).
// This happens because the original parser calls handleDescription() twice for
// the same description string — once on description-mode exit and again on
// the blank-line Rule 18 — without clearing p.description in between.
func (bp *blockParser) routeDescription(desc string) {
	if desc == "" {
		return
	}
	if bp.section.name != "" && bp.section.desc == "" {
		bp.section.desc = desc
		// fall through — do NOT return; the original parser's second
		// handleDescription() call also tries FileDescription.
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
		tag, value := parseTag(stripped)
		if tag == "" {
			i++
			continue
		}
		switch tag {
		case "name", "file":
			bp.doc.FileTitle = value
			i++
		case "brief":
			bp.doc.FileBrief = value
			i++
		case "author":
			bp.doc.Authors = append(bp.doc.Authors, value)
			i++
		case "license":
			bp.doc.License = value
			i++
		case "version":
			bp.doc.Version = value
			i++
		case "section":
			// Only update the name — intentionally preserve bp.section.desc.
			// The original parser never clears p.sectionDesc when p.section
			// changes, so an earlier section's description can persist and be
			// picked up by a later section with no description of its own.
			bp.section.name = value
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
			bp.routeDescription(cleanDescription(strings.Join(parts, "\n")))
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
		tag, value := parseTag(stripped)
		if tag == "" {
			i++
			continue
		}

		switch tag {
		// File-level tags may appear in any block (e.g. @name at the top of a
		// file whose first comment block also documents a function).
		case "name", "file":
			bp.doc.FileTitle = value
			i++
		case "brief":
			bp.doc.FileBrief = value
			i++
		case "author":
			bp.doc.Authors = append(bp.doc.Authors, value)
			i++
		case "license":
			bp.doc.License = value
			i++
		case "version":
			bp.doc.Version = value
			i++

		case "section":
			// Only update the name — preserve bp.section.desc (see parseMetaBlock).
			bp.section.name = value
			i++

		case "internal":
			isInternal = true
			i++

		case "deprecated":
			docblock.IsDeprecated = true
			docblock.DeprecatedMessage = value
			i++

		case "warning":
			docblock.Warnings = append(docblock.Warnings, value)
			i++

		case "description":
			// Route any previously accumulated description through the normal
			// meta routing (file description, section description).
			bp.routeDescription(cleanDescription(pendingDesc))
			pendingDesc = ""
			// Collect the new description.
			var parts []string
			if value != "" {
				parts = append(parts, value)
			}
			cont, next := collectUntilNextTag(lines, i+1)
			for _, l := range cont {
				parts = append(parts, stripDesc(l.Raw))
			}
			i = next
			pendingDesc = strings.Join(parts, "\n")

		case "example":
			cont, next := collectExampleLines(lines, i+1)
			i = next
			var exParts []string
			for _, l := range cont {
				exParts = append(exParts, stripExample(l.Raw))
			}
			docblock.Example = strings.Join(exParts, "\n")

		case "option":
			forms, def, valid := processAtOption(value)
			if valid {
				docblock.Options = append(docblock.Options, OptionEntry{Forms: forms, Definition: def})
			} else {
				bp.warn(lineNum, "Invalid format: @option "+value)
				docblock.BadOptions = append(docblock.BadOptions, value)
			}
			i++

		case "arg":
			if argMatch := bpArgStartRe.FindStringSubmatch(value); argMatch != nil {
				argNumber := argMatch[1]
				sortKey := argNumber
				if sortKey != "@" {
					sortKey = fmt.Sprintf("%03s", sortKey)
				}
				var arg Arg
				if m := bpArgNRe.FindStringSubmatch(value); m != nil {
					arg = Arg{Name: "$" + m[1], Type: m[2], Description: m[3]}
				} else if m := bpArgAtRe.FindStringSubmatch(value); m != nil {
					arg = Arg{Name: "$@", Type: m[1], Description: m[2]}
				}
				tempArgs[sortKey] = arg
			} else {
				// Invalid @arg format — fall through to @option processing,
				// preserving the original awk behaviour.
				bp.warn(lineNum, "Invalid format, processed as @option: @arg "+value)
				forms, def, valid := processAtOption(value)
				if valid {
					docblock.Options = append(docblock.Options, OptionEntry{Forms: forms, Definition: def})
				} else {
					bp.warn(lineNum, "Invalid format: @option "+value)
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
			}
			i++

		case "env":
			if m := bpSetVarRe.FindStringSubmatch(value); m != nil {
				docblock.Env = append(docblock.Env, SetVar{
					Name: m[1], Type: m[2], Description: strings.TrimSpace(m[3]),
				})
			}
			i++

		case "exitcode":
			if m := bpExitCodeRe.FindStringSubmatch(value); m != nil {
				docblock.ExitCodes = append(docblock.ExitCodes, ExitCode{Code: m[1], Description: m[2]})
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
				escapedIndent := regexp.QuoteMeta(indent)
				contRe := regexp.MustCompile(`^` + escapedIndent + `\s+\S.*$`)
				for j < len(lines) && contRe.MatchString(lines[j].Raw) {
					cont := bpStripContRe.ReplaceAllString(lines[j].Raw, "")
					cont = strings.TrimRight(cont, " \t")
					entry += "\n" + cont
					j++
				}
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
			docblock.See = append(docblock.See, parseSeeRef(value))
			i++

		default:
			i++
		}
	}

	if isInternal {
		return
	}

	// Route the final pending description through the same meta routing as
	// the original parser — the description exit triggered handleDescription()
	// before processFunction() consumed p.description. This means a function
	// description can also populate doc.FileDescription (when still empty) or
	// bp.section.desc (when the section is waiting for its description).
	finalDesc := cleanDescription(pendingDesc)
	bp.routeDescription(finalDesc)
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
	bp.section = pendingSection{}

	bp.doc.Functions = append(bp.doc.Functions, docblock)
}
