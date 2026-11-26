#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/harness/execution/step-runtime.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)/tools/harness/execution/step-runtime.sh"

usage() {
  echo "usage: run-vitest-step.sh \"<label>\" -- <exact invocation request.json>" >&2
  exit 2
}

if [[ "$#" -ne 3 ]]; then
  usage
fi

step_label="$1"
shift

if [[ "$1" != "--" ]]; then
  usage
fi
shift

if [[ "$#" -ne 1 ]]; then
  usage
fi

request_file="$1"
output_mode="$(resolve_output_mode)"
step_dir="$(prepare_step_artifact_dir "$step_label")"
run_report="${step_dir}/runner.json"
stdout_log="${step_dir}/stdout.log"
stderr_log="${step_dir}/stderr.log"

if [[ "$output_mode" != "quiet" && "${RUN_STEP_SHOW_BANNER:-1}" == "1" ]]; then
  echo "== ${step_label} =="
fi

failure_details="${step_dir}/vitest-failure-details.json"
run_command=("${NODE_BIN:-node}" "${RUN_STEP_REPO_ROOT}/tools/harness/execution/vitest-invocation-cli.mjs" "$run_report" "$failure_details" "$request_file")

command_text="$(render_command "${run_command[@]}")"
step_capture_start STEP

set +e
run_vitest_command_with_watchdog "$step_label" "$step_dir" "$stdout_log" "$stderr_log" "$output_mode" "${run_command[@]}"
run_status=$?
set -e

step_capture_finish STEP
start_time="${STEP_START_TIME}"
end_time="${STEP_END_TIME}"
duration_ms="${STEP_DURATION_MS}"

set +e
CARTULARY_STEP_VITEST_FAILURE_DETAILS="$failure_details" \
CARTULARY_STEP_VITEST_REQUEST="$request_file" \
CARTULARY_STEP_WATCHDOG_LOG="${CARTULARY_VITEST_WATCHDOG_LOG:-}" \
  step_publish_summary vitest-step label="$step_label" dir="$step_dir" command="$command_text" \
    start_time="$start_time" end_time="$end_time" duration_ms="$duration_ms" wall_duration_ms="${STEP_WALL_DURATION_MS}" \
    exit_status="$run_status" runner_log="$run_report" stdout_log="$stdout_log" stderr_log="$stderr_log" \
    interrupt_signal="${CARTULARY_VITEST_INTERRUPT_SIGNAL:-}"
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
