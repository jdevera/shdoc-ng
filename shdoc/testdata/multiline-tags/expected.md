# multiline-tags

## Overview

Test multi-line continuation for all supported tags.

## Index

* [multi_line_func](#multilinefunc)
* [same_level_func](#samelevelfunc)

### multi_line_func

Function with multi-line tags.

#### Options

* **-o \<path\>** | **--output \<path\>**

  Output path.

#### Arguments

* **$1** (string): Plugin spec in one of these formats:
* **$2** (int): Count of items.

#### Variables set

* **RESULT** (string): The computed result.

#### Exit codes

* **0**: Success.
* **>0**: Failure. Possible causes:

#### Input on stdin

* Input data in CSV format.
  First line is a header.

#### Output on stdout

* Processed output.
  Tab-separated values.

#### Output on stderr

* Progress messages.
  Prefixed with timestamp.

### same_level_func

Same-level comments do not continue tags.

#### Arguments

* **$1** (string): First arg.
* **$2** (string): Second arg.

