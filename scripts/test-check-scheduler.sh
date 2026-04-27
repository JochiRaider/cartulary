#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT_DIR}/scripts/run-check-schedule.mjs"
TEST_OUTPUT_SCRIPT="${ROOT_DIR}/scripts/lib/test-output.sh"
NODE_BIN="${NODE_BIN:-node}"
cleanup_paths=()

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

  assert_file_present "$summary_file" "$target scheduler summary"
  assert_file_present "$events_file" "$target scheduler events"
  "$NODE_BIN" - "$summary_file" "$events_file" "$expected_status" "$expected_failed" "$expected_total" "$expected_event" "$ROOT_DIR" <<'EOF'
const fs = require("node:fs");
const path = require("node:path");
const [summaryFile, eventsFile, expectedStatus, expectedFailed, expectedTotal, expectedEvent, repoRoot] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).filter(Boolean).map((line) => JSON.parse(line));
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
if (summary.schema_id !== "cartulary.check_scheduler_summary.v1") {
  throw new Error(`unexpected summary schema ${summary.schema_id}`);
}
if (summary.status !== expectedStatus) {
  throw new Error(`summary status got ${summary.status} want ${expectedStatus}`);
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
if (!summary.artifacts?.events_jsonl) {
  throw new Error("summary must record scheduler event artifact path");
}
if (!summary.artifacts?.scheduler_logs_dir) {
  throw new Error("summary must record scheduler log artifact path");
}
assertRepoRelativeArtifact(summary.artifacts.events_jsonl, "events_jsonl");
assertRepoRelativeArtifact(summary.artifacts.scheduler_logs_dir, "scheduler_logs_dir");
const schedulerLogsDir = resolveArtifact(summary.artifacts.scheduler_logs_dir);
if (!fs.statSync(schedulerLogsDir).isDirectory()) {
  throw new Error(`scheduler log artifact path must be an existing directory: ${summary.artifacts.scheduler_logs_dir}`);
}
if (events.length === 0) {
  throw new Error("scheduler events must not be empty");
}
if (!events.every((event) => event.schema_id === "cartulary.check_scheduler_event.v1")) {
  throw new Error("unexpected scheduler event schema");
}
if (expectedEvent !== "-" && !events.some((event) => event.event === expectedEvent)) {
  throw new Error(`missing scheduler event ${expectedEvent}`);
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

sleep_key="${target//-/_}"
sleep_key="${sleep_key^^}"
sleep_var="FAKE_SLEEP_${sleep_key}"
sleep_duration="${!sleep_var:-${FAKE_SLEEP_DEFAULT:-0.05}}"

change_active 1
sleep "$sleep_duration"
if [[ "${FAKE_FAIL_TARGET:-}" == "$target" ]]; then
  echo "fake failure for $target" >&2
  change_active -1
  exit 7
fi
change_active -1

echo "fake pass for $target"
write_summary
EOF
  chmod +x "${dir}/fake-make"
}

run_scheduler() {
  local dir="$1"
  local manifest="$2"
  local run_id="$3"
  shift 3

  FAKE_CHECK_SCHEDULER_LOCK="${dir}/lock" \
  FAKE_CHECK_SCHEDULER_ACTIVE="${dir}/active" \
  FAKE_CHECK_SCHEDULER_MAX="${dir}/max" \
  FAKE_CHECK_SCHEDULER_ARGS_LOG="${dir}/make-args.log" \
  FAKE_CHECK_SCHEDULER_EVENT_LOG="${dir}/events.log" \
  FAKE_SLEEP_DEFAULT="${FAKE_SLEEP_DEFAULT:-0.05}" \
  FAKE_SLEEP_ALPHA="${FAKE_SLEEP_ALPHA:-}" \
  FAKE_SLEEP_BETA="${FAKE_SLEEP_BETA:-}" \
  FAKE_FAIL_TARGET="${FAKE_FAIL_TARGET:-}" \
  MAKE="${dir}/fake-make" \
  NODE_BIN="$NODE_BIN" \
  TEST_OUTPUT_SCRIPT="$TEST_OUTPUT_SCRIPT" \
  CARTULARY_TEST_RESULTS_DIR="${dir}/results" \
  CARTULARY_TEST_RUN_ID="$run_id" \
    "$NODE_BIN" "$SCRIPT" --target check --manifest "$manifest" "$@"
}

success_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-success.XXXXXX")"
cleanup_paths+=("$success_dir")
write_fake_make "$success_dir"
success_manifest="${success_dir}/manifest.json"
cat >"$success_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v2",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "cpu": 2, "io": 3, "service_stack": 1 },
      "work_units": [
        { "target": "setup", "weight": 50, "needs": [], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" },
        { "target": "build", "weight": 40, "needs": ["setup"], "resource_claims": { "cpu": "limit" }, "make_jobs": "cpu" },
        { "target": "local", "weight": 30, "needs": ["build"], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" },
        {
          "target": "service",
          "weight": 20,
          "needs": ["build"],
          "resource_claims": { "cpu": "limit", "io": "limit", "service_stack": 1 },
          "make_jobs": "cpu",
          "nested_scheduler": {
            "type": "service_backed",
            "target": "service",
            "manifest": "tools/service_backed_schedule_manifest.json",
            "resource_limit_env": {
              "cpu": "CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT",
              "io": "CARTULARY_SERVICE_BACKED_GO_IO_LIMIT"
            }
          }
        },
        { "target": "meta", "weight": 10, "needs": ["build"], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" }
      ]
    }
  ]
}
JSON
success_output="$(run_scheduler "$success_dir" "$success_manifest" success --summary-targets local,service,meta --summary-groups "check-work=local,service,meta" --resource-limit cpu=2 --resource-limit io=3 2>&1)"
assert_contains "$success_output" "[RUN] check steps=5 targets=3 jobs=2 run_id=success" "success run start"
assert_contains "$success_output" "[CHECK-SCHEDULER] check start work_units=5 capacity={cpu:2,io:3,service_stack:1}" "success concise scheduler start"
assert_contains "$success_output" "top_weighted=setup:50,build:40,local:30,service:20,meta:10" "success concise scheduler start shows top weighted work"
assert_contains "$success_output" "[CHECK-SCHEDULER] check progress completed=0/5 running=1 pending=4 blocked=4 active_groups=setup:1 blocked_by=dependencies unblocks_after=setup slowest_running=setup:" "success concise scheduler progress"
assert_contains "$success_output" "artifacts=tmp/check-scheduler-success" "success concise scheduler progress artifact path"
assert_contains "$success_output" "/results/success/check" "success concise scheduler progress artifact path suffix"
assert_contains "$success_output" "[CHECK-SCHEDULER] check summary status=pass completed=5/5 failed=none slowest=" "success concise scheduler summary"
assert_not_contains "$success_output" "active_resource_claims=" "default scheduler output hides raw active resources"
assert_not_contains "$success_output" "resource_limits=" "default scheduler output hides raw resource limits"
assert_not_contains "$success_output" "claims={" "default scheduler output hides raw claims"
assert_not_contains "$success_output" "[STEP] check" "default scheduler output hides per-unit steps"
assert_not_contains "$success_output" "running_units=" "default scheduler output hides raw running units"
assert_not_contains "$success_output" "blocked_resources=" "default scheduler output hides raw blocked resources"
assert_contains "$success_output" "[PASS] check" "success summary"
assert_contains "$(cat "${success_dir}/make-args.log")" "--output-sync=target -j2 build" "build uses claimed cpu jobs"
success_events="$(cat "${success_dir}/events.log")"
assert_contains "$success_events" "end local" "success local completed"
assert_contains "$success_events" "env service go_cpu=2 go_io=3" "success service forwarded nested scheduler limits"
assert_contains "$success_events" "end service" "success service completed"
assert_contains "$success_events" "end meta" "success meta completed"
assert_not_contains "$success_events" "browser" "success check schedule has no browser tail"
success_summary="${success_dir}/results/success/run-summary.json"
assert_equals "$(json_field "$success_summary" "status")" "pass" "success summary status"
assert_equals "$(json_field "$success_summary" "completed_targets")" "5/5" "success completed"
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
if (!events.some((event) => event.resource_limits?.cpu === 2 && event.resource_limits?.io === 3 && event.resource_limits?.service_stack === 1)) {
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
if (summary.max_active_resource_claims?.cpu !== 2) {
  throw new Error(`max active cpu claims got ${summary.max_active_resource_claims?.cpu} want 2`);
}
if (!events.some((event) => event.running >= 2 && event.active_resource_claims?.cpu === 2)) {
  throw new Error("scheduler events must record two logically admitted cpu work units");
}
const serviceStart = events.find((event) => event.event === "start" && event.work_unit === "service");
if (serviceStart?.nested_scheduler?.forwarded_limits?.CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT !== 2) {
  throw new Error("service start event must record forwarded nested go cpu limit");
}
if (serviceStart?.nested_scheduler?.forwarded_limits?.CARTULARY_SERVICE_BACKED_GO_IO_LIMIT !== 3) {
  throw new Error("service start event must record forwarded nested go io limit");
}
const serviceSummary = summary.nested_scheduler_limits?.find((entry) => entry.work_unit === "service");
if (serviceSummary?.forwarded_limits?.CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT !== 2) {
  throw new Error("summary must record forwarded nested go cpu limit");
}
if (serviceSummary?.forwarded_limits?.CARTULARY_SERVICE_BACKED_GO_IO_LIMIT !== 3) {
  throw new Error("summary must record forwarded nested go io limit");
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

verbose_output="$(VERBOSE=1 run_scheduler "$success_dir" "$success_manifest" verbose --summary-targets local,service,meta --summary-groups "check-work=local,service,meta" --resource-limit cpu=2 --resource-limit io=3 2>&1)"
assert_contains "$verbose_output" "[CHECK-SCHEDULER] check start work_unit=setup claims={cpu:1} active=1 pending=4" "verbose scheduler start telemetry"
assert_contains "$verbose_output" "active_resource_claims={cpu:1}" "verbose scheduler active resource telemetry"
assert_contains "$verbose_output" "resource_limits={cpu:2,io:3,service_stack:1}" "verbose scheduler resource limit telemetry"

makeflags_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-makeflags.XXXXXX")"
cleanup_paths+=("$makeflags_dir")
write_fake_make "$makeflags_dir"
makeflags_manifest="${makeflags_dir}/manifest.json"
cat >"$makeflags_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v2",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "cpu": 1 },
      "work_units": [
        { "target": "alpha", "weight": 1, "needs": [], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" }
      ]
    }
  ]
}
JSON
makeflags_output="$(
  MAKEFLAGS='--jobserver-auth=3,4 -j --trace' \
  MFLAGS='--jobserver-fds=3,4 -j' \
    run_scheduler "$makeflags_dir" "$makeflags_manifest" makeflags --summary-targets alpha --resource-limit cpu=1 2>&1
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
  "schema_id": "cartulary.check_schedule.v2",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "cpu": 2, "service_stack": 1 },
      "work_units": [
        { "target": "alpha", "weight": 30, "needs": [], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" },
        { "target": "beta", "weight": 20, "needs": [], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" },
        { "target": "gamma", "weight": 10, "needs": [], "skipped_summary_targets": ["external-summary"], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" },
        { "target": "delta", "weight": 5, "needs": ["beta"], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" }
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
  run_scheduler "$failure_dir" "$failure_manifest" failure --summary-targets alpha,beta,gamma,external-summary --resource-limit cpu=2 2>&1
)"
failure_status=$?
set -e
assert_equals "$failure_status" "7" "failure exit status"
assert_contains "$failure_output" "fake failure for beta" "failure child output"
assert_contains "$failure_output" "[FAIL] check" "failure summary"
assert_contains "$failure_output" "[CHECK-SCHEDULER] check summary status=fail" "failure scheduler status summary"
assert_contains "$failure_output" "failed=beta" "failure scheduler failed work unit"
failure_events="$(cat "${failure_dir}/events.log")"
assert_contains "$failure_events" "start alpha" "failure alpha started"
assert_contains "$failure_events" "start beta" "failure beta started"
assert_contains "$failure_events" "end alpha" "failure alpha drained"
failure_summary="${failure_dir}/results/failure/run-summary.json"
assert_equals "$(json_field "$failure_summary" "status")" "fail" "failure summary status"
assert_equals "$(json_field "$failure_summary" "aborted_after")" "beta" "failure aborted after"
assert_equals "$(json_field "$failure_summary" "skipped_after_failure.0")" "gamma" "failure skipped target"
assert_equals "$(json_field "$failure_summary" "skipped_after_failure.1")" "external-summary" "failure skipped mapped summary target"
"$NODE_BIN" - "$failure_summary" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.missing_target_summaries.includes("external-summary")) {
  throw new Error("mapped skipped summary target must not be reported missing");
}
EOF
failure_scheduler_summary="${failure_dir}/results/failure/check/scheduler-summary.json"
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
  "schema_id": "cartulary.check_schedule.v2",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "cpu": 1 },
      "work_units": [
        { "target": "alpha", "weight": 1, "needs": ["missing"], "resource_claims": { "cpu": 1 } }
      ]
    }
  ]
}
JSON
set +e
invalid_output="$(run_scheduler "$invalid_dir" "$invalid_manifest" invalid --summary-targets alpha 2>&1)"
invalid_status=$?
set -e
assert_equals "$invalid_status" "1" "invalid dependency status"
assert_contains "$invalid_output" "depends on unknown target missing" "invalid dependency output"

invalid_nested_manifest="${invalid_dir}/invalid-nested-manifest.json"
cat >"$invalid_nested_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v2",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "cpu": 1, "io": 1 },
      "work_units": [
        {
          "target": "service",
          "weight": 1,
          "needs": [],
          "resource_claims": { "cpu": 1 },
          "nested_scheduler": {
            "type": "service_backed",
            "target": "service",
            "manifest": "tools/service_backed_schedule_manifest.json",
            "resource_limit_env": {
              "io": "CARTULARY_SERVICE_BACKED_GO_IO_LIMIT"
            }
          }
        }
      ]
    }
  ]
}
JSON
set +e
invalid_nested_output="$(run_scheduler "$invalid_dir" "$invalid_nested_manifest" invalid-nested --summary-targets service 2>&1)"
invalid_nested_status=$?
set -e
assert_equals "$invalid_nested_status" "1" "invalid nested scheduler status"
assert_contains "$invalid_nested_output" "nested_scheduler.resource_limit_env.io must map a resource claimed by service" "invalid nested scheduler output"

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-dry-run.XXXXXX")"
cleanup_paths+=("$dry_run_dir")
write_fake_make "$dry_run_dir"
dry_run_output="$(
  MAKEFLAGS=n \
    run_scheduler "$dry_run_dir" "$success_manifest" dry-run --summary-targets local,service,meta --resource-limit cpu=2 2>&1
)"
assert_contains "$dry_run_output" "[DRY-RUN] check manifest=" "dry-run output"
assert_contains "$dry_run_output" "resource_limits={cpu:2,io:3,service_stack:1} work_units=5 dependencies=4 top_weighted=setup:50,build:40,local:30,service:20,meta:10" "dry-run compact summary"
assert_not_contains "$dry_run_output" "[DRY-RUN] check unit" "dry-run default hides unit expansion"
assert_not_contains "$dry_run_output" "claims={" "dry-run default hides raw claims"
assert_file_absent "${dry_run_dir}/make-args.log" "dry-run child make"

dry_run_verbose_output="$(
  MAKEFLAGS=n VERBOSE=1 \
    run_scheduler "$dry_run_dir" "$success_manifest" dry-run-verbose --summary-targets local,service,meta --resource-limit cpu=2 2>&1
)"
assert_contains "$dry_run_verbose_output" "[DRY-RUN] check unit setup needs=none claims={cpu:1} make_jobs=1" "verbose dry-run includes unit claims"
assert_contains "$dry_run_verbose_output" "nested_scheduler={\"type\":\"service_backed\",\"target\":\"service\"" "verbose dry-run includes nested scheduler metadata"
