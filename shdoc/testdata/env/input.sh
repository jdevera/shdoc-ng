#!/bin/bash

# @description Sets up the environment.
#
# @env HOME string The user's home directory.
# @env PATH string The system path.
# @arg $1 string The config file.
setup_env() {
    export HOME="$1"
}
