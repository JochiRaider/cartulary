#!/usr/bin/env bash
set -euo pipefail

fail() {
  echo "$*" >&2
  exit 1
}

if [[ "$#" -eq 0 ]]; then
  fail "usage: list-build-inputs.sh <path> [<path>...]"
fi

if ! command -v rg >/dev/null 2>&1; then
  fail "build input discovery requires rg on PATH"
fi

for root in "$@"; do
  if [[ ! -e "$root" ]]; then
    fail "missing build input root: $root"
  fi
done

rg --files -- "$@" | LC_ALL=C sort
