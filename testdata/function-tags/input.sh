# @description Multiline description goes here and
# there
#
# @example
#   some:other:func a b c
#   echo 123
#
# @arg $1 string Some arg.
# @arg $@ any Rest of arguments.
#
# @noargs
#
# @set A string Variable was set
#
# @exitcode 0  If successfull.
# @exitcode >0 On failure
# @exitcode 5  On some error.
#
# @stdin Path to something.
# @stdout Path to something.
# @stderr Stderr description.
#
# @see some:other:func()
# @see Shell documation generator [shdoc](https://github.com/reconquest/shdoc).
some:first:func() {
