## Index

* [tests:value](#testsvalue)

### tests:value

Same, as `tests:eval`, but writes stdout into given variable and
return stderr as expected.

#### Example

```bash
_x() {
    echo "y [$@]"
}
tests:value response _x a b c
tests:assert-equals "$response" "y [a b c]"
```

#### Arguments

* **$1** (string): Variable name.
* **...** (string): String to evaluate.

#### See also

* [tests:eval](#testseval)

