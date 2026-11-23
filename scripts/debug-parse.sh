#!/bin/bash
# Parses a shell script with shdoc-ng and shows the raw JSON output
# for inspecting how tags are parsed. Useful during development.
#
# Usage: scripts/debug-parse.sh <input.sh> [jq-filter]
#
# Examples:
#   scripts/debug-parse.sh test.sh                          # full JSON
#   scripts/debug-parse.sh test.sh '.functions[0].stderr'   # just stderr
#   scripts/debug-parse.sh test.sh '.functions[0].args'     # just args
set -euo pipefail
cd "$(git rev-parse --show-toplevel)"

input="${1:?Usage: $0 <input.sh> [jq-filter]}"
filter="${2:-.}"

go run ./cmd/shdoc-ng generate --format json -i "$input" | jq "$filter"
