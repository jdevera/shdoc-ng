# shdoc-ng

[![CI](https://github.com/jdevera/shdoc-ng/actions/workflows/ci.yml/badge.svg)](https://github.com/jdevera/shdoc-ng/actions/workflows/ci.yml)
[![Release](https://github.com/jdevera/shdoc-ng/actions/workflows/release.yml/badge.svg)](https://github.com/jdevera/shdoc-ng/actions/workflows/release.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/jdevera/shdoc-ng/shdoc.svg)](https://pkg.go.dev/github.com/jdevera/shdoc-ng/shdoc)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Latest Release](https://img.shields.io/github/v/release/jdevera/shdoc-ng)](https://github.com/jdevera/shdoc-ng/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/jdevera/shdoc-ng)](https://goreportcard.com/report/github.com/jdevera/shdoc-ng)
[![Agent Skill](https://img.shields.io/badge/Agent_Skill-shdoc--ng-8A2BE2)](https://github.com/jdevera/shdoc-ng/tree/main/skills/shdoc-ng)

A documentation generator for shell scripts. Reads annotated comments above functions and produces **Markdown**, **HTML**, or **JSON** documentation.

shdoc-ng is a Go reimplementation of [shdoc](https://github.com/reconquest/shdoc) with additional features: multiple output formats, an LSP server, editor integrations, a linting mode, custom templates, and structured JSON output with schema.

## Installation

```bash
# Homebrew
brew install jdevera/tap/shdoc-ng

# Go
go install github.com/jdevera/shdoc-ng/cmd/shdoc-ng@latest

# Or download a binary from the releases page
# https://github.com/jdevera/shdoc-ng/releases
# Linux packages (deb/rpm/apk) are also available there.
```

## Quick Start

Annotate your shell functions with `@tag` comments:

```bash
#!/bin/bash
# @name mylib
# @brief A library for greeting people.

# @description Greet someone by name.
#
# Prints a friendly greeting to stdout.
#
# @example
#   greet "World"
#
# @arg $1 string The name to greet.
# @exitcode 0 If successful.
# @exitcode 1 If no name was provided.
# @stdout The greeting string.
greet() {
    echo "Hello, $1!"
}
```

Generate documentation:

```bash
# Markdown (default)
shdoc-ng generate -i script.sh -o docs.md

# HTML with Catppuccin theme
shdoc-ng generate --format html -i script.sh -o docs.html

# JSON (machine-readable)
shdoc-ng generate --format json -i script.sh

# Pipe-friendly
shdoc-ng generate < script.sh > docs.md
```

The HTML output uses the [Catppuccin](https://catppuccin.com) color palette.

## Supported Tags

### File-level tags

| Tag | Description |
|-----|-------------|
| `@name` / `@file` | Script title |
| `@brief` | One-line summary |
| `@description` | Extended description (multi-line) |
| `@author` | Author name (repeatable) |
| `@license` | License identifier |
| `@version` | Version string |

### Function-level tags

| Tag | Description |
|-----|-------------|
| `@description` | Function description (multi-line) |
| `@example` | Usage example (code block on following lines) |
| `@arg` | Positional argument: `@arg $1 type Description` |
| `@option` | Option flag: `@option -f \| --flag Description` |
| `@noargs` | Declares the function takes no arguments |
| `@set` | Variable set by the function: `@set VAR type Description` |
| `@env` | Environment variable used: `@env VAR type Description` |
| `@exitcode` | Exit code: `@exitcode 0 Description`. Supports prefixes: `>N` (greater than), `!N` (not equal) |
| `@stdin` | Description of stdin usage |
| `@stdout` | Description of stdout output |
| `@stderr` | Description of stderr output |
| `@see` | Cross-reference (function, URL, path, or markdown link) |
| `@section` | Group following functions under a section heading |
| `@internal` | Exclude function from output |
| `@deprecated` | Mark as deprecated, with optional message |
| `@warning` / `@warn` | Usage warning |
| `@label` | Freeform labels for categorization (comma-separated) |

### Tag shorthands

`@desc` = `@description`, `@exit` = `@exitcode`, `@opt` = `@option`, `@warn` = `@warning`

## Commands

```
shdoc-ng generate    Generate documentation from a shell script
shdoc-ng check       Validate annotations (exit 1 if warnings found)
shdoc-ng template    Print a built-in template (for customisation)
shdoc-ng json-schema Print the JSON Schema for the JSON output format
shdoc-ng lsp         Run the LSP server (stdio)
```

### generate

```
shdoc-ng generate [flags]

Flags:
  -i, --input string      Input file (default stdin)
  -o, --output string     Output file (default stdout)
      --format string     Output format: markdown, html, json (default "markdown")
      --template string   Use a custom template file
```

### check

```bash
shdoc-ng check -i script.sh
# Prints warnings to stderr in file:line:col format
# Exits with code 1 if any warnings are found
```

The output format (`file:line:col: warning: message`) integrates with editors and CI tools.

## Custom Templates

Extract a built-in template and modify it:

```bash
shdoc-ng template markdown > my-template.tmpl
shdoc-ng generate --template my-template.tmpl -i script.sh
```

## LSP Server

shdoc-ng includes a Language Server Protocol server that provides:

- **Diagnostics** -- empty tag warnings, invalid formats, `@noargs` conflicts
- **Hover** -- rendered documentation preview on function declarations
- **Completion** -- `@tag` completions inside comment blocks
- **Go to Definition** -- `@see funcName()` resolves to the target function
- **Document Symbols** -- functions and sections in the outline
- **Folding Ranges** -- fold comment blocks
- **Code Actions** -- insert doc block skeleton for undocumented functions

### Editor Setup

**VSCode:** Install the extension from `editors/vscode/`. Set `shdoc-ng.serverCommand` to the path of the `shdoc-ng` binary.

**Neovim:**

```lua
require('shdoc-ng').setup({
    cmd = 'shdoc-ng',
    args = { 'lsp' },
})
```

## Differences from shdoc

shdoc-ng is compatible with the original shdoc annotation format but intentionally deviates in some areas:

- **Sections are sticky** -- all functions after `@section Foo` belong to that section until the next `@section` (original applied only to the next function)
- **Section headers appear once** -- the original repeated the section heading for every function in the section
- **No description routing** -- a function's `@description` only sets the function's description, not the section or file description

See [DEVELOPMENT.md](DEVELOPMENT.md) for the full list of known deviations.

## Pre-commit Hook

shdoc-ng can be used as a [pre-commit](https://pre-commit.com/) hook to check shell script annotations:

```yaml
repos:
  - repo: https://github.com/jdevera/shdoc-ng
    rev: v0.4.0  # check https://github.com/jdevera/shdoc-ng/releases for latest
    hooks:
      - id: shdoc-ng-check
```

## Building from Source

```bash
go build ./cmd/shdoc-ng
go test ./...
```

## License

[MIT](LICENSE) -- based on [shdoc](https://github.com/reconquest/shdoc) by Stanislav Seletskiy & Egor Kovetskiy.
