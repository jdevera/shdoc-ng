package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// renderArgN converts "$N type desc" to "**$N** (type): desc"
var argNRegex = regexp.MustCompile(`^\$([0-9]+)\s+(\S+)\s+`)

func renderArgN(text string) string {
	return argNRegex.ReplaceAllString(text, `**$$${1}** (${2}): `)
}

// renderArgAt converts "$@ type desc" to "**...** (type): desc"
var argAtRegex = regexp.MustCompile(`^\$@\s+(\S+)\s+`)

func renderArgAt(text string) string {
	return argAtRegex.ReplaceAllString(text, `**...** (${1}): `)
}

// renderSet converts "VAR type rest" to "**VAR** (type): rest"
var setRegex = regexp.MustCompile(`^(\S+) (\S+)`)

func renderSet(text string) string {
	return setRegex.ReplaceAllString(text, `**${1}** (${2}):`)
}

// renderExitCode converts "code desc" to "**code**: desc"
var exitCodeRegex = regexp.MustCompile(`([>!]?[0-9]{1,3}) (.*)`)

func renderExitCode(text string) string {
	return exitCodeRegex.ReplaceAllString(text, `**${1}**: ${2}`)
}

// unindent removes common leading whitespace from text lines.
// Matches the awk implementation precisely.
func unindent(text string) string {
	lines := strings.Split(text, "\n")

	// Find first non-empty line and max indent
	// Use -1 as sentinel since Go arrays are 0-indexed (awk uses 1-indexed)
	start := -1
	maxIndent := 0
	for i := 0; i < len(lines); i++ {
		if lines[i] != "" && start == -1 {
			start = i
		}
		// Count leading spaces
		spaces := 0
		for _, ch := range lines[i] {
			if ch == ' ' {
				spaces++
			} else {
				break
			}
		}
		if spaces > maxIndent {
			maxIndent = spaces
		}
	}

	// If no non-empty lines found, return empty
	if start == -1 {
		return ""
	}

	// Find minimum indent from start
	indent := maxIndent
	for i := start; i < len(lines); i++ {
		spaces := 0
		for _, ch := range lines[i] {
			if ch == ' ' {
				spaces++
			} else {
				break
			}
		}
		if spaces < indent {
			indent = spaces
		}
	}

	// Remove indent and join from start
	var result strings.Builder
	for i := start; i < len(lines); i++ {
		if i > start {
			result.WriteString("\n")
		}
		if len(lines[i]) > indent {
			result.WriteString(lines[i][indent:])
		} else {
			result.WriteString(lines[i])
		}
	}

	return result.String()
}

// concat joins two strings with a newline, or returns the non-empty one.
func concat(x, text string) string {
	if x == "" {
		return text
	}
	return x + "\n" + text
}

// renderFuncDoc renders the complete markdown for a single function.
func renderFuncDoc(f *FuncDoc, section *string, sectionDesc *string) string {
	var lines []string

	if *section != "" {
		lines = append(lines, "## "+*section+"\n")
		if *sectionDesc != "" {
			lines = append(lines, *sectionDesc)
			lines = append(lines, "")
		}
		*section = ""
		*sectionDesc = ""
		lines = append(lines, "### "+f.Name+"\n")
	} else {
		lines = append(lines, "### "+f.Name+"\n")
	}

	if f.Description != "" {
		lines = append(lines, f.Description)
		lines = append(lines, "")
	}

	if f.Example != "" {
		lines = append(lines, "#### Example\n")
		lines = append(lines, "```bash")
		lines = append(lines, unindent(f.Example))
		lines = append(lines, "```")
		lines = append(lines, "")
	}

	if len(f.Options) > 0 || len(f.BadOptions) > 0 {
		lines = append(lines, "#### Options\n")

		for _, opt := range f.Options {
			term := renderOptionTerm(opt.Term)
			lines = append(lines, "* "+term+"\n")
			lines = append(lines, "  "+opt.Definition+"\n")
		}

		if len(f.BadOptions) > 0 {
			for _, bad := range f.BadOptions {
				lines = append(lines, "* "+bad)
			}
			lines = append(lines, "")
		}
	}

	if len(f.Args) > 0 {
		lines = append(lines, "#### Arguments\n")

		// Sort args by zero-padded key
		keys := make([]string, 0, len(f.Args))
		for k := range f.Args {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			item := f.Args[k]
			item = renderArgN(item)
			item = renderArgAt(item)
			lines = append(lines, "* "+item)
		}
		lines = append(lines, "")
	}

	if f.NoArgs {
		lines = append(lines, "_Function has no arguments._")
		lines = append(lines, "")
	}

	if len(f.Sets) > 0 {
		lines = append(lines, "#### Variables set\n")
		for _, s := range f.Sets {
			item := renderSet(s)
			lines = append(lines, "* "+item)
		}
		lines = append(lines, "")
	}

	if len(f.ExitCodes) > 0 {
		lines = append(lines, "#### Exit codes\n")
		for _, e := range f.ExitCodes {
			item := renderExitCode(e)
			lines = append(lines, "* "+item)
		}
		lines = append(lines, "")
	}

	if len(f.Stdin) > 0 {
		lines = append(lines, "#### Input on stdin\n")
		for _, s := range f.Stdin {
			// Indent additional lines for markdown list item
			item := strings.ReplaceAll(s, "\n", "\n  ")
			lines = append(lines, "* "+item)
		}
		lines = append(lines, "")
	}

	if len(f.Stdout) > 0 {
		lines = append(lines, "#### Output on stdout\n")
		for _, s := range f.Stdout {
			item := strings.ReplaceAll(s, "\n", "\n  ")
			lines = append(lines, "* "+item)
		}
		lines = append(lines, "")
	}

	if len(f.Stderr) > 0 {
		lines = append(lines, "#### Output on stderr\n")
		for _, s := range f.Stderr {
			item := strings.ReplaceAll(s, "\n", "\n  ")
			lines = append(lines, "* "+item)
		}
		lines = append(lines, "")
	}

	if len(f.See) > 0 {
		lines = append(lines, "#### See also\n")
		for _, s := range f.See {
			lines = append(lines, "* "+renderTocLink(s))
		}
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

// renderDocument assembles the final document output.
func renderDocument(doc *Document) string {
	var parts []string

	if doc.FileTitle != "" {
		parts = append(parts, "# "+doc.FileTitle+"\n")

		if doc.FileBrief != "" {
			parts = append(parts, doc.FileBrief+"\n")
		}

		if doc.FileDescription != "" {
			parts = append(parts, "## Overview\n")
			parts = append(parts, doc.FileDescription+"\n")
		}
	}

	if len(doc.TOC) > 0 {
		parts = append(parts, "## Index\n")
		parts = append(parts, strings.Join(doc.TOC, "\n")+"\n")
	}

	// Join header parts
	header := ""
	for _, p := range parts {
		header = fmt.Sprintf("%s%s\n", header, p)
	}

	// The final output is header + doc string + trailing newline
	// awk: print doc (which adds \n)
	return header + doc.DocStr + "\n"
}
