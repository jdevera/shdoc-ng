package main

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// Parser holds the state for parsing shell script documentation.
type Parser struct {
	doc         Document
	description string
	docblock    FuncDoc

	section     string
	sectionDesc string

	inDescription bool
	inExample     bool

	multiLineDocblockName  string
	multiLineIndentRegex   *regexp.Regexp

	functionDeclaration string
	isInternal          bool

	tempArgs map[string]Arg

	lineNum int
}

// NewParser creates a new Parser.
func NewParser() *Parser {
	return &Parser{
		tempArgs: make(map[string]Arg),
	}
}

func (p *Parser) warn(message string) {
	warnColor := "\033[1;34m"
	colorClear := "\033[1;0m"
	fmt.Fprintf(os.Stderr, "%sline %4d, warning : %s%s\n", warnColor, p.lineNum, message, colorClear)
}

func (p *Parser) reset() {
	p.docblock = FuncDoc{}
	p.tempArgs = make(map[string]Arg)
	p.description = ""
}

func (p *Parser) handleDescription() {
	// Remove empty lines at the start of description
	desc := p.description
	desc = regexp.MustCompile(`^[\s\n]*\n`).ReplaceAllString(desc, "")
	// Remove empty lines at the end of description
	desc = regexp.MustCompile(`[\s\n]*$`).ReplaceAllString(desc, "")
	p.description = desc

	if p.description == "" {
		return
	}

	if p.section != "" && p.sectionDesc == "" {
		p.sectionDesc = p.description
		return
	}

	if p.doc.FileDescription == "" {
		p.doc.FileDescription = p.description
		return
	}
}

func (p *Parser) processFunction(text string) {
	if (!p.docblock.hasDocumentation() && !p.isInternal &&
		p.description == "") || p.inExample {
		return
	}

	if p.isInternal {
		p.isInternal = false
	} else {
		// Convert tempArgs map to sorted slice
		sortKeys := make([]string, 0, len(p.tempArgs))
		for k := range p.tempArgs {
			sortKeys = append(sortKeys, k)
		}
		sort.Strings(sortKeys)
		for _, k := range sortKeys {
			p.docblock.Args = append(p.docblock.Args, p.tempArgs[k])
		}

		// Extract function name
		funcNameRegex := regexp.MustCompile(
			`^\s*(?:function\s+)?([a-zA-Z0-9_\-:.\-]+)\s*(?:\(\s*\))?\s*\{?`,
		)
		m := funcNameRegex.FindStringSubmatch(text)
		funcName := ""
		if m != nil {
			funcName = m[1]
		}

		p.docblock.Name = funcName
		p.docblock.Description = p.description

		// Copy section info to the FuncDoc
		p.docblock.Section = p.section
		p.docblock.SectionDesc = p.sectionDesc
		p.section = ""
		p.sectionDesc = ""

		// Store the function doc for later rendering
		p.doc.Functions = append(p.doc.Functions, p.docblock)
	}

	p.reset()
}

// Regex patterns
var (
	internalRegex   = regexp.MustCompile(`^[\s]*# @internal`)
	deprecatedRegex = regexp.MustCompile(`^[\s]*# @deprecated\s*(.*)`)
	nameFileRegex   = regexp.MustCompile(`^[\s]*# @(name|file) `)
	briefRegex      = regexp.MustCompile(`^[\s]*# @brief `)
	authorRegex     = regexp.MustCompile(`^[\s]*# @author `)
	licenseRegex    = regexp.MustCompile(`^[\s]*# @license `)
	versionRegex    = regexp.MustCompile(`^[\s]*# @version `)

	descriptionTagRegex = regexp.MustCompile(`^[\s]*# @description`)

	// inDescription exit condition from awk:
	// /^[^[[:space:]]*#]|^[[:space:]]*# @[^d]|^[[:space:]]*[^#]|^[[:space:]]*$/
	inDescriptionExitRegex = regexp.MustCompile(`^[^\s#]|^\s*# @[^d]|^\s*[^#]|^\s*$`)

	descriptionStripTagRegex  = regexp.MustCompile(`^\s*# @description\s*`)
	descriptionStripHashRegex = regexp.MustCompile(`^\s*#\s*`)
	descriptionStripEmptyHash = regexp.MustCompile(`^\s*#$`)

	sectionRegex = regexp.MustCompile(`^[\s]*# @section `)

	exampleRegex = regexp.MustCompile(`^[\s]*# @example`)

	// in_example continuation: /^[[:space:]]*#[ ]{1,}/
	inExampleContinueRegex = regexp.MustCompile(`^[\s]*#[ ]+`)

	optionRegexLine = regexp.MustCompile(`^[\t ]*#[\t ]+@option[\t ]+[^\t ]`)
	argRegexLine    = regexp.MustCompile(`^[\t ]*#[\t ]+@arg[\t ]+[^\t ]`)
	noargsRegex     = regexp.MustCompile(`^[\s]*#[\t ]+@noargs[\t ]*$`)

	setRegexLine      = regexp.MustCompile(`^[\s]*# @set `)
	envRegexLine      = regexp.MustCompile(`^[\s]*# @env `)
	exitcodeRegexLine = regexp.MustCompile(`^[\s]*# @exitcode `)
	seeRegexLine      = regexp.MustCompile(`^[\s]*# @see `)
	warningRegexLine  = regexp.MustCompile(`^[\s]*# @warning `)

	// Structured-field parsing
	argNRegex    = regexp.MustCompile(`^\$([0-9]+)\s+(\S+)\s+(.*)$`)
	argAtRegex   = regexp.MustCompile(`^\$@\s+(\S+)\s+(.*)$`)
	setVarRegex  = regexp.MustCompile(`^(\S+)\s+(\S+)\s*(.*)$`)
	exitCodeRegex = regexp.MustCompile(`^([>!]?[0-9]{1,3}) (.*)$`)

	// Multi-line stdin/stdout/stderr
	stdioRegex = regexp.MustCompile(`^([\t ]*#[\t ]+)@(stdin|stdout|stderr)[\t ]+(.*[^\t ])[\t ]*$`)

	// Function declaration with opening brace or paren (subshell)
	funcDeclWithBrace = regexp.MustCompile(
		`^[\t ]*(function[\t ]+)?([a-zA-Z0-9_\-:.\-]+)[\t ]*(\([\t ]*\))?[\t ]*[{(]`,
	)

	// Function declaration without opening brace (for multi-line)
	funcDeclWithoutBrace = regexp.MustCompile(
		`^[\t ]*(function[\t ]+)?([a-zA-Z0-9_\-:.\-]+)[\t ]*(\([\t ]*\))?[\t ]*$`,
	)

	// Lone opening brace or paren (subshell)
	loneBrace = regexp.MustCompile(`^[\t ]*[{(]`)

	// Empty line
	emptyLine = regexp.MustCompile(`^[\t ]*$`)

	// Non-comment line
	nonCommentLine = regexp.MustCompile(`^[^#]*$`)
)

// ProcessLine processes a single line of input.
func (p *Parser) ProcessLine(line string) {
	p.lineNum++

	// Rule 1: @internal
	if internalRegex.MatchString(line) {
		p.isInternal = true
		return
	}

	// Rule 1b: @deprecated
	if deprecatedRegex.MatchString(line) {
		m := deprecatedRegex.FindStringSubmatch(line)
		reason := strings.TrimSpace(m[1])
		if reason == "" {
			p.docblock.Deprecated = " "
		} else {
			p.docblock.Deprecated = reason
		}
		return
	}

	// Rule 1c: @warning
	if warningRegexLine.MatchString(line) {
		stripped := warningRegexLine.ReplaceAllString(line, "")
		p.docblock.Warnings = append(p.docblock.Warnings, stripped)
		return
	}

	// Rule 2: @name/@file
	if nameFileRegex.MatchString(line) {
		stripped := nameFileRegex.ReplaceAllString(line, "")
		p.doc.FileTitle = stripped
		return
	}

	// Rule 3: @brief
	if briefRegex.MatchString(line) {
		stripped := briefRegex.ReplaceAllString(line, "")
		p.doc.FileBrief = stripped
		return
	}

	// Rule 3b: @author (file-level, list)
	if authorRegex.MatchString(line) {
		stripped := authorRegex.ReplaceAllString(line, "")
		p.doc.Authors = append(p.doc.Authors, stripped)
		return
	}

	// Rule 3c: @license (file-level, single)
	if licenseRegex.MatchString(line) {
		stripped := licenseRegex.ReplaceAllString(line, "")
		p.doc.License = stripped
		return
	}

	// Rule 3d: @version (file-level, single)
	if versionRegex.MatchString(line) {
		stripped := versionRegex.ReplaceAllString(line, "")
		p.doc.Version = stripped
		return
	}

	// Rule 4: @description tag
	if descriptionTagRegex.MatchString(line) {
		p.inDescription = true
		p.inExample = false

		p.handleDescription()
		p.reset()

		// NOTE: Fall through to inDescription block (don't return)
	}

	// Rule 5: inDescription continuation or exit
	if p.inDescription {
		if !descriptionTagRegex.MatchString(line) && inDescriptionExitRegex.MatchString(line) {
			// Leave description mode
			p.inDescription = false
			p.handleDescription()
			// Don't return - fall through to process the current line
		} else {
			// Continue collecting description
			stripped := line
			stripped = descriptionStripTagRegex.ReplaceAllString(stripped, "")
			stripped = descriptionStripHashRegex.ReplaceAllString(stripped, "")
			stripped = descriptionStripEmptyHash.ReplaceAllString(stripped, "")
			p.description = concat(p.description, stripped)
			return
		}
	}

	// Rule 6: @section
	if sectionRegex.MatchString(line) {
		stripped := sectionRegex.ReplaceAllString(line, "")
		p.section = stripped
		return
	}

	// Rule 7: @example
	if exampleRegex.MatchString(line) {
		p.inExample = true
		return
	}

	// Rule 8: inExample continuation or exit
	if p.inExample {
		if !inExampleContinueRegex.MatchString(line) {
			// Leave example mode
			p.inExample = false
			// Fall through
		} else {
			// Continue collecting example - strip leading "# " pattern
			stripped := regexp.MustCompile(`^\s*#`).ReplaceAllString(line, "")
			// Concatenate to example (awk uses docblock_concat which is like concat())
			if p.docblock.Example == "" {
				p.docblock.Example = stripped
			} else {
				p.docblock.Example = p.docblock.Example + "\n" + stripped
			}
			return
		}
	}

	// Rule 9: @option
	if optionRegexLine.MatchString(line) {
		optionText := regexp.MustCompile(`^[\t ]*#[\t ]+@option[\t ]+`).ReplaceAllString(line, "")
		optionText = strings.TrimSpace(optionText)

		forms, def, valid := processAtOption(optionText)
		if valid {
			p.docblock.Options = append(p.docblock.Options, OptionEntry{Forms: forms, Definition: def})
		} else {
			p.warn("Invalid format: @option " + optionText)
			p.docblock.BadOptions = append(p.docblock.BadOptions, optionText)
		}
		return
	}

	// Rule 10: @arg
	if argRegexLine.MatchString(line) {
		argText := regexp.MustCompile(`^[\t ]*#[\t ]+@arg[\t ]+`).ReplaceAllString(line, "")
		argText = strings.TrimSpace(argText)

		// Check if it's a numbered arg ($N or $@)
		argMatch := regexp.MustCompile(`^\$([0-9]+|@)\s`).FindStringSubmatch(argText)
		if argMatch != nil {
			argNumber := argMatch[1]
			// Zero-pad numeric arguments for sort order
			sortKey := argNumber
			if sortKey != "@" {
				sortKey = fmt.Sprintf("%03s", sortKey)
			}
			var arg Arg
			if m := argNRegex.FindStringSubmatch(argText); m != nil {
				arg = Arg{Name: "$" + m[1], Type: m[2], Description: m[3]}
			} else if m := argAtRegex.FindStringSubmatch(argText); m != nil {
				arg = Arg{Name: "$@", Type: m[1], Description: m[2]}
			}
			p.tempArgs[sortKey] = arg
			return
		}

		// Invalid @arg format - process as @option with warning
		p.warn("Invalid format, processed as @option: @arg " + argText)
		forms, def, valid := processAtOption(argText)
		if valid {
			p.docblock.Options = append(p.docblock.Options, OptionEntry{Forms: forms, Definition: def})
		} else {
			p.warn("Invalid format: @option " + argText)
			p.docblock.BadOptions = append(p.docblock.BadOptions, argText)
		}
		return
	}

	// Rule 11: @noargs
	if noargsRegex.MatchString(line) {
		p.docblock.NoArgs = true
		return
	}

	// Rule 12: @set
	if setRegexLine.MatchString(line) {
		stripped := setRegexLine.ReplaceAllString(line, "")
		if m := setVarRegex.FindStringSubmatch(stripped); m != nil {
			p.docblock.Sets = append(p.docblock.Sets, SetVar{Name: m[1], Type: m[2], Description: strings.TrimSpace(m[3])})
		}
		return
	}

	// Rule 12b: @env
	if envRegexLine.MatchString(line) {
		stripped := envRegexLine.ReplaceAllString(line, "")
		if m := setVarRegex.FindStringSubmatch(stripped); m != nil {
			p.docblock.Env = append(p.docblock.Env, SetVar{Name: m[1], Type: m[2], Description: strings.TrimSpace(m[3])})
		}
		return
	}

	// Rule 12c: @exitcode
	if exitcodeRegexLine.MatchString(line) {
		stripped := exitcodeRegexLine.ReplaceAllString(line, "")
		if m := exitCodeRegex.FindStringSubmatch(stripped); m != nil {
			p.docblock.ExitCodes = append(p.docblock.ExitCodes, ExitCode{Code: m[1], Description: m[2]})
		}
		return
	}

	// Rule 12: @see
	if seeRegexLine.MatchString(line) {
		// awk: sub(/[[:space:]]*# @see /, "")
		stripped := regexp.MustCompile(`[\s]*# @see `).ReplaceAllString(line, "")
		p.docblock.See = append(p.docblock.See, parseSeeRef(stripped))
		return
	}

	// Rule 13: Multi-line docblock continuation
	if p.multiLineDocblockName != "" {
		if p.multiLineIndentRegex != nil && p.multiLineIndentRegex.MatchString(line) {
			// Continue multi-line entry
			stripped := regexp.MustCompile(`^\s*#\s+`).ReplaceAllString(line, "")
			stripped = strings.TrimRight(stripped, " \t")

			// Append to last item in the appropriate slice
			switch p.multiLineDocblockName {
			case "stdin":
				if len(p.docblock.Stdin) > 0 {
					p.docblock.Stdin[len(p.docblock.Stdin)-1] += "\n" + stripped
				}
			case "stdout":
				if len(p.docblock.Stdout) > 0 {
					p.docblock.Stdout[len(p.docblock.Stdout)-1] += "\n" + stripped
				}
			case "stderr":
				if len(p.docblock.Stderr) > 0 {
					p.docblock.Stderr[len(p.docblock.Stderr)-1] += "\n" + stripped
				}
			}
			return
		}
		// End multi-line
		p.multiLineDocblockName = ""
		p.multiLineIndentRegex = nil
		// Fall through
	}

	// Rule 14: @stdin/@stdout/@stderr
	if m := stdioRegex.FindStringSubmatch(line); m != nil {
		indentation := m[1]
		docblockName := m[2]
		text := m[3]

		switch docblockName {
		case "stdin":
			p.docblock.Stdin = append(p.docblock.Stdin, text)
		case "stdout":
			p.docblock.Stdout = append(p.docblock.Stdout, text)
		case "stderr":
			p.docblock.Stderr = append(p.docblock.Stderr, text)
		}

		// Start multi-line mode
		p.multiLineDocblockName = docblockName
		// Build regex: ^<indentation>\s+\S.*$
		escapedIndent := regexp.QuoteMeta(indentation)
		p.multiLineIndentRegex = regexp.MustCompile(`^` + escapedIndent + `\s+\S.*$`)
		return
	}

	// Rule 15: Function declaration with brace
	if funcDeclWithBrace.MatchString(line) {
		p.processFunction(line)
		return
	}

	// Rule 16: Function declaration without brace (store for later)
	if funcDeclWithoutBrace.MatchString(line) {
		p.functionDeclaration = line
		return
	}

	// Rule 17: Lone brace with stored declaration
	if loneBrace.MatchString(line) && p.functionDeclaration != "" {
		p.processFunction(p.functionDeclaration)
		return
	}

	// Rule: Empty line while waiting for opening bracket
	if emptyLine.MatchString(line) && p.functionDeclaration != "" {
		return
	}

	// Rule 18: Non-comment line (break)
	if nonCommentLine.MatchString(line) {
		p.functionDeclaration = ""
		p.handleDescription()
		p.reset()
		return
	}
}

// Render produces the final document output.
func (p *Parser) Render() string {
	return renderDocument(&p.doc)
}
