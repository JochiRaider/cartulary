#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
HELPER="$ROOT_DIR/scripts/cartulary-runner.mjs"
node_bin="${NODE_BIN:-node}"
cleanup_paths=()

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
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

tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/cartulary-runner-service-backed-target.XXXXXX")"
cleanup_paths+=("$tmp_dir")

fake_test_services="$tmp_dir/fake-test-services.sh"
cat >"$fake_test_services" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" != "run" || "${2:-}" != "--" ]]; then
  echo "unexpected test-services invocation: $*" >&2
  exit 2
fi
shift 2
printf 'test-services %s\n' "$*" >>"${FAKE_WRAPPER_LOG:?}"
exec "$@"
EOF
chmod +x "$fake_test_services"

fake_run_phase="$tmp_dir/fake-run-phase.sh"
cat >"$fake_run_phase" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

label="${1:-}"
if [[ "${2:-}" != "--" ]]; then
  echo "unexpected run-phase invocation: $*" >&2
  exit 2
fi
shift 2
printf 'phase-label=%s\n' "$label" >>"${FAKE_WRAPPER_LOG:?}"
exec "$@"
EOF
chmod +x "$fake_run_phase"

fake_scheduler="$tmp_dir/fake-scheduler.mjs"
cat >"$fake_scheduler" <<'EOF'
import { appendFileSync } from "node:fs";

appendFileSync(
  process.env.FAKE_WRAPPER_LOG,
  `scheduler args=${process.argv.slice(2).join(" ")} make=${process.env.MAKE ?? ""} target=${process.env.NODE_BIN ?? ""}\n`,
);
process.exit(Number(process.env.FAKE_SCHEDULER_STATUS ?? "0"));
EOF

fake_test_output="$tmp_dir/fake-test-output.sh"
cat >"$fake_test_output" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

printf 'summary args=%s\n' "$*" >>"${FAKE_WRAPPER_LOG:?}"
exit "${FAKE_SUMMARY_STATUS:-0}"
EOF
chmod +x "$fake_test_output"

run_case() {
  local name="$1"
  local scheduler_status="$2"
  local summary_status="$3"
  local expected_status="$4"
  local log_file="$tmp_dir/${name}.log"
  local output
  local status

  set +e
  output="$(
    FAKE_WRAPPER_LOG="$log_file" \
    FAKE_SCHEDULER_STATUS="$scheduler_status" \
    FAKE_SUMMARY_STATUS="$summary_status" \
    MAKE="fake-make" \
    NODE_BIN="$node_bin" \
    TEST_SERVICES_BIN="$fake_test_services" \
    RUN_PHASE_SCRIPT="$fake_run_phase" \
    RUN_SERVICE_BACKED_SCHEDULE_SCRIPT="$fake_scheduler" \
    TEST_OUTPUT_SCRIPT="$fake_test_output" \
    TASK_SURFACE_MANIFEST="$ROOT_DIR/tools/task_surface_manifest.json" \
    SERVICE_BACKED_SCHEDULE_MANIFEST="$ROOT_DIR/tools/service_backed_schedule_manifest.json" \
      "$node_bin" "$HELPER" service-backed-target --target test-service-backed --phase-label "test service-backed" --service-wrapper test-services \
      2>&1
  )"
  status=$?
  set -e

  assert_equals "$status" "$expected_status" "$name exit status"
  assert_equals "$output" "" "$name output"
  assert_contains "$(cat "$log_file")" "test-services $fake_run_phase" "$name service wrapper"
  assert_contains "$(cat "$log_file")" "phase-label=test service-backed" "$name phase label"
  assert_contains "$(cat "$log_file")" "scheduler args=--target test-service-backed --manifest $ROOT_DIR/tools/service_backed_schedule_manifest.json --defer-summary" "$name scheduler args"
}

run_case pass 0 0 0
assert_contains "$(cat "$tmp_dir/pass.log")" "summary args=target-summary test-service-backed pass --projection test-service-backed" "pass summary"

run_case scheduler-fail 7 0 7
assert_contains "$(cat "$tmp_dir/scheduler-fail.log")" "summary args=target-summary test-service-backed fail --projection test-service-backed" "scheduler fail summary"

run_case summary-fail 0 9 9
assert_contains "$(cat "$tmp_dir/summary-fail.log")" "summary args=target-summary test-service-backed pass --projection test-service-backed" "summary fail summary"

run_case scheduler-and-summary-fail 7 9 7
assert_contains "$(cat "$tmp_dir/scheduler-and-summary-fail.log")" "summary args=target-summary test-service-backed fail --projection test-service-backed" "scheduler precedence summary"
