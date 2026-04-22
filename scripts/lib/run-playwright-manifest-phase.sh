#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/run-phase-common.sh"

usage() {
  echo "usage: run-playwright-manifest-phase.sh \"<label>\" <phase> <coverage> [<execution_dependency>] -- <playwright test command...>" >&2
  exit 2
}

if [[ "$#" -lt 5 ]]; then
  usage
fi

phase_label="$1"
phase_manifest="$2"
coverage="$3"
shift 3

execution_dependency=""
if [[ "$1" != "--" ]]; then
  execution_dependency="$1"
  shift
fi

if [[ "$1" != "--" ]]; then
  usage
fi
shift

if [[ "$#" -eq 0 ]]; then
  usage
fi

command=("$@")
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
node_bin="${NODE_BIN:-node}"
manifest_script="$repo_root/scripts/lib/phase-manifest.mjs"

mapfile -t manifest_files < <("$node_bin" "$manifest_script" playwright-files "$phase_manifest" "$coverage" "$execution_dependency")
grep_pattern="$("$node_bin" "$manifest_script" playwright-grep "$phase_manifest" "$coverage" "$execution_dependency")"
output_mode="$(resolve_output_mode)"
phase_dir="$(prepare_phase_artifact_dir "$phase_label")"
list_report="${phase_dir}/list.json"
run_report="${phase_dir}/runner.json"
stdout_log="${phase_dir}/stdout.log"
stderr_log="${phase_dir}/stderr.log"
output_dir="${phase_dir}/playwright-output"

if [[ "$output_mode" != "quiet" && "${RUN_PHASE_SHOW_BANNER:-1}" == "1" ]]; then
  echo "== ${phase_label} =="
fi

list_command=("${command[@]}" --list --reporter=json -g "$grep_pattern" "${manifest_files[@]}")
if [[ "$output_mode" == "quiet" ]]; then
  run_command=("${command[@]}" --reporter=json --output "$output_dir" -g "$grep_pattern" "${manifest_files[@]}")
else
  run_command=("${command[@]}" --reporter=dot,json --output "$output_dir" -g "$grep_pattern" "${manifest_files[@]}")
fi
list_command_text="$(render_command "${list_command[@]}")"
run_command_text="$(render_command "${run_command[@]}")"
start_time="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
start_ms="$(date +%s%3N)"

set +e
PLAYWRIGHT_JSON_OUTPUT_FILE="$list_report" "${list_command[@]}" >"$stdout_log" 2>"$stderr_log"
list_status=$?
set -e
if [[ "$list_status" -ne 0 ]]; then
  end_time="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  end_ms="$(date +%s%3N)"
  duration_ms="$((end_ms - start_ms))"
  set +e
  CARTULARY_PHASE_LABEL="$phase_label" \
  CARTULARY_PHASE_DIR="$phase_dir" \
  CARTULARY_PHASE_COMMAND="$list_command_text" \
  CARTULARY_PHASE_START_TIME="$start_time" \
  CARTULARY_PHASE_END_TIME="$end_time" \
  CARTULARY_PHASE_DURATION_MS="$duration_ms" \
  CARTULARY_PHASE_WALL_DURATION_MS="$duration_ms" \
  CARTULARY_PHASE_EXIT_STATUS="$list_status" \
  CARTULARY_PHASE_STDOUT_LOG="$stdout_log" \
  CARTULARY_PHASE_STDERR_LOG="$stderr_log" \
    NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" shell-phase
  set -e
  emit_target_summary fail || true
  exit "$list_status"
fi

set +e
if [[ "$output_mode" != "quiet" ]]; then
  PLAYWRIGHT_JSON_OUTPUT_FILE="$run_report" "${run_command[@]}" > >(tee -a "$stdout_log") 2> >(tee -a "$stderr_log" >&2)
  run_status=$?
else
  PLAYWRIGHT_JSON_OUTPUT_FILE="$run_report" "${run_command[@]}" >>"$stdout_log" 2>>"$stderr_log"
  run_status=$?
fi
set -e

end_time="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
end_ms="$(date +%s%3N)"
duration_ms="$((end_ms - start_ms))"

set +e
CARTULARY_PHASE_LABEL="$phase_label" \
CARTULARY_PHASE_DIR="$phase_dir" \
CARTULARY_PHASE_COMMAND="$run_command_text" \
CARTULARY_PHASE_START_TIME="$start_time" \
CARTULARY_PHASE_END_TIME="$end_time" \
CARTULARY_PHASE_DURATION_MS="$duration_ms" \
CARTULARY_PHASE_WALL_DURATION_MS="$duration_ms" \
CARTULARY_PHASE_EXIT_STATUS="$run_status" \
CARTULARY_PHASE_RUNNER_LOG="$run_report" \
CARTULARY_PLAYWRIGHT_LIST_REPORT="$list_report" \
CARTULARY_PHASE_STDOUT_LOG="$stdout_log" \
CARTULARY_PHASE_STDERR_LOG="$stderr_log" \
CARTULARY_PLAYWRIGHT_OUTPUT_DIR="$output_dir" \
CARTULARY_WEB_E2E_SERVER_LOG="${CARTULARY_WEB_E2E_SERVER_LOG:-}" \
CARTULARY_WEB_E2E_WEB_LOG="${CARTULARY_WEB_E2E_WEB_LOG:-}" \
CARTULARY_MANIFEST_PHASE="$phase_manifest" \
CARTULARY_MANIFEST_COVERAGE="$coverage" \
CARTULARY_MANIFEST_EXECUTION_DEPENDENCY="$execution_dependency" \
  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" playwright-manifest-phase
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
