#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-playwright-webserver-batch.sh"
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

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "$actual" != "$expected" ]]; then
    fail "$label: expected [$expected], got [$actual]"
  fi
}

assert_empty() {
  local value="$1"
  local label="$2"

  if [[ -n "$value" ]]; then
    fail "$label: expected no output, got [$value]"
  fi
}

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/run-playwright-webserver-batch.XXXXXX")"
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

node - "$output_file" "${FAKE_PLAYWRIGHT_MODE:-success}" <<'NODE'
const fs = require("node:fs");
const path = require("node:path");

const [outputFile, mode] = process.argv.slice(2);
const root = process.cwd();
const specs = [];

for (const phase of ["phase1", "phase2", "phase3", "phase4"]) {
  const manifest = JSON.parse(
    fs.readFileSync(path.join(root, "tools", `${phase}_test_map.json`), "utf8"),
  );
  for (const entry of manifest.e2e ?? []) {
    if (
      entry.runner !== "playwright" ||
      entry.coverage !== "authoritative" ||
      entry.execution_dependency !== "browser_functional"
    ) {
      continue;
    }
    if (mode === "mismatch" && (entry.id === "E-2-02" || entry.id === "E-2-03")) {
      continue;
    }
    specs.push({
      title: entry.title,
      file: entry.file.replace(/^apps\/web\/e2e\//, ""),
      tests: [{ results: [{ status: "passed", retry: 0, attachments: [], errors: [] }] }],
    });
  }
}

for (const supportFile of ["phase2.support.spec.ts", "phase3.support.spec.ts"]) {
  const source = fs.readFileSync(path.join(root, "apps", "web", "e2e", supportFile), "utf8");
  for (const match of source.matchAll(/\btest\("([^"]+)"/g)) {
    const failed =
      mode === "support-failure" &&
      supportFile === "phase3.support.spec.ts" &&
      match[1].includes("sort, filter, and group");
    specs.push({
      title: match[1],
      file: supportFile,
      tests: [
        {
          results: [
            failed
              ? {
                  status: "failed",
                  retry: 0,
                  attachments: [],
                  error: { message: "support assertion failed" },
                }
              : { status: "passed", retry: 0, attachments: [], errors: [] },
          ],
        },
      ],
    });
  }
}

fs.writeFileSync(outputFile, `${JSON.stringify({ suites: [{ specs, suites: [] }], errors: [] })}\n`);
NODE

if [[ "${FAKE_PLAYWRIGHT_MODE:-success}" == "support-failure" ]]; then
  exit 1
fi
EOF
chmod +x "$fake_playwright"

success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="batch-success" \
  NODE_BIN="${NODE:-node}" \
    "$HELPER" webserver-backed -- "$fake_playwright"
)"
assert_empty "$success_output" "playwright webserver batch success"
success_root="$tmp_dir/results/batch-success/adhoc"
phase1_summary="$success_root/browser-e2e-functional-phase1-authoritative/phase-summary.json"
phase2_summary="$success_root/browser-e2e-functional-phase2-authoritative/phase-summary.json"
support_summary="$success_root/browser-e2e-support-raw/phase-summary.json"
assert_equals "$(json_field "$phase1_summary" "status")" "pass" "phase1 batch success status"
assert_equals "$(json_field "$phase1_summary" "accounting_mode")" "actual" "phase1 batch accounting"
assert_equals "$(json_field "$phase2_summary" "status")" "pass" "phase2 batch success status"
assert_equals "$(json_field "$phase2_summary" "accounting_mode")" "derived" "phase2 batch accounting"
assert_equals "$(json_field "$support_summary" "status")" "pass" "support batch success status"
assert_equals "$(json_field "$support_summary" "counts.support")" "6" "support batch support count"

set +e
support_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="batch-support-failure" \
  NODE_BIN="${NODE:-node}" \
  FAKE_PLAYWRIGHT_MODE=support-failure \
    "$HELPER" webserver-backed -- "$fake_playwright" \
    2>&1
)"
support_failure_status=$?
set -e
if [[ "$support_failure_status" -eq 0 ]]; then
  fail "playwright webserver batch support failure: expected non-zero exit status"
fi
assert_contains "$support_failure_output" "failure: browser-e2e-support raw" "support failure label"
assert_contains "$support_failure_output" "coverage=support" "support failure coverage"
assert_not_contains "$support_failure_output" "coverage=unmapped" "support failure unmapped coverage"
support_failure_phase1="$tmp_dir/results/batch-support-failure/adhoc/browser-e2e-functional-phase1-authoritative/phase-summary.json"
assert_equals "$(json_field "$support_failure_phase1" "status")" "pass" "support failure leaves phase1 passing"

set +e
mismatch_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="batch-mismatch" \
  NODE_BIN="${NODE:-node}" \
  FAKE_PLAYWRIGHT_MODE=mismatch \
    "$HELPER" webserver-backed -- "$fake_playwright" \
    2>&1
)"
mismatch_status=$?
set -e
if [[ "$mismatch_status" -eq 0 ]]; then
  fail "playwright webserver batch mismatch: expected non-zero exit status"
fi
assert_contains "$mismatch_output" "manifest mismatch: browser-e2e-functional phase2 authoritative" "batch mismatch label"
assert_contains "$mismatch_output" "missing_ids=E-2-02,E-2-03" "batch mismatch missing ids"
assert_contains "$mismatch_output" "selection=" "batch mismatch selection path"
assert_contains "$mismatch_output" "runner=" "batch mismatch runner path"
