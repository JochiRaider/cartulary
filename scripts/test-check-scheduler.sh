#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT_DIR}/scripts/run-check-schedule.mjs"
TEST_OUTPUT_SCRIPT="${ROOT_DIR}/scripts/lib/test-output.sh"
NODE_BIN="${NODE_BIN:-node}"
cleanup_paths=()

unset VERBOSE CI_VERBOSE CARTULARY_OUTPUT_MODE

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
  local progress_file="${dir}/results/${run_id}/${target}/progress-summary.log"

  assert_file_present "$summary_file" "$target scheduler summary"
  assert_file_present "$events_file" "$target scheduler events"
  assert_file_present "$progress_file" "$target progress summary"
  "$NODE_BIN" - "$summary_file" "$events_file" "$progress_file" "$expected_status" "$expected_failed" "$expected_total" "$expected_event" "$ROOT_DIR" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [summaryFile, eventsFile, progressFile, expectedStatus, expectedFailed, expectedTotal, expectedEvent, repoRoot] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).filter(Boolean).map((line) => JSON.parse(line));
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
if (summary.schema_id !== "cartulary.check_scheduler_summary.v6") {
  throw new Error(`unexpected summary schema ${summary.schema_id}`);
}
if (summary.scheduler_kind !== "check") {
  throw new Error(`summary scheduler_kind got ${summary.scheduler_kind} want check`);
}
if (summary.status !== expectedStatus) {
  throw new Error(`summary status got ${summary.status} want ${expectedStatus}`);
}
if (expectedStatus === "fail" && summary.failure_class !== "helper") {
  throw new Error(`summary failure_class got ${summary.failure_class} want helper`);
}
if (expectedStatus === "pass" && summary.failure_class !== null) {
  throw new Error(`passing summary failure_class got ${summary.failure_class}`);
}
if (summary.total_work_units !== Number(expectedTotal)) {
  throw new Error(`total work units got ${summary.total_work_units} want ${expectedTotal}`);
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
if (!progressLog.includes(`[CHECK-SCHEDULER] ${summary.target} summary `)) {
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
if (!events.every((event) => event.schema_id === "cartulary.check_scheduler_event.v5")) {
  throw new Error("unexpected scheduler event schema");
}
events.forEach((event, index) => {
  if (event.event_sequence !== index + 1) {
    throw new Error(`event ${index} sequence got ${event.event_sequence} want ${index + 1}`);
  }
  if (!Number.isInteger(event.monotonic_ms) || event.monotonic_ms < 0) {
    throw new Error(`event ${index} missing monotonic_ms`);
  }
  if (index > 0 && event.monotonic_ms < events[index - 1].monotonic_ms) {
    throw new Error(`event ${index} monotonic_ms regressed`);
  }
  if (typeof event.wall_timestamp !== "string" || Number.isNaN(Date.parse(event.wall_timestamp))) {
    throw new Error(`event ${index} missing wall_timestamp`);
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
  if (!progressLog.includes(`[PROGRESS] ${summary.target} `)) {
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

write_phase_summary() {
  if [[ -z "${CARTULARY_TEST_RESULTS_DIR:-}" || -z "${CARTULARY_TEST_RUN_ID:-}" ]]; then
    return 0
  fi
  local artifact_target="${CARTULARY_TEST_TARGET:-adhoc}"
  local phase_dir="${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${artifact_target}/${target}"
  mkdir -p "$phase_dir"
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
    printf '{"schema_id":"cartulary.service_backed_scheduler_event.v5","target":"partial-service","event":"progress","event_sequence":1,"monotonic_ms":1,"wall_timestamp":"2026-01-01T00:00:00.001Z"' >"${nested_dir}/scheduler-events.jsonl"
    return 0
  fi
  {
    printf 'not-json-diagnostic\n'
    printf '%s\n' '{"schema_id":"cartulary.service_backed_scheduler_event.v5","target":"service","event":"progress","event_sequence":1,"monotonic_ms":1,"wall_timestamp":"2026-01-01T00:00:00.001Z","pending":2,"running":1,"total_work_units":6,"blocked":2,"completed":3,"pending_finalizers":0,"running_finalizers":0,"blocked_reason":"resources","blocked_resources":["go_io"],"waiting_on":["backend-store"],"blocked_units":[],"active_resource_claims":{"go_cpu":1},"resource_limits":{"go_cpu":1,"go_io":1},"active_groups":{"backend-integration":1},"blocked_by":["go_io"],"unblocks_after":"backend-integration/shard-a","slowest_running":{"label":"backend-integration/shard-a","duration_ms":1234}}'
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
  FAKE_SLEEP_PARTIAL_SERVICE="${FAKE_SLEEP_PARTIAL_SERVICE:-}" \
  FAKE_FAIL_TARGET="${FAKE_FAIL_TARGET:-}" \
  MAKE="${dir}/fake-make" \
  NODE_BIN="$NODE_BIN" \
  TEST_OUTPUT_SCRIPT="$TEST_OUTPUT_SCRIPT" \
  VERBOSE="${VERBOSE:-}" \
  CI_VERBOSE="${CI_VERBOSE:-}" \
  CARTULARY_OUTPUT_MODE="${CARTULARY_OUTPUT_MODE:-}" \
  CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS="${CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS:-}" \
  CARTULARY_TEST_RESULTS_DIR="${dir}/results" \
  CARTULARY_TEST_RUN_ID="$run_id" \
    "$NODE_BIN" "$SCRIPT" --target check --manifest "$manifest" "$@"
}

"$NODE_BIN" --input-type=module - "$ROOT_DIR" <<'EOF'
import {
  assertKnownResource,
  browserStageResource,
  isAutoLimitResource,
  normalizeResourceClaims,
  normalizeResourceLimits,
  preferredResourcesForScheduler,
  resolveForwardingProfile,
} from "./scripts/lib/scheduler-resources.mjs";

const fail = (message) => {
  throw new Error(message);
};

if (browserStageResource("webserver-backed") !== "browser_stage_webserver_backed") {
  fail("browser stage lane derivation changed");
}
if (preferredResourcesForScheduler("check").join(",") !== "host_cpu,host_io,service_stack") {
  fail("check resource display order changed");
}
if (!isAutoLimitResource("go_cpu") || !isAutoLimitResource("go_io") || !isAutoLimitResource("browser_stack")) {
  fail("service-backed auto-limit resources are incomplete");
}
if (isAutoLimitResource("host_cpu")) {
  fail("host_cpu must not be an auto-limit resource");
}
try {
  assertKnownResource("cpu", "registry test", { scheduler: "check" });
  fail("retired cpu alias was accepted");
} catch (error) {
  if (!String(error.message).includes("use host_cpu")) {
    throw error;
  }
}
const { limits } = normalizeResourceLimits(
  { host_cpu: 4, host_io: 3, service_stack: 1 },
  "registry test",
  { scheduler: "check" },
);
const claims = normalizeResourceClaims(
  { host_cpu: { mode: "bounded_limit", reserve: 1, min: 1, max: 3 }, host_io: 1, service_stack: 1 },
  "registry test",
  limits,
  { scheduler: "check", allowBounded: true },
);
if (claims.get("host_cpu") !== 3) {
  fail(`bounded host_cpu claim got ${claims.get("host_cpu")}`);
}
const forwarding = resolveForwardingProfile("check_host_to_service_backed_go", claims, "registry test");
if (forwarding.forwardedResourceLimits.get("go_cpu") !== 3 || forwarding.forwardedResourceLimits.get("go_io") !== 1) {
  fail("forwarded service-backed limits were not resolved from host claims");
}
EOF

event_order_dir="$(mktemp -d "${ROOT_DIR}/tmp/scheduler-event-order.XXXXXX")"
cleanup_paths+=("$event_order_dir")
mkdir -p "${event_order_dir}/valid/check" "${event_order_dir}/sequence/check" "${event_order_dir}/monotonic/check" "${event_order_dir}/wall/check" "${event_order_dir}/skew/check"
cat >"${event_order_dir}/valid/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"start","event_sequence":1,"monotonic_ms":0,"wall_timestamp":"2026-01-01T00:00:00.000Z"}
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"finish","event_sequence":2,"monotonic_ms":5,"wall_timestamp":"2026-01-01T00:00:00.005Z"}
JSONL
cat >"${event_order_dir}/sequence/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"start","event_sequence":1,"monotonic_ms":0,"wall_timestamp":"2026-01-01T00:00:00.000Z"}
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"finish","event_sequence":3,"monotonic_ms":5,"wall_timestamp":"2026-01-01T00:00:00.005Z"}
JSONL
cat >"${event_order_dir}/monotonic/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"start","event_sequence":1,"monotonic_ms":10,"wall_timestamp":"2026-01-01T00:00:00.010Z"}
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"finish","event_sequence":2,"monotonic_ms":5,"wall_timestamp":"2026-01-01T00:00:00.005Z"}
JSONL
cat >"${event_order_dir}/wall/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"start","event_sequence":1,"monotonic_ms":0,"wall_timestamp":"2026-01-01T00:00:02.000Z"}
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"finish","event_sequence":2,"monotonic_ms":5,"wall_timestamp":"2026-01-01T00:00:01.000Z"}
JSONL
cat >"${event_order_dir}/skew/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"start","event_sequence":1,"monotonic_ms":0,"wall_timestamp":"2026-01-01T00:00:02.000Z"}
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"clock-skew","event_sequence":2,"monotonic_ms":1,"wall_timestamp":"2026-01-01T00:00:03.000Z"}
{"schema_id":"cartulary.check_scheduler_event.v5","target":"check","event":"finish","event_sequence":3,"monotonic_ms":5,"wall_timestamp":"2026-01-01T00:00:01.000Z"}
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
assert_contains "$sequence_output" "event_sequence got 3, want 2" "event sequence drift fixture output"
assert_equals "$monotonic_status" "1" "monotonic drift fixture status"
assert_contains "$monotonic_output" "monotonic_ms regressed" "monotonic drift fixture output"
assert_equals "$wall_status" "1" "wall drift fixture status"
assert_contains "$wall_output" "wall_timestamp regressed without preceding clock-skew marker" "wall drift fixture output"
assert_contains "$("$NODE_BIN" "$ROOT_DIR/scripts/check-scheduler-event-order-drift.mjs" "${event_order_dir}/skew" 2>&1)" "scheduler event order verified" "clock skew marker drift fixture"

success_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-success.XXXXXX")"
cleanup_paths+=("$success_dir")
write_fake_make "$success_dir"
success_manifest="${success_dir}/manifest.json"
cat >"$success_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v6",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 12, "host_io": 12, "service_stack": 1 },
      "summary_groups": [
        { "name": "check-work", "summary_targets": ["local", "service", "meta"] }
      ],
      "work_units": [
        { "target": "setup", "weight": 50, "needs": [], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "build", "weight": 40, "needs": ["setup"], "resource_claims": { "host_cpu": "limit" }, "make_jobs": "host_cpu" },
        { "target": "local", "weight": 30, "needs": ["build"], "produces_summary_targets": ["local"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        {
          "target": "service",
          "weight": 20,
          "needs": ["build"],
          "produces_summary_targets": ["service"],
          "resource_claims": {
            "host_cpu": { "mode": "bounded_limit", "reserve": 3, "min": 1, "max": 8 },
            "host_io": { "mode": "bounded_limit", "reserve": 4, "min": 1, "max": 10 },
            "service_stack": 1
          },
          "make_jobs": "host_cpu",
          "nested_scheduler": {
            "type": "service_backed",
            "target": "service",
            "manifest": "tools/service_backed_schedule_manifest.json",
            "forwarding": "check_host_to_service_backed_go"
          }
        },
        { "target": "meta", "weight": 10, "needs": ["build"], "produces_summary_targets": ["meta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" }
      ]
    }
  ]
}
JSON
success_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_LOCAL=0.2 FAKE_SLEEP_SERVICE=0.2 run_scheduler "$success_dir" "$success_manifest" success --resource-limit host_cpu=2 --resource-limit host_io=3 2>&1)"
assert_contains "$success_output" "[RUN] check work_units=5 summary_targets=3 helper_units=2 jobs=2 run_id=success" "success run start"
assert_contains "$success_output" "[CHECK-SCHEDULER] check start work_units=5 capacity={host_cpu:2,host_io:3,service_stack:1}" "success concise scheduler start"
assert_contains "$success_output" "top_weighted=setup:50,build:40,local:30,service:20,meta:10" "success concise scheduler start shows top weighted work"
assert_contains "$success_output" "top_weighted=setup:50,build:40,local:30,service:20,meta:10 artifacts=tmp/check-scheduler-success" "success concise scheduler start shows artifact path"
assert_contains "$success_output" "[PROGRESS] check 0/5: check 0/5, running setup, blocked by dependencies, waiting on build,setup, unblocks after setup, slowest setup " "success human scheduler progress"
assert_contains "$success_output" "; bottleneck service 3/6" "success human scheduler progress includes nested service bottleneck"
assert_contains "$success_output" "blocked by go_io, waiting on backend-store, unblocks after backend-integration/shard-a, slowest backend-integration/shard-a 1.23s" "success human nested scheduler progress"
assert_occurrences "$success_output" "bottleneck service 3/6" "1" "success human nested scheduler progress is material-change throttled"
assert_contains "$success_output" "logs tmp/check-scheduler-success" "success human scheduler progress artifact path"
assert_contains "$success_output" "/results/success/check" "success concise scheduler progress artifact path suffix"
assert_not_contains "$success_output" "[CHECK-SCHEDULER] check progress completed_work_units=" "quiet check scheduler hides key/value progress"
assert_not_contains "$success_output" "[CHECK-SCHEDULER] check nested-progress" "quiet check scheduler hides key/value nested progress"
assert_contains "$success_output" "[CHECK-SCHEDULER] check summary status=pass completed_work_units=5/5 failed=none slowest=" "success concise scheduler summary"
assert_contains "$success_output" "slowest=" "success concise scheduler summary includes slowest work"
assert_contains "$success_output" "artifacts=tmp/check-scheduler-success" "success concise scheduler summary artifact path"
assert_not_contains "$success_output" "active_resource_claims=" "default scheduler output hides raw active resources"
assert_not_contains "$success_output" "resource_limits=" "default scheduler output hides raw resource limits"
assert_not_contains "$success_output" "claims={" "default scheduler output hides raw claims"
assert_not_contains "$success_output" "[STEP] check" "default scheduler output hides per-unit steps"
assert_not_contains "$success_output" "running_units=" "default scheduler output hides raw running units"
assert_not_contains "$success_output" "blocked_resources=" "default scheduler output hides raw blocked resources"
assert_contains "$success_output" "[PASS] check" "success summary"
assert_contains "$(cat "${success_dir}/make-args.log")" "--output-sync=target -j2 build" "build uses claimed host_cpu jobs"
success_events="$(cat "${success_dir}/events.log")"
assert_contains "$success_events" "end local" "success local completed"
assert_fake_make_overlap "${success_dir}/events.log" local service "success service overlapped with cheap local work"
assert_contains "$success_events" "env service go_cpu=1 go_io=1" "success service forwarded bounded nested scheduler limits"
assert_contains "$success_events" "end service" "success service completed"
assert_contains "$success_events" "end meta" "success meta completed"
assert_not_contains "$success_events" "browser" "success check schedule has no browser tail"
assert_contains "$success_events" "test-target setup setup" "success setup receives target-owned identity"
assert_file_present "${success_dir}/results/success/setup/setup/phase-summary.json" "success setup target-owned helper phase summary"
assert_file_absent "${success_dir}/results/success/adhoc/setup/phase-summary.json" "success setup helper is not adhoc-owned"
success_summary="${success_dir}/results/success/run-summary.json"
success_target_summary="${success_dir}/results/success/check/target-summary.json"
assert_equals "$(json_field "$success_summary" "status")" "pass" "success summary status"
assert_equals "$(json_field "$success_summary" "schema_id")" "cartulary.test_run_summary.v6" "success run summary schema"
assert_file_present "$success_target_summary" "success check target summary"
assert_equals "$(json_field "$success_target_summary" "target")" "check" "success check target summary identity"
assert_equals "$(json_field "$success_target_summary" "status")" "pass" "success check target summary status"
assert_equals "$(json_field "$success_summary" "work_units.completed")" "5" "success completed work units"
assert_equals "$(json_field "$success_summary" "work_units.total")" "5" "success total work units"
assert_equals "$(json_field "$success_summary" "summary_targets.expected.length")" "3" "success summary target count"
assert_equals "$(json_field "$success_summary" "evidence_targets.present.length")" "3" "success evidence target count"
assert_equals "$(json_field "$success_summary" "helper_units.total")" "2" "success helper unit count"
assert_equals "$(json_field "$success_summary" "helper_units.artifacts.0.target")" "setup" "success helper artifact target"
assert_contains "$(json_field "$success_summary" "helper_units.artifacts.0.phase_summaries.0.artifact")" "/setup/setup/phase-summary.json" "success helper artifact phase summary path"
success_scheduler_summary="${success_dir}/results/success/check/scheduler-summary.json"
success_scheduler_events="${success_dir}/results/success/check/scheduler-events.jsonl"
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
if (!events.some((event) => event.resource_limits?.host_cpu === 2 && event.resource_limits?.host_io === 3 && event.resource_limits?.service_stack === 1)) {
  throw new Error("scheduler events must preserve resource limits");
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
const serviceStart = events.find((event) => event.event === "start" && event.work_unit === "service");
if (serviceStart?.nested_scheduler?.forwarded_limits?.CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT !== 1) {
  throw new Error("service start event must record forwarded bounded nested go host_cpu limit");
}
if (serviceStart?.nested_scheduler?.forwarded_limits?.CARTULARY_SERVICE_BACKED_GO_IO_LIMIT !== 1) {
  throw new Error("service start event must record forwarded bounded nested go host_io limit");
}
const serviceSummary = summary.nested_scheduler_limits?.find((entry) => entry.work_unit === "service");
if (serviceSummary?.forwarded_limits?.CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT !== 1) {
  throw new Error("summary must record forwarded bounded nested go host_cpu limit");
}
if (serviceSummary?.forwarded_limits?.CARTULARY_SERVICE_BACKED_GO_IO_LIMIT !== 1) {
  throw new Error("summary must record forwarded bounded nested go host_io limit");
}
const progressEvent = events.find((event) => event.event === "progress" && event.nested_scheduler_progress?.length > 0);
if (!progressEvent) {
  throw new Error("outer scheduler progress events must record nested scheduler progress");
}
if (!progressEvent.active_groups || progressEvent.blocked_by === undefined || !Object.hasOwn(progressEvent, "unblocks_after") || !Object.hasOwn(progressEvent, "slowest_running")) {
  throw new Error("outer progress event must expose structured v3 progress fields");
}
const nestedProgress = progressEvent.nested_scheduler_progress.find((entry) => entry.work_unit === "service");
if (!nestedProgress) {
  throw new Error("outer progress event must include service nested scheduler snapshot");
}
if (nestedProgress.active_groups?.["backend-integration"] !== 1) {
  throw new Error("nested progress must preserve active_groups");
}
if (!nestedProgress.blocked_by?.includes("go_io")) {
  throw new Error("nested progress must preserve blocked_by");
}
if (!nestedProgress.waiting_on?.includes("backend-store")) {
  throw new Error("nested progress must preserve waiting_on");
}
if (nestedProgress.unblocks_after !== "backend-integration/shard-a") {
  throw new Error(`nested progress unblocks_after got ${nestedProgress.unblocks_after}`);
}
if (nestedProgress.slowest_running?.label !== "backend-integration/shard-a") {
  throw new Error("nested progress must preserve structured slowest_running");
}
const serviceObservation = summary.nested_scheduler_observations?.find((entry) => entry.work_unit === "service");
if (!serviceObservation || serviceObservation.observed_progress_events < 1) {
  throw new Error("summary must record nested scheduler observations");
}
if (serviceObservation.latest_progress?.events_jsonl?.includes("service/scheduler-events.jsonl") !== true) {
  throw new Error("nested observation must include nested event artifact");
}
if (!summary.slowest_running_observations?.some((entry) => entry.source === "nested" && entry.work_unit === "service" && entry.label === "backend-integration/shard-a")) {
  throw new Error("summary must retain nested scheduler slowest running observations");
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

verbose_output="$(VERBOSE=1 run_scheduler "$success_dir" "$success_manifest" verbose --resource-limit host_cpu=2 --resource-limit host_io=3 2>&1)"
assert_contains "$verbose_output" "[CHECK-SCHEDULER] check start work_unit=setup claims={host_cpu:1} active=1 pending=4" "verbose scheduler start telemetry"
assert_contains "$verbose_output" "active_resource_claims={host_cpu:1}" "verbose scheduler active resource telemetry"
assert_contains "$verbose_output" "resource_limits={host_cpu:2,host_io:3,service_stack:1}" "verbose scheduler resource limit telemetry"

machine_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_SERVICE=0.2 CARTULARY_OUTPUT_MODE=machine run_scheduler "$success_dir" "$success_manifest" machine --resource-limit host_cpu=2 --resource-limit host_io=3 2>&1)"
assert_contains "$machine_output" "[CHECK-SCHEDULER] check progress completed_work_units=" "machine scheduler prints key/value progress"
assert_contains "$machine_output" "[CHECK-SCHEDULER] check nested-progress work_unit=service nested_target=service" "machine scheduler prints key/value nested progress"
assert_not_contains "$machine_output" "[PROGRESS] check" "machine scheduler does not print human progress"

partial_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-partial-nested.XXXXXX")"
cleanup_paths+=("$partial_dir")
write_fake_make "$partial_dir"
partial_manifest="${partial_dir}/manifest.json"
cat >"$partial_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v6",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1, "host_io": 1 },
      "work_units": [
        {
          "target": "partial-service",
          "weight": 1,
          "needs": [],
          "produces_summary_targets": ["partial-service"],
          "resource_claims": { "host_cpu": 1, "host_io": 1 },
          "make_jobs": "host_cpu",
          "nested_scheduler": {
            "type": "service_backed",
            "target": "partial-service",
            "manifest": "tools/service_backed_schedule_manifest.json",
            "forwarding": "check_host_to_service_backed_go"
          }
        }
      ]
    }
  ]
}
JSON
partial_output="$(CARTULARY_SCHEDULER_PROGRESS_INTERVAL_MS=25 FAKE_SLEEP_PARTIAL_SERVICE=0.15 run_scheduler "$partial_dir" "$partial_manifest" partial --resource-limit host_cpu=1 --resource-limit host_io=1 2>&1)"
assert_contains "$partial_output" "[PASS] check" "partial nested event does not fail check scheduler"
assert_not_contains "$partial_output" "nested-progress work_unit=partial-service" "partial nested event is ignored until newline-complete"

makeflags_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-makeflags.XXXXXX")"
cleanup_paths+=("$makeflags_dir")
write_fake_make "$makeflags_dir"
makeflags_manifest="${makeflags_dir}/manifest.json"
cat >"$makeflags_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v6",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1 },
      "work_units": [
        { "target": "alpha", "weight": 1, "needs": [], "produces_summary_targets": ["alpha"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" }
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
assert_contains "$makeflags_output" "[PASS] check" "makeflags sanitize summary"
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
  "schema_id": "cartulary.check_schedule.v6",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 2, "service_stack": 1 },
      "work_units": [
        { "target": "alpha", "weight": 30, "needs": [], "produces_summary_targets": ["alpha"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "beta", "weight": 20, "needs": [], "produces_summary_targets": ["beta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "gamma", "weight": 10, "needs": [], "produces_summary_targets": ["gamma", "external-summary"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" },
        { "target": "delta", "weight": 5, "needs": ["beta"], "resource_claims": { "host_cpu": 1 }, "make_jobs": "host_cpu" }
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
assert_equals "$failure_status" "7" "failure exit status"
assert_contains "$failure_output" "fake failure for beta" "failure child output"
assert_contains "$failure_output" "[FAIL] check" "failure summary"
assert_contains "$failure_output" "[CHECK-SCHEDULER] check summary status=fail" "failure scheduler status summary"
assert_contains "$failure_output" "failure_class=helper" "failure scheduler class output"
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
assert_equals "$(json_field "$failure_summary" "failure_class")" "helper" "failure summary class"
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
assert_equals "$(json_field "$failure_scheduler_summary" "failure_class")" "helper" "failure scheduler summary class"
failure_scheduler_events="${failure_dir}/results/failure/check/scheduler-events.jsonl"
assert_check_scheduler_artifacts "$failure_dir" failure check fail beta 4 skip
"$NODE_BIN" - "$failure_scheduler_summary" "$failure_scheduler_events" "$ROOT_DIR" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [summaryFile, eventsFile, repoRoot] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
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

invalid_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-invalid.XXXXXX")"
cleanup_paths+=("$invalid_dir")
write_fake_make "$invalid_dir"
invalid_manifest="${invalid_dir}/manifest.json"
cat >"$invalid_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v6",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1 },
      "work_units": [
        { "target": "alpha", "weight": 1, "needs": ["missing"], "resource_claims": { "host_cpu": 1 } }
      ]
    }
  ]
}
JSON
set +e
invalid_output="$(run_scheduler "$invalid_dir" "$invalid_manifest" invalid 2>&1)"
invalid_status=$?
set -e
assert_equals "$invalid_status" "1" "invalid dependency status"
assert_contains "$invalid_output" "depends on unknown target missing" "invalid dependency output"

invalid_nested_manifest="${invalid_dir}/invalid-nested-manifest.json"
cat >"$invalid_nested_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v6",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 1, "host_io": 1 },
      "work_units": [
        {
          "target": "service",
          "weight": 1,
          "needs": [],
          "resource_claims": { "host_cpu": 1 },
          "nested_scheduler": {
            "type": "service_backed",
            "target": "service",
            "manifest": "tools/service_backed_schedule_manifest.json",
            "forwarding": "check_host_to_service_backed_go"
          }
        }
      ]
    }
  ]
}
JSON
set +e
invalid_nested_output="$(run_scheduler "$invalid_dir" "$invalid_nested_manifest" invalid-nested 2>&1)"
invalid_nested_status=$?
set -e
assert_equals "$invalid_nested_status" "1" "invalid nested scheduler status"
assert_contains "$invalid_nested_output" "source host_io must be claimed by work unit" "invalid nested scheduler output"

invalid_bounded_manifest="${invalid_dir}/invalid-bounded-manifest.json"
cat >"$invalid_bounded_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v6",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "host_cpu": 2 },
      "work_units": [
        {
          "target": "alpha",
          "weight": 1,
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
invalid_bounded_output="$(run_scheduler "$invalid_dir" "$invalid_bounded_manifest" invalid-bounded 2>&1)"
invalid_bounded_status=$?
set -e
assert_equals "$invalid_bounded_status" "1" "invalid bounded claim status"
assert_contains "$invalid_bounded_output" "resource_claims.host_cpu bounded_limit has unknown key legacy" "invalid bounded claim output"

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-dry-run.XXXXXX")"
cleanup_paths+=("$dry_run_dir")
write_fake_make "$dry_run_dir"
dry_run_output="$(
  MAKEFLAGS=n \
    run_scheduler "$dry_run_dir" "$success_manifest" dry-run --resource-limit host_cpu=2 --resource-limit host_io=3 2>&1
)"
assert_contains "$dry_run_output" "[DRY-RUN] check manifest=" "dry-run output"
assert_contains "$dry_run_output" "resource_limits={host_cpu:2,host_io:3,service_stack:1} work_units=5 dependencies=4 top_weighted=setup:50,build:40,local:30,service:20,meta:10" "dry-run compact summary"
assert_not_contains "$dry_run_output" "[DRY-RUN] check unit" "dry-run default hides unit expansion"
assert_not_contains "$dry_run_output" "claims={" "dry-run default hides raw claims"
assert_file_absent "${dry_run_dir}/make-args.log" "dry-run child make"

dry_run_verbose_output="$(
  MAKEFLAGS=n VERBOSE=1 \
    run_scheduler "$dry_run_dir" "$success_manifest" dry-run-verbose --resource-limit host_cpu=2 --resource-limit host_io=3 2>&1
)"
assert_contains "$dry_run_verbose_output" "[DRY-RUN] check unit setup needs=none claims={host_cpu:1} make_jobs=1" "verbose dry-run includes unit claims"
assert_contains "$dry_run_verbose_output" "nested_scheduler={\"type\":\"service_backed\",\"target\":\"service\"" "verbose dry-run includes nested scheduler metadata"
