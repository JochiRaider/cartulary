#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT_DIR}/scripts/run-service-backed-schedule.mjs"
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
  local summary_file="${dir}/results/${run_id}/${target}/scheduler-summary.json"
  local events_file="${dir}/results/${run_id}/${target}/scheduler-events.jsonl"

  assert_file_present "$summary_file" "$target scheduler summary"
  assert_file_present "$events_file" "$target scheduler events"
  "$NODE_BIN" - "$summary_file" "$events_file" "$expected_status" "$expected_blocked" "$expected_event" <<'EOF'
const fs = require("node:fs");
const [summaryFile, eventsFile, expectedStatus, expectedBlocked, expectedEvent] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).filter(Boolean).map((line) => JSON.parse(line));
if (summary.schema_id !== "cartulary.service_backed_scheduler_summary.v4") {
  throw new Error(`unexpected summary schema ${summary.schema_id}`);
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
if (!Array.isArray(summary.slowest_work_units) || summary.slowest_work_units.length === 0) {
  throw new Error("summary must record slowest work");
}
if (!summary.artifacts?.events_jsonl || !summary.artifacts?.scheduler_logs_dir) {
  throw new Error("summary must record scheduler artifact paths");
}
if (!Number.isInteger(summary.max_running_groups) || summary.max_running_groups < 1) {
  throw new Error(`summary max_running_groups got ${summary.max_running_groups}`);
}
if (expectedBlocked !== "-") {
  if (!summary.blocked_resources_seen.includes(expectedBlocked)) {
    throw new Error(`summary missing blocked resource ${expectedBlocked}`);
  }
  if (!summary.blocked_explanations_seen.includes(expectedBlocked)) {
    throw new Error(`summary missing blocked explanation ${expectedBlocked}`);
  }
}
if (events.length === 0) {
  throw new Error("scheduler events must not be empty");
}
if (!events.every((event) => event.schema_id === "cartulary.service_backed_scheduler_event.v4")) {
  throw new Error("unexpected scheduler event schema");
}
if (expectedEvent !== "-" && !events.some((event) => event.event === expectedEvent)) {
  throw new Error(`missing scheduler event ${expectedEvent}`);
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
  echo "fake failure for $target" >&2
  exit 7
fi

echo "fake pass for $target"
write_summary
EOF
  chmod +x "${dir}/fake-make"
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
    if [[ "$#" -ne 3 ]]; then
      echo "usage: fake-go-target finalize-shards <target> <metadata-dir>" >&2
      exit 2
    fi

    target="$2"
    metadata_dir="$3"
    aggregate_dir="${metadata_dir}/aggregate-reports/${target}/fake-aggregate"
    status=0
    summary_status=pass

    log_event "start finalize ${target}"
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

write_manifest() {
  local file="$1"
  local target="$2"
  shift 2

  {
    printf '{\n'
    printf '  "schema_id": "cartulary.service_backed_schedule.v8",\n'
    printf '  "schedules": [\n'
    printf '    { "target": "%s", "resource_limits": { "postgres": 32, "minio": 32, "go_cpu": 6, "go_io": 6, "process": 2, "browser_stack": "auto", "browser_stage_webserver_backed": 1, "browser_stage_isolated": 1, "browser_stage_visual": 1 }, "work_unit_sources": [\n' "$target"
    local first=1
    local source
    for source in "$@"; do
      IFS='|' read -r type name weight claims class browser_stage needs <<<"$source"
      class="${class:-backend}"
      if [[ "$first" -eq 0 ]]; then
        printf ',\n'
      fi
      first=0
      if [[ "$type" == "make_target" ]]; then
        printf '      { "type": "make_target", "class": "%s", "target": "%s", "weight": %s, "resource_claims": {%s}' \
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
  } >"$file"
}

assert_no_shared_backend_integration_shards() {
  "$NODE_BIN" - "$ROOT_DIR" <<'EOF'
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const [root] = process.argv.slice(2);
const shardPlanScript = path.join(root, "scripts/lib/go-shard-plan.mjs");
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
      "resource_limits": { "postgres": 32, "minio": 32, "backend": 4 },
      "children": [
        { "target": "backend-integration", "kind": "backend", "weight": 1, "resource_claims": ["postgres", "minio", "backend"] }
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

  if [[ -x "${dir}/fake-go-target" ]]; then
    go_target_runner="${dir}/fake-go-target"
  fi

  FAKE_SCHEDULER_LOCK="${dir}/lock" \
  FAKE_SCHEDULER_ACTIVE="${dir}/active" \
  FAKE_SCHEDULER_MAX="${dir}/max" \
  FAKE_SCHEDULER_LOG="${dir}/make.log" \
  FAKE_FAIL_TARGET="${FAKE_FAIL_TARGET:-}" \
  FAKE_SCHEDULER_SLEEP="${FAKE_SCHEDULER_SLEEP:-0.2}" \
  FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS="${FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS:-}" \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED="${FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED:-}" \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E="${FAKE_SCHEDULER_SLEEP_BROWSER_E2E:-}" \
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
  VERBOSE="${VERBOSE:-}" \
  CI_VERBOSE="${CI_VERBOSE:-}" \
  CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT= \
  CARTULARY_SERVICE_BACKED_GO_IO_LIMIT= \
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT="${CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT:-}" \
  MAKE="${dir}/fake-make" \
  CARTULARY_TEST_GO_TARGET_RUNNER="${go_target_runner}" \
  NODE_BIN="$NODE_BIN" \
  TEST_OUTPUT_SCRIPT="$TEST_OUTPUT_SCRIPT" \
  CARTULARY_TEST_RESULTS_DIR="${dir}/results" \
  CARTULARY_TEST_RUN_ID="$run_id" \
    "$NODE_BIN" "$SCRIPT" --target "$target" --manifest "$manifest" "$@"
}

weighted_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-weighted.XXXXXX")"
cleanup_paths+=("$weighted_dir")
write_fake_make "$weighted_dir"
weighted_manifest="${weighted_dir}/manifest.json"
write_manifest "$weighted_manifest" test-fast-service-backed \
  'make_target|backend-store|1|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1' \
  'make_target|backend-process|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1' \
  'make_target|backend-integration-support|5|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1'
weighted_output="$(run_scheduler "$weighted_dir" "$weighted_manifest" test-fast-service-backed weighted 2>&1)"
assert_not_contains "$weighted_output" "[STEP] test-fast-service-backed" "default service scheduler output hides per-unit steps"
assert_contains "$weighted_output" "[SCHEDULER] test-fast-service-backed start work_units=3 finalizers=0 capacity={go_cpu:6,go_io:6,browser_stack:" "quiet scheduler shows aggregate start"
assert_contains "$weighted_output" "classes={backend:3}" "quiet scheduler start shows work classes"
assert_contains "$weighted_output" "types={make_target:3}" "quiet scheduler start shows work types"
assert_contains "$weighted_output" "top_weighted=backend-process:10,backend-integration-support:5,backend-store:1" "quiet scheduler start shows top weighted work"
assert_contains "$weighted_output" "artifacts=tmp/service-backed-scheduler-weighted." "quiet scheduler start shows artifact field"
assert_contains "$weighted_output" "/results/weighted/test-fast-service-backed" "quiet scheduler start shows artifact path"
assert_not_contains "$weighted_output" "[SCHEDULER] test-fast-service-backed progress completed_work_units=0/3" "quiet scheduler hides immediate unblocked progress"
assert_contains "$weighted_output" "[SCHEDULER] test-fast-service-backed summary status=pass completed_work_units=3/3 failed=none slowest=" "quiet scheduler shows pass summary"
assert_contains "$weighted_output" "/results/weighted/test-fast-service-backed" "quiet scheduler summary shows artifact path"
assert_not_contains "$weighted_output" "fake pass for backend-store" "quiet scheduler hides successful child logs"
assert_not_contains "$weighted_output" "active_resource_claims=" "default scheduler output hides raw active resources"
assert_not_contains "$weighted_output" "claims={" "default scheduler output hides raw claims"
assert_not_contains "$weighted_output" "running_units=" "default scheduler output hides raw running units"
assert_not_contains "$weighted_output" "blocked_resources=" "default scheduler output hides raw blocked resources"
assert_scheduler_artifacts "$weighted_dir" weighted test-fast-service-backed pass - start

weighted_verbose_output="$(VERBOSE=1 run_scheduler "$weighted_dir" "$weighted_manifest" test-fast-service-backed weighted-verbose 2>&1)"
assert_contains "$weighted_verbose_output" "[SCHEDULER] test-fast-service-backed start work_units=3 finalizers=0 capacity={go_cpu:6,go_io:6,browser_stack:" "verbose scheduler aggregate start"
assert_contains "$weighted_verbose_output" "[SCHEDULER] test-fast-service-backed start work_unit=backend-process claims={go_cpu:1,go_io:1,minio:1,postgres:1,process:1} active=1 pending=2" "verbose scheduler start telemetry"
assert_contains "$weighted_verbose_output" "[SCHEDULER] test-fast-service-backed start work_unit=backend-store claims={go_cpu:1,go_io:1,minio:1,postgres:1} active=3 pending=0 active_resource_claims={go_cpu:3,go_io:3,minio:3,postgres:3,process:1}" "verbose scheduler starts all compatible weighted children"
assert_contains "$weighted_verbose_output" "resource_limits={go_cpu:6,go_io:6,browser_stack:" "verbose scheduler resource limit telemetry includes browser stack"
assert_contains "$weighted_verbose_output" "[SCHEDULER] test-fast-service-backed finish work_unit=backend-process status=0" "verbose scheduler finish telemetry"
assert_contains "$weighted_verbose_output" "fake pass for backend-store" "verbose scheduler replays successful child logs"

resource_block_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-resource-block.XXXXXX")"
cleanup_paths+=("$resource_block_dir")
write_fake_make "$resource_block_dir"
resource_block_manifest="${resource_block_dir}/manifest.json"
write_manifest "$resource_block_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1, "go_cpu": 4, "go_io": 1' \
  'make_target|backend-store|9|"postgres": 1, "minio": 1, "go_cpu": 3, "go_io": 1' \
  'make_target|backend-process|8|"postgres": 1, "minio": 1, "go_cpu": 2, "go_io": 1, "process": 1'
resource_block_output="$(run_scheduler "$resource_block_dir" "$resource_block_manifest" test-fast-service-backed resource-block 2>&1)"
assert_contains "$resource_block_output" "blocked_by=go_cpu unblocks_after=backend-integration" "scheduler go_cpu-blocked progress"
assert_not_contains "$resource_block_output" "blocked_resources=go_cpu" "default go_cpu-blocked output hides raw blocked resources"
assert_not_contains "$resource_block_output" "active_resource_claims=" "default blocked output hides raw active resources"
assert_scheduler_artifacts "$resource_block_dir" resource-block test-fast-service-backed pass go_cpu blocked

backend_capacity_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-backend-capacity.XXXXXX")"
cleanup_paths+=("$backend_capacity_dir")
write_fake_make "$backend_capacity_dir"
backend_capacity_manifest="${backend_capacity_dir}/manifest.json"
write_manifest "$backend_capacity_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1' \
  'make_target|backend-store|9|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1' \
  'make_target|backend-process|8|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1' \
  'make_target|backend-integration-support|7|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1'
backend_capacity_output="$(run_scheduler "$backend_capacity_dir" "$backend_capacity_manifest" test-fast-service-backed backend-capacity 2>&1)"
assert_contains "$backend_capacity_output" "[SCHEDULER] test-fast-service-backed summary status=pass" "quiet go resource model shows success scheduler summary"

io_block_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-host_io-block.XXXXXX")"
cleanup_paths+=("$io_block_dir")
write_fake_make "$io_block_dir"
io_block_manifest="${io_block_dir}/manifest.json"
write_manifest "$io_block_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 4' \
  'make_target|backend-store|9|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 3' \
  'make_target|backend-process|8|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 2, "process": 1'
io_block_output="$(run_scheduler "$io_block_dir" "$io_block_manifest" test-fast-service-backed host_io-block 2>&1)"
assert_contains "$io_block_output" "blocked_by=go_io unblocks_after=backend-integration" "scheduler go_io-blocked progress"

browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-browser.XXXXXX")"
cleanup_paths+=("$browser_dir")
write_fake_make "$browser_dir"
browser_manifest="${browser_dir}/manifest.json"
write_manifest "$browser_manifest" test-service-backed \
  'make_target|backend-process|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend' \
  'make_target|browser-e2e-webserver-backed|9|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed'
browser_output="$(run_scheduler "$browser_dir" "$browser_manifest" test-service-backed browser 2>&1)"
assert_not_contains "$browser_output" "[STEP] test-service-backed" "browser schedule hides default scheduler steps"
assert_contains "$browser_output" "[PASS] test-service-backed kind=aggregate children=2/2 child_tests=2 child_failed=0" "browser schedule aggregate child tests"
assert_contains "$browser_output" "[SCHEDULER] test-service-backed start work_units=2 finalizers=0" "browser quiet scheduler shows aggregate start"
assert_contains "$browser_output" "classes={backend:1,browser:1}" "browser quiet scheduler start shows classes"
assert_contains "$browser_output" "top_weighted=backend-process:10,browser-e2e-webserver-backed:9" "browser quiet scheduler start shows top weighted work"
assert_contains "$browser_output" "[SCHEDULER] test-service-backed summary status=pass" "browser quiet scheduler shows success summary"
assert_not_contains "$browser_output" "claims={browser_stack:1" "browser default output hides resource claims"
assert_scheduler_artifacts "$browser_dir" browser test-service-backed pass - start

eager_finalizer_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-eager-finalizer.XXXXXX")"
cleanup_paths+=("$eager_finalizer_dir")
write_fake_make "$eager_finalizer_dir"
write_fake_go_target_runner "$eager_finalizer_dir"
eager_finalizer_manifest="${eager_finalizer_dir}/manifest.json"
cat >"$eager_finalizer_manifest" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule.v8",
  "schedules": [
    {
      "target": "test-service-backed",
      "resource_limits": { "postgres": 32, "minio": 32, "go_cpu": 64, "go_io": 64, "process": 2, "browser_stack": "auto", "browser_stage_webserver_backed": 1 },
      "work_unit_sources": [
        { "type": "go_shards", "class": "backend", "target": "backend-store", "resource_claims": { "postgres": 1, "minio": 1 } },
        { "type": "make_target", "class": "browser", "target": "browser-e2e-webserver-backed", "browser_stage": "webserver-backed", "weight": 9, "resource_claims": { "postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1 } }
      ]
    }
  ]
}
JSON
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
const browserEnd = indexOf((event) => event.event === "finish" && event.work_unit === "browser-e2e-webserver-backed", "browser finish");
if (!(finalizeStart < browserEnd)) {
  throw new Error("backend-store finalizer waited for browser tail");
}
EOF

separate_finalizer_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-separate-finalizer.XXXXXX")"
cleanup_paths+=("$separate_finalizer_dir")
write_fake_make "$separate_finalizer_dir"
write_fake_go_target_runner "$separate_finalizer_dir"
separate_finalizer_manifest="${separate_finalizer_dir}/manifest.json"
write_manifest "$separate_finalizer_manifest" test-fast-service-backed \
  'go_shards|backend-integration|0|"postgres": 1, "minio": 1' \
  'go_shards|backend-integration-support|0|"postgres": 1, "minio": 1'
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
  'make_target|browser-e2e-webserver-backed|30|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed' \
  'make_target|backend-process|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend'
check_browser_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2
  FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS=0.3
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.05
  run_scheduler "$check_browser_dir" "$check_browser_manifest" check-service-backed check-browser 2>&1
)"
assert_not_contains "$check_browser_output" "[STEP] check-service-backed" "check browser schedule hides default scheduler steps"
assert_contains "$check_browser_output" "[PASS] check-service-backed kind=aggregate children=2/2 child_tests=2 child_failed=0" "check browser aggregate child tests"
check_browser_events="$(cat "${check_browser_dir}/make.log")"
assert_contains "$check_browser_events" "start browser-e2e-webserver-backed" "check browser webserver start"
assert_contains "$check_browser_events" "end browser-e2e-webserver-backed" "check browser webserver end"
assert_contains "$check_browser_events" "end backend-process" "check browser backend end"

dual_browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-dual-browser.XXXXXX")"
cleanup_paths+=("$dual_browser_dir")
write_fake_make "$dual_browser_dir"
dual_browser_manifest="${dual_browser_dir}/manifest.json"
write_manifest "$dual_browser_manifest" check-service-backed \
  'make_target|browser-e2e-webserver-backed|30|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed' \
  'make_target|browser-e2e|20|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_isolated": 1|browser|isolated'
dual_browser_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.2 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E=0.2 \
    run_scheduler "$dual_browser_dir" "$dual_browser_manifest" check-service-backed dual-browser 2>&1
)"
assert_contains "$dual_browser_output" "[PASS] check-service-backed kind=aggregate children=2/2 child_tests=2 child_failed=0" "dual browser aggregate child tests"
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
const isolatedStart = indexOf("start browser-e2e ");
const isolatedEnd = indexOf("end browser-e2e ");
if (!(webStart < isolatedEnd && isolatedStart < webEnd)) {
  throw new Error("distinct browser stages did not overlap when browser_stack capacity allowed it");
}
EOF

dependency_order_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-dependency-order.XXXXXX")"
cleanup_paths+=("$dependency_order_dir")
write_fake_make "$dependency_order_dir"
dependency_order_manifest="${dependency_order_dir}/manifest.json"
write_manifest "$dependency_order_manifest" check-service-backed \
  'make_target|backend-process|30|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend||' \
  'make_target|browser-e2e-webserver-backed|20|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed|' \
  'make_target|browser-e2e|10|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_isolated": 1|browser|isolated|backend-process,browser-e2e-webserver-backed'
dependency_order_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2 \
  FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS=0.15 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.05 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E=0.01 \
    run_scheduler "$dependency_order_dir" "$dependency_order_manifest" check-service-backed dependency-order 2>&1
)"
assert_contains "$dependency_order_output" "blocked_by=dependencies waiting_on=backend-process,browser-e2e-webserver-backed unblocks_after=backend-process" "dependency-blocked browser progress"
assert_scheduler_artifacts "$dependency_order_dir" dependency-order check-service-backed pass - blocked
"$NODE_BIN" - "${dependency_order_dir}/results/dependency-order/check-service-backed/scheduler-events.jsonl" "${dependency_order_dir}/results/dependency-order/check-service-backed/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [eventsFile, summaryFile] = process.argv.slice(2);
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).map((line) => JSON.parse(line));
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (!summary.waiting_on_seen?.includes("backend-process") || !summary.waiting_on_seen?.includes("browser-e2e-webserver-backed")) {
  throw new Error("summary must record service-backed waiting_on targets");
}
const dependencyBlocked = events.find((event) => event.event === "blocked" && event.blocked_reason === "dependencies");
if (!dependencyBlocked) {
  throw new Error("missing service-backed dependency blocked event");
}
if (!dependencyBlocked.waiting_on?.includes("backend-process") || !dependencyBlocked.waiting_on?.includes("browser-e2e-webserver-backed")) {
  throw new Error("dependency blocked event must record direct waiting_on targets");
}
const browserBlocked = dependencyBlocked.blocked_units?.find((entry) => entry.work_unit === "browser-e2e");
if (!browserBlocked?.waiting_on?.includes("backend-process") || !browserBlocked?.waiting_on?.includes("browser-e2e-webserver-backed")) {
  throw new Error("blocked_units must record browser-e2e direct dependencies");
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
const isolatedStart = indexOf("start browser-e2e ");
if (!(backendEnd < isolatedStart && webEnd < isolatedStart)) {
  throw new Error("browser-e2e started before declared dependencies completed");
}
if (!lines.some((line) => line.includes("args --no-print-directory --output-sync=target -j1 browser-e2e"))) {
  throw new Error("service-backed make_target children must run with explicit -j1");
}
EOF

makeflags_sanitize_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-makeflags.XXXXXX")"
cleanup_paths+=("$makeflags_sanitize_dir")
write_fake_make "$makeflags_sanitize_dir"
makeflags_sanitize_manifest="${makeflags_sanitize_dir}/manifest.json"
write_manifest "$makeflags_sanitize_manifest" test-fast-service-backed \
  'make_target|backend-process|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1'
makeflags_sanitize_output="$(
  MAKEFLAGS='--jobserver-auth=3,4 -j --trace' \
  MFLAGS='--jobserver-fds=3,4 -j' \
    run_scheduler "$makeflags_sanitize_dir" "$makeflags_sanitize_manifest" test-fast-service-backed makeflags-sanitize 2>&1
)"
assert_contains "$makeflags_sanitize_output" "[SCHEDULER] test-fast-service-backed summary status=pass" "makeflags sanitize quiet scheduler shows success summary"
assert_not_contains "$(cat "${makeflags_sanitize_dir}/make.log")" "jobserver" "child make env strips inherited jobserver tokens"
assert_not_contains "$(cat "${makeflags_sanitize_dir}/make.log")" "MFLAGS=-j" "child make env strips inherited mflags jobs"
assert_contains "$(cat "${makeflags_sanitize_dir}/make.log")" "MAKEFLAGS=--trace" "child make env preserves non-jobserver make flags"

go_dependency_order_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-go-dependency-order.XXXXXX")"
cleanup_paths+=("$go_dependency_order_dir")
write_fake_make "$go_dependency_order_dir"
write_fake_go_target_runner "$go_dependency_order_dir"
go_dependency_order_manifest="${go_dependency_order_dir}/manifest.json"
write_manifest "$go_dependency_order_manifest" check-service-backed \
  'go_shards|backend-store|0|"postgres": 1, "minio": 1|backend||' \
  'make_target|browser-e2e-webserver-backed|10|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed|backend-store'
go_dependency_order_output="$(
  FAKE_GO_SLEEP_CAPTURE=0.005 \
  FAKE_GO_SLEEP_FINALIZE=0.15 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.01 \
    run_scheduler "$go_dependency_order_dir" "$go_dependency_order_manifest" check-service-backed go-dependency-order 2>&1
)"
assert_contains "$go_dependency_order_output" "blocked_by=dependencies waiting_on=backend-store unblocks_after=backend-store" "go dependency waits for finalizer"
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
  'go_shards|backend-store|0|"postgres": 1, "minio": 1|backend||' \
  'make_target|browser-e2e-webserver-backed|10|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed|backend-store'
set +e
dependency_failure_skip_output="$(
  FAKE_GO_FAIL_SHARD=backend-store-shard-01 \
  FAKE_GO_FINALIZER_FAILURE_STATUS=9 \
    run_scheduler "$dependency_failure_skip_dir" "$dependency_failure_skip_manifest" check-service-backed dependency-failure 2>&1
)"
dependency_failure_skip_status=$?
set -e
assert_equals "$dependency_failure_skip_status" "9" "dependency failure status"
assert_contains "$dependency_failure_skip_output" "skipped=1" "dependency failure summary skipped dependent work"
assert_contains "$dependency_failure_skip_output" "[FAIL] check-service-backed" "dependency failure parent summary"
assert_scheduler_artifacts "$dependency_failure_skip_dir" dependency-failure check-service-backed fail - skip
"$NODE_BIN" - "${dependency_failure_skip_dir}/results/dependency-failure/check-service-backed/scheduler-events.jsonl" "${dependency_failure_skip_dir}/results/dependency-failure/check-service-backed/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [eventsFile, summaryFile] = process.argv.slice(2);
const events = fs.readFileSync(eventsFile, "utf8").trim().split(/\n/).map((line) => JSON.parse(line));
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (!events.some((event) => event.event === "skip" && event.work_unit === "browser-e2e-webserver-backed" && event.skip_reason === "dependency_failure" && event.failed_dependency === "backend-store")) {
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
  'make_target|browser-e2e-webserver-backed|30|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed' \
  'make_target|browser-e2e|20|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_isolated": 1|browser|isolated'
browser_stack_lane_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=1 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.05 \
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E=0.05 \
    run_scheduler "$browser_stack_lane_dir" "$browser_stack_lane_manifest" check-service-backed browser-stack-lane 2>&1
)"
assert_contains "$browser_stack_lane_output" "blocked_by=browser_stack unblocks_after=browser-e2e-webserver-backed" "shared browser stack blocks overlapping browser stages"

same_browser_lane_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-same-browser-lane.XXXXXX")"
cleanup_paths+=("$same_browser_lane_dir")
write_fake_make "$same_browser_lane_dir"
same_browser_lane_manifest="${same_browser_lane_dir}/manifest.json"
write_manifest "$same_browser_lane_manifest" test-fast-service-backed \
  'make_target|backend-process|30|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1, "browser_stage_isolated": 1|backend' \
  'make_target|backend-store|20|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "browser_stage_isolated": 1|backend'
same_browser_lane_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2 \
  FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS=0.05 \
  FAKE_SCHEDULER_SLEEP_BACKEND_STORE=0.05 \
    run_scheduler "$same_browser_lane_dir" "$same_browser_lane_manifest" test-fast-service-backed same-browser-lane 2>&1
)"
assert_contains "$same_browser_lane_output" "blocked_by=browser_stage_isolated unblocks_after=backend-process" "same browser stage lane blocks overlapping work"
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
  throw new Error("same browser_stage_isolated lane work overlapped");
}
EOF

set +e
empty_budget_output="$("$NODE_BIN" "${ROOT_DIR}/scripts/check-postgres-fixture-budget.mjs" --targets "" 2>&1)"
empty_budget_status=$?
set -e
assert_equals "$empty_budget_status" "0" "empty postgres fixture budget target list status"
assert_equals "$empty_budget_output" "" "empty postgres fixture budget target list output"

failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-failure.XXXXXX")"
cleanup_paths+=("$failure_dir")
write_fake_make "$failure_dir"
failure_manifest="${failure_dir}/manifest.json"
write_manifest "$failure_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1' \
  'make_target|backend-store|9|"postgres": 1, "minio": 1'
set +e
failure_output="$(
  FAKE_FAIL_TARGET=backend-store \
    run_scheduler "$failure_dir" "$failure_manifest" test-fast-service-backed failure 2>&1
)"
failure_status=$?
set -e
assert_equals "$failure_status" "7" "child failure status"
assert_contains "$failure_output" "fake failure for backend-store" "child failure output"
assert_contains "$failure_output" "[SCHEDULER] test-fast-service-backed summary status=fail failure_class=helper completed_work_units=2/2 failed=backend-store" "failure scheduler summary"
assert_contains "$failure_output" "[FAIL] test-fast-service-backed" "failure target summary"
assert_scheduler_artifacts "$failure_dir" failure test-fast-service-backed fail - finish

failed_shard_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-failed-shard.XXXXXX")"
cleanup_paths+=("$failed_shard_dir")
write_fake_make "$failed_shard_dir"
write_fake_go_target_runner "$failed_shard_dir"
failed_shard_manifest="${failed_shard_dir}/manifest.json"
write_manifest "$failed_shard_manifest" test-fast-service-backed \
  'go_shards|backend-store|0|"postgres": 1, "minio": 1'
set +e
failed_shard_output="$(
  FAKE_GO_FAIL_SHARD=backend-store-shard-01 \
  FAKE_GO_FINALIZER_FAILURE_STATUS=9 \
    run_scheduler "$failed_shard_dir" "$failed_shard_manifest" test-fast-service-backed failed-shard 2>&1
)"
failed_shard_status=$?
set -e
assert_equals "$failed_shard_status" "9" "failed shard finalizer status"
assert_contains "$failed_shard_output" "fake shard failure for backend-store-shard-01" "failed shard output"
assert_contains "$failed_shard_output" "[SCHEDULER] test-fast-service-backed summary status=fail failure_class=helper completed_work_units=1/1 failed=finalize/backend-store" "failed shard scheduler summary"
assert_contains "$failed_shard_output" "finalizer_failures=1" "failed shard scheduler finalizer failure count"
assert_contains "$failed_shard_output" "[FAIL] test-fast-service-backed" "failed shard parent summary"
assert_scheduler_artifacts "$failed_shard_dir" failed-shard test-fast-service-backed fail - finalize-finish
"$NODE_BIN" - "${failed_shard_dir}/results/failed-shard/test-fast-service-backed/scheduler-summary.json" <<'EOF'
const fs = require("node:fs");
const [summaryFile] = process.argv.slice(2);
const summary = JSON.parse(fs.readFileSync(summaryFile, "utf8"));
if (summary.finalizer_failures !== 1) {
  throw new Error(`expected one finalizer failure, got ${summary.finalizer_failures}`);
}
EOF

defer_summary_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-defer-summary.XXXXXX")"
cleanup_paths+=("$defer_summary_dir")
write_fake_make "$defer_summary_dir"
defer_summary_manifest="${defer_summary_dir}/manifest.json"
write_manifest "$defer_summary_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1' \
  'make_target|backend-store|9|"postgres": 1, "minio": 1'
defer_summary_output="$(run_scheduler "$defer_summary_dir" "$defer_summary_manifest" test-fast-service-backed defer-summary --defer-summary 2>&1)"
assert_not_contains "$defer_summary_output" "[TARGET] start test-fast-service-backed" "defer-summary quiet output hides target start"
assert_file_absent "${defer_summary_dir}/results/defer-summary/test-fast-service-backed/target-summary.json" "defer-summary parent target summary"

unsafe_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-unsafe.XXXXXX")"
cleanup_paths+=("$unsafe_dir")
write_fake_make "$unsafe_dir"
unsafe_manifest="${unsafe_dir}/manifest.json"
write_manifest "$unsafe_manifest" check-service-backed \
  'make_target|backend-unit|10|"postgres": 1, "minio": 1'
set +e
unsafe_output="$(run_scheduler "$unsafe_dir" "$unsafe_manifest" check-service-backed unsafe 2>&1)"
unsafe_status=$?
set -e
assert_equals "$unsafe_status" "1" "unsafe manifest status"
assert_contains "$unsafe_output" "is not service-backed" "unsafe manifest output"

unknown_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-unknown.XXXXXX")"
cleanup_paths+=("$unknown_dir")
write_fake_make "$unknown_dir"
unknown_manifest="${unknown_dir}/manifest.json"
write_manifest "$unknown_manifest" test-fast-service-backed \
  'make_target|unknown-backend-target|10|"postgres": 1, "minio": 1'
set +e
unknown_output="$(run_scheduler "$unknown_dir" "$unknown_manifest" test-fast-service-backed unknown 2>&1)"
unknown_status=$?
set -e
assert_equals "$unknown_status" "1" "unknown manifest status"
assert_contains "$unknown_output" "is not in target-plan" "unknown manifest output"

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-dry-run.XXXXXX")"
cleanup_paths+=("$dry_run_dir")
write_fake_make "$dry_run_dir"
dry_run_manifest="${dry_run_dir}/manifest.json"
write_manifest "$dry_run_manifest" test-service-backed \
  'make_target|browser-e2e-webserver-backed|10|"postgres": 1, "minio": 1, "process": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed'
dry_run_output="$(
  CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT=2 \
  MAKEFLAGS=n \
    run_scheduler "$dry_run_dir" "$dry_run_manifest" test-service-backed dry-run 2>&1
)"
assert_contains "$dry_run_output" "[DRY-RUN] test-service-backed manifest=" "dry-run output"
assert_contains "$dry_run_output" "resource_limits={go_cpu:6,go_io:6,browser_stack:2,minio:32,postgres:32,process:2" "dry-run includes compact resolved resources"
assert_contains "$dry_run_output" "work_units=1 dependencies=0 classes={browser:1} types={make_target:1} finalizers=0 top_weighted=browser-e2e-webserver-backed:10" "dry-run includes compact work summary"
assert_not_contains "$dry_run_output" "claims={" "default dry-run hides per-unit claims"
assert_file_absent "${dry_run_dir}/make.log" "dry-run child make log"

go_shard_dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-go-shard-dry-run.XXXXXX")"
cleanup_paths+=("$go_shard_dry_run_dir")
write_fake_make "$go_shard_dry_run_dir"
go_shard_dry_run_manifest="${go_shard_dry_run_dir}/manifest.json"
write_manifest "$go_shard_dry_run_manifest" test-fast-service-backed \
  'go_shards|backend-store|0|"postgres": 1, "minio": 1'
go_shard_dry_run_output="$(
  VERBOSE=1 \
  MAKEFLAGS=n \
    run_scheduler "$go_shard_dry_run_dir" "$go_shard_dry_run_manifest" test-fast-service-backed go-shard-dry-run 2>&1
)"
expected_go_shard_dry_run_line="$(
  "$NODE_BIN" - "$ROOT_DIR" <<'EOF'
const { execFileSync } = require("node:child_process");
const path = require("node:path");
const [root] = process.argv.slice(2);
const plan = JSON.parse(execFileSync(process.execPath, [path.join(root, "scripts/lib/go-shard-plan.mjs"), "json"], { encoding: "utf8", cwd: root }));
const shard = plan.shards.find((candidate) => candidate.name === "backend-store-shard-01");
if (!shard) {
  process.exit(1);
}
const claimsByProfile = {
  balanced: "{go_cpu:1,go_io:1,minio:1,postgres:1}",
  cpu_heavy: "{go_cpu:2,go_io:1,minio:1,postgres:1}",
  io_heavy: "{go_cpu:1,go_io:2,minio:1,postgres:1}",
  reset_heavy: "{go_cpu:1,go_io:3,minio:1,postgres:1,postgres_reset:1}",
};
process.stdout.write(`backend-store/${shard.name} type=go_shard class=backend profile=${shard.scheduler_profile} claims=${claimsByProfile[shard.scheduler_profile]}`);
EOF
)"
assert_contains "$go_shard_dry_run_output" "work_units=1 dependencies=0 classes={backend:1} types={go_shard:1} finalizers=1" "go_shards dry-run compact summary"
assert_contains "$go_shard_dry_run_output" "$expected_go_shard_dry_run_line" "verbose go_shards dry-run includes per-shard resource claims"

rendered_schedule_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-rendered.XXXXXX")"
cleanup_paths+=("$rendered_schedule_dir")
rendered_schedule_manifest="${rendered_schedule_dir}/service-backed.json"
"$NODE_BIN" "$ROOT_DIR/scripts/render-service-backed-schedule-manifest.mjs" --output "$rendered_schedule_manifest"
"$NODE_BIN" - "$rendered_schedule_manifest" <<'EOF'
const fs = require("node:fs");
const manifest = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const byTarget = new Map((manifest.schedules ?? []).map((schedule) => [schedule.target, schedule]));
const sourceTargets = (target) => (byTarget.get(target)?.work_unit_sources ?? []).map((source) => source.target);
const testSources = sourceTargets("test-service-backed");
const checkSources = sourceTargets("check-service-backed");
const fastSources = sourceTargets("test-fast-service-backed");
if (JSON.stringify(testSources) !== JSON.stringify(checkSources)) {
  throw new Error(`test-service-backed and check-service-backed sources differ: ${testSources} vs ${checkSources}`);
}
if (fastSources.some((target) => target.startsWith("browser-e2e"))) {
  throw new Error(`test-fast-service-backed must remain backend-only, got ${fastSources.join(",")}`);
}
const isolated = (byTarget.get("test-service-backed")?.work_unit_sources ?? []).find(
  (source) => source.target === "browser-e2e",
);
const expectedNeeds = [
  "backend-integration",
  "backend-integration-support",
  "backend-store",
  "backend-process",
  "browser-e2e-webserver-backed",
];
if (JSON.stringify(isolated?.needs ?? []) !== JSON.stringify(expectedNeeds)) {
  throw new Error(`isolated browser needs got ${JSON.stringify(isolated?.needs ?? [])}`);
}
if (manifest.generated?.generator !== "scripts/render-service-backed-schedule-manifest.mjs") {
  throw new Error("rendered schedule must record generator metadata");
}
EOF

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
assert_equals "$invalid_resource_status" "1" "invalid resource manifest status"
assert_contains "$invalid_resource_output" "uses undeclared scheduler resource undeclared-resource" "invalid resource manifest output"

unknown_dependency_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-unknown-dependency.XXXXXX")"
cleanup_paths+=("$unknown_dependency_dir")
write_fake_make "$unknown_dependency_dir"
unknown_dependency_manifest="${unknown_dependency_dir}/manifest.json"
write_manifest "$unknown_dependency_manifest" test-fast-service-backed \
  'make_target|backend-process|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend||missing-target'
set +e
unknown_dependency_output="$(run_scheduler "$unknown_dependency_dir" "$unknown_dependency_manifest" test-fast-service-backed unknown-dependency 2>&1)"
unknown_dependency_status=$?
set -e
assert_equals "$unknown_dependency_status" "1" "unknown dependency manifest status"
assert_contains "$unknown_dependency_output" "depends on unknown target missing-target" "unknown dependency manifest output"

cycle_dependency_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-cycle-dependency.XXXXXX")"
cleanup_paths+=("$cycle_dependency_dir")
write_fake_make "$cycle_dependency_dir"
cycle_dependency_manifest="${cycle_dependency_dir}/manifest.json"
write_manifest "$cycle_dependency_manifest" test-fast-service-backed \
  'make_target|backend-process|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend||backend-store' \
  'make_target|backend-store|9|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1|backend||backend-process'
set +e
cycle_dependency_output="$(run_scheduler "$cycle_dependency_dir" "$cycle_dependency_manifest" test-fast-service-backed cycle-dependency 2>&1)"
cycle_dependency_status=$?
set -e
assert_equals "$cycle_dependency_status" "1" "cycle dependency manifest status"
assert_contains "$cycle_dependency_output" "has a dependency cycle" "cycle dependency manifest output"

removed_backend_resource_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-removed-backend-resource.XXXXXX")"
cleanup_paths+=("$removed_backend_resource_dir")
write_fake_make "$removed_backend_resource_dir"
removed_backend_resource_manifest="${removed_backend_resource_dir}/manifest.json"
write_manifest "$removed_backend_resource_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "backend": 1'
set +e
removed_backend_resource_output="$(run_scheduler "$removed_backend_resource_dir" "$removed_backend_resource_manifest" test-fast-service-backed removed-backend-resource 2>&1)"
removed_backend_resource_status=$?
set -e
assert_equals "$removed_backend_resource_status" "1" "removed backend resource manifest status"
assert_contains "$removed_backend_resource_output" "uses retired resource backend" "removed backend resource manifest output"

invalid_browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-invalid-browser.XXXXXX")"
cleanup_paths+=("$invalid_browser_dir")
write_fake_make "$invalid_browser_dir"
invalid_browser_manifest="${invalid_browser_dir}/manifest.json"
write_manifest "$invalid_browser_manifest" test-service-backed \
  'make_target|browser-e2e-webserver-backed|10|"postgres": 1, "minio": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|missing-stage'
set +e
invalid_browser_output="$(run_scheduler "$invalid_browser_dir" "$invalid_browser_manifest" test-service-backed invalid-browser 2>&1)"
invalid_browser_status=$?
set -e
assert_equals "$invalid_browser_status" "1" "invalid browser manifest status"
assert_contains "$invalid_browser_output" "declares unknown browser_stage missing-stage" "invalid browser manifest output"

invalid_browser_target_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-invalid-browser-target.XXXXXX")"
cleanup_paths+=("$invalid_browser_target_dir")
write_fake_make "$invalid_browser_target_dir"
invalid_browser_target_manifest="${invalid_browser_target_dir}/manifest.json"
write_manifest "$invalid_browser_target_manifest" test-service-backed \
  'make_target|browser-e2e-visual|10|"postgres": 1, "minio": 1, "browser_stack": 1, "browser_stage_webserver_backed": 1|browser|webserver-backed'
set +e
invalid_browser_target_output="$(run_scheduler "$invalid_browser_target_dir" "$invalid_browser_target_manifest" test-service-backed invalid-browser-target 2>&1)"
invalid_browser_target_status=$?
set -e
assert_equals "$invalid_browser_target_status" "1" "invalid browser target manifest status"
assert_contains "$invalid_browser_target_output" "must match browser_stage webserver-backed aggregate target browser-e2e-webserver-backed" "invalid browser target manifest output"

obsolete_browser_resource_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-obsolete-browser-resource.XXXXXX")"
cleanup_paths+=("$obsolete_browser_resource_dir")
write_fake_make "$obsolete_browser_resource_dir"
obsolete_browser_resource_manifest="${obsolete_browser_resource_dir}/manifest.json"
write_manifest "$obsolete_browser_resource_manifest" test-service-backed \
  'make_target|browser-e2e-webserver-backed|10|"postgres": 1, "minio": 1, "browser": 1|browser|webserver-backed'
set +e
obsolete_browser_resource_output="$(run_scheduler "$obsolete_browser_resource_dir" "$obsolete_browser_resource_manifest" test-service-backed obsolete-browser-resource 2>&1)"
obsolete_browser_resource_status=$?
set -e
assert_equals "$obsolete_browser_resource_status" "1" "obsolete browser resource manifest status"
assert_contains "$obsolete_browser_resource_output" "uses retired resource browser" "obsolete browser resource manifest output"

legacy_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-legacy.XXXXXX")"
cleanup_paths+=("$legacy_dir")
write_fake_make "$legacy_dir"
legacy_manifest="${legacy_dir}/manifest.json"
write_legacy_manifest "$legacy_manifest" "cartulary.service_backed_schedule.v2"
set +e
legacy_output="$(run_scheduler "$legacy_dir" "$legacy_manifest" test-fast-service-backed legacy 2>&1)"
legacy_status=$?
set -e
assert_equals "$legacy_status" "1" "legacy manifest status"
assert_contains "$legacy_output" "must declare schema_id cartulary.service_backed_schedule.v8" "legacy manifest output"

jobs_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-jobs.XXXXXX")"
cleanup_paths+=("$jobs_dir")
write_fake_make "$jobs_dir"
jobs_manifest="${jobs_dir}/manifest.json"
write_manifest "$jobs_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1'
set +e
jobs_output="$(
  FAKE_SCHEDULER_LOCK="${jobs_dir}/lock" \
  FAKE_SCHEDULER_ACTIVE="${jobs_dir}/active" \
  FAKE_SCHEDULER_MAX="${jobs_dir}/max" \
  FAKE_SCHEDULER_LOG="${jobs_dir}/make.log" \
  MAKE="${jobs_dir}/fake-make" \
  NODE_BIN="$NODE_BIN" \
  TEST_OUTPUT_SCRIPT="$TEST_OUTPUT_SCRIPT" \
  CARTULARY_TEST_RESULTS_DIR="${jobs_dir}/results" \
  CARTULARY_TEST_RUN_ID="jobs" \
    "$NODE_BIN" "$SCRIPT" --target test-fast-service-backed --jobs 1 --manifest "$jobs_manifest" 2>&1
)"
jobs_status=$?
set -e
assert_equals "$jobs_status" "1" "obsolete jobs flag status"
assert_contains "$jobs_output" "--jobs is obsolete" "obsolete jobs flag output"
