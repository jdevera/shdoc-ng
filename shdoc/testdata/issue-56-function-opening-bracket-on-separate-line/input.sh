#!/sbin/sh

# @description My super function.
#
# @arg $1 string A value to print
say-hello()
{
    if [[ ! "$1" ]]; then
        return 1;
    fi

    echo "Hello $1"
}

