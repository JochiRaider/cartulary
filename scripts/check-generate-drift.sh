#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
GENERATED_PATHS=(
  "internal/gen/contracts"
  "internal/gen/sql"
  "packages/protocol-ts/src/generated"
)

cd "$ROOT_DIR"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "generated artifact drift check must run inside a git work tree" >&2
  exit 1
fi

make generate

if [[ -n "$(git status --short -- "${GENERATED_PATHS[@]}")" ]]; then
  echo "generated artifact drift detected after make generate" >&2
  git status --short -- "${GENERATED_PATHS[@]}" >&2
  echo "diff excerpt (first 200 lines):" >&2
  git --no-pager diff -- "${GENERATED_PATHS[@]}" | sed -n '1,200p' >&2 || true
  exit 1
fi
