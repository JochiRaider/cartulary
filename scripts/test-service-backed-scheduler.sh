#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
SCRIPT="${ROOT_DIR}/scripts/run-service-backed-schedule.mjs"
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

assert_file_absent() {
  local path="$1"
  local label="$2"

  if [[ -e "$path" ]]; then
    fail "$label: expected $path to be absent"
  fi
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

change_active() {
  local delta="$1"
  local active max
  exec 9>"$lock_file"
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
sleep_var="FAKE_SCHEDULER_SLEEP_${sleep_key}"
sleep_duration="${!sleep_var:-${FAKE_SCHEDULER_SLEEP:-0.05}}"

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
    printf '  "schema_id": "cartulary.service_backed_schedule.v5",\n'
    printf '  "schedules": [\n'
    printf '    { "target": "%s", "resource_limits": { "postgres": 32, "minio": 32, "go_cpu": 6, "go_io": 6, "process": 2, "browser": 1 }, "work_unit_sources": [\n' "$target"
    local first=1
    local source
    for source in "$@"; do
      IFS='|' read -r type name weight claims class browser_stage <<<"$source"
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
        printf ' }'
      else
        printf '      { "type": "go_shards", "class": "%s", "target": "%s", "resource_claims": {%s} }' \
          "$class" "$name" "$claims"
      fi
    done
    printf '\n    ] }\n'
    printf '  ]\n'
    printf '}\n'
  } >"$file"
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
  FAKE_GO_SLEEP_FINALIZE="${FAKE_GO_SLEEP_FINALIZE:-}" \
  FAKE_GO_SLEEP_CAPTURE_BACKEND_INTEGRATION_PHASE4_ENTITIES_SHARD_02="${FAKE_GO_SLEEP_CAPTURE_BACKEND_INTEGRATION_PHASE4_ENTITIES_SHARD_02:-}" \
  FAKE_GO_SLEEP_CAPTURE_BACKEND_INTEGRATION_PHASE2_INCIDENTS_SHARD_02="${FAKE_GO_SLEEP_CAPTURE_BACKEND_INTEGRATION_PHASE2_INCIDENTS_SHARD_02:-}" \
  FAKE_GO_SLEEP_FINALIZE_BACKEND_INTEGRATION="${FAKE_GO_SLEEP_FINALIZE_BACKEND_INTEGRATION:-}" \
  FAKE_GO_FAIL_SHARD="${FAKE_GO_FAIL_SHARD:-}" \
  FAKE_GO_FAIL_SHARD_STATUS="${FAKE_GO_FAIL_SHARD_STATUS:-}" \
  FAKE_GO_FAIL_FINALIZER_TARGET="${FAKE_GO_FAIL_FINALIZER_TARGET:-}" \
  FAKE_GO_FINALIZER_FAILURE_STATUS="${FAKE_GO_FINALIZER_FAILURE_STATUS:-}" \
  CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT= \
  CARTULARY_SERVICE_BACKED_GO_IO_LIMIT= \
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
assert_contains "$weighted_output" "[STEP] test-fast-service-backed 1/3 backend-process mode=scheduler jobs=6" "weighted first child"
assert_contains "$weighted_output" "[STEP] test-fast-service-backed 2/3 backend-integration-support mode=scheduler jobs=6" "weighted second child"
assert_contains "$weighted_output" "[STEP] test-fast-service-backed 3/3 backend-store mode=scheduler jobs=6" "weighted third child"
assert_contains "$weighted_output" "[SCHEDULER] test-fast-service-backed start work_unit=backend-process claims={go_cpu:1,go_io:1,minio:1,postgres:1,process:1} active=1 pending=2" "scheduler start telemetry"
assert_contains "$weighted_output" "[SCHEDULER] test-fast-service-backed start work_unit=backend-store claims={go_cpu:1,go_io:1,minio:1,postgres:1} active=3 pending=0 active_resource_claims={go_cpu:3,go_io:3,minio:3,postgres:3,process:1}" "scheduler starts all compatible weighted children"
assert_contains "$weighted_output" "resource_limits={browser:1,go_cpu:6,go_io:6,minio:32,postgres:32,process:2}" "scheduler resource limit telemetry"
assert_contains "$weighted_output" "[SCHEDULER] test-fast-service-backed finish work_unit=backend-process status=0" "scheduler finish telemetry"

resource_block_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-resource-block.XXXXXX")"
cleanup_paths+=("$resource_block_dir")
write_fake_make "$resource_block_dir"
resource_block_manifest="${resource_block_dir}/manifest.json"
write_manifest "$resource_block_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1, "go_cpu": 4, "go_io": 1' \
  'make_target|backend-store|9|"postgres": 1, "minio": 1, "go_cpu": 3, "go_io": 1' \
  'make_target|backend-process|8|"postgres": 1, "minio": 1, "go_cpu": 2, "go_io": 1, "process": 1'
resource_block_output="$(run_scheduler "$resource_block_dir" "$resource_block_manifest" test-fast-service-backed resource-block 2>&1)"
assert_contains "$resource_block_output" "[SCHEDULER] test-fast-service-backed start work_unit=backend-process claims={go_cpu:2,go_io:1,minio:1,postgres:1,process:1} active=2 pending=1 active_resource_claims={go_cpu:6,go_io:2,minio:2,postgres:2,process:1}" "weighted go_cpu resources fill capacity without oversubscription"
assert_contains "$resource_block_output" "[SCHEDULER] test-fast-service-backed blocked reason=resources blocked_resources=go_cpu" "scheduler go_cpu-blocked telemetry"

backend_capacity_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-backend-capacity.XXXXXX")"
cleanup_paths+=("$backend_capacity_dir")
write_fake_make "$backend_capacity_dir"
backend_capacity_manifest="${backend_capacity_dir}/manifest.json"
write_manifest "$backend_capacity_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1' \
  'make_target|backend-store|9|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1' \
  'make_target|backend-process|8|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1' \
  'make_target|backend-integration-support|7|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1' \
  'make_target|phase0-process-e2e|6|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1'
backend_capacity_output="$(run_scheduler "$backend_capacity_dir" "$backend_capacity_manifest" test-fast-service-backed backend-capacity 2>&1)"
assert_contains "$backend_capacity_output" "[SCHEDULER] test-fast-service-backed start work_unit=phase0-process-e2e claims={go_cpu:1,go_io:1,minio:1,postgres:1,process:1} active=5 pending=0 active_resource_claims={go_cpu:5,go_io:5,minio:5,postgres:5,process:2}" "go resource model starts more than four compatible backend work units"

io_block_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-io-block.XXXXXX")"
cleanup_paths+=("$io_block_dir")
write_fake_make "$io_block_dir"
io_block_manifest="${io_block_dir}/manifest.json"
write_manifest "$io_block_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 4' \
  'make_target|backend-store|9|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 3' \
  'make_target|backend-process|8|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 2, "process": 1'
io_block_output="$(run_scheduler "$io_block_dir" "$io_block_manifest" test-fast-service-backed io-block 2>&1)"
assert_contains "$io_block_output" "[SCHEDULER] test-fast-service-backed blocked reason=resources blocked_resources=go_io" "scheduler go_io-blocked telemetry"

browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-browser.XXXXXX")"
cleanup_paths+=("$browser_dir")
write_fake_make "$browser_dir"
browser_manifest="${browser_dir}/manifest.json"
write_manifest "$browser_manifest" test-service-backed \
  'make_target|backend-process|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend' \
  'make_target|browser-e2e-webserver-backed|9|"postgres": 1, "minio": 1, "process": 1, "browser": 1|browser|webserver-backed'
browser_output="$(run_scheduler "$browser_dir" "$browser_manifest" test-service-backed browser 2>&1)"
assert_contains "$browser_output" "[STEP] test-service-backed 1/2 backend-process mode=scheduler jobs=6" "browser schedule backend child"
assert_contains "$browser_output" "[STEP] test-service-backed 2/2 browser-e2e-webserver-backed mode=scheduler jobs=6" "browser schedule browser child"
assert_contains "$browser_output" "claims={browser:1,minio:1,postgres:1,process:1}" "browser resource claims telemetry"
assert_contains "$browser_output" "resource_limits={browser:1,go_cpu:6,go_io:6,minio:32,postgres:32,process:2}" "browser resource limits telemetry"

eager_finalizer_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-eager-finalizer.XXXXXX")"
cleanup_paths+=("$eager_finalizer_dir")
write_fake_make "$eager_finalizer_dir"
write_fake_go_target_runner "$eager_finalizer_dir"
eager_finalizer_manifest="${eager_finalizer_dir}/manifest.json"
cat >"$eager_finalizer_manifest" <<'JSON'
{
  "schema_id": "cartulary.service_backed_schedule.v5",
  "schedules": [
    {
      "target": "test-service-backed",
      "resource_limits": { "postgres": 32, "minio": 32, "go_cpu": 64, "go_io": 64, "process": 2, "browser": 1 },
      "work_unit_sources": [
        { "type": "go_shards", "class": "backend", "target": "backend-store", "resource_claims": { "postgres": 1, "minio": 1 } },
        { "type": "make_target", "class": "browser", "target": "browser-e2e-webserver-backed", "browser_stage": "webserver-backed", "weight": 9, "resource_claims": { "postgres": 1, "minio": 1, "process": 1, "browser": 1 } }
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
assert_contains "$eager_finalizer_output" "[SCHEDULER] test-service-backed finalize-start target=backend-store" "go finalizer starts eagerly"
assert_contains "$eager_finalizer_output" "[SCHEDULER] test-service-backed finalize-finish target=backend-store status=0" "go finalizer finishes eagerly"
assert_contains "$eager_finalizer_output" "aggregate-reports/backend-store/fake-aggregate" "go finalizer uses target-scoped aggregate output"
"$NODE_BIN" - "${eager_finalizer_dir}/make.log" <<'EOF'
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
const finalizeStart = indexOf("start finalize backend-store");
const browserEnd = indexOf("end browser-e2e-webserver-backed");
if (!(finalizeStart < browserEnd)) {
  throw new Error("backend-store finalizer waited for browser tail");
}
EOF

shared_finalizer_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-shared-finalizer.XXXXXX")"
cleanup_paths+=("$shared_finalizer_dir")
write_fake_make "$shared_finalizer_dir"
write_fake_go_target_runner "$shared_finalizer_dir"
shared_finalizer_manifest="${shared_finalizer_dir}/manifest.json"
write_manifest "$shared_finalizer_manifest" test-fast-service-backed \
  'go_shards|backend-integration|0|"postgres": 1, "minio": 1' \
  'go_shards|backend-integration-support|0|"postgres": 1, "minio": 1'
shared_finalizer_output="$(
  FAKE_GO_SLEEP_CAPTURE=0.005 \
  FAKE_GO_SLEEP_CAPTURE_BACKEND_INTEGRATION_PHASE4_ENTITIES_SHARD_02=0.8 \
  FAKE_GO_SLEEP_CAPTURE_BACKEND_INTEGRATION_PHASE2_INCIDENTS_SHARD_02=0.8 \
  FAKE_GO_SLEEP_FINALIZE=0.05 \
  FAKE_GO_SLEEP_FINALIZE_BACKEND_INTEGRATION=1.5 \
    run_scheduler "$shared_finalizer_dir" "$shared_finalizer_manifest" test-fast-service-backed shared-finalizer 2>&1
)"
assert_contains "$shared_finalizer_output" "[SCHEDULER] test-fast-service-backed finalize-start target=backend-integration" "backend-integration finalizer starts"
assert_contains "$shared_finalizer_output" "[SCHEDULER] test-fast-service-backed finalize-start target=backend-integration-support" "backend-integration-support finalizer starts"
assert_contains "$shared_finalizer_output" "aggregate-reports/backend-integration/fake-aggregate" "backend-integration aggregate output is target-scoped"
assert_contains "$shared_finalizer_output" "aggregate-reports/backend-integration-support/fake-aggregate" "backend-integration-support aggregate output is target-scoped"
"$NODE_BIN" - "${shared_finalizer_dir}/make.log" <<'EOF'
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
const sharedEnd = indexOf("end capture backend-integration backend-integration-phase4-entities-shard-02");
const supportUniqueEnd = indexOf("end capture backend-integration-support backend-integration-phase2-incidents-shard-02");
const supportStart = indexOf("start finalize backend-integration-support");
const integrationEnd = indexOf("end finalize backend-integration");
if (!(sharedEnd < supportStart)) {
  throw new Error("backend-integration-support finalizer started before shared shard captured under backend-integration completed");
}
if (!(supportUniqueEnd < supportStart)) {
  throw new Error("backend-integration-support finalizer started before its unique shard completed");
}
if (!(supportStart < integrationEnd)) {
  throw new Error("backend-integration-support finalizer did not run concurrently with backend-integration finalizer");
}
EOF

check_browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-check-browser.XXXXXX")"
cleanup_paths+=("$check_browser_dir")
write_fake_make "$check_browser_dir"
check_browser_manifest="${check_browser_dir}/manifest.json"
write_manifest "$check_browser_manifest" check-service-backed \
  'make_target|browser-e2e-webserver-backed|30|"postgres": 1, "minio": 1, "process": 1, "browser": 1|browser|webserver-backed' \
  'make_target|browser-e2e|20|"postgres": 1, "minio": 1, "process": 1, "browser": 1|browser|isolated' \
  'make_target|backend-process|10|"postgres": 1, "minio": 1, "go_cpu": 1, "go_io": 1, "process": 1|backend'
check_browser_output="$(
  FAKE_SCHEDULER_SLEEP_BACKEND_PROCESS=0.3
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E_WEBSERVER_BACKED=0.05
  FAKE_SCHEDULER_SLEEP_BROWSER_E2E=0.05
  run_scheduler "$check_browser_dir" "$check_browser_manifest" check-service-backed check-browser 2>&1
)"
assert_contains "$check_browser_output" "[STEP] check-service-backed 1/3 browser-e2e-webserver-backed mode=scheduler jobs=6" "check browser webserver child"
assert_contains "$check_browser_output" "[STEP] check-service-backed 2/3 backend-process mode=scheduler jobs=6" "check browser backend child"
assert_contains "$check_browser_output" "[STEP] check-service-backed 3/3 browser-e2e mode=scheduler jobs=6" "check browser isolated child"
check_browser_events="$(cat "${check_browser_dir}/make.log")"
assert_contains "$check_browser_events" "start browser-e2e-webserver-backed" "check browser webserver start"
assert_contains "$check_browser_events" "end browser-e2e-webserver-backed" "check browser webserver end"
assert_contains "$check_browser_events" "start browser-e2e" "check browser isolated start"
assert_contains "$check_browser_events" "end backend-process" "check browser backend end"
"$NODE_BIN" - "${check_browser_dir}/make.log" <<'EOF'
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
const webEnd = indexOf("end browser-e2e-webserver-backed");
const isolatedStart = indexOf("start browser-e2e ");
if (!(webEnd < isolatedStart)) {
  throw new Error("isolated browser batch started before webserver-backed released browser resource");
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
assert_contains "$failure_output" "[FAIL] test-fast-service-backed" "failure target summary"

failed_shard_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-failed-shard.XXXXXX")"
cleanup_paths+=("$failed_shard_dir")
write_fake_make "$failed_shard_dir"
write_fake_go_target_runner "$failed_shard_dir"
failed_shard_manifest="${failed_shard_dir}/manifest.json"
write_manifest "$failed_shard_manifest" test-fast-service-backed \
  'go_shards|backend-store|0|"postgres": 1, "minio": 1'
set +e
failed_shard_output="$(
  FAKE_GO_FAIL_SHARD=backend-store-shared-shard-01 \
  FAKE_GO_FINALIZER_FAILURE_STATUS=9 \
    run_scheduler "$failed_shard_dir" "$failed_shard_manifest" test-fast-service-backed failed-shard 2>&1
)"
failed_shard_status=$?
set -e
assert_equals "$failed_shard_status" "9" "failed shard finalizer status"
assert_contains "$failed_shard_output" "fake shard failure for backend-store-shared-shard-01" "failed shard output"
assert_contains "$failed_shard_output" "[SCHEDULER] test-fast-service-backed finalize-start target=backend-store" "failed shard still finalizes target"
assert_contains "$failed_shard_output" "[SCHEDULER] test-fast-service-backed finalize-finish target=backend-store status=9" "failed shard finalizer reports failure"
assert_contains "$failed_shard_output" "[FAIL] test-fast-service-backed" "failed shard parent summary"

defer_summary_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-defer-summary.XXXXXX")"
cleanup_paths+=("$defer_summary_dir")
write_fake_make "$defer_summary_dir"
defer_summary_manifest="${defer_summary_dir}/manifest.json"
write_manifest "$defer_summary_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1' \
  'make_target|backend-store|9|"postgres": 1, "minio": 1'
defer_summary_output="$(run_scheduler "$defer_summary_dir" "$defer_summary_manifest" test-fast-service-backed defer-summary --defer-summary 2>&1)"
assert_contains "$defer_summary_output" "[TARGET] start test-fast-service-backed" "defer-summary target start"
assert_file_absent "${defer_summary_dir}/results/defer-summary/test-fast-service-backed/target-summary.json" "defer-summary parent target summary"

unsafe_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-unsafe.XXXXXX")"
cleanup_paths+=("$unsafe_dir")
write_fake_make "$unsafe_dir"
unsafe_manifest="${unsafe_dir}/manifest.json"
write_manifest "$unsafe_manifest" check-service-backed \
  'make_target|phase0-process-e2e|10|"postgres": 1, "minio": 1'
set +e
unsafe_output="$(run_scheduler "$unsafe_dir" "$unsafe_manifest" check-service-backed unsafe 2>&1)"
unsafe_status=$?
set -e
assert_equals "$unsafe_status" "1" "unsafe manifest status"
assert_contains "$unsafe_output" "is not check-service-backed safe" "unsafe manifest output"

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
write_manifest "$dry_run_manifest" test-fast-service-backed \
  'make_target|backend-integration|10|"postgres": 1, "minio": 1'
dry_run_output="$(
  MAKEFLAGS=n \
    run_scheduler "$dry_run_dir" "$dry_run_manifest" test-fast-service-backed dry-run 2>&1
)"
assert_contains "$dry_run_output" "[DRY-RUN] test-fast-service-backed manifest=" "dry-run output"
assert_file_absent "${dry_run_dir}/make.log" "dry-run child make log"

go_shard_dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-go-shard-dry-run.XXXXXX")"
cleanup_paths+=("$go_shard_dry_run_dir")
write_fake_make "$go_shard_dry_run_dir"
go_shard_dry_run_manifest="${go_shard_dry_run_dir}/manifest.json"
write_manifest "$go_shard_dry_run_manifest" test-fast-service-backed \
  'go_shards|backend-store|0|"postgres": 1, "minio": 1'
go_shard_dry_run_output="$(
  MAKEFLAGS=n \
    run_scheduler "$go_shard_dry_run_dir" "$go_shard_dry_run_manifest" test-fast-service-backed go-shard-dry-run 2>&1
)"
assert_contains "$go_shard_dry_run_output" "backend-store/backend-store-shared-shard-01 profile=cpu_heavy claims={go_cpu:2,go_io:1,minio:1,postgres:1}" "go_shards dry-run includes per-shard cpu-heavy claims"

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
assert_contains "$invalid_resource_output" "undeclared-resource is not declared in resource_limits" "invalid resource manifest output"

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
assert_contains "$removed_backend_resource_output" "removed generic backend resource" "removed backend resource manifest output"

invalid_browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-invalid-browser.XXXXXX")"
cleanup_paths+=("$invalid_browser_dir")
write_fake_make "$invalid_browser_dir"
invalid_browser_manifest="${invalid_browser_dir}/manifest.json"
write_manifest "$invalid_browser_manifest" test-service-backed \
  'make_target|browser-e2e-visual|10|"postgres": 1, "minio": 1, "browser": 1|browser|visual'
set +e
invalid_browser_output="$(run_scheduler "$invalid_browser_dir" "$invalid_browser_manifest" test-service-backed invalid-browser 2>&1)"
invalid_browser_status=$?
set -e
assert_equals "$invalid_browser_status" "1" "invalid browser manifest status"
assert_contains "$invalid_browser_output" "browser target browser-e2e-visual is not scheduler-safe" "invalid browser manifest output"

invalid_isolated_browser_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-invalid-isolated-browser.XXXXXX")"
cleanup_paths+=("$invalid_isolated_browser_dir")
write_fake_make "$invalid_isolated_browser_dir"
invalid_isolated_browser_manifest="${invalid_isolated_browser_dir}/manifest.json"
write_manifest "$invalid_isolated_browser_manifest" test-service-backed \
  'make_target|browser-e2e|10|"postgres": 1, "minio": 1, "browser": 1|browser|isolated'
set +e
invalid_isolated_browser_output="$(run_scheduler "$invalid_isolated_browser_dir" "$invalid_isolated_browser_manifest" test-service-backed invalid-isolated-browser 2>&1)"
invalid_isolated_browser_status=$?
set -e
assert_equals "$invalid_isolated_browser_status" "1" "invalid isolated browser manifest status"
assert_contains "$invalid_isolated_browser_output" "browser target browser-e2e is not scheduler-safe for test-service-backed" "invalid isolated browser manifest output"

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
assert_contains "$legacy_output" "must declare schema_id cartulary.service_backed_schedule.v5" "legacy manifest output"

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
