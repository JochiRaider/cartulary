#!/usr/bin/env bash
set -euo pipefail

fail() {
  echo "$*" >&2
  exit 1
}

if [[ "$#" -eq 0 ]]; then
  fail "usage: list-build-inputs.sh <path> [<path>...]"
fi

# Called from an artifact recipe after the pinned runtime prerequisite, never
# during Make graph expansion. The shared admission path prunes human material
# before traversal and rejects symlink ancestors before probing descendants.
script_dir="$(unset CDPATH && cd -- "$(dirname "$0")" && pwd)"
"${NODE_BIN:?build input discovery requires the pinned Node runtime}" \
  "$script_dir/../test-catalog/executable-files-cli.mjs" --lines "$PWD" "$@"
