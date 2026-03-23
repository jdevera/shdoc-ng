# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

shdoc-ng is a Go reimplementation of [shdoc](https://github.com/reconquest/shdoc), a shell documentation generator. It reads annotated shell scripts and produces Markdown, HTML, or JSON documentation. The library is `package shdoc` in `shdoc/`; CLIs live under `cmd/`.

## Build & Test Commands

```bash
go build ./cmd/shdoc-ng                # CLI binary (includes LSP server)
go test ./...                           # all tests (conformance + unit)
go test -run TestConformance/option    # single conformance case
```

Usage: `shdoc-ng generate -i script.sh -o output.md` or `shdoc-ng generate < script.sh > output.md`

## Architecture

**Pipeline:** input → `LexLines()` → `SegmentBlocks()` → `ParseDocument()` → `RenderWithTemplate()` → output

- **`shdoc/`** — Library package (`package shdoc`), importable as `github.com/jdevera/shdoc-ng/shdoc`.
  - **`lexer.go`** — Line classification (comment, code, blank).
  - **`segmenter.go`** — Groups comment lines into blocks, identifies function declarations.
  - **`blockparser.go`** — Core parser. Processes blocks into `Document`/`FuncDoc`. All `@tag` parsing, validation, warnings. Tag shorthands (`@desc`→`@description`, etc.) normalized in `ParseTag()`.
  - **`template.go`** — Template rendering with `text/template`. Function map for slugs, markdown helpers.
  - **`templates/`** — Embedded Markdown and HTML (Catppuccin) templates.
  - **`types.go`** — Data structs: `Document`, `FuncDoc`, `OptionEntry`, `Arg`, `SetVar`, etc.
  - **`option.go`** — `@option` format validation and rendering.
  - **`slug.go`** — GitHub-compatible anchors, `@see` link rendering.
  - **`testdata/`** — Conformance test suite.
- **`cmd/shdoc-ng/`** — Cobra CLI with subcommands: `generate`, `check`, `template`, `schema`, `lsp`.
- **`internal/lsp/`** — LSP server (diagnostics, hover, completion, go-to-def, symbols, folding, code actions).
- **`editors/vscode/`** — VSCode extension (syntax highlighting, snippets, LSP client, doc preview).
- **`editors/neovim/`** — Neovim integration (LSP, snippets, syntax).
- **`skills/shdoc-ng/`** — Agent Skills spec skill for documenting shell scripts.

## Conformance Tests

Tests in `shdoc/testdata/*/input.sh` + `expected.md`. Two test modes:

- **`TestConformance`** — validates shdoc-ng behavior. Uses `expected-ng.md` when present (intentional deviations), falls back to `expected.md`.
- **`TestLegacyConformance`** — validates original awk behavior. Always uses `expected.md`, skips cases with `knownDeviations` in `meta.json`.

See `DEVELOPMENT.md` for full details on the test framework and documenting deviations.

## Key Implementation Details

- **Sections are nested** — `Document.Sections []Section`, each containing `Name`, `Description`, and `Functions []FuncDoc`. Functions before any `@section` go in an unnamed section. Sections with no functions are pruned after parsing.
- **Sections are sticky** — all functions after `@section Foo` belong to that section until the next `@section`. This deviates from the original awk (which only applied to the next function).
- **`@arg` with invalid format falls through to `@option` processing** — preserving original awk behaviour.
- **`AllFunctions()`** on `Document` — returns a flat list of all functions across all sections, for code that needs linear iteration.
- **Function descriptions don't route to section descriptions** — unlike the original awk. See `DEVELOPMENT.md` for known deviations.
- **Double description routing** — `routeDescription()` does NOT return early after setting the section description — it falls through to also try `FileDescription`. This mirrors the old parser's dual-call pattern.
- **Function description → FileDescription** — At the end of `parseFuncBlock`, the description is routed via `routeDescription()` BEFORE being set on `docblock.Description`, so the first function's description can populate `doc.FileDescription` if still empty.
- **`@example` continuation** — Uses `^[\s]*#[ ]+` (one or more spaces after `#`). A bare `#` (no space) ends the example. Implemented in `collectExampleLines()`.
- **`@description` collection** — `collectUntilNextTag()` stops at `@`-tagged lines but continues through bare `#` blank comment lines.
- **`@label` parsing** — Comma-separated values on a single line, split and trimmed. Multiple `@label` lines accumulate. Stored as `Labels []string` on `FuncDoc`.

## Reference Implementation

The original gawk shdoc at [reconquest/shdoc](https://github.com/reconquest/shdoc) is the reference implementation. It is not required for building or testing.
