# mylib

A showcase of every shdoc-ng annotation.

## Overview

This library demonstrates all supported tags.

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

## String Utilities

Common string operations.

### greet

Greet someone by name.

Prints a friendly greeting to stdout. Returns non-zero
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

