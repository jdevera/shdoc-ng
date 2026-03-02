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
	Options     []OptionEntry `json:"options,omitempty"`
	BadOptions  []string      `json:"-"`
	Args        []Arg         `json:"args,omitempty"`
	NoArgs      bool          `json:"noargs,omitempty"`
	Sets        []SetVar      `json:"set,omitempty"`
	Env         []SetVar      `json:"env,omitempty"`
	ExitCodes   []ExitCode    `json:"exitcodes,omitempty"`
	Stdin       []string          `json:"stdin,omitempty"`
	Stdout      []string          `json:"stdout,omitempty"`
	Stderr      []string          `json:"stderr,omitempty"`
	See         []SeeRef          `json:"see,omitempty"`
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

// OptionForm represents one form of a command-line option (e.g., "-n" or "--repeat").
// If Value is present and ValueSep is absent, the value is directly adjacent (e.g., "-n<count>").
type OptionForm struct {
	Name     string `json:"name"`
	Value    string `json:"value,omitempty"`
	ValueSep string `json:"value_sep,omitempty"`
}

// OptionEntry represents a valid @option with its parsed forms and description.
type OptionEntry struct {
	Forms      []OptionForm `json:"forms"`
	Definition string       `json:"description"`
}

// SeeRef represents a parsed @see reference.
// Kind is one of: "ref" (anchor), "url", "path", "link" (markdown link), "text" (mixed content).
type SeeRef struct {
	Kind string `json:"kind"`
	Text string `json:"text,omitempty"`
	Href string `json:"href,omitempty"`
}

// Arg represents a parsed @arg entry.
type Arg struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// SetVar represents a parsed @set or @env entry.
type SetVar struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ExitCode represents a parsed @exitcode entry.
type ExitCode struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}
