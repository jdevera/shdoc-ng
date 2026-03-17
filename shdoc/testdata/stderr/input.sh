# @name shdoc @stderr tests
# @brief Test @stderr functionnality.
# @description Tests for shdoc processing of @stderr keyword.
# @description test-stderr dummy function.
# @stderr Standard stderr message.
# @stdout section should appear before stderr section.
# @see see-also-after-stderr-section
# @stderr Stderr message with [markdown link](https://github.com/reconquest/shdoc).
# @stderr       Indented with spaces stderr message.
# @stderr Multiple lines
#   stderr message.
# line outside of multiple lines stderr message (ignored).
# @stderr Failed multiple lines
    # std err message.
    #   @stderr Idented multiple lines
    #       stderr message
# @stderr Stderr message with trailing spaces.      
test-stderr() {
}
