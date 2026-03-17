# @name shdoc @stdin tests
# @brief Test @stdin functionnality.
# @description Tests for shdoc processing of @stdin keyword.
# @description test-stdin dummy function.
# @see see-also-after-stderr-section
# @stdin simple one line message.
# @stdout Stdout section appears after stdin section.
# @stdin         one line message with indentation and trailing spaces.    
    #   @stdin   indented two lines message
    #         to test how indentation is trimmed.
    #   Message without sufficient indentation (ignored).
# @stderr Error output description.
# @stdin tree lines messages    
#     with trailing spaces    
#   and random indentation.
# @exitcode 0 Exit code section appears before stdin section.
test-stdin() {
}
