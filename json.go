package main

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
)

// JSONArg represents a parsed function argument for JSON output.
type JSONArg struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// JSONSetVar represents a parsed @set variable for JSON output.
type JSONSetVar struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// JSONExitCode represents a parsed @exitcode for JSON output.
type JSONExitCode struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// JSONFunc is the JSON-friendly representation of a function.
type JSONFunc struct {
	Name               string         `json:"name"`
	Section            string         `json:"section,omitempty"`
	SectionDescription string         `json:"section_description,omitempty"`
	Description        string         `json:"description,omitempty"`
	Deprecated         string         `json:"deprecated,omitempty"`
	Example            string         `json:"example,omitempty"`
	Options            []OptionEntry  `json:"options,omitempty"`
	Args               []JSONArg      `json:"args,omitempty"`
	NoArgs             bool           `json:"noargs,omitempty"`
	Set                []JSONSetVar   `json:"set,omitempty"`
	ExitCodes          []JSONExitCode `json:"exitcodes,omitempty"`
	Env                []JSONSetVar   `json:"env,omitempty"`
	Stdin              []string       `json:"stdin,omitempty"`
	Stdout             []string       `json:"stdout,omitempty"`
	Stderr             []string       `json:"stderr,omitempty"`
	See                []string       `json:"see,omitempty"`
	Warnings           []string       `json:"warnings,omitempty"`
}

// JSONDocument is the JSON-friendly representation of the full document.
type JSONDocument struct {
	Name        string     `json:"name,omitempty"`
	Brief       string     `json:"brief,omitempty"`
	Description string     `json:"description,omitempty"`
	Authors     []string   `json:"authors,omitempty"`
	License     string     `json:"license,omitempty"`
	Version     string     `json:"version,omitempty"`
	Functions   []JSONFunc `json:"functions,omitempty"`
}

// Regex patterns for parsing raw arg/set/exitcode strings into structured types.
var (
	jsonArgNRegex    = regexp.MustCompile(`^\$([0-9]+)\s+(\S+)\s+(.*)$`)
	jsonArgAtRegex   = regexp.MustCompile(`^\$@\s+(\S+)\s+(.*)$`)
	jsonSetVarRegex  = regexp.MustCompile(`^(\S+)\s+(\S+)\s*(.*)$`)
	jsonExitCodeRegex = regexp.MustCompile(`^([>!]?[0-9]{1,3})\s+(.*)$`)
)

func parseArgs(args map[string]string) []JSONArg {
	if len(args) == 0 {
		return nil
	}

	// Sort by zero-padded key (same order as markdown rendering)
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var result []JSONArg
	for _, k := range keys {
		raw := args[k]
		if m := jsonArgNRegex.FindStringSubmatch(raw); m != nil {
			result = append(result, JSONArg{
				Name:        "$" + m[1],
				Type:        m[2],
				Description: m[3],
			})
		} else if m := jsonArgAtRegex.FindStringSubmatch(raw); m != nil {
			result = append(result, JSONArg{
				Name:        "$@",
				Type:        m[1],
				Description: m[2],
			})
		}
	}
	return result
}

func parseSets(sets []string) []JSONSetVar {
	if len(sets) == 0 {
		return nil
	}
	var result []JSONSetVar
	for _, raw := range sets {
		if m := jsonSetVarRegex.FindStringSubmatch(raw); m != nil {
			desc := strings.TrimSpace(m[3])
			result = append(result, JSONSetVar{
				Name:        m[1],
				Type:        m[2],
				Description: desc,
			})
		}
	}
	return result
}

func parseExitCodes(codes []string) []JSONExitCode {
	if len(codes) == 0 {
		return nil
	}
	var result []JSONExitCode
	for _, raw := range codes {
		if m := jsonExitCodeRegex.FindStringSubmatch(raw); m != nil {
			result = append(result, JSONExitCode{
				Code:        m[1],
				Description: m[2],
			})
		}
	}
	return result
}

func toJSONDocument(doc *Document) *JSONDocument {
	jd := &JSONDocument{
		Name:        doc.FileTitle,
		Brief:       doc.FileBrief,
		Description: doc.FileDescription,
		Authors:     doc.Authors,
		License:     doc.License,
		Version:     doc.Version,
	}

	for i := range doc.Functions {
		f := &doc.Functions[i]
		jf := JSONFunc{
			Name:               f.Name,
			Section:            f.Section,
			SectionDescription: f.SectionDesc,
			Description:        f.Description,
			Deprecated:         f.Deprecated,
			Example:            f.Example,
			Options:            f.Options,
			Args:               parseArgs(f.Args),
			NoArgs:             f.NoArgs,
			Set:                parseSets(f.Sets),
			Env:                parseSets(f.Env),
			ExitCodes:          parseExitCodes(f.ExitCodes),
			Stdin:              f.Stdin,
			Stdout:             f.Stdout,
			Stderr:             f.Stderr,
			See:                f.See,
			Warnings:           f.Warnings,
		}
		jd.Functions = append(jd.Functions, jf)
	}

	return jd
}

// RenderJSON produces JSON output for the parsed document.
func (p *Parser) RenderJSON() (string, error) {
	jd := toJSONDocument(&p.doc)
	data, err := json.MarshalIndent(jd, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}
