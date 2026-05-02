#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-playwright-manifest-phase.sh"
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

assert_empty() {
  local value="$1"
  local label="$2"

  if [[ -n "$value" ]]; then
    fail "$label: expected no output, got [$value]"
  fi
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/run-playwright-manifest-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")
fake_playwright="$tmp_dir/fake-playwright.sh"
cat >"$fake_playwright" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

output_file="${PLAYWRIGHT_JSON_OUTPUT_FILE:-}"
if [[ -z "$output_file" ]]; then
  echo "missing PLAYWRIGHT_JSON_OUTPUT_FILE" >&2
  exit 2
fi
mkdir -p "$(dirname "$output_file")"

if [[ " $* " == *" --list "* ]]; then
  cat >"$output_file" <<'JSON'
{"suites":[{"specs":[{"title":"E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface","tests":[{"results":[],"status":"skipped"}],"file":"phase2.spec.ts"},{"title":"E-2-02 shows incident discovery, raw querystring deep-link retrieval, and promoted-field-only patching on the ordinary incident shell","tests":[{"results":[],"status":"skipped"}],"file":"phase2.spec.ts"},{"title":"E-2-03 lets incident admins manage memberships and hides those controls from non-admin members on the ordinary shell","tests":[{"results":[],"status":"skipped"}],"file":"phase2.spec.ts"}],"suites":[]}],"errors":[]}
JSON
  exit 0
fi

case "${FAKE_PLAYWRIGHT_MODE:-success}" in
  success)
    cat >"$output_file" <<'JSON'
{"suites":[{"specs":[{"title":"E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface","file":"phase2.spec.ts","tests":[{"results":[{"status":"passed","retry":0,"attachments":[],"errors":[]}]}]},{"title":"E-2-02 shows incident discovery, raw querystring deep-link retrieval, and promoted-field-only patching on the ordinary incident shell","file":"phase2.spec.ts","tests":[{"results":[{"status":"passed","retry":0,"attachments":[],"errors":[]}]}]},{"title":"E-2-03 lets incident admins manage memberships and hides those controls from non-admin members on the ordinary shell","file":"phase2.spec.ts","tests":[{"results":[{"status":"passed","retry":0,"attachments":[],"errors":[]}]}]}],"suites":[]}],"errors":[]}
JSON
    ;;
  failure)
    cat >"$output_file" <<'JSON'
{"suites":[{"specs":[{"title":"E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface","file":"phase2.spec.ts","tests":[{"results":[{"status":"failed","retry":0,"attachments":[],"error":{"message":"playwright assertion failed"}}]}]}],"suites":[]}],"errors":[]}
JSON
    exit 1
    ;;
  mismatch)
    cat >"$output_file" <<'JSON'
{"suites":[{"specs":[{"title":"E-2-01 creates an incident, bootstraps the creator as admin, and lands on the workbook surface","file":"phase2.spec.ts","tests":[{"results":[{"status":"passed","retry":0,"attachments":[],"errors":[]}]}]}],"suites":[]}],"errors":[]}
JSON
    ;;
  *)
    echo "unsupported fake playwright mode ${FAKE_PLAYWRIGHT_MODE}" >&2
    exit 2
    ;;
esac
EOF
chmod +x "$fake_playwright"

success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="playwright-manifest-success" \
  NODE_BIN="${NODE:-node}" \
    "$HELPER" "playwright manifest success" phase2 authoritative -- "$fake_playwright"
)"
assert_empty "$success_output" "playwright manifest success"
success_summary="$tmp_dir/results/playwright-manifest-success/adhoc/playwright-manifest-success/phase-summary.json"
assert_contains "$(json_field "$success_summary" "artifacts.selection_json")" "manifest-selection.json" "playwright success selection artifact"
assert_contains "$(json_field "$success_summary" "artifacts.runner_json")" "runner.json" "playwright success runner artifact"

set +e
mismatch_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="playwright-manifest-mismatch" \
  NODE_BIN="${NODE:-node}" \
  FAKE_PLAYWRIGHT_MODE=mismatch \
    "$HELPER" "playwright manifest mismatch" phase2 authoritative -- "$fake_playwright" \
    2>&1
)"
mismatch_status=$?
set -e

if [[ "$mismatch_status" -eq 0 ]]; then
  fail "playwright manifest mismatch: expected non-zero exit status"
fi
assert_contains "$mismatch_output" "manifest mismatch: playwright manifest mismatch" "playwright manifest mismatch label"
assert_contains "$mismatch_output" "missing_ids=E-2-02,E-2-03" "playwright manifest missing ids"
assert_contains "$mismatch_output" "selection=" "playwright manifest mismatch selection path"
assert_contains "$mismatch_output" "runner=" "playwright manifest mismatch runner path"
assert_not_contains "$mismatch_output" "raw=" "playwright manifest mismatch raw path"

set +e
failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="playwright-manifest-failure" \
  NODE_BIN="${NODE:-node}" \
  FAKE_PLAYWRIGHT_MODE=failure \
    "$HELPER" "playwright manifest failure" phase2 authoritative -- "$fake_playwright" \
    2>&1
)"
failure_status=$?
set -e

if [[ "$failure_status" -eq 0 ]]; then
  fail "playwright manifest failure: expected non-zero exit status"
fi
assert_contains "$failure_output" "failure: playwright manifest failure" "playwright manifest failure label"
assert_contains "$failure_output" "selection=" "playwright manifest failure selection path"
assert_contains "$failure_output" "runner=" "playwright manifest failure runner path"
assert_contains "$failure_output" "test_runner=playwright" "playwright manifest failure runner label"
