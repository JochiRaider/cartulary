#!/usr/bin/env bash
# Single-quoted literals below intentionally assert Make/shell text without expansion.
# shellcheck disable=SC2016
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT_DIR}/scripts/run-make-sequence.sh"
task_surface_makefile="$ROOT_DIR/Makefile"
task_surface_generated_make_file="$ROOT_DIR/tools/task_surface.generated.mk"
cleanup_paths=()

# shellcheck source=scripts/lib/task-surface-check-common.sh
source "$ROOT_DIR/scripts/lib/task-surface-check-common.sh"

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "${path}"
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
' "${file}" "${path}"
}

assert_json_field_absent() {
  local file="$1"
  local path="$2"
  local label="$3"

  if "${NODE:-node}" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
process.exit(value === undefined ? 0 : 1);
' "${file}" "${path}"; then
    return 0
  fi
  fail "${label}: expected JSON field [${path}] to be absent"
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "${actual}" != "${expected}" ]]; then
    fail "${label}: expected [${expected}], got [${actual}]"
  fi
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "${haystack}" != *"${needle}"* ]]; then
    fail "${label}: expected output to contain [${needle}]"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "${haystack}" == *"${needle}"* ]]; then
    fail "${label}: expected output not to contain [${needle}]"
  fi
}

assert_file_absent() {
  local path="$1"
  local label="$2"

  if [[ -e "${path}" ]]; then
    fail "${label}: expected ${path} to be absent"
  fi
}

assert_count() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  assert_equals "${actual}" "${expected}" "${label}"
}

assert_output_occurrences() {
  local haystack="$1"
  local needle="$2"
  local expected="$3"
  local label="$4"
  local remaining="${haystack}"
  local actual=0

  while [[ "${remaining}" == *"${needle}"* ]]; do
    remaining="${remaining#*"${needle}"}"
    actual=$((actual + 1))
  done

  assert_equals "${actual}" "${expected}" "${label}"
}

line_count() {
  local pattern="$1"

  grep -Ec "${pattern}" "${ROOT_DIR}/Makefile"
}

write_fake_make() {
  local dir="$1"

  cat >"${dir}/fake-make" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

echo "$*" >>"${FAKE_MAKE_LOG}"

target="${@: -1}"
if [[ -n "${FAKE_MAKE_ENV_LOG:-}" ]]; then
  printf 'target=%s test_target=%s\n' "$target" "${CARTULARY_TEST_TARGET:-}" >>"${FAKE_MAKE_ENV_LOG}"
fi
case "${target}" in
  fail-step)
    exit 7
    ;;
esac

if [[ -n "${CARTULARY_TEST_RESULTS_DIR:-}" && -n "${CARTULARY_TEST_RUN_ID:-}" ]]; then
  mkdir -p "${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}"
  cat >"${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}/target-summary.json" <<JSON
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
    "tests": 0,
    "failed": 0,
    "authoritative": 0,
    "support": 0,
    "unmapped": 0,
    "non_test": 0,
    "authoritative_failed": 0,
    "support_failed": 0,
    "unmapped_failed": 0,
    "non_test_failed": 0,
    "packages": 0
  }
}
JSON
fi
EOF
  chmod +x "${dir}/fake-make"
}

manifest_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-manifest.XXXXXX")"
cleanup_paths+=("${manifest_dir}")
sequence_manifest="${manifest_dir}/task_surface_manifest.json"
"${NODE_BIN:-node}" - "${ROOT_DIR}/tools/task_surface_manifest.json" "${sequence_manifest}" <<'EOF'
const fs = require("node:fs");
const [source, destination] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(source, "utf8"));
for (const name of ["alpha", "beta", "missing-target", "fail-step", "smoke", "aggregate-missing", "fail-smoke", "dry-run"]) {
  if (!manifest.targets.some((target) => target.name === name)) {
    manifest.targets.push({ name, classification: "helper_only", included_in: ["helper_only"] });
  }
}
manifest.make_recipes ??= {};
for (const name of ["alpha", "beta", "missing-target", "fail-step", "smoke", "aggregate-missing", "fail-smoke", "dry-run"]) {
  manifest.make_recipes[name] ??= { type: "alias", prerequisites: [] };
}
manifest.sequences.smoke = {
  summary_groups: [
    { name: "alpha-group", summary_targets: ["alpha"] },
    { name: "beta-group", summary_targets: ["beta"] },
  ],
  steps: [
    { type: "step", target: "alpha", produces_summary_targets: ["alpha"] },
    { type: "parallel", target: "beta", jobs: 3, produces_summary_targets: ["beta"] },
  ],
};
manifest.sequences["aggregate-missing"] = {
  summary_groups: [],
  steps: [{ type: "step", target: "alpha", produces_summary_targets: ["alpha", "missing-target"] }],
};
manifest.sequences["fail-smoke"] = {
  summary_groups: [],
  steps: [
    { type: "step", target: "alpha", produces_summary_targets: ["alpha"] },
    { type: "step", target: "fail-step" },
    { type: "step", target: "beta", produces_summary_targets: ["beta"] },
  ],
};
manifest.sequences["dry-run"] = {
  summary_groups: [],
  steps: [{ type: "step", target: "alpha", produces_summary_targets: ["alpha"] }],
};
fs.writeFileSync(destination, `${JSON.stringify(manifest, null, 2)}\n`);
EOF

success_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-success.XXXXXX")"
cleanup_paths+=("${success_dir}")
write_fake_make "${success_dir}"
success_results="${success_dir}/results"
success_output="$(
  VERBOSE="" \
  CI_VERBOSE="" \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_OUTPUT_MODE="" \
  MAKE="${success_dir}/fake-make" \
  FAKE_MAKE_LOG="${success_dir}/make.log" \
  FAKE_MAKE_ENV_LOG="${success_dir}/make-env.log" \
  CARTULARY_TEST_RESULTS_DIR="${success_results}" \
  CARTULARY_TEST_RUN_ID="success" \
  TASK_SURFACE_MANIFEST="${sequence_manifest}" \
    "${SCRIPT}" --sequence smoke \
    2>&1
)"
assert_not_contains "${success_output}" "[RUN]" "success run start output"
assert_not_contains "${success_output}" "[STEP]" "success step output"
assert_contains "${success_output}" "[RESULT] target=smoke status=pass" "success run summary output"
assert_contains "${success_output}" "[ARTIFACTS] target=smoke" "success artifact output"
success_summary="${success_results}/success/run-summary.json"
assert_equals "$(json_field "${success_summary}" "status")" "pass" "success status"
assert_equals "$(json_field "${success_summary}" "work_units.completed")" "2" "success completed work units"
assert_equals "$(json_field "${success_summary}" "work_units.total")" "2" "success total work units"
assert_equals "$(json_field "${success_summary}" "summary_targets.expected.0")" "alpha" "success summary target 0"
assert_equals "$(json_field "${success_summary}" "summary_targets.expected.1")" "beta" "success summary target 1"
assert_equals "$(json_field "${success_summary}" "summary_groups.0.name")" "alpha-group" "success group 0"
assert_equals "$(json_field "${success_summary}" "summary_groups.0.summary_targets.0")" "alpha" "success group target 0"
assert_equals "$(json_field "${success_summary}" "summary_groups.0.wall_duration_ms")" "1000" "success group wall duration"
assert_equals "$(json_field "${success_summary}" "summary_groups.0.critical_path_wall_duration_ms")" "1000" "success group critical path duration"
assert_equals "$(json_field "${success_summary}" "summary_groups.0.teardown_duration_ms")" "0" "success group teardown duration"
assert_json_field_absent "${success_summary}" "duration_ms" "success legacy run duration"
assert_json_field_absent "${success_summary}" "summary_groups.0.duration_ms" "success legacy group duration"
assert_contains "$(cat "${success_dir}/make.log")" "--output-sync=target -j3 beta" "parallel make invocation"
assert_contains "$(cat "${success_dir}/make-env.log")" "target=alpha test_target=" "sequence serial target env not forwarded"
assert_contains "$(cat "${success_dir}/make-env.log")" "target=beta test_target=" "sequence parallel target env not forwarded"

aggregate_missing_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-aggregate-missing.XXXXXX")"
cleanup_paths+=("${aggregate_missing_dir}")
write_fake_make "${aggregate_missing_dir}"
aggregate_missing_results="${aggregate_missing_dir}/results"
set +e
aggregate_missing_output="$(
  VERBOSE="" \
  CI_VERBOSE="" \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_OUTPUT_MODE="" \
  MAKE="${aggregate_missing_dir}/fake-make" \
  FAKE_MAKE_LOG="${aggregate_missing_dir}/make.log" \
  FAKE_MAKE_ENV_LOG="${aggregate_missing_dir}/make-env.log" \
  CARTULARY_TEST_RESULTS_DIR="${aggregate_missing_results}" \
  CARTULARY_TEST_RUN_ID="aggregate-missing" \
  TASK_SURFACE_MANIFEST="${sequence_manifest}" \
    "${SCRIPT}" --sequence aggregate-missing \
    2>&1
)"
aggregate_missing_status=$?
set -e
assert_equals "${aggregate_missing_status}" "1" "aggregate missing target exit status"
assert_not_contains "${aggregate_missing_output}" "[RUN]" "aggregate missing run start output"
assert_not_contains "${aggregate_missing_output}" "[STEP]" "aggregate missing step output"
assert_contains "${aggregate_missing_output}" "[FAIL] target=aggregate-missing" "aggregate missing target run summary output"
assert_output_occurrences "${aggregate_missing_output}" "[FAIL] target=aggregate-missing" "1" "aggregate missing single failure block"
aggregate_missing_summary="${aggregate_missing_results}/aggregate-missing/run-summary.json"
assert_equals "$(json_field "${aggregate_missing_summary}" "status")" "fail" "aggregate missing target status"
assert_equals "$(json_field "${aggregate_missing_summary}" "summary_targets.missing.0")" "missing-target" "aggregate missing target list"

failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-failure.XXXXXX")"
cleanup_paths+=("${failure_dir}")
write_fake_make "${failure_dir}"
failure_results="${failure_dir}/results"
set +e
failure_output="$(
  VERBOSE="" \
  CI_VERBOSE="" \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_OUTPUT_MODE="" \
  MAKE="${failure_dir}/fake-make" \
  FAKE_MAKE_LOG="${failure_dir}/make.log" \
  FAKE_MAKE_ENV_LOG="${failure_dir}/make-env.log" \
  CARTULARY_TEST_RESULTS_DIR="${failure_results}" \
  CARTULARY_TEST_RUN_ID="failure" \
  TASK_SURFACE_MANIFEST="${sequence_manifest}" \
    "${SCRIPT}" --sequence fail-smoke \
    2>&1
)"
failure_status=$?
set -e
assert_equals "${failure_status}" "7" "failure child exit status"
assert_not_contains "${failure_output}" "[RUN]" "failure run start output"
assert_not_contains "${failure_output}" "[STEP]" "failure step output"
assert_contains "${failure_output}" "[FAIL] target=fail-smoke" "failure run summary output"
assert_output_occurrences "${failure_output}" "[FAIL] target=fail-smoke" "1" "sequence single failure block"
failure_summary="${failure_results}/failure/run-summary.json"
assert_equals "$(json_field "${failure_summary}" "status")" "fail" "failure status"
assert_equals "$(json_field "${failure_summary}" "work_units.completed")" "1" "failure completed work units"
assert_equals "$(json_field "${failure_summary}" "work_units.total")" "3" "failure total work units"
assert_equals "$(json_field "${failure_summary}" "work_units.aborted_after")" "fail-step" "failure aborted_after"
assert_equals "$(json_field "${failure_summary}" "counts.non_test_failed")" "1" "failure non-test count"
assert_equals "$(json_field "${failure_summary}" "failure_class")" "helper" "failure class"
assert_equals "$(json_field "${failure_summary}" "failure_classes.helper")" "1" "failure helper count"
assert_contains "$(cat "${failure_dir}/make-env.log")" "target=fail-step test_target=" "sequence failing helper target env not forwarded"

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-dry-run.XXXXXX")"
cleanup_paths+=("${dry_run_dir}")
write_fake_make "${dry_run_dir}"
dry_run_output="$(
  VERBOSE="" \
  CI_VERBOSE="" \
  CARTULARY_OUTPUT_MODE="" \
  MAKEFLAGS="n" \
  MAKE="${dry_run_dir}/fake-make" \
  FAKE_MAKE_LOG="${dry_run_dir}/make.log" \
  FAKE_MAKE_ENV_LOG="${dry_run_dir}/make-env.log" \
  CARTULARY_TEST_RESULTS_DIR="${dry_run_dir}/results" \
  CARTULARY_TEST_RUN_ID="dry-run" \
  TASK_SURFACE_MANIFEST="${sequence_manifest}" \
    "${SCRIPT}" --sequence dry-run \
    2>&1
)"
assert_not_contains "${dry_run_output}" "[RUN]" "script dry-run run start output"
assert_not_contains "${dry_run_output}" "[STEP]" "script dry-run step output"
assert_file_absent "${dry_run_dir}/results/dry-run/run-summary.json" "script dry-run summary"
assert_contains "$(cat "${dry_run_dir}/make.log")" "--no-print-directory alpha" "script dry-run child make"
assert_contains "$(cat "${dry_run_dir}/make-env.log")" "target=alpha test_target=" "script dry-run target env not forwarded"

invalid_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-invalid.XXXXXX")"
cleanup_paths+=("${invalid_dir}")
write_fake_make "${invalid_dir}"
set +e
invalid_output="$(
  MAKE="${invalid_dir}/fake-make" \
  FAKE_MAKE_LOG="${invalid_dir}/make.log" \
    "${SCRIPT}" --summary-targets alpha \
    2>&1
)"
invalid_status=$?
set -e
assert_equals "${invalid_status}" "2" "invalid usage status"
assert_contains "${invalid_output}" "usage: run-make-sequence.sh --sequence <name>" "invalid usage output"
assert_file_absent "${invalid_dir}/make.log" "invalid usage child make log"

makefile_content="$(cat "${ROOT_DIR}/Makefile")"
service_schedule_target_content="$(cat "${ROOT_DIR}/scripts/cartulary-runner.mjs")"
generated_make="$(cat "${ROOT_DIR}/tools/task_surface.generated.mk")"
generated_phony_line="$(printf '%s\n' "${generated_make}" | sed -n 's/^\\.PHONY: //p')"
manifest_content="$(cat "${ROOT_DIR}/tools/task_surface_manifest.json")"
assert_count "$(line_count '^RUN_MAKE_SEQUENCE_SCRIPT :=')" "1" "run sequence helper declaration"
assert_count "$(line_count '^RUN_HARNESS_SMOKE_SCRIPT :=')" "1" "harness smoke helper declaration"
assert_count "$(line_count '^RUN_SERVICE_BACKED_SCHEDULE_SCRIPT :=')" "1" "service-backed scheduler helper declaration"
assert_count "$(line_count '^RUN_CHECK_SCHEDULE_SCRIPT :=')" "1" "check scheduler helper declaration"
assert_contains "${makefile_content}" "include tools/task_surface.generated.mk" "Makefile includes generated task surface"
assert_contains "${generated_make}" 'RUN_MAKE_NODE_TOOL = env NODE_BIN="$(NODE_BIN)" $(2) ./scripts/run-make-node-tool.sh $(1)' "generated Make node-tool macro"
assert_not_contains "${generated_make}" 'BASELINE_FILE="$(BASELINE_FILE)" CARTULARY_TEST_RESULTS_DIR="$(CARTULARY_TEST_RESULTS_DIR)" CARTULARY_TEST_RUN_ID="$(CARTULARY_TEST_RUN_ID)" DETAIL="$(DETAIL)"' "generated Make old global node-tool env block"
assert_not_contains "${generated_make}" "TASK_SURFACE_HARNESS_TIER_" "generated Make harness tier variables"
assert_not_contains "${generated_phony_line}" "harness-smoke-toolchain-pins" "generated Make harness leaf targets"
assert_not_contains "${generated_phony_line}" "run-harness-smoke-fast-all" "generated Make fast harness aggregate leaf"
assert_contains "${generated_make}" "summary-target --target check-go-test-duration-baseline-drift --child-target go-test-duration-baseline-drift --status pass" "generated summary target runner invocation"
assert_contains "${generated_make}" 'MAKE_BIN="$(MAKE)"' "generated summary target make bin"
assert_contains "${generated_make}" 'RUN_PHASE_SCRIPT="$(RUN_PHASE_SCRIPT)"' "generated summary target run phase"
assert_not_contains "${generated_make}" '$(MAKE) --no-print-directory go-test-duration-baseline-drift' "generated summary target old child make"
assert_not_contains "${generated_make}" '$(TARGET_SUMMARY) check-go-test-duration-baseline-drift pass' "generated summary target old summary call"
assert_count "$(line_count '^RUN_MAKE_SEQUENCE_SCRIPT :=')" "1" "run sequence helper declaration"
assert_count "$(line_count '^RUN_HARNESS_SMOKE_SCRIPT :=')" "1" "harness smoke helper declaration"
assert_count "$(line_count '^RUN_SERVICE_BACKED_SCHEDULE_SCRIPT :=')" "1" "service-backed scheduler helper declaration"
assert_count "$(line_count '^RUN_CHECK_SCHEDULE_SCRIPT :=')" "1" "check scheduler helper declaration"

test_block="$(extract_target_definition test)"
test_fast_block="$(extract_target_definition test-fast)"
ci_block="$(extract_target_definition ci)"
release_check_block="$(extract_target_definition release-check)"
check_block="$(extract_target_definition check)"
run_harness_smoke_fast_block="$(extract_target_definition run-harness-smoke-fast)"
run_harness_smoke_extended_block="$(extract_target_definition run-harness-smoke-extended)"
run_harness_smoke_full_block="$(extract_target_definition run-harness-smoke-full)"
check_harness_smoke_block="$(extract_target_definition check-harness-smoke)"
test_service_backed_block="$(extract_target_definition test-service-backed)"
test_fast_service_backed_block="$(extract_target_definition test-fast-service-backed)"
check_service_backed_block="$(extract_target_definition check-service-backed)"
task_guide_block="$(extract_target_definition task-guide)"
target_plan_block="$(extract_target_definition target-plan)"
fixture_report_block="$(extract_target_definition fixture-report)"
go_test_duration_baseline_drift_block="$(extract_target_definition go-test-duration-baseline-drift)"
browser_e2e_duration_baselines_block="$(extract_target_definition browser-e2e-duration-baselines)"
assert_contains "${test_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT)' "make test helper invocation"
assert_contains "${test_block}" "--sequence test" "make test manifest sequence"
assert_contains "${test_fast_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT)' "make test-fast helper invocation"
assert_contains "${test_fast_block}" "--sequence test-fast" "make test-fast manifest sequence"
assert_contains "${ci_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT)' "make ci helper invocation"
assert_contains "${ci_block}" "--sequence ci" "make ci manifest sequence"
assert_contains "${ci_block}" "ci: export CI := 1" "make ci exports CI"
assert_contains "${release_check_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT)' "make release-check helper invocation"
assert_contains "${release_check_block}" "--sequence release-check" "make release-check manifest sequence"
assert_not_contains "${test_block}" "--summary-profile test" "make test no inline summary profile"
assert_not_contains "${test_block}" "--parallel-step test-local:3 --step test-service-backed" "make test no inline sequence"
assert_not_contains "${test_block}" "--step browser-e2e" "make test no final serial browser step"
assert_not_contains "${test_block}" "--step test-isolated" "make test old split browser sequence"
assert_not_contains "${test_block}" "completed=" "make test inline completed counter"
assert_not_contains "${test_block}" "total=" "make test inline total counter"
assert_contains "${check_block}" '$(RUN_CHECK_SCHEDULE_SCRIPT)' "make check scheduler invocation"
assert_not_contains "${check_block}" "--summary-profile check" "make check no copied summary profile"
assert_contains "${check_block}" '$(TASK_SURFACE_CHECK_SCHEDULER_OVERRIDE_ENV)' "make check scheduler override env forwarding"
assert_not_contains "${check_block}" '--resource-limit host_cpu=' "make check no default cpu CLI resource override"
assert_not_contains "${check_block}" '--resource-limit host_io=' "make check no default io CLI resource override"
assert_not_contains "${check_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT)' "make check no longer uses serial sequence helper"
assert_not_contains "${check_block}" "--step browser-e2e" "make check no final serial browser step"
assert_not_contains "${check_block}" "--step check-isolated" "make check old split browser sequence"
assert_not_contains "${check_block}" "completed=" "make check inline completed counter"
assert_not_contains "${check_block}" "total=" "make check inline total counter"
assert_contains "${run_harness_smoke_fast_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier fast --jobs "$(HARNESS_SMOKE_JOBS)"' "run-harness-smoke-fast manifest runner"
assert_contains "${run_harness_smoke_extended_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier extended --jobs "$(HARNESS_SMOKE_JOBS)"' "run-harness-smoke-extended manifest runner"
assert_contains "${run_harness_smoke_full_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier full --jobs "$(HARNESS_SMOKE_JOBS)"' "run-harness-smoke-full manifest runner"
assert_contains "${check_harness_smoke_block}" "run-harness-smoke-fast" "check-harness-smoke fast tier invocation"
assert_contains "${check_harness_smoke_block}" "--projection check-harness-smoke" "check-harness-smoke summary projection"
assert_contains "${task_guide_block}" '$(call RUN_MAKE_NODE_TOOL,task-guide,ROLE="$(ROLE)" PHASE="$(PHASE)" JSON="$(JSON)" CARTULARY_TEST_RESULTS_DIR="$(CARTULARY_TEST_RESULTS_DIR)")' "task-guide node-tool scoped env"
assert_not_contains "${task_guide_block}" 'BASELINE_FILE="$(BASELINE_FILE)"' "task-guide omits unrelated baseline env"
assert_contains "${target_plan_block}" '$(call RUN_MAKE_NODE_TOOL,target-plan,)' "target-plan node-tool empty env"
assert_not_contains "${target_plan_block}" 'PHASE="$(PHASE)"' "target-plan omits phase env"
assert_contains "${fixture_report_block}" '$(call RUN_MAKE_NODE_TOOL,fixture-report,FIXTURE_THRESHOLD_MS="$(FIXTURE_THRESHOLD_MS)" FIXTURE_TOP="$(FIXTURE_TOP)" RUN_ID="$(RUN_ID)" TARGET="$(TARGET)" JSON="$(JSON)" RESULTS_DIR="$(RESULTS_DIR)" CARTULARY_TEST_RESULTS_DIR="$(CARTULARY_TEST_RESULTS_DIR)")' "fixture-report node-tool scoped env"
assert_not_contains "${fixture_report_block}" 'ROLE="$(ROLE)"' "fixture-report omits unrelated role env"
assert_contains "${go_test_duration_baseline_drift_block}" '$(call RUN_MAKE_NODE_TOOL,go-test-duration-baseline-drift,GO_TEST_DURATION_BASELINE="$(GO_TEST_DURATION_BASELINE)" RESULTS_DIR="$(RESULTS_DIR)" CARTULARY_TEST_RESULTS_DIR="$(CARTULARY_TEST_RESULTS_DIR)" CARTULARY_TEST_RUN_ID="$(CARTULARY_TEST_RUN_ID)")' "go duration drift node-tool current-run env"
assert_contains "${browser_e2e_duration_baselines_block}" '$(call RUN_MAKE_NODE_TOOL,browser-e2e-duration-baselines,BROWSER_E2E_DURATION_BASELINE="$(BROWSER_E2E_DURATION_BASELINE)" RESULTS_DIR="$(RESULTS_DIR)")' "browser duration refresh node-tool explicit baseline env"
assert_not_contains "${manifest_content}" "\"summary_profiles\"" "manifest copied summary profiles"
assert_not_contains "${manifest_content}" "\"summary_projection\"" "manifest copied summary projections"
env NODE_BIN="${NODE_BIN:-node}" "${NODE_BIN:-node}" --input-type=module - "${ROOT_DIR}/tools/task_surface_manifest.json" "${ROOT_DIR}/tools/service_backed_schedule_manifest.json" "${ROOT_DIR}/tools/browser_e2e_batch_manifest.json" <<'EOF'
import fs from "node:fs";
import { loadSummaryTopologyContext, resolveSummaryGroups } from "./scripts/lib/summary-topology.mjs";
import { loadBrowserBatchStages } from "./scripts/lib/browser-batch-manifest.mjs";

const [manifestFile, serviceManifest, browserManifest] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const context = loadSummaryTopologyContext({
  taskSurfaceManifest: manifest,
  serviceBackedScheduleManifestPath: serviceManifest,
  browserBatchManifestPath: browserManifest,
});
const sequence = manifest.sequences.test;
const testTargets = sequence.steps.flatMap((step) => step.produces_summary_targets ?? []);
for (const target of ["backend-unit", "frontend-typecheck", "frontend-unit", "test-service-backed"]) {
  if (!testTargets.includes(target)) {
    throw new Error(`test sequence must produce ${target}`);
  }
}
const groups = resolveSummaryGroups(context, sequence.summary_groups);
const local = groups.find((group) => group.name === "local");
const expectedLocal = ["backend-unit", "frontend-typecheck", "frontend-unit"];
if (JSON.stringify(local?.summaryTargets) !== JSON.stringify(expectedLocal)) {
  throw new Error("test local summary group must include local leaf targets");
}
const browser = groups.find((group) => group.name === "browser");
const browserStages = loadBrowserBatchStages(browserManifest);
const webserverBackedStage = browserStages.get("webserver-backed");
const isolatedStage = browserStages.get("isolated");
if (!webserverBackedStage || !isolatedStage) {
  throw new Error("browser batch manifest must declare webserver-backed and isolated stages");
}
const expectedBrowser = [
  webserverBackedStage.target,
  ...(isolatedStage.summaryChildren.length > 0 ? isolatedStage.summaryChildren : [isolatedStage.target]),
];
if (JSON.stringify(browser?.summaryTargets) !== JSON.stringify(expectedBrowser)) {
  throw new Error("test browser summary group must derive browser leaves from schedules");
}
const testFastTargets = manifest.sequences["test-fast"].steps.flatMap((step) => step.produces_summary_targets ?? []);
for (const target of ["backend-unit", "frontend-typecheck", "frontend-unit", "test-fast-service-backed"]) {
  if (!testFastTargets.includes(target)) {
    throw new Error(`test-fast sequence must produce ${target}`);
  }
}
const ciTargets = manifest.sequences.ci.steps.flatMap((step) => step.produces_summary_targets ?? []);
for (const target of ["check", "run-harness-smoke-extended"]) {
  if (!ciTargets.includes(target)) {
    throw new Error(`ci sequence must produce ${target}`);
  }
}
const releaseTargets = manifest.sequences["release-check"].steps.flatMap((step) => step.produces_summary_targets ?? []);
for (const target of ["check", "run-harness-smoke-extended"]) {
  if (!releaseTargets.includes(target)) {
    throw new Error(`release-check sequence must produce ${target}`);
  }
}
EOF
assert_contains "${manifest_content}" "\"harness_checks\"" "manifest logical harness checks"
assert_contains "${manifest_content}" "harness-smoke-run-make-sequence-fast" "harness smoke fast make sequence check"
assert_contains "${manifest_content}" "harness-smoke-check-scheduler-smoke" "harness smoke fast check scheduler smoke"
assert_contains "${manifest_content}" "harness-smoke-service-backed-scheduler-smoke" "harness smoke fast service-backed scheduler smoke"
assert_not_contains "${manifest_content}" "harness-smoke-run-go-target-fast" "retired fast go target harness smoke"
assert_contains "${test_service_backed_block}" 'service-backed-target --target test-service-backed --phase-label "test service-backed" --service-wrapper test-services' "test service-backed scheduler invocation"
assert_contains "${test_fast_service_backed_block}" 'service-backed-target --target test-fast-service-backed --phase-label "test-fast service-backed" --service-wrapper test-services' "test-fast service-backed scheduler invocation"
assert_contains "${check_service_backed_block}" 'service-backed-target --target check-service-backed --phase-label "check service-backed" --service-wrapper test-services' "check service-backed scheduler invocation"
assert_contains "${service_schedule_target_content}" '"--children"' "service-backed runner summary children"
assert_not_contains "${test_service_backed_block}" "--jobs" "test service-backed fixed scheduler jobs"
assert_not_contains "${test_fast_service_backed_block}" "--jobs" "test-fast service-backed fixed scheduler jobs"
assert_not_contains "${check_service_backed_block}" "--jobs" "check service-backed fixed scheduler jobs"
assert_not_contains "${makefile_content}" "RUN_SUMMARY =" "unused run summary helper variable"
assert_not_contains "${makefile_content}" "RUN_SUMMARY_CMD =" "unused run summary command variable"
assert_not_contains "${makefile_content}" "bash -lc './scripts/test-check-toolchain-pins.sh &&" "old serialized harness smoke chain"

for target in test-fast test ci release-check run-harness-smoke-fast run-harness-smoke-extended run-harness-smoke-full; do
  make_dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-make-n-${target}.XXXXXX")"
  cleanup_paths+=("${make_dry_run_dir}")
  make_dry_run_output="$(
    CARTULARY_TEST_RESULTS_DIR="${make_dry_run_dir}/results" \
    CARTULARY_TEST_RUN_ID="make-n-${target}" \
      make -n --no-print-directory "${target}" \
      2>&1
  )"
  if [[ "${target}" == run-harness-smoke-* ]]; then
    assert_contains "${make_dry_run_output}" "scripts/run-harness-smoke.mjs --tier ${target#run-harness-smoke-}" "make -n ${target} helper command"
  else
    assert_contains "${make_dry_run_output}" "scripts/run-make-sequence.sh --sequence ${target}" "make -n ${target} helper command"
  fi
  assert_file_absent "${make_dry_run_dir}/results/make-n-${target}/run-summary.json" "make -n ${target} summary"
done

check_dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-make-n-check.XXXXXX")"
cleanup_paths+=("${check_dry_run_dir}")
check_dry_run_output="$(
  CARTULARY_TEST_RESULTS_DIR="${check_dry_run_dir}/results" \
  CARTULARY_TEST_RUN_ID="make-n-check" \
    make -n --no-print-directory check \
    2>&1
)"
assert_contains "${check_dry_run_output}" "scripts/run-check-schedule.mjs --target check" "make -n check scheduler command"
assert_not_contains "${check_dry_run_output}" "--step browser-e2e" "make -n check no final browser step"
assert_file_absent "${check_dry_run_dir}/results/make-n-check/run-summary.json" "make -n check summary"
