#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/run-frontend-unit.sh"
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

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/run-frontend-unit-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")
runtime_dir="$tmp_dir/runtime"
mkdir -p "$runtime_dir/bin"
ln -s "$(command -v "${NODE:-node}")" "$runtime_dir/bin/node"

fake_pnpm="$runtime_dir/bin/pnpm"
cat >"$fake_pnpm" <<'EOF'
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

"${NODE_BIN:-node}" - "$output_file" "${FAKE_FRONTEND_UNIT_MODE:-success}" <<'NODE'
const fs = require("node:fs");
const path = require("node:path");

const [outputFile, mode] = process.argv.slice(2);
const root = process.cwd();
const phaseFiles = ["phase1", "phase2", "phase3"];
const authoritative = [];

for (const phase of phaseFiles) {
  const manifest = JSON.parse(fs.readFileSync(path.join(root, "tools", `${phase}_test_map.json`), "utf8"));
  for (const row of manifest.unit ?? []) {
    if (
      row.runner === "vitest" &&
      row.coverage === "authoritative" &&
      row.execution_dependency === "frontend_unit"
    ) {
      authoritative.push(row);
    }
  }
}

const statusFor = (index, fallback = "passed") => {
  if (mode === "authoritative-failure" && index === 0) {
    return "failed";
  }
  return fallback;
};

const assertion = (title, status) => ({
  ancestorTitles: ["frontend-unit smoke"],
  fullName: `frontend-unit smoke ${title}`,
  status,
  title,
  failureMessages: status === "failed" ? ["frontend unit smoke failure"] : [],
  meta: {},
  tags: [],
});

const byFile = new Map();
for (const [index, row] of authoritative.entries()) {
  const absolute = path.join(root, row.file);
  const entries = byFile.get(absolute) ?? [];
  entries.push(assertion(row.title, statusFor(index)));
  byFile.set(absolute, entries);
}

byFile.set(path.join(root, "apps/web/src/App.phase1.support.test.tsx"), [
  assertion("Phase 1 support smoke keeps ordinary shell helpers stable", "passed"),
]);
byFile.set(path.join(root, "apps/web/src/Unmapped.frontend-unit.test.tsx"), [
  assertion(
    "unmapped frontend residual smoke",
    mode === "residual-failure" ? "failed" : "passed",
  ),
]);

const testResults = [...byFile.entries()].map(([name, assertionResults]) => {
  const failed = assertionResults.some((entry) => entry.status === "failed");
  return {
    assertionResults,
    status: failed ? "failed" : "passed",
    message: failed ? "frontend unit smoke failure" : "",
    name,
  };
});
const tests = testResults.flatMap((entry) => entry.assertionResults);
const failedTests = tests.filter((entry) => entry.status === "failed");
fs.writeFileSync(outputFile, `${JSON.stringify({
  numTotalTestSuites: testResults.length,
  numPassedTestSuites: testResults.filter((entry) => entry.status === "passed").length,
  numFailedTestSuites: testResults.filter((entry) => entry.status === "failed").length,
  numPendingTestSuites: 0,
  numTotalTests: tests.length,
  numPassedTests: tests.length - failedTests.length,
  numFailedTests: failedTests.length,
  numPendingTests: 0,
  numTodoTests: 0,
  success: failedTests.length === 0,
  testResults,
})}\n`);
NODE

if [[ "${FAKE_FRONTEND_UNIT_MODE:-success}" == *failure ]]; then
  exit 1
fi
EOF
chmod +x "$fake_pnpm"

run_case() {
  local name="$1"
  local mode="$2"
  local expected_status="$3"
  local results_dir="$tmp_dir/results-$name"
  local stdout_log="$tmp_dir/$name.stdout.log"
  local stderr_log="$tmp_dir/$name.stderr.log"

  set +e
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_TARGET=frontend-unit \
  CARTULARY_TEST_RESULTS_DIR="$results_dir" \
  CARTULARY_TEST_RUN_ID="$name" \
  NODE_RUNTIME_DIR="$runtime_dir" \
  NODE_BIN="$runtime_dir/bin/node" \
  PNPM="$fake_pnpm" \
  FAKE_FRONTEND_UNIT_MODE="$mode" \
    "$HELPER" >"$stdout_log" 2>"$stderr_log"
  local status=$?
  set -e

  if [[ "$expected_status" == "pass" && "$status" -ne 0 ]]; then
    cat "$stderr_log" >&2
    fail "$name: expected pass, got status $status"
  fi
  if [[ "$expected_status" == "fail" && "$status" -eq 0 ]]; then
    fail "$name: expected fail"
  fi

  printf '%s\n' "$results_dir/$name/frontend-unit/target-summary.json"
}

success_summary="$(run_case success success pass)"
assert_equals "$(json_field "$success_summary" "own.counts.tests")" "15" "success total tests"
assert_equals "$(json_field "$success_summary" "own.counts.authoritative")" "13" "success authoritative count"
assert_equals "$(json_field "$success_summary" "own.counts.support")" "1" "success support count"
assert_equals "$(json_field "$success_summary" "own.counts.unmapped")" "1" "success unmapped count"
assert_equals "$(json_field "$success_summary" "own.accounting_modes.actual")" "1" "success raw actual phase"
assert_equals "$(json_field "$success_summary" "own.accounting_modes.derived")" "4" "success derived slices"

residual_summary="$(run_case residual residual-failure fail)"
assert_equals "$(json_field "$residual_summary" "own.counts.failed")" "1" "residual failure count"
assert_equals "$(json_field "$residual_summary" "own.counts.unmapped_failed")" "1" "residual unmapped failure count"
assert_equals "$(json_field "$residual_summary" "own.counts.authoritative_failed")" "0" "residual authoritative failure count"

authoritative_summary="$(run_case authoritative authoritative-failure fail)"
assert_equals "$(json_field "$authoritative_summary" "own.counts.failed")" "1" "authoritative failure count"
assert_equals "$(json_field "$authoritative_summary" "own.counts.authoritative_failed")" "1" "authoritative authoritative failure count"
assert_equals "$(json_field "$authoritative_summary" "own.counts.unmapped_failed")" "0" "authoritative unmapped failure count"
