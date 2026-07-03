#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/harness/core/run-phase-common.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)/tools/harness/core/run-phase-common.sh"

usage() {
  echo "usage: run-vitest-phase.sh \"<label>\" -- <vitest run command...>" >&2
  exit 2
}

if [[ "$#" -lt 3 ]]; then
  usage
fi

phase_label="$1"
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
phase_dir="$(prepare_phase_artifact_dir "$phase_label")"
run_report="${phase_dir}/runner.json"
stdout_log="${phase_dir}/stdout.log"
stderr_log="${phase_dir}/stderr.log"

if [[ "$output_mode" != "quiet" && "${RUN_PHASE_SHOW_BANNER:-1}" == "1" ]]; then
  echo "== ${phase_label} =="
fi

if [[ "$output_mode" == "quiet" ]]; then
  run_command=("${command[@]}" --reporter=json --outputFile="$run_report")
else
  run_command=("${command[@]}" --reporter=dot --reporter=json --outputFile.json="$run_report")
fi

command_text="$(render_command "${run_command[@]}")"
phase_capture_start PHASE

set +e
run_vitest_command_with_watchdog "$phase_label" "$phase_dir" "$stdout_log" "$stderr_log" "$output_mode" "${run_command[@]}"
run_status=$?
set -e

phase_capture_finish PHASE
start_time="${PHASE_START_TIME}"
end_time="${PHASE_END_TIME}"
duration_ms="${PHASE_DURATION_MS}"

set +e
CARTULARY_PHASE_LABEL="$phase_label" \
CARTULARY_PHASE_DIR="$phase_dir" \
CARTULARY_PHASE_COMMAND="$command_text" \
CARTULARY_PHASE_START_TIME="$start_time" \
CARTULARY_PHASE_END_TIME="$end_time" \
CARTULARY_PHASE_DURATION_MS="$duration_ms" \
CARTULARY_PHASE_WALL_DURATION_MS="${PHASE_WALL_DURATION_MS}" \
CARTULARY_PHASE_EXIT_STATUS="$run_status" \
CARTULARY_PHASE_RUNNER_LOG="$run_report" \
CARTULARY_PHASE_STDOUT_LOG="$stdout_log" \
CARTULARY_PHASE_STDERR_LOG="$stderr_log" \
CARTULARY_PHASE_WATCHDOG_LOG="${CARTULARY_VITEST_WATCHDOG_LOG:-}" \
  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" vitest-phase
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
