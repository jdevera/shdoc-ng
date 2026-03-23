# mylib

A showcase of every shdoc-ng annotation.

## Overview

This library **demonstrates** all supported tags.

Features include:

* Function documentation
* Deprecation notices
* Grouped sections

#### Authors

* Alice
* Bob

#### License

MIT

#### Version

2.0.0

## Index

* [greet](#greet)
* [farewell](#farewell)
* [old_hello](#old_hello)
* [old_noop](#old_noop)
* [braceless_func](#braceless_func)
* [subshell_func](#subshell_func)
* [multi_arg_func](#multi_arg_func)
* [option_variety](#option_variety)
* [env_and_set_func](#env_and_set_func)
* [multi_exitcode](#multi_exitcode)
* [noargs_but_uses_params](#noargs_but_uses_params)
* [noargs_with_awk](#noargs_with_awk)
* [old_parse](#old_parse)
* [new_parse](#new_parse)
* [empty_tags_func](#empty_tags_func)
* [see_references](#see_references)
* [single_line_doc](#single_line_doc)
* [multi_line_doc](#multi_line_doc)
* [keyword_no_parens](#keyword_no_parens)
* [keyword_with_parens](#keyword_with_parens)
* [multiline_io](#multiline_io)

## String Utilities

Common string operations.

### greet

Greet someone by name.

Prints a friendly **greeting** to stdout. Returns non-zero
if no name is provided.

#### Warnings

* Do not pass empty strings.

#### Example

```bash
greet "World"
greet "$(whoami)"
```

#### Options

* **-u** | **--uppercase**

  Uppercase the name.

* **-n \<count\>** | **--repeat=\<count\>**

  Repeat the greeting.

* **-r** | **--reverse**

  Reverse the greeting.

#### Arguments

* **$1** (string): The name to greet.
* **...** (string): Additional names.

#### Variables set

* **LAST_GREETED** (string): The last name that was greeted.

#### Environment variables

* **HOME** (string): The user's home directory.

#### Exit codes

* **0**: If successful.
* **1**: If no name was provided.

#### Input on stdin

* A fallback name, read if $1 is empty.

#### Output on stdout

* The greeting string.

#### Output on stderr

* A warning when no name is given.

#### Labels

`couldbeascript` · `greeting`

#### See also

* [farewell()](#farewell)
* [https://www.gnu.org/software/bash/](https://www.gnu.org/software/bash/)
* [./lib/helpers.sh](./lib/helpers.sh)
* [Bash manual](https://www.gnu.org/software/bash/manual/)
* See also the [bash guide](https://tldp.org/LDP/abs/html/) and [https://mywiki.wooledge.org/BashFAQ](https://mywiki.wooledge.org/BashFAQ).

### farewell

Say goodbye.

#### Arguments

* **$1** (string): The name to bid farewell.

#### Exit codes

* **0**: Always.

#### See also

* [greet()](#greet)

## Deprecated Utilities

Functions kept for backwards compatibility.

### old_hello

**DEPRECATED:** Use greet() instead.

This function is obsolete.

### old_noop

**DEPRECATED.**

### braceless_func

A function whose opening brace
is on the next line.

_Function has no arguments._

### subshell_func

A subshell function using parens
instead of braces.

#### Arguments

* **$1** (string): A command to run in a subshell.

## LSP & Highlight Testing

Cases to exercise syntax highlighting, diagnostics, and LSP features.

### multi_arg_func

Multiple arguments with different types.

#### Arguments

* **$1** (string): The input file path.
* **$2** (int): Number of lines to read.
* **$3** (bool): Whether to trim whitespace.
* **...** (string): Remaining arguments forwarded to helper.

### option_variety

Various option forms.

#### Options

* **-v**

  Enable verbose mode.

* **--dry-run**

  Do not actually execute.

* **-o \<path\>** | **--output \<path\>**

  Write output to file.

* **-f** | **--force**

  Force overwrite.

* **--timeout=\<seconds\>**

  Timeout in seconds.

### env_and_set_func

Exercises @set and @env highlighting.

_Function has no arguments._

#### Variables set

* **RESULT_CODE** (int): The result code.
* **OUTPUT_PATH** (string): The generated output path.

#### Environment variables

* **LANG** (string): The current locale.
* **TERM** (string): The terminal type.

### multi_exitcode

Multiple exit codes with prefix variants.

_Function has no arguments._

#### Exit codes

* **0**: Success.
* **1**: General error.
* **2**: Misuse of shell builtins.
* **127**: Command not found.
* **>0**: Any failure.
* **!0**: Non-zero exit.

### noargs_but_uses_params

Should warn: uses $1 despite @noargs.

_Function has no arguments._

### noargs_with_awk

Should NOT warn: $1 is inside single quotes (awk).

_Function has no arguments._

### old_parse

**DEPRECATED:** Replaced by new_parse().

Old parser, use new_parse() instead.

#### Arguments

* **$1** (string): Input text.

### new_parse

New parser.

#### Arguments

* **$1** (string): Input text.

### empty_tags_func

#### Environment variables

* **FOO** (): 

#### Input on stdin

* 

#### Output on stdout

* 

#### Output on stderr

* 

#### See also

* [](#)

### see_references

A function that references others via @see.

_Function has no arguments._

#### See also

* [greet()](#greet)
* [farewell()](#farewell)
* [old_parse()](#old_parse)

### single_line_doc

This is a single-line doc block (no fold arrow expected).

### multi_line_doc

This is a multi-line doc block.

It spans several comment lines, so it should
be foldable in the editor.

#### Arguments

* **$1** (string): Something.

#### Exit codes

* **0**: Always.

### keyword_no_parens

Uses the 'function' keyword without parens.

_Function has no arguments._

### keyword_with_parens

Uses the 'function' keyword with parens.

#### Arguments

* **$1** (string): A value.

### multiline_io

Demonstrates multi-line IO docs.

_Function has no arguments._

#### Input on stdin

* Lines of text, one per line.
  Each line should be tab-separated.
  Empty lines are skipped.

#### Output on stdout

* Processed output.
  First line is a header.
  Subsequent lines are data rows.

#### Output on stderr

* Progress messages.
  Prefixed with the current timestamp.

