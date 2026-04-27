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

if [[ -n "${FAKE_PLAYWRIGHT_INVOCATIONS:-}" ]]; then
  project="(none)"
  previous=""
  for arg in "$@"; do
    if [[ "$previous" == "--project" ]]; then
      project="$arg"
      break
    fi
    previous="$arg"
  done
  printf 'project=%s files=%s\n' "$project" "${CARTULARY_PLAYWRIGHT_FUNCTIONAL_FILES//$'\n'/,}" >>"$FAKE_PLAYWRIGHT_INVOCATIONS"
fi

node - "$output_file" "${FAKE_PLAYWRIGHT_MODE:-success}" "$@" <<'NODE'
const fs = require("node:fs");
const path = require("node:path");

const [outputFile, mode, ...args] = process.argv.slice(2);
const root = process.cwd();
const specs = [];
const baseTimeMs = Date.parse("2026-04-24T00:00:00.000Z");
let timingIndex = 0;
const projectIndex = args.indexOf("--project");
const project = projectIndex === -1 ? "functional" : args[projectIndex + 1];
const functionalFiles = new Set(
  (process.env.CARTULARY_PLAYWRIGHT_FUNCTIONAL_FILES ?? "")
    .split(/\r?\n/u)
    .map((file) => file.trim().replace(/^apps\/web\/e2e\//u, ""))
    .filter(Boolean),
);
const functionalGrep = new RegExp(process.env.CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP ?? ".*");

function fakeResult(status, extra = {}) {
  const duration = 100 + timingIndex;
  const startTime = new Date(baseTimeMs + timingIndex * 1000).toISOString();
  timingIndex += 1;
  return {
    status,
    retry: 0,
    duration,
    startTime,
    attachments: [],
    errors: [],
    ...extra,
  };
}

if (project === "functional") {
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
      const file = entry.file.replace(/^apps\/web\/e2e\//, "");
      if (functionalFiles.size > 0 && !functionalFiles.has(file)) {
        continue;
      }
      if (!functionalGrep.test(entry.title)) {
        continue;
      }
      if (mode === "mismatch" && (entry.id === "E-2-02" || entry.id === "E-2-03")) {
        continue;
      }
      specs.push({
        title: entry.title,
        file,
        tests: [{ results: [fakeResult("passed")] }],
      });
    }
  }
}

if (project === "support") {
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
                    ...fakeResult("failed"),
                    error: { message: "support assertion failed" },
                  }
                : fakeResult("passed"),
            ],
          },
        ],
      });
    }
  }
}

fs.writeFileSync(outputFile, `${JSON.stringify({ suites: [{ specs, suites: [] }], errors: [] })}\n`);
NODE

if [[ "${FAKE_PLAYWRIGHT_MODE:-success}" == "support-failure" ]]; then
  exit 1
fi
EOF
chmod +x "$fake_playwright"

success_invocations="$tmp_dir/batch-success-invocations.log"
success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="batch-success" \
  NODE_BIN="${NODE:-node}" \
  FAKE_PLAYWRIGHT_INVOCATIONS="$success_invocations" \
    "$HELPER" webserver-backed -- "$fake_playwright"
)"
assert_empty "$success_output" "playwright webserver batch success"
assert_contains "$(cat "$success_invocations")" "project=functional" "functional shard invocation"
assert_contains "$(cat "$success_invocations")" "project=support" "support project invocation"
"${NODE:-node}" - "$success_invocations" <<'NODE'
const fs = require("node:fs");
const lines = fs.readFileSync(process.argv[2], "utf8").trim().split(/\n/u).filter(Boolean);
const functional = lines.filter((line) => line.startsWith("project=functional "));
if (functional.length < 2) {
  throw new Error(`expected at least two functional shard invocations, got ${functional.length}`);
}
if (!functional.some((line) => line.includes("phase4.workbook.spec.ts"))) {
  throw new Error("expected a functional shard containing phase4.workbook.spec.ts");
}
if (!functional.some((line) => line.includes("phase4.mentions.spec.ts"))) {
  throw new Error("expected a functional shard containing phase4.mentions.spec.ts");
}
if (functional.some((line) =>
  line.includes("phase4.autoresolve.spec.ts") &&
  line.includes("phase4.mentions.spec.ts") &&
  line.includes("phase4.merge.spec.ts") &&
  line.includes("phase4.workbook.spec.ts")
)) {
  throw new Error("phase4 specs were not split across duration-balanced shards");
}
NODE
success_root="$tmp_dir/results/batch-success/adhoc"
phase1_summary="$success_root/browser-e2e-functional-phase1-authoritative/phase-summary.json"
phase2_summary="$success_root/browser-e2e-functional-phase2-authoritative/phase-summary.json"
phase4_summary="$success_root/browser-e2e-functional-phase4-authoritative/phase-summary.json"
support_summary="$success_root/browser-e2e-support-raw/phase-summary.json"
assert_equals "$(json_field "$phase1_summary" "status")" "pass" "phase1 batch success status"
assert_equals "$(json_field "$phase1_summary" "accounting_mode")" "actual" "phase1 batch accounting"
assert_equals "$(json_field "$phase2_summary" "status")" "pass" "phase2 batch success status"
assert_equals "$(json_field "$phase2_summary" "accounting_mode")" "actual" "phase2 batch accounting"
assert_equals "$(json_field "$phase4_summary" "accounting_mode")" "actual" "phase4 batch accounting"
assert_equals "$(json_field "$support_summary" "status")" "pass" "support batch success status"
assert_equals "$(json_field "$support_summary" "counts.support")" "6" "support batch support count"
phase1_timing="$success_root/browser-e2e-functional-phase1-authoritative/playwright-timing.json"
phase4_timing="$success_root/browser-e2e-functional-phase4-authoritative/playwright-timing.json"
assert_equals "$(json_field "$phase1_timing" "source")" "playwright_result_timestamps" "phase1 timing source"
assert_equals "$(json_field "$phase1_timing" "phase")" "phase1" "phase1 timing phase"
assert_equals "$(json_field "$phase4_timing" "files.0.file")" "apps/web/e2e/phase4.autoresolve.spec.ts" "phase4 timing first file"
assert_equals "$(json_field "$phase4_timing" "files.1.file")" "apps/web/e2e/phase4.mentions.spec.ts" "phase4 timing second file"
assert_equals "$(json_field "$phase4_timing" "files.2.file")" "apps/web/e2e/phase4.merge.spec.ts" "phase4 timing third file"
assert_equals "$(json_field "$phase4_timing" "files.3.file")" "apps/web/e2e/phase4.workbook.spec.ts" "phase4 timing fourth file"
NODE_BIN="${NODE:-node}" CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" CARTULARY_TEST_RUN_ID="batch-success" \
  "$ROOT_DIR/scripts/lib/test-output.sh" target-summary adhoc pass >/dev/null
success_target_summary="$success_root/target-summary.json"
assert_equals "$(json_field "$success_target_summary" "accounting_modes.actual")" "4" "batch target actual phase count"
assert_equals "$(json_field "$success_target_summary" "accounting_modes.derived")" "1" "batch target derived phase count"

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

batch_manifest="$ROOT_DIR/tools/browser_e2e_batch_manifest.json"
batch_runner="$ROOT_DIR/scripts/run-browser-e2e-batch.sh"
batch_manifest_summary="$("${NODE:-node}" - "$batch_manifest" <<'NODE'
const fs = require("node:fs");

const manifest = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const stages = new Map((manifest.stages ?? []).map((stage) => [stage.name, stage]));
for (const required of ["webserver-backed", "isolated"]) {
  if (!stages.has(required)) {
    throw new Error(`missing browser batch stage ${required}`);
  }
}
if (stages.has("all")) {
  throw new Error("browser batch manifest must not keep removed all stage");
}
const isolated = stages.get("isolated");
const isolatedTargets = isolated.groups.map((group) => group.target).join(",");
const resetLabels = [
  ...isolated.groups.map((group) => group.reset_before ?? ""),
].filter(Boolean).join(",");
process.stdout.write(`${isolatedTargets}\n${resetLabels}\n`);
NODE
)"
assert_contains "$batch_manifest_summary" "browser-e2e-stateful,browser-e2e-measurement,browser-e2e-visual" "isolated batch targets"
assert_contains "$batch_manifest_summary" "stateful-to-measurement" "isolated batch stateful reset"
assert_contains "$batch_manifest_summary" "measurement-to-visual" "isolated batch visual reset"
assert_contains "$(cat "$batch_runner")" 'target-summary "$target"' "batch runner child summary"
assert_contains "$(cat "$batch_runner")" "reset-web-e2e-stack.sh" "batch runner reset boundary"
