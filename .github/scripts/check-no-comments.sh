#!/usr/bin/env bash
set -euo pipefail

hits=$(
  grep -rnE '(^[[:space:]]*//|[^:"'"'"'`][[:space:]]//[[:space:]])' --include='*.go' . \
    | grep -v 'internal/database/ent/' \
    | grep -vE '//go:(build|generate|embed)' \
    || true
)

if [ -n "$hits" ]; then
  echo "::error::hand-written comments are not allowed (see CLAUDE.md); only //go: directives and generated code under internal/database/ent are exempt"
  echo "$hits"
  exit 1
fi

echo "no hand-written comments found"
