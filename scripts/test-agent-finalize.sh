#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
NODE_BIN="${NODE_BIN:-node}"
SCRIPT="$ROOT_DIR/scripts/agent-finalize.sh"
RUN_PHASE="$ROOT_DIR/scripts/lib/run-phase.sh"
TMP_DIR="$(mktemp -d "$ROOT_DIR/tmp/agent-finalize-test.XXXXXX")"

cleanup() {
  rm -rf "$TMP_DIR"
}

trap cleanup EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"
  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected [$needle] in [$haystack]"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"
  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: unexpected [$needle] in [$haystack]"
  fi
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"
  if [[ "$actual" != "$expected" ]]; then
    fail "$label: expected [$expected], got [$actual]"
  fi
}

json_field() {
  local file="$1"
  local expr="$2"
  "$NODE_BIN" - "$file" "$expr" <<'JS'
const fs = require("node:fs");
const [file, expr] = process.argv.slice(2);
const value = JSON.parse(fs.readFileSync(file, "utf8"));
const result = Function("value", `return ${expr}`)(value);
if (Array.isArray(result)) {
  process.stdout.write(result.join("\n"));
} else if (result === null || result === undefined) {
  process.stdout.write("");
} else {
  process.stdout.write(String(result));
}
JS
}

write_fake_make() {
  local file="$1"
  cat >"$file" <<'SH'
#!/usr/bin/env bash
set -euo pipefail
target="${@: -1}"
printf '%s\n' "$target" >>"$FAKE_MAKE_LOG"
if [[ "${FAKE_FAIL_TARGET:-}" == "$target" ]]; then
  if [[ -n "${CARTULARY_TEST_RESULTS_DIR:-}" && -n "${CARTULARY_TEST_RUN_ID:-}" ]]; then
    mkdir -p "${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}"
    cat >"${CARTULARY_TEST_RESULTS_DIR}/${CARTULARY_TEST_RUN_ID}/${target}/tool-run-summary.json" <<JSON
{
  "schema_id": "cartulary.tool_run_summary.v3",
  "target": "${target}",
  "status": "fail",
  "failure_class": "product",
  "failure_reason": "test_assertion_failure",
  "failures": [
    {
      "headline": "${target} synthetic assertion"
    }
  ]
}
JSON
  fi
  printf '%s synthetic failure\n' "$target" >&2
  exit 17
fi
SH
  chmod +x "$file"
}

write_retained_run() {
  local dir="$1"
  mkdir -p "$dir/check" "$dir/backend-unit/backend-unit" "$dir/check/check"
  cat >"$dir/check/tool-run-summary.json" <<'JSON'
{
  "schema_id": "cartulary.tool_run_summary.v3",
  "target": "check",
  "status": "pass"
}
JSON
  cat >"$dir/check/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.check_scheduler_summary.v9",
  "target": "check",
  "status": "pass"
}
JSON
  cat >"$dir/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"scheduler-start","seq":1,"monotonic_ms":0,"emitted_at":"2026-01-01T00:00:00.000Z"}
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"scheduler-finish","seq":2,"monotonic_ms":10,"emitted_at":"2026-01-01T00:00:00.010Z"}
JSONL
  cat >"$dir/backend-unit/target-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_target_summary.v4",
  "target": "backend-unit",
  "status": "pass"
}
JSON
  cat >"$dir/backend-unit/backend-unit/phase-summary.json" <<'JSON'
{
  "schema_id": "cartulary.test_phase_summary.v3",
  "target": "backend-unit",
  "status": "pass"
}
JSON
}

write_service_backed_only_run() {
  local dir="$1"
  mkdir -p "$dir/check-service-backed"
  cat >"$dir/check-service-backed/tool-run-summary.json" <<'JSON'
{
  "schema_id": "cartulary.tool_run_summary.v3",
  "target": "check-service-backed",
  "status": "pass"
}
JSON
}

write_incomplete_retained_run() {
  local dir="$1"
  mkdir -p "$dir/check"
  cat >"$dir/check/tool-run-summary.json" <<'JSON'
{
  "schema_id": "cartulary.tool_run_summary.v3",
  "target": "check",
  "status": "pass"
}
JSON
  cat >"$dir/check/scheduler-summary.json" <<'JSON'
{
  "schema_id": "cartulary.check_scheduler_summary.v9",
  "target": "check",
  "status": "pass"
}
JSON
  cat >"$dir/check/scheduler-events.jsonl" <<'JSONL'
{"schema_id":"cartulary.scheduler_event.v6","target":"check","event":"scheduler-start","seq":1,"monotonic_ms":0,"emitted_at":"2026-01-01T00:00:00.000Z"}
JSONL
}

assert_invalid_results_dir() {
  local label="$1"
  local retained_root="$2"
  local expected_class="$3"
  local expected_reason="$4"
  local expected_headline="$5"
  local scenario_dir="$TMP_DIR/invalid-${label}"
  mkdir -p "$scenario_dir"
  local scenario_make="$scenario_dir/fake-make"
  local scenario_log="$scenario_dir/make.log"
  write_fake_make "$scenario_make"
  set +e
  CARTULARY_TEST_RESULTS_DIR="$scenario_dir/results" \
  CARTULARY_TEST_RUN_ID="$label" \
  CARTULARY_PHASE_ARTIFACT_DIR="$scenario_dir/results/$label/agent-finalize/agent-finalize" \
  MAKE="$scenario_make" \
  FAKE_MAKE_LOG="$scenario_log" \
  RESULTS_DIR="$retained_root" \
    "$SCRIPT" >"$scenario_dir/stdout.log" 2>"$scenario_dir/stderr.log"
  local status=$?
  set -e
  if [[ "$status" -eq 0 ]]; then
    fail "$label invalid RESULTS_DIR unexpectedly passed"
  fi
  if [[ -f "$scenario_log" ]]; then
    fail "$label preflight failure must not run mutating child targets"
  fi
  local summary="$scenario_dir/results/$label/agent-finalize/finalize-summary.json"
  assert_equals "$(json_field "$summary" 'value.results_dir_status')" "invalid" "$label results dir status"
  assert_equals "$(json_field "$summary" 'value.failures[0].action_id')" "structure_ledger_refresh" "$label failure action"
  assert_equals "$(json_field "$summary" 'value.failures[0].substep_id')" "retained-run-preflight" "$label failure substep"
  assert_equals "$(json_field "$summary" 'value.failures[0].failure_class')" "$expected_class" "$label failure class"
  assert_equals "$(json_field "$summary" 'value.failures[0].failure_reason')" "$expected_reason" "$label failure reason"
  assert_contains "$(json_field "$summary" 'value.failures[0].headline')" "$expected_headline" "$label failure headline"
}

success_dir="$TMP_DIR/success"
mkdir -p "$success_dir"
success_make="$success_dir/fake-make"
success_log="$success_dir/make.log"
write_fake_make "$success_make"
CARTULARY_TEST_RESULTS_DIR="$success_dir/results" \
CARTULARY_TEST_RUN_ID="success" \
CARTULARY_PHASE_ARTIFACT_DIR="$success_dir/results/success/agent-finalize/agent-finalize" \
MAKE="$success_make" \
FAKE_MAKE_LOG="$success_log" \
RESULTS_DIR="" \
  "$SCRIPT"
summary="$success_dir/results/success/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$summary" 'value.schema_id')" "cartulary.agent_finalize_summary.v2" "no RESULTS_DIR schema"
assert_equals "$(json_field "$summary" 'value.actions.map((action) => action.action_id)')" $'structure_ledger_refresh\nschema_shape_validation\nduration_baseline_coverage' "no RESULTS_DIR action selection"
assert_equals "$(json_field "$summary" 'value.status')" "pass" "no RESULTS_DIR status"
assert_equals "$(json_field "$summary" 'value.duration.status')" "skipped" "no RESULTS_DIR duration status"
assert_equals "$(json_field "$summary" 'value.run_checks.status')" "skipped" "no RESULTS_DIR run checks"
assert_not_contains "$(cat "$success_log")" "duration-baseline-drift-suite" "no RESULTS_DIR skips retained-run drift"

retained_dir="$TMP_DIR/retained-run"
write_retained_run "$retained_dir"
results_dir="$TMP_DIR/with-results"
mkdir -p "$results_dir"
results_make="$results_dir/fake-make"
results_log="$results_dir/make.log"
write_fake_make "$results_make"
CARTULARY_TEST_RESULTS_DIR="$results_dir/results" \
CARTULARY_TEST_RUN_ID="with-results" \
CARTULARY_PHASE_ARTIFACT_DIR="$results_dir/results/with-results/agent-finalize/agent-finalize" \
MAKE="$results_make" \
FAKE_MAKE_LOG="$results_log" \
RESULTS_DIR="$retained_dir" \
  "$SCRIPT"
results_summary="$results_dir/results/with-results/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$results_summary" 'value.actions.map((action) => action.action_id)')" $'structure_ledger_refresh\nschema_shape_validation\nduration_baseline_refresh\nduration_baseline_coverage\nduration_baseline_drift_validation\nscheduler_drift_validation' "RESULTS_DIR action selection"
assert_equals "$(json_field "$results_summary" 'value.actions[0].substeps[0].id')" "retained-run-preflight" "RESULTS_DIR preflight is private substep"
assert_equals "$(json_field "$results_summary" 'value.results_dir_status')" "valid" "RESULTS_DIR valid"
assert_equals "$(json_field "$results_summary" 'value.duration.status')" "refreshed" "RESULTS_DIR duration refreshed"
assert_equals "$(json_field "$results_summary" 'value.run_checks.status')" "pass" "RESULTS_DIR run checks pass"
assert_equals "$(json_field "$results_summary" 'value.actions.find((action) => action.action_id === "duration_baseline_refresh").substeps.map((substep) => substep.id).slice(-2)')" $'phase-schedules-after-duration-baselines\nphase-schedule-drift-after-duration-baselines' "RESULTS_DIR refreshes schedules after duration baselines"

assert_invalid_results_dir \
  "missing-results" \
  "$TMP_DIR/missing-results" \
  "config" \
  "configuration_error" \
  "RESULTS_DIR does not exist"

service_backed_only_dir="$TMP_DIR/service-backed-only"
write_service_backed_only_run "$service_backed_only_dir"
assert_invalid_results_dir \
  "service-backed-only" \
  "$service_backed_only_dir" \
  "config" \
  "configuration_error" \
  "must be a successful full warm make check retained run root"

failed_retained_dir="$TMP_DIR/failed-retained"
write_retained_run "$failed_retained_dir"
cat >"$failed_retained_dir/check/tool-run-summary.json" <<'JSON'
{
  "schema_id": "cartulary.tool_run_summary.v3",
  "target": "check",
  "status": "fail"
}
JSON
assert_invalid_results_dir \
  "failed-retained" \
  "$failed_retained_dir" \
  "artifact" \
  "artifact_error" \
  "must identify a passing warm check run"

incomplete_retained_dir="$TMP_DIR/incomplete-retained"
write_incomplete_retained_run "$incomplete_retained_dir"
assert_invalid_results_dir \
  "incomplete-retained" \
  "$incomplete_retained_dir" \
  "artifact" \
  "artifact_error" \
  "scheduler, target, and phase summary artifact families"

non_warm_retained_dir="$TMP_DIR/non-warm-retained"
write_retained_run "$non_warm_retained_dir"
rm "$non_warm_retained_dir/check/scheduler-summary.json"
assert_invalid_results_dir \
  "non-warm-retained" \
  "$non_warm_retained_dir" \
  "artifact" \
  "artifact_error" \
  "required for warm scheduler checks"

contaminated_retained_dir="$TMP_DIR/contaminated-retained"
write_retained_run "$contaminated_retained_dir"
mkdir -p "$contaminated_retained_dir/check-service-backed/service-session"
cat >"$contaminated_retained_dir/check-service-backed/service-session/service-scope.json" <<'JSON'
{
  "postgres": {
    "startup": {
      "final_status": "pass",
      "retry_count": 1
    }
  }
}
JSON
assert_invalid_results_dir \
  "contaminated-retained" \
  "$contaminated_retained_dir" \
  "artifact" \
  "artifact_error" \
  "contaminated timing evidence"

failure_dir="$TMP_DIR/child-fail"
mkdir -p "$failure_dir"
failure_make="$failure_dir/fake-make"
failure_log="$failure_dir/make.log"
write_fake_make "$failure_make"
set +e
CARTULARY_TEST_RESULTS_DIR="$failure_dir/results" \
CARTULARY_TEST_RUN_ID="child-fail" \
CARTULARY_PHASE_ARTIFACT_DIR="$failure_dir/results/child-fail/agent-finalize/agent-finalize" \
MAKE="$failure_make" \
FAKE_MAKE_LOG="$failure_log" \
FAKE_FAIL_TARGET="phase-schedules" \
RESULTS_DIR="" \
  "$SCRIPT" >"$failure_dir/stdout.log" 2>"$failure_dir/stderr.log"
failure_status=$?
set -e
if [[ "$failure_status" -eq 0 ]]; then
  fail "child failure unexpectedly passed"
fi
assert_equals "$(cat "$failure_log")" $'phase-ledgers\nphase-ledger-drift\nphase-schedules' "child failure fail-fast order"
failure_summary="$failure_dir/results/child-fail/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$failure_summary" 'value.failures[0].action_id')" "structure_ledger_refresh" "child failure action propagation"
assert_equals "$(json_field "$failure_summary" 'value.failures[0].substep_id')" "phase-schedules" "child failure substep propagation"
assert_equals "$(json_field "$failure_summary" 'value.failures[0].failure_class')" "product" "child failure class propagation"
assert_equals "$(json_field "$failure_summary" 'value.actions.filter((action) => action.status === "skipped").length')" "2" "child failure skipped actions"
assert_equals "$(json_field "$failure_summary" 'value.actions.flatMap((action) => action.substeps).filter((step) => step.status === "skipped").length')" "3" "child failure skipped substeps"

wrapper_dir="$TMP_DIR/wrapper"
mkdir -p "$wrapper_dir"
wrapper_make="$wrapper_dir/fake-make"
wrapper_log="$wrapper_dir/make.log"
write_fake_make "$wrapper_make"
(
  cd "$ROOT_DIR"
  CARTULARY_TEST_TARGET=agent-finalize \
  CARTULARY_TEST_RESULTS_DIR="$wrapper_dir/results" \
  CARTULARY_TEST_RUN_ID="wrapper-summary" \
  MAKE="$wrapper_make" \
  FAKE_MAKE_LOG="$wrapper_log" \
  RESULTS_DIR="" \
    "$RUN_PHASE" "agent-finalize" -- bash "$SCRIPT"
) >"$wrapper_dir/stdout.log" 2>"$wrapper_dir/stderr.log"
wrapper_output="$(cat "$wrapper_dir/stdout.log")"
assert_contains "$wrapper_output" "[FINALIZE] generated=" "wrapper summary finalize line"
assert_contains "$wrapper_output" "finalize_json=agent-finalize/finalize-summary.json" "wrapper artifact finalize ref"
assert_contains "$(cat "$wrapper_dir/results/wrapper-summary/agent-finalize/tool-run-summary.json")" "finalize_summary" "tool summary finalize artifact"

machine_dir="$TMP_DIR/machine"
mkdir -p "$machine_dir"
machine_make="$machine_dir/fake-make"
machine_log="$machine_dir/make.log"
write_fake_make "$machine_make"
(
  cd "$ROOT_DIR"
  CARTULARY_OUTPUT_MODE=machine \
  CARTULARY_TEST_TARGET=agent-finalize \
  CARTULARY_TEST_RESULTS_DIR="$machine_dir/results" \
  CARTULARY_TEST_RUN_ID="machine" \
  MAKE="$machine_make" \
  FAKE_MAKE_LOG="$machine_log" \
  RESULTS_DIR="" \
    "$RUN_PHASE" "agent-finalize" -- bash "$SCRIPT"
) >"$machine_dir/stdout.log" 2>"$machine_dir/stderr.log"
assert_equals "$(cat "$machine_dir/stderr.log")" "" "machine stderr"
machine_stdout="$(cat "$machine_dir/stdout.log")"
assert_contains "$machine_stdout" '"schema_id":"cartulary.tool_run_summary.v3"' "machine tool summary"
if [[ "$machine_stdout" == *"[FINALIZE]"* || "$machine_stdout" == *"[RESULT]"* ]]; then
  fail "machine output must not include human lines"
fi

echo "agent-finalize harness tests passed"
