# @name shdoc @option tests for options
# @brief Test @option functionnality for options.
# @description Tests for shdoc processing of @option keyword.
# @option -2 Run twice as fast.
# @option -h Show help message.
# @option --help Show help message.
# @option -h | --help Show help message.
# @option -v<my value> Set a value for short option (joined).
# @option -v <my value> Set a value for short option (space separated).
# @option option with invalid format.
# @option --value=<my value> Set a value for long option (= joined).
# @set ARG_TESTED A variable set by the function.
# @option ---another option with invalid format.
# @option --value <my value> Set a value for long option (space separated).
# @example
#   test-arg 'my-tested-argument'
#
# @option -v <my value> | --value <my value> Set a value.
# @option -v<my value> | --value=<my value> Set a value (joined).
# @option --value=<my value>  |  -v<my value> |   --longer-value <my value> Set a value via three different options.
# @arg -v<my value> | --value=<my value> Set a value described by @arg.
test-arg() {
}
