#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-vitest-manifest-phase.sh"
cleanup_paths=()

unset VERBOSE CI_VERBOSE CARTULARY_OUTPUT_MODE

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

assert_empty() {
  local value="$1"
  local label="$2"

  if [[ -n "$value" ]]; then
    fail "$label: expected no output, got [$value]"
  fi
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "$actual" != "$expected" ]]; then
    fail "$label: expected [$expected], got [$actual]"
  fi
}

json_field() {
  local file="$1"
  local path="$2"

  "${NODE:-node}" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
if (value === undefined || value === null) {
  process.exit(1);
}
process.stdout.write(String(value));
' "$file" "$path"
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/run-vitest-manifest-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")
fake_vitest="$tmp_dir/fake-vitest.sh"
cat >"$fake_vitest" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

output_file=""
for arg in "$@"; do
  case "$arg" in
    --outputFile=*)
      output_file="${arg#--outputFile=}"
      ;;
    --outputFile.json=*)
      output_file="${arg#--outputFile.json=}"
      ;;
  esac
done

if [[ -z "$output_file" ]]; then
  echo "missing output file" >&2
  exit 2
fi

mkdir -p "$(dirname "$output_file")"

case "${FAKE_VITEST_MODE:-success}" in
  success)
    # This fixture must mirror the Phase 3 authoritative Vitest manifest.
    cat >"$output_file" <<'JSON'
{"numTotalTestSuites":3,"numPassedTestSuites":3,"numFailedTestSuites":0,"numPendingTestSuites":0,"numTotalTests":6,"numPassedTests":6,"numFailedTests":0,"numPendingTests":0,"numTodoTests":0,"success":true,"testResults":[{"assertionResults":[{"ancestorTitles":["Phase 3 Timeline workbook autosave coverage"],"fullName":"Phase 3 Timeline workbook autosave coverage Phase 3 U-3-05 autosaves on Enter, Tab, blur, and paste completion without a Save button and keeps exact save-state labels","status":"passed","title":"Phase 3 U-3-05 autosaves on Enter, Tab, blur, and paste completion without a Save button and keeps exact save-state labels","failureMessages":[],"meta":{},"tags":[]}],"status":"passed","message":"","name":"/home/askahn/code/cartulary/apps/web/src/WorkbookShell.phase3.autosave.test.tsx"},{"assertionResults":[{"ancestorTitles":["Phase 3 Timeline workbook payload coverage"],"fullName":"Phase 3 Timeline workbook payload coverage Phase 3 U-3-12 builds zero-field Timeline create payloads only for explicit blank-row creation","status":"passed","title":"Phase 3 U-3-12 builds zero-field Timeline create payloads only for explicit blank-row creation","failureMessages":[],"meta":{},"tags":[]},{"ancestorTitles":["Phase 3 Timeline workbook payload coverage"],"fullName":"Phase 3 Timeline workbook payload coverage Phase 3 U-3-13 creates an explicit blank Timeline row with only client_txn_id and suppresses duplicate pending submits","status":"passed","title":"Phase 3 U-3-13 creates an explicit blank Timeline row with only client_txn_id and suppresses duplicate pending submits","failureMessages":[],"meta":{},"tags":[]}],"status":"passed","message":"","name":"/home/askahn/code/cartulary/apps/web/src/WorkbookShell.phase3.payload.test.tsx"},{"assertionResults":[{"ancestorTitles":["Phase 3 Timeline workbook grid coverage"],"fullName":"Phase 3 Timeline workbook grid coverage Phase 3 U-3-GRID-01 binds Timeline grid columns from the active view_schema and commits writable cells by field_key","status":"passed","title":"Phase 3 U-3-GRID-01 binds Timeline grid columns from the active view_schema and commits writable cells by field_key","failureMessages":[],"meta":{},"tags":[]},{"ancestorTitles":["Phase 3 Timeline workbook grid coverage"],"fullName":"Phase 3 Timeline workbook grid coverage Phase 3 U-3-GRID-02 binds saved rows by record_id and row_version instead of visible row index","status":"passed","title":"Phase 3 U-3-GRID-02 binds saved rows by record_id and row_version instead of visible row index","failureMessages":[],"meta":{},"tags":[]},{"ancestorTitles":["Phase 3 Timeline workbook grid coverage"],"fullName":"Phase 3 Timeline workbook grid coverage Phase 3 U-3-GRID-03 keeps sorted and filtered local edits bound to the original record_id, base_row_version, and field_key","status":"passed","title":"Phase 3 U-3-GRID-03 keeps sorted and filtered local edits bound to the original record_id, base_row_version, and field_key","failureMessages":[],"meta":{},"tags":[]}],"status":"passed","message":"","name":"/home/askahn/code/cartulary/apps/web/src/WorkbookShell.phase3.grid.test.tsx"}]}
JSON
    ;;
  mismatch)
    cat >"$output_file" <<'JSON'
{"numTotalTestSuites":1,"numPassedTestSuites":1,"numFailedTestSuites":0,"numPendingTestSuites":0,"numTotalTests":1,"numPassedTests":1,"numFailedTests":0,"numPendingTests":0,"numTodoTests":0,"success":true,"testResults":[{"assertionResults":[{"ancestorTitles":["Phase 3 Timeline workbook autosave coverage"],"fullName":"Phase 3 Timeline workbook autosave coverage wrong title","status":"passed","title":"Phase 3 support wrong title","failureMessages":[],"meta":{},"tags":[]}],"status":"passed","message":"","name":"/home/askahn/code/cartulary/apps/web/src/WorkbookShell.phase3.autosave.test.tsx"}]}
JSON
    ;;
  suite_load_failure)
    cat >"$output_file" <<'JSON'
{"numTotalTestSuites":1,"numPassedTestSuites":0,"numFailedTestSuites":1,"numPendingTestSuites":0,"numTotalTests":0,"numPassedTests":0,"numFailedTests":0,"numPendingTests":0,"numTodoTests":0,"success":false,"testResults":[{"assertionResults":[],"status":"failed","message":"ReferenceError: window is not defined","name":"/home/askahn/code/cartulary/apps/web/src/WorkbookShell.phase3.autosave.test.tsx"}]}
JSON
    exit 1
    ;;
  timeout)
    sleep 30
    ;;
  *)
    echo "unsupported fake vitest mode ${FAKE_VITEST_MODE}" >&2
    exit 2
    ;;
esac

echo "JSON report written to $output_file"
EOF
chmod +x "$fake_vitest"

success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  NODE_BIN="${NODE:-node}" \
    "$HELPER" "vitest manifest smoke" phase3 authoritative frontend_unit -- "$fake_vitest"
)"
assert_empty "$success_output" "vitest manifest success"

set +e
mismatch_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  NODE_BIN="${NODE:-node}" \
  FAKE_VITEST_MODE=mismatch \
    "$HELPER" "vitest manifest mismatch" phase3 authoritative frontend_unit -- "$fake_vitest" \
    2>&1
)"
mismatch_status=$?
set -e

if [[ "$mismatch_status" -eq 0 ]]; then
  fail "vitest manifest mismatch: expected non-zero exit status"
fi
assert_contains "$mismatch_output" "manifest mismatch: vitest manifest mismatch" "vitest manifest mismatch label"
assert_contains "$mismatch_output" "missing_ids=U-3-05,U-3-12,U-3-13,U-3-GRID-01,U-3-GRID-02,U-3-GRID-03" "vitest manifest missing id"

suite_load_results="$tmp_dir/results"
set +e
suite_load_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$suite_load_results" \
  CARTULARY_TEST_RUN_ID="suite-load" \
  NODE_BIN="${NODE:-node}" \
  FAKE_VITEST_MODE=suite_load_failure \
    "$HELPER" "vitest manifest suite load" phase3 authoritative frontend_unit -- "$fake_vitest" \
    2>&1
)"
suite_load_status=$?
set -e

if [[ "$suite_load_status" -eq 0 ]]; then
  fail "vitest manifest suite load: expected non-zero exit status"
fi
assert_contains "$suite_load_output" "failure: vitest manifest suite load" "vitest manifest suite load label"
assert_contains "$suite_load_output" "symbol_or_title=(suite load)" "vitest manifest suite load title"
assert_contains "$suite_load_output" "message=ReferenceError: window is not defined" "vitest manifest suite load message"
phase_summary="$suite_load_results/suite-load/adhoc/vitest-manifest-suite-load/phase-summary.json"
assert_equals "$(json_field "$phase_summary" "counts.failed")" "1" "vitest manifest suite load failed count"
assert_equals "$(json_field "$phase_summary" "counts.authoritative_failed")" "1" "vitest manifest suite load authoritative failed count"

timeout_results="$tmp_dir/results-timeout"
set +e
timeout_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$timeout_results" \
  CARTULARY_TEST_RUN_ID="timeout" \
  CARTULARY_VITEST_WATCHDOG_SECONDS=1 \
  CARTULARY_WATCHDOG_KILL_GRACE_SECONDS=1 \
  NODE_BIN="${NODE:-node}" \
  FAKE_VITEST_MODE=timeout \
    "$HELPER" "vitest manifest timeout" phase3 authoritative frontend_unit -- "$fake_vitest" \
    2>&1
)"
timeout_status=$?
set -e

if [[ "$timeout_status" -eq 0 ]]; then
  fail "vitest manifest timeout: expected non-zero exit status"
fi
assert_contains "$timeout_output" "failure: vitest manifest timeout" "vitest manifest timeout failure label"
assert_contains "$timeout_output" "coverage=non_test" "vitest manifest timeout non-test coverage"
assert_contains "$timeout_output" "vitest watchdog timed out before runner.json was written" "vitest manifest timeout message"
timeout_summary="$timeout_results/timeout/adhoc/vitest-manifest-timeout/phase-summary.json"
assert_equals "$(json_field "$timeout_summary" "counts.failed")" "1" "vitest manifest timeout failed count"
assert_equals "$(json_field "$timeout_summary" "counts.non_test_failed")" "1" "vitest manifest timeout non-test failed count"
assert_contains "$(json_field "$timeout_summary" "artifacts.watchdog_json")" "watchdog.json" "vitest manifest timeout watchdog artifact"
