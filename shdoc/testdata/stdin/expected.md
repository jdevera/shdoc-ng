# shdoc @stdin tests

Test @stdin functionnality.

## Overview

Tests for shdoc processing of @stdin keyword.

## Index

* [test-stdin](#test-stdin)

### test-stdin

test-stdin dummy function.

#### Exit codes

* **0**: Exit code section appears before stdin section.

#### Input on stdin

* simple one line message.
* one line message with indentation and trailing spaces.
* indented two lines message
  to test how indentation is trimmed.
* tree lines messages
  with trailing spaces
  and random indentation.

#### Output on stdout

* Stdout section appears after stdin section.

#### Output on stderr

* Error output description.

#### See also

* [see-also-after-stderr-section](#see-also-after-stderr-section)

