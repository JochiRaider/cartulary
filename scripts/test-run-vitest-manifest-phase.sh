#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-vitest-manifest-phase.sh"
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
    cat >"$output_file" <<'JSON'
{"numTotalTestSuites":2,"numPassedTestSuites":2,"numFailedTestSuites":0,"numPendingTestSuites":0,"numTotalTests":1,"numPassedTests":1,"numFailedTests":0,"numPendingTests":0,"numTodoTests":0,"success":true,"testResults":[{"assertionResults":[{"ancestorTitles":["Phase 3 Timeline workbook"],"fullName":"Phase 3 Timeline workbook Phase 3 U-3-05 autosaves on Enter, Tab, blur, and paste completion without a Save button and keeps exact save-state labels","status":"passed","title":"Phase 3 U-3-05 autosaves on Enter, Tab, blur, and paste completion without a Save button and keeps exact save-state labels","failureMessages":[],"meta":{},"tags":[]}],"status":"passed","message":"","name":"/home/askahn/code/cartulary/apps/web/src/App.test.tsx"}]}
JSON
    ;;
  mismatch)
    cat >"$output_file" <<'JSON'
{"numTotalTestSuites":1,"numPassedTestSuites":1,"numFailedTestSuites":0,"numPendingTestSuites":0,"numTotalTests":1,"numPassedTests":1,"numFailedTests":0,"numPendingTests":0,"numTodoTests":0,"success":true,"testResults":[{"assertionResults":[{"ancestorTitles":["Phase 3 Timeline workbook"],"fullName":"Phase 3 Timeline workbook wrong title","status":"passed","title":"Phase 3 support wrong title","failureMessages":[],"meta":{},"tags":[]}],"status":"passed","message":"","name":"/home/askahn/code/cartulary/apps/web/src/App.test.tsx"}]}
JSON
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
assert_contains "$mismatch_output" "missing_ids=U-3-05" "vitest manifest missing id"
