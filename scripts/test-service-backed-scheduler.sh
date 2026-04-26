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
  "duration_ms": 1,
  "wall_duration_ms": 1,
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

change_active 1
sleep "${FAKE_SCHEDULER_SLEEP:-0.05}"
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

write_manifest() {
  local file="$1"
  local target="$2"
  shift 2

  {
    printf '{\n'
    printf '  "schema_id": "cartulary.service_backed_schedule.v1",\n'
    printf '  "schedules": [\n'
    printf '    { "target": "%s", "children": [\n' "$target"
    local first=1
    local child
    for child in "$@"; do
      IFS='|' read -r name kind weight exclusive <<<"$child"
      if [[ "$first" -eq 0 ]]; then
        printf ',\n'
      fi
      first=0
      printf '      { "target": "%s", "kind": "%s", "weight": %s, "resource_tags": ["postgres", "minio"%s], "exclusive_tags": [%s] }' \
        "$name" "$kind" "$weight" "$([[ "$kind" == browser ]] && printf ', "browser"' || true)" "$exclusive"
    done
    printf '\n    ] }\n'
    printf '  ]\n'
    printf '}\n'
  } >"$file"
}

run_scheduler() {
  local dir="$1"
  local manifest="$2"
  local target="$3"
  local jobs="$4"
  local run_id="$5"

  FAKE_SCHEDULER_LOCK="${dir}/lock" \
  FAKE_SCHEDULER_ACTIVE="${dir}/active" \
  FAKE_SCHEDULER_MAX="${dir}/max" \
  FAKE_SCHEDULER_LOG="${dir}/make.log" \
  FAKE_FAIL_TARGET="${FAKE_FAIL_TARGET:-}" \
  FAKE_SCHEDULER_SLEEP="${FAKE_SCHEDULER_SLEEP:-0.05}" \
  MAKE="${dir}/fake-make" \
  NODE_BIN="$NODE_BIN" \
  TEST_OUTPUT_SCRIPT="$TEST_OUTPUT_SCRIPT" \
  CARTULARY_TEST_RESULTS_DIR="${dir}/results" \
  CARTULARY_TEST_RUN_ID="$run_id" \
    "$NODE_BIN" "$SCRIPT" --target "$target" --jobs "$jobs" --manifest "$manifest"
}

weighted_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-weighted.XXXXXX")"
cleanup_paths+=("$weighted_dir")
write_fake_make "$weighted_dir"
weighted_manifest="${weighted_dir}/manifest.json"
write_manifest "$weighted_manifest" test-fast-service-backed \
  'backend-store|backend|1|' \
  'backend-process|backend|10|' \
  'backend-integration-support|backend|5|'
weighted_output="$(run_scheduler "$weighted_dir" "$weighted_manifest" test-fast-service-backed 1 weighted 2>&1)"
assert_contains "$weighted_output" "[STEP] test-fast-service-backed 1/3 backend-process mode=scheduler jobs=1" "weighted first child"
assert_contains "$weighted_output" "[STEP] test-fast-service-backed 2/3 backend-integration-support mode=scheduler jobs=1" "weighted second child"
assert_contains "$weighted_output" "[STEP] test-fast-service-backed 3/3 backend-store mode=scheduler jobs=1" "weighted third child"

parallel_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-parallel.XXXXXX")"
cleanup_paths+=("$parallel_dir")
write_fake_make "$parallel_dir"
parallel_manifest="${parallel_dir}/manifest.json"
write_manifest "$parallel_manifest" test-fast-service-backed \
  'backend-integration|backend|10|' \
  'backend-store|backend|9|' \
  'backend-process|backend|8|' \
  'backend-integration-support|backend|7|'
run_scheduler "$parallel_dir" "$parallel_manifest" test-fast-service-backed 2 parallel >/dev/null
assert_equals "$(cat "${parallel_dir}/max")" "2" "parallel max active"

exclusive_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-exclusive.XXXXXX")"
cleanup_paths+=("$exclusive_dir")
write_fake_make "$exclusive_dir"
exclusive_manifest="${exclusive_dir}/manifest.json"
write_manifest "$exclusive_manifest" test-fast-service-backed \
  'backend-integration|backend|10|"postgres"' \
  'backend-store|backend|9|"postgres"' \
  'backend-process|backend|8|"postgres"'
run_scheduler "$exclusive_dir" "$exclusive_manifest" test-fast-service-backed 3 exclusive >/dev/null
assert_equals "$(cat "${exclusive_dir}/max")" "1" "exclusive tag max active"

failure_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-failure.XXXXXX")"
cleanup_paths+=("$failure_dir")
write_fake_make "$failure_dir"
failure_manifest="${failure_dir}/manifest.json"
write_manifest "$failure_manifest" test-fast-service-backed \
  'backend-integration|backend|10|' \
  'backend-store|backend|9|'
set +e
failure_output="$(
  FAKE_FAIL_TARGET=backend-store \
    run_scheduler "$failure_dir" "$failure_manifest" test-fast-service-backed 2 failure 2>&1
)"
failure_status=$?
set -e
assert_equals "$failure_status" "7" "child failure status"
assert_contains "$failure_output" "fake failure for backend-store" "child failure output"
assert_contains "$failure_output" "[FAIL] test-fast-service-backed" "failure target summary"

unsafe_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-unsafe.XXXXXX")"
cleanup_paths+=("$unsafe_dir")
write_fake_make "$unsafe_dir"
unsafe_manifest="${unsafe_dir}/manifest.json"
write_manifest "$unsafe_manifest" check-service-backed \
  'phase0-process-e2e|backend|10|'
set +e
unsafe_output="$(run_scheduler "$unsafe_dir" "$unsafe_manifest" check-service-backed 1 unsafe 2>&1)"
unsafe_status=$?
set -e
assert_equals "$unsafe_status" "1" "unsafe manifest status"
assert_contains "$unsafe_output" "is not check-service-backed safe" "unsafe manifest output"

unknown_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-unknown.XXXXXX")"
cleanup_paths+=("$unknown_dir")
write_fake_make "$unknown_dir"
unknown_manifest="${unknown_dir}/manifest.json"
write_manifest "$unknown_manifest" test-fast-service-backed \
  'unknown-backend-target|backend|10|'
set +e
unknown_output="$(run_scheduler "$unknown_dir" "$unknown_manifest" test-fast-service-backed 1 unknown 2>&1)"
unknown_status=$?
set -e
assert_equals "$unknown_status" "1" "unknown manifest status"
assert_contains "$unknown_output" "is not in target-plan" "unknown manifest output"

dry_run_dir="$(mktemp -d "${ROOT_DIR}/tmp/service-backed-scheduler-dry-run.XXXXXX")"
cleanup_paths+=("$dry_run_dir")
write_fake_make "$dry_run_dir"
dry_run_manifest="${dry_run_dir}/manifest.json"
write_manifest "$dry_run_manifest" test-fast-service-backed \
  'backend-integration|backend|10|'
dry_run_output="$(
  MAKEFLAGS=n \
    run_scheduler "$dry_run_dir" "$dry_run_manifest" test-fast-service-backed 1 dry-run 2>&1
)"
assert_contains "$dry_run_output" "[DRY-RUN] test-fast-service-backed jobs=1" "dry-run output"
assert_file_absent "${dry_run_dir}/make.log" "dry-run child make log"
