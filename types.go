package main

// Document holds the top-level file metadata and stored function docs.
type Document struct {
	FileTitle       string    `json:"name,omitempty"`
	FileBrief       string    `json:"brief,omitempty"`
	FileDescription string    `json:"description,omitempty"`
	Authors         []string  `json:"authors,omitempty"`
	License         string    `json:"license,omitempty"`
	Version         string    `json:"version,omitempty"`
	Functions       []FuncDoc `json:"functions,omitempty"`
}

// FuncDoc holds the parsed documentation for a single function.
type FuncDoc struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Example     string            `json:"example,omitempty"`
	Options     []OptionEntry     `json:"options,omitempty"`
	BadOptions  []string          `json:"-"`
	Args        map[string]string `json:"-"`
	NoArgs      bool              `json:"noargs,omitempty"`
	Sets        []string          `json:"-"`
	Env         []string          `json:"-"`
	ExitCodes   []string          `json:"-"`
	Stdin       []string          `json:"stdin,omitempty"`
	Stdout      []string          `json:"stdout,omitempty"`
	Stderr      []string          `json:"stderr,omitempty"`
	See         []string          `json:"see,omitempty"`
	Warnings    []string          `json:"warnings,omitempty"`
	Deprecated  string            `json:"deprecated,omitempty"`
	Section     string            `json:"section,omitempty"`
	SectionDesc string            `json:"section_description,omitempty"`
}

// hasDocumentation returns true if the FuncDoc has any documentation content.
func (f *FuncDoc) hasDocumentation() bool {
	return len(f.Options) > 0 || len(f.BadOptions) > 0 ||
		len(f.Args) > 0 || f.NoArgs ||
		len(f.Sets) > 0 || len(f.Env) > 0 ||
		len(f.ExitCodes) > 0 ||
		len(f.Stdin) > 0 || len(f.Stdout) > 0 ||
		len(f.Stderr) > 0 || len(f.See) > 0 ||
		len(f.Warnings) > 0 ||
		f.Deprecated != "" || f.Example != ""
}

// OptionEntry represents a valid @option with its rendered term and description.
type OptionEntry struct {
	Term       string `json:"term"`
	Definition string `json:"description"`
}
