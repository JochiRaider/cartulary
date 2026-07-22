#!/usr/bin/env bash
# Single-quoted literals below intentionally assert Make/shell text without expansion.
# shellcheck disable=SC2016
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
SCRIPT="${ROOT_DIR}/tools/harness/execution/run-make-sequence.sh"
task_surface_makefile="$ROOT_DIR/Makefile"
task_surface_generated_make_file="$ROOT_DIR/tools/task_surface.generated.mk"
cleanup_paths=()

unset VERBOSE CI_VERBOSE CARTULARY_OUTPUT_MODE CARTULARY_SUPPRESS_CHILD_SUCCESS

# shellcheck source=tools/harness/test-support/task-surface-check-common.sh
source "$ROOT_DIR/tools/harness/test-support/task-surface-check-common.sh"
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "$ROOT_DIR/tools/harness/test-support/harness-scratch.sh"

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

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "${actual}" != "${expected}" ]]; then
    fail "${label}: expected [${expected}], got [${actual}]"
  fi
}

assert_file_absent() {
  local path="$1"
  local label="$2"

  if [[ -e "${path}" ]]; then
    fail "${label}: expected ${path} to be absent"
  fi
}

assert_file_present() {
  local path="$1"
  local label="$2"

  if [[ ! -f "${path}" ]]; then
    fail "${label}: expected ${path} to exist"
  fi
}

json_field() {
  local file="$1"
  local path="$2"

  "${NODE_BIN:-node}" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
if (value === undefined || value === null) {
  process.exit(1);
}
process.stdout.write(String(value));
' "${file}" "${path}"
}

assert_output_budget() {
  local manifest="$1"
  local target="$2"
  local stdout_file="$3"
  local stderr_file="$4"
  local label="$5"

  "${NODE_BIN:-node}" - "${manifest}" "${target}" "${stdout_file}" "${stderr_file}" "${label}" <<'EOF'
const fs = require("node:fs");
const [manifestPath, targetName, stdoutFile, stderrFile, label] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
const target = manifest.targets.find((entry) => entry.name === targetName);
if (!target?.output_policy?.success_budget) {
  throw new Error(`${label}: missing success budget for ${targetName}`);
}
const budget = target.output_policy.success_budget;
const readText = (file) => fs.existsSync(file) ? fs.readFileSync(file, "utf8") : "";
const lineCount = (text) => {
  if (text.length === 0) return 0;
  const trimmed = text.endsWith("\n") ? text.slice(0, -1) : text;
  return trimmed.length === 0 ? 0 : trimmed.split(/\r?\n/).length;
};
const checks = [
  ["stdout_lines", lineCount(readText(stdoutFile))],
  ["stdout_bytes", Buffer.byteLength(readText(stdoutFile))],
  ["stderr_lines", lineCount(readText(stderrFile))],
  ["stderr_bytes", Buffer.byteLength(readText(stderrFile))],
];
for (const [key, actual] of checks) {
  const limit = budget[key];
  if (Number.isInteger(limit) && actual > limit) {
    throw new Error(`${label}: ${key} ${actual} exceeds budget ${limit}`);
  }
}
EOF
}

assert_single_machine_json() {
  local stdout_file="$1"
  local stderr_file="$2"
  local expected_target="$3"
  local label="$4"
  shift 4

  "${NODE_BIN:-node}" - "${stdout_file}" "${stderr_file}" "${expected_target}" "${label}" "$@" <<'EOF'
const fs = require("node:fs");
const [stdoutFile, stderrFile, expectedTarget, label, ...expectedRoles] = process.argv.slice(2);
const stdout = fs.readFileSync(stdoutFile, "utf8");
const stderr = fs.readFileSync(stderrFile, "utf8");
if (stderr !== "") {
  throw new Error(`${label}: expected empty stderr in machine mode, got ${JSON.stringify(stderr)}`);
}
if (stdout.includes("[RESULT]") || stdout.includes("[ARTIFACTS]")) {
  throw new Error(`${label}: machine mode must not include human summary lines`);
}
const lines = stdout.split(/\r?\n/).filter((line) => line.trim() !== "");
if (lines.length !== 1) {
  throw new Error(`${label}: expected exactly one JSON line, got ${lines.length}`);
}
const summary = JSON.parse(lines[0]);
if (summary.schema_id !== "cartulary.tool_run_summary.v5") {
  throw new Error(`${label}: unexpected schema ${summary.schema_id}`);
}
if (summary.target !== expectedTarget) {
  throw new Error(`${label}: expected target ${expectedTarget}, got ${summary.target}`);
}
for (const field of ["started_at", "completed_at"]) {
  if (typeof summary[field] !== "string" || summary[field].trim() === "" || Number.isNaN(Date.parse(summary[field]))) {
    throw new Error(`${label}: ${field} must be a non-empty timestamp`);
  }
}
const roles = new Set((summary.summary_artifacts ?? []).map((artifact) => artifact.role));
for (const role of expectedRoles) {
  if (!roles.has(role)) {
    throw new Error(`${label}: missing summary artifact role ${role}`);
  }
}
EOF
}

write_fake_make() {
  local dir="$1"

  cat >"${dir}/fake-make" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

echo "$*" >>"${FAKE_MAKE_LOG}"
if [[ -n "${FAKE_MAKE_ENV_LOG:-}" ]]; then
  printf 'target=%s skip=%s satisfied=%s check_cpu=%s check_io=%s service_cpu=%s service_io=%s\n' \
    "${@: -1}" \
    "${CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES:-unset}" \
    "${CARTULARY_SEQUENCE_PREREQUISITES_SATISFIED:-unset}" \
    "${CHECK_HOST_CPU_JOBS:-unset}" \
    "${CHECK_HOST_IO_JOBS:-unset}" \
    "${CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT:-unset}" \
    "${CARTULARY_SERVICE_BACKED_GO_IO_LIMIT:-unset}" \
    >>"${FAKE_MAKE_ENV_LOG}"
fi

if [[ -n "${CARTULARY_TEST_RESULTS_DIR:-}" && -n "${CARTULARY_TEST_RUN_ID:-}" ]]; then
  write_summary() {
    local summary_target="$1"
    mkdir -p "${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${summary_target}"
    cat >"${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${summary_target}/target-summary.json" <<JSON
{
  "target": "${summary_target}",
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
    "steps": 1,
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
  }
  targets=()
  for arg in "$@"; do
    case "${arg}" in
      -*|*=*) ;;
      *) targets+=("${arg}") ;;
    esac
  done
  if [[ "${#targets[@]}" -eq 0 ]]; then
    targets=("${@: -1}")
  fi
  for target in "${targets[@]}"; do
    case "${target}" in
      test-local)
        write_summary backend-unit
        write_summary frontend-typecheck
        write_summary frontend-unit
        ;;
      test-fast-service-backed)
        write_summary build-operator
        write_summary build-server-harness
        write_summary backend-integration
        write_summary backend-integration-support
        write_summary backend-store
        write_summary backend-process
        write_summary test-fast-service-backed
        ;;
      check)
        write_summary check
        ;;
      build)
        write_summary build-web
        write_summary build-server
        write_summary build-migrate
        write_summary build-operator
        ;;
      run-harness-smoke-extended)
        write_summary run-harness-smoke-extended
        ;;
      seaweedfs-release-gate)
        write_summary seaweedfs-compatibility
        write_summary seaweedfs-migration-preservation
        write_summary seaweedfs-release-gate
        ;;
      release-browser-readiness)
        write_summary browser-e2e-support
        write_summary browser-e2e-visual
        write_summary browser-e2e-a11y
        write_summary release-browser-readiness
        ;;
      *)
        write_summary "${target}"
        ;;
    esac
  done
fi
EOF
  chmod +x "${dir}/fake-make"
}

manifest_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-manifest.XXXXXX")"
cleanup_paths+=("${manifest_dir}")
sequence_manifest="${manifest_dir}/task_surface_manifest.json"
"${NODE_BIN:-node}" - "${ROOT_DIR}/tools/task_surface_manifest.json" "${sequence_manifest}" <<'EOF'
const fs = require("node:fs");
const [source, destination] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(source, "utf8"));
for (const name of ["alpha", "beta", "gamma", "smoke", "generic-resource", "dry-run"]) {
  let target = manifest.targets.find((entry) => entry.name === name);
  if (!target) {
    target = { name, target_class: "internal_helper", default_inclusion_sets: [], lifecycle_state: "candidate_child" };
    manifest.targets.push(target);
  }
  if (["alpha", "smoke"].includes(name)) {
    target.output_policy = {
      output_class: name === "smoke" ? "aggregate_summary_with_artifacts" : "summary_with_artifacts",
      artifact_policy: name === "smoke" ? "run_and_target_summaries" : "tool_run_summary",
      success_budget: { stdout_lines: 2, stdout_bytes: 1500, stderr_lines: 0, stderr_bytes: 0 },
      failure_budget: { stderr_lines: 35, stderr_bytes: 6000, excerpt_lines: 25, excerpt_bytes: 4096 },
      raw_stream_policy: "never_default",
      summary_schema: "cartulary.tool_run_summary.v5",
    };
  }
  manifest.make_recipes[name] = { type: "aggregate", prerequisites: ["help"] };
}
manifest.sequences.smoke = {
  execution_mode: "dag",
  max_jobs: 3,
  resource_limits: { process: 3 },
  summary_groups: [
    { name: "alpha-group", summary_targets: ["alpha"] },
    { name: "beta-group", summary_targets: ["beta"] },
    { name: "gamma-group", summary_targets: ["gamma"] },
  ],
  steps: [
    { type: "step", target: "alpha", priority: 30, needs: [], resource_claims: { process: 1 }, produces_summary_targets: ["alpha"] },
    { type: "parallel", target: "beta", priority: 20, needs: ["alpha"], resource_claims: { process: 1 }, jobs: 3, produces_summary_targets: ["beta"] },
    { type: "step", target: "gamma", priority: 10, needs: ["beta"], resource_claims: { process: 1 }, skip_prerequisites: true, produces_summary_targets: ["gamma"] },
  ],
};
manifest.make_recipes.smoke = { type: "sequence", prerequisites: [], sequence: "smoke" };
manifest.sequences["generic-resource"] = {
  execution_mode: "dag",
  max_jobs: 3,
  resource_limits: { host_io: 1, process: 3 },
  summary_groups: [],
  steps: [
    { type: "step", target: "alpha", priority: 20, needs: [], resource_claims: { host_io: 1, process: 1 }, produces_summary_targets: ["alpha"] },
    { type: "step", target: "beta", priority: 10, needs: [], resource_claims: { host_io: 1, process: 1 }, produces_summary_targets: ["beta"] },
    { type: "step", target: "gamma", priority: 0, needs: ["alpha", "beta"], resource_claims: { process: 1 }, produces_summary_targets: ["gamma"] },
  ],
};
manifest.make_recipes["generic-resource"] = { type: "sequence", prerequisites: [], sequence: "generic-resource" };
manifest.sequences["dry-run"] = {
  execution_mode: "serial",
  max_jobs: 1,
  resource_limits: { process: 1 },
  summary_groups: [],
  steps: [{ type: "step", target: "alpha", priority: 10, needs: [], resource_claims: { process: 1 }, produces_summary_targets: ["alpha"] }],
};
manifest.make_recipes["dry-run"] = { type: "sequence", prerequisites: [], sequence: "dry-run" };
fs.writeFileSync(destination, `${JSON.stringify(manifest, null, 2)}\n`);
EOF

"${NODE_BIN:-node}" --input-type=module - <<'EOF'
import {
  estimateSequenceHostCPULimit,
  estimateSequenceHostIOLimit,
  estimateSequenceProcessLimit,
} from "./tools/harness/scheduler/scheduler-resource-policy.mjs";

if (
  estimateSequenceHostCPULimit(24) !== 20 ||
  estimateSequenceHostIOLimit(20, 24) !== 24 ||
  estimateSequenceProcessLimit(24) !== 8
) {
  throw new Error("24-way sequence capacity must resolve to CPU=20 IO=24 process=8");
}
EOF

success_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-success.XXXXXX")"
cleanup_paths+=("${success_dir}")
write_fake_make "${success_dir}"
success_output="$(
  VERBOSE="" \
  CI_VERBOSE="" \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_OUTPUT_MODE="" \
  MAKE="${success_dir}/fake-make" \
  FAKE_MAKE_LOG="${success_dir}/make.log" \
  FAKE_MAKE_ENV_LOG="${success_dir}/make-env.log" \
  CARTULARY_CHECK_SCHEDULER_SKIP_PREREQUISITES=1 \
  CARTULARY_SEQUENCE_PREREQUISITES_SATISFIED=1 \
  CARTULARY_TEST_RESULTS_DIR="${success_dir}/results" \
  CARTULARY_TEST_RUN_ID="success" \
  TASK_SURFACE_MANIFEST="${sequence_manifest}" \
    "${SCRIPT}" --sequence smoke \
    2>&1
)"
assert_not_contains "${success_output}" "[RUN]" "success run start output"
assert_not_contains "${success_output}" "[STEP]" "success step output"
assert_contains "${success_output}" "[RESULT] target=smoke status=pass" "success run summary output"
assert_contains "${success_output}" "[ARTIFACTS] target=smoke" "success artifact output"
assert_file_present "${success_dir}/results/success/smoke/target-summary.json" "success target summary"
assert_equals "$(json_field "${success_dir}/results/success/smoke/target-summary.json" "target")" "smoke" "success target summary identity"
assert_file_present "${success_dir}/results/success/smoke/scheduler-events.jsonl" "success scheduler events"
assert_file_present "${success_dir}/results/success/smoke/scheduler-summary.json" "success scheduler summary"
assert_file_present "${success_dir}/results/success/smoke/sequence-events.jsonl" "success sequence events"
"${NODE_BIN:-node}" - \
  "${success_dir}/results/success/smoke/scheduler-events.jsonl" \
  "${success_dir}/results/success/smoke/scheduler-summary.json" \
  "${success_dir}/results/success/smoke/sequence-events.jsonl" <<'EOF'
const fs = require("node:fs");
const [schedulerEventsFile, schedulerSummaryFile, sequenceEventsFile] = process.argv.slice(2);
const lines = (file) => fs.readFileSync(file, "utf8").trim().split(/\r?\n/).map(JSON.parse);
const schedulerEvents = lines(schedulerEventsFile);
const schedulerSummary = JSON.parse(fs.readFileSync(schedulerSummaryFile, "utf8"));
const sequenceEvents = lines(sequenceEventsFile);
if (schedulerSummary.schema_id !== "cartulary.sequence_scheduler_summary.v1" || schedulerSummary.scheduler_kind !== "sequence") {
  throw new Error("sequence must retain its typed shared-scheduler summary");
}
for (const [index, event] of schedulerEvents.entries()) {
  if (event.schema_id !== "cartulary.scheduler_event.v7" || event.scheduler_kind !== "sequence" || event.seq !== index + 1) {
    throw new Error("sequence scheduler events must be contiguous v7 evidence");
  }
}
const terminalSchedulerEvent = schedulerEvents.at(-1);
const edgeTokens = terminalSchedulerEvent.dependency_edges.map((edge) => `${edge.from}->${edge.to}`);
if (JSON.stringify(edgeTokens) !== JSON.stringify(["alpha->beta", "beta->gamma"])) {
  throw new Error(`unexpected retained dependency edges ${JSON.stringify(edgeTokens)}`);
}
if (!terminalSchedulerEvent.work_unit_states.every((state) => state.terminal_state === "passed")) {
  throw new Error("all successful sequence work units must be terminal in retained state");
}
for (const [index, event] of sequenceEvents.entries()) {
  if (event.schema_id !== "cartulary.harness_sequence_event.v1" || event.seq !== index + 1) {
    throw new Error("sequence lifecycle events must be contiguous typed evidence");
  }
  if (index > 0 && event.monotonic_ms < sequenceEvents[index - 1].monotonic_ms) {
    throw new Error("sequence lifecycle evidence must be monotonic");
  }
}
if (sequenceEvents[0].event !== "sequence_started" || sequenceEvents.at(-1).event !== "sequence_finished") {
  throw new Error("sequence lifecycle must have exact start and finish boundaries");
}
const terminalByTarget = new Map(sequenceEvents.filter((event) => event.event === "step_finished").map((event) => [event.target, event]));
for (const [target, index] of [["alpha", 1], ["beta", 2], ["gamma", 3]]) {
  const terminal = terminalByTarget.get(target);
  if (terminal?.step_index !== index || terminal.mode !== "dag" || terminal.status !== "pass") {
    throw new Error(`sequence lifecycle lost authored identity for ${target}`);
  }
}
EOF
assert_contains "$(cat "${success_dir}/make.log")" "--output-sync=target -j3 beta" "parallel make invocation"
assert_contains "$(cat "${success_dir}/make-env.log")" "target=alpha skip=unset satisfied=unset" "normal sequence step clears inherited prerequisite state"
assert_contains "$(cat "${success_dir}/make-env.log")" "target=beta skip=unset satisfied=unset" "parallel sequence step clears inherited prerequisite state"
assert_contains "$(cat "${success_dir}/make-env.log")" "target=gamma skip=1 satisfied=1" "explicit skip sequence step owns prerequisite state"

generic_resource_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-generic-resource.XXXXXX")"
cleanup_paths+=("${generic_resource_dir}")
write_fake_make "${generic_resource_dir}"
MAKE="${generic_resource_dir}/fake-make" \
FAKE_MAKE_LOG="${generic_resource_dir}/make.log" \
CARTULARY_TEST_RESULTS_DIR="${generic_resource_dir}/results" \
CARTULARY_TEST_RUN_ID="generic-resource" \
TASK_SURFACE_MANIFEST="${sequence_manifest}" \
  "${SCRIPT}" --sequence generic-resource >/dev/null
"${NODE_BIN:-node}" - "${generic_resource_dir}/results/generic-resource/generic-resource/scheduler-summary.json" "${generic_resource_dir}/results/generic-resource/generic-resource/scheduler-events.jsonl" <<'EOF'
const fs = require("node:fs");
const [summaryFile, eventsFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\r?\n/).map(JSON.parse);
if (summary.resource_limits.host_io !== 1 || summary.max_active_resource_claims.host_io !== 1) {
  throw new Error("generic logical resource capacity was not enforced");
}
if (!events.some((event) => event.event === "blocked" && event.blocked_resources.includes("host_io"))) {
  throw new Error("generic logical resource contention was not retained");
}
EOF

leaf_budget_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-leaf-budget.XXXXXX")"
cleanup_paths+=("${leaf_budget_dir}")
CARTULARY_TEST_TARGET=alpha \
CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
CARTULARY_TEST_RESULTS_DIR="${leaf_budget_dir}/results" \
CARTULARY_TEST_RUN_ID="leaf-budget" \
  "${ROOT_DIR}/tools/harness/execution/run-step.sh" "alpha smoke" -- true \
  >/dev/null 2>&1
CARTULARY_OUTPUT_MODE="" \
CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
CARTULARY_TEST_RESULTS_DIR="${leaf_budget_dir}/results" \
CARTULARY_TEST_RUN_ID="leaf-budget" \
TASK_SURFACE_MANIFEST="${sequence_manifest}" \
  "${ROOT_DIR}/tools/harness/output/test-output.sh" target-summary alpha pass \
  >"${leaf_budget_dir}/stdout.log" \
  2>"${leaf_budget_dir}/stderr.log"
assert_output_budget "${sequence_manifest}" alpha "${leaf_budget_dir}/stdout.log" "${leaf_budget_dir}/stderr.log" "leaf success budget"
assert_contains "$(cat "${leaf_budget_dir}/stdout.log")" "[RESULT] target=alpha status=pass" "leaf success budget result"

retained_biome_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-retained-biome.XXXXXX")"
cleanup_paths+=("${retained_biome_dir}")
cat >"${retained_biome_dir}/Makefile" <<EOF
lint-biome:
	@CARTULARY_SUPPRESS_CHILD_SUCCESS=1 CARTULARY_TEST_TARGET=lint-biome CARTULARY_TEST_RESULTS_DIR="${retained_biome_dir}/results" CARTULARY_TEST_RUN_ID=retained-biome "${ROOT_DIR}/tools/harness/execution/run-step.sh" "lint biome" -- bash -lc 'printf "%s\n" "apps/web/src/example.ts:12:8 lint/style/noNonNullAssertion" "  ! Forbidden non-null assertion."; exit 1'; status=\$\$?; if [ "\$\$status" -eq 0 ]; then CARTULARY_OUTPUT_MODE=quiet CARTULARY_TEST_RESULTS_DIR="${retained_biome_dir}/results" CARTULARY_TEST_RUN_ID=retained-biome "${ROOT_DIR}/tools/harness/output/test-output.sh" target-summary lint-biome pass --quiet-success --suppress-machine-output --preserve-existing-tool-summary; summary_status=\$\$?; else CARTULARY_OUTPUT_MODE=quiet CARTULARY_TEST_RESULTS_DIR="${retained_biome_dir}/results" CARTULARY_TEST_RUN_ID=retained-biome "${ROOT_DIR}/tools/harness/output/test-output.sh" target-summary lint-biome fail --quiet-success --suppress-machine-output --preserve-existing-tool-summary; summary_status=\$\$?; fi; if [ "\$\$summary_status" -ne 0 ]; then exit "\$\$summary_status"; fi; exit "\$\$status"
EOF
set +e
make --no-print-directory -f "${retained_biome_dir}/Makefile" lint-biome \
  >"${retained_biome_dir}/stdout.log" \
  2>"${retained_biome_dir}/stderr.log"
retained_biome_status=$?
set -e
assert_equals "${retained_biome_status}" "2" "retained biome outer Make status"
assert_equals "$(cat "${retained_biome_dir}/stdout.log")" "" "retained biome stdout"
retained_biome_stderr="$(cat "${retained_biome_dir}/stderr.log")"
assert_contains "${retained_biome_stderr}" "[FAIL] target=lint-biome exit_code=1 failure_class=harness reason=tool_diagnostic_failure" "retained biome failure line"
assert_contains "${retained_biome_stderr}" "[ARTIFACTS] target=lint-biome" "retained biome artifact line"
assert_contains "${retained_biome_stderr}" "[RERUN] command=\"make lint-biome\"" "retained biome rerun line"
assert_contains "${retained_biome_stderr}" "[INVESTIGATE] command=\"make explain-target TARGET=lint-biome DETAIL=artifacts\"" "retained biome investigation line"
retained_biome_summary="${retained_biome_dir}/results/retained-biome/lint-biome/tool-run-summary.json"
assert_equals "$(json_field "${retained_biome_summary}" "failure_reason")" "tool_diagnostic_failure" "retained biome summary reason"
assert_equals "$(json_field "${retained_biome_summary}" "exit_code")" "1" "retained biome summary exit code"

suppressed_machine_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-suppressed-machine.XXXXXX")"
cleanup_paths+=("${suppressed_machine_dir}")
CARTULARY_TEST_TARGET=alpha \
CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
CARTULARY_TEST_RESULTS_DIR="${suppressed_machine_dir}/results" \
CARTULARY_TEST_RUN_ID="suppressed-machine" \
  "${ROOT_DIR}/tools/harness/execution/run-step.sh" "alpha smoke" -- true \
  >/dev/null 2>&1
CARTULARY_OUTPUT_MODE=machine \
CARTULARY_SUPPRESS_CHILD_SUCCESS=1 \
CARTULARY_TEST_RESULTS_DIR="${suppressed_machine_dir}/results" \
CARTULARY_TEST_RUN_ID="suppressed-machine" \
TASK_SURFACE_MANIFEST="${sequence_manifest}" \
  "${ROOT_DIR}/tools/harness/output/test-output.sh" target-summary alpha pass \
  >"${suppressed_machine_dir}/stdout.log" \
  2>"${suppressed_machine_dir}/stderr.log"
assert_equals "$(cat "${suppressed_machine_dir}/stdout.log")" "" "suppressed child machine stdout"
assert_equals "$(cat "${suppressed_machine_dir}/stderr.log")" "" "suppressed child machine stderr"
assert_file_present "${suppressed_machine_dir}/results/suppressed-machine/alpha/tool-run-summary.json" "suppressed child machine artifact"

sequence_budget_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-sequence-budget.XXXXXX")"
cleanup_paths+=("${sequence_budget_dir}")
write_fake_make "${sequence_budget_dir}"
VERBOSE="" \
CI_VERBOSE="" \
CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
CARTULARY_OUTPUT_MODE="" \
MAKE="${sequence_budget_dir}/fake-make" \
FAKE_MAKE_LOG="${sequence_budget_dir}/make.log" \
CARTULARY_TEST_RESULTS_DIR="${sequence_budget_dir}/results" \
CARTULARY_TEST_RUN_ID="sequence-budget" \
TASK_SURFACE_MANIFEST="${sequence_manifest}" \
  "${SCRIPT}" --sequence smoke \
  >"${sequence_budget_dir}/stdout.log" \
  2>"${sequence_budget_dir}/stderr.log"
assert_output_budget "${sequence_manifest}" smoke "${sequence_budget_dir}/stdout.log" "${sequence_budget_dir}/stderr.log" "sequence success budget"

machine_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-machine.XXXXXX")"
cleanup_paths+=("${machine_dir}")
write_fake_make "${machine_dir}"
VERBOSE="" \
CI_VERBOSE="" \
CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
CARTULARY_OUTPUT_MODE=machine \
MAKE="${machine_dir}/fake-make" \
FAKE_MAKE_LOG="${machine_dir}/make.log" \
CARTULARY_TEST_RESULTS_DIR="${machine_dir}/results" \
CARTULARY_TEST_RUN_ID="machine" \
TASK_SURFACE_MANIFEST="${sequence_manifest}" \
  "${SCRIPT}" --sequence smoke \
  >"${machine_dir}/stdout.log" \
  2>"${machine_dir}/stderr.log"
assert_single_machine_json \
  "${machine_dir}/stdout.log" \
  "${machine_dir}/stderr.log" \
  smoke \
  "machine sequence summary" \
  tool_run_summary \
  target_summary \
  run_summary \
  run_tool_run_summary
assert_file_present "${machine_dir}/results/machine/run-summary.json" "machine sequence run summary artifact"
assert_file_present "${machine_dir}/results/machine/smoke/tool-run-summary.json" "machine sequence target tool summary artifact"

lifecycle_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-lifecycle.XXXXXX")"
cleanup_paths+=("${lifecycle_dir}")
run_start_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="${lifecycle_dir}/results" \
  CARTULARY_TEST_RUN_ID="lifecycle" \
    "${ROOT_DIR}/tools/harness/output/test-output.sh" run-start "sequence label" --steps 2 --summary-targets 1 --helper-units 0 --jobs 3 --force
)"
assert_equals "${run_start_output}" "[RUN] sequence label work_units=2 summary_targets=1 helper_units=0 jobs=3 run_id=lifecycle" "forced run-start lifecycle output"
step_start_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="${lifecycle_dir}/results" \
  CARTULARY_TEST_RUN_ID="lifecycle" \
    "${ROOT_DIR}/tools/harness/output/test-output.sh" step-start "sequence label" 1 2 alpha --mode parallel --jobs 3 --force
)"
assert_equals "${step_start_output}" "[STEP] sequence label 1/2 alpha mode=parallel jobs=3" "forced step-start lifecycle output"
target_start_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="${lifecycle_dir}/results" \
  CARTULARY_TEST_RUN_ID="lifecycle" \
    "${ROOT_DIR}/tools/harness/output/test-output.sh" target-start alpha --children beta,gamma --service-backed 1 --expected-steps 2 --expected-tests 4 --force
)"
assert_equals "${target_start_output}" "[TARGET] start alpha service_backed=1 expected_steps=2 expected_tests=4 children=beta,gamma" "forced target-start lifecycle output"

go_json_stream_output="$(
  printf '%s\n%s\n%s\n' '{"Output":"hello\n"}' 'not-json' '{"Output":"done"}' |
    NODE_BIN="${NODE_BIN:-}" "${ROOT_DIR}/tools/harness/output/test-output.sh" go-json-stream
)"
assert_equals "${go_json_stream_output}" $'hello\ndone' "go-json-stream unwraps Go output lines"

for aggregate_sequence in test-fast ci release-check; do
  aggregate_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-${aggregate_sequence}.XXXXXX")"
  cleanup_paths+=("${aggregate_dir}")
  write_fake_make "${aggregate_dir}"
  aggregate_output="$(
    VERBOSE="" \
    CI_VERBOSE="" \
    CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
    CARTULARY_OUTPUT_MODE="" \
    MAKE="${aggregate_dir}/fake-make" \
    FAKE_MAKE_LOG="${aggregate_dir}/make.log" \
    FAKE_MAKE_ENV_LOG="${aggregate_dir}/make-env.log" \
    CARTULARY_TEST_RESULTS_DIR="${aggregate_dir}/results" \
    CARTULARY_TEST_RUN_ID="${aggregate_sequence}" \
    TASK_SURFACE_MANIFEST="${ROOT_DIR}/tools/task_surface_manifest.json" \
      "${SCRIPT}" --sequence "${aggregate_sequence}" \
      2>&1
  )"
  assert_contains "${aggregate_output}" "[RESULT] target=${aggregate_sequence} status=pass" "${aggregate_sequence} run summary output"
  assert_contains "${aggregate_output}" "[ARTIFACTS] target=${aggregate_sequence}" "${aggregate_sequence} artifact output"
  assert_file_present "${aggregate_dir}/results/${aggregate_sequence}/${aggregate_sequence}/target-summary.json" "${aggregate_sequence} target summary"
  assert_equals "$(json_field "${aggregate_dir}/results/${aggregate_sequence}/${aggregate_sequence}/target-summary.json" "target")" "${aggregate_sequence}" "${aggregate_sequence} target summary identity"
  assert_equals "$(json_field "${aggregate_dir}/results/${aggregate_sequence}/run-summary.json" "label")" "${aggregate_sequence}" "${aggregate_sequence} run summary identity"
  if [[ "${aggregate_sequence}" == "ci" ]]; then
    assert_contains "$(cat "${aggregate_dir}/make-env.log")" "target=check skip=unset satisfied=unset check_cpu=20 check_io=24" "ci forwards the exact nested check budget"
    "${NODE_BIN:-node}" - "${aggregate_dir}/results/${aggregate_sequence}/${aggregate_sequence}/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const summary = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
if (summary.resource_limits.host_cpu !== 20 || summary.resource_limits.host_io !== 24 || summary.resource_limits.process !== 8) {
  throw new Error(`unexpected adaptive capacity ${JSON.stringify(summary.resource_limits)}`);
}
const nested = summary.nested_scheduler_limits?.find((item) => item.target === "check");
if (nested?.forwarding_profile !== "sequence_to_check" ||
    nested.mappings?.find((item) => item.env_variable === "CHECK_HOST_CPU_JOBS")?.amount !== 20 ||
    nested.mappings?.find((item) => item.env_variable === "CHECK_HOST_IO_JOBS")?.amount !== 24) {
  throw new Error("nested check forwarding evidence is incomplete");
}
EOF
  fi
  if [[ "${aggregate_sequence}" == "release-check" ]]; then
    assert_contains "$(cat "${aggregate_dir}/make-env.log")" "target=release-browser-readiness skip=1 satisfied=1 check_cpu=unset check_io=unset service_cpu=2 service_io=4" "release forwards the exact nested service budget"
  fi
done

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-dry-run.XXXXXX")"
cleanup_paths+=("${dry_run_dir}")
write_fake_make "${dry_run_dir}"
dry_run_output="$(
  VERBOSE="" \
  CI_VERBOSE="" \
  CARTULARY_OUTPUT_MODE="" \
  MAKEFLAGS="n" \
  MAKE="${dry_run_dir}/fake-make" \
  FAKE_MAKE_LOG="${dry_run_dir}/make.log" \
  CARTULARY_TEST_RESULTS_DIR="${dry_run_dir}/results" \
  CARTULARY_TEST_RUN_ID="dry-run" \
  TASK_SURFACE_MANIFEST="${sequence_manifest}" \
    "${SCRIPT}" --sequence dry-run \
    2>&1
)"
assert_not_contains "${dry_run_output}" "[RUN]" "script dry-run run start output"
assert_not_contains "${dry_run_output}" "[STEP]" "script dry-run step output"
assert_file_absent "${dry_run_dir}/results/dry-run/run-summary.json" "script dry-run summary"
assert_contains "${dry_run_output}" "[DRY-RUN] dry-run" "script scheduler dry-run plan"
assert_file_absent "${dry_run_dir}/make.log" "script dry-run child make"

harness_quiet_dir="$(mktemp -d "${ROOT_DIR}/tmp/harness-smoke-quiet.XXXXXX")"
cleanup_paths+=("${harness_quiet_dir}")
mkdir -p "${harness_quiet_dir}/scripts"
cat >"${harness_quiet_dir}/scripts/check-a.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 0
EOF
cat >"${harness_quiet_dir}/scripts/check-b.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 0
EOF
cat >"${harness_quiet_dir}/scripts/check-env.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${CARTULARY_SUPPRESS_CHILD_SUCCESS+x}" == "x" ]]; then
  echo "CARTULARY_SUPPRESS_CHILD_SUCCESS leaked into harness smoke check" >&2
  exit 9
fi
exit 0
EOF
chmod +x \
  "${harness_quiet_dir}/scripts/check-a.sh" \
  "${harness_quiet_dir}/scripts/check-b.sh" \
  "${harness_quiet_dir}/scripts/check-env.sh"
harness_manifest="${harness_quiet_dir}/manifest.json"
"${NODE_BIN:-node}" - "${ROOT_DIR}/tools/task_surface_manifest.json" "${harness_manifest}" "${harness_quiet_dir#"${ROOT_DIR}"/}/scripts" <<'EOF'
const fs = require("node:fs");
const [source, destination, scriptDir] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(source, "utf8"));
for (const check of manifest.harness_checks) {
  delete check.gate_smoke_role;
}
const checks = ["harness-quiet-a", "harness-quiet-b", "harness-quiet-env"];
manifest.harness_tiers.fast = { checks };
manifest.harness_checks.push(
  { name: "harness-quiet-a", gate_smoke_role: "public_make_wrapper", backing_scripts: [`${scriptDir}/check-a.sh`] },
  { name: "harness-quiet-b", gate_smoke_role: "check_scheduler_semantic", backing_scripts: [`${scriptDir}/check-b.sh`] },
  { name: "harness-quiet-env", gate_smoke_role: "service_backed_scheduler_semantic", backing_scripts: [`${scriptDir}/check-env.sh`] },
);
fs.writeFileSync(destination, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
harness_quiet_output="$(
  VERBOSE="" \
  CI_VERBOSE="" \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="${harness_quiet_dir}/results" \
  CARTULARY_TEST_RUN_ID="quiet" \
  TASK_SURFACE_MANIFEST="${harness_manifest}" \
    "${NODE_BIN:-node}" "${ROOT_DIR}/tools/harness/smoke/run-harness-smoke-cli.mjs" --tier fast --jobs 2 --manifest "${harness_manifest}" \
    2>&1
)"
assert_equals "${harness_quiet_output}" "" "quiet harness internal success output"
check_harness_quiet_output="$(
  VERBOSE="" \
  CI_VERBOSE="" \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_TEST_RESULTS_DIR="${harness_quiet_dir}/results" \
  CARTULARY_TEST_RUN_ID="quiet" \
    "${ROOT_DIR}/tools/harness/output/test-output.sh" target-summary check-harness-smoke pass --children harness-quiet-a,harness-quiet-b,harness-quiet-env \
    2>&1
)"
assert_contains "${check_harness_quiet_output}" "[RESULT] target=check-harness-smoke status=pass" "quiet check harness aggregate summary"
assert_contains "${check_harness_quiet_output}" "[ARTIFACTS] target=check-harness-smoke" "quiet check harness artifact summary"
assert_not_contains "${check_harness_quiet_output}" "[CHILD]" "quiet check harness hides child detail"

harness_failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/harness-smoke-failure.XXXXXX")"
cleanup_paths+=("${harness_failure_dir}")
mkdir -p "${harness_failure_dir}/scripts"
cat >"${harness_failure_dir}/scripts/check-fail.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 7
EOF
cat >"${harness_failure_dir}/scripts/check-skipped.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 0
EOF
chmod +x "${harness_failure_dir}/scripts/check-fail.sh" "${harness_failure_dir}/scripts/check-skipped.sh"
harness_failure_manifest="${harness_failure_dir}/manifest.json"
"${NODE_BIN:-node}" - "${ROOT_DIR}/tools/task_surface_manifest.json" "${harness_failure_manifest}" "${harness_failure_dir#"${ROOT_DIR}"/}/scripts" <<'EOF'
const fs = require("node:fs");
const [source, destination, scriptDir] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(source, "utf8"));
for (const check of manifest.harness_checks) {
  delete check.gate_smoke_role;
}
const checks = ["harness-fail-a", "harness-skipped-b", "harness-skipped-c"];
manifest.harness_tiers.fast = { checks };
manifest.harness_checks.push(
  { name: "harness-fail-a", gate_smoke_role: "public_make_wrapper", backing_scripts: [`${scriptDir}/check-fail.sh`] },
  { name: "harness-skipped-b", gate_smoke_role: "check_scheduler_semantic", backing_scripts: [`${scriptDir}/check-skipped.sh`] },
  { name: "harness-skipped-c", gate_smoke_role: "service_backed_scheduler_semantic", backing_scripts: [`${scriptDir}/check-skipped.sh`] },
);
fs.writeFileSync(destination, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
set +e
harness_failure_output="$(
  VERBOSE="" \
  CI_VERBOSE="" \
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_TEST_RESULTS_DIR="${harness_failure_dir}/results" \
  CARTULARY_TEST_RUN_ID="failure" \
  TASK_SURFACE_MANIFEST="${harness_failure_manifest}" \
    "${NODE_BIN:-node}" "${ROOT_DIR}/tools/harness/smoke/run-harness-smoke-cli.mjs" --tier fast --jobs 1 --manifest "${harness_failure_manifest}" \
    2>&1
)"
harness_failure_status=$?
set -e
assert_equals "${harness_failure_status}" "1" "failing harness normalizes unknown child status"
assert_contains "${harness_failure_output}" "[FAIL] target=run-harness-smoke-fast" "failing harness reports concise failure"
assert_contains "${harness_failure_output}" "[ARTIFACTS] target=run-harness-smoke-fast root=" "failing harness reports artifacts"
assert_not_contains "${harness_failure_output}" "[CHILD-MISSING] run-harness-smoke-fast harness-skipped-b" "failing harness does not report skipped child missing"
assert_not_contains "${harness_failure_output}" "missing child target summary: harness-skipped-b" "failing harness does not create missing child artifact failure"
harness_failure_summary="${harness_failure_dir}/results/failure/run-harness-smoke-fast/target-summary.json"
assert_equals "$(json_field "${harness_failure_summary}" "children.missing.length")" "0" "failing harness skipped child missing list"
assert_equals "$(json_field "${harness_failure_summary}" "children.skipped.0.target")" "harness-skipped-b" "failing harness skipped child target"
assert_equals "$(json_field "${harness_failure_summary}" "children.skipped.0.reason")" "schedule_stopped_after_failure" "failing harness skipped child reason"
assert_equals "$(json_field "${harness_failure_summary}" "children.skipped.0.failed_dependency")" "harness-fail-a" "failing harness skipped child dependency"
assert_equals "$(json_field "${harness_failure_summary}" "children.failed_targets.0")" "harness-fail-a" "failing harness failed child"
check_harness_failure_output="$(
  CARTULARY_OUTPUT_MODE=quiet \
  CARTULARY_SUPPRESS_CHILD_SUCCESS=0 \
  CARTULARY_TEST_RESULTS_DIR="${harness_failure_dir}/results" \
  CARTULARY_TEST_RUN_ID="failure" \
  TASK_SURFACE_MANIFEST="${harness_failure_manifest}" \
    "${ROOT_DIR}/tools/harness/output/test-output.sh" target-summary check-harness-smoke fail \
      --projection check-harness-smoke \
      --skipped-from-child run-harness-smoke-fast \
    2>&1
)"
assert_contains "${check_harness_failure_output}" "[FAIL] target=check-harness-smoke" "projected check harness reports concise failure"
assert_contains "${check_harness_failure_output}" "[ARTIFACTS] target=check-harness-smoke root=" "projected check harness reports artifacts"
assert_not_contains "${check_harness_failure_output}" "[CHILD-MISSING] check-harness-smoke harness-skipped-b" "projected check harness does not report skipped child missing"
assert_not_contains "${check_harness_failure_output}" "missing child target summary: harness-skipped-b" "projected check harness avoids skipped child artifact failure"
check_harness_failure_summary="${harness_failure_dir}/results/failure/check-harness-smoke/target-summary.json"
assert_equals "$(json_field "${check_harness_failure_summary}" "children.skipped.0.target")" "harness-skipped-b" "projected check harness skipped child target"
assert_equals "$(json_field "${check_harness_failure_summary}" "children.missing.length")" "0" "projected check harness skipped child missing list"

invalid_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-invalid.XXXXXX")"
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
if [[ "${invalid_status}" != "2" ]]; then
  fail "invalid usage status: expected [2], got [${invalid_status}]"
fi
assert_contains "${invalid_output}" "usage: run-make-sequence.sh --sequence <name>" "invalid usage output"
assert_file_absent "${invalid_dir}/make.log" "invalid usage child make log"

set +e
invalid_suppress_output="$(
  "${ROOT_DIR}/tools/harness/output/test-output.sh" run-summary smoke pass 0 0 - --suppress-machine-output=1 \
    2>&1
)"
invalid_suppress_status=$?
set -e
if [[ "${invalid_suppress_status}" != "1" ]]; then
  fail "invalid suppress-machine-output status: expected [1], got [${invalid_suppress_status}]"
fi
assert_contains "${invalid_suppress_output}" "unknown run-summary option --suppress-machine-output=1" "invalid suppress-machine-output output"

env NODE_BIN="${NODE_BIN:-node}" "${NODE_BIN:-node}" - "${ROOT_DIR}/tools/task_surface_manifest.json" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");

const [manifestPath] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
const repoRoot = path.dirname(path.dirname(manifestPath));
const { fast, execution, extended, lifecycle, full } = manifest.harness_tiers;

function fail(message) {
  console.error(message);
  process.exit(1);
}

const fullOnlyChecks = new Set(["harness-smoke-tool-output-real-targets"]);
const expectedFullBase = [...fast.checks, ...extended.checks, ...lifecycle.checks];
const filteredFull = full.checks.filter((check) => !fullOnlyChecks.has(check));
if (JSON.stringify(filteredFull) !== JSON.stringify(expectedFullBase)) {
  fail("full harness tier must contain fast + extended + lifecycle tiers in order");
}
for (const check of fullOnlyChecks) {
  if (!full.checks.includes(check)) {
    fail(`full harness tier missing ${check}`);
  }
}

const expectedFast = [
  "harness-smoke-public-make-wrapper",
  "harness-smoke-check-scheduler-smoke",
  "harness-smoke-service-backed-scheduler-smoke",
];
if (JSON.stringify(fast.checks) !== JSON.stringify(expectedFast)) {
  fail(`fast harness tier must be the check-gate smoke set, got ${fast.checks.join(",")}`);
}

const checksByName = new Map(manifest.harness_checks.map((check) => [check.name, check]));
const requiredScratchHelpers = new Map([
  ["tools/harness/tests/test-public-make-wrapper-smoke.sh", ["cartulary_harness_mktemp_dir \"public-make-wrapper.XXXXXX\""]],
  ["tools/harness/scheduler/tests/test-check-scheduler.sh", [
    "cartulary_harness_mktemp_dir \"check-scheduler-smoke.XXXXXX\"",
    "cartulary_harness_mktemp_dir \"check-scheduler-smoke-service-timing.XXXXXX\"",
  ]],
  ["tools/harness/scheduler/tests/test-service-backed-scheduler.sh", ["cartulary_harness_mktemp_dir \"service-backed-scheduler-smoke.XXXXXX\""]],
]);
for (const check of fast.checks) {
  for (const script of checksByName.get(check)?.backing_scripts ?? []) {
    if (!script.endsWith(".sh")) {
      continue;
    }
    const content = fs.readFileSync(path.join(repoRoot, script), "utf8");
    if (!content.includes("harness-scratch.sh")) {
      fail(`${script} must source harness-scratch.sh for fast smoke scratch`);
    }
    for (const required of requiredScratchHelpers.get(script) ?? []) {
      if (!content.includes(required)) {
        fail(`${script} missing ${required}`);
      }
    }
    if (/mktemp -d "\$\{?ROOT_DIR\}?\/(?:tmp|\.cartulary\/tmp)\/(?:public-make-wrapper|check-scheduler-smoke|check-scheduler-smoke-service-timing|service-backed-scheduler-smoke)\.XXXXXX"/.test(content)) {
      fail(`${script} must not create fast smoke fixtures with raw repo-local mktemp`);
    }
    if (/CARTULARY_HARNESS_SCRATCH_ROOT:-\$\{ROOT_DIR\}\/\.cartulary\/tmp/.test(content)) {
      fail(`${script} must not default harness scratch inside the repository`);
    }
  }
}

const tierMembership = new Map();
for (const [tier, checks] of [["fast", fast.checks], ["extended", extended.checks], ["lifecycle", lifecycle.checks]]) {
  for (const check of checks) {
    if (tierMembership.has(check)) {
      fail(`${check} is present in both ${tierMembership.get(check)} and ${tier}`);
    }
    tierMembership.set(check, tier);
  }
}

const expectedExecutionChecks = [
  "harness-smoke-run-make-sequence-fast",
  "harness-smoke-cartulary-runner-service-backed-target",
  "harness-smoke-make-node-tools",
];
if (JSON.stringify(execution.checks) !== JSON.stringify(expectedExecutionChecks)) {
  fail(`execution harness tier must be the execution wrapper smoke set, got ${execution.checks.join(",")}`);
}
for (const target of expectedExecutionChecks) {
  if (tierMembership.get(target) !== "extended") {
    fail(`${target} must remain in extended harness smoke`);
  }
}

for (const target of ["harness-smoke-run-make-sequence", "harness-smoke-run-go-target"]) {
  if (tierMembership.get(target) !== "extended") {
    fail(`${target} must stay in extended harness smoke`);
  }
}
for (const target of ["harness-smoke-make-node-tools", "harness-smoke-check-scheduler"]) {
  if (tierMembership.get(target) !== "extended") {
    fail(`${target} must stay in extended harness smoke`);
  }
}
for (const retired of [
  "harness-smoke-frontend-evidence-audit",
  "harness-smoke-guidance-core",
  "harness-smoke-guidance-matrix",
  "harness-smoke-run-go-target-fast",
  "harness-smoke-service-backed-scheduler-fast",
]) {
  if (manifest.harness_checks.some((check) => check.name === retired) || tierMembership.has(retired)) {
    fail(`${retired} must be retired from harness smoke`);
  }
}
EOF

run_fast_block="$(extract_target_definition run-harness-smoke-fast)"
run_extended_block="$(extract_target_definition run-harness-smoke-extended)"
run_execution_block="$(extract_target_definition run-harness-smoke-execution)"
run_lifecycle_block="$(extract_target_definition run-harness-smoke-lifecycle)"
run_full_block="$(extract_target_definition run-harness-smoke-full)"
check_harness_smoke_block="$(extract_target_definition check-harness-smoke)"
release_check_block="$(extract_target_definition release-check)"
ci_block="$(extract_target_definition ci)"
test_fast_block="$(extract_target_definition test-fast)"

assert_contains "${test_fast_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT) --sequence test-fast' "test-fast sequence runner"
assert_not_contains "${run_fast_block}" '$(FRONTEND_INSTALL_STAMP)' "fast harness smoke does not require frontend install"
assert_not_contains "${run_execution_block}" '$(FRONTEND_INSTALL_STAMP)' "execution harness smoke does not require frontend install"
assert_contains "${run_fast_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier fast --jobs "$(HARNESS_SMOKE_JOBS)"' "fast harness manifest runner"
assert_contains "${run_extended_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier extended --jobs "$(HARNESS_SMOKE_JOBS)"' "extended harness manifest runner"
assert_contains "${run_execution_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier execution --jobs "$(HARNESS_SMOKE_JOBS)"' "execution harness manifest runner"
assert_contains "${run_lifecycle_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier lifecycle --jobs "$(HARNESS_SMOKE_JOBS)"' "lifecycle harness manifest runner"
assert_contains "${run_full_block}" '$(RUN_HARNESS_SMOKE_SCRIPT) --tier full --jobs "$(HARNESS_SMOKE_JOBS)"' "full harness manifest runner"
assert_contains "${check_harness_smoke_block}" "run-harness-smoke-fast" "check harness fast tier"
assert_contains "${check_harness_smoke_block}" "--projection check-harness-smoke" "check harness summary projection"
assert_contains "${ci_block}" "ci: export CI := 1" "CI target sets CI output-mode signal"
assert_contains "${ci_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT) --sequence ci' "CI target delegates through canonical Make sequence"
assert_contains "${release_check_block}" '$(RUN_MAKE_SEQUENCE_SCRIPT) --sequence release-check' "release-check sequence runner"
assert_contains "$(cat "${ROOT_DIR}/tools/task_surface_manifest.json")" "tools/harness/tests/test-run-make-sequence-fast.sh" "fast make-sequence smoke backing script"
assert_contains "$(cat "${ROOT_DIR}/tools/task_surface_manifest.json")" "tools/harness/scheduler/tests/test-check-scheduler.sh" "fast check scheduler smoke backing script"
assert_contains "$(cat "${ROOT_DIR}/tools/task_surface_manifest.json")" "tools/harness/scheduler/tests/test-service-backed-scheduler.sh" "fast service-backed scheduler smoke backing script"

"${NODE_BIN:-node}" --input-type=module <<'EOF'
import assert from "node:assert/strict";
import {
  chmodSync,
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  rmSync,
  symlinkSync,
  statSync,
  writeFileSync,
} from "node:fs";
import path from "node:path";

import {
  HarnessConfigError,
  generateRunId,
  redactString,
  redactValue,
  resolveHarnessConfig,
  resolveOutputMode,
  runCleanup,
  secureWriteFile,
  validateSchemaSync,
} from "./tools/harness/contract/harness-contract.mjs";
import {
  classifyExecutionFailure,
  classifyExecutionFailureReason,
  publicExitCodeForFailure,
  publicExitCodeForFailures,
  publicExitCodeForSummary,
} from "./tools/harness/contract/failure-taxonomy.mjs";

assert.equal(
  resolveOutputMode({ CARTULARY_OUTPUT_MODE: "quiet", VERBOSE: "1" }, "backend-unit"),
  "quiet",
);
assert.equal(
  resolveOutputMode({ VERBOSE: "1", CI_VERBOSE: "1" }, "backend-unit"),
  "verbose",
);
assert.equal(resolveOutputMode({ CI_VERBOSE: "1" }, "backend-unit"), "ci");
assert.equal(resolveOutputMode({ CI: "1" }, "backend-unit"), "ci");
assert.equal(resolveOutputMode({}, "ci"), "ci");
assert.equal(classifyExecutionFailure("deployable-shape"), "artifact");
assert.equal(classifyExecutionFailureReason("lint-biome"), "unknown_failure");

assert.throws(
  () => resolveOutputMode({ CARTULARY_OUTPUT_MODE: "bogus" }, "backend-unit"),
  (error) =>
    error instanceof HarnessConfigError &&
    error.failure_reason === "configuration_error" &&
    error.exit_code === 2,
);
assert.throws(
  () => resolveHarnessConfig("help", { CARTULARY_OUTPUT_MODE: "machine" }),
  (error) =>
    error instanceof HarnessConfigError &&
    error.failure_reason === "usage_error" &&
    error.exit_code === 2,
);

assert.equal(publicExitCodeForFailure({ failure_reason: "configuration_error" }), 2);
assert.equal(publicExitCodeForFailure({ failure_reason: "fixture_error" }), 3);
assert.equal(publicExitCodeForFailure({ failure_reason: "resource_conflict" }), 4);
assert.equal(publicExitCodeForFailure({ failure_reason: "test_assertion_failure" }), 10);
assert.equal(publicExitCodeForFailure({ failure_reason: "tool_diagnostic_failure" }), 1);
assert.equal(publicExitCodeForFailure({ failure_reason: "artifact_error" }), 11);
assert.equal(publicExitCodeForFailure({ failure_reason: "cleanup_error" }), 12);
assert.equal(publicExitCodeForFailure({ failure_reason: "duration_baseline_drift" }), 13);
assert.equal(publicExitCodeForFailure({ failure_reason: "timeout_failure" }), 13);
assert.equal(publicExitCodeForFailure({ failure_reason: "cancelled_or_interrupted" }, { signal: "SIGINT" }), 130);
assert.equal(publicExitCodeForFailures([
  { failure_reason: "cleanup_error" },
  { failure_reason: "test_assertion_failure" },
]), 10);
assert.equal(publicExitCodeForSummary({
  status: "fail",
  failure_reason: "child_target_failure",
  failure_class: "harness",
  failures: [{ failure_reason: "child_target_failure", child_target: "child" }],
}, {
  childSummaries: [{ target: "child", failure_reason: "test_assertion_failure", failure_class: "product" }],
}), 10);
assert.equal(publicExitCodeForSummary({
  status: "fail",
  failure_reason: "child_target_failure",
  failure_class: "harness",
  failures: [{ failure_reason: "child_target_failure", child_target: "missing" }],
}, {
  childSummaries: [],
}), 1);

const resolved = resolveHarnessConfig("backend-unit", {
  CARTULARY_TEST_RESULTS_DIR: "tmp/results",
  CARTULARY_TEST_RUN_ID: "run-a",
  CHECK_HOST_CPU_JOBS: "2",
  RANDOM_JOBS: "bad",
});
assert.equal(resolved.output_mode, "summary");
assert.equal(resolved.resource_limits.scheduler_overrides.CHECK_HOST_CPU_JOBS, 2);

const unrelatedSuffixResolved = resolveHarnessConfig("backend-unit", {
  CARTULARY_TEST_RESULTS_DIR: "tmp/results",
  CARTULARY_TEST_RUN_ID: "run-a",
  FOO_JOBS: "abc",
  BAR_WORKERS: "abc",
  BAZ_SHARDS: "abc",
});
assert.equal(unrelatedSuffixResolved.output_mode, "summary");

assert.throws(
  () =>
    resolveHarnessConfig("backend-unit", {
      CARTULARY_TEST_RESULTS_DIR: "",
      CARTULARY_TEST_RUN_ID: "run-a",
    }),
  HarnessConfigError,
);
assert.throws(
  () =>
    resolveHarnessConfig("backend-unit", {
      CARTULARY_TEST_RESULTS_DIR: "tmp/results",
      CARTULARY_TEST_RUN_ID: "run-a",
      PLAYWRIGHT_WORKERS: "abc",
    }),
  HarnessConfigError,
);
assert.throws(
  () =>
    resolveHarnessConfig("backend-unit", {
      CARTULARY_TEST_RESULTS_DIR: "tmp/results",
      CARTULARY_TEST_RUN_ID: "run-a",
      CHECK_HOST_CPU_JOBS: "0",
    }),
  HarnessConfigError,
);
assert.throws(
  () =>
    resolveHarnessConfig("backend-unit", {
      CARTULARY_TEST_RESULTS_DIR: "tmp/results",
      CARTULARY_TEST_RUN_ID: "run-a",
      CARTULARY_PGTEST_ADMIN_DSN: "postgres://admin",
    }),
  HarnessConfigError,
);

const secretText = [
  "Authorization: Bearer abc.def.secret",
  "Cookie: session=super-cookie",
  "Set-Cookie: cartulary=super-cookie",
  "X-Cartulary-Test-Route-Token: route-secret",
  "postgres://cartulary:supersecret@127.0.0.1:5432/postgres",
  "--token route-secret --password=supersecret",
  "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJ1In0.signature",
  "-----BEGIN PRIVATE KEY-----\nsecret\n-----END PRIVATE KEY-----",
].join("\n");
const redactedText = redactString(secretText);
for (const secret of ["abc.def.secret", "super-cookie", "route-secret", "supersecret", "signature", "BEGIN PRIVATE KEY"]) {
  assert.equal(redactedText.includes(secret), false, `raw redaction leaked ${secret}: ${redactedText}`);
}
assert.match(redactedText, /\[REDACTED/u);

const redactedJSON = redactValue({
  nested: {
    Authorization: "Bearer nested.secret",
    cookie: "session=nested-cookie",
    CARTULARY_S3TEST_SECRET_ACCESS_KEY: "object-store-secret",
  },
  service_sessions: [
    {
      target: "service-timing-suite",
      cleanup_status: "pass",
      setup_duration_ms: 12,
      session_token: "nested-token",
    },
  ],
  args: ["--token", "nested-token"],
});
assert.equal(redactedJSON.nested.Authorization, "[REDACTED]");
assert.equal(redactedJSON.nested.cookie, "[REDACTED]");
assert.equal(redactedJSON.nested.CARTULARY_S3TEST_SECRET_ACCESS_KEY, "[REDACTED]");
assert.equal(Array.isArray(redactedJSON.service_sessions), true);
assert.equal(redactedJSON.service_sessions[0].target, "service-timing-suite");
assert.equal(redactedJSON.service_sessions[0].cleanup_status, "pass");
assert.equal(redactedJSON.service_sessions[0].setup_duration_ms, 12);
assert.equal(JSON.stringify(redactedJSON).includes("nested-token"), false);

const tempRoot = mkdtempSync(path.join(process.cwd(), "tmp", "harness-contract."));
try {
  const resultRoot = path.join(tempRoot, "results");
  const staleRunRoot = path.join(resultRoot, "stale-run");
  mkdirSync(staleRunRoot, { recursive: true });
  writeFileSync(path.join(staleRunRoot, "stale.txt"), "stale\n");
  assert.throws(
    () =>
      resolveHarnessConfig(
        "bootstrap-node-runtime",
        {
          CARTULARY_TEST_RESULTS_DIR: resultRoot,
          CARTULARY_TEST_RUN_ID: "stale-run",
        },
        { prepareRetainedArtifacts: true },
      ),
    HarnessConfigError,
  );
  const nestedResolved = resolveHarnessConfig(
    "bootstrap-node-runtime",
    {
      CARTULARY_TEST_RESULTS_DIR: resultRoot,
      CARTULARY_TEST_RUN_ID: "stale-run",
      CARTULARY_SUPPRESS_CHILD_SUCCESS: "1",
    },
    { prepareRetainedArtifacts: true },
  );
  assert.equal(nestedResolved.run_root, staleRunRoot);
  assert.equal((statSync(nestedResolved.run_root).mode & 0o077), 0);

  const emptyRunRoot = path.join(resultRoot, "empty-run");
  mkdirSync(emptyRunRoot, { recursive: true });
  const emptyResolved = resolveHarnessConfig(
    "bootstrap-node-runtime",
    {
      CARTULARY_TEST_RESULTS_DIR: resultRoot,
      CARTULARY_TEST_RUN_ID: "empty-run",
    },
    { prepareRetainedArtifacts: true },
  );
  assert.equal(emptyResolved.run_root, emptyRunRoot);
  assert.equal(existsSync(emptyRunRoot), true);
  assert.equal((statSync(emptyRunRoot).mode & 0o077), 0);

  const generatedAt = new Date("2026-01-02T03:04:05Z");
  const generatedBase = generateRunId(generatedAt, 1234);
  mkdirSync(path.join(resultRoot, generatedBase), { recursive: true });
  mkdirSync(path.join(resultRoot, `${generatedBase}-n1`), { recursive: true });
  const generatedResolved = resolveHarnessConfig(
    "bootstrap-node-runtime",
    { CARTULARY_TEST_RESULTS_DIR: resultRoot },
    {
      prepareRetainedArtifacts: true,
      materializeGeneratedRunId: true,
      now: generatedAt,
      pid: 1234,
    },
  );
  assert.equal(generatedResolved.run_id, `${generatedBase}-n2`);
  assert.equal(existsSync(generatedResolved.run_root), true);
  assert.equal((statSync(generatedResolved.run_root).mode & 0o077), 0);

  const secureFile = path.join(generatedResolved.run_root, "secure.json");
  secureWriteFile(secureFile, JSON.stringify({ status: "ok" }));
  assert.equal((statSync(secureFile).mode & 0o077), 0);
  assert.equal(readFileSync(secureFile, "utf8").includes("ok"), true);

  assert.throws(
    () => runCleanup({ scope: "clean", candidates: [path.join(process.cwd(), "apps")], includeTmp: false, stdout: { write() {} } }),
    (error) => error instanceof HarnessConfigError && /protected_root/u.test(error.message),
  );

  const externalFile = path.join(tempRoot, "external.txt");
  writeFileSync(externalFile, "keep\n");
  const cleanupLink = path.join(resultRoot, "link-to-external");
  symlinkSync(externalFile, cleanupLink);
  runCleanup({ scope: "clean", candidates: [cleanupLink], includeTmp: false, stdout: { write() {} } });
  assert.equal(existsSync(cleanupLink), false);
  assert.equal(readFileSync(externalFile, "utf8"), "keep\n");

  const unsafeRoot = path.join(tempRoot, "unsafe-root");
  mkdirSync(unsafeRoot, { recursive: true });
  chmodSync(unsafeRoot, 0o777);
  assert.throws(
    () =>
      resolveHarnessConfig(
        "bootstrap-node-runtime",
        {
          CARTULARY_TEST_RESULTS_DIR: unsafeRoot,
          CARTULARY_TEST_RUN_ID: "run-a",
        },
        { prepareRetainedArtifacts: true },
      ),
    HarnessConfigError,
  );
} finally {
  rmSync(tempRoot, { recursive: true, force: true });
}

assert.throws(
  () => validateSchemaSync("cartulary.test_step_summary.v1", {
    schema_id: "cartulary.test_step_summary.v1",
  }),
  /validation failed/u,
);

EOF

assert_equals "$(json_field "${retained_biome_dir}/results/retained-biome/lint-biome/lint-biome/step-summary.json" "schema_id")" "cartulary.test_step_summary.v1" "retained biome step summary schema"
assert_equals "$(json_field "${success_dir}/results/success/smoke/target-summary.json" "schema_id")" "cartulary.test_target_summary.v4" "success target summary schema"
assert_equals "$(json_field "${machine_dir}/results/machine/run-summary.json" "schema_id")" "cartulary.test_run_summary.v6" "machine run summary schema"

verbose_ci_dry_run="$(
  CARTULARY_TEST_RESULTS_DIR="${ROOT_DIR}/tmp/run-make-sequence-fast-output-mode-results" \
  CARTULARY_TEST_RUN_ID="verbose-ci" \
  VERBOSE=1 \
  CI_VERBOSE=1 \
    make -n --no-print-directory lint-biome \
    2>&1
)"
assert_not_contains "${verbose_ci_dry_run}" "--reporter=summary" "VERBOSE must win over CI_VERBOSE in Make output mode"

quiet_verbose_dry_run="$(
  CARTULARY_TEST_RESULTS_DIR="${ROOT_DIR}/tmp/run-make-sequence-fast-output-mode-results" \
  CARTULARY_TEST_RUN_ID="quiet-verbose" \
  CARTULARY_OUTPUT_MODE=quiet \
  VERBOSE=1 \
    make -n --no-print-directory lint-biome \
    2>&1
)"
assert_contains "${quiet_verbose_dry_run}" "--reporter=summary" "explicit CARTULARY_OUTPUT_MODE must win over VERBOSE in Make output mode"

invalid_preflight_output="$(
  set +e
  CARTULARY_TEST_RESULTS_DIR="" \
  CARTULARY_TEST_RUN_ID="empty-results" \
    make --no-print-directory help \
    2>&1
  printf 'status=%s\n' "$?"
)"
assert_contains "${invalid_preflight_output}" "failure_reason=configuration_error" "empty result root fails in preflight"
assert_contains "${invalid_preflight_output}" "status=2" "empty result root exits 2"
assert_not_contains "${invalid_preflight_output}" "Cartulary compact workflow task surface" "empty result root stops child help output"

collision_preflight_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-run-root-collision.XXXXXX")"
cleanup_paths+=("${collision_preflight_dir}")
mkdir -p "${collision_preflight_dir}/results/stale-run"
printf 'stale\n' >"${collision_preflight_dir}/results/stale-run/stale.txt"
collision_preflight_output="$(
  set +e
  env -u CARTULARY_HARNESS_IDENTITY_PREPARED -u CARTULARY_TEST_TARGET \
    CARTULARY_TEST_RESULTS_DIR="${collision_preflight_dir}/results" \
    CARTULARY_TEST_RUN_ID="stale-run" \
    make --no-print-directory doctor \
    2>&1
  printf 'status=%s\n' "$?"
)"
assert_contains "${collision_preflight_output}" "failure_reason=configuration_error" "non-empty run root fails in preflight"
assert_contains "${collision_preflight_output}" "non-empty run root" "non-empty run root diagnostic"
assert_contains "${collision_preflight_output}" "status=2" "non-empty run root exits 2"
assert_file_absent "${collision_preflight_dir}/results/stale-run/tool-run-summary.json" "collision preflight stops wrapper summary"

generated_make="$(cat "${task_surface_generated_make_file}")"
assert_contains "${generated_make}" "backend-integration: export CARTULARY_TEST_TARGET ?= backend-integration" "backend-integration recipe exists"
assert_contains "${generated_make}" "CARTULARY_HARNESS_IDENTITY_PREPARED=1 GO_TEST_PACKAGE_PARALLELISM" "backend-integration child runner reuses prepared public identity"

stale_embed_dir="$(cartulary_harness_mktemp_dir "run-make-sequence-fast-stale-embed.XXXXXX")"
cleanup_paths+=("${stale_embed_dir}")
mkdir -p "${stale_embed_dir}/web-dist/assets" "${stale_embed_dir}/embed/dist" "${stale_embed_dir}/frontend-embed" "${stale_embed_dir}/gomod" "${stale_embed_dir}/gocache"
printf '<div id="root"></div>\n' >"${stale_embed_dir}/web-dist/index.html"
printf 'asset\n' >"${stale_embed_dir}/web-dist/assets/app.js"
printf 'source=stale\n' >"${stale_embed_dir}/frontend-embed/web-assets.stamp"
printf 'stale server\n' >"${stale_embed_dir}/server"
fake_go="${stale_embed_dir}/fake-go"
cat >"${fake_go}" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "run" && "${2:-}" == "./tools/embedwebassets" ]]; then
  output=""
  asset_manifest=""
  client_support_registry=""
  while [[ "$#" -gt 0 ]]; do
    case "$1" in
      --output)
        output="$2"
        shift 2
        ;;
      --asset-manifest)
        asset_manifest="$2"
        shift 2
        ;;
      --client-support-registry)
        client_support_registry="$2"
        shift 2
        ;;
      *)
        shift
        ;;
    esac
  done
  if [[ -z "${output}" || -z "${asset_manifest}" || -z "${client_support_registry}" ]]; then
    echo "fake embedwebassets requires all packaged outputs" >&2
    exit 2
  fi
  mkdir -p "$(dirname "${output}")"
  printf 'fake embedded web archive\n' >"${output}"
  printf '{"assets":[],"schema_id":"cartulary.client_asset_set_manifest.v1"}\n' >"${asset_manifest}"
  printf '{"asset_set_sha256":"fake","client_build_class":"standard","client_build_id":"fake","profiles":[],"schema_id":"cartulary.client_extension_support_registry.v1"}\n' >"${client_support_registry}"
  exit 0
fi

output=""
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    -o)
      output="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done
if [[ -n "${output}" ]]; then
  mkdir -p "$(dirname "${output}")"
  printf 'fake server\n' >"${output}"
fi
EOF
chmod +x "${fake_go}"
CARTULARY_TEST_RESULTS_DIR="${stale_embed_dir}/results" \
CARTULARY_TEST_RUN_ID="stale-embed" \
CARTULARY_BUILD_CACHE_DIR="${stale_embed_dir}/cache/build" \
GO="${fake_go}" \
GO_CACHE_DIR="${stale_embed_dir}/gocache" \
GO_MOD_CACHE_DIR="${stale_embed_dir}/gomod" \
SERVER_BIN="${stale_embed_dir}/server" \
WEB_DIST_INDEX="${stale_embed_dir}/web-dist/index.html" \
EMBEDDED_WEB_ASSET_DIR="${stale_embed_dir}/embed/dist" \
EMBEDDED_WEB_ASSET_ARCHIVE="${stale_embed_dir}/embed/dist/web-assets.zip" \
EMBEDDED_WEB_ASSET_STAMP="${stale_embed_dir}/frontend-embed/web-assets.stamp" \
EMBEDDED_WEB_ASSET_READY_STAMP="${stale_embed_dir}/frontend-embed/web-assets.ready" \
  make -e --no-print-directory build-server \
  >/dev/null
assert_file_present "${stale_embed_dir}/embed/dist/web-assets.zip" "stale embedded web archive is rebuilt"
assert_file_present "${stale_embed_dir}/embed/dist/client-asset-set-manifest.json" "stale client asset manifest is rebuilt"
assert_file_present "${stale_embed_dir}/embed/dist/client-extension-support-registry.json" "stale client support registry is rebuilt"
assert_file_present "${stale_embed_dir}/frontend-embed/web-assets.stamp" "stale embedded web stamp is refreshed"
assert_file_present "${stale_embed_dir}/frontend-embed/web-assets.ready" "stale embedded web ready stamp is refreshed"
assert_file_absent "${stale_embed_dir}/embed/dist/index.html" "stale embedded web fixture must not restore legacy loose index"
assert_contains "$(cat "${stale_embed_dir}/server")" "fake server" "stale embedded web refresh rebuilds server"

for target in run-harness-smoke-fast run-harness-smoke-execution run-harness-smoke-extended run-harness-smoke-lifecycle run-harness-smoke-full; do
  make_dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/run-make-sequence-fast-make-n-${target}.XXXXXX")"
  cleanup_paths+=("${make_dry_run_dir}")
  make_dry_run_output="$(
    CARTULARY_TEST_RESULTS_DIR="${make_dry_run_dir}/results" \
    CARTULARY_TEST_RUN_ID="make-n-${target}" \
      make -n --no-print-directory "${target}" \
      2>&1
  )"
  assert_contains "${make_dry_run_output}" "tools/harness/smoke/run-harness-smoke-cli.mjs --tier ${target#run-harness-smoke-}" "make -n ${target} helper command"
  assert_file_absent "${make_dry_run_dir}/results/make-n-${target}/run-summary.json" "make -n ${target} summary"
done
