# shdoc @option tests for options

Test @option functionnality for options.

## Overview

Tests for shdoc processing of @option keyword.

## Index

* [test-arg](#test-arg)

### test-arg

Tests for shdoc processing of @option keyword.

#### Example

```bash
test-arg 'my-tested-argument'
```

#### Options

* **-2**

  Run twice as fast.

* **-h**

  Show help message.

* **--help**

  Show help message.

* **-h** | **--help**

  Show help message.

* **-v\<my value\>**

  Set a value for short option (joined).

* **-v \<my value\>**

  Set a value for short option (space separated).

* **--value=\<my value\>**

  Set a value for long option (= joined).

* **--value \<my value\>**

  Set a value for long option (space separated).

* **-v \<my value\>** | **--value \<my value\>**

  Set a value.

* **-v\<my value\>** | **--value=\<my value\>**

  Set a value (joined).

* **--value=\<my value\>** | **-v\<my value\>** | **--longer-value \<my value\>**

  Set a value via three different options.

* **-v\<my value\>** | **--value=\<my value\>**

  Set a value described by @arg.

* option with invalid format.
* ---another option with invalid format.

#### Variables set

* **ARG_TESTED** (A): variable set by the function.

