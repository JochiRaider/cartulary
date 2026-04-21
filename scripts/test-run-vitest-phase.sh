#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-vitest-phase.sh"
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

assert_empty() {
  local value="$1"
  local label="$2"

  if [[ -n "$value" ]]; then
    fail "$label: expected no output, got [$value]"
  fi
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
