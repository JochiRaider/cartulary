#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT_DIR}/scripts/run-check-schedule.mjs"
TEST_OUTPUT_SCRIPT="${ROOT_DIR}/scripts/lib/test-output.sh"
NODE_BIN="${NODE_BIN:-node}"
cleanup_paths=()
SUITE="${1:-all}"

case "$SUITE" in
  all | smoke) ;;
  *)
    echo "usage: test-check-scheduler.sh [smoke]" >&2
    exit 2
    ;;
esac

unset VERBOSE CI_VERBOSE CARTULARY_OUTPUT_MODE CARTULARY_SUPPRESS_CHILD_SUCCESS LINT_SHELL_STRICT

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

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: expected output not to contain [$needle]"
  fi
}

assert_occurrences() {
  local haystack="$1"
  local needle="$2"
  local expected="$3"
  local label="$4"
  local remaining="$haystack"
  local actual=0

  while [[ "$remaining" == *"$needle"* ]]; do
    remaining="${remaining#*"$needle"}"
    actual=$((actual + 1))
  done

  assert_equals "$actual" "$expected" "$label"
}

assert_file_absent() {
  local path="$1"
  local label="$2"

  if [[ -e "$path" ]]; then
    fail "$label: expected $path to be absent"
  fi
}

assert_file_present() {
  local path="$1"
  local label="$2"

  if [[ ! -f "$path" ]]; then
    fail "$label: expected $path to exist"
  fi
}

assert_check_scheduler_artifacts() {
  local dir="$1"
  local run_id="$2"
  local target="$3"
  local expected_status="$4"
  local expected_failed="$5"
  local expected_total="$6"
  local expected_event="$7"
  local summary_file="${dir}/results/${run_id}/${target}/scheduler-summary.json"
  local events_file="${dir}/results/${run_id}/${target}/scheduler-events.jsonl"
  local pressure_file="${dir}/results/${run_id}/${target}/pressure-summary.json"
  local progress_file="${dir}/results/${run_id}/${target}/progress-summary.log"

  assert_file_present "$summary_file" "$target scheduler summary"
  assert_file_present "$events_file" "$target scheduler events"
  assert_file_present "$pressure_file" "$target pressure summary"
  assert_file_present "$progress_file" "$target progress summary"
  "$NODE_BIN" - "$summary_file" "$events_file" "$pressure_file" "$progress_file" "$expected_status" "$expected_failed" "$expected_total" "$expected_event" "$ROOT_DIR" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [summaryFile, eventsFile, pressureFile, progressFile, expectedStatus, expectedFailed, expectedTotal, expectedEvent, repoRoot] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).filter(Boolean).map((line) => JSON.parse(line));
const pressure = JSON.parse(fs.readFileSync(pressureFile, "utf8"));
const progressLog = fs.readFileSync(progressFile, "utf8");
const resolveArtifact = (artifactPath) => path.resolve(repoRoot, artifactPath);
const assertRepoRelativeArtifact = (artifactPath, label) => {
  if (!artifactPath || typeof artifactPath !== "string") {
    throw new Error(`${label} must be a non-empty string`);
  }
  if (path.isAbsolute(artifactPath)) {
    throw new Error(`${label} must be repo-relative, got ${artifactPath}`);
  }
  if (artifactPath.includes("cartulary-check-schedule-")) {
    throw new Error(`${label} must not point at obsolete temp scheduler logs, got ${artifactPath}`);
  }
};
if (summary.schema_id !== "cartulary.check_scheduler_summary.v9") {
  throw new Error(`unexpected summary schema ${summary.schema_id}`);
}
if (summary.scheduler_kind !== "check") {
  throw new Error(`summary scheduler_kind got ${summary.scheduler_kind} want check`);
}
if (summary.status !== expectedStatus) {
  throw new Error(`summary status got ${summary.status} want ${expectedStatus}`);
}
if (expectedStatus === "fail" && summary.failure_class !== "harness") {
  throw new Error(`summary failure_class got ${summary.failure_class} want harness`);
}
if (expectedStatus === "pass" && summary.failure_class !== null) {
  throw new Error(`passing summary failure_class got ${summary.failure_class}`);
}
if (summary.total_work_units !== Number(expectedTotal)) {
  throw new Error(`total work units got ${summary.total_work_units} want ${expectedTotal}`);
}
if (!Number.isInteger(summary.scheduler_started_monotonic_ms) || summary.scheduler_started_monotonic_ms !== 0) {
  throw new Error(`scheduler_started_monotonic_ms got ${summary.scheduler_started_monotonic_ms} want 0`);
}
if (!Number.isInteger(summary.scheduler_completed_monotonic_ms) || summary.scheduler_completed_monotonic_ms < 0) {
  throw new Error(`scheduler_completed_monotonic_ms got ${summary.scheduler_completed_monotonic_ms}`);
}
if (!Number.isInteger(summary.scheduler_total_duration_ms) || summary.scheduler_total_duration_ms < summary.scheduler_completed_monotonic_ms - summary.scheduler_started_monotonic_ms) {
  throw new Error(`scheduler_total_duration_ms got ${summary.scheduler_total_duration_ms}`);
}
if (typeof summary.scheduler_started_at !== "string" || Number.isNaN(Date.parse(summary.scheduler_started_at))) {
  throw new Error("summary must record scheduler_started_at");
}
if (typeof summary.scheduler_completed_at !== "string" || Number.isNaN(Date.parse(summary.scheduler_completed_at))) {
  throw new Error("summary must record scheduler_completed_at");
}
const failed = summary.failed_work_unit ?? null;
if (expectedFailed === "-") {
  if (failed !== null) {
    throw new Error(`expected no failed work unit, got ${failed}`);
  }
} else if (failed !== expectedFailed) {
  throw new Error(`failed work unit got ${failed} want ${expectedFailed}`);
}
if (!Array.isArray(summary.slowest_work_units) || summary.slowest_work_units.length === 0) {
  throw new Error("summary must record slowest work");
}
if (!summary.resource_limits || Object.keys(summary.resource_limits).length === 0) {
  throw new Error("summary must record resource limits");
}
if (!Number.isInteger(summary.max_running_work_units) || summary.max_running_work_units < 1) {
  throw new Error(`summary max_running_work_units got ${summary.max_running_work_units}`);
}
if (!Number.isInteger(summary.max_running_groups) || summary.max_running_groups < 1) {
  throw new Error(`summary max_running_groups got ${summary.max_running_groups}`);
}
if (!summary.max_active_resource_claims || Object.keys(summary.max_active_resource_claims).length === 0) {
  throw new Error("summary must record max active resource claims");
}
if (!Array.isArray(summary.blocked_reasons_seen)) {
  throw new Error("summary must record blocked reasons");
}
if (!Array.isArray(summary.top_blockers) || summary.top_blockers.length > 5) {
  throw new Error("summary must record capped top blockers");
}
for (const [index, blocker] of summary.top_blockers.entries()) {
  if (!["dependency", "resource"].includes(blocker.kind)) {
    throw new Error(`top_blockers[${index}].kind got ${blocker.kind}`);
  }
  if (typeof blocker.name !== "string" || blocker.name.trim() === "") {
    throw new Error(`top_blockers[${index}] must record a blocker name`);
  }
  if (blocker.blocker !== `${blocker.kind}:${blocker.name}`) {
    throw new Error(`top_blockers[${index}] blocker key is not canonical`);
  }
  if (!Number.isInteger(blocker.count) || blocker.count < 1) {
    throw new Error(`top_blockers[${index}] must record a positive count`);
  }
}
if (!Array.isArray(summary.nested_scheduler_limits)) {
  throw new Error("summary must record nested scheduler limits as an array");
}
if (!Array.isArray(summary.nested_scheduler_observations)) {
  throw new Error("summary must record nested scheduler observations as an array");
}
if (summary.finalizer_count !== 0 || summary.finalizer_failures !== 0) {
  throw new Error("check scheduler summary must record zero finalizers");
}
if (!Array.isArray(summary.finalizer_timings) || summary.finalizer_timings.length !== 0) {
  throw new Error("check scheduler summary must record empty finalizer timings");
}
if (!summary.artifacts?.events_jsonl) {
  throw new Error("summary must record scheduler event artifact path");
}
if (!summary.artifacts?.scheduler_logs_dir) {
  throw new Error("summary must record scheduler log artifact path");
}
if (!summary.artifacts?.progress_summary_log) {
  throw new Error("summary must record progress summary artifact path");
}
assertRepoRelativeArtifact(summary.artifacts.events_jsonl, "events_jsonl");
assertRepoRelativeArtifact(summary.artifacts.scheduler_logs_dir, "scheduler_logs_dir");
assertRepoRelativeArtifact(summary.artifacts.progress_summary_log, "progress_summary_log");
if (pressure.schema_id !== "cartulary.scheduler_pressure_summary.v1") {
  throw new Error(`unexpected pressure schema ${pressure.schema_id}`);
}
for (const field of [
  "target",
  "scheduler_kind",
  "status",
  "target_counts",
  "lane_duration_ms",
  "resource_claim_counts",
  "fixture_class_counts",
  "slowest_work_units",
  "reused_accounting_counts",
  "readiness_attribution_counts",
  "generated_at",
]) {
  if (!Object.hasOwn(pressure, field)) {
    throw new Error(`pressure summary missing ${field}`);
  }
}
if (pressure.target !== summary.target || pressure.status !== summary.status) {
  throw new Error("pressure summary target/status must match scheduler summary");
}
if (typeof pressure.generated_at !== "string" || Number.isNaN(Date.parse(pressure.generated_at))) {
  throw new Error("pressure summary generated_at must be a timestamp");
}
if (
  pressure.reused_accounting_counts === null ||
  Array.isArray(pressure.reused_accounting_counts) ||
  typeof pressure.reused_accounting_counts !== "object" ||
  pressure.readiness_attribution_counts === null ||
  Array.isArray(pressure.readiness_attribution_counts) ||
  typeof pressure.readiness_attribution_counts !== "object"
) {
  throw new Error("pressure summary accounting fields must be objects");
}
const schedulerLogsDir = resolveArtifact(summary.artifacts.scheduler_logs_dir);
if (!fs.statSync(schedulerLogsDir).isDirectory()) {
  throw new Error(`scheduler log artifact path must be an existing directory: ${summary.artifacts.scheduler_logs_dir}`);
}
if (!fs.statSync(resolveArtifact(summary.artifacts.progress_summary_log)).isFile()) {
  throw new Error(`progress summary artifact path must be an existing file: ${summary.artifacts.progress_summary_log}`);
}
if (!progressLog.includes(`[CHECK-SCHEDULER] ${summary.target} start `)) {
  throw new Error("progress summary log must retain scheduler start");
}
if (!progressLog.includes(`[SUMMARY] target=${summary.target} `)) {
  throw new Error("progress summary log must retain scheduler summary");
}
if (!Array.isArray(summary.progress_snapshots) || summary.progress_snapshots.length > 8) {
  throw new Error("summary must record capped progress snapshots");
}
if (!Array.isArray(summary.slowest_running_observations) || summary.slowest_running_observations.length > 5) {
  throw new Error("summary must record capped slowest running observations");
}
if (events.length === 0) {
  throw new Error("scheduler events must not be empty");
}
if (events[0]?.event !== "scheduler-start") {
  throw new Error(`first scheduler event got ${events[0]?.event} want scheduler-start`);
}
if (events[events.length - 1]?.event !== "scheduler-finish") {
  throw new Error(`final scheduler event got ${events[events.length - 1]?.event} want scheduler-finish`);
}
if (!events.every((event) => event.schema_id === "cartulary.scheduler_event.v6")) {
  throw new Error("unexpected scheduler event schema");
}
events.forEach((event, index) => {
  if (event.seq !== index + 1) {
    throw new Error(`event ${index} sequence got ${event.seq} want ${index + 1}`);
  }
  if (!Number.isInteger(event.monotonic_ms) || event.monotonic_ms < 0) {
    throw new Error(`event ${index} missing monotonic_ms`);
  }
  if (index > 0 && event.monotonic_ms < events[index - 1].monotonic_ms) {
    throw new Error(`event ${index} monotonic_ms regressed`);
  }
  if (typeof event.emitted_at !== "string" || Number.isNaN(Date.parse(event.emitted_at))) {
    throw new Error(`event ${index} missing emitted_at`);
  }
  if (Object.hasOwn(event, "timestamp")) {
    throw new Error(`event ${index} must not emit legacy timestamp`);
  }
});
if (expectedEvent !== "-" && !events.some((event) => event.event === expectedEvent)) {
  throw new Error(`missing scheduler event ${expectedEvent}`);
}
const progressEvents = events.filter((event) => event.event === "progress");
if (progressEvents.length > 0) {
  if (summary.progress_snapshots.length === 0) {
    throw new Error("summary must retain progress snapshots when progress events were emitted");
  }
  if (!progressLog.includes(`[PROGRESS] target=${summary.target} `)) {
    throw new Error("progress summary log must retain human progress lines");
  }
  if (progressEvents.some((event) => event.slowest_running || event.nested_scheduler_progress?.some((progress) => progress.slowest_running)) && summary.slowest_running_observations.length === 0) {
    throw new Error("summary must retain slowest running observations");
  }
}
if (!events.some((event) => event.resource_limits && Object.keys(event.resource_limits).length > 0)) {
  throw new Error("events must include resource limits");
}
const recordedLogPaths = [
  ...summary.slowest_work_units.map((unit) => unit.log_file),
  ...events.map((event) => event.log_file).filter(Boolean),
];
if (recordedLogPaths.length === 0) {
  throw new Error("scheduler artifacts must record work-unit log paths");
}
for (const [index, logFile] of recordedLogPaths.entries()) {
  assertRepoRelativeArtifact(logFile, `log_file[${index}]`);
  if (!logFile.includes("/scheduler-logs/")) {
    throw new Error(`log_file[${index}] must live under scheduler-logs, got ${logFile}`);
  }
  const absoluteLogFile = resolveArtifact(logFile);
  const relativeToLogDir = path.relative(schedulerLogsDir, absoluteLogFile);
  if (relativeToLogDir.startsWith("..") || path.isAbsolute(relativeToLogDir)) {
    throw new Error(`log_file[${index}] must be inside scheduler_logs_dir, got ${logFile}`);
  }
  if (!fs.statSync(absoluteLogFile).isFile()) {
    throw new Error(`log_file[${index}] must be readable after scheduler exit, got ${logFile}`);
  }
}
EOF
}

json_field() {
  local file="$1"
  local path="$2"

  "$NODE_BIN" -e '
const fs = require("node:fs");
const [file, path] = process.argv.slice(1);
const value = path.split(".").reduce((current, key) => current?.[key], JSON.parse(fs.readFileSync(file, "utf8")));
if (value === undefined || value === null) {
  process.exit(1);
}
process.stdout.write(String(value));
' "$file" "$path"
}

assert_output_budget() {
  local manifest="$1"
  local target="$2"
  local stdout_file="$3"
  local stderr_file="$4"
  local label="$5"

  "$NODE_BIN" - "$manifest" "$target" "$stdout_file" "$stderr_file" "$label" <<'EOF'
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
for (const [key, actual] of [
  ["stdout_lines", lineCount(readText(stdoutFile))],
  ["stdout_bytes", Buffer.byteLength(readText(stdoutFile))],
  ["stderr_lines", lineCount(readText(stderrFile))],
  ["stderr_bytes", Buffer.byteLength(readText(stderrFile))],
]) {
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

  "$NODE_BIN" - "$stdout_file" "$stderr_file" "$expected_target" "$label" "$@" <<'EOF'
const fs = require("node:fs");
const [stdoutFile, stderrFile, expectedTarget, label, ...roleArgs] = process.argv.slice(2);
const logSeparator = roleArgs.indexOf("--log");
const expectedSummaryRoles = logSeparator === -1 ? roleArgs : roleArgs.slice(0, logSeparator);
const expectedLogRoles = logSeparator === -1 ? [] : roleArgs.slice(logSeparator + 1);
const stdout = fs.readFileSync(stdoutFile, "utf8");
const stderr = fs.readFileSync(stderrFile, "utf8");
if (stderr !== "") {
  throw new Error(`${label}: expected empty stderr in machine mode, got ${JSON.stringify(stderr)}`);
}
if (stdout.includes("[RESULT]") || stdout.includes("[ARTIFACTS]") || stdout.includes("[PROGRESS]")) {
  throw new Error(`${label}: machine mode must not include human summary or progress lines`);
}
const lines = stdout.split(/\r?\n/).filter((line) => line.trim() !== "");
if (lines.length !== 1) {
  throw new Error(`${label}: expected exactly one JSON line, got ${lines.length}`);
}
const summary = JSON.parse(lines[0]);
if (summary.schema_id !== "cartulary.tool_run_summary.v3") {
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
const summaryRoles = new Set((summary.summary_artifacts ?? []).map((artifact) => artifact.role));
for (const role of expectedSummaryRoles) {
  if (!summaryRoles.has(role)) {
    throw new Error(`${label}: missing summary artifact role ${role}`);
  }
}
const logRoles = new Set((summary.log_artifacts ?? []).map((artifact) => artifact.role));
for (const role of expectedLogRoles) {
  if (!logRoles.has(role)) {
    throw new Error(`${label}: missing log artifact role ${role}`);
  }
}
EOF
}

assert_fake_make_overlap() {
  local event_log="$1"
  local left_target="$2"
  local right_target="$3"
  local message="$4"

  "$NODE_BIN" - "$event_log" "$left_target" "$right_target" "$message" <<'EOF'
const fs = require("node:fs");
const [eventLog, leftTarget, rightTarget, message] = process.argv.slice(2);
const lines = fs.readFileSync(eventLog, "utf8").trim().split(/\n/).filter(Boolean);
const expected = new Set([leftTarget, rightTarget]);
const started = new Set();
const ended = new Set();
const running = new Set();
let overlapped = false;

for (const line of lines) {
  const match = /^(start|end) (\S+) active=(\d+)$/.exec(line);
  if (!match) {
    continue;
  }
  const [, action, target, activeText] = match;
  if (!expected.has(target)) {
    continue;
  }
  if (action === "start") {
    started.add(target);
    running.add(target);
  } else {
    ended.add(target);
    running.delete(target);
  }
  const active = Number(activeText);
  if (running.has(leftTarget) && running.has(rightTarget) && active >= 2) {
    overlapped = true;
  }
}

for (const target of expected) {
  if (!started.has(target)) {
    throw new Error(`${message}: missing start event for ${target}`);
  }
  if (!ended.has(target)) {
    throw new Error(`${message}: missing end event for ${target}`);
  }
}
if (!overlapped) {
  throw new Error(`${message}: ${leftTarget} and ${rightTarget} never overlapped with active>=2`);
}
EOF
}

write_fake_make() {
  local dir="$1"

  cat >"${dir}/fake-make" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

echo "$*" >>"${FAKE_CHECK_SCHEDULER_ARGS_LOG:?}"
target="${@: -1}"
lock_file="${FAKE_CHECK_SCHEDULER_LOCK:?}"
active_file="${FAKE_CHECK_SCHEDULER_ACTIVE:?}"
max_file="${FAKE_CHECK_SCHEDULER_MAX:?}"
event_log="${FAKE_CHECK_SCHEDULER_EVENT_LOG:?}"

mkdir -p "$(dirname "$active_file")"
touch "$active_file" "$max_file" "$event_log"

if [[ -n "${CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT:-}" || -n "${CARTULARY_SERVICE_BACKED_GO_IO_LIMIT:-}" ]]; then
  printf 'env %s go_cpu=%s go_io=%s\n' "$target" "${CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT:-}" "${CARTULARY_SERVICE_BACKED_GO_IO_LIMIT:-}" >>"$event_log"
fi
printf 'envflags %s MAKEFLAGS=%s MFLAGS=%s\n' "$target" "${MAKEFLAGS-}" "${MFLAGS-}" >>"$event_log"
printf 'test-target %s %s\n' "$target" "${CARTULARY_TEST_TARGET:-}" >>"$event_log"
printf 'strict-env %s %s\n' "$target" "${LINT_SHELL_STRICT:-unset}" >>"$event_log"

change_active() {
  local delta="$1"
  local active max action
  {
    flock 9
    active="$(cat "$active_file" 2>/dev/null || true)"
    max="$(cat "$max_file" 2>/dev/null || true)"
    active="${active:-0}"
    max="${max:-0}"
    active=$((active + delta))
    printf '%s\n' "$active" >"$active_file"
    if (( active > max )); then
      printf '%s\n' "$active" >"$max_file"
    fi
    if [[ "$delta" -gt 0 ]]; then
      action=start
    else
      action=end
    fi
    printf '%s %s active=%s\n' "$action" "$target" "$active" >>"$event_log"
  } 9>"$lock_file"
}

write_summary() {
  if [[ -z "${CARTULARY_TEST_RESULTS_DIR:-}" || -z "${CARTULARY_TEST_RUN_ID:-}" ]]; then
    return 0
  fi
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
}

write_failure_summary() {
  if [[ -z "${CARTULARY_TEST_RESULTS_DIR:-}" || -z "${CARTULARY_TEST_RUN_ID:-}" ]]; then
    return 0
  fi
  mkdir -p "${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}"
  cat >"${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}/target-summary.json" <<JSON
{
  "target": "${target}",
  "status": "fail",
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
    "failed": 1,
    "authoritative": 1,
    "support": 0,
    "unmapped": 0,
    "non_test": 0,
    "authoritative_failed": 1,
    "support_failed": 0,
    "unmapped_failed": 0,
    "non_test_failed": 0,
    "packages": 1
  },
  "failure_class": "product",
  "failure_reason": "test_assertion_failure",
  "failure_classes": { "product": 1, "config": 0, "infra": 0, "harness": 1, "artifact": 0, "timing": 0, "interrupted": 0, "unknown": 0 },
  "failure_reasons": { "usage_error": 0, "configuration_error": 0, "preflight_error": 0, "service_start_error": 0, "service_readiness_timeout": 0, "fixture_error": 0, "resource_conflict": 0, "test_assertion_failure": 1, "child_target_failure": 0, "scheduler_accounting_error": 0, "artifact_error": 0, "cleanup_error": 0, "duration_baseline_drift": 0, "timeout_failure": 0, "cancelled_or_interrupted": 0, "unknown_failure": 1 },
  "failures": [
    { "failure_class": "harness", "failure_reason": "unknown_failure", "kind": "failure", "source": "shell", "target": "${target}", "runner": "shell", "label": "(shell command)", "message": "command exited with status 1", "artifact": ".cartulary/test-results/${CARTULARY_TEST_RUN_ID}/${target}/${target}/stderr.log" },
    { "failure_class": "product", "failure_reason": "test_assertion_failure", "kind": "failure", "source": "vitest", "target": "${target}", "runner": "vitest", "label": "synthetic runner assertion", "message": "synthetic product failure", "artifact": ".cartulary/test-results/${CARTULARY_TEST_RUN_ID}/${target}/raw/runner.json" }
  ],
  "failure_headline": "vitest synthetic runner assertion: synthetic product failure"
}
JSON
}

write_phase_summary() {
  if [[ -z "${CARTULARY_TEST_RESULTS_DIR:-}" || -z "${CARTULARY_TEST_RUN_ID:-}" ]]; then
    return 0
  fi
  local artifact_target="${CARTULARY_TEST_TARGET:-adhoc}"
  local phase_dir="${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${artifact_target}/${target}"
  mkdir -p "$phase_dir"
  printf '{"ok":true}\n' >"${phase_dir}/runner.json"
  printf 'phase stdout for %s\n' "$target" >"${phase_dir}/stdout.log"
  printf 'phase stderr for %s\n' "$target" >"${phase_dir}/stderr.log"
  cat >"${phase_dir}/phase-summary.json" <<JSON
{
  "schema_id": "cartulary.test_phase_summary.v3",
  "label": "${target}",
  "target": "${artifact_target}",
  "runner": "shell",
  "status": "pass",
  "start_time": "2026-01-01T00:00:00Z",
  "end_time": "2026-01-01T00:00:01Z",
  "wall_duration_ms": 1,
  "critical_path_wall_duration_ms": 1,
  "counts": {
    "phases": 1,
    "tests": 0,
    "failed": 0,
    "authoritative": 0,
    "support": 0,
    "raw": 0,
    "tooling_support": 0,
    "unowned_regression": 0,
    "unmapped": 0,
    "non_test": 0,
    "non_test_failed": 0,
    "packages": 0
  },
  "artifacts": {
    "runner_json": "${phase_dir}/runner.json",
    "stdout_log": "${phase_dir}/stdout.log",
    "stderr_log": "${phase_dir}/stderr.log"
  }
}
JSON
}

write_nested_scheduler_progress() {
  if [[ "$target" != "service" && "$target" != "partial-service" ]]; then
    return 0
  fi
  if [[ -z "${CARTULARY_TEST_RESULTS_DIR:-}" || -z "${CARTULARY_TEST_RUN_ID:-}" ]]; then
    return 0
  fi
  local nested_dir="${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}"
  mkdir -p "$nested_dir"
  if [[ "$target" == "partial-service" ]]; then
    printf '{"schema_id":"cartulary.scheduler_event.v6","target":"partial-service","event":"progress","seq":1,"monotonic_ms":1,"emitted_at":"2026-01-01T00:00:00.001Z"' >"${nested_dir}/scheduler-events.jsonl"
    return 0
  fi
  {
    printf 'not-json-diagnostic\n'
    printf '%s\n' '{"schema_id":"cartulary.scheduler_event.v6","target":"service","event":"progress","seq":1,"monotonic_ms":1,"emitted_at":"2026-01-01T00:00:00.001Z","pending":2,"running":1,"total_work_units":6,"blocked":2,"completed":3,"pending_finalizers":0,"running_finalizers":0,"blocked_reason":"resources","blocked_resources":["go_io"],"waiting_on":["backend-store"],"blocked_units":[],"active_resource_claims":{"go_cpu":1},"resource_limits":{"go_cpu":1,"go_io":1},"active_groups":{"backend-integration":1},"blocked_by":["go_io"],"unblocks_after":"backend-integration/shard-a","slowest_running":{"label":"backend-integration/shard-a","duration_ms":1234}}'
  } >"${nested_dir}/scheduler-events.jsonl"
}

sleep_key="${target//-/_}"
sleep_key="${sleep_key^^}"
sleep_var="FAKE_SLEEP_${sleep_key}"
sleep_duration="${!sleep_var:-${FAKE_SLEEP_DEFAULT:-0.05}}"

change_active 1
write_nested_scheduler_progress
sleep "$sleep_duration"
if [[ "${FAKE_FAIL_TARGET:-}" == "$target" ]]; then
  echo "fake failure for $target" >&2
  if [[ "${FAKE_FAIL_WITH_SUMMARY_TARGET:-}" == "$target" ]]; then
    write_failure_summary
  fi
  change_active -1
  exit 7
fi
change_active -1

echo "fake pass for $target"
write_phase_summary
write_summary
EOF
  chmod +x "${dir}/fake-make"
}

run_scheduler() {
  local dir="$1"
  local manifest="$2"
  local run_id="$3"
  shift 3
  local normalized_manifest="${manifest}.normalized"

  "$NODE_BIN" - "$manifest" "$normalized_manifest" <<'EOF'
const fs = require("node:fs");
const [manifestFile, normalizedFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));

if (manifest.schema_id !== "cartulary.scheduler_manifest.v1") {
  fs.copyFileSync(manifestFile, normalizedFile);
  process.exit(0);
}

const serviceTarget = (unit) => unit.service_session?.target ?? unit.serviceSession?.target ?? unit.target;
const defaultCommand = (unit) => {
  const kind = unit.kind ?? "make_target";
  if (kind === "service_session") {
    return { type: "service_session_start", service_target: serviceTarget(unit) };
  }
  if (kind === "service_complete") {
    return { type: "service_complete", service_target: serviceTarget(unit) };
  }
  if (kind === "browser_stage_session") {
    return { type: "browser_stage_session_start", service_target: serviceTarget(unit), browser_stage: unit.browser_stage };
  }
  if (kind === "browser_group") {
    return { type: "browser_group", service_target: serviceTarget(unit), browser_stage: unit.browser_stage, group_id: unit.browser_group?.id ?? unit.browser_group ?? unit.target };
  }
  if (kind === "browser_stage_complete") {
    return { type: "browser_stage_complete", service_target: serviceTarget(unit), browser_stage: unit.browser_stage };
  }
  if (kind === "go_shard") {
    return { type: "go_shard", target: unit.target, shard: unit.shard, service_target: serviceTarget(unit) };
  }
  if (kind === "go_shard_finalize" || kind === "aggregate_finalize") {
    return { type: "go_shard_finalize", target: unit.target, service_target: serviceTarget(unit) };
  }
  return { type: "make_target", target: unit.target };
};

const schedules = (manifest.schedules ?? []).map((schedule) => ({
  ...schedule,
  scheduler_kind: schedule.scheduler_kind ?? "check",
  stop_on_first_failure: schedule.stop_on_first_failure ?? true,
  progress_tick_seconds: schedule.progress_tick_seconds ?? 30,
  validate_timing: schedule.validate_timing ?? true,
  work_units: (schedule.work_units ?? []).map((unit) => ({
    ...unit,
    command: unit.command ?? defaultCommand(unit),
  })),
  finalizers: schedule.finalizers ?? [],
}));

fs.writeFileSync(
  normalizedFile,
  `${JSON.stringify({
    schema_id: "cartulary.scheduler_manifest.v1",
    generated: manifest.generated ?? {
      generator: "scripts/test-check-scheduler.sh",
      source: "fixture",
    },
    schedules,
  }, null, 2)}\n`,
);
EOF

  env \
  FAKE_CHECK_SCHEDULER_LOCK="${dir}/lock" \
  FAKE_CHECK_SCHEDULER_ACTIVE="${dir}/active" \
  FAKE_CHECK_SCHEDULER_MAX="${dir}/max" \
  FAKE_CHECK_SCHEDULER_ARGS_LOG="${dir}/make-args.log" \
  FAKE_CHECK_SCHEDULER_EVENT_LOG="${dir}/events.log" \
  FAKE_SLEEP_DEFAULT="${FAKE_SLEEP_DEFAULT:-0.05}" \
  FAKE_SLEEP_ALPHA="${FAKE_SLEEP_ALPHA:-}" \
  FAKE_SLEEP_BETA="${FAKE_SLEEP_BETA:-}" \
  FAKE_SLEEP_LOCAL="${FAKE_SLEEP_LOCAL:-}" \
  FAKE_SLEEP_META="${FAKE_SLEEP_META:-}" \
  FAKE_SLEEP_SERVICE="${FAKE_SLEEP_SERVICE:-}" \
  FAKE_SLEEP_CHECK_SERVICE_BACKED="${FAKE_SLEEP_CHECK_SERVICE_BACKED:-}" \
  FAKE_SLEEP_MIGRATION_DRIFT="${FAKE_SLEEP_MIGRATION_DRIFT:-}" \
  FAKE_SLEEP_PARTIAL_SERVICE="${FAKE_SLEEP_PARTIAL_SERVICE:-}" \
  FAKE_FAIL_TARGET="${FAKE_FAIL_TARGET:-}" \
  FAKE_FAIL_WITH_SUMMARY_TARGET="${FAKE_FAIL_WITH_SUMMARY_TARGET:-}" \
  MAKE="${dir}/fake-make" \
  NODE_BIN="$NODE_BIN" \
  TEST_OUTPUT_SCRIPT="$TEST_OUTPUT_SCRIPT" \
  TEST_SERVICES_BIN="${TEST_SERVICES_BIN:-}" \
  SCHEDULER_MANIFEST="${SCHEDULER_MANIFEST:-${ROOT_DIR}/tools/scheduler_manifest.json}" \
  VERBOSE="${VERBOSE:-}" \
  CI_VERBOSE="${CI_VERBOSE:-}" \
  CARTULARY_OUTPUT_MODE="${CARTULARY_OUTPUT_MODE:-}" \
  CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS="${CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS:-}" \
  CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT="${CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT:-}" \
  CARTULARY_SERVICE_BACKED_GO_IO_LIMIT="${CARTULARY_SERVICE_BACKED_GO_IO_LIMIT:-}" \
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT="${CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT:-}" \
  CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT="${CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT:-}" \
  CARTULARY_TEST_RESULTS_DIR="${dir}/results" \
  CARTULARY_TEST_RUN_ID="$run_id" \
    "$NODE_BIN" "$SCRIPT" --target check --manifest "$normalized_manifest" "$@"
}

if [[ "$SUITE" == "smoke" ]]; then
smoke_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-smoke.XXXXXX")"
cleanup_paths+=("$smoke_dir")
write_fake_make "$smoke_dir"
smoke_manifest="${smoke_dir}/manifest.json"
cat >"$smoke_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "generated": {
    "generator": "scripts/test-check-scheduler.sh",
    "source": "smoke fixture"
  },
  "schedules": [
    {
      "target": "check",
      "scheduler_kind": "check",
      "capacity_profile": "check_default",
      "resource_limits": { "host_cpu": 12, "host_io": 12, "suite_service_stack": 1, "migration_scratch_postgres": 1 },
      "stop_on_first_failure": true,
      "progress_tick_seconds": 30,
      "validate_timing": true,
      "summary_groups": [
        { "name": "check-work", "summary_targets": ["local", "meta"] }
      ],
      "work_units": [
        { "target": "setup", "weight_ms": 30, "needs": [], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "command": { "type": "make_target", "target": "setup" } },
        { "target": "local", "weight_ms": 20, "needs": ["setup"], "produces_summary_targets": ["local"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "command": { "type": "make_target", "target": "local" } },
        { "target": "meta", "weight_ms": 10, "needs": ["setup"], "produces_summary_targets": ["meta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "command": { "type": "make_target", "target": "meta" } }
      ],
      "finalizers": []
    }
  ]
}
JSON
smoke_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_DEFAULT=0.01 run_scheduler "$smoke_dir" "$smoke_manifest" smoke --resource-limit host_cpu=2 --resource-limit host_io=2 2>&1)"
assert_contains "$smoke_output" "[CHECK-SCHEDULER] check start work_units=3 capacity={host_cpu:2,host_io:2,suite_service_stack:1,migration_scratch_postgres:1}" "smoke scheduler start"
assert_contains "$smoke_output" "[SUMMARY] target=check status=pass work_units=3/3" "smoke scheduler pass summary"
assert_not_contains "$smoke_output" "[STEP] check" "smoke output hides per-unit steps"
assert_check_scheduler_artifacts "$smoke_dir" smoke check pass - 3 finish

smoke_service_timing_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-smoke-service-timing.XXXXXX")"
cleanup_paths+=("$smoke_service_timing_dir")
write_fake_make "$smoke_service_timing_dir"
cat >"${smoke_service_timing_dir}/fake-test-services" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

mode="${1:-}"
shift || true

env_file=""
lease_file=""
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --env-file)
      env_file="${2:-}"
      shift 2
      ;;
    --lease-file | --lease)
      lease_file="${2:-}"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

case "$mode" in
  start-suite)
    mkdir -p "$(dirname "$env_file")" "$(dirname "$lease_file")"
    sleep 0.03
    printf '{"CARTULARY_TEST_SERVICES_ACTIVE":"1","CARTULARY_FAKE_SERVICE_READY":"1"}\n' >"$env_file"
    printf '{"schema_id":"cartulary.test_services.lease.v1","lease_id":"fake","suite_id":"fake","target":"service-timing-suite","cleanup_state":"ready","resources":[]}\n' >"$lease_file"
    ;;
  terminate-suite)
    sleep 0.02
    ;;
  record-lifecycle)
    ;;
  *)
    echo "unexpected fake test-services mode ${mode}" >&2
    exit 2
    ;;
esac
EOF
chmod +x "${smoke_service_timing_dir}/fake-test-services"
smoke_service_timing_manifest="${smoke_service_timing_dir}/manifest.json"
cat >"$smoke_service_timing_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "generated": {
    "generator": "scripts/test-check-scheduler.sh",
    "source": "smoke service timing fixture"
  },
  "schedules": [
    {
      "target": "check",
      "scheduler_kind": "check",
      "capacity_profile": "check_default",
      "resource_limits": { "host_cpu": 1, "host_io": 1, "suite_service_stack": 1 },
      "stop_on_first_failure": true,
      "progress_tick_seconds": 30,
      "validate_timing": true,
      "summary_groups": [
        { "name": "check-work", "summary_targets": ["service-timing-suite"] }
      ],
      "work_units": [
        {
          "id": "service-timing-suite:service-session",
          "kind": "service_session",
          "target": "service-timing-suite",
          "label": "service-timing-suite/service-session",
          "weight_ms": 10,
          "needs": [],
          "completion_keys": ["service_session:service-timing-suite"],
          "resource_claims": { "host_cpu": 1, "host_io": 1, "suite_service_stack": 1 },
          "retained_resource_claims": { "suite_service_stack": 1 },
          "service_session": { "target": "service-timing-suite" },
          "command": { "type": "service_session_start", "service_target": "service-timing-suite" }
        },
        {
          "id": "service-timing-suite:service-child",
          "kind": "service_make_target",
          "target": "service-timing-suite",
          "label": "service-timing-suite/child",
          "aggregate_target": "service-timing-suite",
          "weight_ms": 10,
          "needs": ["service_session:service-timing-suite"],
          "completion_keys": ["service-child"],
          "failure_keys": ["service-child"],
          "resource_claims": { "host_cpu": 1 },
          "service_session": { "target": "service-timing-suite" },
          "command": { "type": "make_target", "target": "service-timing-suite", "service_target": "service-timing-suite" }
        },
        {
          "id": "service-timing-suite:complete",
          "kind": "service_complete",
          "target": "service-timing-suite",
          "label": "service-timing-suite/complete",
          "weight_ms": 1,
          "needs": ["service-child"],
          "completion_keys": ["service-timing-suite"],
          "failure_keys": ["service-timing-suite"],
          "produces_summary_targets": ["service-timing-suite"],
          "count_in_total": false,
          "counts_started": false,
          "resource_claims": {},
          "service_session": { "target": "service-timing-suite" },
          "command": { "type": "service_complete", "service_target": "service-timing-suite" }
        }
      ],
      "finalizers": []
    }
  ]
}
JSON
smoke_service_timing_output="$(
  TEST_SERVICES_BIN="${smoke_service_timing_dir}/fake-test-services" \
    run_scheduler "$smoke_service_timing_dir" "$smoke_service_timing_manifest" smoke-service-timing 2>&1
)"
assert_contains "$smoke_service_timing_output" "[SUMMARY] target=check status=pass" "smoke service timing scheduler success"
smoke_service_timing_summary="${smoke_service_timing_dir}/results/smoke-service-timing/check/scheduler-summary.json"
assert_file_present "$smoke_service_timing_summary" "smoke service timing scheduler summary"
"$NODE_BIN" - "$smoke_service_timing_summary" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const session = summary.service_sessions?.find((entry) => entry.target === "service-timing-suite");
if (!session) {
  throw new Error("missing service timing session summary");
}
for (const field of ["setup_duration_ms", "ready_at_monotonic_ms", "child_work_started_at_monotonic_ms", "cleanup_duration_ms"]) {
  if (!Number.isFinite(session[field]) || session[field] < 0) {
    throw new Error(`${field} must be a non-negative number, got ${session[field]}`);
  }
}
if (session.cleanup_status !== "pass") {
  throw new Error(`cleanup_status got ${session.cleanup_status}`);
}
if (session.ready_at_monotonic_ms > session.child_work_started_at_monotonic_ms) {
  throw new Error("child work cannot start before service readiness");
}
EOF

smoke_dry_run_output="$(MAKEFLAGS=n run_scheduler "$smoke_dir" "$smoke_manifest" smoke-dry-run --resource-limit host_cpu=2 --resource-limit host_io=2 2>&1)"
assert_contains "$smoke_dry_run_output" "[DRY-RUN] check manifest=" "smoke dry-run output"
assert_contains "$smoke_dry_run_output" "resource_limits={host_cpu:2,host_io:2,suite_service_stack:1,migration_scratch_postgres:1} work_units=3 dependencies=2 classes={:3} types={make_target:3} top_weighted=setup:30,local:20,meta:10" "smoke dry-run compact summary"
assert_not_contains "$smoke_dry_run_output" "[DRY-RUN] check unit" "smoke dry-run hides unit expansion"
exit 0
fi

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import { mkdtempSync, writeFileSync } from "node:fs";
import os from "node:os";
import path from "node:path";
import {
  browserStageResource,
  checkHostCPUMaxAutoLimit,
  estimateCheckHostCPULimit,
  estimateCheckHostIOLimit,
  estimatePostgresCloneAutoLimit,
  estimatePostgresResetAutoLimit,
  isAutoLimitResource,
  loadSchedulerResourceRegistry,
  maxResourceClaims,
  normalizeResourceClaims,
  normalizeResourceLimits,
  resourceLimitsForCapacityProfile,
  resourceOverrideEnvVariablesForScheduler,
  preferredResourcesForScheduler,
  resolveAutoResourceLimits,
} from "./scripts/lib/scheduler-resources.mjs";

const fail = (message) => {
  throw new Error(message);
};

const validSchedulerRegistry = () => ({
  schema_id: "cartulary.scheduler_resource_registry.v4",
  resources: [
    {
      name: "host_cpu",
      display_name: "host CPU",
      schedulers: ["check"],
      display_order: 10,
      capacity: {
        default_limit: 1,
        max_limit: 256,
      },
    },
  ],
  templates: [
    {
      name: "browser_stage",
      prefix: "browser_stage_",
      display_name: "browser stage",
      schedulers: ["service_backed"],
      display_order: 100,
      max_limit: 8,
    },
  ],
  capacity_profiles: [
    {
      name: "check_default",
      scheduler: "check",
      resources: ["host_cpu"],
    },
  ],
  forwarding_profiles: [],
});

if (browserStageResource("webserver-backed") !== "browser_stage_webserver_backed") {
  fail("browser stage lane derivation changed");
}
if (preferredResourcesForScheduler("check").join(",") !== "host_cpu,host_io,suite_service_stack,migration_scratch_postgres,browser_stack,object_store,postgres,process,postgres_reset,postgres_clone") {
  fail("check resource display order changed");
}
if (resourceOverrideEnvVariablesForScheduler("check").join(",") !== "CHECK_HOST_CPU_JOBS,CHECK_HOST_IO_JOBS,CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT,CARTULARY_SERVICE_BACKED_POSTGRES_RESET_LIMIT,CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT") {
  fail("check override env names changed");
}
if (!isAutoLimitResource("go_cpu") || !isAutoLimitResource("go_io") || !isAutoLimitResource("browser_stack") || !isAutoLimitResource("postgres_reset") || !isAutoLimitResource("postgres_clone")) {
  fail("service-backed auto-limit resources are incomplete");
}
if (!isAutoLimitResource("host_cpu") || !isAutoLimitResource("host_io")) {
  fail("check host resources must be auto-limit resources");
}
const checkProfile = resourceLimitsForCapacityProfile("check_default", "registry test", {
  scheduler: "check",
  allowAuto: true,
});
if (
  checkProfile.limits.get("host_cpu") !== "auto" ||
  checkProfile.limits.get("host_io") !== "auto" ||
  checkProfile.limits.get("suite_service_stack") !== 1 ||
  checkProfile.limits.get("migration_scratch_postgres") !== 1
) {
  fail("check_default capacity profile changed");
}
const expectedCheckCPU = estimateCheckHostCPULimit();
const expectedCheckIO = estimateCheckHostIOLimit(new Map([["host_cpu", expectedCheckCPU]]));
if (
  expectedCheckCPU < 1 ||
  expectedCheckCPU > checkHostCPUMaxAutoLimit ||
  expectedCheckIO !== expectedCheckCPU
) {
  fail("check host auto-limit estimates changed");
}
if (checkProfile.sources.get("host_cpu") !== "registry:check_default") {
  fail(`check_default source got ${checkProfile.sources.get("host_cpu")}`);
}
const serviceProfile = resourceLimitsForCapacityProfile("service_backed_full", "registry test", {
  scheduler: "service_backed",
  allowAuto: true,
});
if (serviceProfile.limits.get("go_cpu") !== "auto" || serviceProfile.limits.get("go_io") !== "auto" || serviceProfile.limits.get("browser_stack") !== "auto" || serviceProfile.limits.get("postgres_reset") !== "auto" || serviceProfile.limits.get("postgres_clone") !== "auto") {
  fail("service_backed_full auto limits changed");
}
if (estimatePostgresCloneAutoLimit(new Map([["host_cpu", 12], ["host_io", 12]])) !== 6) {
  fail("postgres clone auto limit must resolve to 6 on the supported 12 CPU/IO profile");
}
if (estimatePostgresCloneAutoLimit(new Map([["go_cpu", 6], ["go_io", 8]]), { cpuResources: ["go_cpu"], ioResources: ["go_io"] }) !== 6) {
  fail("postgres clone auto limit must use the calibrated service-backed floor");
}
if (estimatePostgresResetAutoLimit(new Map([["host_io", 12]])) !== 4) {
  fail("postgres reset auto limit must resolve to 4 on the supported 12 IO profile");
}
if (estimatePostgresResetAutoLimit(new Map([["go_io", 8]]), { ioResources: ["go_io"] }) !== 2) {
  fail("postgres reset auto limit must scale with the service-backed IO lane budget");
}
const envResolved = normalizeResourceLimits(
  { host_cpu: "auto", host_io: "auto", suite_service_stack: 1, migration_scratch_postgres: 1 },
  "registry env test",
  {
    scheduler: "check",
    capacityProfile: "check_default",
    allowAuto: true,
    env: { CHECK_HOST_CPU_JOBS: "5" },
  },
);
if (envResolved.limits.get("host_cpu") !== 5 || envResolved.sources.get("host_cpu") !== "env:CHECK_HOST_CPU_JOBS") {
  fail("check host_cpu env override did not resolve from the registry");
}
const cliResolved = normalizeResourceLimits(
  { host_cpu: "auto", host_io: "auto", suite_service_stack: 1, migration_scratch_postgres: 1 },
  "registry cli test",
  {
    scheduler: "check",
    capacityProfile: "check_default",
    allowAuto: true,
    env: { CHECK_HOST_CPU_JOBS: "5" },
    overrides: new Map([["host_cpu", 4]]),
  },
);
if (cliResolved.limits.get("host_cpu") !== 4 || cliResolved.sources.get("host_cpu") !== "cli") {
  fail("check host_cpu CLI override must win over env override");
}
const declaredClaimUnits = [
  { resourceClaims: new Map([["host_cpu", 6], ["host_io", 3]]) },
];
const autoFloored = resolveAutoResourceLimits(
  new Map([["host_cpu", "auto"], ["host_io", "auto"]]),
  new Map([["host_cpu", "registry:test"], ["host_io", "registry:test"]]),
  "registry auto floor test",
  {
    check_host_cpu: () => 2,
    check_host_io: () => 1,
  },
  maxResourceClaims(declaredClaimUnits),
);
if (autoFloored.resourceLimits.get("host_cpu") !== 6 || autoFloored.resourceLimits.get("host_io") !== 3) {
  fail("auto resource limits must not resolve below the largest declared claim");
}
try {
  resolveAutoResourceLimits(
    new Map([["host_cpu", 5]]),
    new Map([["host_cpu", "cli"]]),
    "registry override floor test",
    {},
    new Map([["host_cpu", 6]]),
  );
  fail("resource override below largest declared claim was accepted");
} catch (error) {
  if (!String(error.message).includes("resource_limits.host_cpu must be >= largest declared claim 6")) {
    throw error;
  }
}
try {
  normalizeResourceLimits(
    { host_cpu: "auto", host_io: "auto", suite_service_stack: 1, migration_scratch_postgres: 1 },
    "registry env bound test",
    {
      scheduler: "check",
      capacityProfile: "check_default",
      allowAuto: true,
      env: { CHECK_HOST_CPU_JOBS: "257" },
    },
  );
  fail("oversized host_cpu env override was accepted");
} catch (error) {
  if (!String(error.message).includes("resource_limits.host_cpu must be <= 256")) {
    throw error;
  }
}
try {
  normalizeResourceLimits(
    { host_cpu: "auto", host_io: "auto", suite_service_stack: 1, migration_scratch_postgres: 1 },
    "registry cli bound test",
    {
      scheduler: "check",
      capacityProfile: "check_default",
      allowAuto: true,
      overrides: new Map([["host_cpu", 257]]),
    },
  );
  fail("oversized host_cpu CLI override was accepted");
} catch (error) {
  if (!String(error.message).includes("resource_limits.host_cpu must be <= 256")) {
    throw error;
  }
}
try {
  normalizeResourceLimits(
    { browser_stage_visual: 9 },
    "browser stage bound test",
    { scheduler: "check" },
  );
  fail("oversized browser stage limit was accepted");
} catch (error) {
  if (!String(error.message).includes("resource_limits.browser_stage_visual must be <= 8")) {
    throw error;
  }
}
const invalidRegistryPath = path.join(mkdtempSync(path.join(os.tmpdir(), "cartulary-registry-test-")), "registry.json");
const invalidCapacityRegistry = validSchedulerRegistry();
invalidCapacityRegistry.resources[0].capacity.auto_policy = "bad_policy";
writeFileSync(
  invalidRegistryPath,
  `${JSON.stringify(invalidCapacityRegistry)}\n`,
);
try {
  loadSchedulerResourceRegistry(invalidRegistryPath);
  fail("invalid capacity descriptor was accepted");
} catch (error) {
  if (!String(error.message).includes("exactly one of default_limit or auto_policy")) {
    throw error;
  }
}
const unknownRegistryKeyPath = path.join(mkdtempSync(path.join(os.tmpdir(), "cartulary-registry-key-test-")), "registry.json");
const unknownKeyRegistry = validSchedulerRegistry();
unknownKeyRegistry.retired_aliases = [];
writeFileSync(
  unknownRegistryKeyPath,
  `${JSON.stringify(unknownKeyRegistry)}\n`,
);
try {
  loadSchedulerResourceRegistry(unknownRegistryKeyPath);
  fail("unknown registry top-level key was accepted");
} catch (error) {
  if (!String(error.message).includes("unknown key retired_aliases")) {
    throw error;
  }
}
const { limits } = normalizeResourceLimits(
  { host_cpu: 4, host_io: 3, suite_service_stack: 1, migration_scratch_postgres: 1 },
  "registry test",
  { scheduler: "check" },
);
const claims = normalizeResourceClaims(
  { host_cpu: { mode: "bounded_limit", reserve: 1, min: 1, max: 3 }, host_io: 1, suite_service_stack: 1 },
  "registry test",
  limits,
  { scheduler: "check", allowBounded: true },
);
if (claims.get("host_cpu") !== 3) {
  fail(`bounded host_cpu claim got ${claims.get("host_cpu")}`);
}
EOF

event_order_dir="$(mktemp -d "${ROOT_DIR}/tmp/scheduler-event-order.XXXXXX")"
cleanup_paths+=("$event_order_dir")
mkdir -p "${event_order_dir}/valid/check" "${event_order_dir}/sequence/check" "${event_order_dir}/monotonic/check" "${event_order_dir}/wall/check" "${event_order_dir}/skew/check"
cat >"${event_order_dir}/valid/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"start","seq":1,"monotonic_ms":0,"emitted_at":"2026-01-01T00:00:00.000Z"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"finish","seq":2,"monotonic_ms":5,"emitted_at":"2026-01-01T00:00:00.005Z"}
JSONL
cat >"${event_order_dir}/sequence/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"start","seq":1,"monotonic_ms":0,"emitted_at":"2026-01-01T00:00:00.000Z"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"finish","seq":3,"monotonic_ms":5,"emitted_at":"2026-01-01T00:00:00.005Z"}
JSONL
cat >"${event_order_dir}/monotonic/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"start","seq":1,"monotonic_ms":10,"emitted_at":"2026-01-01T00:00:00.010Z"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"finish","seq":2,"monotonic_ms":5,"emitted_at":"2026-01-01T00:00:00.005Z"}
JSONL
cat >"${event_order_dir}/wall/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"start","seq":1,"monotonic_ms":0,"emitted_at":"2026-01-01T00:00:02.000Z"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"finish","seq":2,"monotonic_ms":5,"emitted_at":"2026-01-01T00:00:01.000Z"}
JSONL
cat >"${event_order_dir}/skew/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"start","seq":1,"monotonic_ms":0,"emitted_at":"2026-01-01T00:00:02.000Z"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"clock-skew","seq":2,"monotonic_ms":1,"emitted_at":"2026-01-01T00:00:03.000Z"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"finish","seq":3,"monotonic_ms":5,"emitted_at":"2026-01-01T00:00:01.000Z"}
JSONL
assert_contains "$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-event-order-drift.mjs" "${event_order_dir}/valid" 2>&1)" "scheduler event order verified" "valid scheduler event order drift fixture"
set +e
sequence_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-event-order-drift.mjs" "${event_order_dir}/sequence" 2>&1)"
sequence_status=$?
monotonic_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-event-order-drift.mjs" "${event_order_dir}/monotonic" 2>&1)"
monotonic_status=$?
wall_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-event-order-drift.mjs" "${event_order_dir}/wall" 2>&1)"
wall_status=$?
set -e
assert_equals "$sequence_status" "1" "event sequence drift fixture status"
assert_contains "$sequence_output" "seq got 3, want 2" "event sequence drift fixture output"
assert_equals "$monotonic_status" "1" "monotonic drift fixture status"
assert_contains "$monotonic_output" "monotonic_ms regressed" "monotonic drift fixture output"
assert_equals "$wall_status" "1" "wall drift fixture status"
assert_contains "$wall_output" "emitted_at regressed without preceding clock-skew marker" "wall drift fixture output"
assert_contains "$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-event-order-drift.mjs" "${event_order_dir}/skew" 2>&1)" "scheduler event order verified" "clock skew marker drift fixture"

summary_timing_dir="$(mktemp -d "${ROOT_DIR}/tmp/scheduler-summary-timing.XXXXXX")"
cleanup_paths+=("$summary_timing_dir")
mkdir -p "${summary_timing_dir}/valid/check" "${summary_timing_dir}/stale/check"
cat >"${summary_timing_dir}/valid/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"scheduler-start","seq":1,"monotonic_ms":0,"emitted_at":"2026-01-01T00:00:00.000Z"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"scheduler-finish","seq":2,"monotonic_ms":120000,"emitted_at":"2026-01-01T00:02:00.000Z"}
JSONL
cat >"${summary_timing_dir}/valid/check/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.check_scheduler_summary.v9",
  "target": "check",
  "status": "pass",
  "scheduler_kind": "check",
  "scheduler_started_monotonic_ms": 0,
  "scheduler_completed_monotonic_ms": 120000,
  "scheduler_total_duration_ms": 120000,
  "scheduler_started_at": "2026-01-01T00:00:00.000Z",
  "scheduler_completed_at": "2026-01-01T00:02:00.000Z",
  "critical_path_wall_duration_ms": 120000,
  "critical_path_units": [],
  "critical_path_blockers": [],
  "critical_path_terminal_unit": null
}
JSON
cat >"${summary_timing_dir}/valid/check/target-summary.json" <<'JSON'
{
  "schema_id": "cartulary.target_summary.v5",
  "target": "check",
  "status": "pass",
  "start_time": "2026-01-01T00:00:00.000Z",
  "end_time": "2026-01-01T00:02:00.000Z",
  "wall_duration_ms": 120000,
  "critical_path_wall_duration_ms": 120000
}
JSON
cat >"${summary_timing_dir}/valid/check/tool-run-summary.json" <<'JSON'
{
  "schema_id": "cartulary.tool_run_summary.v3",
  "target": "check",
  "status": "pass",
  "completed_at": "2026-01-01T00:02:00.000Z",
  "duration_ms": 120000,
  "scheduler_timing": {
    "scheduler_total_duration_ms": 120000
  },
  "extensions": {}
}
JSON
cat >"${summary_timing_dir}/valid/run-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_run_summary.v6",
  "label": "check",
  "status": "pass",
  "end_time": "2026-01-01T00:02:00.000Z",
  "wall_duration_ms": 120000,
  "critical_path_wall_duration_ms": 120000
}
JSON
cat >"${summary_timing_dir}/valid/tool-run-summary.json" <<'JSON'
{
  "schema_id": "cartulary.tool_run_summary.v3",
  "target": "check",
  "status": "pass",
  "completed_at": "2026-01-01T00:02:00.000Z",
  "duration_ms": 120000,
  "scheduler_timing": {
    "scheduler_total_duration_ms": 120000
  },
  "extensions": {}
}
JSON
cp "${summary_timing_dir}/valid/check/scheduler-events.jsonl" "${summary_timing_dir}/stale/check/scheduler-events.jsonl"
cp "${summary_timing_dir}/valid/check/scheduler-summary.json" "${summary_timing_dir}/stale/check/scheduler-summary.json"
cp "${summary_timing_dir}/valid/check/target-summary.json" "${summary_timing_dir}/stale/check/target-summary.json"
cp "${summary_timing_dir}/valid/run-summary.json" "${summary_timing_dir}/stale/run-summary.json"
cat >"${summary_timing_dir}/stale/check/tool-run-summary.json" <<'JSON'
{
  "schema_id": "cartulary.tool_run_summary.v3",
  "target": "check",
  "status": "pass",
  "completed_at": "2026-01-01T00:01:10.000Z",
  "duration_ms": 70000,
  "scheduler_timing": {
    "scheduler_total_duration_ms": 120000
  },
  "extensions": {}
}
JSON
cp "${summary_timing_dir}/valid/tool-run-summary.json" "${summary_timing_dir}/stale/tool-run-summary.json"
assert_contains "$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" "${summary_timing_dir}/valid" 2>&1)" "scheduler summary timing verified" "valid scheduler summary timing fixture"
"$NODE_BIN" --input-type=module - "${summary_timing_dir}/warm" <<'EOF'
import { mkdirSync, writeFileSync } from "node:fs";
import path from "node:path";

const [root] = process.argv.slice(2);
const schemaID = "cartulary.scheduler_event.v6";
const startWallMs = Date.parse("2026-01-01T00:00:00.000Z");
const emittedAt = (ms) => new Date(startWallMs + ms).toISOString();

function writeJSON(file, value) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function writeEvents(file, events) {
  mkdirSync(path.dirname(file), { recursive: true });
  writeFileSync(file, `${events.map((event) => JSON.stringify(event)).join("\n")}\n`);
}

function baseSummary(target, durationMs) {
  return {
    schema_id: "cartulary.target_summary.v5",
    target,
    status: "pass",
    start_time: emittedAt(0),
    end_time: emittedAt(durationMs),
    wall_duration_ms: durationMs,
    critical_path_wall_duration_ms: durationMs,
  };
}

function toolSummary(target, durationMs) {
  return {
    schema_id: "cartulary.tool_run_summary.v3",
    target,
    status: "pass",
    completed_at: emittedAt(durationMs),
    duration_ms: durationMs,
    scheduler_timing: {
      scheduler_total_duration_ms: durationMs,
    },
    extensions: {},
  };
}

  function writeWarmRun(
    name,
    {
      serviceMs = 58000,
      measurement = false,
      browserSkew = false,
      coldProvisioning = false,
      unexpectedReuse = false,
      fixtureOverBudget = false,
      forbiddenVisual = false,
    } = {},
  ) {
  const runDir = path.join(root, name);
  const checkDir = path.join(runDir, "check");
  const serviceDir = path.join(runDir, "check-service-backed");
  const events = [];
  let seq = 1;
  const push = (event) => events.push({ schema_id: schemaID, target: "check", seq: seq++, ...event });
  const start = (id, monotonicMs, extra = {}) =>
    push({
      event: "start",
      monotonic_ms: monotonicMs,
      emitted_at: emittedAt(monotonicMs),
      work_unit_id: id,
      work_unit: id,
      ...extra,
    });
  const finish = (id, monotonicMs, durationMs = monotonicMs, extra = {}) =>
    push({
      event: "finish",
      monotonic_ms: monotonicMs,
      emitted_at: emittedAt(monotonicMs),
      work_unit_id: id,
      work_unit: id,
      status: 0,
      duration_ms: durationMs,
      ...extra,
    });

  push({ event: "scheduler-start", monotonic_ms: 0, emitted_at: emittedAt(0) });
  if (coldProvisioning) {
    start("testservices-build", 0, {
      work_unit_type: "make_target",
      aggregate_target: "testservices-build",
    });
    finish("testservices-build", 30000, 30000);
  }
  if (unexpectedReuse) {
    start("lint-shell", 0, {
      work_unit_type: "make_target",
      aggregate_target: "lint-shell",
    });
    finish("lint-shell", 20, 20, {
      extensions: {
        "cartulary.scheduler_accounting": {
          accounting_mode: "reused",
          cache_outcome: "hit",
        },
      },
    });
  }
  start("check-service-backed", 0, {
    work_unit_type: "make_target",
    aggregate_target: "check-service-backed",
    nested_scheduler: { type: "service_backed", target: "check-service-backed" },
  });

  const goDurations = [10000, 11000, 12000];
  for (const [index, durationMs] of goDurations.entries()) {
    start(`check-service-backed:backend-integration:backend-integration-shard-${index + 1}`, 0, {
      work_unit_type: "go_shard",
      aggregate_target: "backend-integration",
    });
    finish(`check-service-backed:backend-integration:backend-integration-shard-${index + 1}`, durationMs, durationMs);
  }

  if (measurement) {
    start("check-service-backed:browser-stage-session:measurement", 0, {
      work_unit_type: "browser_stage_session",
      aggregate_target: "browser-e2e-measurement",
    });
    finish("check-service-backed:browser-stage-session:measurement", 13000, 13000);
  }

  const browserDurations = browserSkew ? [20000, 21000, 40000] : [20000, 21000, 22000];
  for (const [index, durationMs] of browserDurations.entries()) {
    start(`check-service-backed:browser-e2e-webserver-backed:browser-functional-shard-${index + 1}`, 0, {
      work_unit_type: "browser_group",
      aggregate_target: "browser-e2e-webserver-backed",
    });
    finish(
      `check-service-backed:browser-e2e-webserver-backed:browser-functional-shard-${index + 1}`,
      durationMs,
      durationMs,
    );
  }
  if (forbiddenVisual) {
    start("check-service-backed:browser-e2e-visual:visual-smoke", 0, {
      work_unit_type: "browser_group",
      aggregate_target: "browser-e2e-visual",
    });
    finish(
      "check-service-backed:browser-e2e-visual:visual-smoke",
      36000,
      36000,
    );
  }

  finish("check-service-backed", serviceMs, serviceMs);
  push({ event: "scheduler-finish", monotonic_ms: serviceMs, emitted_at: emittedAt(serviceMs) });
  writeEvents(path.join(checkDir, "scheduler-events.jsonl"), events);
  writeJSON(path.join(checkDir, "scheduler-summary.json"), {
    schema_id: "cartulary.check_scheduler_summary.v9",
    target: "check",
    status: "pass",
    scheduler_kind: "check",
    scheduler_started_monotonic_ms: 0,
    scheduler_completed_monotonic_ms: serviceMs,
    scheduler_total_duration_ms: serviceMs,
    scheduler_started_at: emittedAt(0),
    scheduler_completed_at: emittedAt(serviceMs),
    critical_path_wall_duration_ms: serviceMs,
    critical_path_units: [],
    critical_path_blockers: [],
    critical_path_terminal_unit: null,
  });
  writeJSON(path.join(checkDir, "target-summary.json"), baseSummary("check", serviceMs));
  writeJSON(path.join(checkDir, "tool-run-summary.json"), toolSummary("check", serviceMs));
  const runSummary = {
    schema_id: "cartulary.test_run_summary.v6",
    label: "check",
    status: "pass",
    end_time: emittedAt(serviceMs),
    wall_duration_ms: serviceMs,
    critical_path_wall_duration_ms: serviceMs,
  };
  if (fixtureOverBudget) {
    runSummary.fixture = {
      by_strategy: [
        {
          service: "postgres",
          operation: "database-reset",
          fixture_policy: "package_reset",
          count: 31,
          total_duration_ms: 61000,
        },
      ],
    };
  }
  writeJSON(path.join(runDir, "run-summary.json"), runSummary);
  writeJSON(path.join(runDir, "tool-run-summary.json"), toolSummary("check", serviceMs));
  writeJSON(path.join(serviceDir, "target-summary.json"), baseSummary("check-service-backed", serviceMs));
}

writeWarmRun("valid");
writeWarmRun("overbudget", { serviceMs: 70000 });
writeWarmRun("measurement", { measurement: true });
writeWarmRun("skewed", { browserSkew: true });
writeWarmRun("cold-provisioning", { coldProvisioning: true });
writeWarmRun("unexpected-reuse", { unexpectedReuse: true });
writeWarmRun("fixture-overbudget", { fixtureOverBudget: true });
writeWarmRun("visual-in-default", { forbiddenVisual: true });
EOF
assert_contains "$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" --target check --warm-check-budget-ms 60000 --warm-check-balance-ratio 1.25 "${summary_timing_dir}/warm/valid" 2>&1)" "warm check scheduler health verified" "valid warm check health fixture"
set +e
warm_overbudget_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" --target check --warm-check-budget-ms 60000 --warm-check-balance-ratio 1.25 "${summary_timing_dir}/warm/overbudget" 2>&1)"
warm_overbudget_status=$?
set -e
assert_equals "$warm_overbudget_status" "1" "warm check overbudget fixture status"
assert_contains "$warm_overbudget_output" "warm duration 70000ms exceeds budget 60000ms" "warm check overbudget fixture output"
set +e
warm_measurement_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" --target check --warm-check-budget-ms 60000 --warm-check-balance-ratio 1.25 "${summary_timing_dir}/warm/measurement" 2>&1)"
warm_measurement_status=$?
set -e
assert_equals "$warm_measurement_status" "1" "warm check measurement fixture status"
assert_contains "$warm_measurement_output" "default warm check includes explicit browser evidence unit check-service-backed:browser-stage-session:measurement" "warm check measurement fixture output"
set +e
warm_skew_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" --target check --warm-check-budget-ms 60000 --warm-check-balance-ratio 1.25 "${summary_timing_dir}/warm/skewed" 2>&1)"
warm_skew_status=$?
set -e
assert_equals "$warm_skew_status" "1" "warm check skew fixture status"
assert_contains "$warm_skew_output" "browser-e2e-webserver-backed functional" "warm check skew fixture output"
set +e
warm_cold_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" --target check --warm-check-budget-ms 60000 --warm-check-balance-ratio 1.25 "${summary_timing_dir}/warm/cold-provisioning" 2>&1)"
warm_cold_status=$?
set -e
assert_equals "$warm_cold_status" "1" "warm check cold provisioning fixture status"
assert_contains "$warm_cold_output" "warm readiness unit testservices-build duration 30000ms exceeds warm threshold 15000ms" "warm check cold provisioning fixture output"
set +e
warm_reuse_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" --target check --warm-check-budget-ms 60000 --warm-check-balance-ratio 1.25 "${summary_timing_dir}/warm/unexpected-reuse" 2>&1)"
warm_reuse_status=$?
set -e
assert_equals "$warm_reuse_status" "1" "warm check unexpected reuse fixture status"
assert_contains "$warm_reuse_output" "unexpected reused work lint-shell is not allowed in the current check profile" "warm check unexpected reuse fixture output"
set +e
warm_fixture_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" --target check --warm-check-budget-ms 60000 --warm-check-balance-ratio 1.25 "${summary_timing_dir}/warm/fixture-overbudget" 2>&1)"
warm_fixture_status=$?
set -e
assert_equals "$warm_fixture_status" "1" "warm check fixture budget fixture status"
assert_contains "$warm_fixture_output" "package-reset fixture count 31 exceeds warm budget 30" "warm check fixture budget count output"
assert_contains "$warm_fixture_output" "package-reset fixture duration 61000ms exceeds warm budget 60000ms" "warm check fixture budget duration output"
set +e
warm_visual_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" --target check --warm-check-budget-ms 60000 --warm-check-balance-ratio 1.25 "${summary_timing_dir}/warm/visual-in-default" 2>&1)"
warm_visual_status=$?
set -e
assert_equals "$warm_visual_status" "1" "warm check visual-in-default fixture status"
assert_contains "$warm_visual_output" "default warm check includes explicit browser evidence unit check-service-backed:browser-e2e-visual:visual-smoke" "warm check visual-in-default fixture output"
mkdir -p "${summary_timing_dir}/critical/linked/check" "${summary_timing_dir}/critical/unlinked/check"
cat >"${summary_timing_dir}/critical/linked/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"scheduler-start","seq":1,"monotonic_ms":0,"emitted_at":"2026-01-01T00:00:00.000Z"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"start","seq":2,"monotonic_ms":1,"emitted_at":"2026-01-01T00:00:00.001Z","work_unit_id":"setup","work_unit":"setup"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"finish","seq":3,"monotonic_ms":50000,"emitted_at":"2026-01-01T00:00:50.000Z","work_unit_id":"setup","work_unit":"setup","status":0,"duration_ms":50000}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"start","seq":4,"monotonic_ms":50000,"emitted_at":"2026-01-01T00:00:50.000Z","work_unit_id":"build","work_unit":"build"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"finish","seq":5,"monotonic_ms":120000,"emitted_at":"2026-01-01T00:02:00.000Z","work_unit_id":"build","work_unit":"build","status":0,"duration_ms":70000}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"scheduler-finish","seq":6,"monotonic_ms":120000,"emitted_at":"2026-01-01T00:02:00.000Z"}
JSONL
cat >"${summary_timing_dir}/critical/linked/check/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.check_scheduler_summary.v9",
  "target": "check",
  "status": "pass",
  "scheduler_kind": "check",
  "completed_work_units": 2,
  "scheduler_started_monotonic_ms": 0,
  "scheduler_completed_monotonic_ms": 120000,
  "scheduler_total_duration_ms": 120000,
  "scheduler_started_at": "2026-01-01T00:00:00.000Z",
  "scheduler_completed_at": "2026-01-01T00:02:00.000Z",
  "critical_path_wall_duration_ms": 120000,
  "critical_path_units": [
    { "id": "setup", "label": "setup", "kind": "work_unit", "aggregate_target": "setup", "duration_ms": 50000, "started_monotonic_ms": 0, "finished_monotonic_ms": 50000, "needs": [], "completion_keys": ["setup"] },
    { "id": "build", "label": "build", "kind": "work_unit", "aggregate_target": "build", "duration_ms": 70000, "started_monotonic_ms": 50000, "finished_monotonic_ms": 120000, "needs": ["setup"], "completion_keys": ["build"] }
  ],
  "critical_path_blockers": [],
  "critical_path_terminal_unit": { "id": "build", "label": "build", "kind": "work_unit", "aggregate_target": "build", "duration_ms": 70000, "started_monotonic_ms": 50000, "finished_monotonic_ms": 120000, "needs": ["setup"], "completion_keys": ["build"] }
}
JSON
cp "${summary_timing_dir}/valid/check/target-summary.json" "${summary_timing_dir}/critical/linked/check/target-summary.json"
cp "${summary_timing_dir}/valid/check/tool-run-summary.json" "${summary_timing_dir}/critical/linked/check/tool-run-summary.json"
cp "${summary_timing_dir}/valid/run-summary.json" "${summary_timing_dir}/critical/linked/run-summary.json"
cp "${summary_timing_dir}/valid/tool-run-summary.json" "${summary_timing_dir}/critical/linked/tool-run-summary.json"
cp "${summary_timing_dir}/critical/linked/check/scheduler-events.jsonl" "${summary_timing_dir}/critical/unlinked/check/scheduler-events.jsonl"
cp "${summary_timing_dir}/critical/linked/check/scheduler-summary.json" "${summary_timing_dir}/critical/unlinked/check/scheduler-summary.json"
cp "${summary_timing_dir}/critical/linked/check/target-summary.json" "${summary_timing_dir}/critical/unlinked/check/target-summary.json"
cp "${summary_timing_dir}/critical/linked/check/tool-run-summary.json" "${summary_timing_dir}/critical/unlinked/check/tool-run-summary.json"
cp "${summary_timing_dir}/critical/linked/run-summary.json" "${summary_timing_dir}/critical/unlinked/run-summary.json"
cp "${summary_timing_dir}/critical/linked/tool-run-summary.json" "${summary_timing_dir}/critical/unlinked/tool-run-summary.json"
"$NODE_BIN" --input-type=module - "${summary_timing_dir}/critical/unlinked/check/scheduler-summary.json" <<'EOF'
import { readFileSync, writeFileSync } from "node:fs";
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(readFileSync(summaryFile, "utf8"));
summary.critical_path_units[1].needs = ["missing"];
summary.critical_path_terminal_unit.needs = ["missing"];
writeFileSync(summaryFile, `${JSON.stringify(summary, null, 2)}\n`);
EOF
assert_contains "$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" "${summary_timing_dir}/critical/linked" 2>&1)" "scheduler summary timing verified" "linked critical path timing fixture"
set +e
critical_path_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" "${summary_timing_dir}/critical/unlinked" 2>&1)"
critical_path_status=$?
set -e
assert_equals "$critical_path_status" "1" "critical path continuity drift fixture status"
assert_contains "$critical_path_output" "build is not linked to previous unit setup" "critical path continuity drift output"
set +e
summary_timing_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" "${summary_timing_dir}/stale" 2>&1)"
summary_timing_status=$?
set -e
assert_equals "$summary_timing_status" "1" "scheduler summary timing drift fixture status"
assert_contains "$summary_timing_output" "summary completed before final scheduler event" "scheduler summary timing completed-before-final output"
assert_contains "$summary_timing_output" "duration 70000ms is below scheduler total 120000ms" "scheduler summary timing duration output"

parent_work_unit_dir="$(mktemp -d "${ROOT_DIR}/tmp/scheduler-parent-work-unit.XXXXXX")"
cleanup_paths+=("$parent_work_unit_dir")
mkdir -p "${parent_work_unit_dir}/stale/check" "${parent_work_unit_dir}/stale/check-service-backed"
cat >"${parent_work_unit_dir}/stale/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"scheduler-start","seq":1,"monotonic_ms":0,"emitted_at":"2026-01-01T00:00:00.000Z"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"start","seq":2,"monotonic_ms":0,"emitted_at":"2026-01-01T00:00:00.000Z","work_unit_id":"check-service-backed","work_unit":"check-service-backed","work_unit_type":"make_target","aggregate_target":"check-service-backed","nested_scheduler":{"type":"service_backed","target":"check-service-backed"}}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"finish","seq":3,"monotonic_ms":121233,"emitted_at":"2026-01-01T00:02:01.233Z","work_unit_id":"check-service-backed","work_unit":"check-service-backed","status":0,"duration_ms":121233}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"scheduler-finish","seq":4,"monotonic_ms":121233,"emitted_at":"2026-01-01T00:02:01.233Z"}
JSONL
cat >"${parent_work_unit_dir}/stale/check/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.check_scheduler_summary.v9",
  "target": "check",
  "status": "pass",
  "scheduler_kind": "check",
  "scheduler_started_monotonic_ms": 0,
  "scheduler_completed_monotonic_ms": 121233,
  "scheduler_total_duration_ms": 121233,
  "scheduler_started_at": "2026-01-01T00:00:00.000Z",
  "scheduler_completed_at": "2026-01-01T00:02:01.233Z"
}
JSON
cat >"${parent_work_unit_dir}/stale/check/target-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_target_summary.v4",
  "target": "check",
  "status": "pass",
  "start_time": "2026-01-01T00:00:00.000Z",
  "end_time": "2026-01-01T00:02:01.233Z",
  "wall_duration_ms": 121233,
  "critical_path_wall_duration_ms": 121233
}
JSON
cat >"${parent_work_unit_dir}/stale/check/tool-run-summary.json" <<'JSON'
{
  "schema_id": "cartulary.tool_run_summary.v3",
  "target": "check",
  "status": "pass",
  "completed_at": "2026-01-01T00:02:01.233Z",
  "duration_ms": 121233,
  "scheduler_timing": {
    "scheduler_total_duration_ms": 121233
  },
  "extensions": {}
}
JSON
cat >"${parent_work_unit_dir}/stale/check-service-backed/target-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_target_summary.v4",
  "target": "check-service-backed",
  "status": "pass",
  "start_time": "2026-01-01T00:01:04.292Z",
  "end_time": "2026-01-01T00:02:01.233Z",
  "wall_duration_ms": 56941,
  "critical_path_wall_duration_ms": 56941
}
JSON
cat >"${parent_work_unit_dir}/stale/check-service-backed/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_scheduler_summary.v9",
  "target": "check-service-backed",
  "status": "pass",
  "scheduler_kind": "service_backed",
  "scheduler_started_monotonic_ms": 64292,
  "scheduler_completed_monotonic_ms": 121233,
  "scheduler_total_duration_ms": 56941,
  "scheduler_started_at": "2026-01-01T00:01:04.292Z",
  "scheduler_completed_at": "2026-01-01T00:02:01.233Z"
}
JSON
set +e
parent_work_unit_output="$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-summary-timing-drift.mjs" "${parent_work_unit_dir}/stale" 2>&1)"
parent_work_unit_status=$?
set -e
assert_equals "$parent_work_unit_status" "1" "parent scheduler work-unit drift status"
assert_contains "$parent_work_unit_output" "duration 56941ms is below parent scheduler work-unit check-service-backed duration 121233ms" "parent scheduler work-unit duration output"
assert_contains "$parent_work_unit_output" "critical path duration 56941ms is below parent scheduler work-unit check-service-backed duration 121233ms" "parent scheduler work-unit critical output"

full_envelope_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-full-envelope.XXXXXX")"
cleanup_paths+=("$full_envelope_dir")
mkdir -p \
  "${full_envelope_dir}/results/envelope/check-service-backed" \
  "${full_envelope_dir}/results/envelope/browser-e2e-webserver-backed" \
  "${full_envelope_dir}/results/envelope/_shared/test-services/suite/events"
cat >"${full_envelope_dir}/results/envelope/check-service-backed/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.service_backed_scheduler_summary.v9",
  "target": "check-service-backed",
  "status": "pass",
  "scheduler_kind": "service_backed",
  "scheduler_started_monotonic_ms": 0,
  "scheduler_completed_monotonic_ms": 56941,
  "scheduler_total_duration_ms": 56941,
  "scheduler_started_at": "2026-01-01T00:01:04.292Z",
  "scheduler_completed_at": "2026-01-01T00:02:01.233Z"
}
JSON
cat >"${full_envelope_dir}/results/envelope/browser-e2e-webserver-backed/target-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_target_summary.v4",
  "target": "browser-e2e-webserver-backed",
  "status": "pass",
  "start_time": "2026-01-01T00:01:04.292Z",
  "end_time": "2026-01-01T00:02:01.233Z",
  "wall_duration_ms": 56941,
  "critical_path_wall_duration_ms": 56941,
  "executed_duration_ms": 56941,
  "logical_duration_ms": 56941,
  "counts": {
    "phases": 1,
    "tests": 1,
    "failed": 0
  }
}
JSON
cat >"${full_envelope_dir}/results/envelope/_shared/test-services/suite/events/service-wait.json" <<'JSON'
{
  "type": "timing-span",
  "status": "pass",
  "details": {
    "target": "check-service-backed",
    "bucket": "service_wait",
    "label": "test-services start postgres",
    "start_time": "2026-01-01T00:00:00.000Z",
    "end_time": "2026-01-01T00:00:57.074Z",
    "duration_ms": 57074,
    "status": "pass"
  }
}
JSON
CARTULARY_TEST_RESULTS_DIR="${full_envelope_dir}/results" \
CARTULARY_TEST_RUN_ID="envelope" \
  "$TEST_OUTPUT_SCRIPT" target-summary check-service-backed pass --children browser-e2e-webserver-backed --quiet-success
full_envelope_summary="${full_envelope_dir}/results/envelope/check-service-backed/target-summary.json"
full_envelope_timing="${full_envelope_dir}/results/envelope/check-service-backed/target-timing.json"
assert_equals "$(json_field "$full_envelope_summary" "wall_duration_ms")" "121233" "service-backed full envelope wall duration"
assert_equals "$(json_field "$full_envelope_summary" "critical_path_wall_duration_ms")" "121233" "service-backed full envelope critical duration"
assert_equals "$(json_field "$full_envelope_summary" "scheduler_timing.scheduler_total_duration_ms")" "56941" "service-backed nested scheduler provenance"
assert_equals "$(json_field "$full_envelope_summary" "totals.slowest_lifecycle_bucket.name")" "service_wait" "service-backed full envelope slowest bucket"
assert_equals "$(json_field "$full_envelope_timing" "slowest_lifecycle_bucket.name")" "service_wait" "service-backed target timing slowest bucket"

success_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-success.XXXXXX")"
cleanup_paths+=("$success_dir")
write_fake_make "$success_dir"
success_manifest="${success_dir}/manifest.json"
check_auto_capacity="$(
  cd "$ROOT_DIR" &&
    "$NODE_BIN" --input-type=module - <<'EOF'
import {
  estimateCheckHostCPULimit,
  estimateCheckHostIOLimit,
} from "./scripts/lib/scheduler-resources.mjs";

const hostCPU = estimateCheckHostCPULimit();
const hostIO = estimateCheckHostIOLimit(new Map([["host_cpu", hostCPU]]));
process.stdout.write(`${hostCPU},${hostIO}`);
EOF
)"
check_auto_cpu="${check_auto_capacity%,*}"
check_auto_io="${check_auto_capacity#*,}"
cat >"$success_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "capacity_profile": "check_default",
      "resource_limits": { "host_cpu": "auto", "host_io": "auto", "suite_service_stack": 1, "migration_scratch_postgres": 1 },
      "summary_groups": [
        { "name": "check-work", "summary_targets": ["local", "service", "meta"] }
      ],
      "work_units": [
        { "target": "setup", "weight_ms": 50, "needs": [], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "build", "weight_ms": 40, "needs": ["setup"], "resource_claims": { "host_cpu": "limit" }, "make_jobs": "host_cpu" },
        { "target": "local", "weight_ms": 30, "needs": ["build"], "produces_summary_targets": ["local"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "env": { "LINT_SHELL_STRICT": "1" } },
        {
          "target": "service",
          "weight_ms": 20,
          "needs": ["build"],
          "produces_summary_targets": ["service"],
          "resource_claims": {
            "host_cpu": { "mode": "bounded_limit", "reserve": 3, "min": 1, "max": 8 },
            "host_io": { "mode": "bounded_limit", "reserve": 4, "min": 1, "max": 10 },
            "suite_service_stack": 1
          },
          "make_jobs": "host_cpu"
        },
        { "target": "meta", "weight_ms": 10, "needs": ["build"], "produces_summary_targets": ["meta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" }
      ]
    }
  ]
}
JSON
default_capacity_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_LOCAL=0.01 FAKE_SLEEP_SERVICE=0.01 run_scheduler "$success_dir" "$success_manifest" default-capacity 2>&1)"
assert_contains "$default_capacity_output" "[CHECK-SCHEDULER] check start work_units=5 capacity={host_cpu:${check_auto_cpu},host_io:${check_auto_io},suite_service_stack:1,migration_scratch_postgres:1}" "default capacity comes from registry"
"$NODE_BIN" - "${success_dir}/results/default-capacity/check/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.resource_limit_sources?.host_cpu !== "auto:check_host_cpu") {
  throw new Error(`default host_cpu source got ${summary.resource_limit_sources?.host_cpu}`);
}
if (summary.resource_limit_sources?.host_io !== "auto:check_host_io") {
  throw new Error(`default host_io source got ${summary.resource_limit_sources?.host_io}`);
}
EOF
env_capacity_output="$(CHECK_HOST_CPU_JOBS=5 CHECK_HOST_IO_JOBS=4 CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_LOCAL=0.01 FAKE_SLEEP_SERVICE=0.01 run_scheduler "$success_dir" "$success_manifest" env-capacity 2>&1)"
assert_contains "$env_capacity_output" "[CHECK-SCHEDULER] check start work_units=5 capacity={host_cpu:5,host_io:4,suite_service_stack:1,migration_scratch_postgres:1}" "env capacity overrides registry default"
"$NODE_BIN" - "${success_dir}/results/env-capacity/check/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.resource_limit_sources?.host_cpu !== "env:CHECK_HOST_CPU_JOBS") {
  throw new Error(`env host_cpu source got ${summary.resource_limit_sources?.host_cpu}`);
}
if (summary.resource_limit_sources?.host_io !== "env:CHECK_HOST_IO_JOBS") {
  throw new Error(`env host_io source got ${summary.resource_limit_sources?.host_io}`);
}
EOF

cpu_constrained_manifest="${success_dir}/cpu-constrained-manifest.json"
cat >"$cpu_constrained_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "capacity_profile": "check_default",
      "resource_limits": { "host_cpu": "auto", "host_io": "auto", "suite_service_stack": 1, "migration_scratch_postgres": 1 },
      "summary_groups": [
        { "name": "io-heavy-work", "summary_targets": ["io-heavy"] }
      ],
      "work_units": [
        { "target": "io-heavy", "weight_ms": 10, "needs": [], "produces_summary_targets": ["io-heavy"], "resource_claims": { "host_cpu": 1, "host_io": 3 }, "make_jobs": "host_cpu" }
      ]
    }
  ]
}
JSON
cpu_constrained_output="$(CHECK_HOST_CPU_JOBS=2 CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_LOCAL=0.01 run_scheduler "$success_dir" "$cpu_constrained_manifest" cpu-constrained 2>&1)"
assert_contains "$cpu_constrained_output" "[CHECK-SCHEDULER] check start work_units=1 capacity={host_cpu:2,host_io:3,suite_service_stack:1,migration_scratch_postgres:1}" "auto host_io must not resolve below declared claims under constrained host_cpu"

browser_auto_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-browser-auto.XXXXXX")"
cleanup_paths+=("$browser_auto_dir")
write_fake_make "$browser_auto_dir"
browser_auto_manifest="${browser_auto_dir}/manifest.json"
cat >"$browser_auto_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": {
        "host_cpu": 12,
        "host_io": 12,
        "suite_service_stack": 1,
        "migration_scratch_postgres": 1,
        "browser_stack": "auto",
        "postgres_clone": "auto",
        "browser_stage_webserver_backed": 1,
        "browser_stage_stateful": 1,
        "browser_stage_measurement": 1,
        "browser_stage_visual": 1,
        "process": 4
      },
      "summary_groups": [
        { "name": "browser", "summary_targets": ["browser-e2e-webserver-backed", "browser-e2e-stateful", "browser-e2e-measurement", "browser-e2e-visual"] }
      ],
      "work_units": [
        { "target": "browser-e2e-webserver-backed", "weight_ms": 40, "needs": [], "produces_summary_targets": ["browser-e2e-webserver-backed"], "resource_claims": { "host_cpu": 1, "host_io": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1 }, "make_jobs": "host_cpu" },
        { "target": "browser-e2e-stateful", "weight_ms": 30, "needs": [], "produces_summary_targets": ["browser-e2e-stateful"], "resource_claims": { "host_cpu": 1, "host_io": 1, "process": 1, "browser_stack": 1, "browser_stage_stateful": 1 }, "make_jobs": "host_cpu" },
        { "target": "browser-e2e-measurement", "weight_ms": 20, "needs": [], "produces_summary_targets": ["browser-e2e-measurement"], "resource_claims": { "host_cpu": "limit", "host_io": "limit", "process": "limit", "browser_stack": "limit", "browser_stage_measurement": 1 }, "make_jobs": "host_cpu" },
        { "target": "browser-e2e-visual", "weight_ms": 10, "needs": [], "produces_summary_targets": ["browser-e2e-visual"], "resource_claims": { "host_cpu": 1, "host_io": 1, "process": 1, "browser_stack": 1, "browser_stage_visual": 1 }, "make_jobs": "host_cpu" }
      ]
    }
  ]
}
JSON
browser_auto_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_DEFAULT=0.2 run_scheduler "$browser_auto_dir" "$browser_auto_manifest" browser-auto 2>&1)"
assert_contains "$browser_auto_output" "browser_stack:4" "check browser stack auto capacity resolves to all four browser stages"
assert_equals "$(cat "${browser_auto_dir}/max")" "2" "check browser auto capacity keeps limit-claim measurement from overlapping sibling stages"
"$NODE_BIN" - "${browser_auto_dir}/results/browser-auto/check/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.resource_limits?.browser_stack !== 4) {
  throw new Error(`browser_stack limit got ${summary.resource_limits?.browser_stack}`);
}
if (summary.resource_limit_sources?.browser_stack !== "auto:service_backed_browser_stack") {
  throw new Error(`browser_stack source got ${summary.resource_limit_sources?.browser_stack}`);
}
if (summary.resource_limits?.postgres_clone !== 6) {
  throw new Error(`postgres_clone limit got ${summary.resource_limits?.postgres_clone}`);
}
if (summary.resource_limit_sources?.postgres_clone !== "auto:service_backed_postgres_clone") {
  throw new Error(`postgres_clone source got ${summary.resource_limit_sources?.postgres_clone}`);
}
EOF
browser_override_output="$(CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=3 CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT=5 CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_DEFAULT=0.01 run_scheduler "$browser_auto_dir" "$browser_auto_manifest" browser-override 2>&1)"
assert_contains "$browser_override_output" "browser_stack:3" "check browser stack env override applies to flattened check schedule"
"$NODE_BIN" - "${browser_auto_dir}/results/browser-override/check/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.resource_limit_sources?.browser_stack !== "env:CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT") {
  throw new Error(`browser_stack override source got ${summary.resource_limit_sources?.browser_stack}`);
}
if (summary.resource_limit_sources?.postgres_clone !== "env:CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT") {
  throw new Error(`postgres_clone override source got ${summary.resource_limit_sources?.postgres_clone}`);
}
EOF

priority_reservation_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-priority-reservation.XXXXXX")"
cleanup_paths+=("$priority_reservation_dir")
write_fake_make "$priority_reservation_dir"
priority_reservation_manifest="${priority_reservation_dir}/manifest.json"
cat >"$priority_reservation_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 2, "host_io": 1, "suite_service_stack": 1, "migration_scratch_postgres": 1 },
      "summary_groups": [
        { "name": "priority-reservation", "summary_targets": ["alpha", "build-server", "low-cpu", "low-io"] }
      ],
      "work_units": [
        { "target": "alpha", "weight_ms": 50, "needs": [], "produces_summary_targets": ["alpha"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "build-server", "weight_ms": 40, "needs": [], "produces_summary_targets": ["build-server"], "resource_claims": { "host_cpu": 2 }, "make_jobs": "host_cpu" },
        { "target": "low-cpu", "weight_ms": 30, "needs": [], "produces_summary_targets": ["low-cpu"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "low-io", "weight_ms": 20, "needs": [], "produces_summary_targets": ["low-io"], "resource_claims": { "host_io": 1 }, "make_jobs": 1 }
      ]
    }
  ]
}
JSON
priority_reservation_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_ALPHA=0.2 FAKE_SLEEP_DEFAULT=0.01 run_scheduler "$priority_reservation_dir" "$priority_reservation_manifest" priority-reservation 2>&1)"
assert_contains "$priority_reservation_output" "[SUMMARY] target=check status=pass work_units=4/4" "priority reservation scheduler pass summary"
"$NODE_BIN" - "${priority_reservation_dir}/events.log" <<'EOF'
const fs = require("node:fs");
const [eventLog] = process.argv.slice(2);
const lines = fs.readFileSync(eventLog, "utf8").trim().split(/\n/).filter(Boolean);
const indexOf = (needle) => {
  const index = lines.findIndex((line) => line.startsWith(needle));
  if (index === -1) {
    throw new Error(`missing event ${needle}\n${lines.join("\n")}`);
  }
  return index;
};
const startAlpha = indexOf("start alpha ");
const startLowIO = indexOf("start low-io ");
const endLowIO = indexOf("end low-io ");
const endAlpha = indexOf("end alpha ");
const startBuild = indexOf("start build-server ");
const startLowCPU = indexOf("start low-cpu ");
if (!(startAlpha < startLowIO && startLowIO < endLowIO && endLowIO < endAlpha)) {
  throw new Error("unrelated host_io work must backfill while build-server waits for host_cpu");
}
if (!(endAlpha < startBuild && startBuild < startLowCPU)) {
  throw new Error("lower-priority host_cpu work must not backfill before build-server starts");
}
EOF

service_priority_reservation_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-service-priority-reservation.XXXXXX")"
cleanup_paths+=("$service_priority_reservation_dir")
write_fake_make "$service_priority_reservation_dir"
service_priority_reservation_manifest="${service_priority_reservation_dir}/manifest.json"
cat >"$service_priority_reservation_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 2, "host_io": 1, "suite_service_stack": 1, "migration_scratch_postgres": 1 },
      "summary_groups": [
        { "name": "service-priority-reservation", "summary_targets": ["alpha", "service-child", "static-child", "drift-io"] }
      ],
      "work_units": [
        { "target": "alpha", "priority": 36000, "weight_ms": 50, "needs": [], "produces_summary_targets": ["alpha"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "service-child", "priority": 35000, "weight_ms": 40, "needs": [], "produces_summary_targets": ["service-child"], "resource_claims": { "host_cpu": 2 }, "make_jobs": "host_cpu" },
        { "target": "static-child", "priority": 13000, "weight_ms": 30, "needs": [], "produces_summary_targets": ["static-child"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "drift-io", "priority": 11000, "weight_ms": 20, "needs": [], "produces_summary_targets": ["drift-io"], "resource_claims": { "host_io": 1 }, "make_jobs": 1 }
      ]
    }
  ]
}
JSON
service_priority_reservation_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_ALPHA=0.2 FAKE_SLEEP_DEFAULT=0.01 run_scheduler "$service_priority_reservation_dir" "$service_priority_reservation_manifest" service-priority-reservation 2>&1)"
assert_contains "$service_priority_reservation_output" "[SUMMARY] target=check status=pass work_units=4/4" "service-backed priority reservation scheduler pass summary"
"$NODE_BIN" - "${service_priority_reservation_dir}/events.log" <<'EOF'
const fs = require("node:fs");
const [eventLog] = process.argv.slice(2);
const lines = fs.readFileSync(eventLog, "utf8").trim().split(/\n/).filter(Boolean);
const indexOf = (needle) => {
  const index = lines.findIndex((line) => line.startsWith(needle));
  if (index === -1) {
    throw new Error(`missing event ${needle}\n${lines.join("\n")}`);
  }
  return index;
};
const startAlpha = indexOf("start alpha ");
const startDriftIO = indexOf("start drift-io ");
const endDriftIO = indexOf("end drift-io ");
const endAlpha = indexOf("end alpha ");
const startService = indexOf("start service-child ");
const startStatic = indexOf("start static-child ");
if (!(startAlpha < startDriftIO && startDriftIO < endDriftIO && endDriftIO < endAlpha)) {
  throw new Error("unrelated lower-priority IO work may run while ready service-backed CPU work waits");
}
if (!(endAlpha < startService && startService < startStatic)) {
  throw new Error("lower-priority overlapping host_cpu work must not start before ready service-backed work");
}
EOF

scheduler_priority_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-scheduler-priority.XXXXXX")"
cleanup_paths+=("$scheduler_priority_dir")
write_fake_make "$scheduler_priority_dir"
scheduler_priority_manifest="${scheduler_priority_dir}/manifest.json"
cat >"$scheduler_priority_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1, "host_io": 1, "suite_service_stack": 1, "migration_scratch_postgres": 1 },
      "summary_groups": [
        { "name": "scheduler-priority", "summary_targets": ["alpha", "beta"] }
      ],
      "work_units": [
        { "target": "alpha", "priority": 0, "weight_ms": 100, "needs": [], "produces_summary_targets": ["alpha"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "beta", "priority": 10, "weight_ms": 1, "needs": [], "produces_summary_targets": ["beta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" }
      ]
    }
  ]
}
JSON
scheduler_priority_output="$(FAKE_SLEEP_DEFAULT=0.01 run_scheduler "$scheduler_priority_dir" "$scheduler_priority_manifest" scheduler-priority 2>&1)"
assert_contains "$scheduler_priority_output" "[SUMMARY] target=check status=pass work_units=2/2" "scheduler priority pass summary"
"$NODE_BIN" - "${scheduler_priority_dir}/events.log" <<'EOF'
const fs = require("node:fs");
const [eventLog] = process.argv.slice(2);
const lines = fs.readFileSync(eventLog, "utf8").trim().split(/\n/).filter(Boolean);
const startBeta = lines.findIndex((line) => line.startsWith("start beta "));
const startAlpha = lines.findIndex((line) => line.startsWith("start alpha "));
if (startBeta === -1 || startAlpha === -1 || startBeta > startAlpha) {
  throw new Error(`priority must outrank duration weight, got\n${lines.join("\n")}`);
}
EOF
success_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_LOCAL=0.2 FAKE_SLEEP_SERVICE=0.2 run_scheduler "$success_dir" "$success_manifest" success --resource-limit host_cpu=2 --resource-limit host_io=3 2>&1)"
assert_not_contains "$success_output" "[RUN] check" "success hides legacy run start"
assert_contains "$success_output" "[CHECK-SCHEDULER] check start work_units=5 capacity={host_cpu:2,host_io:3,suite_service_stack:1,migration_scratch_postgres:1}" "success concise scheduler start"
assert_contains "$success_output" "[PROGRESS] target=check completed=0/5" "success human scheduler progress"
assert_contains "$success_output" "blocker=dependencies" "success human scheduler progress explains blocker"
assert_not_contains "$success_output" "bottleneck=service:3/6" "success flattened service work has no nested bottleneck"
assert_not_contains "$success_output" "[CHECK-SCHEDULER] check progress completed_work_units=" "quiet check scheduler hides key/value progress"
assert_not_contains "$success_output" "[CHECK-SCHEDULER] check nested-progress" "quiet check scheduler hides key/value nested progress"
assert_contains "$success_output" "[SUMMARY] target=check status=pass work_units=5/5 total_wall_time=" "success concise scheduler summary"
assert_contains "$success_output" " failed=none slowest=" "success concise scheduler summary preserves failed and slowest field order"
assert_contains "$success_output" "slowest=" "success concise scheduler summary includes slowest work"
assert_contains "$success_output" "blockers=" "success concise scheduler summary includes blockers"
assert_not_contains "$success_output" "active_resource_claims=" "default scheduler output hides raw active resources"
assert_not_contains "$success_output" "resource_limits=" "default scheduler output hides raw resource limits"
assert_not_contains "$success_output" "claims={" "default scheduler output hides raw claims"
assert_not_contains "$success_output" "[STEP] check" "default scheduler output hides per-unit steps"
assert_not_contains "$success_output" "running_units=" "default scheduler output hides raw running units"
assert_not_contains "$success_output" "blocked_resources=" "default scheduler output hides raw blocked resources"
assert_contains "$success_output" "[RESULT] target=check status=pass" "success summary"
assert_contains "$(cat "${success_dir}/make-args.log")" "--output-sync=target -j2 build" "build uses claimed host_cpu jobs"
success_events="$(cat "${success_dir}/events.log")"
assert_contains "$success_events" "end local" "success local completed"
assert_contains "$success_events" "strict-env local 1" "success child make receives target-scoped env"
assert_contains "$success_events" "strict-env service unset" "success sibling child make does not inherit target-scoped env"
assert_fake_make_overlap "${success_dir}/events.log" local service "success service overlapped with cheap local work"
assert_not_contains "$success_events" "env service go_cpu=" "success service receives no forwarded nested scheduler limits"
assert_contains "$success_events" "end service" "success service completed"
assert_contains "$success_events" "end meta" "success meta completed"
assert_not_contains "$success_events" "browser" "success check schedule has no browser tail"
assert_contains "$success_events" "test-target setup setup" "success setup receives target-owned identity"
assert_file_present "${success_dir}/results/success/setup/setup/phase-summary.json" "success setup target-owned helper phase summary"
assert_file_absent "${success_dir}/results/success/adhoc/setup/phase-summary.json" "success setup helper is not adhoc-owned"
success_summary="${success_dir}/results/success/run-summary.json"
success_target_summary="${success_dir}/results/success/check/target-summary.json"
success_scheduler_summary="${success_dir}/results/success/check/scheduler-summary.json"
success_scheduler_events="${success_dir}/results/success/check/scheduler-events.jsonl"
assert_contains "$(cat "${success_dir}/results/success/check/progress-summary.log")" "total_wall_time=" "success progress summary retains wall time field"
SUCCESS_OUTPUT="$success_output" "$NODE_BIN" - "$success_scheduler_summary" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const output = process.env.SUCCESS_OUTPUT ?? "";
const match = output.match(/\[SUMMARY\] target=check status=pass work_units=5\/5 total_wall_time=([0-9]+\.[0-9]{2}s) failed=none slowest=/);
if (!match) {
  throw new Error("success scheduler summary must include total_wall_time before failed");
}
const reportedMs = Math.round(Number(match[1].slice(0, -1)) * 1000);
if (!Number.isInteger(summary.scheduler_total_duration_ms) || summary.scheduler_total_duration_ms < 0) {
  throw new Error(`scheduler_total_duration_ms got ${summary.scheduler_total_duration_ms}`);
}
if (Math.abs(reportedMs - summary.scheduler_total_duration_ms) > 10) {
  throw new Error(`total_wall_time ${reportedMs}ms does not match scheduler_total_duration_ms ${summary.scheduler_total_duration_ms}ms`);
}
EOF
assert_equals "$(json_field "$success_summary" "status")" "pass" "success summary status"
assert_equals "$(json_field "$success_summary" "schema_id")" "cartulary.test_run_summary.v6" "success run summary schema"
assert_file_present "$success_target_summary" "success check target summary"
assert_equals "$(json_field "$success_target_summary" "target")" "check" "success check target summary identity"
assert_equals "$(json_field "$success_target_summary" "status")" "pass" "success check target summary status"
assert_equals "$(json_field "$success_summary" "wall_duration_ms")" "$(json_field "$success_scheduler_summary" "scheduler_total_duration_ms")" "success run summary wall uses scheduler total"
assert_equals "$(json_field "$success_target_summary" "wall_duration_ms")" "$(json_field "$success_scheduler_summary" "scheduler_total_duration_ms")" "success target summary wall uses scheduler total"
assert_equals "$(json_field "${success_dir}/results/success/tool-run-summary.json" "scheduler_timing.scheduler_total_duration_ms")" "$(json_field "$success_scheduler_summary" "scheduler_total_duration_ms")" "success run tool summary scheduler timing"
assert_equals "$(json_field "${success_dir}/results/success/check/tool-run-summary.json" "scheduler_timing.scheduler_total_duration_ms")" "$(json_field "$success_scheduler_summary" "scheduler_total_duration_ms")" "success target tool summary scheduler timing"
assert_equals "$(json_field "$success_summary" "work_units.completed")" "5" "success completed work units"
assert_equals "$(json_field "$success_summary" "work_units.total")" "5" "success total work units"
assert_equals "$(json_field "$success_summary" "summary_targets.expected.length")" "3" "success summary target count"
assert_equals "$(json_field "$success_summary" "evidence_targets.present.length")" "3" "success evidence target count"
assert_equals "$(json_field "$success_summary" "helper_units.total")" "2" "success helper unit count"
assert_equals "$(json_field "$success_summary" "helper_units.artifacts.0.target")" "setup" "success helper artifact target"
assert_contains "$(json_field "$success_summary" "helper_units.artifacts.0.phase_summaries.0.artifact")" "/setup/setup/phase-summary.json" "success helper artifact phase summary path"
assert_contains "$(json_field "$success_summary" "helper_units.artifacts.0.phase_summaries.0.runner_json")" "/setup/setup/runner.json" "success helper artifact runner path"
assert_contains "$(json_field "$success_summary" "helper_units.artifacts.0.phase_summaries.0.stdout_log")" "/setup/setup/stdout.log" "success helper artifact stdout path"
assert_contains "$(json_field "$success_summary" "helper_units.artifacts.0.phase_summaries.0.stderr_log")" "/setup/setup/stderr.log" "success helper artifact stderr path"
assert_check_scheduler_artifacts "$success_dir" success check pass - 5 finish
assert_equals "$(json_field "$success_scheduler_summary" "completed_work_units")" "5" "success scheduler completed count"
"$NODE_BIN" - "$success_scheduler_summary" "$success_scheduler_events" <<'EOF'
const fs = require("node:fs");
const [summaryFile, eventsFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).map((line) => JSON.parse(line));
if (!events.some((event) => event.active_resource_claims && Object.keys(event.active_resource_claims).length > 0)) {
  throw new Error("scheduler events must preserve active resource claims");
}
if (
  !events.some(
    (event) =>
      event.resource_limits?.host_cpu === 2 &&
      event.resource_limits?.host_io === 3 &&
      event.resource_limits?.suite_service_stack === 1 &&
      event.resource_limits?.migration_scratch_postgres === 1,
  )
) {
  throw new Error("scheduler events must preserve resource limits");
}
if (summary.resource_limit_sources?.host_cpu !== "cli" || summary.resource_limit_sources?.host_io !== "cli") {
  throw new Error("scheduler summary must record CLI resource-limit override sources");
}
if (summary.max_running_work_units !== 2) {
  throw new Error(`max running work units got ${summary.max_running_work_units} want 2`);
}
if (summary.max_running_groups < 1) {
  throw new Error(`max running groups got ${summary.max_running_groups} want at least 1`);
}
if (!summary.blocked_explanations_seen.includes("dependencies")) {
  throw new Error("summary must record dependency blocked explanation");
}
if (!summary.top_blockers?.some((entry) => entry.kind === "dependency" && entry.name === "setup" && entry.count > 0)) {
  throw new Error("summary must record top dependency blocker");
}
if (!summary.waiting_on_seen?.includes("setup")) {
  throw new Error("summary must record dependency waiting_on target");
}
const dependencyBlocked = events.find((event) => event.event === "blocked" && event.blocked_reason === "dependencies");
if (!dependencyBlocked) {
  throw new Error("scheduler events must record dependency blocked event");
}
if (!dependencyBlocked.waiting_on?.includes("setup")) {
  throw new Error("dependency blocked event must record waiting_on setup");
}
const buildBlocked = dependencyBlocked.blocked_units?.find((entry) => entry.work_unit === "build");
if (!buildBlocked?.waiting_on?.includes("setup")) {
  throw new Error("blocked_units must record build waiting on setup");
}
if (summary.max_active_resource_claims?.host_cpu !== 2) {
  throw new Error(`max active host_cpu claims got ${summary.max_active_resource_claims?.host_cpu} want 2`);
}
if (!events.some((event) => event.running >= 2 && event.active_resource_claims?.host_cpu === 2)) {
  throw new Error("scheduler events must record two logically admitted host_cpu work units");
}
if (events.some((event) => event.nested_scheduler)) {
  throw new Error("flattened service work must not emit nested scheduler start metadata");
}
if (events.some((event) => event.nested_scheduler_progress?.length > 0)) {
  throw new Error("flattened service work must not emit nested scheduler progress");
}
if ((summary.nested_scheduler_limits ?? []).length !== 0 || (summary.nested_scheduler_observations ?? []).length !== 0) {
  throw new Error("flattened service work must not record nested scheduler summaries");
}
for (const [index, event] of events.entries()) {
  const limits = event.resource_limits ?? {};
  for (const [resource, amount] of Object.entries(event.active_resource_claims ?? {})) {
    const limit = limits[resource];
    if (!Number.isInteger(limit)) {
      throw new Error(`event ${index} active claim ${resource} has no integer resource limit`);
    }
    if (amount > limit) {
      throw new Error(`event ${index} active claim ${resource}=${amount} exceeds limit ${limit}`);
    }
  }
}
EOF

blocker_clarity_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-blocker-clarity.XXXXXX")"
cleanup_paths+=("$blocker_clarity_dir")
write_fake_make "$blocker_clarity_dir"
blocker_clarity_manifest="${blocker_clarity_dir}/manifest.json"
cat >"$blocker_clarity_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1 },
      "summary_groups": [
        { "name": "blocker-clarity", "summary_targets": ["alpha", "beta", "meta"] }
      ],
      "work_units": [
        { "target": "alpha", "weight_ms": 3, "needs": [], "produces_summary_targets": ["alpha"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "beta", "weight_ms": 2, "needs": [], "produces_summary_targets": ["beta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "meta", "weight_ms": 1, "needs": ["alpha"], "produces_summary_targets": ["meta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" }
      ]
    }
  ]
}
JSON
FAKE_SLEEP_ALPHA=0.15 \
  run_scheduler "$blocker_clarity_dir" "$blocker_clarity_manifest" blocker-clarity --resource-limit host_cpu=1 \
  >"${blocker_clarity_dir}/stdout.log" \
  2>"${blocker_clarity_dir}/stderr.log"
blocker_clarity_output="$(cat "${blocker_clarity_dir}/stdout.log")"
assert_contains "$blocker_clarity_output" "[PROGRESS] target=check completed=0/3" "quiet scheduler emits progress when blocker state first changes"
assert_contains "$blocker_clarity_output" "blocker=dependencies,host_cpu" "quiet scheduler explains dependency and resource blockers"
assert_occurrences "$blocker_clarity_output" "[PROGRESS] target=check" "2" "quiet blocker-change progress is capped"
assert_output_budget "${ROOT_DIR}/tools/task_surface_manifest.json" check "${blocker_clarity_dir}/stdout.log" "${blocker_clarity_dir}/stderr.log" "scheduler blocker clarity budget"
"$NODE_BIN" - "${blocker_clarity_dir}/results/blocker-clarity/check/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (!summary.blocked_explanations_seen?.includes("dependencies")) {
  throw new Error("summary must record dependency blocker explanations from blocked events");
}
if (!summary.blocked_explanations_seen?.includes("host_cpu")) {
  throw new Error("summary must record resource blocker explanations from blocked events");
}
const hostCPU = summary.top_blockers?.find((entry) => entry.kind === "resource" && entry.name === "host_cpu");
if (!hostCPU || hostCPU.count < 1) {
  throw new Error("summary must rank host_cpu as a top resource blocker");
}
const alpha = summary.top_blockers?.find((entry) => entry.kind === "dependency" && entry.name === "alpha");
if (!alpha || alpha.count < 1) {
  throw new Error("summary must rank alpha as a top dependency blocker");
}
EOF

success_budget_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-budget.XXXXXX")"
cleanup_paths+=("$success_budget_dir")
write_fake_make "$success_budget_dir"
success_budget_manifest="${success_budget_dir}/manifest.json"
cat >"$success_budget_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1 },
      "work_units": [
        { "target": "alpha", "weight_ms": 1, "needs": [], "produces_summary_targets": ["alpha"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" }
      ]
    }
  ]
}
JSON
FAKE_SLEEP_ALPHA=0.01 \
  run_scheduler "$success_budget_dir" "$success_budget_manifest" success-budget --resource-limit host_cpu=1 \
  >"${success_budget_dir}/stdout.log" \
  2>"${success_budget_dir}/stderr.log"
assert_output_budget "${ROOT_DIR}/tools/task_surface_manifest.json" check "${success_budget_dir}/stdout.log" "${success_budget_dir}/stderr.log" "scheduler success budget"

verbose_output="$(VERBOSE=1 run_scheduler "$success_dir" "$success_manifest" verbose --resource-limit host_cpu=2 --resource-limit host_io=3 2>&1)"
assert_contains "$verbose_output" "[CHECK-SCHEDULER] check start work_unit=setup claims={host_cpu:1} active=1 pending=4" "verbose scheduler start telemetry"
assert_contains "$verbose_output" "active_resource_claims={host_cpu:1}" "verbose scheduler active resource telemetry"
assert_contains "$verbose_output" "resource_limits={host_cpu:2,host_io:3,suite_service_stack:1,migration_scratch_postgres:1}" "verbose scheduler resource limit telemetry"

split_lane_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-split-lanes.XXXXXX")"
cleanup_paths+=("$split_lane_dir")
write_fake_make "$split_lane_dir"
split_lane_manifest="${split_lane_dir}/manifest.json"
cat >"$split_lane_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": {
        "host_cpu": 2,
        "host_io": 2,
        "suite_service_stack": 1,
        "migration_scratch_postgres": 1
      },
      "summary_groups": [
        { "name": "split-lane-work", "summary_targets": ["check-service-backed", "migration-drift"] }
      ],
      "work_units": [
        {
          "target": "build-migrate",
          "weight_ms": 30,
          "needs": [],
          "resource_claims": { "host_cpu": 1 },
          "make_jobs": "host_cpu"
        },
        {
          "target": "check-service-backed",
          "weight_ms": 20,
          "needs": ["build-migrate"],
          "produces_summary_targets": ["check-service-backed"],
          "resource_claims": { "host_cpu": 1, "host_io": 1, "suite_service_stack": 1 },
          "make_jobs": "host_cpu"
        },
        {
          "target": "migration-drift",
          "weight_ms": 10,
          "needs": ["build-migrate"],
          "produces_summary_targets": ["migration-drift"],
          "resource_claims": { "host_cpu": 1, "host_io": 1, "migration_scratch_postgres": 1 },
          "make_jobs": "host_cpu"
        }
      ]
    }
  ]
}
JSON
set +e
FAKE_SLEEP_CHECK_SERVICE_BACKED=0.2 \
FAKE_SLEEP_MIGRATION_DRIFT=0.2 \
  run_scheduler "$split_lane_dir" "$split_lane_manifest" split-lanes \
  >"${split_lane_dir}/stdout.log" \
  2>"${split_lane_dir}/stderr.log"
split_lane_status=$?
set -e
if [[ "$split_lane_status" != "0" ]]; then
  cat "${split_lane_dir}/stdout.log"
  cat "${split_lane_dir}/stderr.log" >&2
  fail "split service lanes scheduler fixture failed"
fi
assert_equals "$(cat "${split_lane_dir}/max")" "2" "split service lanes allow service-backed and migration drift concurrency"
assert_contains "$(cat "${split_lane_dir}/events.log")" "start migration-drift" "migration drift starts in split service lane fixture"

machine_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-machine.XXXXXX")"
cleanup_paths+=("$machine_dir")
write_fake_make "$machine_dir"
CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 \
FAKE_SLEEP_SERVICE=0.2 \
CARTULARY_OUTPUT_MODE=machine \
  run_scheduler "$machine_dir" "$success_manifest" machine --resource-limit host_cpu=2 --resource-limit host_io=3 \
  >"${machine_dir}/stdout.log" \
  2>"${machine_dir}/stderr.log"
assert_single_machine_json \
  "${machine_dir}/stdout.log" \
  "${machine_dir}/stderr.log" \
  check \
  "machine check scheduler summary" \
  tool_run_summary \
  target_summary \
  run_summary \
  run_tool_run_summary \
  scheduler_summary \
  scheduler_events \
  --log \
  scheduler_progress \
  scheduler_logs

partial_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-partial-nested.XXXXXX")"
cleanup_paths+=("$partial_dir")
write_fake_make "$partial_dir"
partial_manifest="${partial_dir}/manifest.json"
cat >"$partial_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1, "host_io": 1 },
      "work_units": [
        {
          "target": "partial-service",
          "weight_ms": 1,
          "needs": [],
          "produces_summary_targets": ["partial-service"],
          "resource_claims": { "host_cpu": 1, "host_io": 1 },
          "make_jobs": "host_cpu"
        }
      ]
    }
  ]
}
JSON
partial_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_PARTIAL_SERVICE=0.15 run_scheduler "$partial_dir" "$partial_manifest" partial --resource-limit host_cpu=1 --resource-limit host_io=1 2>&1)"
assert_contains "$partial_output" "[RESULT] target=check status=pass" "partial nested event does not fail check scheduler"
assert_not_contains "$partial_output" "nested-progress work_unit=partial-service" "partial nested event is ignored until newline-complete"

makeflags_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-makeflags.XXXXXX")"
cleanup_paths+=("$makeflags_dir")
write_fake_make "$makeflags_dir"
makeflags_manifest="${makeflags_dir}/manifest.json"
cat >"$makeflags_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1 },
      "work_units": [
        { "target": "alpha", "weight_ms": 1, "needs": [], "produces_summary_targets": ["alpha"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" }
      ]
    }
  ]
}
JSON
makeflags_output="$(
  MAKEFLAGS='--jobserver-auth=3,4 -j --trace' \
  MFLAGS='--jobserver-fds=3,4 -j' \
    run_scheduler "$makeflags_dir" "$makeflags_manifest" makeflags --resource-limit host_cpu=1 2>&1
)"
assert_contains "$makeflags_output" "[RESULT] target=check status=pass" "makeflags sanitize summary"
makeflags_events="$(cat "${makeflags_dir}/events.log")"
assert_not_contains "$makeflags_events" "jobserver" "check child make env strips inherited jobserver tokens"
assert_not_contains "$makeflags_events" "MFLAGS=-j" "check child make env strips inherited mflags jobs"
assert_contains "$makeflags_events" "MAKEFLAGS=--trace" "check child make env preserves non-jobserver make flags"

failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-failure.XXXXXX")"
cleanup_paths+=("$failure_dir")
write_fake_make "$failure_dir"
failure_manifest="${failure_dir}/manifest.json"
cat >"$failure_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 2, "suite_service_stack": 1 },
      "work_units": [
        { "target": "alpha", "weight_ms": 30, "needs": [], "produces_summary_targets": ["alpha"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "beta", "weight_ms": 20, "needs": [], "produces_summary_targets": ["beta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "gamma", "weight_ms": 10, "needs": [], "produces_summary_targets": ["gamma", "external-summary"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "delta", "weight_ms": 5, "needs": ["beta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" }
      ]
    }
  ]
}
JSON
set +e
failure_output="$(
  FAKE_FAIL_TARGET=beta
  FAKE_SLEEP_ALPHA=0.2
  FAKE_SLEEP_BETA=0.01
  run_scheduler "$failure_dir" "$failure_manifest" failure --resource-limit host_cpu=2 2>&1
)"
failure_status=$?
set -e
assert_equals "$failure_status" "1" "failure exit status"
assert_contains "$failure_output" "fake failure for beta" "failure child output"
assert_contains "$failure_output" "[FAIL] target=check" "failure summary"
assert_occurrences "$failure_output" "[FAIL] target=check" "1" "failure single check failure block"
assert_contains "$failure_output" "[SUMMARY] target=check status=fail" "failure scheduler status summary"
assert_contains "$failure_output" "failure_class=harness" "failure scheduler class output"
assert_contains "$failure_output" " total_wall_time=" "failure scheduler wall time output"
assert_contains "$failure_output" "failed=beta" "failure scheduler failed work unit"
failure_events="$(cat "${failure_dir}/events.log")"
assert_contains "$failure_events" "start alpha" "failure alpha started"
assert_contains "$failure_events" "start beta" "failure beta started"
assert_contains "$failure_events" "end alpha" "failure alpha drained"
failure_summary="${failure_dir}/results/failure/run-summary.json"
failure_target_summary="${failure_dir}/results/failure/check/target-summary.json"
assert_equals "$(json_field "$failure_summary" "status")" "fail" "failure summary status"
assert_file_present "$failure_target_summary" "failure check target summary"
assert_equals "$(json_field "$failure_target_summary" "target")" "check" "failure check target summary identity"
assert_equals "$(json_field "$failure_target_summary" "status")" "fail" "failure check target summary status"
assert_equals "$(json_field "$failure_summary" "failure_class")" "harness" "failure summary class"
assert_equals "$(json_field "$failure_summary" "work_units.aborted_after")" "beta" "failure aborted after"
assert_equals "$(json_field "$failure_summary" "summary_targets.skipped_after_failure.0")" "gamma" "failure skipped target"
assert_equals "$(json_field "$failure_summary" "summary_targets.skipped_after_failure.1")" "external-summary" "failure skipped mapped summary target"
"$NODE_BIN" - "$failure_summary" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.summary_targets.missing.includes("external-summary")) {
  throw new Error("mapped skipped summary target must not be reported missing");
}
EOF
failure_scheduler_summary="${failure_dir}/results/failure/check/scheduler-summary.json"
assert_equals "$(json_field "$failure_scheduler_summary" "failure_class")" "harness" "failure scheduler summary class"
FAILURE_OUTPUT="$failure_output" "$NODE_BIN" - "$failure_scheduler_summary" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const output = process.env.FAILURE_OUTPUT ?? "";
const match = output.match(/\[SUMMARY\] target=check status=fail failure_class=harness reason=unknown_failure work_units=\d+\/4 total_wall_time=([0-9]+\.[0-9]{2}s) failed=beta/);
if (!match) {
  throw new Error("failure scheduler summary must include total_wall_time before failed");
}
const reportedMs = Math.round(Number(match[1].slice(0, -1)) * 1000);
if (Math.abs(reportedMs - summary.scheduler_total_duration_ms) > 10) {
  throw new Error(`failure total_wall_time ${reportedMs}ms does not match scheduler_total_duration_ms ${summary.scheduler_total_duration_ms}ms`);
}
EOF
failure_scheduler_events="${failure_dir}/results/failure/check/scheduler-events.jsonl"
assert_check_scheduler_artifacts "$failure_dir" failure check fail beta 4 skip
"$NODE_BIN" - "$failure_scheduler_summary" "$failure_scheduler_events" "$ROOT_DIR" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [summaryFile, eventsFile, repoRoot] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.failed_work_unit_detail?.aggregate_target !== "beta") {
  throw new Error(`failed work aggregate target got ${summary.failed_work_unit_detail?.aggregate_target}`);
}
if (summary.failed_work_unit_detail?.label !== "beta") {
  throw new Error(`failed work label got ${summary.failed_work_unit_detail?.label}`);
}
const skipped = new Map(summary.skipped_work_units.map((unit) => [unit.id, unit]));
if (skipped.get("delta")?.reason !== "dependency_failure") {
  throw new Error("delta must be marked skipped by dependency failure");
}
const gamma = skipped.get("gamma");
if (gamma && gamma.reason !== "schedule_stopped_after_failure") {
  throw new Error(`gamma skip reason got ${gamma.reason} want schedule_stopped_after_failure`);
}
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).map((line) => JSON.parse(line));
if (!events.some((event) => event.event === "skip" && event.work_unit === "delta" && event.skip_reason === "dependency_failure")) {
  throw new Error("scheduler events must record dependency-failure skips");
}
const slowestByLabel = new Map(summary.slowest_work_units.map((unit) => [unit.label, unit]));
for (const [label, expectedText] of [
  ["alpha", "fake pass for alpha"],
  ["beta", "fake failure for beta"],
]) {
  const logFile = slowestByLabel.get(label)?.log_file;
  if (!logFile) {
    throw new Error(`failure summary must preserve ${label} work-unit log`);
  }
  if (path.isAbsolute(logFile) || logFile.includes("cartulary-check-schedule-")) {
    throw new Error(`${label} log path must be persisted and repo-relative, got ${logFile}`);
  }
  const contents = fs.readFileSync(path.resolve(repoRoot, logFile), "utf8");
  if (!contents.includes(expectedText)) {
    throw new Error(`${label} log must remain readable after scheduler exit`);
  }
}
EOF

classified_failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-classified-failure.XXXXXX")"
cleanup_paths+=("$classified_failure_dir")
write_fake_make "$classified_failure_dir"
classified_failure_manifest="${classified_failure_dir}/manifest.json"
cat >"$classified_failure_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 2 },
      "work_units": [
        { "target": "alpha", "weight_ms": 30, "needs": [], "produces_summary_targets": ["alpha"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "beta", "weight_ms": 20, "needs": [], "produces_summary_targets": ["beta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "gamma", "weight_ms": 10, "needs": ["beta"], "produces_summary_targets": ["gamma"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" }
      ]
    }
  ]
}
JSON
set +e
classified_failure_output="$(
  FAKE_FAIL_TARGET=beta
  FAKE_FAIL_WITH_SUMMARY_TARGET=beta
  FAKE_SLEEP_ALPHA=0.08
  FAKE_SLEEP_BETA=0.01
  run_scheduler "$classified_failure_dir" "$classified_failure_manifest" classified-failure --resource-limit host_cpu=2 2>&1
)"
classified_failure_status=$?
set -e
assert_equals "$classified_failure_status" "10" "classified child failure exit status"
assert_contains "$classified_failure_output" "[SUMMARY] target=check status=fail failure_class=product reason=test_assertion_failure" "classified child scheduler output"
assert_contains "$classified_failure_output" "failed=beta" "classified child failed work unit"
classified_failure_scheduler_summary="${classified_failure_dir}/results/classified-failure/check/scheduler-summary.json"
classified_failure_target_summary="${classified_failure_dir}/results/classified-failure/check/target-summary.json"
classified_failure_tool_summary="${classified_failure_dir}/results/classified-failure/check/tool-run-summary.json"
assert_equals "$(json_field "$classified_failure_scheduler_summary" "failure_class")" "product" "classified scheduler summary class"
assert_equals "$(json_field "$classified_failure_scheduler_summary" "failure_reason")" "test_assertion_failure" "classified scheduler summary reason"
assert_equals "$(json_field "$classified_failure_scheduler_summary" "failure_headline")" "product reason=test_assertion_failure failure: synthetic product failure" "classified scheduler summary headline"
assert_equals "$(json_field "$classified_failure_scheduler_summary" "failed_work_unit")" "beta" "classified scheduler failed work unit"
assert_equals "$(json_field "$classified_failure_target_summary" "failure_class")" "product" "classified target summary class"
assert_equals "$(json_field "$classified_failure_target_summary" "failure_reason")" "test_assertion_failure" "classified target summary reason"
assert_equals "$(json_field "$classified_failure_target_summary" "failure_headline")" "product reason=test_assertion_failure failure: synthetic product failure" "classified target summary headline"
assert_equals "$(json_field "$classified_failure_tool_summary" "failure_class")" "product" "classified tool summary class"
assert_equals "$(json_field "$classified_failure_tool_summary" "failure_reason")" "test_assertion_failure" "classified tool summary reason"

service_skip_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-service-skip.XXXXXX")"
cleanup_paths+=("$service_skip_dir")
write_fake_make "$service_skip_dir"
service_skip_manifest="${service_skip_dir}/manifest.json"
cat >"$service_skip_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1, "host_io": 1, "suite_service_stack": 1 },
      "summary_groups": [
        { "name": "check-work", "summary_targets": ["lint-biome", "check-service-backed"] }
      ],
      "work_units": [
        { "target": "setup", "weight_ms": 110, "needs": [], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "lint-biome", "weight_ms": 100, "needs": ["setup"], "produces_summary_targets": ["lint-biome"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "backend-store", "weight_ms": 80, "needs": ["setup"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "service_session": { "target": "check-service-backed" } },
        { "target": "backend-integration", "weight_ms": 70, "needs": ["setup"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "service_session": { "target": "check-service-backed" } },
        { "target": "backend-integration-support", "weight_ms": 60, "needs": ["setup"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "service_session": { "target": "check-service-backed" } },
        { "target": "backend-process", "weight_ms": 50, "needs": ["setup"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "service_session": { "target": "check-service-backed" } },
        { "target": "browser-e2e-webserver-backed", "weight_ms": 40, "needs": ["setup"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "service_session": { "target": "check-service-backed" } },
        { "target": "browser-e2e-stateful", "weight_ms": 30, "needs": ["setup"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "service_session": { "target": "check-service-backed" } },
        { "target": "browser-e2e-measurement", "weight_ms": 20, "needs": ["setup"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "service_session": { "target": "check-service-backed" } },
        { "target": "browser-e2e-visual", "weight_ms": 10, "needs": ["setup"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "service_session": { "target": "check-service-backed" } },
        { "target": "browser-e2e-a11y", "weight_ms": 10, "needs": ["setup"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu", "service_session": { "target": "check-service-backed" } },
        {
          "target": "check-service-backed",
          "weight_ms": 1,
          "needs": [
            "backend-store",
            "backend-integration",
            "backend-integration-support",
            "backend-process",
            "browser-e2e-webserver-backed",
            "browser-e2e-stateful",
            "browser-e2e-measurement",
            "browser-e2e-visual",
            "browser-e2e-a11y"
          ],
          "produces_summary_targets": ["check-service-backed"],
          "resource_claims": {},
          "service_session": { "target": "check-service-backed" }
        }
      ]
    }
  ]
}
JSON
set +e
service_skip_output="$(
  FAKE_FAIL_TARGET=lint-biome \
    run_scheduler "$service_skip_dir" "$service_skip_manifest" service-skip 2>&1
)"
service_skip_status=$?
set -e
assert_equals "$service_skip_status" "1" "service skip exit status"
assert_contains "$service_skip_output" "[FAIL] target=check" "service skip check failure"
assert_not_contains "$service_skip_output" "missing child target summary: backend-store" "service skip avoids backend-store missing artifact"
service_skip_summary="${service_skip_dir}/results/service-skip/check-service-backed/target-summary.json"
assert_file_present "$service_skip_summary" "service skip check-service-backed summary"
assert_equals "$(json_field "$service_skip_summary" "status")" "fail" "service skip aggregate status"
assert_equals "$(json_field "$service_skip_summary" "children.skipped.0.target")" "backend-integration" "service skip first skipped child target"
assert_equals "$(json_field "$service_skip_summary" "children.skipped.0.reason")" "schedule_stopped_after_failure" "service skip child reason"
assert_equals "$(json_field "$service_skip_summary" "children.skipped.0.failed_dependency")" "lint-biome" "service skip child failed dependency"
assert_equals "$(json_field "$service_skip_summary" "children.missing.length")" "0" "service skip children not missing"
assert_equals "$(json_field "$service_skip_summary" "own.counts.non_test_failed")" "0" "service skip aggregate avoids artifact failure"

service_no_lease_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-service-no-lease.XXXXXX")"
cleanup_paths+=("$service_no_lease_dir")
write_fake_make "$service_no_lease_dir"
cat >"${service_no_lease_dir}/fake-test-services" <<EOF
#!/usr/bin/env bash
set -euo pipefail

log_file="${service_no_lease_dir}/test-services.log"
mode="\${1:-}"
shift || true

case "\$mode" in
  start-suite)
    printf 'test-services start-suite\n' >>"\$log_file"
    echo "fake test-services start failure before lease" >&2
    exit 9
    ;;
  terminate-suite)
    printf 'test-services terminate-suite\n' >>"\$log_file"
    echo "terminate-suite should not run without a lease" >&2
    exit 11
    ;;
  record-lifecycle)
    printf 'test-services record-lifecycle\n' >>"\$log_file"
    ;;
  *)
    echo "unexpected fake test-services mode \${mode}" >&2
    exit 2
    ;;
esac
EOF
chmod +x "${service_no_lease_dir}/fake-test-services"
service_no_lease_manifest="${service_no_lease_dir}/manifest.json"
cat >"$service_no_lease_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1, "host_io": 1, "suite_service_stack": 1 },
      "summary_groups": [
        { "name": "check-work", "summary_targets": ["service-no-lease-suite"] }
      ],
      "work_units": [
        {
          "id": "service-no-lease-suite:service-session",
          "kind": "service_session",
          "target": "service-no-lease-suite",
          "label": "service-no-lease-suite/service-session",
          "weight_ms": 10,
          "needs": [],
          "completion_keys": ["service_session:service-no-lease-suite"],
          "resource_claims": { "host_cpu": 1, "host_io": 1, "suite_service_stack": 1 },
          "retained_resource_claims": { "suite_service_stack": 1 },
          "service_session": { "target": "service-no-lease-suite" }
        },
        {
          "id": "service-no-lease-suite:complete",
          "kind": "service_complete",
          "target": "service-no-lease-suite",
          "label": "service-no-lease-suite/complete",
          "weight_ms": 1,
          "needs": ["service_session:service-no-lease-suite"],
          "completion_keys": ["service-no-lease-suite"],
          "failure_keys": ["service-no-lease-suite"],
          "produces_summary_targets": ["service-no-lease-suite"],
          "count_in_total": false,
          "counts_started": false,
          "resource_claims": {},
          "service_session": { "target": "service-no-lease-suite" }
        }
      ]
    }
  ]
}
JSON
set +e
service_no_lease_output="$(
  TEST_SERVICES_BIN="${service_no_lease_dir}/fake-test-services" \
    run_scheduler "$service_no_lease_dir" "$service_no_lease_manifest" service-no-lease 2>&1
)"
service_no_lease_status=$?
set -e
assert_equals "$service_no_lease_status" "2" "service no-lease startup failure status"
assert_contains "$service_no_lease_output" "fake test-services start failure before lease" "service no-lease preserves startup failure"
assert_not_contains "$service_no_lease_output" "terminate-suite should not run without a lease" "service no-lease cleanup skips missing lease"
assert_not_contains "$(cat "${service_no_lease_dir}/test-services.log")" "test-services terminate-suite" "service no-lease does not invoke terminate-suite"
service_no_lease_scheduler_summary="${service_no_lease_dir}/results/service-no-lease/check/scheduler-summary.json"
assert_file_present "$service_no_lease_scheduler_summary" "service no-lease scheduler summary"
assert_equals "$("$NODE_BIN" -e 'const fs=require("node:fs"); const summary=JSON.parse(fs.readFileSync(process.argv[1],"utf8")); console.log(summary.service_sessions?.[0]?.cleanup_status ?? "missing");' "$service_no_lease_scheduler_summary")" "skipped_no_lease" "service no-lease cleanup status"

invalid_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-invalid.XXXXXX")"
cleanup_paths+=("$invalid_dir")
write_fake_make "$invalid_dir"
invalid_manifest="${invalid_dir}/manifest.json"
cat >"$invalid_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1 },
      "work_units": [
        { "target": "alpha", "weight_ms": 1, "needs": ["missing"], "resource_claims": { "host_cpu": 1 } }
      ]
    }
  ]
}
JSON
set +e
invalid_output="$(run_scheduler "$invalid_dir" "$invalid_manifest" invalid 2>&1)"
invalid_status=$?
set -e
assert_equals "$invalid_status" "2" "invalid dependency status"
assert_contains "$invalid_output" "depends on unknown completion key missing" "invalid dependency output"

invalid_env_manifest="${invalid_dir}/invalid-env-manifest.json"
cat >"$invalid_env_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1 },
      "work_units": [
        { "target": "alpha", "weight_ms": 1, "env": { "CARTULARY_TEST_TARGET": "beta" }, "resource_claims": { "host_cpu": 1 } }
      ]
    }
  ]
}
JSON
set +e
run_scheduler "$invalid_dir" "$invalid_env_manifest" invalid-env >/dev/null 2>&1
invalid_env_status=$?
set -e
assert_equals "$invalid_env_status" "2" "invalid env status"

invalid_retained_manifest="${invalid_dir}/invalid-retained-manifest.json"
cat >"$invalid_retained_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1, "host_io": 1 },
      "work_units": [
        {
          "target": "service",
          "weight_ms": 1,
          "needs": [],
          "resource_claims": { "host_cpu": 1 },
          "retained_resource_claims": { "host_io": 1 }
        }
      ]
    }
  ]
}
JSON
set +e
run_scheduler "$invalid_dir" "$invalid_retained_manifest" invalid-retained >/dev/null 2>&1
invalid_retained_status=$?
set -e
assert_equals "$invalid_retained_status" "2" "invalid retained resource status"

invalid_bounded_manifest="${invalid_dir}/invalid-bounded-manifest.json"
cat >"$invalid_bounded_manifest" <<'JSON'
{
  "schema_id": "cartulary.scheduler_manifest.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 2 },
      "work_units": [
        {
          "target": "alpha",
          "weight_ms": 1,
          "needs": [],
          "resource_claims": {
            "host_cpu": { "mode": "bounded_limit", "reserve": 1, "min": 1, "max": 2, "legacy": true }
          }
        }
      ]
    }
  ]
}
JSON
set +e
run_scheduler "$invalid_dir" "$invalid_bounded_manifest" invalid-bounded >/dev/null 2>&1
invalid_bounded_status=$?
set -e
assert_equals "$invalid_bounded_status" "2" "invalid bounded claim status"

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-dry-run.XXXXXX")"
cleanup_paths+=("$dry_run_dir")
write_fake_make "$dry_run_dir"
dry_run_output="$(
  MAKEFLAGS=n \
    run_scheduler "$dry_run_dir" "$success_manifest" dry-run --resource-limit host_cpu=2 --resource-limit host_io=3 2>&1
)"
assert_contains "$dry_run_output" "[DRY-RUN] check manifest=" "dry-run output"
assert_contains "$dry_run_output" "resource_limits={host_cpu:2,host_io:3" "dry-run compact resource summary"
assert_contains "$dry_run_output" "work_units=5" "dry-run compact work-unit summary"
assert_not_contains "$dry_run_output" "[DRY-RUN] check unit" "dry-run default hides unit expansion"
assert_not_contains "$dry_run_output" "claims={" "dry-run default hides raw claims"
assert_file_absent "${dry_run_dir}/make-args.log" "dry-run child make"

dry_run_verbose_output="$(
  MAKEFLAGS=n VERBOSE=1 \
    run_scheduler "$dry_run_dir" "$success_manifest" dry-run-verbose --resource-limit host_cpu=2 --resource-limit host_io=3 2>&1
)"
assert_contains "$dry_run_verbose_output" "[DRY-RUN] check unit setup needs=none claims={host_cpu:1} make_jobs=1" "verbose dry-run includes unit claims"
assert_not_contains "$dry_run_verbose_output" "nested_scheduler=" "verbose dry-run omits obsolete nested scheduler metadata"
