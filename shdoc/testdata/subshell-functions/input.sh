# @description Inline subshell function.
inline_subshell() (
    echo "runs in subshell"
)

# @description Subshell paren on next line.
nextline_subshell()
(
    echo "runs in subshell"
)

# @description Keyword form with subshell.
function keyword_subshell (
    echo "runs in subshell"
)
