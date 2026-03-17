#!/usr/bin/env bash
# @name TestIndented
# @brief Testing indented functions

# @description Top-level function.
# @noargs
top_func() {
    echo "top"
}

if [[ 1 -eq 2 ]]; then

    # @description Indented function inside a conditional.
    # @noargs
    indented_func() {
        echo "indented"
    }
fi
