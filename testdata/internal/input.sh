# @name TestInternal
# @brief Testing @internal tag

# @description Public function that should appear.
# @arg $1 string Name
foo() {
    echo "$1"
}

# @internal
# @description Internal function that should be hidden.
# @arg $1 string Secret
bar() {
    echo "$1"
}

# @description Another public function.
baz() {
    echo "baz"
}
