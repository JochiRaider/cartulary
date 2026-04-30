#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
MAKE_HELPER="${MAKE:-make}"
TASK_GUIDE="$ROOT_DIR/scripts/print-task-guide.mjs"
EXPLAIN_PHASE="$ROOT_DIR/scripts/print-explain-phase.mjs"
EXPLAIN_TARGET="$ROOT_DIR/scripts/print-explain-target.mjs"
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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle]"
  fi
}

assert_fails() {
  local label="$1"
  shift

  local output
  local status
  set +e
  output="$("$@" 2>&1)"
  status=$?
  set -e

  if [[ "$status" -eq 0 ]]; then
    fail "$label: expected failure"
  fi
  printf '%s' "$output"
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/task-guidance.XXXXXX")"
cleanup_paths+=("$tmp_dir")
results_dir="$tmp_dir/results"
mkdir -p "$results_dir/run-a/backend-store"
printf '{"target":"backend-store","status":"pass"}\n' >"$results_dir/run-a/backend-store/target-summary.json"
mkdir -p "$results_dir/run-b" "$results_dir/run-c" "$results_dir/run-d/frontend-unit" "$results_dir/run-e/frontend-unit"
printf '{"label":"check","status":"pass"}\n' >"$results_dir/run-b/run-summary.json"
printf '{"label":"ci","status":"pass"}\n' >"$results_dir/run-c/run-summary.json"
printf '{"target":"not-frontend-unit","status":"pass"}\n' >"$results_dir/run-d/frontend-unit/target-summary.json"
printf '{"target":"frontend-unit","status":"pass"}\n' >"$results_dir/run-e/frontend-unit/target-summary.json"
expected_results_files="$(find "$results_dir" -type f | wc -l | tr -d '[:space:]')"

default_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$TASK_GUIDE")"
assert_contains "$default_output" "Cartulary task guide" "default task-guide header"
assert_contains "$default_output" "local-dev:" "default task-guide local-dev role"
assert_contains "$default_output" "feature-dev:" "default task-guide feature-dev role"
assert_contains "$default_output" "latest_artifact=none" "default task-guide reports missing artifacts"

for role in local-dev feature-dev phase-author ci-investigator release; do
  role_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$TASK_GUIDE" --role "$role")"
  assert_contains "$role_output" "role=$role" "task-guide role header"
  assert_contains "$role_output" "$role:" "task-guide role section"
done

phase_role_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$TASK_GUIDE" --role phase-author --phase phase1)"
assert_contains "$phase_role_output" "phase focus: phase1" "task-guide phase focus"
assert_contains "$phase_role_output" "make explain-phase PHASE=phase1" "task-guide phase command"

guide_json="$tmp_dir/task-guide.json"
"$NODE_BIN" "$TASK_GUIDE" --role feature-dev --json >"$guide_json"
"$NODE_BIN" -e 'JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8"))' "$guide_json"

phase_output="$("$NODE_BIN" "$EXPLAIN_PHASE" --phase phase1)"
assert_contains "$phase_output" "Cartulary phase guidance: phase1" "explain-phase header"
assert_contains "$phase_output" "targets:" "explain-phase targets"
assert_contains "$phase_output" "make backend-store" "explain-phase backend-store target"
assert_contains "$phase_output" "ledger: docs/testing/phase1_coverage_ledger.md" "explain-phase ledger"

phase_json="$tmp_dir/explain-phase.json"
"$NODE_BIN" "$EXPLAIN_PHASE" --phase phase1 --json >"$phase_json"
"$NODE_BIN" - "$phase_json" <<'EOF'
const fs = require("node:fs");
const phase = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
if (phase.phase !== "phase1" || !Array.isArray(phase.targets) || phase.targets.length === 0) {
  process.exit(1);
}
EOF

unknown_phase_output="$(assert_fails "unknown phase" "$NODE_BIN" "$EXPLAIN_PHASE" --phase phase99)"
assert_contains "$unknown_phase_output" "unknown phase phase99" "unknown phase error"

backend_summary="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$EXPLAIN_TARGET" --target backend-store)"
assert_contains "$backend_summary" "Cartulary target guidance: backend-store" "backend-store explain-target header"
assert_contains "$backend_summary" "services: Postgres,MinIO" "backend-store service requirements"
assert_contains "$backend_summary" "latest_artifact: tmp/task-guidance" "backend-store latest artifact"

backend_rows="$("$NODE_BIN" "$EXPLAIN_TARGET" --target backend-store --detail rows)"
assert_contains "$backend_rows" "rows:" "backend-store rows"
assert_contains "$backend_rows" "U-1-05" "backend-store Go rows"

frontend_output="$("$NODE_BIN" "$EXPLAIN_TARGET" --target frontend-unit)"
assert_contains "$frontend_output" "Cartulary target guidance: frontend-unit" "frontend-unit explain-target"
assert_contains "$frontend_output" "phase_coverage:" "frontend-unit phase coverage"

frontend_identity_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$EXPLAIN_TARGET" --target frontend-unit)"
assert_contains "$frontend_identity_output" "latest_artifact: tmp/task-guidance" "matching target summary latest artifact"
assert_not_contains "$frontend_identity_output" "run-d/frontend-unit/target-summary.json" "mismatched target summary ignored"

test_fast_mismatch_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$EXPLAIN_TARGET" --target test-fast)"
assert_contains "$test_fast_mismatch_output" "latest_artifact: none" "mismatched check run summary ignored for test-fast"

release_mismatch_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$EXPLAIN_TARGET" --target release-check)"
assert_contains "$release_mismatch_output" "latest_artifact: none" "mismatched check run summary ignored for release-check"

ci_match_output="$(CARTULARY_TEST_RESULTS_DIR="$results_dir" "$NODE_BIN" "$EXPLAIN_TARGET" --target ci)"
assert_contains "$ci_match_output" "run-c/run-summary.json" "matching ci run summary accepted"

browser_output="$("$NODE_BIN" "$EXPLAIN_TARGET" --target browser-e2e-webserver-backed)"
assert_contains "$browser_output" "browser stack" "browser explain-target service requirements"
assert_contains "$browser_output" "webserver-backed browser batch" "browser explain-target scheduler"

check_output="$("$NODE_BIN" "$EXPLAIN_TARGET" --target check)"
assert_contains "$check_output" "check scheduler" "check explain-target scheduler"
assert_contains "$check_output" "phase_coverage:" "check explain-target phase coverage"

target_artifacts="$("$NODE_BIN" "$EXPLAIN_TARGET" --target backend-store --detail artifacts)"
assert_contains "$target_artifacts" "expected:" "target artifact expected paths"
assert_contains "$target_artifacts" "<run-id>/backend-store/target-summary.json" "target artifact expected summary"

target_json="$tmp_dir/explain-target.json"
"$NODE_BIN" "$EXPLAIN_TARGET" --target browser-e2e-webserver-backed --json >"$target_json"
"$NODE_BIN" -e 'JSON.parse(require("node:fs").readFileSync(process.argv[1], "utf8"))' "$target_json"

unknown_target_output="$(assert_fails "unknown target" "$NODE_BIN" "$EXPLAIN_TARGET" --target no-such-target)"
assert_contains "$unknown_target_output" "unknown target no-such-target" "unknown target error"

make_task_guide="$("$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" task-guide ROLE=feature-dev)"
assert_contains "$make_task_guide" "role=feature-dev" "make task-guide role"

make_phase="$("$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-phase PHASE=phase1)"
assert_contains "$make_phase" "Cartulary phase guidance: phase1" "make explain-phase"

make_target="$(
  CARTULARY_TEST_RESULTS_DIR="$results_dir" \
    "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store DETAIL=rows
)"
assert_contains "$make_target" "U-1-05" "make explain-target rows"

detail_zero_output="$(assert_fails "obsolete DETAIL=0" "$MAKE_HELPER" --no-print-directory -C "$ROOT_DIR" explain-target TARGET=backend-store DETAIL=0)"
assert_contains "$detail_zero_output" "usage: print-explain-target.mjs" "DETAIL=0 rejected"

if [[ -d "$results_dir" ]] && [[ "$(find "$results_dir" -type f | wc -l | tr -d '[:space:]')" != "$expected_results_files" ]]; then
  fail "guidance commands must not create test report artifacts"
fi
