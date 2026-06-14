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
target=""
for arg in "$@"; do
  if [[ "$arg" == --* || "$arg" == *=* ]]; then
    continue
  fi
  target="$arg"
  break
done
if [[ -z "$target" ]]; then
  printf 'fake make could not identify target in args: %s\n' "$*" >&2
  exit 2
fi
printf '%s\n' "$target" >>"$FAKE_MAKE_LOG"
if [[ -n "${FAKE_MAKE_ENV_LOG:-}" ]]; then
  printf '%s\tRESULTS_DIR=%s\tARGS=%s\n' "$target" "${RESULTS_DIR:-}" "$*" >>"$FAKE_MAKE_ENV_LOG"
fi
if [[ "${FAKE_REJECT_RESULTS_DIR_LEAK:-}" == "1" ]]; then
  case "$target" in
    phase-ledgers | phase-ledger-drift | phase-schedules | phase-schedule-drift | json-shape-check | go-test-duration-baseline-coverage)
      if [[ -n "${RESULTS_DIR:-}" || -n "${CARTULARY_MAKE_ORIGIN_RESULTS_DIR:-}" || " ${MAKEFLAGS:-} " == *" RESULTS_DIR="* ]]; then
        printf 'RESULTS_DIR leaked into non-retained substep %s\n' "$target" >&2
        exit 2
      fi
      if [[ -n "${ALLOW_OLDER_RESULTS_DIR:-}" || -n "${CARTULARY_MAKE_ORIGIN_ALLOW_OLDER_RESULTS_DIR:-}" || " ${MAKEFLAGS:-} " == *" ALLOW_OLDER_RESULTS_DIR="* ]]; then
        printf 'ALLOW_OLDER_RESULTS_DIR leaked into non-retained substep %s\n' "$target" >&2
        exit 2
      fi
      ;;
  esac
fi
if [[ -n "${FAKE_MUTATE_TRACKED_FILE:-}" && "${FAKE_MUTATE_TARGET:-}" == "$target" ]]; then
  printf 'synthetic mutation from %s\n' "$target" >"${FAKE_MUTATE_ROOT:-.}/${FAKE_MUTATE_TRACKED_FILE}"
fi
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
  "schema_id": "cartulary.check_scheduler_summary.v10",
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
  "schema_id": "cartulary.check_scheduler_summary.v10",
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
  CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1 \
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
  assert_equals "$(json_field "$summary" 'value.failures[0].action_id')" "scheduler_drift_validation" "$label failure action"
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
CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1 \
MAKE="$success_make" \
FAKE_MAKE_LOG="$success_log" \
RESULTS_DIR="" \
  "$SCRIPT"
summary="$success_dir/results/success/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$summary" 'value.schema_id')" "cartulary.agent_finalize_summary.v3" "no RESULTS_DIR schema"
assert_equals "$(json_field "$summary" 'value.actions.map((action) => action.action_id)')" $'structure_ledger_refresh\nschema_shape_validation\nduration_baseline_refresh\nduration_baseline_coverage\nduration_baseline_drift_validation\nscheduler_drift_validation' "no RESULTS_DIR action registry"
assert_equals "$(json_field "$summary" 'value.actions.filter((action) => action.execution_state === "not_selected").length')" "3" "no RESULTS_DIR retained actions not selected"
assert_equals "$(json_field "$summary" 'value.status')" "pass" "no RESULTS_DIR status"
assert_equals "$(json_field "$summary" 'value.duration.status')" "skipped" "no RESULTS_DIR duration status"
assert_equals "$(json_field "$summary" 'value.run_checks.status')" "skipped" "no RESULTS_DIR run checks"
assert_equals "$(json_field "$summary" 'value.retained_run_selection.status')" "skipped" "no RESULTS_DIR retained selection skipped"
assert_not_contains "$(cat "$success_log")" "duration-baseline-drift-suite" "no RESULTS_DIR skips retained-run drift"

cache_dir="$TMP_DIR/cache"
cache_output="$TMP_DIR/cache-output.txt"
printf 'cache output v1\n' >"$cache_output"
cache_first="$TMP_DIR/cache-first"
mkdir -p "$cache_first"
cache_first_make="$cache_first/fake-make"
cache_first_log="$cache_first/make.log"
write_fake_make "$cache_first_make"
CARTULARY_TEST_RESULTS_DIR="$cache_first/results" \
CARTULARY_TEST_RUN_ID="cache-first" \
CARTULARY_PHASE_ARTIFACT_DIR="$cache_first/results/cache-first/agent-finalize/agent-finalize" \
CARTULARY_AGENT_FINALIZE_ACTION_CACHE_DIR="$cache_dir" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT="stable" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT="$cache_output" \
MAKE="$cache_first_make" \
FAKE_MAKE_LOG="$cache_first_log" \
RESULTS_DIR="" \
  "$SCRIPT"
cache_first_summary="$cache_first/results/cache-first/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$cache_first_summary" 'value.actions.filter((action) => action.execution_state === "executed").length')" "3" "cache first run executes selected actions"
assert_equals "$(json_field "$cache_first_summary" 'value.actions.filter((action) => action.cache.state === "miss").length')" "3" "cache first run records selected misses"

cache_second="$TMP_DIR/cache-second"
mkdir -p "$cache_second"
cache_second_log="$cache_second/make.log"
CARTULARY_TEST_RESULTS_DIR="$cache_second/results" \
CARTULARY_TEST_RUN_ID="cache-second" \
CARTULARY_PHASE_ARTIFACT_DIR="$cache_second/results/cache-second/agent-finalize/agent-finalize" \
CARTULARY_AGENT_FINALIZE_ACTION_CACHE_DIR="$cache_dir" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT="stable" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT="$cache_output" \
MAKE="$cache_first_make" \
FAKE_MAKE_LOG="$cache_second_log" \
RESULTS_DIR="" \
  "$SCRIPT"
cache_second_summary="$cache_second/results/cache-second/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$cache_second_summary" 'value.actions.filter((action) => action.execution_state === "reused").length')" "3" "cache second run reuses selected actions"
assert_equals "$(json_field "$cache_second_summary" 'value.actions.filter((action) => action.cache.state === "hit").length')" "3" "cache second run reports hits"
if [[ -f "$cache_second_log" ]]; then
  fail "cache hit run must not invoke fake make"
fi

cache_disabled="$TMP_DIR/cache-disabled"
mkdir -p "$cache_disabled"
cache_disabled_make="$cache_disabled/fake-make"
cache_disabled_log="$cache_disabled/make.log"
write_fake_make "$cache_disabled_make"
CARTULARY_TEST_RESULTS_DIR="$cache_disabled/results" \
CARTULARY_TEST_RUN_ID="cache-disabled" \
CARTULARY_PHASE_ARTIFACT_DIR="$cache_disabled/results/cache-disabled/agent-finalize/agent-finalize" \
CARTULARY_AGENT_FINALIZE_ACTION_CACHE_DIR="$cache_dir" \
CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1 \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT="stable" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT="$cache_output" \
MAKE="$cache_disabled_make" \
FAKE_MAKE_LOG="$cache_disabled_log" \
RESULTS_DIR="" \
  "$SCRIPT"
cache_disabled_summary="$cache_disabled/results/cache-disabled/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$cache_disabled_summary" 'value.actions.filter((action) => action.cache.state === "disabled").length')" "3" "cache disabled reports disabled selected actions"
assert_contains "$(cat "$cache_disabled_log")" "phase-ledgers" "cache disabled executes fake make"

cache_changed="$TMP_DIR/cache-input-changed"
mkdir -p "$cache_changed"
cache_changed_make="$cache_changed/fake-make"
cache_changed_log="$cache_changed/make.log"
write_fake_make "$cache_changed_make"
CARTULARY_TEST_RESULTS_DIR="$cache_changed/results" \
CARTULARY_TEST_RUN_ID="cache-input-changed" \
CARTULARY_PHASE_ARTIFACT_DIR="$cache_changed/results/cache-input-changed/agent-finalize/agent-finalize" \
CARTULARY_AGENT_FINALIZE_ACTION_CACHE_DIR="$cache_dir" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT="changed" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT="$cache_output" \
MAKE="$cache_changed_make" \
FAKE_MAKE_LOG="$cache_changed_log" \
RESULTS_DIR="" \
  "$SCRIPT"
cache_changed_summary="$cache_changed/results/cache-input-changed/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$cache_changed_summary" 'value.actions.filter((action) => action.cache.state === "miss").length')" "3" "cache input change reports misses"
assert_contains "$(cat "$cache_changed_log")" "phase-ledgers" "cache input change executes fake make"

rm "$cache_output"
cache_output_missing="$TMP_DIR/cache-output-missing"
mkdir -p "$cache_output_missing"
cache_output_missing_make="$cache_output_missing/fake-make"
cache_output_missing_log="$cache_output_missing/make.log"
write_fake_make "$cache_output_missing_make"
CARTULARY_TEST_RESULTS_DIR="$cache_output_missing/results" \
CARTULARY_TEST_RUN_ID="cache-output-missing" \
CARTULARY_PHASE_ARTIFACT_DIR="$cache_output_missing/results/cache-output-missing/agent-finalize/agent-finalize" \
CARTULARY_AGENT_FINALIZE_ACTION_CACHE_DIR="$cache_dir" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT="stable" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT="$cache_output" \
MAKE="$cache_output_missing_make" \
FAKE_MAKE_LOG="$cache_output_missing_log" \
RESULTS_DIR="" \
  "$SCRIPT"
cache_output_missing_summary="$cache_output_missing/results/cache-output-missing/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$cache_output_missing_summary" 'value.actions.filter((action) => action.cache.reason_code === "output_missing").length')" "3" "cache output missing reason"
assert_contains "$(cat "$cache_output_missing_log")" "phase-ledgers" "cache output missing executes fake make"

printf 'cache output v1\n' >"$cache_output"
cache_corrupt_seed="$TMP_DIR/cache-corrupt-seed"
mkdir -p "$cache_corrupt_seed"
cache_corrupt_seed_make="$cache_corrupt_seed/fake-make"
write_fake_make "$cache_corrupt_seed_make"
CARTULARY_TEST_RESULTS_DIR="$cache_corrupt_seed/results" \
CARTULARY_TEST_RUN_ID="cache-corrupt-seed" \
CARTULARY_PHASE_ARTIFACT_DIR="$cache_corrupt_seed/results/cache-corrupt-seed/agent-finalize/agent-finalize" \
CARTULARY_AGENT_FINALIZE_ACTION_CACHE_DIR="$cache_dir" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT="corrupt" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT="$cache_output" \
MAKE="$cache_corrupt_seed_make" \
FAKE_MAKE_LOG="$cache_corrupt_seed/make.log" \
RESULTS_DIR="" \
  "$SCRIPT"
cache_corrupt_seed_summary="$cache_corrupt_seed/results/cache-corrupt-seed/agent-finalize/finalize-summary.json"
corrupt_record_rel="$(json_field "$cache_corrupt_seed_summary" 'value.actions[0].cache.record_path')"
printf '{not valid json\n' >"$ROOT_DIR/$corrupt_record_rel"
cache_corrupt="$TMP_DIR/cache-corrupt"
mkdir -p "$cache_corrupt"
cache_corrupt_log="$cache_corrupt/make.log"
CARTULARY_TEST_RESULTS_DIR="$cache_corrupt/results" \
CARTULARY_TEST_RUN_ID="cache-corrupt" \
CARTULARY_PHASE_ARTIFACT_DIR="$cache_corrupt/results/cache-corrupt/agent-finalize/agent-finalize" \
CARTULARY_AGENT_FINALIZE_ACTION_CACHE_DIR="$cache_dir" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_INPUT="corrupt" \
CARTULARY_AGENT_FINALIZE_TEST_CACHE_OUTPUT="$cache_output" \
MAKE="$cache_corrupt_seed_make" \
FAKE_MAKE_LOG="$cache_corrupt_log" \
RESULTS_DIR="" \
  "$SCRIPT"
cache_corrupt_summary="$cache_corrupt/results/cache-corrupt/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$cache_corrupt_summary" 'value.actions[0].cache.state')" "corrupt" "cache corrupt state"
assert_contains "$(cat "$cache_corrupt_log")" "phase-ledgers" "cache corrupt executes fake make"

retained_dir="$TMP_DIR/retained-run"
write_retained_run "$retained_dir"
results_dir="$TMP_DIR/with-results"
mkdir -p "$results_dir"
results_make="$results_dir/fake-make"
results_log="$results_dir/make.log"
results_env_log="$results_dir/make-env.log"
write_fake_make "$results_make"
CARTULARY_TEST_RESULTS_DIR="$results_dir/results" \
CARTULARY_TEST_RUN_ID="with-results" \
CARTULARY_PHASE_ARTIFACT_DIR="$results_dir/results/with-results/agent-finalize/agent-finalize" \
CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1 \
MAKE="$results_make" \
FAKE_MAKE_LOG="$results_log" \
FAKE_MAKE_ENV_LOG="$results_env_log" \
FAKE_REJECT_RESULTS_DIR_LEAK=1 \
CARTULARY_MAKE_ORIGIN_RESULTS_DIR="command line" \
CARTULARY_MAKE_ORIGIN_ALLOW_OLDER_RESULTS_DIR="undefined" \
MAKEFLAGS="--no-print-directory -- RESULTS_DIR=$retained_dir" \
RESULTS_DIR="$retained_dir" \
  "$SCRIPT"
results_summary="$results_dir/results/with-results/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$results_summary" 'value.actions.map((action) => action.action_id)')" $'scheduler_drift_validation\nstructure_ledger_refresh\nschema_shape_validation\nduration_baseline_refresh\nduration_baseline_coverage\nduration_baseline_drift_validation' "RESULTS_DIR action selection"
assert_equals "$(json_field "$results_summary" 'value.actions[0].substeps[0].id')" "retained-run-preflight" "RESULTS_DIR preflight is private substep"
assert_equals "$(json_field "$results_summary" 'value.results_dir_status')" "valid" "RESULTS_DIR valid"
assert_equals "$(json_field "$results_summary" 'value.retained_run_selection.status')" "latest" "RESULTS_DIR selected latest root"
assert_equals "$(json_field "$results_summary" 'value.retained_run_selection.supplied_is_latest')" "true" "RESULTS_DIR latest flag"
assert_equals "$(json_field "$results_summary" 'value.duration.status')" "refreshed" "RESULTS_DIR duration refreshed"
assert_equals "$(json_field "$results_summary" 'value.run_checks.status')" "pass" "RESULTS_DIR run checks pass"
assert_equals "$(json_field "$results_summary" 'value.actions.find((action) => action.action_id === "duration_baseline_refresh").substeps.map((substep) => substep.id).slice(-2)')" $'phase-schedules-after-duration-baselines\nphase-schedule-drift-after-duration-baselines' "RESULTS_DIR refreshes schedules after duration baselines"
assert_contains "$(cat "$results_env_log")" $'scheduler-event-order-drift\tRESULTS_DIR='"$retained_dir" "scheduler health substep receives retained run first"
assert_contains "$(cat "$results_env_log")" $'scheduler-summary-timing-drift\tRESULTS_DIR='"$retained_dir"$'\tARGS=--no-print-directory scheduler-summary-timing-drift TARGET=check ' "scheduler timing substep selects retained check target"
assert_contains "$(cat "$results_env_log")" $'go-test-duration-baselines\tRESULTS_DIR='"$retained_dir" "RESULTS_DIR substep receives retained run"
assert_contains "$(cat "$results_env_log")" $'phase-schedules\tRESULTS_DIR=' "RESULTS_DIR is stripped from non-retained substeps"
assert_not_contains "$(cat "$results_env_log")" $'phase-schedules\tRESULTS_DIR='"$retained_dir" "RESULTS_DIR must not leak into non-retained phase-schedules substep"

older_parent="$TMP_DIR/retained-selection"
older_retained_dir="$older_parent/20260101T000000Z-old"
newer_retained_dir="$older_parent/20260102T000000Z-new"
write_retained_run "$older_retained_dir"
write_retained_run "$newer_retained_dir"
assert_invalid_results_dir \
  "older-retained" \
  "$older_retained_dir" \
  "config" \
  "configuration_error" \
  "older than the latest successful full warm check"

older_override_dir="$TMP_DIR/older-override"
mkdir -p "$older_override_dir"
older_override_make="$older_override_dir/fake-make"
older_override_log="$older_override_dir/make.log"
write_fake_make "$older_override_make"
CARTULARY_TEST_RESULTS_DIR="$older_override_dir/results" \
CARTULARY_TEST_RUN_ID="older-override" \
CARTULARY_PHASE_ARTIFACT_DIR="$older_override_dir/results/older-override/agent-finalize/agent-finalize" \
CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1 \
MAKE="$older_override_make" \
FAKE_MAKE_LOG="$older_override_log" \
ALLOW_OLDER_RESULTS_DIR=1 \
RESULTS_DIR="$older_retained_dir" \
  "$SCRIPT"
older_override_summary="$older_override_dir/results/older-override/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$older_override_summary" 'value.retained_run_selection.status')" "older_with_override" "older override selection status"
assert_equals "$(json_field "$older_override_summary" 'value.retained_run_selection.latest_results_dir')" "$(cd "$ROOT_DIR" && realpath --relative-to="$ROOT_DIR" "$newer_retained_dir")" "older override latest root"

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
  "partial service-backed run root"

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
CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1 \
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
assert_equals "$(json_field "$failure_summary" 'value.actions.filter((action) => action.execution_state === "skipped_after_failure").length')" "2" "child failure skipped-after-failure actions"
assert_equals "$(json_field "$failure_summary" 'value.actions.filter((action) => action.execution_state === "not_selected").length')" "3" "child failure retained actions not selected"
assert_equals "$(json_field "$failure_summary" 'value.actions.flatMap((action) => action.substeps).filter((step) => step.status === "skipped").length')" "12" "child failure skipped substeps"

rollback_dir="$TMP_DIR/rollback"
mkdir -p "$rollback_dir"
rollback_make="$rollback_dir/fake-make"
rollback_log="$rollback_dir/make.log"
rollback_before="$rollback_dir/browser-baseline-before.json"
cp "$ROOT_DIR/tools/browser_e2e_duration_baselines.json" "$rollback_before"
write_fake_make "$rollback_make"
set +e
CARTULARY_TEST_RESULTS_DIR="$rollback_dir/results" \
CARTULARY_TEST_RUN_ID="rollback" \
CARTULARY_PHASE_ARTIFACT_DIR="$rollback_dir/results/rollback/agent-finalize/agent-finalize" \
CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1 \
MAKE="$rollback_make" \
FAKE_MAKE_LOG="$rollback_log" \
FAKE_MUTATE_ROOT="$ROOT_DIR" \
FAKE_MUTATE_TARGET="phase-schedules" \
FAKE_MUTATE_TRACKED_FILE="tools/browser_e2e_duration_baselines.json" \
FAKE_FAIL_TARGET="phase-schedules" \
RESULTS_DIR="" \
  "$SCRIPT" >"$rollback_dir/stdout.log" 2>"$rollback_dir/stderr.log"
rollback_status=$?
set -e
if [[ "$rollback_status" -eq 0 ]]; then
  fail "rollback mutation failure unexpectedly passed"
fi
if ! cmp -s "$rollback_before" "$ROOT_DIR/tools/browser_e2e_duration_baselines.json"; then
  fail "failed finalizer must restore tracked files mutated before failure"
fi
rollback_summary="$rollback_dir/results/rollback/agent-finalize/finalize-summary.json"
assert_equals "$(json_field "$rollback_summary" 'value.mutation_rollback.status')" "restored" "rollback summary status"
assert_equals "$(json_field "$rollback_summary" 'value.mutation_rollback.restored_file_count')" "1" "rollback restored file count"
assert_contains "$(json_field "$rollback_summary" 'value.mutation_rollback.restored_files')" "tools/browser_e2e_duration_baselines.json" "rollback restored file list"
assert_equals "$(json_field "$rollback_summary" 'value.updated_files.length')" "0" "rollback leaves no tracked updated files"

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
  CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1 \
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
  CARTULARY_AGENT_FINALIZE_DISABLE_ACTION_CACHE=1 \
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
