# shdoc-ng: Comprehensive Review & Roadmap to Published Tool

## 1. High-Level Overview

**shdoc-ng** is a Go reimplementation of [shdoc](https://github.com/reconquest/shdoc) (338 stars, last active Oct 2023, effectively abandoned). It parses annotated shell scripts and generates Markdown, HTML, or JSON documentation. Beyond the original's scope, it adds an LSP server, editor integrations (VSCode + Neovim), a linting mode, custom templates, and structured JSON output with schema.

**Pipeline:** `input → LexLines() → SegmentBlocks() → ParseDocument() → RenderWithTemplate() → output`

**Stats:** ~3,150 lines of Go across 22 files, 37 conformance test cases, all tests passing.

---

## 2. Competitive Landscape

shdoc-ng is **unique in this space** — no other shell documentation tool offers LSP support, multiple output formats, editor integrations, and linting combined.

| Tool | Lang | Stars | Last Active | LSP | Multi-format | Linting |
|------|------|-------|-------------|-----|--------------|---------|
| **shdoc** (original) | Awk | 338 | 2023-10 | No | Markdown only | No |
| **shdoc-ng** | Go | — | Active | Yes | MD/HTML/JSON | Yes |
| **bashdoc** (Rust) | Rust | 31 | 2024-12 | No | No | No |
| **tomdoc.sh** | Shell | 69 | 2022-03 | No | No | No |
| **shocco** | Perl | 207 | Archived | No | HTML only | No |
| **bash-doxygen** | sed | 96 | 2019-02 | No | Via Doxygen | No |
| **zsdoc** | Shell | 18 | 2025-11 | No | No | No |

**Note:** bash-language-server (2,646 stars) is complementary — it's a general bash LSP with no documentation annotation awareness.

Sources: GitHub repos for each tool ([reconquest/shdoc](https://github.com/reconquest/shdoc), [dustinknopoff/bashdoc](https://github.com/dustinknopoff/bashdoc), [tests-always-included/tomdoc.sh](https://github.com/tests-always-included/tomdoc.sh), [rtomayko/shocco](https://github.com/rtomayko/shocco), [Anvil/bash-doxygen](https://github.com/Anvil/bash-doxygen), [z-shell/zsdoc](https://github.com/z-shell/zsdoc))

---

## 3. Code Review Findings

### 3.1 Strengths

- **Clean pipeline architecture** — lexer/segmenter/parser/renderer separation is textbook good design
- **Excellent test framework** — dual conformance suite (shdoc-ng vs legacy awk), 37 test cases, deviation documentation via `meta.json`
- **Drop-in compatibility** — preserves original shdoc annotation format and awk quirks (like `@arg` fallthrough to `@option`)
- **Template-driven output** — users can customize without touching Go code
- **Full LSP implementation** — diagnostics, hover, completion, go-to-def, symbols, folding, code actions — impressive scope
- **Schema generation** — auto-generates JSON Schema from struct reflection with `desc` tags

### 3.2 Functional Issues

#### ~~F1. Module path is not a valid Go module path~~ DONE

~~`go.mod` declares `module shdoc-ng` — not importable as a library.~~

Fixed: module path is now `github.com/jdevera/shdoc-ng`. All imports updated. Build and tests pass.

#### F3. `BpStripRe` is exported but should be internal
The `BpStripRe` regex is exported (capital B) and used by the LSP, but this leaks parser internals. The LSP should use a public function instead.

#### F4. `processAtOption` compiles regex on every call
In `option.go:80`, `pipeRe := regexp.MustCompile(...)` is compiled inside `processAtOption()`. This should be a package-level `var` like the other regexes.

#### F5. LSP has no tests
`internal/lsp/` has 831 lines and zero test files. This is the most complex module in the project.

#### F6. `cmd/shdoc-ng/` has no tests
The CLI subcommands (generate, check, schema, template, lsp) have no test coverage at all.

#### F7. Sorting in generate doesn't match sorting in test
`generate.go` uses `sort.SliceStable` (preserving section order), but `shdoc_test.go:TestSortOutput` uses `sort.Slice` (unstable). The test doesn't exercise the actual CLI sorting logic.

#### F8. `unindent` counts max indent unnecessarily
`render.go` first finds `maxIndent`, then finds `indent` (minimum). The `maxIndent` pass is wasted work — it's computed but only used as the initial value for the minimum search.

#### ~~F9. No version information in binary~~ DONE
Added `version` var with ldflags injection. `shdoc-ng --version` works (`dev` locally, real version via GoReleaser).

#### ~~F10. Makefile build target was wrong~~ DONE
Removed. Go's toolchain and GoReleaser cover all build needs without a Makefile.

### 3.3 UX Issues

#### ~~U1. No `--version` flag~~ DONE
Added via Cobra's built-in `Version` field + ldflags injection.

#### U2. No `--quiet` / `--no-warnings` flag
Warnings always go to stderr. Users piping output may want to suppress them.

#### U3. `generate` is not the default command
Running `shdoc-ng < script.sh` does nothing useful — you must run `shdoc-ng generate < script.sh`. For a tool meant to be part of pipelines, the primary action should work without a subcommand. Consider making `generate` the default when no subcommand is given.

#### U4. No stdin auto-detection
If input is `-` (default) but stdin is a terminal (not a pipe), the tool blocks silently. Should print usage hint or error.

#### U5. Missing man page
Unix tools should ship a man page. GoReleaser can generate these.

#### U6. No shell completion
Cobra has built-in support for generating bash/zsh/fish completions. This is a one-line addition but significantly improves discoverability.

#### U7. `check` exit code semantics
`check` exits 1 on any warning. Some warnings are informational (e.g., duplicate tags). Consider `--strict` vs default behavior, or exit code 0 for warnings, 1 for errors.

#### U8. Error messages don't suggest `--help`
When users mistype a command or flag, error messages should hint at `--help`.

#### U9. No progress or summary in `check` mode
`check` either prints warnings or nothing. A summary line like "3 warnings in script.sh" or "script.sh: OK" would improve the experience.

#### U10. HTML output includes full CSS inline
The 34K HTML template embeds the entire Catppuccin theme. No option to reference an external CSS file or use a minimal theme.

### 3.4 Code Quality Observations

#### C1. `blockparser.go` is a 542-line switch statement
The `parseFuncBlock` method is a single function with a large `switch` on tag names. While readable, each case could be extracted to a method for testability and maintainability.

#### C2. LSP server is a single 831-line file
`internal/lsp/lsp.go` handles initialization, parsing, diagnostics, hover, completion, definition, symbols, folding, and code actions — all in one file. Should be split by capability.

#### C3. Global mutable state in LSP
`var store = map[string]*docState{}` with a `sync.RWMutex` — this works but makes testing hard. Consider a `Server` struct that holds state.

#### ~~C4. No godoc comments on exported functions~~ (partially addressed)
Most exported functions now have comments. Could be improved for pkg.go.dev rendering.

#### ~~C5. Import alias `shdoc "shdoc-ng"` is awkward~~ DONE
Fixed by F1 — module path is now `github.com/jdevera/shdoc-ng`. Import aliases are still used for brevity (`shdoc "github.com/jdevera/shdoc-ng"`) which is idiomatic.

#### C6. No error types
All errors are `fmt.Errorf` strings. For a library, typed errors (or at minimum sentinel errors) would help consumers handle specific failure modes.

---

## 4. Missing Infrastructure for Publication

| Item | Status | Priority |
|------|--------|----------|
| README.md | **DONE** | ~~P0~~ |
| LICENSE file | **DONE** (MIT, honouring original authors) | ~~P0~~ |
| Valid Go module path | **DONE** (`github.com/jdevera/shdoc-ng`) | ~~P0~~ |
| GitHub repo | **DONE** (private, `github.com/jdevera/shdoc-ng`) | ~~P0~~ |
| ~~`.goreleaser.yaml`~~ | **DONE** | ~~P0~~ |
| ~~GitHub Actions CI~~ | **DONE** (ci.yml + release.yml) | ~~P0~~ |
| ~~`--version` flag~~ | **DONE** | ~~P0~~ |
| ~~Fix Makefile~~ | **DONE** (removed) | ~~P0~~ |
| Homebrew tap | Missing | **P1** |
| Shell completions | Missing | **P1** |
| Man page | Missing | **P2** |
| pkg.go.dev presence | Ready (push + tag needed) | **P1** |
| AUR package | Missing | **P2** |
| `.github/CONTRIBUTING.md` | Missing | **P2** |

---

## 5. Roadmap to Published Tool

### Phase 0: Foundation (must-do before any release)

- [x] **Add LICENSE** — MIT, honouring original shdoc authors
- [x] **Fix module path** — `github.com/jdevera/shdoc-ng`, all imports updated
- [x] **Write README.md** — installation, usage, full tag reference, LSP setup, differences from original
- [x] **Create GitHub repo** — private at `github.com/jdevera/shdoc-ng`
- [x] **Add `--version`** — inject via `ldflags` at build time, use `cobra.Command.Version`
- [x] **Remove Makefile** — Go toolchain + GoReleaser covers all build needs

### Phase 1: CI & Release Pipeline

- [x] **GitHub Actions CI** — `go test ./...`, `go vet ./...`, `golangci-lint`, matrix across Go 1.25/1.26
- [x] **GoReleaser config** — `.goreleaser.yaml` with multi-platform builds (linux/darwin/windows, amd64/arm64), version injection via ldflags
- [ ] **Push to GitHub** — initial push to the private repo
- [ ] **Tag v0.1.0** — first release with GitHub Release assets
- [ ] **Homebrew tap** — create `homebrew-tap` repo, add `brews` section to goreleaser config
- [ ] **Shell completions** — add `completion` subcommand via Cobra's built-in support

### Phase 2: Quality & Polish

- [ ] **Default command** — make `shdoc-ng` without subcommand behave like `shdoc-ng generate` (or at least print useful help)
- [ ] **Add `--quiet` flag** — suppress warnings on stderr
- [ ] **Terminal detection** — warn when stdin is a TTY and no `-i` flag given
- [ ] **Fix regex-in-loop** — move `pipeRe` to package level in `option.go`
- [ ] **Unexport `BpStripRe`** — provide a public helper function instead
- [ ] **Add CLI tests** — test the `generate`, `check`, `schema`, `template` subcommands end-to-end
- [ ] **Add LSP tests** — at minimum, test diagnostics and hover output
- [ ] **Add `--strict` to check** — differentiate warnings from errors

### Phase 3: Community & Distribution

- [ ] **Make repo public**
- [ ] **pkg.go.dev** — auto-indexed once public + tagged
- [ ] **AUR package** — add `aurs` section to goreleaser
- [ ] **deb/rpm packages** — add `nfpms` section to goreleaser
- [ ] **VSCode Marketplace** — publish the extension (requires publisher account)
- [ ] **Announce** — blog post, Reddit (r/bash, r/golang, r/commandline), Hacker News
- [ ] **CONTRIBUTING.md** — guide for contributors
- [ ] **Issue templates** — bug report, feature request

### Phase 4: Future Features (post-launch)

- [ ] **Watch mode** — `shdoc-ng generate --watch` for live doc regeneration
- [ ] **Multi-file support** — generate docs from a directory of scripts
- [ ] **Custom tags** — user-defined tags via config
- [ ] **Doctest** — validate `@example` blocks actually run
- [ ] **GitHub Pages integration** — auto-publish docs from CI
- [ ] **Man page generation** — via cobra-man or mango

---

## 6. What's Next (ordered by impact)

1. ~~**Add `--version` flag**~~ DONE
2. ~~**GitHub Actions CI + GoReleaser**~~ DONE — gates quality, builds trust; add `ci.yml` with test + lint
4. **GoReleaser config** — `.goreleaser.yaml` for multi-platform binaries
5. **Push to GitHub + tag v0.1.0** — first release
6. **Shell completions** — low effort, high UX value
7. **Default command or better no-args help** — first impression for new users

---

## 7. Recommended CI Workflow

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.25', '1.26']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
      - run: go test -race ./...
      - run: go vet ./...
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - uses: golangci/golangci-lint-action@v8

# .github/workflows/release.yml
name: Release
on:
  push:
    tags: ['v*']
permissions:
  contents: write
jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: stable
      - uses: goreleaser/goreleaser-action@v7
        with:
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 8. Summary Assessment

**shdoc-ng is a well-architected tool with a clear competitive advantage in a niche that has no active competition.** The core parsing pipeline is solid, the conformance test suite is thorough, and the feature set (LSP, multi-format output, linting) is genuinely unique among shell documentation generators.

The foundation is now in place (LICENSE, README, go.mod, GitHub repo). The remaining barriers to a v0.1.0 release are **`--version` flag, Makefile fix, CI, and GoReleaser** — all achievable in a short sprint.

The code itself has some issues (regex-in-loop, exported internals, monolithic LSP file, no CLI/LSP tests) but nothing that blocks a v0.1.0 release. These are quality-of-life improvements for Phase 2.

**Bottom line:** Foundation done. Next step is the CI/release pipeline — about a day of work to reach a tagged v0.1.0 with multi-platform binaries.
