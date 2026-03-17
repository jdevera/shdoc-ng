# multiline-tags

## Overview

Test multi-line continuation for all supported tags.

## Index

* [multi_line_func](#multi_line_func)
* [same_level_func](#same_level_func)

### multi_line_func

**DEPRECATED:** Use new_func() instead.
This function will be removed in v2.0.

Function with multi-line tags.

#### Warnings

* This function modifies global state.
  Make sure to save context first.

#### Options

* **-o \<path\>** | **--output \<path\>**

  Output path.
  If not specified, defaults to stdout.

#### Arguments

* **$1** (string): Plugin spec in one of these formats:
  - `name` — local plugin
  - `user/repo` — GitHub repo
* **$2** (int): Count of items.

#### Variables set

* **RESULT** (string): The computed result.
  May contain newlines.

#### Environment variables

* **HOME** (string): The user's home directory.
  Used to resolve ~ in paths.

#### Exit codes

* **0**: Success.
* **>0**: Failure. Possible causes:
  - Invalid input format
  - Missing required files

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

