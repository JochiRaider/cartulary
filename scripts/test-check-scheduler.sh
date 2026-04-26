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

assert_file_absent() {
  local path="$1"
  local label="$2"

  if [[ -e "$path" ]]; then
    fail "$label: expected $path to be absent"
  fi
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

change_active() {
  local delta="$1"
  local active max action
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
  if [[ "$delta" -gt 0 ]]; then
    action=start
  else
    action=end
  fi
  printf '%s %s active=%s\n' "$action" "$target" "$active" >>"$event_log"
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
  "schema_id": "cartulary.check_schedule.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "cpu": 2, "service_stack": 1 },
      "work_units": [
        { "target": "setup", "weight": 50, "needs": [], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" },
        { "target": "build", "weight": 40, "needs": ["setup"], "resource_claims": { "cpu": "limit" }, "make_jobs": "cpu" },
        { "target": "local", "weight": 30, "needs": ["build"], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" },
        { "target": "service", "weight": 20, "needs": ["build"], "resource_claims": { "cpu": 1, "service_stack": 1 }, "make_jobs": "cpu" },
        { "target": "meta", "weight": 10, "needs": ["build"], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" }
      ]
    }
  ]
}
JSON
success_output="$(run_scheduler "$success_dir" "$success_manifest" success --summary-targets local,service,meta --summary-groups "check-work=local,service,meta" --resource-limit cpu=2 2>&1)"
assert_contains "$success_output" "[RUN] check steps=5 targets=3 jobs=2 run_id=success" "success run start"
assert_contains "$success_output" "[STEP] check 1/5 setup mode=scheduler jobs=1" "success setup step"
assert_contains "$success_output" "[STEP] check 2/5 build mode=scheduler jobs=2" "success build step"
assert_contains "$success_output" "[STEP] check 3/5 local mode=scheduler jobs=1" "success higher-weight local step"
assert_contains "$success_output" "[STEP] check 4/5 service mode=scheduler jobs=1" "success service step"
assert_contains "$success_output" "[PASS] check" "success summary"
assert_equals "$(cat "${success_dir}/max")" "2" "success resource limit"
assert_contains "$(cat "${success_dir}/make-args.log")" "--output-sync=target -j2 build" "build uses claimed cpu jobs"
success_summary="${success_dir}/results/success/run-summary.json"
assert_equals "$(json_field "$success_summary" "status")" "pass" "success summary status"
assert_equals "$(json_field "$success_summary" "completed_targets")" "5/5" "success completed"

failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-failure.XXXXXX")"
cleanup_paths+=("$failure_dir")
write_fake_make "$failure_dir"
failure_manifest="${failure_dir}/manifest.json"
cat >"$failure_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v1",
  "schedules": [
    {
      "target": "check",
      "resource_limits": { "cpu": 2, "service_stack": 1 },
      "work_units": [
        { "target": "alpha", "weight": 30, "needs": [], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" },
        { "target": "beta", "weight": 20, "needs": [], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" },
        { "target": "gamma", "weight": 10, "needs": [], "resource_claims": { "cpu": 1 }, "make_jobs": "cpu" }
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
  run_scheduler "$failure_dir" "$failure_manifest" failure --summary-targets alpha,beta,gamma --resource-limit cpu=2 2>&1
)"
failure_status=$?
set -e
assert_equals "$failure_status" "7" "failure exit status"
assert_contains "$failure_output" "fake failure for beta" "failure child output"
assert_contains "$failure_output" "[FAIL] check" "failure summary"
failure_events="$(cat "${failure_dir}/events.log")"
assert_contains "$failure_events" "start alpha" "failure alpha started"
assert_contains "$failure_events" "start beta" "failure beta started"
assert_contains "$failure_events" "end alpha" "failure alpha drained"
failure_summary="${failure_dir}/results/failure/run-summary.json"
assert_equals "$(json_field "$failure_summary" "status")" "fail" "failure summary status"
assert_equals "$(json_field "$failure_summary" "aborted_after")" "beta" "failure aborted after"

invalid_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-invalid.XXXXXX")"
cleanup_paths+=("$invalid_dir")
write_fake_make "$invalid_dir"
invalid_manifest="${invalid_dir}/manifest.json"
cat >"$invalid_manifest" <<'JSON'
{
  "schema_id": "cartulary.check_schedule.v1",
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

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/check-scheduler-dry-run.XXXXXX")"
cleanup_paths+=("$dry_run_dir")
write_fake_make "$dry_run_dir"
dry_run_output="$(
  MAKEFLAGS=n \
    run_scheduler "$dry_run_dir" "$success_manifest" dry-run --summary-targets local,service,meta --resource-limit cpu=2 2>&1
)"
assert_contains "$dry_run_output" "[DRY-RUN] check manifest=" "dry-run output"
assert_file_absent "${dry_run_dir}/make-args.log" "dry-run child make"
