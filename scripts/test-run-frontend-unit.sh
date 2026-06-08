#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
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

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

artifact_abs_path() {
  local value="$1"
  if [[ "$value" = /* ]]; then
    printf '%s\n' "$value"
    return
  fi
  printf '%s\n' "$ROOT_DIR/$value"
}

assert_artifact_present() {
  local value="$1"
  local label="$2"
  local file
  file="$(artifact_abs_path "$value")"
  if [[ ! -f "$file" ]]; then
    fail "$label: missing artifact [$value]"
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
const phaseRegistry = JSON.parse(fs.readFileSync(path.join(root, "tools", "phase_registry.json"), "utf8"));
const phaseFiles = (phaseRegistry.phases ?? [])
  .filter((entry) => entry.status === "active")
  .sort((left, right) => left.order - right.order || left.phase.localeCompare(right.phase))
  .map((entry) => entry.manifest_path);
const manifestSections = ["unit", "integration", "e2e", "visual"];
const authoritative = [];
const manifestSupport = [];

for (const manifestPath of phaseFiles) {
  const manifest = JSON.parse(fs.readFileSync(path.join(root, manifestPath), "utf8"));
  for (const section of manifestSections) {
    for (const row of manifest[section] ?? []) {
      if (row.runner !== "vitest" || row.execution_dependency !== "frontend_unit") {
        continue;
      }
      if (row.coverage === "authoritative") {
        authoritative.push(row);
      } else {
        manifestSupport.push(row);
      }
    }
  }
}

const statusFor = (index, fallback = "passed") => {
  if ((mode === "authoritative-failure" || mode === "stack-failure") && index === 0) {
    return "failed";
  }
  return fallback;
};

const assertion = (title, status) => ({
  ancestorTitles: ["frontend-unit smoke"],
  fullName: `frontend-unit smoke ${title}`,
  status,
  title,
  failureMessages:
    status !== "failed"
      ? []
      : mode === "stack-failure"
        ? [
            `Error: STACK_TRACE_ERROR
    at task (file:///tmp/vitest/chunk-artifact.js:1784:27)
    at ${path.join(root, "apps/web/src/WorkbookShell.phase3.grid.test.tsx")}:56:3`,
          ]
        : ["frontend unit smoke failure"],
  meta: {},
  tags: [],
});

const byFile = new Map();
for (const [index, row] of authoritative.entries()) {
  const absolute = path.join(root, row.file);
  const entries = byFile.get(absolute) ?? [];
  const titles = Array.isArray(row.titles) ? row.titles : [row.title];
  for (const title of titles.filter(Boolean)) {
    entries.push(assertion(title, statusFor(index)));
  }
  byFile.set(absolute, entries);
}

for (const [index, row] of manifestSupport.entries()) {
  const absolute = path.join(root, row.file);
  const entries = byFile.get(absolute) ?? [];
  const titles = Array.isArray(row.titles) ? row.titles : [row.title];
  for (const title of titles.filter(Boolean)) {
    entries.push(
      assertion(
        title,
        mode === "manifest-support-failure" && index === 0
          ? "failed"
          : "passed",
      ),
    );
  }
  byFile.set(absolute, entries);
}

const frontendPhaseEntries = [];
const frontendRegistry = JSON.parse(fs.readFileSync(path.join(root, "tools", "frontend_phase_registry.json"), "utf8"));
function rowIsInActiveTargetScope(phase, row) {
  if (phase.status === "active") {
    return true;
  }
  if (phase.status !== "planned") {
    return false;
  }
  return row.claim_status === "implemented" || row.claim_status === "stale";
}
for (const entry of (frontendRegistry.phases ?? [])) {
  const frontendPhase = JSON.parse(
    fs.readFileSync(path.join(root, entry.manifest_path), "utf8"),
  );
  for (const row of frontendPhase.rows ?? []) {
    if (!rowIsInActiveTargetScope(entry, row)) {
      continue;
    }
    if (
      row.claim_status !== "implemented" ||
      !(row.targets ?? []).some((target) => target.target_name === "frontend-unit") ||
      (row.scenario_titles ?? []).length === 0
    ) {
      continue;
    }
  for (const [index, title] of (row.scenario_titles ?? []).entries()) {
    if (
      mode === "frontend-row-missing-implemented" &&
      row.id === "FE-I-P1-01" &&
      index === 0
    ) {
      continue;
    }
    frontendPhaseEntries.push(
      assertion(
        title,
        mode === "frontend-row-failure" && row.id === "FE-I-P1-01" && index === 0
          ? "failed"
          : "passed",
      ),
    );
  }
  }
}
byFile.set(path.join(root, "apps/web/src/App.phase1.test.tsx"), [
  ...(byFile.get(path.join(root, "apps/web/src/App.phase1.test.tsx")) ?? []),
  ...frontendPhaseEntries,
]);

byFile.set(path.join(root, "apps/web/src/App.phase1.support.test.tsx"), [
  assertion("Phase 1 support smoke keeps ordinary shell helpers stable", "passed"),
]);
byFile.set(path.join(root, "apps/web/src/Unmapped.frontend-unit.test.tsx"), [
  assertion(
    "classified frontend residual smoke",
    mode === "residual-failure" ? "failed" : "passed",
  ),
]);
if (mode === "unknown-failure") {
  byFile.set(path.join(root, "apps/web/src/Unknown.frontend-unit.test.tsx"), [
    assertion("unknown frontend residual smoke", "failed"),
  ]);
}

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
  echo "frontend unit smoke stdout"
  echo "frontend unit smoke stderr" >&2
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
frontend_counts="$("${NODE:-node}" - "$ROOT_DIR" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");

const [root] = process.argv.slice(2);
const registry = JSON.parse(fs.readFileSync(path.join(root, "tools", "phase_registry.json"), "utf8"));
const sections = ["unit", "integration", "e2e", "visual"];
let authoritative = 0;
let support = 0;
let frontendAuthoritative = 0;
let frontendSupport = 0;
const phases = new Set();
for (const entry of (registry.phases ?? [])
  .filter((phase) => phase.status === "active")
  .sort((left, right) => left.order - right.order || left.phase.localeCompare(right.phase))) {
  const manifest = JSON.parse(fs.readFileSync(path.join(root, entry.manifest_path), "utf8"));
  for (const section of sections) {
    for (const row of manifest[section] ?? []) {
      if (
        row.runner === "vitest" &&
        row.execution_dependency === "frontend_unit"
      ) {
        const count = Array.isArray(row.titles) ? row.titles.length : 1;
        if (row.coverage === "authoritative") {
          authoritative += count;
        } else {
          support += count;
        }
        phases.add(entry.phase);
      }
    }
  }
}
const frontendRegistry = JSON.parse(fs.readFileSync(path.join(root, "tools", "frontend_phase_registry.json"), "utf8"));
function rowIsInActiveTargetScope(phase, row) {
  if (phase.status === "active") {
    return true;
  }
  if (phase.status !== "planned") {
    return false;
  }
  return row.claim_status === "implemented" || row.claim_status === "stale";
}
for (const entry of (frontendRegistry.phases ?? [])) {
  const frontendPhase = JSON.parse(
    fs.readFileSync(path.join(root, entry.manifest_path), "utf8"),
  );
  for (const row of frontendPhase.rows ?? []) {
  if (!rowIsInActiveTargetScope(entry, row)) {
    continue;
  }
  if (
    row.claim_status !== "implemented" ||
    !(row.targets ?? []).some((target) => target.target_name === "frontend-unit") ||
    (row.scenario_titles ?? []).length === 0
  ) {
    continue;
  }
  const count = (row.scenario_titles ?? []).length;
  if (row.evidence_class === "product_conformance") {
    frontendAuthoritative += count;
  } else {
    frontendSupport += count;
  }
  }
}
process.stdout.write(`${authoritative},${support},${phases.size},${frontendAuthoritative},${frontendSupport}`);
EOF
)"
IFS=',' read -r base_authoritative base_support phase_count frontend_authoritative frontend_support <<<"$frontend_counts"
expected_authoritative="$(( base_authoritative + frontend_authoritative ))"
expected_support="$(( base_support + 1 + frontend_support ))"
expected_derived="$(( phase_count + 1 ))"
expected_total="$(( expected_authoritative + expected_support + 1 ))"
assert_equals "$(json_field "$success_summary" "own.counts.tests")" "$expected_total" "success total tests"
assert_equals "$(json_field "$success_summary" "own.counts.authoritative")" "$expected_authoritative" "success authoritative count"
assert_equals "$(json_field "$success_summary" "own.counts.support")" "$expected_support" "success support count"
assert_equals "$(json_field "$success_summary" "own.counts.unowned_regression")" "1" "success unowned regression count"
assert_equals "$(json_field "$success_summary" "own.counts.unmapped")" "0" "success unmapped count"
assert_equals "$(json_field "$success_summary" "own.accounting_modes.actual")" "1" "success raw actual phase"
assert_equals "$(json_field "$success_summary" "own.accounting_modes.derived")" "$expected_derived" "success derived slices"
"${NODE:-node}" - "$success_summary" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [summaryPath] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryPath, "utf8"));
const accounting = summary.extensions?.["cartulary.frontend_row_accounting"];
if (!accounting) {
  throw new Error("frontend-unit target summary must include frontend row accounting");
}
const artifactRel = summary.artifacts?.frontend_row_accounting;
if (!artifactRel) {
  throw new Error("frontend-unit target summary must reference frontend row accounting artifact");
}
const artifact = JSON.parse(
  fs.readFileSync(path.resolve(process.cwd(), artifactRel), "utf8"),
);
if (artifact.schema_id !== "cartulary.frontend_row_accounting.v3") {
  throw new Error(`frontend row accounting artifact has wrong schema: ${artifact.schema_id}`);
}
if (accounting.accounting_scope?.mode !== "active_target") {
  throw new Error(`frontend-unit broad target must use active target accounting scope: ${JSON.stringify(accounting.accounting_scope)}`);
}
if (JSON.stringify(artifact) !== JSON.stringify(accounting)) {
  throw new Error("frontend row accounting artifact must match compatibility extension");
}
const byID = new Map((accounting.rows ?? []).map((row) => [row.row_id, row]));
const plannedNonAccountable = (accounting.rows ?? []).filter((row) =>
  row.phase_status === "planned" &&
  ["blocked", "not_implemented", "retired"].includes(row.claim_status)
);
if (plannedNonAccountable.length > 0) {
  throw new Error(`active target accounting must exclude planned blocked/not-implemented rows: ${JSON.stringify(plannedNonAccountable)}`);
}
const feu5 = byID.get("FE-U-P5-01");
if (!feu5 || feu5.closure_status !== "closed") {
  throw new Error(`implemented FE-U-P5-01 must be closed in success accounting: ${JSON.stringify(feu5)}`);
}
const fei = byID.get("FE-I-P1-01");
if (!fei || fei.closure_status !== "closed") {
  throw new Error(`FE-I-P1-01 must be closed in success accounting: ${JSON.stringify(fei)}`);
}
if (fei.scenarios.filter((scenario) => scenario.status === "passed").length !== 11) {
  throw new Error("FE-I-P1-01 must retain per-scenario passed accounting");
}
const fei3 = byID.get("FE-I-P3-01");
if (!fei3 || fei3.closure_status !== "closed") {
  throw new Error(`FE-I-P3-01 must be closed in success accounting: ${JSON.stringify(fei3)}`);
}
if (fei3.scenarios.filter((scenario) => scenario.status === "passed").length !== 1) {
  throw new Error("FE-I-P3-01 must retain per-scenario passed accounting");
}
const toolSummary = JSON.parse(
  fs.readFileSync(path.join(path.dirname(summaryPath), "tool-run-summary.json"), "utf8"),
);
if (!toolSummary.extensions?.["cartulary.frontend_row_accounting"]) {
  throw new Error("frontend-unit tool summary must include frontend row accounting");
}
if (!toolSummary.summary_artifacts?.some((entry) =>
  entry.role === "frontend_row_accounting" && entry.path === artifactRel
)) {
  throw new Error("frontend-unit tool summary must reference frontend row accounting artifact");
}
EOF

success_target_dir="${success_summary%/target-summary.json}"
selected_scope_root="$tmp_dir/results-selected/selected/frontend-unit"
mkdir -p "$selected_scope_root"
cp -R "$success_target_dir/." "$selected_scope_root/"
CARTULARY_OUTPUT_MODE=quiet \
CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results-selected" \
CARTULARY_TEST_RUN_ID="selected" \
NODE_BIN="$runtime_dir/bin/node" \
  "$ROOT_DIR/scripts/lib/test-output.sh" target-summary frontend-unit pass \
  --frontend-row-accounting-scope selected_rows \
  --frontend-row-accounting-phase-namespace frontend \
  --frontend-row-accounting-phase FE-P3 \
  --frontend-row-accounting-row-ids FE-I-P1-01 >/dev/null
selected_scope_summary="$selected_scope_root/target-summary.json"
"${NODE:-node}" - "$selected_scope_summary" <<'EOF'
const fs = require("node:fs");
const [summaryPath] = process.argv.slice(2);
const accounting = JSON.parse(fs.readFileSync(summaryPath, "utf8"))
  .extensions?.["cartulary.frontend_row_accounting"];
if (accounting.accounting_scope?.mode !== "selected_rows") {
  throw new Error(`selected scope was not retained: ${JSON.stringify(accounting.accounting_scope)}`);
}
if (accounting.accounting_scope.phase !== "FE-P3") {
  throw new Error("selected scope must retain frontend phase");
}
const rowIDs = (accounting.rows ?? []).map((row) => row.row_id);
if (rowIDs.join(",") !== "FE-I-P1-01") {
  throw new Error(`selected scope must emit only FE-I-P1-01, got ${rowIDs.join(",")}`);
}
EOF

disabled_scope_root="$tmp_dir/results-disabled/disabled/frontend-unit"
mkdir -p "$disabled_scope_root"
cp -R "$success_target_dir/." "$disabled_scope_root/"
CARTULARY_OUTPUT_MODE=quiet \
CARTULARY_TEST_RESULTS_DIR="$tmp_dir/results-disabled" \
CARTULARY_TEST_RUN_ID="disabled" \
NODE_BIN="$runtime_dir/bin/node" \
  "$ROOT_DIR/scripts/lib/test-output.sh" target-summary frontend-unit pass \
  --frontend-row-accounting-scope disabled \
  --frontend-row-accounting-phase-namespace base \
  --frontend-row-accounting-phase phase9 >/dev/null
disabled_scope_summary="$disabled_scope_root/target-summary.json"
"${NODE:-node}" - "$disabled_scope_summary" <<'EOF'
const fs = require("node:fs");
const [summaryPath] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryPath, "utf8"));
const accounting = summary.extensions?.["cartulary.frontend_row_accounting"];
if (summary.status !== "pass") {
  throw new Error(`disabled scope target summary must pass, got ${summary.status}`);
}
if (accounting.accounting_scope?.mode !== "disabled") {
  throw new Error(`disabled scope was not retained: ${JSON.stringify(accounting.accounting_scope)}`);
}
if ((accounting.rows ?? []).length !== 0) {
  throw new Error(`disabled scope must emit zero rows, got ${accounting.rows.length}`);
}
EOF

residual_summary="$(run_case residual residual-failure fail)"
assert_equals "$(json_field "$residual_summary" "own.counts.failed")" "1" "residual failure count"
assert_equals "$(json_field "$residual_summary" "own.counts.unowned_regression_failed")" "1" "residual unowned regression failure count"
assert_equals "$(json_field "$residual_summary" "own.counts.unmapped_failed")" "0" "residual unmapped failure count"
assert_equals "$(json_field "$residual_summary" "own.counts.authoritative_failed")" "0" "residual authoritative failure count"
"${NODE:-node}" - "$residual_summary" <<'EOF'
const fs = require("node:fs");
const [summaryPath] = process.argv.slice(2);
const accounting = JSON.parse(fs.readFileSync(summaryPath, "utf8"))
  .extensions?.["cartulary.frontend_row_accounting"];
const fei = new Map((accounting?.rows ?? []).map((row) => [row.row_id, row])).get("FE-I-P1-01");
if (!fei || fei.closure_status !== "blocked_by_target") {
  throw new Error(`unrelated target failure must block, not fail, FE-I-P1-01: ${JSON.stringify(fei)}`);
}
EOF

unknown_summary="$(run_case unknown unknown-failure fail)"
assert_equals "$(json_field "$unknown_summary" "own.counts.failed")" "1" "unknown residual failure count"
assert_equals "$(json_field "$unknown_summary" "own.counts.unmapped_failed")" "1" "unknown residual unmapped failure count"
assert_equals "$(json_field "$unknown_summary" "own.counts.unowned_regression_failed")" "0" "unknown residual unowned regression failure count"

authoritative_summary="$(run_case authoritative authoritative-failure fail)"
assert_equals "$(json_field "$authoritative_summary" "own.counts.failed")" "1" "authoritative failure count"
assert_equals "$(json_field "$authoritative_summary" "own.counts.authoritative_failed")" "1" "authoritative authoritative failure count"
assert_equals "$(json_field "$authoritative_summary" "own.counts.unmapped_failed")" "0" "authoritative unmapped failure count"
manifest_support_summary="$(run_case manifest-support manifest-support-failure fail)"
assert_equals "$(json_field "$manifest_support_summary" "own.counts.failed")" "1" "manifest support failure count"
assert_equals "$(json_field "$manifest_support_summary" "own.counts.support_failed")" "1" "manifest support support failure count"
assert_equals "$(json_field "$manifest_support_summary" "own.counts.authoritative_failed")" "0" "manifest support authoritative failure count"
assert_equals "$(json_field "$manifest_support_summary" "own.counts.unmapped_failed")" "0" "manifest support unmapped failure count"
frontend_row_summary="$(run_case frontend-row frontend-row-failure fail)"
assert_equals "$(json_field "$frontend_row_summary" "own.counts.failed")" "1" "frontend row failure count"
assert_equals "$(json_field "$frontend_row_summary" "own.counts.authoritative_failed")" "1" "frontend row authoritative failure count"
"${NODE:-node}" - "$frontend_row_summary" <<'EOF'
const fs = require("node:fs");
const [summaryPath] = process.argv.slice(2);
const accounting = JSON.parse(fs.readFileSync(summaryPath, "utf8"))
  .extensions?.["cartulary.frontend_row_accounting"];
const fei = new Map((accounting?.rows ?? []).map((row) => [row.row_id, row])).get("FE-I-P1-01");
if (!fei || fei.closure_status !== "failed") {
  throw new Error(`mapped FE-I-P1-01 assertion failure must mark the row failed: ${JSON.stringify(fei)}`);
}
if (!fei.scenarios.some((scenario) => scenario.status === "failed")) {
  throw new Error("mapped FE-I-P1-01 failure must retain failed scenario accounting");
}
EOF
frontend_row_missing_summary="$(run_case frontend-row-missing frontend-row-missing-implemented fail)"
assert_equals "$(json_field "$frontend_row_missing_summary" "status")" "fail" "implemented frontend row missing status"
assert_equals "$(json_field "$frontend_row_missing_summary" "failure_class")" "harness" "implemented frontend row missing failure class"
assert_equals "$(json_field "$frontend_row_missing_summary" "failure_reason")" "frontend_row_accounting" "implemented frontend row missing failure reason"
"${NODE:-node}" - "$frontend_row_missing_summary" <<'EOF'
const fs = require("node:fs");
const [summaryPath] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryPath, "utf8"));
const accounting = summary.extensions?.["cartulary.frontend_row_accounting"];
const fei = new Map((accounting?.rows ?? []).map((row) => [row.row_id, row])).get("FE-I-P1-01");
if (!fei || fei.closure_status !== "missing") {
  throw new Error(`implemented FE-I-P1-01 missing scenario must fail target accounting: ${JSON.stringify(fei)}`);
}
if (!summary.failures?.some((failure) => failure.source === "frontend-row-accounting")) {
  throw new Error("target summary must include a frontend-row-accounting failure");
}
EOF
"${NODE:-node}" - "$success_target_dir" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [targetDir] = process.argv.slice(2);
const summaries = fs.readdirSync(targetDir, { withFileTypes: true })
  .filter((entry) => entry.isDirectory())
  .map((entry) => path.join(targetDir, entry.name, "phase-summary.json"))
  .filter((summaryPath) => fs.existsSync(summaryPath))
  .map((summaryPath) => JSON.parse(fs.readFileSync(summaryPath, "utf8")));
const authoritativePhase9 = summaries.find((summary) =>
  summary.label === "frontend-unit phase9 authoritative"
);
if (!authoritativePhase9) {
  throw new Error("frontend-unit smoke must emit a phase9 authoritative summary");
}
const authoritativeIDs = new Set(
  (authoritativePhase9.inventory ?? [])
    .filter((entry) => entry.coverage === "authoritative")
    .map((entry) => entry.id),
);
if (!authoritativeIDs.has("I-9-GRID-01")) {
  throw new Error("phase9 authoritative Vitest summary must include I-9-GRID-01");
}
const residual = summaries.find((summary) => summary.label === "frontend-unit residual");
if (!residual) {
  throw new Error("frontend-unit smoke must emit a residual summary");
}
const residualIDs = new Set((residual.inventory ?? []).map((entry) => entry.id));
if (residualIDs.has("I-9-GRID-01")) {
  throw new Error("frontend-unit residual summary must exclude I-9-GRID-01");
}
EOF
authoritative_phase_summary="${authoritative_summary%/target-summary.json}/frontend-unit-vitest/phase-summary.json"
runner_json="$(json_field "$authoritative_phase_summary" "artifacts.runner_json")"
stdout_log="$(json_field "$authoritative_phase_summary" "artifacts.stdout_log")"
stderr_log="$(json_field "$authoritative_phase_summary" "artifacts.stderr_log")"
assert_contains "$runner_json" "/raw/frontend-unit/runner.json" "authoritative failure runner artifact path"
assert_contains "$stdout_log" "/raw/frontend-unit/stdout.log" "authoritative failure stdout artifact path"
assert_contains "$stderr_log" "/raw/frontend-unit/stderr.log" "authoritative failure stderr artifact path"
assert_artifact_present "$runner_json" "authoritative failure runner artifact"
assert_artifact_present "$stdout_log" "authoritative failure stdout artifact"
assert_artifact_present "$stderr_log" "authoritative failure stderr artifact"

stack_summary="$(run_case stack stack-failure fail)"
stack_target_dir="${stack_summary%/target-summary.json}"
"${NODE:-node}" - "$stack_target_dir" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [targetDir] = process.argv.slice(2);
const summaries = fs.readdirSync(targetDir, { withFileTypes: true })
  .filter((entry) => entry.isDirectory())
  .map((entry) => path.join(targetDir, entry.name, "phase-summary.json"))
  .filter((summaryPath) => fs.existsSync(summaryPath))
  .map((summaryPath) => JSON.parse(fs.readFileSync(summaryPath, "utf8")));
const dossier = summaries
  .flatMap((summary) => summary.dossiers ?? [])
  .find((entry) => String(entry.message ?? "").includes("STACK_TRACE_ERROR"));
if (!dossier) {
  throw new Error("stack failure summary must include fallback STACK_TRACE_ERROR diagnostic");
}
if (!String(dossier.message).includes("first_app_frame=apps/web/src/WorkbookShell.phase3.grid.test.tsx:56")) {
  throw new Error(`stack failure summary did not include first app frame: ${dossier.message}`);
}
if (!(dossier.diagnostic_tags ?? []).includes("vitest_stack_trace_error")) {
  throw new Error("stack failure summary must include vitest_stack_trace_error diagnostic tag");
}
EOF
