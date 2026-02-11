# @name TestUndocumented
# @brief Testing undocumented functions

undocumented_one() {
    echo "no docs"
}

# @description A documented function.
documented_func() {
    echo "has docs"
}

undocumented_two() {
    echo "also no docs"
}

# @description Another documented function.
another_documented() {
    echo "also has docs"
}
