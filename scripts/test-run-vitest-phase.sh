#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-vitest-phase.sh"
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

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/run-vitest-phase-smoke.XXXXXX")"
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
    cat >"$output_file" <<'JSON'
{"numTotalTestSuites":1,"numPassedTestSuites":1,"numFailedTestSuites":0,"numPendingTestSuites":0,"numTotalTests":1,"numPassedTests":1,"numFailedTests":0,"numPendingTests":0,"numTodoTests":0,"success":true,"testResults":[{"assertionResults":[{"ancestorTitles":[],"fullName":"raw success","status":"passed","title":"raw success","failureMessages":[],"meta":{},"tags":[]}],"status":"passed","message":"","name":"/home/askahn/code/cartulary/apps/web/src/raw-success.test.ts"}]}
JSON
    ;;
  failure)
    cat >"$output_file" <<'JSON'
{"numTotalTestSuites":1,"numPassedTestSuites":0,"numFailedTestSuites":1,"numPendingTestSuites":0,"numTotalTests":1,"numPassedTests":0,"numFailedTests":1,"numPendingTests":0,"numTodoTests":0,"success":false,"testResults":[{"assertionResults":[{"ancestorTitles":[],"fullName":"raw failure","status":"failed","title":"raw failure","failureMessages":["AssertionError: expected 1 to be 2\n    at raw"],"meta":{},"tags":[]}],"status":"failed","message":"","name":"/home/askahn/code/cartulary/apps/web/src/raw-failure.test.ts"}]}
JSON
    exit 1
    ;;
  package_failure)
    cat >"$output_file" <<'JSON'
{"numTotalTestSuites":1,"numPassedTestSuites":0,"numFailedTestSuites":1,"numPendingTestSuites":0,"numTotalTests":1,"numPassedTests":0,"numFailedTests":1,"numPendingTests":0,"numTodoTests":0,"success":false,"testResults":[{"assertionResults":[{"ancestorTitles":[],"fullName":"raw package failure","status":"failed","title":"raw package failure","failureMessages":["AssertionError: package failure\n    at package"],"meta":{},"tags":[]}],"status":"failed","message":"","name":"/home/askahn/code/cartulary/packages/test-utils/src/index.test.ts"}]}
JSON
    exit 1
    ;;
  suite_load_failure)
    cat >"$output_file" <<'JSON'
{"numTotalTestSuites":1,"numPassedTestSuites":0,"numFailedTestSuites":1,"numPendingTestSuites":0,"numTotalTests":0,"numPassedTests":0,"numFailedTests":0,"numPendingTests":0,"numTodoTests":0,"success":false,"testResults":[{"assertionResults":[],"status":"failed","message":"ReferenceError: window is not defined","name":"/home/askahn/code/cartulary/apps/web/src/raw-suite-load.test.ts"}]}
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
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  NODE_BIN="${NODE:-node}" \
    "$HELPER" "vitest raw smoke" -- "$fake_vitest"
)"
assert_empty "$success_output" "vitest raw success"

set +e
failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  NODE_BIN="${NODE:-node}" \
  FAKE_VITEST_MODE=failure \
    "$HELPER" "vitest raw failure" -- "$fake_vitest" \
    2>&1
)"
failure_status=$?
set -e

if [[ "$failure_status" -eq 0 ]]; then
  fail "vitest raw failure: expected non-zero exit status"
fi
assert_contains "$failure_output" "failure: vitest raw failure" "vitest raw failure label"
assert_contains "$failure_output" "runner=vitest" "vitest raw failure runner"
assert_contains "$failure_output" "symbol_or_title=raw failure" "vitest raw failure title"
assert_contains "$failure_output" "message=AssertionError: expected 1 to be 2" "vitest raw failure message"

set +e
package_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  NODE_BIN="${NODE:-node}" \
  FAKE_VITEST_MODE=package_failure \
    "$HELPER" "vitest raw package failure" -- "$fake_vitest" \
    2>&1
)"
package_failure_status=$?
set -e

if [[ "$package_failure_status" -eq 0 ]]; then
  fail "vitest raw package failure: expected non-zero exit status"
fi
assert_contains "$package_failure_output" "package_or_file=packages/test-utils/src/index.test.ts" "vitest raw package failure owner"
assert_contains "$package_failure_output" "reproduce=pnpm --dir apps/web exec vitest run ../../packages/test-utils/src/index.test.ts -t 'raw package failure$'" "vitest raw package failure reproduce"

suite_load_results="$tmp_dir/results"
set +e
suite_load_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$suite_load_results" \
  CARTULARY_TEST_RUN_ID="suite-load" \
  NODE_BIN="${NODE:-node}" \
  FAKE_VITEST_MODE=suite_load_failure \
    "$HELPER" "vitest raw suite load" -- "$fake_vitest" \
    2>&1
)"
suite_load_status=$?
set -e

if [[ "$suite_load_status" -eq 0 ]]; then
  fail "vitest raw suite load: expected non-zero exit status"
fi
assert_contains "$suite_load_output" "failure: vitest raw suite load" "vitest raw suite load label"
assert_contains "$suite_load_output" "symbol_or_title=(suite load)" "vitest raw suite load title"
assert_contains "$suite_load_output" "message=ReferenceError: window is not defined" "vitest raw suite load message"
phase_summary="$suite_load_results/suite-load/adhoc/vitest-raw-suite-load/phase-summary.json"
assert_equals "$(json_field "$phase_summary" "counts.failed")" "1" "vitest raw suite load failed count"
assert_equals "$(json_field "$phase_summary" "counts.unmapped_failed")" "1" "vitest raw suite load unmapped failed count"

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
    "$HELPER" "vitest raw timeout" -- "$fake_vitest" \
    2>&1
)"
timeout_status=$?
set -e

if [[ "$timeout_status" -eq 0 ]]; then
  fail "vitest raw timeout: expected non-zero exit status"
fi
assert_contains "$timeout_output" "failure: vitest raw timeout" "vitest raw timeout failure label"
assert_contains "$timeout_output" "coverage=non_test" "vitest raw timeout non-test coverage"
assert_contains "$timeout_output" "vitest watchdog timed out before runner.json was written" "vitest raw timeout message"
timeout_summary="$timeout_results/timeout/adhoc/vitest-raw-timeout/phase-summary.json"
assert_equals "$(json_field "$timeout_summary" "counts.failed")" "1" "vitest raw timeout failed count"
assert_equals "$(json_field "$timeout_summary" "counts.non_test_failed")" "1" "vitest raw timeout non-test failed count"
assert_contains "$(json_field "$timeout_summary" "artifacts.watchdog_json")" "watchdog.json" "vitest raw timeout watchdog artifact"
