#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_HELPER="${NODE_BIN:-node}"
MAKE_HELPER="${MAKE:-make}"
PLAN_SCRIPT="$ROOT_DIR/scripts/print-target-plan.mjs"
cleanup_paths=()

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
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

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/target-plan-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")

json_a="$tmp_dir/target-plan-a.json"
json_b="$tmp_dir/target-plan-b.json"
"$NODE_HELPER" "$PLAN_SCRIPT" --json >"$json_a"
"$NODE_HELPER" "$PLAN_SCRIPT" --json >"$json_b"

"$NODE_HELPER" -e 'JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8"))' "$json_a"
cmp -s "$json_a" "$json_b" || fail "target-plan JSON must be deterministic across invocations"

backend_store_output="$("$NODE_HELPER" "$PLAN_SCRIPT" --target backend-store)"
for phase in phase1 phase2 phase3 phase4; do
  assert_contains "$backend_store_output" "$phase unit authoritative" "backend-store target plan"
done

results_dir="$tmp_dir/results"
make_output="$(
  CARTULARY_TEST_RESULTS_DIR="$results_dir" \
    "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store
)"
assert_contains "$make_output" "backend-store" "make explain-target"
if [[ -d "$results_dir" ]] && [[ -n "$(find "$results_dir" -mindepth 1 -print -quit)" ]]; then
  fail "make explain-target must not create test report artifacts"
fi
