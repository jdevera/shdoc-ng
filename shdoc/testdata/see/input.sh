# @name shdoc @see tests
# @brief Test @see functionnality.
# @description Tests for shdoc processing of @see keyword.

# @description test-see dummy function.
# @see test-failing-see
# @see test-working-see test-failing-see
# @see some:other:func()
# @see .string-starting-by-dot
# @see ./some/relative/path
# @see ../some/other/relative/path
# @see /some/absolute/path
# @see .../some/strange/string
# @see https://github.com/reconquest/shdoc
# @see file://var/log/syslog
# @see shdoc: https://github.com/reconquest/shdoc
# @see shdoc (https://github.com/reconquest/shdoc) and https://github.com/reconquest/import.bash
# @see [shdoc](https://github.com/reconquest/shdoc)
# @see Shell documation generator [shdoc](https://github.com/reconquest/shdoc).
# @see Shell documation generator [shdoc](https://github.com/reconquest/shdoc) (and https://github.com/reconquest/import.bash).
test-working-see() {
}

