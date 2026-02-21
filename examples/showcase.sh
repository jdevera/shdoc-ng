#!/bin/bash
# @name mylib
# @brief A showcase of every shdoc-ng annotation.
# @description
#     This library demonstrates all supported tags.
#
#     Features include:
#      * Function documentation
#      * Deprecation notices
#      * Grouped sections

# @section String Utilities
# @description Common string operations.

# @description Greet someone by name.
#
# Prints a friendly greeting to stdout. Returns non-zero
# if no name is provided.
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
# @exitcode 0 If successful.
# @exitcode 1 If no name was provided.
#
# @stdin A fallback name, read if $1 is empty.
# @stdout The greeting string.
# @stderr A warning when no name is given.
#
# @see farewell()
# @see [Bash scripting guide](https://tldp.org/LDP/abs/html/).
greet() {
    if [[ -z "$1" ]]; then
        echo "Warning: no name provided" >&2
        return 1
    fi
    LAST_GREETED="$1"
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
