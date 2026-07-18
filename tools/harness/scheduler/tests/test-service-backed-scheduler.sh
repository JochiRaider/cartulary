#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../../.." && pwd)"
SCRIPT="${ROOT_DIR}/tools/harness/scheduler/service-backed-schedule-cli.mjs"
TEST_OUTPUT_SCRIPT="${ROOT_DIR}/tools/harness/output/test-output.sh"
NODE_BIN="${NODE_BIN:-node}"
cleanup_paths=()
SUITE="${1:-all}"
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "${ROOT_DIR}/tools/harness/test-support/harness-scratch.sh"

case "$SUITE" in
  all | smoke | fast | matrix) ;;
  *)
    echo "usage: test-service-backed-scheduler.sh [smoke|fast|matrix]" >&2
    exit 2
    ;;
esac

unset VERBOSE CI_VERBOSE CARTULARY_OUTPUT_MODE CARTULARY_SUPPRESS_CHILD_SUCCESS

cleanup() {
  if [[ "${KEEP_TEST_TMP:-0}" == "1" ]]; then
    printf 'keeping test tmp paths: %s\n' "${cleanup_paths[*]}" >&2
    return
  fi
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

assert_file_present() {
  local path="$1"
  local label="$2"

  if [[ ! -f "$path" ]]; then
    fail "$label: expected $path to exist"
  fi
}

assert_file_absent() {
  local path="$1"
  local label="$2"

  if [[ -e "$path" ]]; then
    fail "$label: expected $path to be absent"
  fi
}

assert_scheduler_artifacts() {
  local dir="$1"
  local run_id="$2"
  local target="$3"
  local expected_status="$4"
  local expected_blocked="$5"
  local expected_event="$6"
  local expected_failure_class="${7:-harness}"
  local expected_failure_reason="${8:-}"
  local summary_file="${dir}/results/${run_id}/${target}/scheduler-summary.json"
  local events_file="${dir}/results/${run_id}/${target}/scheduler-events.jsonl"
  local progress_file="${dir}/results/${run_id}/${target}/progress-summary.log"

  assert_file_present "$summary_file" "$target scheduler summary"
  assert_file_present "$events_file" "$target scheduler events"
  assert_file_present "$progress_file" "$target progress summary"
  "$NODE_BIN" - "$summary_file" "$events_file" "$progress_file" "$expected_status" "$expected_blocked" "$expected_event" "$ROOT_DIR" "$expected_failure_class" "$expected_failure_reason" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [summaryFile, eventsFile, progressFile, expectedStatus, expectedBlocked, expectedEvent, repoRoot, expectedFailureClass, expectedFailureReason] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).filter(Boolean).map((line) => JSON.parse(line));
const progressLog = fs.readFileSync(progressFile, "utf8");
const summaryDir = path.dirname(summaryFile);
const resolveArtifact = (artifactPath) =>
  path.isAbsolute(artifactPath)
    ? artifactPath
    : path.resolve(repoRoot, artifactPath);
const assertRepoRelativeArtifact = (artifactPath, label) => {
  if (!artifactPath || typeof artifactPath !== "string") {
    throw new Error(`${label} must be a non-empty string`);
  }
  if (path.isAbsolute(artifactPath)) {
    const relativeToSummary = path.relative(summaryDir, artifactPath);
    if (relativeToSummary.startsWith("..") || path.isAbsolute(relativeToSummary)) {
      throw new Error(`${label} absolute path must stay under the retained scheduler directory, got ${artifactPath}`);
    }
  }
};
if (summary.schema_id !== "cartulary.service_backed_scheduler_summary.v10") {
  throw new Error(`unexpected summary schema ${summary.schema_id}`);
}
if (summary.scheduler_kind !== "service_backed") {
  throw new Error(`summary scheduler_kind got ${summary.scheduler_kind} want service_backed`);
}
if (summary.status !== expectedStatus) {
  throw new Error(`summary status got ${summary.status} want ${expectedStatus}`);
}
if (!Array.isArray(summary.observed_failed_work_units)) {
  throw new Error("summary must record observed_failed_work_units");
}
if (expectedStatus === "pass" && summary.observed_failed_work_units.length !== 0) {
  throw new Error("passing summary must not record observed failed work units");
}
if (expectedStatus === "fail" && summary.failure_class !== expectedFailureClass) {
  throw new Error(`summary failure_class got ${summary.failure_class} want ${expectedFailureClass}`);
}
if (expectedStatus === "fail" && expectedFailureReason && summary.failure_reason !== expectedFailureReason) {
  throw new Error(`summary failure_reason got ${summary.failure_reason} want ${expectedFailureReason}`);
}
if (expectedStatus === "pass" && summary.failure_class !== null) {
  throw new Error(`passing summary failure_class got ${summary.failure_class}`);
}
if (!Array.isArray(summary.slowest_work_units) || summary.slowest_work_units.length === 0) {
  throw new Error("summary must record slowest work");
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
if (!summary.resource_limits || Object.keys(summary.resource_limits).length === 0) {
  throw new Error("summary must record resource limits");
}
if (!Number.isInteger(summary.max_running_work_units) || summary.max_running_work_units < 1) {
  throw new Error(`summary max_running_work_units got ${summary.max_running_work_units}`);
}
if (!summary.artifacts?.events_jsonl || !summary.artifacts?.scheduler_logs_dir || !summary.artifacts?.progress_summary_log) {
  throw new Error("summary must record scheduler artifact paths");
}
assertRepoRelativeArtifact(summary.artifacts.progress_summary_log, "progress_summary_log");
if (!fs.statSync(resolveArtifact(summary.artifacts.progress_summary_log)).isFile()) {
  throw new Error(`progress summary artifact path must be an existing file: ${summary.artifacts.progress_summary_log}`);
}
if (!progressLog.includes(`[SCHEDULER] ${summary.target} start `)) {
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
if (!Array.isArray(summary.nested_scheduler_limits) || summary.nested_scheduler_limits.length !== 0) {
  throw new Error("service-backed summary must record empty nested scheduler limits");
}
if (!Array.isArray(summary.nested_scheduler_observations) || summary.nested_scheduler_observations.length !== 0) {
  throw new Error("service-backed summary must record empty nested scheduler observations");
}
if (!Number.isInteger(summary.finalizer_count) || !Number.isInteger(summary.finalizer_failures)) {
  throw new Error("summary must record finalizer counts");
}
if (!Array.isArray(summary.finalizer_timings)) {
  throw new Error("summary must record finalizer timings");
}
if (expectedBlocked !== "-") {
  if (!summary.blocked_resources_seen.includes(expectedBlocked)) {
    throw new Error(`summary missing blocked resource ${expectedBlocked}`);
  }
  if (!summary.blocked_explanations_seen.includes(expectedBlocked)) {
    throw new Error(`summary missing blocked explanation ${expectedBlocked}`);
  }
  if (!summary.top_blockers.some((entry) => entry.kind === "resource" && entry.name === expectedBlocked && entry.count > 0)) {
    throw new Error(`summary missing top resource blocker ${expectedBlocked}`);
  }
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
  if (progressEvents.some((event) => event.slowest_running) && summary.slowest_running_observations.length === 0) {
    throw new Error("summary must retain slowest running observations");
  }
}
if (!events.some((event) => event.resource_limits && Object.keys(event.resource_limits).length > 0)) {
  throw new Error("events must include resource limits");
}
if (!events.some((event) => event.resource_claims && Object.keys(event.resource_claims).length > 0)) {
  throw new Error("events must include resource claims");
}
const progress = events.find((event) => event.event === "progress");
if (progress) {
  if (!progress.active_groups || !Array.isArray(progress.blocked_by) || !Object.hasOwn(progress, "unblocks_after") || !Object.hasOwn(progress, "slowest_running")) {
    throw new Error("progress events must include structured v3 progress fields");
  }
  if (!Number.isInteger(progress.total_work_units) || !Number.isInteger(progress.blocked)) {
    throw new Error("progress events must include total_work_units and blocked counts");
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

write_fake_make() {
  local dir="$1"

  cat >"${dir}/fake-make" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

target="${@: -1}"
lock_file="${FAKE_SCHEDULER_LOCK:?}"
active_file="${FAKE_SCHEDULER_ACTIVE:?}"
max_file="${FAKE_SCHEDULER_MAX:?}"
log_file="${FAKE_SCHEDULER_LOG:?}"

mkdir -p "$(dirname "$active_file")"
touch "$active_file" "$max_file" "$log_file"

log_line() {
  {
    flock 8
    printf '%s\n' "$*" >>"$log_file"
  } 8>"$lock_file"
}

change_active() {
  local delta="$1"
  local active max
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
    printf '%s %s active=%s\n' "$([[ "$delta" -gt 0 ]] && printf start || printf end)" "$target" "$active" >>"$log_file"
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
    "steps": 1,
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
}

write_failure_summary() {
  if [[ -z "${CARTULARY_TEST_RESULTS_DIR:-}" || -z "${CARTULARY_TEST_RUN_ID:-}" ]]; then
    return 0
  fi
  local failure_class="${FAKE_FAIL_FAILURE_CLASS:-product}"
  local failure_reason="${FAKE_FAIL_FAILURE_REASON:-test_assertion_failure}"
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
    "steps": 1,
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
  "failure_class": "${failure_class}",
  "failure_reason": "${failure_reason}",
  "failures": [
    {
      "failure_class": "${failure_class}",
      "failure_reason": "${failure_reason}",
      "kind": "test",
      "target": "${target}",
      "label": "${target} assertion",
      "message": "fake assertion failure for ${target}"
    }
  ]
}
JSON
}

sleep_key="${target//-/_}"
sleep_key="${sleep_key^^}"
sleep_var="FAKE_SCHEDULER_SLEEP_${sleep_key}"
sleep_duration="${!sleep_var:-${FAKE_SCHEDULER_SLEEP:-0.05}}"

log_line "args $*"
log_line "env ${target} MAKEFLAGS=${MAKEFLAGS-} MFLAGS=${MFLAGS-}"
change_active 1
sleep "$sleep_duration"
change_active -1

if [[ "${FAKE_FAIL_TARGET:-}" == "$target" ]]; then
  if [[ "${FAKE_FAIL_WRITES_SUMMARY:-}" == "1" ]]; then
    write_failure_summary
  fi
  echo "fake failure for $target" >&2
  exit 7
fi

echo "fake pass for $target"
write_summary
EOF
  chmod +x "${dir}/fake-make"

  cat >"${dir}/fake-browser-session" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

lock_file="${FAKE_SCHEDULER_LOCK:?}"
active_file="${FAKE_SCHEDULER_ACTIVE:?}"
max_file="${FAKE_SCHEDULER_MAX:?}"
log_file="${FAKE_SCHEDULER_LOG:?}"
mode=""
env_file=""
lease_file=""

mkdir -p "$(dirname "$active_file")"
touch "$active_file" "$max_file" "$log_file"

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
    printf '%s browser-session %s stage=%s active=%s\n' "$action" "${CARTULARY_TEST_TARGET:-}" "${CARTULARY_BROWSER_STAGE:-}" "$active" >>"$log_file"
  } 9>"$lock_file"
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --session-start)
      mode=start
      shift
      ;;
    --session-stop)
      mode=stop
      shift
      ;;
    --env-file)
      env_file="${2:-}"
      shift 2
      ;;
    --lease-file)
      lease_file="${2:-}"
      shift 2
      ;;
    *)
      echo "unexpected fake browser session arg $1" >&2
      exit 2
      ;;
  esac
done

case "$mode" in
  start)
    if [[ -z "$env_file" || -z "$lease_file" ]]; then
      echo "fake browser session start requires env and lease files" >&2
      exit 2
    fi
    mkdir -p "$(dirname "$env_file")" "$(dirname "$lease_file")"
    change_active 1
    sleep "${FAKE_BROWSER_SESSION_SLEEP:-${FAKE_SCHEDULER_SLEEP:-0.05}}"
    cat >"$env_file" <<JSON
{
  "CARTULARY_PLAYWRIGHT_EXTERNAL_SERVER": "1",
  "CARTULARY_WEB_E2E_API_ORIGIN": "http://127.0.0.1:18080",
  "CARTULARY_WEB_E2E_PUBLIC_ORIGIN": "http://127.0.0.1:14173"
}
JSON
    cat >"$lease_file" <<JSON
{
  "schema_id": "cartulary.web_e2e_session_lease.v1",
  "target": "${CARTULARY_TEST_TARGET:-}",
  "stage": "${CARTULARY_BROWSER_STAGE:-}"
}
JSON
    change_active -1
    ;;
  stop)
    if [[ -z "$lease_file" ]]; then
      echo "fake browser session stop requires lease file" >&2
      exit 2
    fi
    {
      flock 8
      printf 'stop browser-session %s stage=%s lease=%s\n' "${CARTULARY_TEST_TARGET:-}" "${CARTULARY_BROWSER_STAGE:-}" "$lease_file" >>"$log_file"
    } 8>"$lock_file"
    rm -f "$lease_file"
    ;;
  *)
    echo "fake browser session requires --session-start or --session-stop" >&2
    exit 2
    ;;
esac
EOF
  chmod +x "${dir}/fake-browser-session"

  cat >"${dir}/fake-browser-group" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

lock_file="${FAKE_SCHEDULER_LOCK:?}"
active_file="${FAKE_SCHEDULER_ACTIVE:?}"
max_file="${FAKE_SCHEDULER_MAX:?}"
log_file="${FAKE_SCHEDULER_LOG:?}"
target="${CARTULARY_BROWSER_GROUP_TARGET:-${CARTULARY_TEST_TARGET:-browser-group}}"
sleep_key="${target//-/_}"
sleep_key="${sleep_key^^}"
sleep_var="FAKE_SCHEDULER_SLEEP_${sleep_key}"
sleep_duration="${!sleep_var:-${FAKE_SCHEDULER_SLEEP:-0.05}}"

mkdir -p "$(dirname "$active_file")"
touch "$active_file" "$max_file" "$log_file"

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
    printf '%s %s group=%s kind=%s stage=%s active=%s\n' \
      "$action" "$target" "${CARTULARY_BROWSER_GROUP_NAME:-}" "${CARTULARY_BROWSER_GROUP_KIND:-}" "${CARTULARY_BROWSER_STAGE:-}" "$active" >>"$log_file"
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
    "steps": 1,
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
}

change_active 1
sleep "$sleep_duration"
change_active -1

if [[ "${FAKE_FAIL_TARGET:-}" == "$target" ]]; then
  echo "fake browser group failure for $target" >&2
  exit 7
fi

write_summary
echo "fake browser group pass for $target"
EOF
  chmod +x "${dir}/fake-browser-group"
}

write_fake_go_target_runner() {
  local dir="$1"

  cat >"${dir}/fake-go-target" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

log_file="${FAKE_SCHEDULER_LOG:?}"
lock_file="${FAKE_SCHEDULER_LOCK:?}"

sanitize_key() {
  local value="$1"
  value="${value//[^[:alnum:]]/_}"
  printf '%s\n' "${value^^}"
}

sleep_for() {
  local prefix="$1"
  local name="$2"
  local fallback="$3"
  local key specific

  if [[ "$prefix" == "FAKE_GO_SLEEP_CAPTURE" && -n "${FAKE_GO_SLEEP_CAPTURE_SHARD:-}" && "${FAKE_GO_SLEEP_CAPTURE_SHARD}" == "$name" ]]; then
    if [[ -z "${FAKE_GO_SLEEP_CAPTURE_SHARD_DURATION:-}" ]]; then
      echo "FAKE_GO_SLEEP_CAPTURE_SHARD_DURATION is required when FAKE_GO_SLEEP_CAPTURE_SHARD matches $name" >&2
      exit 2
    fi
    printf '%s\n' "${FAKE_GO_SLEEP_CAPTURE_SHARD_DURATION}"
    return 0
  fi

  if [[ "$prefix" == "FAKE_GO_SLEEP_FINALIZE" && -n "${FAKE_GO_SLEEP_FINALIZE_TARGET:-}" && "${FAKE_GO_SLEEP_FINALIZE_TARGET}" == "$name" ]]; then
    if [[ -z "${FAKE_GO_SLEEP_FINALIZE_TARGET_DURATION:-}" ]]; then
      echo "FAKE_GO_SLEEP_FINALIZE_TARGET_DURATION is required when FAKE_GO_SLEEP_FINALIZE_TARGET matches $name" >&2
      exit 2
    fi
    printf '%s\n' "${FAKE_GO_SLEEP_FINALIZE_TARGET_DURATION}"
    return 0
  fi

  key="$(sanitize_key "$name")"
  specific="${prefix}_${key}"
  if [[ -n "${!specific:-}" ]]; then
    printf '%s\n' "${!specific}"
    return 0
  fi
  if [[ -n "${!prefix:-}" ]]; then
    printf '%s\n' "${!prefix}"
    return 0
  fi
  printf '%s\n' "$fallback"
}

log_event() {
  exec 9>"$lock_file"
  flock 9
  printf '%s\n' "$*" >>"$log_file"
}

write_summary() {
  local target="$1"
  local status="$2"

  if [[ -z "${CARTULARY_TEST_RESULTS_DIR:-}" || -z "${CARTULARY_TEST_RUN_ID:-}" ]]; then
    return 0
  fi

  mkdir -p "${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}"
  cat >"${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}/target-summary.json" <<JSON
{
  "target": "${target}",
  "status": "${status}",
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
}

case "${1:-}" in
  capture-shard)
    if [[ "$#" -ne 4 ]]; then
      echo "usage: fake-go-target capture-shard <target> <shard> <metadata-dir>" >&2
      exit 2
    fi

    target="$2"
    shard="$3"
    metadata_dir="$4"
    report_dir="${metadata_dir}/fake-reports/${shard}"
    status=0

    log_event "start capture ${target} ${shard}"
    sleep "$(sleep_for FAKE_GO_SLEEP_CAPTURE "$shard" 0.01)"
    mkdir -p "$report_dir" "${metadata_dir}/fake-failed-shards"
    : >"${report_dir}/runner.jsonl"
    : >"${report_dir}/stderr.log"
    printf 'fake go test %s %s\n' "$target" "$shard" >"${report_dir}/command.txt"
    printf '2026-01-01T00:00:00Z\n' >"${report_dir}/start_time.txt"
    printf '2026-01-01T00:00:01Z\n' >"${report_dir}/end_time.txt"
    printf '1\n' >"${report_dir}/duration_ms.txt"
    if [[ "${FAKE_GO_FAIL_SHARD:-}" == "$shard" ]]; then
      status="${FAKE_GO_FAIL_SHARD_STATUS:-5}"
      printf '%s\n' "$target" >"${metadata_dir}/fake-failed-shards/${shard}"
      printf 'fake shard failure for %s\n' "$shard" >&2
    fi
    printf '%s\n' "$status" >"${report_dir}/exit_status.txt"
    printf '%s\n%s\n' "$report_dir" actual >"${metadata_dir}/${shard}.meta"
    log_event "end capture ${target} ${shard}"
    ;;
  finalize-shards)
    if [[ "$#" -lt 3 ]]; then
      echo "usage: fake-go-target finalize-shards <target> <metadata-dir> [shard...]" >&2
      exit 2
    fi

    target="$2"
    metadata_dir="$3"
    shift 3
    shards=("$@")
    aggregate_dir="${metadata_dir}/aggregate-reports/${target}/fake-aggregate"
    status=0
    summary_status=pass
    shard_list="${shards[*]:-}"

    log_event "start finalize ${target} shards=${shard_list}"
    if [[ -n "${FAKE_GO_EXPECT_FINALIZE_TARGET:-}" && "${FAKE_GO_EXPECT_FINALIZE_TARGET}" == "$target" ]]; then
      actual_shards="$(IFS=,; printf '%s' "${shards[*]}")"
      if [[ "$actual_shards" != "${FAKE_GO_EXPECT_FINALIZE_SHARDS:-}" ]]; then
        echo "finalize shards for ${target} got [${actual_shards}] want [${FAKE_GO_EXPECT_FINALIZE_SHARDS:-}]" >&2
        exit 8
      fi
    fi
    if [[ -n "${FAKE_GO_FORBID_FINALIZE_SHARD:-}" ]]; then
      for shard in "${shards[@]}"; do
        if [[ "$shard" == "${FAKE_GO_FORBID_FINALIZE_SHARD}" ]]; then
          echo "forbidden finalize shard ${shard} for ${target}" >&2
          exit 8
        fi
      done
    fi
    sleep "$(sleep_for FAKE_GO_SLEEP_FINALIZE "$target" 0.05)"
    mkdir -p "$aggregate_dir"
    printf 'fake aggregate for %s\n' "$target" >"${aggregate_dir}/artifact.txt"
    if [[ "${FAKE_GO_FAIL_FINALIZER_TARGET:-}" == "$target" || -n "$(find "${metadata_dir}/fake-failed-shards" -type f -print -quit 2>/dev/null)" ]]; then
      status="${FAKE_GO_FINALIZER_FAILURE_STATUS:-9}"
      summary_status=fail
    fi
    write_summary "$target" "$summary_status"
    printf 'fake finalized %s status=%s aggregate_dir=%s\n' "$target" "$summary_status" "$aggregate_dir"
    log_event "end finalize ${target}"
    exit "$status"
    ;;
  *)
    echo "usage: fake-go-target <capture-shard|finalize-shards> ..." >&2
    exit 2
    ;;
esac
EOF
  chmod +x "${dir}/fake-go-target"
}

expand_source_manifest() {
  local source_file="$1"
  local output_file="$2"

  "$NODE_BIN" --input-type=module - "$ROOT_DIR" "$source_file" "$output_file" <<'EOF'
import fs from "node:fs";
import {
  expandServiceBackedSchedule,
} from "./tools/harness/execution/service-backed/schedule-planning.mjs";

const [repoRoot, sourceFile, outputFile] = process.argv.slice(2);
const sourceManifest = JSON.parse(fs.readFileSync(sourceFile, "utf8"));
const schedules = sourceManifest.schedules.map((sourceSchedule) => {
  const { work_unit_sources: _workUnitSources, ...schedule } = sourceSchedule;
  return {
    ...schedule,
    scheduler_kind: "service_backed",
    stop_on_first_failure: false,
    progress_tick_seconds: 30,
    validate_timing: true,
    work_units: expandServiceBackedSchedule({
      repoRoot,
      serviceSchedule: sourceSchedule,
    }),
    finalizers: [],
  };
});
fs.writeFileSync(
  outputFile,
  `${JSON.stringify({
    schema_id: "cartulary.scheduler_manifest.v2",
    generated: {
      generator: "tools/harness/scheduler/tests/test-service-backed-scheduler.sh",
      source: "smoke fixture",
    },
    schedules,
  }, null, 2)}\n`,
);
EOF
  rm -f "$source_file"
}

write_manifest() {
  local file="$1"
  local target="$2"
  local source_file="${file}.sources"
  shift 2

  {
    printf '{\n'
    printf '  "schema_id": "cartulary.service_backed_schedule_sources.v1",\n'
    printf '  "schedules": [\n'
    printf '    { "target": "%s", "resource_limits": { "postgres": 32, "object_store": 32, "go_cpu": 6, "go_io": 6, "postgres_clone": 8, "postgres_reset": 8, "process": 2, "browser_stack": "auto", "browser_stage_webserver_backed": 1, "browser_stage_stateful": 1, "browser_stage_measurement": 1, "browser_stage_visual": 1 }, "work_unit_sources": [\n' "$target"
    local first=1
    local source
	    for source in "$@"; do
	      IFS='|' read -r type name weight claims class browser_stage needs go_shard_mode <<<"$source"
	      class="${class:-backend}"
      if [[ "$first" -eq 0 ]]; then
        printf ',\n'
      fi
      first=0
	      if [[ "$type" == "make_target" && "$class" == "browser" ]]; then
	        local group_kind="$browser_stage"
	        local group_name="$browser_stage"
	        if [[ "$browser_stage" == "webserver-backed" ]]; then
	          group_kind="support"
	          group_name="support"
	        fi
	        printf '      { "type": "browser_stage", "class": "browser", "target": "%s", "browser_stage": "%s", "weight_ms": %s, "resource_claims": {%s}, "groups": [{ "id": "%s:%s", "name": "%s", "kind": "%s", "target": "%s", "aggregate_target": "%s", "coverage": "authoritative", "execution_dependency": "browser_%s", "weight_ms": %s, "resource_claims": { "go_cpu": 1, "go_io": 1 } }]' \
	          "$name" "$browser_stage" "$weight" "$claims" "$name" "$group_name" "$group_name" "$group_kind" "$name" "$name" "${browser_stage//-/_}" "$weight"
	        if [[ -n "${needs:-}" ]]; then
	          printf ', "needs": ['
	          local first_need=1
	          local need
	          IFS=',' read -r -a need_list <<<"$needs"
	          for need in "${need_list[@]}"; do
	            if [[ "$first_need" -eq 0 ]]; then
	              printf ', '
	            fi
	            first_need=0
	            printf '"%s"' "$need"
	          done
	          printf ']'
	        fi
	        printf ' }'
	      elif [[ "$type" == "make_target" ]]; then
	        printf '      { "type": "make_target", "class": "%s", "target": "%s", "weight_ms": %s, "resource_claims": {%s}' \
	          "$class" "$name" "$weight" "$claims"
        if [[ -n "${browser_stage:-}" ]]; then
          printf ', "browser_stage": "%s"' "$browser_stage"
        fi
        if [[ -n "${needs:-}" ]]; then
          printf ', "needs": ['
          local first_need=1
          local need
          IFS=',' read -r -a need_list <<<"$needs"
          for need in "${need_list[@]}"; do
            if [[ "$first_need" -eq 0 ]]; then
              printf ', '
            fi
            first_need=0
            printf '"%s"' "$need"
          done
          printf ']'
        fi
        printf ' }'
      else
        printf '      { "type": "go_shards", "class": "%s", "target": "%s", "resource_claims": {%s}' \
          "$class" "$name" "$claims"
        if [[ "${go_shard_mode:-}" == "default_check" ]]; then
          printf ', "default_check_required": true'
        fi
        if [[ -n "${needs:-}" ]]; then
          printf ', "needs": ['
          local first_need=1
          local need
          IFS=',' read -r -a need_list <<<"$needs"
          for need in "${need_list[@]}"; do
            if [[ "$first_need" -eq 0 ]]; then
              printf ', '
            fi
            first_need=0
            printf '"%s"' "$need"
          done
          printf ']'
        fi
        printf ' }'
      fi
    done
	    printf '\n    ] }\n'
	    printf '  ]\n'
	    printf '}\n'
	  } >"$source_file"

  expand_source_manifest "$source_file" "$file"
}

assert_no_shared_backend_integration_shards() {
  "$NODE_BIN" - "$ROOT_DIR" <<'EOF'
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const [root] = process.argv.slice(2);
const shardPlanScript = path.join(root, "tools/harness/backend/go-shard-plan.mjs");
const runPlan = (...args) =>
  execFileSync(process.execPath, [shardPlanScript, ...args], { encoding: "utf8", cwd: root });
const plan = JSON.parse(runPlan("json"));
const mixed = plan.shards.filter((candidate) => candidate.has_authoritative && candidate.has_support);
const shared = plan.shards.filter((candidate) => candidate.shared_across_targets);
if (mixed.length > 0 || shared.length > 0) {
  console.error(`expected no cross-target shared backend integration shards, got mixed=${mixed.map((candidate) => candidate.name).join(",")} shared=${shared.map((candidate) => candidate.name).join(",")}`);
  process.exit(1);
}
EOF
}

write_legacy_manifest() {
  local file="$1"
  local schema_id="$2"

  cat >"$file" <<JSON
{
  "schema_id": "${schema_id}",
  "schedules": [
    {
      "target": "test-fast-service-backed",
      "resource_limits": { "postgres": 32, "object_store": 32, "backend": 4 },
      "children": [
        { "target": "backend-integration", "kind": "backend", "weight_ms": 1, "resource_claims": ["postgres", "object_store", "backend"] }
      ]
    }
  ]
}
JSON
}

run_scheduler() {
  local dir="$1"
  local manifest="$2"
  local target="$3"
  local run_id="$4"
  shift 4
  local go_target_runner=""
  local browser_session_script=""
  local browser_group_runner=""

  if [[ -x "${dir}/fake-go-target" ]]; then
    go_target_runner="${dir}/fake-go-target"
  fi
  if [[ -x "${dir}/fake-browser-session" ]]; then
    browser_session_script="${dir}/fake-browser-session"
  fi
  if [[ -x "${dir}/fake-browser-group" ]]; then
    browser_group_runner="${dir}/fake-browser-group"
  fi

  env \
  FAKE_SCHEDULER_LOCK="${dir}/lock" \
  FAKE_SCHEDULER_ACTIVE="${dir}/active" \
  FAKE_SCHEDULER_MAX="${dir}/max" \
  FAKE_SCHEDULER_LOG="${dir}/make.log" \
  FAKE_FAIL_TARGET="${FAKE_FAIL_TARGET:-}" \
  FAKE_FAIL_WRITES_SUMMARY="${FAKE_FAIL_WRITES_SUMMARY:-}" \
  FAKE_FAIL_FAILURE_CLASS="${FAKE_FAIL_FAILURE_CLASS:-}" \
  FAKE_FAIL_FAILURE_REASON="${FAKE_FAIL_FAILURE_REASON:-}" \
  FAKE_SCHEDULER_SLEEP="${FAKE_SCHEDULER_SLEEP:-0.2}" \
  FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS="${FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS:-}" \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED="${FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED:-}" \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_STATEFUL="${FAKE_SCHEDULER_SLEEP_BROWSER_E2E_STATEFUL:-}" \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_MEASUREMENT="${FAKE_SCHEDULER_SLEEP_BROWSER_E2E_MEASUREMENT:-}" \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_VISUAL="${FAKE_SCHEDULER_SLEEP_BROWSER_E2E_VISUAL:-}" \
  FAKE_GO_SLEEP_CAPTURE="${FAKE_GO_SLEEP_CAPTURE:-}" \
  FAKE_GO_SLEEP_CAPTURE_SHARD="${FAKE_GO_SLEEP_CAPTURE_SHARD:-}" \
  FAKE_GO_SLEEP_CAPTURE_SHARD_DURATION="${FAKE_GO_SLEEP_CAPTURE_SHARD_DURATION:-}" \
  FAKE_GO_SLEEP_FINALIZE="${FAKE_GO_SLEEP_FINALIZE:-}" \
  FAKE_GO_SLEEP_FINALIZE_TARGET="${FAKE_GO_SLEEP_FINALIZE_TARGET:-}" \
  FAKE_GO_SLEEP_FINALIZE_TARGET_DURATION="${FAKE_GO_SLEEP_FINALIZE_TARGET_DURATION:-}" \
  FAKE_GO_FAIL_SHARD="${FAKE_GO_FAIL_SHARD:-}" \
  FAKE_GO_FAIL_SHARD_STATUS="${FAKE_GO_FAIL_SHARD_STATUS:-}" \
  FAKE_GO_FAIL_FINALIZER_TARGET="${FAKE_GO_FAIL_FINALIZER_TARGET:-}" \
  FAKE_GO_FINALIZER_FAILURE_STATUS="${FAKE_GO_FINALIZER_FAILURE_STATUS:-}" \
  FAKE_GO_EXPECT_FINALIZE_TARGET="${FAKE_GO_EXPECT_FINALIZE_TARGET:-}" \
  FAKE_GO_EXPECT_FINALIZE_SHARDS="${FAKE_GO_EXPECT_FINALIZE_SHARDS:-}" \
  FAKE_GO_FORBID_FINALIZE_SHARD="${FAKE_GO_FORBID_FINALIZE_SHARD:-}" \
  VERBOSE="${VERBOSE:-}" \
  CI_VERBOSE="${CI_VERBOSE:-}" \
  CARTULARY_OUTPUT_MODE="${CARTULARY_OUTPUT_MODE:-}" \
  CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT="${CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT:-}" \
  CARTULARY_SERVICE_BACKED_GO_IO_LIMIT="${CARTULARY_SERVICE_BACKED_GO_IO_LIMIT:-}" \
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT="${CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT:-}" \
  CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT="${CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT:-}" \
  CARTULARY_BROWSER_E2E_SESSION_SCRIPT="${browser_session_script}" \
  CARTULARY_BROWSER_E2E_GROUP_RUNNER="${browser_group_runner}" \
  MAKE="${dir}/fake-make" \
  CARTULARY_TEST_GO_TARGET_RUNNER="${go_target_runner}" \
  NODE_BIN="$NODE_BIN" \
  TEST_OUTPUT_SCRIPT="$TEST_OUTPUT_SCRIPT" \
  CARTULARY_TEST_RESULTS_DIR="${dir}/results" \
  CARTULARY_TEST_RUN_ID="$run_id" \
    "$NODE_BIN" "$SCRIPT" --target "$target" --manifest "$manifest" "$@"
}

if [[ "$SUITE" == "smoke" ]]; then
smoke_dir="$(cartulary_harness_mktemp_dir "service-backed-scheduler-smoke.XXXXXX")"
cleanup_paths+=("$smoke_dir")
write_fake_make "$smoke_dir"
smoke_manifest="${smoke_dir}/manifest.json"
write_manifest "$smoke_manifest" test-fast-service-backed \
  'make_target|backend-store|10|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1' \
  'make_target|backend-process|9|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend||backend-store'
smoke_output="$(FAKE_SCHEDULER_SLEEP=0.01 run_scheduler "$smoke_dir" "$smoke_manifest" test-fast-service-backed smoke 2>&1)"
assert_contains "$smoke_output" "[SCHEDULER] test-fast-service-backed start work_units=2 finalizers=0 capacity={go_cpu:6,go_io:6,browser_stack:" "smoke scheduler start"
assert_contains "$smoke_output" "[SUMMARY] target=test-fast-service-backed status=pass work_units=2/2 failed=none slowest=" "smoke scheduler pass summary"
assert_not_contains "$smoke_output" "[STEP] test-fast-service-backed" "smoke output hides per-unit steps"
assert_scheduler_artifacts "$smoke_dir" smoke test-fast-service-backed pass - start

smoke_dry_run_output="$(MAKEFLAGS=n run_scheduler "$smoke_dir" "$smoke_manifest" test-fast-service-backed smoke-dry-run 2>&1)"
assert_contains "$smoke_dry_run_output" "[DRY-RUN] test-fast-service-backed manifest=" "smoke dry-run output"
assert_contains "$smoke_dry_run_output" "work_units=2 dependencies=1 classes={backend:2} types={make_target:2}" "smoke dry-run compact summary"
assert_not_contains "$smoke_dry_run_output" "claims={" "smoke dry-run hides raw claims"
exit 0
fi

weighted_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-weighted.XXXXXX")"
cleanup_paths+=("$weighted_dir")
write_fake_make "$weighted_dir"
weighted_manifest="${weighted_dir}/manifest.json"
write_manifest "$weighted_manifest" test-fast-service-backed \
  'make_target|backend-store|1|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1' \
  'make_target|backend-process|10|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1' \
  'make_target|backend-integration-support|5|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1'
weighted_output="$(run_scheduler "$weighted_dir" "$weighted_manifest" test-fast-service-backed weighted 2>&1)"
assert_not_contains "$weighted_output" "[STEP] test-fast-service-backed" "default service scheduler output hides per-unit steps"
assert_contains "$weighted_output" "[SCHEDULER] test-fast-service-backed start work_units=3 finalizers=0 capacity={go_cpu:6,go_io:6,browser_stack:" "quiet scheduler shows aggregate start"
assert_not_contains "$weighted_output" "[SCHEDULER] test-fast-service-backed progress completed_work_units=0/3" "quiet scheduler hides immediate unblocked progress"
assert_contains "$weighted_output" "[SUMMARY] target=test-fast-service-backed status=pass work_units=3/3 failed=none slowest=" "quiet scheduler shows pass summary"
assert_not_contains "$weighted_output" "fake pass for backend-store" "quiet scheduler hides successful child logs"
assert_not_contains "$weighted_output" "active_resource_claims=" "default scheduler output hides raw active resources"
assert_not_contains "$weighted_output" "claims={" "default scheduler output hides raw claims"
assert_not_contains "$weighted_output" "running_units=" "default scheduler output hides raw running units"
assert_not_contains "$weighted_output" "blocked_resources=" "default scheduler output hides raw blocked resources"
assert_scheduler_artifacts "$weighted_dir" weighted test-fast-service-backed pass - start

weighted_verbose_output="$(VERBOSE=1 run_scheduler "$weighted_dir" "$weighted_manifest" test-fast-service-backed weighted-verbose 2>&1)"
assert_contains "$weighted_verbose_output" "[SCHEDULER] test-fast-service-backed start work_units=3 finalizers=0 capacity={go_cpu:6,go_io:6,browser_stack:" "verbose scheduler aggregate start"
assert_contains "$weighted_verbose_output" "[SCHEDULER] test-fast-service-backed start work_unit=backend-process claims={go_cpu:1,go_io:1,object_store:1,postgres:1,process:1} active=1 pending=2" "verbose scheduler start telemetry"
assert_contains "$weighted_verbose_output" "[SCHEDULER] test-fast-service-backed start work_unit=backend-store claims={go_cpu:1,go_io:1,object_store:1,postgres:1} active=3 pending=0 active_resource_claims={go_cpu:3,go_io:3,object_store:3,postgres:3,process:1}" "verbose scheduler starts all compatible weighted children"
assert_contains "$weighted_verbose_output" "resource_limits={go_cpu:6,go_io:6,browser_stack:" "verbose scheduler resource limit telemetry includes browser stack"
assert_contains "$weighted_verbose_output" "[SCHEDULER] test-fast-service-backed finish work_unit=backend-process status=0" "verbose scheduler finish telemetry"
assert_contains "$weighted_verbose_output" "fake pass for backend-store" "verbose scheduler replays successful child logs"

if [[ "$SUITE" != "fast" ]]; then
auto_capacity_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-auto-capacity.XXXXXX")"
cleanup_paths+=("$auto_capacity_dir")
write_fake_make "$auto_capacity_dir"
auto_capacity_manifest="${auto_capacity_dir}/manifest.json"
cat >"${auto_capacity_manifest}.sources" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule_sources.v1",
  "schedules": [
    {
      "target": "test-service-backed",
      "capacity_profile": "service_backed_full",
      "resource_limits": {
        "postgres": 32,
        "object_store": 32,
        "go_cpu": "auto",
        "go_io": "auto",
        "postgres_reset": "auto",
        "postgres_clone": "auto",
        "process": 6,
        "browser_stack": "auto"
      },
      "work_unit_sources": [
        { "type": "make_target", "class": "backend", "target": "backend-process", "weight_ms": 1, "resource_claims": { "postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1 } }
      ]
    }
  ]
}
JSON
expand_source_manifest "${auto_capacity_manifest}.sources" "$auto_capacity_manifest"
auto_capacity_output="$(run_scheduler "$auto_capacity_dir" "$auto_capacity_manifest" test-service-backed auto-capacity 2>&1)"
assert_contains "$auto_capacity_output" "[SCHEDULER] test-service-backed start work_units=1 finalizers=0 capacity={" "auto capacity resolves through registry policies"
"$NODE_BIN" - "${auto_capacity_dir}/results/auto-capacity/test-service-backed/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const sources = summary.resource_limit_sources ?? {};
if (sources.go_cpu !== "auto:service_backed_go_cpu") {
  throw new Error(`go_cpu auto source got ${sources.go_cpu}`);
}
if (sources.go_io !== "auto:service_backed_go_io") {
  throw new Error(`go_io auto source got ${sources.go_io}`);
}
if (sources.browser_stack !== "auto:service_backed_browser_stack") {
  throw new Error(`browser_stack auto source got ${sources.browser_stack}`);
}
if (sources.postgres_reset !== "auto:service_backed_postgres_reset") {
  throw new Error(`postgres_reset auto source got ${sources.postgres_reset}`);
}
if (sources.postgres_clone !== "auto:service_backed_postgres_clone") {
  throw new Error(`postgres_clone auto source got ${sources.postgres_clone}`);
}
if (sources.process !== "registry:service_backed_full") {
  throw new Error(`process registry source got ${sources.process}`);
}
EOF
env_capacity_output="$(
  CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT=3 \
  CARTULARY_SERVICE_BACKED_GO_IO_LIMIT=4 \
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2 \
  CARTULARY_SERVICE_BACKED_POSTGRES_RESET_LIMIT=3 \
  CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT=5 \
    run_scheduler "$auto_capacity_dir" "$auto_capacity_manifest" test-service-backed env-capacity 2>&1
)"
assert_contains "$env_capacity_output" "[SCHEDULER] test-service-backed start work_units=1 finalizers=0 capacity={go_cpu:3,go_io:4,browser_stack:2" "service-backed env capacity overrides registry policies"
"$NODE_BIN" - "${auto_capacity_dir}/results/env-capacity/test-service-backed/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const sources = summary.resource_limit_sources ?? {};
if (sources.go_cpu !== "env:CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT") {
  throw new Error(`go_cpu env source got ${sources.go_cpu}`);
}
if (sources.go_io !== "env:CARTULARY_SERVICE_BACKED_GO_IO_LIMIT") {
  throw new Error(`go_io env source got ${sources.go_io}`);
}
if (sources.browser_stack !== "env:CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT") {
  throw new Error(`browser_stack env source got ${sources.browser_stack}`);
}
if (sources.postgres_reset !== "env:CARTULARY_SERVICE_BACKED_POSTGRES_RESET_LIMIT") {
  throw new Error(`postgres_reset env source got ${sources.postgres_reset}`);
}
if (sources.postgres_clone !== "env:CARTULARY_SERVICE_BACKED_POSTGRES_CLONE_LIMIT") {
  throw new Error(`postgres_clone env source got ${sources.postgres_clone}`);
}
EOF
fi

resource_block_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-resource-block.XXXXXX")"
cleanup_paths+=("$resource_block_dir")
write_fake_make "$resource_block_dir"
resource_block_manifest="${resource_block_dir}/manifest.json"
write_manifest "$resource_block_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "object_store": 1, "go_cpu": 4, "go_io": 1' \
  'make_target|backend-store|9|"postgres": 1, "object_store": 1, "go_cpu": 3, "go_io": 1' \
  'make_target|backend-process|8|"postgres": 1, "object_store": 1, "go_cpu": 2, "go_io": 1, "process": 1'
resource_block_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=1 run_scheduler "$resource_block_dir" "$resource_block_manifest" test-fast-service-backed resource-block 2>&1)"
assert_contains "$resource_block_output" "[PROGRESS] target=test-fast-service-backed completed=0/3" "scheduler go_cpu-blocked human progress"
assert_contains "$resource_block_output" "blocker=go_cpu" "scheduler go_cpu-blocked human progress explains blocker"
assert_not_contains "$resource_block_output" "[SCHEDULER] test-fast-service-backed progress completed_work_units=" "quiet scheduler hides key/value progress"
assert_not_contains "$resource_block_output" "blocked_by=go_cpu" "quiet scheduler hides key/value blocked progress"
assert_not_contains "$resource_block_output" "unblocks_after=backend-integration" "quiet scheduler hides key/value unblock progress"
assert_not_contains "$resource_block_output" "blocked_resources=go_cpu" "default go_cpu-blocked output hides raw blocked resources"
assert_not_contains "$resource_block_output" "active_resource_claims=" "default blocked output hides raw active resources"
assert_scheduler_artifacts "$resource_block_dir" resource-block test-fast-service-backed pass go_cpu blocked

if [[ "$SUITE" != "fast" ]]; then
resource_block_machine_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-machine.XXXXXX")"
cleanup_paths+=("$resource_block_machine_dir")
write_fake_make "$resource_block_machine_dir"
CARTULARY_OUTPUT_MODE=machine \
  run_scheduler "$resource_block_machine_dir" "$resource_block_manifest" test-fast-service-backed resource-block-machine \
  >"${resource_block_machine_dir}/stdout.log" \
  2>"${resource_block_machine_dir}/stderr.log"
assert_single_machine_json \
  "${resource_block_machine_dir}/stdout.log" \
  "${resource_block_machine_dir}/stderr.log" \
  test-fast-service-backed \
  "machine service-backed scheduler summary" \
  tool_run_summary \
  target_summary \
  scheduler_summary \
  scheduler_events \
  --log \
  scheduler_progress \
  scheduler_logs

backend_capacity_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-backend-capacity.XXXXXX")"
cleanup_paths+=("$backend_capacity_dir")
write_fake_make "$backend_capacity_dir"
backend_capacity_manifest="${backend_capacity_dir}/manifest.json"
write_manifest "$backend_capacity_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1' \
  'make_target|backend-store|9|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1' \
  'make_target|backend-process|8|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1' \
  'make_target|backend-integration-support|7|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1'
backend_capacity_output="$(run_scheduler "$backend_capacity_dir" "$backend_capacity_manifest" test-fast-service-backed backend-capacity 2>&1)"
assert_contains "$backend_capacity_output" "[SUMMARY] target=test-fast-service-backed status=pass" "quiet go resource model shows success scheduler summary"

io_block_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-host_io-block.XXXXXX")"
cleanup_paths+=("$io_block_dir")
write_fake_make "$io_block_dir"
io_block_manifest="${io_block_dir}/manifest.json"
write_manifest "$io_block_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 4' \
  'make_target|backend-store|9|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 3' \
  'make_target|backend-process|8|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 2, "process": 1'
io_block_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=1 run_scheduler "$io_block_dir" "$io_block_manifest" test-fast-service-backed host_io-block 2>&1)"
assert_contains "$io_block_output" "[PROGRESS] target=test-fast-service-backed completed=0/3" "scheduler go_io-blocked human progress"
assert_contains "$io_block_output" "blocker=go_io" "scheduler go_io-blocked human progress explains blocker"
fi

browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-browser.XXXXXX")"
cleanup_paths+=("$browser_dir")
write_fake_make "$browser_dir"
browser_manifest="${browser_dir}/manifest.json"
write_manifest "$browser_manifest" test-service-backed \
  'make_target|backend-process|10|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend' \
  'make_target|browser-e2e-webserver-backed|9|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed'
browser_output="$(run_scheduler "$browser_dir" "$browser_manifest" test-service-backed browser 2>&1)"
assert_not_contains "$browser_output" "[STEP] test-service-backed" "browser schedule hides default scheduler steps"
assert_contains "$browser_output" "[RESULT] target=test-service-backed status=pass" "browser schedule aggregate child tests"
assert_contains "$browser_output" "[SCHEDULER] test-service-backed start work_units=3 finalizers=0" "browser quiet scheduler shows aggregate start"
assert_contains "$browser_output" "[SUMMARY] target=test-service-backed status=pass" "browser quiet scheduler shows success summary"
assert_not_contains "$browser_output" "claims={browser_stack:1" "browser default output hides resource claims"
assert_scheduler_artifacts "$browser_dir" browser test-service-backed pass - start

eager_finalizer_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-eager-finalizer.XXXXXX")"
cleanup_paths+=("$eager_finalizer_dir")
write_fake_make "$eager_finalizer_dir"
write_fake_go_target_runner "$eager_finalizer_dir"
eager_finalizer_manifest="${eager_finalizer_dir}/manifest.json"
cat >"${eager_finalizer_manifest}.sources" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule_sources.v1",
  "schedules": [
    {
      "target": "test-service-backed",
      "resource_limits": { "postgres": 32, "object_store": 32, "go_cpu": 64, "go_io": 64, "postgres_clone": 8, "postgres_reset": 8, "process": 2, "browser_stack": "auto", "browser_stage_webserver_backed": 1 },
      "work_unit_sources": [
        { "type": "go_shards", "class": "backend", "target": "backend-store", "resource_claims": { "postgres": 1, "object_store": 1 } },
        {
          "type": "browser_stage",
          "class": "browser",
          "target": "browser-e2e-webserver-backed",
          "browser_stage": "webserver-backed",
          "weight_ms": 9,
          "resource_claims": { "postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1 },
          "groups": [
            { "id": "browser-e2e-webserver-backed:support", "name": "support", "kind": "support", "target": "browser-e2e-webserver-backed", "aggregate_target": "browser-e2e-webserver-backed", "coverage": "authoritative", "execution_dependency": "browser_support", "weight_ms": 9, "resource_claims": { "go_cpu": 1, "go_io": 1 } }
          ]
        }
      ]
    }
  ]
}
JSON
expand_source_manifest "${eager_finalizer_manifest}.sources" "$eager_finalizer_manifest"
eager_finalizer_output="$(
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=5 \
  FAKE_GO_SLEEP_CAPTURE=0.01 \
  FAKE_GO_SLEEP_FINALIZE=0.05 \
    run_scheduler "$eager_finalizer_dir" "$eager_finalizer_manifest" test-service-backed eager-finalizer 2>&1
)"
assert_not_contains "$eager_finalizer_output" "aggregate-reports/backend-store/fake-aggregate" "quiet scheduler hides successful finalizer log output"
assert_contains "$(cat "${eager_finalizer_dir}/results/eager-finalizer/test-service-backed/scheduler-logs/finalize-backend-store.log")" "aggregate-reports/backend-store/fake-aggregate" "go finalizer uses target-scoped aggregate output"
assert_scheduler_artifacts "$eager_finalizer_dir" eager-finalizer test-service-backed pass - finalize-start
"$NODE_BIN" - "${eager_finalizer_dir}/results/eager-finalizer/test-service-backed/scheduler-events.jsonl" <<'EOF'
const fs = require("node:fs");
const [eventsFile] = process.argv.slice(2);
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).map((line) => JSON.parse(line));
const indexOf = (predicate, label) => {
  const index = events.findIndex(predicate);
  if (index === -1) {
    throw new Error(`missing ${label}`);
  }
  return index;
};
const finalizeStart = indexOf((event) => event.event === "finalize-start" && event.finalizer === "backend-store", "backend-store finalize start");
const browserEnd = indexOf((event) => event.event === "finish" && event.work_unit === "browser-e2e-webserver-backed/support", "browser group finish");
if (!(finalizeStart < browserEnd)) {
  throw new Error("backend-store finalizer waited for browser tail");
}
EOF

if [[ "$SUITE" != "fast" ]]; then
separate_finalizer_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-separate-finalizer.XXXXXX")"
cleanup_paths+=("$separate_finalizer_dir")
write_fake_make "$separate_finalizer_dir"
write_fake_go_target_runner "$separate_finalizer_dir"
separate_finalizer_manifest="${separate_finalizer_dir}/manifest.json"
write_manifest "$separate_finalizer_manifest" test-fast-service-backed \
  'go_shards|backend-integration|0|"postgres": 1, "object_store": 1' \
  'go_shards|backend-integration-support|0|"postgres": 1, "object_store": 1'
assert_no_shared_backend_integration_shards
separate_finalizer_output="$(
  FAKE_GO_SLEEP_CAPTURE=0.005 \
  FAKE_GO_SLEEP_FINALIZE=0.05 \
  FAKE_GO_SLEEP_FINALIZE_TARGET=backend-integration \
  FAKE_GO_SLEEP_FINALIZE_TARGET_DURATION=1.5 \
    run_scheduler "$separate_finalizer_dir" "$separate_finalizer_manifest" test-fast-service-backed separate-finalizer 2>&1
)"
assert_not_contains "$separate_finalizer_output" "aggregate-reports/backend-integration/fake-aggregate" "quiet scheduler hides backend-integration finalizer log output"
assert_contains "$(cat "${separate_finalizer_dir}/results/separate-finalizer/test-fast-service-backed/scheduler-logs/finalize-backend-integration.log")" "aggregate-reports/backend-integration/fake-aggregate" "backend-integration aggregate output is target-scoped"
assert_contains "$(cat "${separate_finalizer_dir}/results/separate-finalizer/test-fast-service-backed/scheduler-logs/finalize-backend-integration-support.log")" "aggregate-reports/backend-integration-support/fake-aggregate" "backend-integration-support aggregate output is target-scoped"
assert_scheduler_artifacts "$separate_finalizer_dir" separate-finalizer test-fast-service-backed pass - finalize-start
"$NODE_BIN" - "${separate_finalizer_dir}/make.log" <<'EOF'
const fs = require("node:fs");
const [logFile] = process.argv.slice(2);
const lines = fs.readFileSync(logFile, "utf8").trim().split(/\n/);
const indexOf = (needle) => {
  const index = lines.findIndex((line) => line.includes(needle));
  if (index === -1) {
    throw new Error(`missing ${needle}`);
  }
  return index;
};
indexOf("start finalize backend-integration");
indexOf("start finalize backend-integration-support");
EOF

check_browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-check-browser.XXXXXX")"
cleanup_paths+=("$check_browser_dir")
write_fake_make "$check_browser_dir"
check_browser_manifest="${check_browser_dir}/manifest.json"
write_manifest "$check_browser_manifest" check-service-backed \
  'make_target|browser-e2e-webserver-backed|30|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed' \
  'make_target|backend-process|10|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend'
check_browser_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2
  FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS=0.3
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.05
  run_scheduler "$check_browser_dir" "$check_browser_manifest" check-service-backed check-browser 2>&1
)"
assert_not_contains "$check_browser_output" "[STEP] check-service-backed" "check browser schedule hides default scheduler steps"
assert_contains "$check_browser_output" "[RESULT] target=check-service-backed status=pass" "check browser aggregate child tests"
check_browser_events="$(cat "${check_browser_dir}/make.log")"
assert_contains "$check_browser_events" "start browser-e2e-webserver-backed" "check browser webserver start"
assert_contains "$check_browser_events" "end browser-e2e-webserver-backed" "check browser webserver end"
assert_contains "$check_browser_events" "end backend-process" "check browser backend end"

dual_browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-dual-browser.XXXXXX")"
cleanup_paths+=("$dual_browser_dir")
write_fake_make "$dual_browser_dir"
dual_browser_manifest="${dual_browser_dir}/manifest.json"
write_manifest "$dual_browser_manifest" check-service-backed \
  'make_target|browser-e2e-webserver-backed|30|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed' \
  'make_target|browser-e2e-stateful|20|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_stateful": 1|browser|stateful'
dual_browser_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.2 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_STATEFUL=0.2 \
    run_scheduler "$dual_browser_dir" "$dual_browser_manifest" check-service-backed dual-browser 2>&1
)"
assert_contains "$dual_browser_output" "[RESULT] target=check-service-backed status=pass" "dual browser aggregate child tests"
assert_scheduler_artifacts "$dual_browser_dir" dual-browser check-service-backed pass - start
"$NODE_BIN" - "${dual_browser_dir}/make.log" <<'EOF'
const fs = require("node:fs");
const [logFile] = process.argv.slice(2);
const lines = fs.readFileSync(logFile, "utf8").trim().split(/\n/);
const indexOf = (needle) => {
  const index = lines.findIndex((line) => line.includes(needle));
  if (index === -1) {
    throw new Error(`missing ${needle}`);
  }
  return index;
};
const webStart = indexOf("start browser-e2e-webserver-backed");
const webEnd = indexOf("end browser-e2e-webserver-backed");
const statefulStart = indexOf("start browser-e2e-stateful");
const statefulEnd = indexOf("end browser-e2e-stateful");
if (!(webStart < statefulEnd && statefulStart < webEnd)) {
  throw new Error("distinct browser stages did not overlap when browser_stack capacity allowed it");
}
EOF

browser_auto_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-browser-auto.XXXXXX")"
cleanup_paths+=("$browser_auto_dir")
write_fake_make "$browser_auto_dir"
browser_auto_manifest="${browser_auto_dir}/manifest.json"
cat >"${browser_auto_manifest}.sources" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule_sources.v1",
  "schedules": [
    {
      "target": "check-service-backed",
      "resource_limits": {
        "postgres": 32,
        "object_store": 32,
        "go_cpu": 8,
        "go_io": 8,
        "process": 6,
        "browser_stack": "auto",
        "browser_stage_webserver_backed": 1,
        "browser_stage_stateful": 2,
        "browser_stage_measurement": 1,
        "browser_stage_visual": 1
      },
      "work_unit_sources": [
        {
          "type": "browser_stage",
          "class": "browser",
          "target": "browser-e2e-webserver-backed",
          "browser_stage": "webserver-backed",
          "weight_ms": 40,
          "resource_claims": { "postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1 },
          "groups": [
            { "id": "browser-e2e-webserver-backed:browser-functional-shard-01", "name": "browser-functional-shard-01", "kind": "functional_shard", "target": "browser-e2e-webserver-backed", "aggregate_target": "browser-e2e-webserver-backed", "coverage": "authoritative", "execution_dependency": "browser_functional", "shard_name": "browser-functional-shard-01", "shard_index": 0, "shard_count": 2, "entry_ids": ["module.auth.browser.login-shell"], "weight_ms": 41, "resource_claims": { "go_cpu": 1, "go_io": 1 } },
            { "id": "browser-e2e-webserver-backed:browser-functional-shard-02", "name": "browser-functional-shard-02", "kind": "functional_shard", "target": "browser-e2e-webserver-backed", "aggregate_target": "browser-e2e-webserver-backed", "coverage": "authoritative", "execution_dependency": "browser_functional", "shard_name": "browser-functional-shard-02", "shard_index": 1, "shard_count": 2, "entry_ids": ["module.auth.browser.invalid-login"], "weight_ms": 40, "resource_claims": { "go_cpu": 1, "go_io": 1 } },
            { "id": "browser-e2e-webserver-backed:support", "name": "support", "kind": "support", "target": "browser-e2e-webserver-backed", "aggregate_target": "browser-e2e-webserver-backed", "coverage": "authoritative", "execution_dependency": "browser_support", "weight_ms": 40, "resource_claims": { "go_cpu": 1, "go_io": 1 } }
          ]
        },
        {
          "type": "browser_stage",
          "class": "browser",
          "target": "browser-e2e-stateful",
          "browser_stage": "stateful",
          "weight_ms": 30,
          "resource_claims": { "postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_stateful": 1 },
          "groups": [
            { "id": "browser-e2e-stateful:stateful-one", "name": "stateful-one", "kind": "stateful_partition", "target": "browser-e2e-stateful", "aggregate_target": "browser-e2e-stateful", "coverage": "authoritative", "execution_dependency": "browser_stateful", "browser_session_group": "isolated-stateful-one", "browser_session_isolation_reason": "stateful partition isolation", "weight_ms": 30, "resource_claims": { "go_cpu": 1, "go_io": 1 } },
            { "id": "browser-e2e-stateful:stateful-two", "name": "stateful-two", "kind": "stateful_partition", "target": "browser-e2e-stateful", "aggregate_target": "browser-e2e-stateful", "coverage": "authoritative", "execution_dependency": "browser_stateful", "browser_session_group": "isolated-stateful-two", "browser_session_isolation_reason": "stateful partition isolation", "weight_ms": 30, "resource_claims": { "go_cpu": 1, "go_io": 1 } }
          ]
        },
        {
          "type": "browser_stage",
          "class": "browser",
          "target": "browser-e2e-measurement",
          "browser_stage": "measurement",
          "needs": ["browser-e2e-webserver-backed", "browser-e2e-stateful", "browser-e2e-visual"],
          "weight_ms": 20,
          "resource_claims": { "postgres": "limit", "object_store": "limit", "process": "limit", "browser_stack": "limit", "go_cpu": "limit", "go_io": "limit", "browser_stage_measurement": 1 },
          "groups": [
            { "id": "browser-e2e-measurement:measurement", "name": "measurement", "kind": "measurement", "target": "browser-e2e-measurement", "aggregate_target": "browser-e2e-measurement", "coverage": "authoritative", "execution_dependency": "browser_measurement", "weight_ms": 20, "resource_claims": { "go_cpu": 1, "go_io": 1 } }
          ]
        },
        {
          "type": "browser_stage",
          "class": "browser",
          "target": "browser-e2e-visual",
          "browser_stage": "visual",
          "weight_ms": 10,
          "resource_claims": { "postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_visual": 1 },
          "groups": [
            { "id": "browser-e2e-visual:visual", "name": "visual", "kind": "visual", "target": "browser-e2e-visual", "aggregate_target": "browser-e2e-visual", "coverage": "authoritative", "execution_dependency": "browser_visual", "weight_ms": 10, "resource_claims": { "go_cpu": 1, "go_io": 1 } }
          ]
        }
      ]
    }
  ]
}
JSON
expand_source_manifest "${browser_auto_manifest}.sources" "$browser_auto_manifest"
browser_auto_output="$(
  FAKE_SCHEDULER_SLEEP=0.2 \
    run_scheduler "$browser_auto_dir" "$browser_auto_manifest" check-service-backed browser-auto 2>&1
)"
assert_contains "$browser_auto_output" "browser_stack:5" "service-backed browser stack auto capacity counts isolated stateful sessions"
assert_equals "$(cat "${browser_auto_dir}/max")" "6" "service-backed browser auto capacity overlaps non-measurement browser groups while measurement stays isolated"
"$NODE_BIN" - "${browser_auto_dir}/results/browser-auto/check-service-backed/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.resource_limits?.browser_stack !== 5) {
  throw new Error(`browser_stack limit got ${summary.resource_limits?.browser_stack}`);
}
if (summary.resource_limit_sources?.browser_stack !== "auto:service_backed_browser_stack") {
  throw new Error(`browser_stack source got ${summary.resource_limit_sources?.browser_stack}`);
}
EOF
"$NODE_BIN" - "${browser_auto_dir}/make.log" <<'EOF'
const fs = require("node:fs");
const [logFile] = process.argv.slice(2);
const lines = fs.readFileSync(logFile, "utf8").trim().split(/\n/);
const groupLine = (action, group) => (line) =>
  line.startsWith(`${action} browser-e2e-webserver-backed `) && line.includes(`group=${group}`);
const functionalOneStart = lines.findIndex(groupLine("start", "browser-functional-shard-01"));
const functionalOneEnd = lines.findIndex((line) =>
  line.startsWith("end browser-e2e-webserver-backed ") && line.includes("group=browser-functional-shard-01")
);
const functionalTwoStart = lines.findIndex(groupLine("start", "browser-functional-shard-02"));
const functionalTwoEnd = lines.findIndex(groupLine("end", "browser-functional-shard-02"));
const supportStart = lines.findIndex(groupLine("start", "support"));
const supportEnd = lines.findIndex(groupLine("end", "support"));
if (
  functionalOneStart === -1 ||
  functionalOneEnd === -1 ||
  functionalTwoStart === -1 ||
  functionalTwoEnd === -1 ||
  supportStart === -1 ||
  supportEnd === -1 ||
  functionalOneStart > functionalTwoEnd ||
  functionalTwoStart > functionalOneEnd ||
  supportStart > functionalOneEnd ||
  functionalOneStart > supportEnd ||
  supportStart > functionalTwoEnd ||
  functionalTwoStart > supportEnd
) {
  throw new Error(`webserver-backed groups must overlap after the stage session, got\n${lines.join("\n")}`);
}
EOF

browser_backend_overlap_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-browser-backend-overlap.XXXXXX")"
cleanup_paths+=("$browser_backend_overlap_dir")
write_fake_make "$browser_backend_overlap_dir"
browser_backend_overlap_manifest="${browser_backend_overlap_dir}/manifest.json"
write_manifest "$browser_backend_overlap_manifest" check-service-backed \
  'make_target|backend-process|30|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend' \
  'make_target|browser-e2e-stateful|20|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_stateful": 1|browser|stateful'
browser_backend_overlap_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2 \
  FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS=0.25 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_STATEFUL=0.05 \
    run_scheduler "$browser_backend_overlap_dir" "$browser_backend_overlap_manifest" check-service-backed browser-backend-overlap 2>&1
)"
assert_contains "$browser_backend_overlap_output" "[RESULT] target=check-service-backed status=pass" "browser backend overlap aggregate child tests"
"$NODE_BIN" - "${browser_backend_overlap_dir}/make.log" <<'EOF'
const fs = require("node:fs");
const [logFile] = process.argv.slice(2);
const lines = fs.readFileSync(logFile, "utf8").trim().split(/\n/);
const indexOf = (needle) => {
  const index = lines.findIndex((line) => line.includes(needle));
  if (index === -1) {
    throw new Error(`missing ${needle}`);
  }
  return index;
};
const backendEnd = indexOf("end backend-process");
const statefulStart = indexOf("start browser-e2e-stateful");
if (!(statefulStart < backendEnd)) {
  throw new Error("browser-e2e-stateful waited for backend-process without a declared dependency");
}
EOF

scheduler_priority_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-priority.XXXXXX")"
cleanup_paths+=("$scheduler_priority_dir")
write_fake_make "$scheduler_priority_dir"
scheduler_priority_manifest="${scheduler_priority_dir}/manifest.json"
cat >"${scheduler_priority_manifest}.sources" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule_sources.v1",
  "schedules": [
    {
      "target": "check-service-backed",
      "resource_limits": { "postgres": 32, "object_store": 32, "go_cpu": 1, "go_io": 1, "process": 2 },
      "work_unit_sources": [
        { "type": "make_target", "class": "backend", "target": "backend-store", "priority": 0, "weight_ms": 100, "resource_claims": { "postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1 } },
        { "type": "make_target", "class": "backend", "target": "backend-process", "priority": 10, "weight_ms": 1, "resource_claims": { "postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1 } }
      ]
    }
  ]
}
JSON
expand_source_manifest "${scheduler_priority_manifest}.sources" "$scheduler_priority_manifest"
scheduler_priority_output="$(FAKE_SCHEDULER_SLEEP=0.01 run_scheduler "$scheduler_priority_dir" "$scheduler_priority_manifest" check-service-backed scheduler-priority 2>&1)"
assert_contains "$scheduler_priority_output" "[RESULT] target=check-service-backed status=pass" "service-backed scheduler priority pass"
"$NODE_BIN" - "${scheduler_priority_dir}/make.log" <<'EOF'
const fs = require("node:fs");
const [logFile] = process.argv.slice(2);
const lines = fs.readFileSync(logFile, "utf8").trim().split(/\n/);
const processStart = lines.findIndex((line) => line.includes("start backend-process"));
const storeStart = lines.findIndex((line) => line.includes("start backend-store"));
if (processStart === -1 || storeStart === -1 || processStart > storeStart) {
  throw new Error(`priority must outrank duration weight, got\n${lines.join("\n")}`);
}
EOF

dependency_order_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-dependency-order.XXXXXX")"
cleanup_paths+=("$dependency_order_dir")
write_fake_make "$dependency_order_dir"
dependency_order_manifest="${dependency_order_dir}/manifest.json"
write_manifest "$dependency_order_manifest" check-service-backed \
  'make_target|backend-process|30|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend||' \
  'make_target|browser-e2e-webserver-backed|20|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed|' \
  'make_target|browser-e2e-stateful|10|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_stateful": 1|browser|stateful|backend-process'
dependency_order_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2 \
  CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=1 \
  FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS=0.05 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.15 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_STATEFUL=0.01 \
    run_scheduler "$dependency_order_dir" "$dependency_order_manifest" check-service-backed dependency-order 2>&1
)"
assert_contains "$dependency_order_output" "[PROGRESS] target=check-service-backed" "dependency-blocked browser human progress"
assert_contains "$dependency_order_output" "blocker=dependencies" "dependency-blocked browser human progress explains blocker"
assert_scheduler_artifacts "$dependency_order_dir" dependency-order check-service-backed pass - blocked
"$NODE_BIN" - "${dependency_order_dir}/results/dependency-order/check-service-backed/scheduler-events.jsonl" "${dependency_order_dir}/results/dependency-order/check-service-backed/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [eventsFile, summaryFile] = process.argv.slice(2);
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).map((line) => JSON.parse(line));
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (!summary.waiting_on_seen?.includes("backend-process")) {
  throw new Error("summary must record service-backed waiting_on targets");
}
const dependencyBlocked = events.find((event) => event.event === "blocked" && event.blocked_reason === "dependencies");
if (!dependencyBlocked) {
  throw new Error("missing service-backed dependency blocked event");
}
if (!dependencyBlocked.waiting_on?.includes("backend-process")) {
  throw new Error("dependency blocked event must record direct waiting_on targets");
}
const browserBlocked = dependencyBlocked.blocked_units?.find((entry) => entry.work_unit === "browser-e2e-stateful/stage-session");
if (!browserBlocked?.waiting_on?.includes("backend-process") || browserBlocked?.waiting_on?.includes("browser-e2e-webserver-backed")) {
  throw new Error("blocked_units must record only browser-e2e-stateful backend dependency");
}
EOF
"$NODE_BIN" - "${dependency_order_dir}/make.log" <<'EOF'
const fs = require("node:fs");
const [logFile] = process.argv.slice(2);
const lines = fs.readFileSync(logFile, "utf8").trim().split(/\n/);
const indexOf = (needle) => {
  const index = lines.findIndex((line) => line.includes(needle));
  if (index === -1) {
    throw new Error(`missing ${needle}`);
  }
  return index;
};
const backendEnd = indexOf("end backend-process");
const webEnd = indexOf("end browser-e2e-webserver-backed");
const statefulStart = indexOf("start browser-e2e-stateful");
if (!(backendEnd < statefulStart)) {
  throw new Error("browser-e2e-stateful started before declared backend dependencies completed");
}
if (!(statefulStart < webEnd)) {
  throw new Error("browser-e2e-stateful incorrectly waited for browser-e2e-webserver-backed");
}
if (!lines.some((line) => line.includes("args --no-print-directory --output-sync=target -j1 backend-process"))) {
  throw new Error("service-backed make_target children must run with explicit -j1");
}
EOF

makeflags_sanitize_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-makeflags.XXXXXX")"
cleanup_paths+=("$makeflags_sanitize_dir")
write_fake_make "$makeflags_sanitize_dir"
makeflags_sanitize_manifest="${makeflags_sanitize_dir}/manifest.json"
write_manifest "$makeflags_sanitize_manifest" test-fast-service-backed \
  'make_target|backend-process|10|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1'
makeflags_sanitize_output="$(
  MAKEFLAGS='--jobserver-auth=3,4 -j --trace' \
  MFLAGS='--jobserver-fds=3,4 -j' \
    run_scheduler "$makeflags_sanitize_dir" "$makeflags_sanitize_manifest" test-fast-service-backed makeflags-sanitize 2>&1
)"
assert_contains "$makeflags_sanitize_output" "[SUMMARY] target=test-fast-service-backed status=pass" "makeflags sanitize quiet scheduler shows success summary"
assert_not_contains "$(cat "${makeflags_sanitize_dir}/make.log")" "jobserver" "child make env strips inherited jobserver tokens"
assert_not_contains "$(cat "${makeflags_sanitize_dir}/make.log")" "MFLAGS=-j" "child make env strips inherited mflags jobs"
assert_contains "$(cat "${makeflags_sanitize_dir}/make.log")" "MAKEFLAGS=--trace" "child make env preserves non-jobserver make flags"

go_dependency_order_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-go-dependency-order.XXXXXX")"
cleanup_paths+=("$go_dependency_order_dir")
write_fake_make "$go_dependency_order_dir"
write_fake_go_target_runner "$go_dependency_order_dir"
go_dependency_order_manifest="${go_dependency_order_dir}/manifest.json"
write_manifest "$go_dependency_order_manifest" check-service-backed \
  'go_shards|backend-store|0|"postgres": 1, "object_store": 1|backend||' \
  'make_target|browser-e2e-webserver-backed|10|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed|backend-store'
go_dependency_order_output="$(
  CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=1 \
  FAKE_GO_SLEEP_CAPTURE=0.005 \
  FAKE_GO_SLEEP_FINALIZE=0.15 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.01 \
    run_scheduler "$go_dependency_order_dir" "$go_dependency_order_manifest" check-service-backed go-dependency-order 2>&1
)"
assert_contains "$go_dependency_order_output" "blocker=dependencies" "go dependency waits for finalizer"
"$NODE_BIN" - "${go_dependency_order_dir}/make.log" <<'EOF'
const fs = require("node:fs");
const [logFile] = process.argv.slice(2);
const lines = fs.readFileSync(logFile, "utf8").trim().split(/\n/);
const indexOf = (needle) => {
  const index = lines.findIndex((line) => line.includes(needle));
  if (index === -1) {
    throw new Error(`missing ${needle}`);
  }
  return index;
};
const finalizeEnd = indexOf("end finalize backend-store");
const browserStart = indexOf("start browser-e2e-webserver-backed");
if (!(finalizeEnd < browserStart)) {
  throw new Error("dependent browser work started before backend-store finalizer completed");
}
EOF

dependency_failure_skip_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-dependency-failure.XXXXXX")"
cleanup_paths+=("$dependency_failure_skip_dir")
write_fake_make "$dependency_failure_skip_dir"
write_fake_go_target_runner "$dependency_failure_skip_dir"
dependency_failure_skip_manifest="${dependency_failure_skip_dir}/manifest.json"
write_manifest "$dependency_failure_skip_manifest" check-service-backed \
  'go_shards|backend-store|0|"postgres": 1, "object_store": 1|backend||' \
  'make_target|browser-e2e-webserver-backed|10|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed|backend-store'
dependency_failure_shard="$(
  "$NODE_BIN" - "$dependency_failure_skip_manifest" <<'EOF'
const fs = require("node:fs");
const manifest = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const shard = manifest.schedules?.[0]?.work_units?.find((unit) => unit.kind === "go_shard")?.shard;
if (!shard) process.exit(1);
process.stdout.write(shard);
EOF
)"
set +e
dependency_failure_skip_output="$(
  FAKE_GO_FAIL_SHARD="$dependency_failure_shard" \
  FAKE_GO_FINALIZER_FAILURE_STATUS=9 \
    run_scheduler "$dependency_failure_skip_dir" "$dependency_failure_skip_manifest" check-service-backed dependency-failure 2>&1
)"
dependency_failure_skip_status=$?
set -e
assert_equals "$dependency_failure_skip_status" "1" "dependency failure status"
assert_contains "$dependency_failure_skip_output" "skipped=" "dependency failure summary skipped dependent work"
assert_contains "$dependency_failure_skip_output" "[FAIL] target=check-service-backed" "dependency failure parent summary"
assert_occurrences "$dependency_failure_skip_output" "[FAIL] target=check-service-backed" "1" "dependency failure single parent failure block"
assert_scheduler_artifacts "$dependency_failure_skip_dir" dependency-failure check-service-backed fail - skip
"$NODE_BIN" - "${dependency_failure_skip_dir}/results/dependency-failure/check-service-backed/scheduler-events.jsonl" "${dependency_failure_skip_dir}/results/dependency-failure/check-service-backed/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [eventsFile, summaryFile] = process.argv.slice(2);
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).map((line) => JSON.parse(line));
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (!events.some((event) => event.event === "skip" && event.work_unit === "browser-e2e-webserver-backed/stage-session" && event.skip_reason === "dependency_failure" && event.failed_dependency === "backend-store")) {
  throw new Error("missing dependency failure skip event");
}
if (summary.skipped_work_units?.[0]?.failed_dependency !== "backend-store") {
  throw new Error("summary must record skipped dependency");
}
EOF

browser_stack_lane_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-browser-stack.XXXXXX")"
cleanup_paths+=("$browser_stack_lane_dir")
write_fake_make "$browser_stack_lane_dir"
browser_stack_lane_manifest="${browser_stack_lane_dir}/manifest.json"
write_manifest "$browser_stack_lane_manifest" check-service-backed \
  'make_target|browser-e2e-webserver-backed|30|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed' \
  'make_target|browser-e2e-stateful|20|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_stateful": 1|browser|stateful'
browser_stack_lane_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=1 \
  CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=1 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.05 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_STATEFUL=0.05 \
    run_scheduler "$browser_stack_lane_dir" "$browser_stack_lane_manifest" check-service-backed browser-stack-lane 2>&1
)"
assert_contains "$browser_stack_lane_output" "blocker=browser_stack" "shared browser stack blocks overlapping browser stages"

same_browser_lane_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-same-browser-lane.XXXXXX")"
cleanup_paths+=("$same_browser_lane_dir")
write_fake_make "$same_browser_lane_dir"
same_browser_lane_manifest="${same_browser_lane_dir}/manifest.json"
write_manifest "$same_browser_lane_manifest" test-fast-service-backed \
  'make_target|backend-process|30|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1, "browser_stage_stateful": 1|backend' \
  'make_target|backend-store|20|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "browser_stage_stateful": 1|backend'
same_browser_lane_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2 \
  CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=1 \
  FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS=0.05 \
  FAKE_SCHEDULER_SLEEP_BACKEND_STORE=0.05 \
    run_scheduler "$same_browser_lane_dir" "$same_browser_lane_manifest" test-fast-service-backed same-browser-lane 2>&1
)"
assert_contains "$same_browser_lane_output" "blocker=browser_stage_stateful" "same browser stage lane blocks overlapping work"
"$NODE_BIN" - "${same_browser_lane_dir}/make.log" <<'EOF'
const fs = require("node:fs");
const [logFile] = process.argv.slice(2);
const lines = fs.readFileSync(logFile, "utf8").trim().split(/\n/);
const indexOf = (needle) => {
  const index = lines.findIndex((line) => line.includes(needle));
  if (index === -1) {
    throw new Error(`missing ${needle}`);
  }
  return index;
};
const processEnd = indexOf("end backend-process");
const storeStart = indexOf("start backend-store");
if (!(processEnd < storeStart)) {
  throw new Error("same browser_stage_stateful lane work overlapped");
}
EOF

set +e
empty_budget_output="$("$NODE_BIN" "${ROOT_DIR}/tools/harness/backend/postgres-fixture-budget-cli.mjs" --targets "" 2>&1)"
empty_budget_status=$?
set -e
assert_equals "$empty_budget_status" "0" "empty postgres fixture budget target list status"
assert_equals "$empty_budget_output" "" "empty postgres fixture budget target list output"

fixture_shape_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-fixture-shape.XXXXXX")"
cleanup_paths+=("$fixture_shape_dir")
write_fake_make "$fixture_shape_dir"
fixture_shape_manifest="${fixture_shape_dir}/manifest.json"
write_manifest "$fixture_shape_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "object_store": 1'
fixture_events_dir="${fixture_shape_dir}/results/fixture-shape-fail/_shared/test-services/suite/events"
mkdir -p "$fixture_events_dir"
cat >"${fixture_events_dir}/001-postgres-db-created.json" <<'JSON'
{
  "type": "postgres-db-created",
  "kind": "template-clone",
  "timestamp": "2026-05-28T00:00:00Z",
  "name": "fixture_shape_1",
  "details": {
    "target": "backend-integration",
    "fixture_policy": "package_reset",
    "reuse_scope": "package-reused",
    "caller_package": "internal/app",
    "test_name": "SyntheticFixtureShapeFailure"
  }
}
JSON
cat >"${fixture_events_dir}/002-postgres-db-created.json" <<'JSON'
{
  "type": "postgres-db-created",
  "kind": "template-clone",
  "timestamp": "2026-05-28T00:00:01Z",
  "name": "fixture_shape_2",
  "details": {
    "target": "backend-integration",
    "fixture_policy": "package_reset",
    "reuse_scope": "package-reused",
    "caller_package": "internal/app",
    "test_name": "SyntheticFixtureShapeFailure"
  }
}
JSON
set +e
fixture_shape_output="$(
  FAKE_SCHEDULER_SLEEP=0.01 \
    run_scheduler "$fixture_shape_dir" "$fixture_shape_manifest" test-fast-service-backed fixture-shape-fail 2>&1
)"
fixture_shape_status=$?
set -e
assert_equals "$fixture_shape_status" "3" "postgres fixture shape scheduler status"
assert_contains "$fixture_shape_output" "postgres-fixture-shape" "fixture shape failure label"
assert_contains "$fixture_shape_output" "reason=fixture_error" "fixture shape failure reason"
assert_contains "$fixture_shape_output" "internal/app package database creates got 2, budget 0" "fixture shape failure package diagnostic"
assert_not_contains "$fixture_shape_output" "package reset duration got" "fixture shape failure does not enforce reset duration"

group_clone_shape_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-group-clone-shape.XXXXXX")"
cleanup_paths+=("$group_clone_shape_dir")
write_fake_make "$group_clone_shape_dir"
group_clone_shape_manifest="${group_clone_shape_dir}/manifest.json"
write_manifest "$group_clone_shape_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "object_store": 1'
group_clone_events_dir="${group_clone_shape_dir}/results/group-clone-shape-fail/_shared/test-services/suite/events"
mkdir -p "$group_clone_events_dir"
for index in $(seq 1 20); do
  event_id="$(printf '%03d' "$index")"
  cat >"${group_clone_events_dir}/${event_id}-postgres-db-created.json" <<JSON
{
  "type": "postgres-db-created",
  "kind": "template-clone",
  "timestamp": "2026-05-28T00:01:00Z",
  "name": "group_clone_shape_${event_id}",
  "details": {
    "target": "backend-integration",
    "fixture_policy": "group_clone",
    "reuse_scope": "group-reused",
    "caller_package": "internal/modules/evidence",
    "caller_file": "internal/modules/evidence/objectstore_dependency_test.go",
    "test_name": "SyntheticGroupCloneBudgetMiss/subcase_${event_id}",
    "reuse_group": "internal/modules/evidence:SyntheticGroupCloneBudgetMiss:subcase_${event_id}"
  }
}
JSON
done
set +e
group_clone_shape_output="$(
  FAKE_SCHEDULER_SLEEP=0.01 \
    run_scheduler "$group_clone_shape_dir" "$group_clone_shape_manifest" test-fast-service-backed group-clone-shape-fail 2>&1
)"
group_clone_shape_status=$?
set -e
assert_equals "$group_clone_shape_status" "3" "postgres group clone fixture shape scheduler status"
assert_contains "$group_clone_shape_output" "postgres-fixture-shape" "group clone fixture shape failure label"
assert_contains "$group_clone_shape_output" "backend-integration exceeded postgres group clone budget" "group clone fixture shape over-budget diagnostic"
assert_contains "$group_clone_shape_output" "actual_sources=" "group clone fixture shape actual sources"
assert_contains "$group_clone_shape_output" "SyntheticGroupCloneBudgetMiss" "group clone fixture shape names actual test"
assert_contains "$group_clone_shape_output" "planned_manifest_symbols=" "group clone fixture shape planned symbols"
assert_contains "$group_clone_shape_output" "TestAttachRouteContract_Integration" "group clone fixture shape names planned manifest symbol"

fi

failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-failure.XXXXXX")"
cleanup_paths+=("$failure_dir")
write_fake_make "$failure_dir"
failure_manifest="${failure_dir}/manifest.json"
write_manifest "$failure_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "object_store": 1' \
  'make_target|backend-store|9|"postgres": 1, "object_store": 1'
set +e
failure_output="$(
  FAKE_FAIL_WRITES_SUMMARY=1 \
  FAKE_FAIL_TARGET=backend-store \
    run_scheduler "$failure_dir" "$failure_manifest" test-fast-service-backed failure 2>&1
)"
failure_status=$?
set -e
assert_equals "$failure_status" "10" "child failure status"
assert_contains "$failure_output" "fake failure for backend-store" "child failure output"
assert_contains "$failure_output" "[SUMMARY] target=test-fast-service-backed status=fail failure_class=product reason=test_assertion_failure work_units=2/2 failed=backend-store" "failure scheduler summary"
assert_contains "$failure_output" "[FAIL] target=test-fast-service-backed" "failure target summary"
assert_occurrences "$failure_output" "[FAIL] target=test-fast-service-backed" "1" "failure single target failure block"
assert_contains "$failure_output" "[FAIL] target=test-fast-service-backed exit_code=10 failure_class=product" "failure target class"
assert_contains "$failure_output" "reason=" "failure target origin"
assert_contains "$failure_output" "work_unit=backend-store" "failure target work unit"
assert_contains "$failure_output" "child_target=backend-store" "failure target child"
assert_contains "$failure_output" "duration_ms=" "failure target duration"
assert_contains "$failure_output" "summary_json=test-fast-service-backed/tool-run-summary.json" "failure summary json"
assert_contains "$failure_output" "log_artifact=test-fast-service-backed/progress-summary.log" "failure log artifact"
assert_contains "$failure_output" "scheduler_json=test-fast-service-backed/scheduler-summary.json" "failure scheduler json"
assert_contains "$failure_output" "[RERUN] command=\"make test-fast-service-backed\"" "failure rerun command"
assert_contains "$failure_output" "[INVESTIGATE] command=\"make explain-target TARGET=test-fast-service-backed DETAIL=artifacts\"" "failure investigate command"
assert_scheduler_artifacts "$failure_dir" failure test-fast-service-backed fail - finish product test_assertion_failure

if [[ "$SUITE" != "fast" ]]; then
failed_shard_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-failed-shard.XXXXXX")"
cleanup_paths+=("$failed_shard_dir")
write_fake_make "$failed_shard_dir"
write_fake_go_target_runner "$failed_shard_dir"
failed_shard_manifest="${failed_shard_dir}/manifest.json"
write_manifest "$failed_shard_manifest" test-fast-service-backed \
  'go_shards|backend-store|0|"postgres": 1, "object_store": 1'
failed_shard_name="$(
  "$NODE_BIN" - "$failed_shard_manifest" <<'EOF'
const fs = require("node:fs");
const manifest = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const shard = manifest.schedules?.[0]?.work_units?.find((unit) => unit.kind === "go_shard")?.shard;
if (!shard) process.exit(1);
process.stdout.write(shard);
EOF
)"
set +e
failed_shard_output="$(
  FAKE_GO_FAIL_SHARD="$failed_shard_name" \
  FAKE_GO_FINALIZER_FAILURE_STATUS=9 \
    run_scheduler "$failed_shard_dir" "$failed_shard_manifest" test-fast-service-backed failed-shard 2>&1
)"
failed_shard_status=$?
set -e
assert_equals "$failed_shard_status" "1" "failed shard finalizer status"
assert_contains "$failed_shard_output" "fake shard failure for $failed_shard_name" "failed shard output"
assert_contains "$failed_shard_output" "[SUMMARY] target=test-fast-service-backed status=fail" "failed shard scheduler summary"
assert_contains "$failed_shard_output" "finalizer_failures=1" "failed shard scheduler finalizer failure count"
assert_contains "$failed_shard_output" "[FAIL] target=test-fast-service-backed" "failed shard parent summary"
assert_contains "$failed_shard_output" "work_unit=finalize/backend-store" "failed shard finalizer work unit"
assert_contains "$failed_shard_output" "child_target=backend-store" "failed shard finalizer child target"
assert_occurrences "$failed_shard_output" "[FAIL] target=test-fast-service-backed" "1" "failed shard single parent failure block"
assert_scheduler_artifacts "$failed_shard_dir" failed-shard test-fast-service-backed fail - finalize-finish
"$NODE_BIN" - "${failed_shard_dir}/results/failed-shard/test-fast-service-backed/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.finalizer_failures !== 1) {
  throw new Error(`expected one finalizer failure, got ${summary.finalizer_failures}`);
}
if (summary.finalizer_count !== 1) {
  throw new Error(`expected one finalizer, got ${summary.finalizer_count}`);
}
if (summary.failed_work_unit_detail?.aggregate_target !== "backend-store") {
  throw new Error(`finalizer failure aggregate target got ${summary.failed_work_unit_detail?.aggregate_target}`);
}
if (summary.failed_work_unit_detail?.label !== "finalize/backend-store") {
  throw new Error(`finalizer failure label got ${summary.failed_work_unit_detail?.label}`);
}
const timing = summary.finalizer_timings?.find((entry) => entry.id === "finalize:backend-store");
if (!timing) {
  throw new Error("expected failed backend-store finalizer timing");
}
if (timing.label !== "finalize/backend-store" || timing.status !== 9) {
  throw new Error(`unexpected finalizer timing ${JSON.stringify(timing)}`);
}
if (!Number.isInteger(timing.duration_ms) || timing.duration_ms < 0) {
  throw new Error(`finalizer duration must be non-negative integer, got ${timing.duration_ms}`);
}
if (!timing.log_file || timing.log_file.startsWith("/")) {
  throw new Error(`finalizer log path must be repo-relative, got ${timing.log_file}`);
}
EOF

selected_shard_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-selected-shards.XXXXXX")"
cleanup_paths+=("$selected_shard_dir")
write_fake_make "$selected_shard_dir"
write_fake_go_target_runner "$selected_shard_dir"
selected_shard_manifest="${selected_shard_dir}/manifest.json"
write_manifest "$selected_shard_manifest" test-fast-service-backed \
  'go_shards|backend-integration|0|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1|backend|||default_check'
selected_shards="$(
  "$NODE_BIN" - "$selected_shard_manifest" <<'EOF'
const fs = require("node:fs");
const [manifestFile] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(manifestFile, "utf8"));
const unit = manifest.schedules[0].work_units.find(
  (candidate) =>
    candidate.command?.type === "go_shard_finalize" &&
    candidate.command?.target === "backend-integration",
);
if (!unit) {
  throw new Error("missing backend-integration go_shard_finalize unit");
}
if (!Array.isArray(unit.shard_names) || unit.shard_names.length === 0) {
  throw new Error("backend-integration finalizer selected no shard_names");
}
if (unit.shard_names.includes("backend-integration-testutil-shard-01")) {
  throw new Error("default-check shard selection must omit backend-integration-testutil-shard-01");
}
process.stdout.write(unit.shard_names.join(","));
EOF
)"
selected_shard_output="$(
  FAKE_SCHEDULER_SLEEP=0.01 \
  FAKE_GO_EXPECT_FINALIZE_TARGET=backend-integration \
  FAKE_GO_EXPECT_FINALIZE_SHARDS="$selected_shards" \
  FAKE_GO_FORBID_FINALIZE_SHARD=backend-integration-testutil-shard-01 \
    run_scheduler "$selected_shard_dir" "$selected_shard_manifest" test-fast-service-backed selected-shards 2>&1
)"
assert_contains "$selected_shard_output" "[SUMMARY] target=test-fast-service-backed status=pass" "selected shard scheduler pass"
assert_contains "$(cat "${selected_shard_dir}/make.log")" "start finalize backend-integration shards=${selected_shards//,/ }" "selected shard finalizer argv"
assert_not_contains "$(cat "${selected_shard_dir}/make.log")" "backend-integration-testutil-shard-01" "selected shard finalizer omits testutil shard"

defer_summary_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-defer-summary.XXXXXX")"
cleanup_paths+=("$defer_summary_dir")
write_fake_make "$defer_summary_dir"
defer_summary_manifest="${defer_summary_dir}/manifest.json"
write_manifest "$defer_summary_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "object_store": 1' \
  'make_target|backend-store|9|"postgres": 1, "object_store": 1'
defer_summary_output="$(run_scheduler "$defer_summary_dir" "$defer_summary_manifest" test-fast-service-backed defer-summary --defer-summary 2>&1)"
assert_not_contains "$defer_summary_output" "[TARGET] start test-fast-service-backed" "defer-summary quiet output hides target start"
assert_file_absent "${defer_summary_dir}/results/defer-summary/test-fast-service-backed/target-summary.json" "defer-summary parent target summary"

unsafe_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-unsafe.XXXXXX")"
cleanup_paths+=("$unsafe_dir")
write_fake_make "$unsafe_dir"
unsafe_manifest="${unsafe_dir}/manifest.json"
write_manifest "$unsafe_manifest" check-service-backed \
  'make_target|backend-unit|10|"postgres": 1, "object_store": 1'
set +e
unsafe_output="$(run_scheduler "$unsafe_dir" "$unsafe_manifest" check-service-backed unsafe 2>&1)"
unsafe_status=$?
set -e
assert_equals "$unsafe_status" "2" "unsafe manifest status"
assert_contains "$unsafe_output" "is not service-backed" "unsafe manifest output"

unknown_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-unknown.XXXXXX")"
cleanup_paths+=("$unknown_dir")
write_fake_make "$unknown_dir"
unknown_manifest="${unknown_dir}/manifest.json"
write_manifest "$unknown_manifest" test-fast-service-backed \
  'make_target|unknown-backend-target|10|"postgres": 1, "object_store": 1'
set +e
unknown_output="$(run_scheduler "$unknown_dir" "$unknown_manifest" test-fast-service-backed unknown 2>&1)"
unknown_status=$?
set -e
assert_equals "$unknown_status" "2" "unknown manifest status"
assert_contains "$unknown_output" "is not in target-plan" "unknown manifest output"

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-dry-run.XXXXXX")"
cleanup_paths+=("$dry_run_dir")
write_fake_make "$dry_run_dir"
dry_run_manifest="${dry_run_dir}/manifest.json"
write_manifest "$dry_run_manifest" test-service-backed \
  'make_target|browser-e2e-webserver-backed|10|"postgres": 1, "object_store": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed'
dry_run_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2 \
  MAKEFLAGS=n \
    run_scheduler "$dry_run_dir" "$dry_run_manifest" test-service-backed dry-run 2>&1
)"
assert_contains "$dry_run_output" "[DRY-RUN] test-service-backed manifest=" "dry-run output"
assert_contains "$dry_run_output" "resource_limits={go_cpu:6,go_io:6,browser_stack:2,object_store:32,postgres:32,process:2" "dry-run includes compact resolved resources"
assert_contains "$dry_run_output" "work_units=2" "dry-run includes compact browser stage work summary"
assert_contains "$dry_run_output" "finalizers=0" "dry-run excludes browser stage completion from finalizers"
assert_not_contains "$dry_run_output" "claims={" "default dry-run hides per-unit claims"
assert_file_absent "${dry_run_dir}/make.log" "dry-run child make log"

go_shard_dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-go-shard-dry-run.XXXXXX")"
cleanup_paths+=("$go_shard_dry_run_dir")
write_fake_make "$go_shard_dry_run_dir"
go_shard_dry_run_manifest="${go_shard_dry_run_dir}/manifest.json"
write_manifest "$go_shard_dry_run_manifest" test-fast-service-backed \
  'go_shards|backend-store|0|"postgres": 1, "object_store": 1'
go_shard_dry_run_output="$(
  VERBOSE=1 \
  MAKEFLAGS=n \
    run_scheduler "$go_shard_dry_run_dir" "$go_shard_dry_run_manifest" test-fast-service-backed go-shard-dry-run 2>&1
)"
expected_go_shard_dry_run_line="$(
  "$NODE_BIN" - "$go_shard_dry_run_manifest" <<'EOF'
const fs = require("node:fs");
const manifest = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const shard = manifest.schedules?.[0]?.work_units?.find((unit) => unit.kind === "go_shard");
if (!shard?.shard) {
  process.exit(1);
}
const claims = Object.entries(shard.resource_claims)
  .sort(([left], [right]) => left.localeCompare(right))
  .map(([name, value]) => `${name}:${value}`)
  .join(",");
process.stdout.write(`backend-store/${shard.shard} type=go_shard class=backend profile=${shard.scheduler_profile} claims={${claims}}`);
EOF
)"
assert_contains "$go_shard_dry_run_output" "finalizers=1" "go_shards dry-run compact summary"
assert_contains "$go_shard_dry_run_output" "$expected_go_shard_dry_run_line" "verbose go_shards dry-run includes per-shard resource claims"

rendered_schedule_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-rendered.XXXXXX")"
cleanup_paths+=("$rendered_schedule_dir")
rendered_schedule_manifest="${rendered_schedule_dir}/service-backed.json"
"$NODE_BIN" "$ROOT_DIR/tools/harness/generated-artifacts/render-service-backed-schedule-manifest.mjs" --output "$rendered_schedule_manifest"
"$NODE_BIN" --input-type=module - "$ROOT_DIR" "$rendered_schedule_manifest" <<'EOF'
import fs from "node:fs";
import path from "node:path";
import { pathToFileURL } from "node:url";

const [root, manifestPath] = process.argv.slice(2);
process.chdir(root);
const { compareExecutionDependencies } = await import(
  pathToFileURL(path.join(root, "tools/harness/execution/execution-dependencies.mjs"))
);
const { collectTargetPlanRows, findTargetDescriptor } = await import(
  pathToFileURL(path.join(root, "tools/harness/backend/backend-target-plan.mjs"))
);

const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8"));
const byTarget = new Map((manifest.schedules ?? []).map((schedule) => [schedule.target, schedule]));
const sourceTargets = (target) => (byTarget.get(target)?.work_unit_sources ?? []).map((source) => source.target);
const testSources = sourceTargets("test-service-backed");
const checkSources = sourceTargets("check-service-backed");
const fastSources = sourceTargets("test-fast-service-backed");
const backendSourceTargets = (target) =>
  (byTarget.get(target)?.work_unit_sources ?? [])
    .filter((source) => source.class === "backend")
    .map((source) => source.target);
function minDependency(target) {
  const descriptor = findTargetDescriptor(target, root);
  return [
    ...(descriptor?.executionDependencies ?? []),
    ...(descriptor?.supportTargets ?? []),
  ].sort(compareExecutionDependencies)[0] ?? "";
}
const expectedBackendTargets = Array.from(
  new Set(
    collectTargetPlanRows(root)
      .filter((row) => row.runner_family === "go_test" && row.service_backed && row.check_service_backed_safe)
      .map((row) => row.target),
  ),
).sort(
  (left, right) =>
    compareExecutionDependencies(minDependency(left), minDependency(right)) ||
    left.localeCompare(right),
);
const expectedCheckBackendTargets = expectedBackendTargets.filter(
  (target) => target !== "backend-integration-support",
);
for (const target of ["test-service-backed", "test-fast-service-backed", "check-service-backed"]) {
  const actualBackendTargets = backendSourceTargets(target);
  const expectedTargets = target === "check-service-backed" ? expectedCheckBackendTargets : expectedBackendTargets;
  if (JSON.stringify(actualBackendTargets) !== JSON.stringify(expectedTargets)) {
    throw new Error(`${target} backend sources got ${actualBackendTargets} want ${expectedTargets}`);
  }
}
const expectedCheckSources = testSources.filter(
  (target) =>
    ![
      "backend-integration-support",
      "browser-e2e-measurement",
      "browser-e2e-visual",
      "browser-e2e-a11y",
    ].includes(target),
);
if (JSON.stringify(expectedCheckSources) !== JSON.stringify(checkSources)) {
  throw new Error(`check-service-backed sources got ${checkSources} want ${expectedCheckSources}`);
}
if (fastSources.some((target) => target.startsWith("browser-e2e"))) {
  throw new Error(`test-fast-service-backed must remain backend-only, got ${fastSources.join(",")}`);
}
for (const excluded of ["browser-e2e-functional", "browser-e2e-support"]) {
  if (testSources.includes(excluded)) {
    throw new Error(`untagged helper stage ${excluded} must not enter service-backed schedules`);
  }
}
if (testSources.includes("browser-e2e")) {
  throw new Error("browser-e2e aggregate must not enter service-backed schedules");
}
const expectedBrowserTargets = [
  "browser-e2e-webserver-backed",
  "browser-e2e-stateful",
  "browser-e2e-measurement",
  "browser-e2e-visual",
  "browser-e2e-a11y",
];
const actualBrowserTargets = (byTarget.get("test-service-backed")?.work_unit_sources ?? [])
  .filter((source) => source.class === "browser")
  .map((source) => source.target);
if (JSON.stringify(actualBrowserTargets) !== JSON.stringify(expectedBrowserTargets)) {
  throw new Error(`service-backed browser sources got ${JSON.stringify(actualBrowserTargets)}`);
}
const webserverSource = (byTarget.get("test-service-backed")?.work_unit_sources ?? []).find(
  (candidate) => candidate.target === "browser-e2e-webserver-backed",
);
const webserverGroups = webserverSource?.groups ?? [];
const selectedWebserverRows = webserverGroups.flatMap((group) => group.selected_row_ids ?? []);
if (
  webserverGroups.length === 0 ||
  webserverGroups.some(
    (group) =>
      group.kind !== "duration_balanced_specs" ||
      !Array.isArray(group.selected_row_ids) ||
      group.selected_row_ids.length === 0,
  ) ||
  new Set(selectedWebserverRows).size !== selectedWebserverRows.length
) {
  throw new Error("browser-e2e-webserver-backed groups must close unique semantic row selections");
}
if ((webserverSource?.groups ?? []).some((group) => group.priority !== 36000)) {
  throw new Error("browser-e2e-webserver-backed groups must carry critical-path scheduler priority");
}
for (const target of ["browser-e2e-stateful", "browser-e2e-visual"]) {
  const source = (byTarget.get("test-service-backed")?.work_unit_sources ?? []).find(
    (candidate) => candidate.target === target,
  );
  if (JSON.stringify(source?.needs ?? []) !== JSON.stringify([])) {
    throw new Error(`${target} needs got ${JSON.stringify(source?.needs ?? [])}`);
  }
}
const checkBrowserSources = (byTarget.get("check-service-backed")?.work_unit_sources ?? [])
  .filter((source) => source.class === "browser");
const sessionGroups = new Map(checkBrowserSources.map((source) => [source.browser_stage, source.browser_session_group]));
const expectedSessionGroups = new Map([
  ["webserver-backed", "default-check-browser-shared"],
  ["stateful", "default-check-stateful-isolated"],
]);
for (const [stage, expectedGroup] of expectedSessionGroups.entries()) {
  if (sessionGroups.get(stage) !== expectedGroup) {
    throw new Error(`check-service-backed ${stage} session group got ${sessionGroups.get(stage)} want ${expectedGroup}`);
  }
}
if (checkBrowserSources.some((source) => source.browser_stage === "measurement")) {
  throw new Error("check-service-backed must not include measurement browser session work");
}
const statefulSource = checkBrowserSources.find((source) => source.browser_stage === "stateful");
if (!statefulSource?.browser_session_isolation_reason) {
  throw new Error("check-service-backed isolated stateful browser session must declare an isolation reason");
}
const measurementSource = (byTarget.get("test-service-backed")?.work_unit_sources ?? []).find(
  (candidate) => candidate.target === "browser-e2e-measurement",
);
const expectedMeasurementNeeds = [
  "browser-e2e-webserver-backed",
  "browser-e2e-stateful",
  "browser-e2e-visual",
  "browser-e2e-a11y",
];
if (JSON.stringify(measurementSource?.needs ?? []) !== JSON.stringify(expectedMeasurementNeeds)) {
  throw new Error(`browser-e2e-measurement needs got ${JSON.stringify(measurementSource?.needs ?? [])}`);
}
for (const retiredNeed of ["backend-store", "backend-integration", "backend-integration-support", "backend-process"]) {
  if ((measurementSource?.needs ?? []).includes(retiredNeed)) {
    throw new Error(`browser-e2e-measurement must not depend on ${retiredNeed}`);
  }
}
if (manifest.generated?.generator !== "tools/harness/generated-artifacts/render-service-backed-schedule-manifest.mjs") {
  throw new Error("rendered schedule must record generator metadata");
}
EOF

invalid_derivation_manifest="${rendered_schedule_dir}/invalid-topology.json"
cp "$ROOT_DIR/tools/service_backed_make_target_duration_baselines.json" \
  "$rendered_schedule_dir/service_backed_make_target_duration_baselines.json"
"$NODE_BIN" - "$ROOT_DIR/tools/execution_topology_manifest.json" "$invalid_derivation_manifest" <<'EOF'
const fs = require("node:fs");
const [sourcePath, outputPath] = process.argv.slice(2);
const manifest = JSON.parse(fs.readFileSync(sourcePath, "utf8"));
manifest.browser_e2e_batch.stages.push({
  name: "tagged-without-rows",
  target: "browser-e2e-support",
  schedule_tags: ["service_backed_full"],
  groups: [
    {
      name: "missing-owner-plan-rows",
      target: "browser-e2e-support",
      kind: "support",
      coverage: "authoritative",
      execution_dependency: "browser_support",
      workers: "default",
    },
  ],
});
fs.writeFileSync(outputPath, `${JSON.stringify(manifest, null, 2)}\n`);
EOF
set +e
invalid_derivation_output="$(
  "$NODE_BIN" "$ROOT_DIR/tools/harness/generated-artifacts/render-service-backed-schedule-manifest.mjs" \
    --topology "$invalid_derivation_manifest" \
    --output "${rendered_schedule_dir}/invalid-service-backed.json" \
    2>&1
)"
invalid_derivation_status=$?
set -e
if [[ "$invalid_derivation_status" -eq 0 ]]; then
  fail "tagged browser schedule stage without matching catalog rows must fail schedule derivation"
fi
assert_contains "$invalid_derivation_output" "tagged-without-rows" "tagged browser stage without catalog rows"

invalid_resource_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-invalid-resource.XXXXXX")"
cleanup_paths+=("$invalid_resource_dir")
write_fake_make "$invalid_resource_dir"
invalid_resource_manifest="${invalid_resource_dir}/manifest.json"
write_manifest "$invalid_resource_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "undeclared-resource": 1'
set +e
invalid_resource_output="$(run_scheduler "$invalid_resource_dir" "$invalid_resource_manifest" test-fast-service-backed invalid-resource 2>&1)"
invalid_resource_status=$?
set -e
assert_equals "$invalid_resource_status" "2" "invalid resource manifest status"
assert_contains "$invalid_resource_output" "uses undeclared scheduler resource undeclared-resource" "invalid resource manifest output"

unknown_dependency_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-unknown-dependency.XXXXXX")"
cleanup_paths+=("$unknown_dependency_dir")
write_fake_make "$unknown_dependency_dir"
unknown_dependency_manifest="${unknown_dependency_dir}/manifest.json"
write_manifest "$unknown_dependency_manifest" test-fast-service-backed \
  'make_target|backend-process|10|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend||missing-target'
set +e
unknown_dependency_output="$(run_scheduler "$unknown_dependency_dir" "$unknown_dependency_manifest" test-fast-service-backed unknown-dependency 2>&1)"
unknown_dependency_status=$?
set -e
assert_equals "$unknown_dependency_status" "2" "unknown dependency manifest status"
assert_contains "$unknown_dependency_output" "depends on unknown completion key missing-target" "unknown dependency manifest output"

cycle_dependency_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-cycle-dependency.XXXXXX")"
cleanup_paths+=("$cycle_dependency_dir")
write_fake_make "$cycle_dependency_dir"
cycle_dependency_manifest="${cycle_dependency_dir}/manifest.json"
write_manifest "$cycle_dependency_manifest" test-fast-service-backed \
  'make_target|backend-process|10|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend||backend-store' \
  'make_target|backend-store|9|"postgres": 1, "object_store": 1, "go_cpu": 1, "go_io": 1|backend||backend-process'
set +e
cycle_dependency_output="$(run_scheduler "$cycle_dependency_dir" "$cycle_dependency_manifest" test-fast-service-backed cycle-dependency 2>&1)"
cycle_dependency_status=$?
set -e
assert_equals "$cycle_dependency_status" "2" "cycle dependency manifest status"
assert_contains "$cycle_dependency_output" "has a dependency cycle" "cycle dependency manifest output"

invalid_browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-invalid-browser.XXXXXX")"
cleanup_paths+=("$invalid_browser_dir")
write_fake_make "$invalid_browser_dir"
invalid_browser_manifest="${invalid_browser_dir}/manifest.json"
write_manifest "$invalid_browser_manifest" test-service-backed \
  'make_target|browser-e2e-webserver-backed|10|"postgres": 1, "object_store": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|missing-stage'
set +e
invalid_browser_output="$(run_scheduler "$invalid_browser_dir" "$invalid_browser_manifest" test-service-backed invalid-browser 2>&1)"
invalid_browser_status=$?
set -e
assert_equals "$invalid_browser_status" "2" "invalid browser manifest status"
assert_contains "$invalid_browser_output" "declares unknown browser_stage missing-stage" "invalid browser manifest output"

invalid_browser_target_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-invalid-browser-target.XXXXXX")"
cleanup_paths+=("$invalid_browser_target_dir")
write_fake_make "$invalid_browser_target_dir"
invalid_browser_target_manifest="${invalid_browser_target_dir}/manifest.json"
write_manifest "$invalid_browser_target_manifest" test-service-backed \
  'make_target|browser-e2e-visual|10|"postgres": 1, "object_store": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed'
set +e
invalid_browser_target_output="$(run_scheduler "$invalid_browser_target_dir" "$invalid_browser_target_manifest" test-service-backed invalid-browser-target 2>&1)"
invalid_browser_target_status=$?
set -e
assert_equals "$invalid_browser_target_status" "2" "invalid browser target manifest status"
assert_contains "$invalid_browser_target_output" "must match browser_stage webserver-backed aggregate target browser-e2e-webserver-backed" "invalid browser target manifest output"
fi

legacy_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-legacy.XXXXXX")"
cleanup_paths+=("$legacy_dir")
write_fake_make "$legacy_dir"
legacy_manifest="${legacy_dir}/manifest.json"
write_legacy_manifest "$legacy_manifest" "cartulary.service_backed_schedule.v11"
set +e
legacy_output="$(run_scheduler "$legacy_dir" "$legacy_manifest" test-fast-service-backed legacy 2>&1)"
legacy_status=$?
set -e
assert_equals "$legacy_status" "2" "legacy manifest status"
assert_contains "$legacy_output" "must declare schema_id cartulary.scheduler_manifest.v2" "legacy manifest output"

if [[ "$SUITE" != "fast" ]]; then
unknown_option_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-unknown-option.XXXXXX")"
cleanup_paths+=("$unknown_option_dir")
write_fake_make "$unknown_option_dir"
unknown_option_manifest="${unknown_option_dir}/manifest.json"
write_manifest "$unknown_option_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "object_store": 1'
set +e
unknown_option_output="$(
  env \
  FAKE_SCHEDULER_LOCK="${unknown_option_dir}/lock" \
  FAKE_SCHEDULER_ACTIVE="${unknown_option_dir}/active" \
  FAKE_SCHEDULER_MAX="${unknown_option_dir}/max" \
  FAKE_SCHEDULER_LOG="${unknown_option_dir}/make.log" \
  MAKE="${unknown_option_dir}/fake-make" \
  NODE_BIN="$NODE_BIN" \
  TEST_OUTPUT_SCRIPT="$TEST_OUTPUT_SCRIPT" \
  CARTULARY_TEST_RESULTS_DIR="${unknown_option_dir}/results" \
  CARTULARY_TEST_RUN_ID="unknown-option" \
    "$NODE_BIN" "$SCRIPT" --target test-fast-service-backed --unknown-option --manifest "$unknown_option_manifest" 2>&1
)"
unknown_option_status=$?
set -e
assert_equals "$unknown_option_status" "2" "unknown option status"
assert_contains "$unknown_option_output" "usage: run-service-backed-schedule.mjs" "unknown option output"
fi
