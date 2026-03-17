# @name multiline-tags
# @description Test multi-line continuation for all supported tags.

# @description Function with multi-line tags.
#
# @arg $1 string Plugin spec in one of these formats:
#     - `name` — local plugin
#     - `user/repo` — GitHub repo
# @arg $2 int Count of items.
#
# @option -o <path> | --output <path> Output path.
#     If not specified, defaults to stdout.
#
# @set RESULT string The computed result.
#     May contain newlines.
#
# @env HOME string The user's home directory.
#     Used to resolve ~ in paths.
#
# @exitcode 0 Success.
# @exitcode >0 Failure. Possible causes:
#     - Invalid input format
#     - Missing required files
#
# @warning This function modifies global state.
#     Make sure to save context first.
#
# @deprecated Use new_func() instead.
#     This function will be removed in v2.0.
#
# @stdin Input data in CSV format.
#     First line is a header.
#
# @stdout Processed output.
#     Tab-separated values.
#
# @stderr Progress messages.
#     Prefixed with timestamp.
multi_line_func() {
    :
}

# @description Same-level comments do not continue tags.
#
# @arg $1 string First arg.
# This is NOT a continuation.
# @arg $2 string Second arg.
same_level_func() {
    :
}
