# VSCode Extension Manual Testing Checklist

## Setup

```bash
# Rebuild LSP binary
go build -o ~/.local/share/go/bin/shdoc-lsp ./cmd/shdoc-lsp

# Compile extension
cd editors/vscode && npm run compile

# Launch dev host
code --extensionDevelopmentPath=/Users/jdevera/devel/contributions/shdoc-ng/editors/vscode
```

**After rebuilding `shdoc-lsp`, restart the dev host** (Cmd+Shift+P -> "Reload Window" or close/reopen).

Open a `.sh` file (e.g. `examples/showcase.sh`) to test.

---

## 1. Syntax Highlighting

Use the doc blocks in `examples/showcase.sh` to verify.

### Tag keywords
- [ ] All `@tag` names highlighted as keywords: `@description`, `@arg`, `@option`, `@example`, `@exitcode`, `@see`, `@warning`, `@deprecated`, `@noargs`, `@internal`, `@name`, `@brief`, `@file`, `@author`, `@license`, `@version`, `@section`, `@set`, `@env`, `@stdin`, `@stdout`, `@stderr`

### @arg lines
- [ ] `$1`, `$2`, `$3` highlighted as parameter
- [ ] `$@` highlighted as parameter
- [ ] Type (`string`, `int`, `bool`) highlighted as type
- [ ] Description text is NOT specially highlighted

### @option lines
- [ ] Short flags (`-u`, `-n`, `-f`, `-o`, `-r`) highlighted as flag
- [ ] Long flags (`--uppercase`, `--repeat`, `--reverse`, `--dry-run`, `--output`, `--force`, `--timeout`) highlighted as flag
- [ ] Flags after `/` separator (`-r/--reverse`) — both highlighted
- [ ] Flags after `|` separator (`-u | --uppercase`) — both highlighted
- [ ] `<count>`, `<path>`, `<seconds>` highlighted as placeholder variable
- [ ] `=<count>` form (`--repeat=<count>`) — placeholder highlighted
- [ ] Description text is NOT specially highlighted

### @set / @env lines
- [ ] Variable name (`LAST_GREETED`, `RESULT_CODE`, `HOME`, `LANG`) highlighted differently
- [ ] Type (`string`, `int`) highlighted as type

### @exitcode lines
- [ ] Exit code numbers (`0`, `1`, `2`, `127`) highlighted as number

### Negative cases
- [ ] `$1` is NOT highlighted outside `@arg` lines (e.g. in `@stdin`, `@description`)
- [ ] `--flag` style text is NOT highlighted outside `@option` lines
- [ ] `-ng` in words like `shdoc-ng` is NOT highlighted as a flag
- [ ] Plain description text stays comment-colored

## 2. Hover

- [ ] Hover over function declaration line (e.g. `greet() {`) -> rendered markdown preview
- [ ] Preview shows: function name heading, description, args, options, exit codes, etc.
- [ ] Preview does NOT show "Index" or TOC
- [ ] Preview does NOT show section headers from the file
- [ ] Hovering over comment lines inside the doc block does NOT show hover
- [ ] Hover on a `# @see old_parse()` line -> shows **old_parse's** docs (not the containing function's)
- [ ] Hover on a `@see` referencing a non-existent function -> no hover

## 3. Diagnostics (Warnings)

All of these must be inside a doc block (comment block directly above a `funcname() {` line).

### Empty value warnings
- [ ] `# @author` (empty) -> warning: "empty @author value"
- [ ] `# @arg` (empty) -> warning about empty/invalid arg
- [ ] `# @option` (empty) -> warning about empty option
- [ ] `# @description` with no text on same or continuation lines -> warning
- [ ] `# @example` with no indented continuation lines -> warning
- [ ] `# @version` (empty) -> warning
- [ ] `# @see` (empty) -> warning
- [ ] `# @set` (empty or invalid format) -> warning
- [ ] `# @env` (empty or invalid format) -> warning
- [ ] `# @exitcode` (empty or invalid format) -> warning
- [ ] `# @stdin` / `# @stdout` / `# @stderr` with no continuation -> warning
- [ ] `# @warning` (empty) -> warning
- [ ] Valid tags with values -> NO warning

### Deprecated strikethrough
- [ ] Function with `@deprecated` -> declaration line shows strikethrough
- [ ] Deprecated with message -> hint says "deprecated: <message>"
- [ ] Deprecated without message -> hint says "deprecated"
- [ ] Non-deprecated functions -> no strikethrough

### @noargs vs positional params
- [ ] `@noargs` function using `$1`, `$@`, `$#` in body -> warning on each usage
- [ ] `@noargs` function using `$1` inside single quotes (e.g. `awk '{print $1}'`) -> NO warning
- [ ] Functions without `@noargs` using `$1` -> NO warning

## 4. Completion

- [ ] On a `# ` line inside a doc block, type `@` -> completion list appears with all tags
- [ ] Type `@au` -> filters to `@author`
- [ ] Type `@desc` -> filters to `@description`
- [ ] Selecting a completion does NOT produce double `@@`
- [ ] Completing on empty `# ` line (no `@` yet) -> also offers tags
- [ ] Tags from both meta (`@name`, `@author`, etc.) and func (`@arg`, `@option`, etc.) categories appear

## 5. Document Symbols (Cmd+Shift+O)

- [ ] Opens symbol list showing documented function names
- [ ] Each entry has a "function" icon
- [ ] Selecting a symbol jumps to the doc block start

## 6. Go to Definition (F12 / Cmd+Click)

- [ ] On a `# @see other_func()` line, F12 or Cmd+click -> jumps to `other_func`'s doc block
- [ ] Works only when target function exists and is documented in the same file
- [ ] On non-`@see` lines -> no action

## 7. Folding

- [ ] Multi-line doc blocks have a fold arrow in the gutter
- [ ] Clicking the fold arrow collapses the comment block
- [ ] Single-line doc blocks do NOT show a fold arrow

## 8. Code Action (Lightbulb)

- [ ] Write an undocumented function: `myfunc() {` with no comment block above
- [ ] Place cursor on that line -> lightbulb appears (or Cmd+.)
- [ ] Action: "Insert shdoc-ng doc block for myfunc"
- [ ] Selecting it inserts a skeleton doc block above the function
- [ ] Already-documented functions do NOT get this action

## 9. Snippets

- [ ] `@desc` -> expands to `@description` + continuation line (no extra `#` prefix)
- [ ] `@arg` -> expands to `@arg $1 string Description.`
- [ ] `@opt` -> expands to `@option --flag Description.`
- [ ] `@exit` -> expands to `@exitcode 0 Description.`
- [ ] `@ex` -> expands to `@example` + continuation line
- [ ] `@see` -> expands to `@see`
- [ ] `@warn` -> expands to `@warning`
- [ ] `@dep` -> expands to `@deprecated`
- [ ] `docblock` -> full doc block skeleton WITH `#` prefixes (starts a new block)
- [ ] None of the `@`-triggered snippets add an extra `#` (user is already in a comment)

## 10. Documentation Preview

- [ ] Command palette: "shdoc-ng: Preview Documentation" -> opens markdown preview to the side
- [ ] Preview icon appears in editor title bar for `.sh` files
- [ ] Preview shows rendered documentation (TOC, function docs, etc.)
- [ ] Editing the source file updates the preview automatically
- [ ] Note: scroll sync between source and preview is not supported (generated content)
