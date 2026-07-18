#!/usr/bin/env bash
set -euo pipefail

# shellcheck source=tools/harness/execution/step-runtime.sh
source "$(cd "$(dirname "${BASH_SOURCE[0]}")/../../.." && pwd)/tools/harness/execution/step-runtime.sh"

usage() {
  echo "usage: run-go-step.sh \"<label>\" \"<pattern>\" -- <go test command...>" >&2
  exit 2
}

if [[ "$#" -lt 4 ]]; then
  usage
fi

step="$1"
pattern="$2"
shift 2

if [[ "$1" != "--" ]]; then
  usage
fi
shift

if [[ "$#" -eq 0 ]]; then
  usage
fi

command=("$@")
test_index=-1
for i in "${!command[@]}"; do
  if [[ "${command[$i]}" == "test" ]]; then
    test_index="$i"
    break
  fi
done

if [[ "$test_index" -lt 1 ]]; then
  echo "usage: run-go-step.sh \"<label>\" \"<pattern>\" -- <go test command...>" >&2
  echo "expected a go test command after --" >&2
  exit 2
fi

prefix=("${command[@]:0:$test_index}")
suffix=("${command[@]:$((test_index + 1))}")
run_command=("${prefix[@]}" test -json -run "$pattern" "${suffix[@]}")
output_mode="$(resolve_output_mode)"
step_dir="$(prepare_step_artifact_dir "$step")"
log_file="${step_dir}/runner.jsonl"
stderr_file="${step_dir}/stderr.log"
command_text="$(render_command "${run_command[@]}")"
step_capture_start STEP

if [[ "$output_mode" != "quiet" && "${RUN_STEP_SHOW_BANNER:-1}" == "1" ]]; then
  echo "== ${step} =="
fi

set +e
if [[ "$output_mode" != "quiet" ]]; then
  "${run_command[@]}" > >(tee "$log_file" | stream_go_json_output) 2> >(tee "$stderr_file" >&2)
  run_status=$?
else
  "${run_command[@]}" >"$log_file" 2>"$stderr_file"
  run_status=$?
fi
set -e

step_capture_finish STEP
start_time="${STEP_START_TIME}"
end_time="${STEP_END_TIME}"
duration_ms="${STEP_DURATION_MS}"

set +e
CARTULARY_STEP_LABEL="$step" \
CARTULARY_STEP_DIR="$step_dir" \
CARTULARY_STEP_COMMAND="$command_text" \
CARTULARY_STEP_START_TIME="$start_time" \
CARTULARY_STEP_END_TIME="$end_time" \
CARTULARY_STEP_LOGICAL_DURATION_MS="$duration_ms" \
CARTULARY_STEP_EXECUTED_DURATION_MS="$duration_ms" \
CARTULARY_STEP_WALL_DURATION_MS="${STEP_WALL_DURATION_MS}" \
CARTULARY_STEP_EXIT_STATUS="$run_status" \
CARTULARY_STEP_RUNNER_LOG="$log_file" \
CARTULARY_STEP_STDERR_LOG="$stderr_file" \
  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" go-step
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
