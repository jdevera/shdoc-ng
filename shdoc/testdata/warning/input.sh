#!/bin/bash

# @description A dangerous function.
#
# @warning This will delete files.
# @warning Use with caution.
# @arg $1 string The target path.
danger() {
    rm -rf "$1"
}
