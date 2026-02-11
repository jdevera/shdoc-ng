# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

shdoc-ng is a Go reimplementation of [shdoc](https://github.com/reconquest/shdoc), a shell documentation generator. It reads annotated shell scripts from stdin and produces Markdown on stdout. Zero external dependencies.

## Build & Test Commands

```bash
go build -o shdoc-ng .     # build binary
go test ./...               # run all tests (includes conformance)
go test -run TestConformance/option  # run a single conformance case
```

Usage: `./shdoc-ng < script.sh > output.md`

## Architecture

The pipeline is: **stdin → Parser (line-by-line state machine) → Renderer → stdout**

- **`parser.go`** (~400 lines) — The core. A line-by-line state machine in `ProcessLine()` that matches awk rule order exactly. Rules are checked sequentially (not with early returns everywhere — some rules intentionally fall through to later ones). State variables track whether we're inside a `@description`, `@example`, or multi-line `@stdin/@stdout/@stderr` block.
- **`render.go`** — Assembles Markdown output. `renderFuncDoc()` produces one function's documentation; `renderDocument()` joins the file header, TOC, and all function docs. Also contains `unindent()` for `@example` blocks.
- **`slug.go`** — GitHub-compatible anchor generation (`slug()`) and `@see`/TOC link rendering (`renderTocLink()`). Handles bare URLs, markdown links, relative/absolute paths.
- **`option.go`** — Validates `@option` format (short/long flags with optional values, pipe-separated) and renders terms with bold/escaped angle brackets.
- **`types.go`** — Data structs: `Document` (file-level), `FuncDoc` (per-function docblock), `OptionEntry`.
- **`main.go`** — Thin CLI wrapper.

## Conformance Tests

Tests live in `testdata/*/input.sh` + `expected.md` (18 cases). `shdoc_test.go` feeds each input through the parser and compares output **byte-for-byte** against expected. These cases were extracted from the original shdoc test suite.

To add a test case: create `testdata/<name>/input.sh` and `testdata/<name>/expected.md`. It will be picked up automatically.

## Key Implementation Details

- **Rule order in `parser.go` must match the awk original exactly.** The `@description` tag rule intentionally does NOT return — it falls through to the `inDescription` continuation block on the same line. Changing rule order will break conformance.
- **`handleDescription()` does not clear `description`** — it routes to `fileDescription` or `sectionDesc` but only `reset()` clears the field. This means description stays available for `processFunction()` to consume.
- **`@arg` with invalid format falls through to `@option` processing** — it's re-processed as an option entry, not just warned about.
- **`unindent()` uses `-1` as sentinel for `start`** because Go arrays are 0-indexed (the awk original used 0 since `split()` produces 1-indexed arrays).

## Reference Implementation

`shdoc-awk/` (if present) contains the original gawk shdoc with its `.git` history. It is gitignored and not required for building or testing.
