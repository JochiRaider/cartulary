#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)"
VERBOSE_MODE="${CI_VERBOSE:-${VERBOSE:-0}}"

is_verbose() {
  [[ "$VERBOSE_MODE" == "1" ]]
}

run_phase() {
  local phase="$1"
  shift
  local -a command=("$@")

  echo "== ${phase} =="
  if is_verbose; then
    "${command[@]}"
    return
  fi

  local log_file
  log_file="$(mktemp -t cartulary-ci-XXXX.log)"
  if "${command[@]}" >"$log_file" 2>&1; then
    rm -f "$log_file"
    return
  fi

  echo "phase failed: ${phase}" >&2
  echo "failing command: ${command[*]}" >&2
  echo "----- phase output begin -----" >&2
  cat "$log_file" >&2
  echo "----- phase output end -----" >&2
  rm -f "$log_file"
  exit 1
}

cd "$ROOT_DIR"
run_phase "make generate" make generate
run_phase "make test" make test
run_phase "make lint" make lint
run_phase "make check" make check

echo "provider-neutral CI contract passed"
