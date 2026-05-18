#!/usr/bin/env bash
# Single-quoted literals below intentionally assert shell text without expansion.
# shellcheck disable=SC2016
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/lib/run-playwright-webserver-batch.sh"
cleanup_paths=()

unset VERBOSE CI_VERBOSE CARTULARY_OUTPUT_MODE CARTULARY_SUPPRESS_CHILD_SUCCESS

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
  selected_ids_log="${CARTULARY_MANIFEST_SELECTED_IDS:-}"
  for arg in "$@"; do
    if [[ "$previous" == "--project" ]]; then
      project="$arg"
      break
    fi
    previous="$arg"
  done
  printf 'project=%s worker_count=%s worker_offset=%s files=%s selected_ids=%s\n' \
    "$project" \
    "${CARTULARY_PLAYWRIGHT_WORKER_COUNT:-}" \
    "${CARTULARY_PLAYWRIGHT_WORKER_INDEX_OFFSET:-}" \
    "${CARTULARY_PLAYWRIGHT_FUNCTIONAL_FILES//$'\n'/,}" \
    "${selected_ids_log//$'\n'/,}" >>"$FAKE_PLAYWRIGHT_INVOCATIONS"
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
const supportFiles = new Set(
  (process.env.CARTULARY_PLAYWRIGHT_SUPPORT_FILES ?? "")
    .split(/\r?\n/u)
    .map((file) => file.trim().replace(/^apps\/web\/e2e\//u, "").replace(/^e2e\//u, ""))
    .filter(Boolean),
);
const supportGrep = new RegExp(process.env.CARTULARY_PLAYWRIGHT_SUPPORT_GREP ?? ".*");

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
  const registry = JSON.parse(fs.readFileSync(path.join(root, "tools", "phase_registry.json"), "utf8"));
  if (registry.schema_id !== "cartulary.phase_registry.v1") {
    throw new Error("tools/phase_registry.json must declare schema_id cartulary.phase_registry.v1");
  }
  const phases = (registry.phases ?? [])
    .filter((entry) => entry.status === "active")
    .sort((left, right) => left.order - right.order || left.phase.localeCompare(right.phase))
    .map((entry) => {
      const manifest = JSON.parse(fs.readFileSync(path.join(root, entry.manifest_path), "utf8"));
      if (manifest.schema_id !== "cartulary.phase_test_map.v1") {
        throw new Error(`${entry.manifest_path} must declare schema_id cartulary.phase_test_map.v1`);
      }
      if (manifest.phase !== entry.phase) {
        throw new Error(`${entry.manifest_path} must declare phase ${entry.phase}`);
      }
      return manifest.phase;
    });
  for (const phase of phases) {
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
      const titles = Array.isArray(entry.titles) ? entry.titles : [entry.title];
      const matchedTitles = titles.filter((title) => functionalGrep.test(title));
      if (matchedTitles.length === 0) {
        continue;
      }
      for (const title of matchedTitles) {
        if (mode === "mismatch" && (entry.id === "E-2-02" || entry.id === "E-2-03")) {
          continue;
        }
        specs.push({
          title,
          file,
          tests: [{ results: [fakeResult("passed")] }],
        });
      }
    }
  }
}

if (project === "support") {
  for (const supportFile of [...supportFiles].sort()) {
    const source = fs.readFileSync(path.join(root, "apps", "web", "e2e", supportFile), "utf8");
    for (const match of source.matchAll(/\btest\("([^"]+)"/g)) {
      if (!supportGrep.test(match[1])) {
        continue;
      }
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

fake_pnpm="$tmp_dir/fake-pnpm.sh"
cat >"$fake_pnpm" <<EOF
#!/usr/bin/env bash
set -euo pipefail

if [[ "\${1:-}" == "--dir" ]]; then
  shift 2
fi
if [[ "\${1:-}" == "exec" ]]; then
  shift
fi
if [[ "\${1:-}" == "playwright" && "\${2:-}" == "test" ]]; then
  shift 2
fi

exec "$fake_playwright" "\$@"
EOF
chmod +x "$fake_pnpm"

success_invocations="$tmp_dir/batch-success-invocations.log"
success_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="batch-success" \
  BROWSER_E2E_FUNCTIONAL_SHARDS=2 \
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
const offsets = functional
  .map((line) => line.match(/worker_offset=([^ ]*)/)?.[1] ?? "")
  .sort();
if (offsets.join(",") !== "0,1") {
  throw new Error(`expected functional shard worker offsets 0,1, got ${offsets.join(",")}`);
}
for (const line of functional) {
  if (!line.includes("worker_count=2 ")) {
    throw new Error(`expected functional shard worker_count=2, got ${line}`);
  }
}
if (!functional.some((line) => line.includes("phase4.workbook.generic.spec.ts"))) {
  throw new Error("expected a functional shard containing phase4.workbook.generic.spec.ts");
}
if (!functional.some((line) => line.includes("phase4.mentions.resolve.spec.ts"))) {
  throw new Error("expected a functional shard containing phase4.mentions.resolve.spec.ts");
}
if (functional.some((line) =>
  line.includes("phase4.autoresolve.spec.ts") &&
  line.includes("phase4.mentions.lifecycle.spec.ts") &&
  line.includes("phase4.mentions.resolve.spec.ts") &&
  line.includes("phase4.merge.spec.ts") &&
  line.includes("phase4.workbook.assessments.spec.ts") &&
  line.includes("phase4.workbook.generic.spec.ts")
)) {
  throw new Error("phase4 entries were not split across duration-balanced shards");
}
NODE
success_root="$tmp_dir/results/batch-success/adhoc"
phase1_summary="$success_root/browser-e2e-functional-phase1-authoritative/phase-summary.json"
phase2_summary="$success_root/browser-e2e-functional-phase2-authoritative/phase-summary.json"
phase4_summary="$success_root/browser-e2e-functional-phase4-authoritative/phase-summary.json"
assert_equals "$(json_field "$phase1_summary" "status")" "pass" "phase1 batch success status"
assert_equals "$(json_field "$phase1_summary" "accounting_mode")" "actual" "phase1 batch accounting"
assert_equals "$(json_field "$phase2_summary" "status")" "pass" "phase2 batch success status"
assert_equals "$(json_field "$phase2_summary" "accounting_mode")" "actual" "phase2 batch accounting"
assert_equals "$(json_field "$phase4_summary" "accounting_mode")" "actual" "phase4 batch accounting"
support_phase2_summary="$success_root/browser-e2e-support-phase2-supplemental/phase-summary.json"
support_phase3_summary="$success_root/browser-e2e-support-phase3-supplemental/phase-summary.json"
assert_equals "$(json_field "$support_phase2_summary" "status")" "pass" "phase2 support batch success status"
assert_equals "$(json_field "$support_phase2_summary" "counts.support")" "3" "phase2 support batch support count"
assert_equals "$(json_field "$support_phase3_summary" "status")" "pass" "phase3 support batch success status"
assert_equals "$(json_field "$support_phase3_summary" "counts.support")" "3" "phase3 support batch support count"
phase1_timing="$success_root/browser-e2e-functional-phase1-authoritative/playwright-timing.json"
phase4_timing="$success_root/browser-e2e-functional-phase4-authoritative/playwright-timing.json"
assert_equals "$(json_field "$phase1_timing" "source")" "playwright_result_timestamps" "phase1 timing source"
assert_equals "$(json_field "$phase1_timing" "phase")" "phase1" "phase1 timing phase"
assert_equals "$(json_field "$phase4_timing" "files.0.file")" "apps/web/e2e/phase4.autoresolve.spec.ts" "phase4 timing first file"
assert_equals "$(json_field "$phase4_timing" "files.1.file")" "apps/web/e2e/phase4.mentions.lifecycle.spec.ts" "phase4 timing second file"
assert_equals "$(json_field "$phase4_timing" "files.2.file")" "apps/web/e2e/phase4.mentions.resolve.spec.ts" "phase4 timing third file"
assert_equals "$(json_field "$phase4_timing" "files.3.file")" "apps/web/e2e/phase4.merge.spec.ts" "phase4 timing fourth file"
assert_equals "$(json_field "$phase4_timing" "files.4.file")" "apps/web/e2e/phase4.workbook.assessments.spec.ts" "phase4 timing fifth file"
assert_equals "$(json_field "$phase4_timing" "files.5.file")" "apps/web/e2e/phase4.workbook.generic.spec.ts" "phase4 timing sixth file"
assert_equals "$(json_field "$phase4_timing" "entries.0.id")" "E-4-01" "phase4 timing first entry"
assert_equals "$(json_field "$phase4_timing" "entries.5.id")" "E-4-06" "phase4 timing sixth entry"
NODE_BIN="${NODE:-node}" CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" CARTULARY_TEST_RUN_ID="batch-success" \
  "$ROOT_DIR/scripts/lib/test-output.sh" target-summary adhoc pass >/dev/null
success_target_summary="$success_root/target-summary.json"
assert_equals "$(json_field "$success_target_summary" "kind")" "leaf" "batch target summary kind"
assert_equals "$(json_field "$success_target_summary" "totals.accounting_modes.actual")" "4" "batch target actual phase count"
assert_equals "$(json_field "$success_target_summary" "totals.accounting_modes.derived")" "2" "batch target derived phase count"

single_shard_invocations="$tmp_dir/batch-single-shard-invocations.log"
single_shard_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="batch-single-shard" \
  NODE_BIN="${NODE:-node}" \
  FAKE_PLAYWRIGHT_INVOCATIONS="$single_shard_invocations" \
    "$HELPER" functional-shard browser-functional-shard-01 0 2 -- "$fake_playwright"
)"
assert_empty "$single_shard_output" "playwright webserver single shard success"
"${NODE:-node}" - "$single_shard_invocations" <<'NODE'
const fs = require("node:fs");
const lines = fs.readFileSync(process.argv[2], "utf8").trim().split(/\n/u).filter(Boolean);
const functional = lines.filter((line) => line.startsWith("project=functional "));
if (functional.length !== 1) {
  throw new Error(`expected one functional shard invocation, got ${functional.length}`);
}
if (!functional[0].includes("worker_count=2 ") || !functional[0].includes("worker_offset=0 ")) {
  throw new Error(`single shard worker routing was not preserved: ${functional[0]}`);
}
if (!functional[0].includes("selected_ids=E-")) {
  throw new Error(`single shard must pass selected manifest IDs: ${functional[0]}`);
}
if (lines.some((line) => line.startsWith("project=support "))) {
  throw new Error("functional-shard mode must not run support project");
}
NODE
single_shard_root="$tmp_dir/results/batch-single-shard/adhoc"
single_shard_phase1="$single_shard_root/browser-e2e-functional-phase1-authoritative-browser-functional-shard-01/phase-summary.json"
assert_equals "$(json_field "$single_shard_phase1" "status")" "pass" "single shard phase1 status"
assert_equals "$(json_field "$single_shard_phase1" "accounting_mode")" "actual" "single shard phase1 accounting"

phase_filter_invocations="$tmp_dir/batch-phase-filter-invocations.log"
phase_filter_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
  CARTULARY_PHASE_SLICE_PHASE=phase4 \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="batch-phase-filter" \
  BROWSER_E2E_FUNCTIONAL_SHARDS=2 \
  NODE_BIN="${NODE:-node}" \
  FAKE_PLAYWRIGHT_INVOCATIONS="$phase_filter_invocations" \
    "$HELPER" webserver-backed -- "$fake_playwright"
)"
assert_empty "$phase_filter_output" "playwright webserver batch phase-filter success"
assert_contains "$(cat "$phase_filter_invocations")" "phase4.workbook.generic.spec.ts" "phase-filter functional shard includes phase4 file"
assert_not_contains "$(cat "$phase_filter_invocations")" "phase2.spec.ts" "phase-filter functional shard excludes phase2 file"
phase_filter_root="$tmp_dir/results/batch-phase-filter/adhoc"
phase_filter_phase4_summary="$phase_filter_root/browser-e2e-functional-phase4-authoritative/phase-summary.json"
assert_equals "$(json_field "$phase_filter_phase4_summary" "status")" "pass" "phase-filter phase4 functional status"
if [[ -e "$phase_filter_root/browser-e2e-functional-phase2-authoritative/phase-summary.json" ]]; then
  fail "phase-filtered browser batch must not emit phase2 functional summary"
fi
if [[ -e "$phase_filter_root/browser-e2e-support-phase2-supplemental/phase-summary.json" ]]; then
  fail "phase-filtered browser batch must not emit phase2 support summary"
fi

set +e
support_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="batch-support-failure" \
  BROWSER_E2E_FUNCTIONAL_SHARDS=2 \
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
assert_contains "$support_failure_output" "failure: browser-e2e-support phase3 supplemental" "support failure label"
assert_contains "$support_failure_output" "coverage=support" "support failure coverage"
assert_not_contains "$support_failure_output" "coverage=unmapped" "support failure unmapped coverage"
support_failure_phase1="$tmp_dir/results/batch-support-failure/adhoc/browser-e2e-functional-phase1-authoritative/phase-summary.json"
assert_equals "$(json_field "$support_failure_phase1" "status")" "pass" "support failure leaves phase1 passing"

set +e
mismatch_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="batch-mismatch" \
  BROWSER_E2E_FUNCTIONAL_SHARDS=2 \
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
batch_group_selections="$("${NODE:-node}" "$ROOT_DIR/scripts/lib/browser-batch-manifest.mjs" group-selections "$batch_manifest")"
assert_contains \
  "$batch_group_selections" \
  $'webserver-backed\tfunctional-support\tbrowser-e2e-webserver-backed\tduration_balanced_specs\tauthoritative\tbrowser_functional' \
  "browser batch group selection metadata"
assert_contains \
  "$batch_group_selections" \
  $'support\tsupport\tbrowser-e2e-support\tsupport\tsupplemental\tbrowser_support' \
  "browser support group selection metadata"
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
assert_contains "$(cat "$batch_runner")" "--defer-summary" "batch runner deferred summary option"
assert_contains "$(cat "$batch_runner")" "reset-web-e2e-stack.sh" "batch runner reset boundary"

summary_ownership_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="summary-ownership" \
  CARTULARY_TEST_TARGET="browser-e2e-webserver-backed" \
  BROWSER_E2E_FUNCTIONAL_SHARDS=2 \
  NODE_BIN="${NODE:-node}" \
  PNPM="$fake_pnpm" \
    "$batch_runner" webserver-backed --defer-summary 2>/dev/null
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="summary-ownership" \
  CARTULARY_TEST_TARGET="browser-e2e-webserver-backed" \
  CARTULARY_TIMING_BUCKET="teardown" \
  CARTULARY_TIMING_LABEL="browser-e2e stop owned processes" \
  CARTULARY_TIMING_START_TIME="2026-04-24T00:00:10.000Z" \
  CARTULARY_TIMING_END_TIME="2026-04-24T00:00:11.000Z" \
  CARTULARY_TIMING_DURATION_MS="1000" \
    "$ROOT_DIR/scripts/lib/test-output.sh" timing-span
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="summary-ownership" \
  NODE_BIN="${NODE:-node}" \
    "$ROOT_DIR/scripts/lib/test-output.sh" target-summary browser-e2e-webserver-backed pass
)"
summary_pass_count="$(printf '%s\n' "$summary_ownership_output" | grep -c '^\[RESULT\] target=browser-e2e-webserver-backed status=pass')"
assert_equals "$summary_pass_count" "1" "webserver-backed authoritative summary line count"
assert_contains "$summary_ownership_output" "slowest=teardown:1000" "webserver-backed authoritative summary includes teardown"

browser_aggregate_results="$tmp_dir/results/browser-aggregate"
for target in browser-e2e-stateful browser-e2e-measurement browser-e2e-visual; do
  mkdir -p "$browser_aggregate_results/$target"
  cat >"$browser_aggregate_results/$target/target-summary.json" <<JSON
{
  "target": "${target}",
  "status": "pass",
  "start_time": "2026-01-01T00:00:00Z",
  "end_time": "2026-01-01T00:00:01Z",
  "executed_duration_ms": 1,
  "logical_duration_ms": 1,
  "reused_duration_ms": 0,
  "derived_duration_ms": 0,
  "wall_duration_ms": 1,
  "critical_path_wall_duration_ms": 1,
  "teardown_duration_ms": 0,
  "counts": {
    "phases": 1,
    "tests": 1,
    "failed": 0,
    "authoritative": 1,
    "support": 0,
    "unmapped": 0,
    "non_test": 0,
    "authoritative_failed": 0,
    "support_failed": 0,
    "unmapped_failed": 0,
    "non_test_failed": 0,
    "packages": 1
  }
}
JSON
done
browser_aggregate_output="$(
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results" \
  CARTULARY_TEST_RUN_ID="browser-aggregate" \
    "$ROOT_DIR/scripts/lib/test-output.sh" target-summary browser-e2e pass --projection browser-e2e \
    2>&1
)"
assert_contains "$browser_aggregate_output" "[RESULT] target=browser-e2e status=pass" "browser aggregate child tests"
browser_aggregate_summary="$browser_aggregate_results/browser-e2e/target-summary.json"
assert_equals "$(json_field "$browser_aggregate_summary" "children.counts.tests")" "3" "browser aggregate JSON child tests"
assert_equals "$(json_field "$browser_aggregate_summary" "totals.counts.tests")" "3" "browser aggregate JSON total tests"
