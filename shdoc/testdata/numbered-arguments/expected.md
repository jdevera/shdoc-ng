# shdoc @arg tests

Test @arg functionnality.

## Overview

Tests for shdoc processing of @arg keyword.

## Index

* [test-arg](#test-arg)

### test-arg

Tests for shdoc processing of @arg keyword.

#### Example

```bash
test-arg 'my-tested-argument'
```

#### Arguments

* **$1** (string): 1st arg.
* **$2** (string): 2nd arg.
* **$3** (bool): 3rd arg with indentation and trailing spaces.
* **$4** (int): 4th arg.
* **$5** (int): 5th arg.
* **$6** (string): 6th arg.
* **$7** (string): 7th arg with indentation before #.
* **$8** (array[]): 8th arg with indentation between # and @arg.
* **$9** (string): 9th arg with indentation between @arg and number.
* **$10** (string): 10th arg.
* **$11** (string): 11th arg.
* **...** (string): All other arguments.

#### Variables set

* **ARG_TESTED** (A): variable set by the function.

