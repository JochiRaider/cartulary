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

assert_json_number_gt() {
  local file="$1"
  local field="$2"
  local threshold="$3"
  local label="$4"

  "$node_bin" - "$file" "$field" "$threshold" "$label" <<'EOF'
const fs = require("node:fs");
const [file, field, thresholdText, label] = process.argv.slice(2);
const data = JSON.parse(fs.readFileSync(file, "utf8"));
const value = field.split(".").reduce((current, part) => current?.[part], data);
const threshold = Number(thresholdText);
if (typeof value !== "number" || !(value > threshold)) {
  throw new Error(`${label}: expected ${field} > ${threshold}, got ${value}`);
}
EOF
}

assert_json_field_equals() {
  local file="$1"
  local field="$2"
  local expected="$3"
  local label="$4"

  "$node_bin" - "$file" "$field" "$expected" "$label" <<'EOF'
const fs = require("node:fs");
const [file, field, expected, label] = process.argv.slice(2);
const data = JSON.parse(fs.readFileSync(file, "utf8"));
const value = field.split(".").reduce((current, part) => current?.[part], data);
if (String(value) !== expected) {
  throw new Error(`${label}: expected ${field}=${expected}, got ${value}`);
}
EOF
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
expected_children="backend-store,backend-integration,backend-integration-support,backend-process,browser-e2e-webserver-backed,browser-e2e"
assert_contains "$(cat "$tmp_dir/pass.log")" "summary args=target-summary test-service-backed pass --children $expected_children" "pass summary"

run_case scheduler-fail 7 0 7
assert_contains "$(cat "$tmp_dir/scheduler-fail.log")" "summary args=target-summary test-service-backed fail --children $expected_children" "scheduler fail summary"

run_case summary-fail 0 9 9
assert_contains "$(cat "$tmp_dir/summary-fail.log")" "summary args=target-summary test-service-backed pass --children $expected_children" "summary fail summary"

run_case scheduler-and-summary-fail 7 9 7
assert_contains "$(cat "$tmp_dir/scheduler-and-summary-fail.log")" "summary args=target-summary test-service-backed fail --children $expected_children" "scheduler precedence summary"

summary_make="$tmp_dir/fake-summary-make.sh"
cat >"$summary_make" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

printf 'summary-make target=%s test-target=%s\n' "${*: -1}" "${CARTULARY_TEST_TARGET:-}" >>"${FAKE_SUMMARY_MAKE_LOG:?}"
case "${*: -1}" in
  child-pass)
    sleep 0.05
    exit 0
    ;;
  child-fail)
    sleep 0.05
    echo "child failed" >&2
    exit 6
    ;;
  *)
    echo "unexpected child target: ${*: -1}" >&2
    exit 2
    ;;
esac
EOF
chmod +x "$summary_make"

summary_pass_results="$tmp_dir/summary-pass-results"
summary_pass_log="$tmp_dir/summary-pass-make.log"
summary_pass_output="$(
  FAKE_SUMMARY_MAKE_LOG="$summary_pass_log" \
  MAKE_BIN="$summary_make" \
  NODE_BIN="$node_bin" \
  RUN_PHASE_SCRIPT="$ROOT_DIR/scripts/lib/run-phase.sh" \
  TEST_OUTPUT_SCRIPT="$ROOT_DIR/scripts/lib/test-output.sh" \
  TASK_SURFACE_MANIFEST="$ROOT_DIR/tools/task_surface_manifest.json" \
  CARTULARY_TEST_RESULTS_DIR="$summary_pass_results" \
  CARTULARY_TEST_RUN_ID="summary-pass" \
    "$node_bin" "$HELPER" summary-target --target check-summary-pass --child-target child-pass --status pass --phase-label "summary pass child" \
    2>&1
)"
assert_contains "$summary_pass_output" "[PASS] check-summary-pass kind=leaf" "summary pass output"
assert_contains "$(cat "$summary_pass_log")" "summary-make target=child-pass test-target=check-summary-pass" "summary pass child target env"
summary_pass_file="$summary_pass_results/summary-pass/check-summary-pass/target-summary.json"
assert_json_field_equals "$summary_pass_file" "status" "pass" "summary pass status"
assert_json_number_gt "$summary_pass_file" "wall_duration_ms" "0" "summary pass wall duration"
assert_json_number_gt "$summary_pass_file" "critical_path_wall_duration_ms" "0" "summary pass critical duration"
assert_json_number_gt "$summary_pass_file" "executed_duration_ms" "0" "summary pass executed duration"
assert_json_number_gt "$summary_pass_file" "logical_duration_ms" "0" "summary pass logical duration"

summary_fail_results="$tmp_dir/summary-fail-results"
summary_fail_log="$tmp_dir/summary-fail-make.log"
set +e
summary_fail_output="$(
  FAKE_SUMMARY_MAKE_LOG="$summary_fail_log" \
  MAKE_BIN="$summary_make" \
  NODE_BIN="$node_bin" \
  RUN_PHASE_SCRIPT="$ROOT_DIR/scripts/lib/run-phase.sh" \
  TEST_OUTPUT_SCRIPT="$ROOT_DIR/scripts/lib/test-output.sh" \
  TASK_SURFACE_MANIFEST="$ROOT_DIR/tools/task_surface_manifest.json" \
  CARTULARY_TEST_RESULTS_DIR="$summary_fail_results" \
  CARTULARY_TEST_RUN_ID="summary-fail" \
    "$node_bin" "$HELPER" summary-target --target check-summary-fail --child-target child-fail --status pass --phase-label "summary fail child" \
    2>&1
)"
summary_fail_status=$?
set -e
assert_equals "$summary_fail_status" "6" "summary child failure exit status"
assert_contains "$summary_fail_output" "[FAIL] check-summary-fail" "summary fail output"
assert_json_field_equals "$summary_fail_results/summary-fail/check-summary-fail/target-summary.json" "status" "fail" "summary fail status"

projection_run_phase="$tmp_dir/projection-run-phase.sh"
cat >"$projection_run_phase" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${2:-}" != "--" ]]; then
  echo "unexpected projection run-phase invocation: $*" >&2
  exit 2
fi
shift 2
printf 'projection child=%s target=%s\n' "${*: -1}" "${CARTULARY_TEST_TARGET:-}" >>"${FAKE_PROJECTION_LOG:?}"
exit 0
EOF
chmod +x "$projection_run_phase"

projection_test_output="$tmp_dir/projection-test-output.sh"
cat >"$projection_test_output" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf 'projection-summary args=%s\n' "$*" >>"${FAKE_PROJECTION_LOG:?}"
exit 0
EOF
chmod +x "$projection_test_output"

projection_log="$tmp_dir/projection.log"
FAKE_PROJECTION_LOG="$projection_log" \
MAKE_BIN="$summary_make" \
NODE_BIN="$node_bin" \
RUN_PHASE_SCRIPT="$projection_run_phase" \
TEST_OUTPUT_SCRIPT="$projection_test_output" \
TASK_SURFACE_MANIFEST="$ROOT_DIR/tools/task_surface_manifest.json" \
  "$node_bin" "$HELPER" summary-target --target check-summary-projection --child-target child-pass --status pass --phase-label "summary projection child" --projection check-harness-smoke
assert_contains "$(cat "$projection_log")" "projection child=child-pass target=check-summary-projection" "summary projection child"
assert_contains "$(cat "$projection_log")" "projection-summary args=target-summary check-summary-projection pass --projection check-harness-smoke --skipped-from-child child-pass" "summary projection args"
