# Development Guide

## Project Structure

```
shdoc-ng/
├── *.go                    # Core library (package shdoc)
├── cmd/
│   ├── shdoc-ng/           # CLI tool
│   └── shdoc-lsp/          # LSP server for editor integration
├── templates/
│   ├── markdown.tmpl       # Default Markdown output template
│   └── html.tmpl           # HTML output template (Catppuccin theme)
├── editors/
│   ├── vscode/             # VSCode extension (syntax, snippets, LSP client, preview)
│   └── neovim/             # Neovim config (LSP, snippets, syntax)
├── testdata/               # Conformance test suite
├── examples/
│   └── showcase.sh         # Comprehensive example exercising all features
└── docs/                   # Design docs and checklists
```

## Build & Test

```bash
go build -o shdoc-ng ./cmd/shdoc-ng       # CLI binary
go build -o shdoc-lsp ./cmd/shdoc-lsp     # LSP binary
go test ./...                               # all tests
go test -run TestConformance/option        # single conformance case
go test -v -run TestLegacyConformance      # legacy awk conformance
```

## Architecture

The root package is `package shdoc` (a library). CLIs import it.

**Pipeline:** input → Lexer → Segmenter → Block Parser → Renderer → output

- **`lexer.go`** — `LexLines()` classifies each line as comment, code, or blank.
- **`segmenter.go`** — `SegmentBlocks()` groups consecutive comment lines into blocks, identifies which precede function declarations.
- **`blockparser.go`** — `ParseDocument()` processes blocks into a `Document` with `FuncDoc` entries. Handles all `@tag` parsing, validation, and warnings.
- **`template.go`** — `RenderWithTemplate()` renders a `Document` using Go `text/template`.
- **`templates/`** — Embedded template files for Markdown and HTML output.
- **`types.go`** — Data structs: `Document`, `FuncDoc`, `OptionEntry`, `Arg`, `SetVar`, etc.
- **`option.go`** — `@option` format validation and rendering.
- **`slug.go`** — GitHub-compatible anchor generation, `@see` link rendering.

## Tag Shorthands

These shorthands are normalized to their full form in `ParseTag()`:

| Shorthand | Full form      |
|-----------|----------------|
| `@desc`   | `@description` |
| `@exit`   | `@exitcode`    |
| `@opt`    | `@option`      |
| `@warn`   | `@warning`     |

## Conformance Test Suite

Test cases live in `testdata/<name>/` with these files:

| File              | Purpose                                              |
|-------------------|------------------------------------------------------|
| `input.sh`        | Input shell script                                   |
| `expected.md`     | Expected output matching the original awk shdoc      |
| `expected-ng.md`  | (optional) Expected output for shdoc-ng when we intentionally deviate |
| `meta.json`       | (optional) Metadata: tags and known deviations       |

### How it works

- **`TestConformance`** — validates shdoc-ng behavior. Uses `expected-ng.md` when present, falls back to `expected.md`.
- **`TestLegacyConformance`** — validates original awk behavior. Always uses `expected.md`. Skips cases with `knownDeviations`.

### Adding a test case

Create `testdata/<name>/input.sh` and `testdata/<name>/expected.md`. It will be picked up automatically with the `compat` tag.

### Documenting intentional deviations

When shdoc-ng intentionally differs from the original awk:

1. Keep `expected.md` untouched (original awk output).
2. Create `expected-ng.md` with shdoc-ng's corrected output.
3. Add or update `meta.json` with `knownDeviations` explaining why:

```json
{
  "tags": ["compat"],
  "knownDeviations": [
    "Brief description of what changed and why"
  ]
}
```

## Known Deviations from Original shdoc

### Sections are sticky

In the original awk, `@section` only applied to the immediately following function. In shdoc-ng, all functions after `@section Foo` belong to that section until the next `@section` or end of file.

### No description routing from functions to sections

In the original awk, a function's `@description` could leak into the section description and/or the file-level overview description via a shared `handleDescription()` function. This was an artifact of the awk implementation's shared mutable state, not intentional design.

In shdoc-ng:
- A function's `@description` only sets the function's description.
- Section descriptions only come from `@description` in meta blocks (comment blocks not followed by a function).
- File-level overview (`FileDescription`) can still be populated from a function's description if no explicit file-level `@description` exists (for backwards compatibility with files that rely on this).

### Section headers are not repeated

In the original awk, the section heading was emitted for every function in the section. In shdoc-ng, `IsFirstInSection` is set on the first function in each section, and templates use this flag to emit the heading only once.

## Linting

The CLI supports a `--lint` flag for validation-only mode:

```bash
shdoc-ng --lint -i script.sh    # prints warnings to stderr, exits 1 if any
```

Output uses the standard `file:line:col: warning: message` format for tool integration.

## LSP Server

`cmd/shdoc-lsp/` provides:

- **Diagnostics** — empty tag warnings, invalid formats, `@noargs` + positional param conflicts
- **Hover** — rendered doc preview on function declaration lines, `@see` reference preview
- **Completion** — `@tag` completions inside comment blocks
- **Go to Definition** — `@see funcName()` resolves to target's doc block
- **Document Symbols** — file title, sections (as namespace), functions (as string) hierarchy
- **Folding** — comment block folding ranges
- **Code Actions** — insert doc block skeleton for undocumented functions
- **Deprecated** — strikethrough on `@deprecated` function declarations

## VSCode Extension

`editors/vscode/` provides:

- TextMate grammar injection for `@tag` syntax highlighting
- Snippets for common annotations
- LSP client connecting to `shdoc-lsp` via stdio
- Documentation preview command (renders to Markdown, opens in built-in preview)

### Testing the extension

See `docs/vscode-testing-checklist.md` for the full manual testing checklist.

```bash
# Build
go build -o ~/.local/share/go/bin/shdoc-lsp ./cmd/shdoc-lsp
go build -o ~/.local/share/go/bin/shdoc-ng ./cmd/shdoc-ng
cd editors/vscode && npm run compile

# Symlink binaries (once)
ln -s ~/.local/share/go/bin/shdoc-lsp ~/.local/bin/shdoc-lsp
ln -s ~/.local/share/go/bin/shdoc-ng ~/.local/bin/shdoc-ng

# Launch dev host
code --extensionDevelopmentPath=editors/vscode

# After rebuilding shdoc-lsp, reload the dev host window
```
