#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/harness/execution/step-runtime.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)/tools/harness/execution/step-runtime.sh"

usage() {
  echo "usage: run-playwright-step.sh \"<label>\" -- <playwright test command...>" >&2
  exit 2
}

if [[ "$#" -lt 3 ]]; then
  usage
fi

step_label="$1"
shift

if [[ "$1" != "--" ]]; then
  usage
fi
shift

if [[ "$#" -eq 0 ]]; then
  usage
fi

command=("$@")
output_mode="$(resolve_output_mode)"
step_dir="$(prepare_step_artifact_dir "$step_label")"
run_report="${step_dir}/runner.json"
stdout_log="${step_dir}/stdout.log"
stderr_log="${step_dir}/stderr.log"
output_dir="${step_dir}/playwright-output"

if [[ "$output_mode" != "quiet" && "${RUN_STEP_SHOW_BANNER:-1}" == "1" ]]; then
  echo "== ${step_label} =="
fi

if [[ "$output_mode" == "quiet" ]]; then
  run_command=("${command[@]}" --reporter=json --output "$output_dir")
else
  run_command=("${command[@]}" "--reporter=dot,json" --output "$output_dir")
fi

command_text="$(render_command "${run_command[@]}")"
step_capture_start STEP

set +e
if [[ "$output_mode" != "quiet" ]]; then
  PLAYWRIGHT_JSON_OUTPUT_FILE="$run_report" "${run_command[@]}" > >(tee "$stdout_log") 2> >(tee "$stderr_log" >&2)
  run_status=$?
else
  PLAYWRIGHT_JSON_OUTPUT_FILE="$run_report" "${run_command[@]}" >"$stdout_log" 2>"$stderr_log"
  run_status=$?
fi
set -e

step_capture_finish STEP
start_time="${STEP_START_TIME}"
end_time="${STEP_END_TIME}"
duration_ms="${STEP_DURATION_MS}"

set +e
CARTULARY_STEP_LABEL="$step_label" \
CARTULARY_STEP_DIR="$step_dir" \
CARTULARY_STEP_COMMAND="$command_text" \
CARTULARY_STEP_START_TIME="$start_time" \
CARTULARY_STEP_END_TIME="$end_time" \
CARTULARY_STEP_LOGICAL_DURATION_MS="$duration_ms" \
CARTULARY_STEP_EXECUTED_DURATION_MS="$duration_ms" \
CARTULARY_STEP_WALL_DURATION_MS="${STEP_WALL_DURATION_MS}" \
CARTULARY_STEP_EXIT_STATUS="$run_status" \
CARTULARY_STEP_RUNNER_LOG="$run_report" \
CARTULARY_STEP_STDOUT_LOG="$stdout_log" \
CARTULARY_STEP_STDERR_LOG="$stderr_log" \
CARTULARY_PLAYWRIGHT_OUTPUT_DIR="$output_dir" \
CARTULARY_WEB_E2E_SERVER_LOG="${CARTULARY_WEB_E2E_SERVER_LOG:-}" \
CARTULARY_WEB_E2E_WEB_LOG="${CARTULARY_WEB_E2E_WEB_LOG:-}" \
  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" playwright-step
helper_status=$?
set -e

if [[ "$run_status" -eq 0 && "$helper_status" -eq 0 ]]; then
  exit 0
fi

emit_target_summary fail || true

if [[ "$run_status" -ne 0 ]]; then
  exit "$run_status"
fi

exit "$helper_status"
