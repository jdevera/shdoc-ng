# @name Flexible option separators
# @description Test that / and tight | work as option separators.

# @description A function with various separator styles.
# @option -s/--sequence  Run in sequence.
# @option -s / --sequence  Run in sequence (spaced slash).
# @option -s|--sequence  Run in sequence (tight pipe).
# @option -v<val>/--value=<val>  Set a value (slash with values).
# @option -v <val> / --value=<val>  Set a value (spaced slash with values).
# @option -a | -b/-c|--delta  Mixed separators.
flexible_func() {
    :
}
