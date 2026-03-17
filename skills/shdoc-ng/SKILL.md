---
name: shdoc-ng
description: Add or update shdoc-ng documentation annotations in shell scripts. Use when asked to document shell functions, add doc blocks, or generate shell script documentation.
---

# Document Shell Scripts with shdoc-ng

Add or update documentation annotations in shell scripts following shdoc-ng syntax. After editing, validate with `shdoc-ng check`.

## Instructions

The user may specify one or more shell script paths to document. If no path is given, ask which file(s) to document.

1. Read the target shell file(s) completely before making changes.
2. Identify all public functions (skip functions prefixed with `_` unless asked).
3. For each function, read its body to understand arguments, exit codes, side effects, and I/O.
4. Add or update doc blocks directly above each function declaration.
5. If the file lacks a file-level header block, add one at the top.
6. After all edits, run `shdoc-ng check -i <file>` and fix any warnings.

## shdoc-ng Annotation Syntax

Doc blocks are consecutive `#` comment lines immediately above a function declaration. A blank line or code line between the comment block and the function breaks the association.

### File-level tags (meta block)

A meta block is a doc comment block at the top of the file that is NOT followed by a function declaration. It describes the file/library itself.

```bash
#!/bin/bash
# @name my_library
# @brief One-line summary of what this library does.
# @description
#     Longer description with **markdown** support.
#
#     Can span multiple paragraphs.
#
# @author Author Name
# @license MIT
# @version 1.0.0
```

| Tag            | Format                          | Notes |
|----------------|---------------------------------|-------|
| `@name`        | `@name <library-name>`          | Sets the document title. `@file` is an alias. |
| `@brief`       | `@brief <one-liner>`            | Short summary shown under the title. |
| `@description` | `@description` or `@description <text>` | Multi-line: continuation lines are `#` comment lines until the next tag or end of block. Supports markdown. `@desc` is a shorthand. |
| `@author`      | `@author <name>`                | Repeatable for multiple authors. |
| `@license`     | `@license <license>`            | |
| `@version`     | `@version <version>`            | |

### Section tags (meta block)

Sections group functions under a heading. Place a section tag in a meta block (not above a function). All functions after it belong to that section until the next `@section`.

```bash
# @section String Utilities
# @description Helper functions for string manipulation.
```

| Tag          | Format                       | Notes |
|--------------|------------------------------|-------|
| `@section`   | `@section <Section Name>`    | Starts a new section. Sticky: applies to all following functions. |

### Function-level tags

Place these in a comment block directly above the function declaration.

```bash
# @description Greet someone by name.
#
# Prints a greeting to stdout. Supports **markdown** in descriptions.
#
# @arg $1 string The name to greet.
# @arg $@ string Additional names.
#
# @option -u | --uppercase  Uppercase the name.
# @option -n <count> | --repeat=<count>  Repeat the greeting.
# @option -f/--force  Force overwrite.
#
# @set LAST_GREETED string The last name greeted.
# @set COUNT
# @env HOME string The user's home directory.
#
# @exitcode 0 If successful.
# @exitcode >0 If an error occurred.
#
# @stdin A fallback name if $1 is empty.
# @stdout The greeting string.
# @stderr A warning when no name is given.
#
# @example
#   greet "World"
#   greet "$(whoami)"
#
# @see other_func()
# @see https://example.com
#
# @warning Do not pass empty strings.
# @deprecated Use new_greet() instead.
greet() {
    echo "Hello, $1!"
}
```

#### Tag reference

| Tag            | Format | Lines | Notes |
|----------------|--------|-------|-------|
| `@description` | `@description <text>` or `@description` + continuation lines | Multi | Markdown. `@desc` is a shorthand. |
| `@arg`         | `@arg $N type Description` | Multi | `$1`..`$9`, `$@`. Type is required (e.g. `string`, `int`, `bool`). |
| `@option`      | `@option <flags> Description` | Multi | Flag forms: `-f`, `--flag`, `-f/--flag`, `-f \| --flag`, `--flag=<value>`, `-f <value>`. `@opt` is a shorthand. |
| `@noargs`      | `@noargs` | — | Marks function as taking no arguments. Mutually exclusive with `@arg`. |
| `@set`         | `@set VARNAME [type [Description]]` | Multi | Documents a variable the function sets/exports. |
| `@env`         | `@env VARNAME [type [Description]]` | Multi | Documents an environment variable the function reads. |
| `@exitcode`    | `@exitcode [>\|!]N Description` | Multi | Supports prefixes: `>N` (greater than), `!N` (not equal). `@exit` is a shorthand. |
| `@stdin`       | `@stdin Description` | Multi | Continuation lines indented past the tag. Repeatable. |
| `@stdout`      | `@stdout Description` | Multi | Same as `@stdin`. Repeatable. |
| `@stderr`      | `@stderr Description` | Multi | Same as `@stdin`. Repeatable. |
| `@example`     | `@example` + indented continuation lines | Multi | Lines must have at least one space after `#`. |
| `@see`         | `@see func_name()` or `@see <URL>` or `@see [text](URL)` | Multi | Cross-references. Function refs use `()` suffix. |
| `@warning`     | `@warning <text>` | Multi | `@warn` is a shorthand. |
| `@deprecated`  | `@deprecated [message]` | Multi | Message is optional. |
| `@internal`    | `@internal` | — | Excludes the function from generated documentation. |

#### Tag shorthands

| Shorthand | Expands to     |
|-----------|----------------|
| `@desc`   | `@description` |
| `@opt`    | `@option`      |
| `@exit`   | `@exitcode`    |
| `@warn`   | `@warning`     |

### Multi-line values

**Descriptions** continue on subsequent `#` lines until the next `@tag` or end of block:

```bash
# @description First line.
#
# Second paragraph with **bold**.
```

**Examples** require indentation (at least one space after `#`):

```bash
# @example
#   command --flag arg
#   another_command
```

**stdin/stdout/stderr** support multi-line continuations indented past the tag:

```bash
# @stdout Processed output.
#   First line is a header.
#   Subsequent lines are data rows.
```

## Common patterns

### Minimal function doc

```bash
# @description Brief explanation of what the function does.
my_func() {
```

### Function with no arguments

```bash
# @description Reset all counters.
# @noargs
# @exitcode 0 Always.
reset_counters() {
```

### Internal/private helper

```bash
# @internal
_helper() {
```

## Validation

After editing, always run:

```bash
shdoc-ng check -i <file>
```

Fix any warnings before finishing. Common warnings:
- Empty tag values (e.g. `@author` with no name)
- Invalid `@arg` format (missing `$N`, type, or description)
- Invalid `@set`/`@env` format (missing variable name)
- Invalid `@exitcode` format (missing numeric code)
- `@noargs` on a function that uses `$1`, `$@`, `$#` in its body
