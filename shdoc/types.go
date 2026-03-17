package shdoc

// Document holds the top-level file metadata and stored function docs.
type Document struct {
	FileTitle       string    `json:"name,omitempty"        desc:"Title or name of the shell script (@file)"`
	FileBrief       string    `json:"brief,omitempty"       desc:"One-line summary of the script (@brief)"`
	FileDescription string    `json:"description,omitempty" desc:"Extended description of the script (@description)"`
	Authors         []string  `json:"authors,omitempty"     desc:"List of authors (@author)"`
	License         string    `json:"license,omitempty"     desc:"License identifier (@license)"`
	Version         string    `json:"version,omitempty"     desc:"Version string (@version)"`
	Functions       []FuncDoc `json:"functions,omitempty"   desc:"Documented functions found in the script"`
}

// FuncDoc holds the parsed documentation for a single function.
type FuncDoc struct {
	Name        string        `json:"name"                        desc:"Function name as it appears in the source"`
	Description string        `json:"description,omitempty"       desc:"Description of the function (@description)"`
	Example     string        `json:"example,omitempty"           desc:"Example usage (@example)"`
	Options     []OptionEntry `json:"options,omitempty"           desc:"Command-line options the function accepts (@option)"`
	BadOptions  []string      `json:"-"`
	Args        []Arg         `json:"args,omitempty"              desc:"Positional arguments the function accepts (@arg)"`
	IsNoArgs    bool          `json:"is_noargs,omitempty"         desc:"True if the function explicitly takes no arguments (@noargs)"`
	Sets        []SetVar      `json:"set,omitempty"               desc:"Variables set by the function (@set)"`
	Env         []SetVar      `json:"env,omitempty"               desc:"Environment variables used or modified (@env)"`
	ExitCodes   []ExitCode    `json:"exitcodes,omitempty"         desc:"Exit codes and their meanings (@exitcode)"`
	Stdin       []string      `json:"stdin,omitempty"             desc:"Description of stdin usage (@stdin)"`
	Stdout      []string      `json:"stdout,omitempty"            desc:"Description of stdout output (@stdout)"`
	Stderr      []string      `json:"stderr,omitempty"            desc:"Description of stderr output (@stderr)"`
	See         []SeeRef      `json:"see,omitempty"               desc:"See also references (@see)"`
	Warnings    []string      `json:"warnings,omitempty"          desc:"Warnings about usage or behavior (@warning)"`
	IsDeprecated    bool          `json:"is_deprecated,omitempty"      desc:"True if the function is deprecated (@deprecated)"`
	DeprecatedMessage string      `json:"deprecated_message,omitempty" desc:"Deprecation notice, if provided (@deprecated message)"`
	Section          string `json:"section,omitempty"              desc:"Section the function belongs to (@section)"`
	SectionDesc      string `json:"section_description,omitempty"  desc:"Description of the section"`
	IsFirstInSection bool   `json:"is_first_in_section,omitempty"  desc:"True if this is the first function in its section"`
}

// hasDocumentation returns true if the FuncDoc has any documentation content.
func (f *FuncDoc) hasDocumentation() bool {
	return len(f.Options) > 0 || len(f.BadOptions) > 0 ||
		len(f.Args) > 0 || f.IsNoArgs ||
		len(f.Sets) > 0 || len(f.Env) > 0 ||
		len(f.ExitCodes) > 0 ||
		len(f.Stdin) > 0 || len(f.Stdout) > 0 ||
		len(f.Stderr) > 0 || len(f.See) > 0 ||
		len(f.Warnings) > 0 ||
		f.IsDeprecated || f.Example != ""
}

// OptionForm represents one form of a command-line option (e.g., "-n" or "--repeat").
// If Value is present and ValueSep is absent, the value is directly adjacent (e.g., "-n<count>").
type OptionForm struct {
	Name     string `json:"name"               desc:"The flag name (e.g., \"-n\" or \"--repeat\")"`
	Value    string `json:"value,omitempty"    desc:"Placeholder for the option's value (e.g., \"count\")"`
	ValueSep string `json:"value_sep,omitempty" desc:"Separator between flag and value (\"=\" or \" \"; absent means value is adjacent)"`
}

// OptionEntry represents a valid @option with its parsed forms and description.
type OptionEntry struct {
	Forms      []OptionForm `json:"forms"       desc:"All forms of the option (e.g., short and long flags)"`
	Definition string       `json:"description" desc:"Description of what the option does"`
}

// SeeRef represents a parsed @see reference.
// Kind is one of: "ref" (anchor), "url", "path", "link" (markdown link), "text" (mixed content).
type SeeRef struct {
	Kind string `json:"kind"           desc:"Reference type: ref (anchor), url, path, link (markdown link), or text"`
	Text string `json:"text,omitempty" desc:"Display text for the reference"`
	Href string `json:"href,omitempty" desc:"URL or path target for the reference"`
}

// Arg represents a parsed @arg entry.
type Arg struct {
	Name        string `json:"name"        desc:"Argument name"`
	Type        string `json:"type"        desc:"Argument type (e.g., String, Number)"`
	Description string `json:"description" desc:"Description of the argument"`
}

// SetVar represents a parsed @set or @env entry.
type SetVar struct {
	Name        string `json:"name"        desc:"Variable name"`
	Type        string `json:"type"        desc:"Variable type"`
	Description string `json:"description" desc:"Description of the variable"`
}

// ExitCode represents a parsed @exitcode entry.
type ExitCode struct {
	Code        string `json:"code"        desc:"Exit code value (e.g., 0, 1, or a signal name)"`
	Description string `json:"description" desc:"Meaning of this exit code"`
}
