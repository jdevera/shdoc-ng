package main

// Document holds the top-level file metadata and accumulated output.
type Document struct {
	FileTitle       string
	FileBrief       string
	FileDescription string
	TOC             []string // rendered TOC items (e.g. "* [name](#slug)")
	DocStr          string   // accumulated rendered function docs
}

// FuncDoc holds the parsed documentation for a single function.
type FuncDoc struct {
	Name        string
	Description string
	Example     string
	Options     []OptionEntry
	BadOptions  []string
	Args        map[string]string // zero-padded key ("001", "@") → raw "$N type desc"
	NoArgs      bool
	Sets        []string
	ExitCodes   []string
	Stdin       []string
	Stdout      []string
	Stderr      []string
	See        []string
	Deprecated string
}

// OptionEntry represents a valid @option with its rendered term and description.
type OptionEntry struct {
	Term       string // e.g. "-v <val> | --value <val>"
	Definition string
}
