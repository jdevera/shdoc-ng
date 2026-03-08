#!/bin/bash
# @name mylib
# @brief A showcase of every shdoc-ng annotation.
# @author Alice
# @author Bob
# @license MIT
# @version 2.0.0
# @description
#     This library **demonstrates** all supported tags.
#
#     Features include:
#
#      * Function documentation
#      * Deprecation notices
#      * Grouped sections

# @section String Utilities
# @description Common string operations.

# @brief Greet someone by name.
# @description Greet someone by name.
#
# Prints a friendly **greeting** to stdout. Returns non-zero
# if no name is provided.
#
# @warn Do not pass empty strings.
#
# @example
#   greet "World"
#   greet "$(whoami)"
#
# @option -u | --uppercase  Uppercase the name.
# @option -n <count> | --repeat=<count>  Repeat the greeting.
# @option -r/--reverse  Reverse the greeting.
#
# @arg $1 string The name to greet.
# @arg $@ string Additional names.
#
# @set LAST_GREETED string The last name that was greeted.
#
# @env HOME string The user's home directory.
#
# @exitcode 0 If successful.
# @exitcode 1 If no name was provided.
#
# @stdin A fallback name, read if $1 is empty.
# @stdout The greeting string.
# @stderr A warning when no name is given.
#
# @see farewell()
# @see https://www.gnu.org/software/bash/
# @see ./lib/helpers.sh
# @see [Bash manual](https://www.gnu.org/software/bash/manual/)
# @see See also the [bash guide](https://tldp.org/LDP/abs/html/) and https://mywiki.wooledge.org/BashFAQ.
greet() {
    if [[ -z "$1" ]]; then
        echo "Warning: no name provided" >&2
        return 1
    fi
    export LAST_GREETED="$1"
    echo "Hello, $1!"
}

# @description Say goodbye.
#
# @arg $1 string The name to bid farewell.
#
# @exitcode 0 Always.
#
# @see greet()
farewell() {
    echo "Goodbye, ${1:-stranger}!"
}

# @section Deprecated Utilities
# @description Functions kept for backwards compatibility.

# @description This function is obsolete.
# @deprecated Use greet() instead.
old_hello() {
    echo "Hello!"
}

# @deprecated
old_noop() {
    :
}

# @internal
_private_helper() {
    :
}

# @description A function whose opening brace
# is on the next line.
#
# @noargs
braceless_func()
{
    echo "I work too"
}

# @description A subshell function using parens
# instead of braces.
#
# @arg $1 string A command to run in a subshell.
subshell_func() (
    eval "$1"
)

# @section LSP & Highlight Testing
# @description Cases to exercise syntax highlighting, diagnostics, and LSP features.

# --- Highlighting: @arg variations ---

# @description Multiple arguments with different types.
#
# @arg $1 string The input file path.
# @arg $2 int Number of lines to read.
# @arg $3 bool Whether to trim whitespace.
# @arg $@ string Remaining arguments forwarded to helper.
multi_arg_func() {
    echo "$1" "$2" "$3" "$@"
}

# --- Highlighting: @option variations ---

# @desc Various option forms.
#
# @opt -v Enable verbose mode.
# @opt --dry-run Do not actually execute.
# @opt -o <path> | --output <path>  Write output to file.
# @opt -f/--force  Force overwrite.
# @opt --timeout=<seconds>  Timeout in seconds.
option_variety() {
    echo "options"
}

# --- Highlighting: @set and @env ---

# @description Exercises @set and @env highlighting.
#
# @set RESULT_CODE int The result code.
# @set OUTPUT_PATH string The generated output path.
# @env LANG string The current locale.
# @env TERM string The terminal type.
#
# @noargs
env_and_set_func() {
    export RESULT_CODE=0
    export OUTPUT_PATH="/tmp/out"
}

# --- Highlighting: @exitcode numbers ---

# @desc Multiple exit codes.
#
# @exit 0 Success.
# @exit 1 General error.
# @exit 2 Misuse of shell builtins.
# @exit 127 Command not found.
#
# @noargs
multi_exitcode() {
    return 0
}

# --- Diagnostics: @noargs with positional params ---

# @desc Should warn: uses $1 despite @noargs.
#
# @noargs
noargs_but_uses_params() {
    echo "$1"
    echo "$@"
    echo "$#"
}

# @description Should NOT warn: $1 is inside single quotes (awk).
#
# @noargs
noargs_with_awk() {
    echo "hello" | awk '{print $1}'
    echo "hello" | sed -n 's/h/H/p'
}

# --- Diagnostics: deprecated strikethrough ---

# @description Old parser, use new_parse() instead.
# @deprecated Replaced by new_parse().
#
# @arg $1 string Input text.
old_parse() {
    echo "$1"
}

# @description New parser.
#
# @arg $1 string Input text.
new_parse() {
    echo "parsed: $1"
    old_parse "$1"
}

# --- Diagnostics: empty tag values (should all warn) ---

# @description
# @arg
# @option
# @exitcode
# @see
# @warning
# @set
# @env
# @stdin
# @stdout
# @stderr
# @example
# @env FOO
empty_tags_func() {
    :
}

# --- LSP: go-to-definition via @see ---

# @description A function that references others via @see.
#
# @see greet()
# @see farewell()
# @see old_parse()
#
# @noargs
see_references() {
    :
}

# --- LSP: code action target (undocumented function) ---

undocumented_function() {
    echo "I have no doc block — the LSP should offer to insert one."
}

another_undocumented() {
    echo "Me neither."
}

# --- LSP: folding (multi-line vs single-line doc blocks) ---

# @description This is a single-line doc block (no fold arrow expected).
single_line_doc() {
    :
}

# @description This is a multi-line doc block.
#
# It spans several comment lines, so it should
# be foldable in the editor.
#
# @arg $1 string Something.
# @exitcode 0 Always.
multi_line_doc() {
    echo "$1"
}

# --- Edge case: function keyword syntax ---

# @description Uses the 'function' keyword without parens.
#
# @noargs
function keyword_no_parens {
    echo "function keyword, no parens"
    multi_line_doc a
}

# @description Uses the 'function' keyword with parens.
#
# @arg $1 string A value.
function keyword_with_parens() {
    echo "$1"
}

# --- Edge case: multi-line stdin/stdout/stderr ---

# @description Demonstrates multi-line IO docs.
#
# @stdin Lines of text, one per line.
#   Each line should be tab-separated.
#   Empty lines are skipped.
#
# @stdout Processed output.
#   First line is a header.
#   Subsequent lines are data rows.
#
# @stderr Progress messages.
#   Prefixed with the current timestamp.
#
# @noargs
multiline_io() {
    while IFS= read -r line; do
        echo "$line"
    done
}
# this is a regular comment (awk foo)
