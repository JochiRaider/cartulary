#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-phase.sh"
cleanup_paths=()

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -f "$path"
  done
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output to omit [$needle]"
  fi
}

extract_log_path() {
  printf '%s\n' "$1" | sed -n 's/^phase log: //p' | tail -n 1
}

quiet_success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
    "$HELPER" "quiet success" -- bash -lc 'echo hidden-success-output'
)"
assert_contains "$quiet_success_output" "== quiet success ==" "quiet success banner"
assert_not_contains "$quiet_success_output" "hidden-success-output" "quiet success suppresses routine output"

success_log_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_OUTPUT_ALLOW_SUCCESS_LOG=1 \
    "$HELPER" "success log replay" -- bash -lc 'echo keep-this-warning >&2'
)"
assert_contains "$success_log_output" "== success log replay ==" "success log replay banner"
assert_contains "$success_log_output" "keep-this-warning" "success log replay output"

set +e
short_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
    "$HELPER" "short failure" -- bash -lc 'echo short-failure >&2; exit 7' \
    2>&1
)"
short_failure_status=$?
set -e
if [[ "$short_failure_status" -ne 7 ]]; then
  fail "short failure: expected exit status 7, got $short_failure_status"
fi
assert_contains "$short_failure_output" "== short failure ==" "short failure banner"
assert_contains "$short_failure_output" "phase failed: short failure" "short failure label"
assert_contains "$short_failure_output" "failing command: bash -lc" "short failure command"
assert_contains "$short_failure_output" "short-failure" "short failure output"
short_failure_log="$(extract_log_path "$short_failure_output")"
if [[ -z "$short_failure_log" || ! -f "$short_failure_log" ]]; then
  fail "short failure: expected a preserved phase log path"
fi
cleanup_paths+=("$short_failure_log")

set +e
long_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
    "$HELPER" "long failure" -- bash -lc 'for i in $(seq 1 230); do echo "line-$i"; done; exit 9' \
    2>&1
)"
long_failure_status=$?
set -e
if [[ "$long_failure_status" -ne 9 ]]; then
  fail "long failure: expected exit status 9, got $long_failure_status"
fi
assert_contains "$long_failure_output" "== long failure ==" "long failure banner"
assert_contains "$long_failure_output" "line-1" "long failure first excerpt"
assert_contains "$long_failure_output" "line-40" "long failure first excerpt boundary"
assert_not_contains "$long_failure_output" "line-70" "long failure middle omission"
assert_contains "$long_failure_output" "line-71" "long failure last excerpt start"
assert_contains "$long_failure_output" "line-230" "long failure last excerpt end"
long_failure_log="$(extract_log_path "$long_failure_output")"
if [[ -z "$long_failure_log" || ! -f "$long_failure_log" ]]; then
  fail "long failure: expected a preserved phase log path"
fi
cleanup_paths+=("$long_failure_log")

verbose_override_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  VERBOSE=1 \
    "$HELPER" "verbose override" -- bash -lc 'echo verbose-stream'
)"
assert_contains "$verbose_override_output" "== verbose override ==" "verbose override banner"
assert_contains "$verbose_override_output" "verbose-stream" "verbose override output"
