# @name shdoc @stdout tests
# @brief Test @stdout functionnality.
# @description Tests for shdoc processing of @stdout keyword.
# @description test-stdout dummy function.
# @see see-also-after-stderr-section
# @stdout simple one line message.
# @stdout         one line message with indentation and trailing spaces.    
    #   @stdout   indented two lines message
    #         to test how indentation is trimmed.
    #   Message without sufficient indentation (ignored).
# @stderr Error output description.
# @stdout tree lines messages    
#     with trailing spaces    
#   and random indentation.  
# @stdin Input stream description.
test-stdout() {
}
