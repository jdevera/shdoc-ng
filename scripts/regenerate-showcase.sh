#!/bin/bash
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

for fmt in markdown html json; do
    ext="${fmt}"
    [ "$fmt" = "markdown" ] && ext="md"
    go run ./cmd/shdoc-ng generate --format "$fmt" \
        -i examples/showcase.sh \
        -o "examples/showcase.${ext}"
done

git add examples/showcase.md examples/showcase.html examples/showcase.json
