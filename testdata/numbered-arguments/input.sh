# @name shdoc @arg tests
# @brief Test @arg functionnality.
# @description Tests for shdoc processing of @arg keyword.
# @arg $4 int 4th arg.
# @arg $6 string 6th arg.
# @set ARG_TESTED A variable set by the function.
# @arg $5 int 5th arg.
# @arg $@ string All other arguments.
# @arg $1 string 1st arg.
# @example
#   test-arg 'my-tested-argument'
#
# @arg $3    bool    3rd arg with indentation and trailing spaces.        
# @arg $2 string 2nd arg.
    # @arg $7 string 7th arg with indentation before #.
#       @arg $8 array[] 8th arg with indentation between # and @arg.
# @arg      $9 string 9th arg with indentation between @arg and number.
# @arg $10 string 10th arg.
# @arg $11 string 11th arg.
test-arg() {
}
