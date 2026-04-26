#!/usr/bin/env bash
set -euo pipefail

source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/run-phase-common.sh"

usage() {
  echo "usage: run-playwright-webserver-batch.sh <webserver-backed|functional> -- <playwright test command...>" >&2
  exit 2
}

if [[ "$#" -lt 3 ]]; then
  usage
fi

mode="$1"
shift

if [[ "$1" != "--" ]]; then
  usage
fi
shift

if [[ "$#" -eq 0 ]]; then
  usage
fi

case "$mode" in
  webserver-backed | functional)
    ;;
  *)
    usage
    ;;
esac

command=("$@")
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd "$script_dir/../.." && pwd)"
node_bin="${NODE_BIN:-node}"
manifest_script="$repo_root/scripts/lib/phase-manifest.mjs"
functional_phases=()
functional_specs=()
while IFS= read -r phase; do
  if [[ -z "$phase" ]]; then
    continue
  fi
  count="$("$node_bin" "$manifest_script" playwright-count "$phase" authoritative browser_functional)"
  if [[ "$count" == "0" ]]; then
    continue
  fi
  functional_phases+=("$phase")
  functional_specs+=("${phase}:authoritative:browser_functional")
done < <("$node_bin" "$manifest_script" list-phases)

if [[ "${#functional_specs[@]}" -eq 0 ]]; then
  echo "no authoritative browser_functional Playwright manifest rows found" >&2
  exit 1
fi

functional_grep="$("$node_bin" "$manifest_script" playwright-grep-many "${functional_specs[@]}")"
functional_files="$("$node_bin" "$manifest_script" playwright-files-many "${functional_specs[@]}")"
output_mode="$(resolve_output_mode)"
batch_dir="$(prepare_target_support_dir "playwright-${mode}-batch")"
run_report="${batch_dir}/runner.json"
stdout_log="${batch_dir}/stdout.log"
stderr_log="${batch_dir}/stderr.log"
output_dir="${batch_dir}/playwright-output"
projects=(--project functional)
if [[ "$mode" == "webserver-backed" ]]; then
  projects+=(--project support)
fi

if [[ "$output_mode" == "quiet" ]]; then
  run_command=("${command[@]}" --reporter=json --output "$output_dir" "${projects[@]}")
else
  run_command=("${command[@]}" --reporter=dot,json --output "$output_dir" "${projects[@]}")
fi

command_text="$(render_command "${run_command[@]}")"

if [[ "$output_mode" != "quiet" && "${RUN_PHASE_SHOW_BANNER:-1}" == "1" ]]; then
  echo "== browser-e2e-${mode} batch =="
fi

phase_capture_start BATCH

set +e
if [[ "$output_mode" != "quiet" ]]; then
  CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP="$functional_grep" \
  CARTULARY_PLAYWRIGHT_FUNCTIONAL_FILES="$functional_files" \
  PLAYWRIGHT_JSON_OUTPUT_FILE="$run_report" \
    "${run_command[@]}" > >(tee "$stdout_log") 2> >(tee "$stderr_log" >&2)
  run_status=$?
else
  CARTULARY_PLAYWRIGHT_FUNCTIONAL_GREP="$functional_grep" \
  CARTULARY_PLAYWRIGHT_FUNCTIONAL_FILES="$functional_files" \
  PLAYWRIGHT_JSON_OUTPUT_FILE="$run_report" \
    "${run_command[@]}" >"$stdout_log" 2>"$stderr_log"
  run_status=$?
fi
set -e

phase_capture_finish BATCH
start_time="${BATCH_START_TIME}"
end_time="${BATCH_END_TIME}"
duration_ms="${BATCH_DURATION_MS}"

if [[ ! -s "$run_report" ]]; then
  cat >"$run_report" <<JSON
{
  "suites": [],
  "errors": [
    {
      "message": "playwright did not produce a JSON report"
    }
  ]
}
JSON
fi

emit_playwright_manifest_slice() {
  local label="$1"
  local phase="$2"
  local accounting_mode="$3"
  local logical_duration_ms="$4"
  local executed_duration_ms="$5"
  local wall_duration_ms="$6"
  local phase_dir
  local selection_report
  local helper_status

  phase_dir="$(prepare_phase_artifact_dir "$label")"
  selection_report="${phase_dir}/manifest-selection.json"
  "$node_bin" "$manifest_script" playwright-selection-report "$phase" authoritative browser_functional >"$selection_report"

  set +e
  CARTULARY_REPORT_SLICE=1 \
  CARTULARY_PHASE_ACCOUNTING_MODE="$accounting_mode" \
  CARTULARY_PHASE_LABEL="$label" \
  CARTULARY_PHASE_DIR="$phase_dir" \
  CARTULARY_PHASE_COMMAND="$command_text" \
  CARTULARY_PHASE_START_TIME="$start_time" \
  CARTULARY_PHASE_END_TIME="$end_time" \
  CARTULARY_PHASE_DURATION_MS="$logical_duration_ms" \
  CARTULARY_PHASE_LOGICAL_DURATION_MS="$logical_duration_ms" \
  CARTULARY_PHASE_EXECUTED_DURATION_MS="$executed_duration_ms" \
  CARTULARY_PHASE_WALL_DURATION_MS="$wall_duration_ms" \
  CARTULARY_PHASE_EXIT_STATUS="$run_status" \
  CARTULARY_PHASE_RUNNER_LOG="$run_report" \
  CARTULARY_PLAYWRIGHT_SELECTION_REPORT="$selection_report" \
  CARTULARY_PHASE_STDOUT_LOG="$stdout_log" \
  CARTULARY_PHASE_STDERR_LOG="$stderr_log" \
  CARTULARY_PLAYWRIGHT_OUTPUT_DIR="$output_dir" \
  CARTULARY_WEB_E2E_SERVER_LOG="${CARTULARY_WEB_E2E_SERVER_LOG:-}" \
  CARTULARY_WEB_E2E_WEB_LOG="${CARTULARY_WEB_E2E_WEB_LOG:-}" \
  CARTULARY_MANIFEST_PHASE="$phase" \
  CARTULARY_MANIFEST_COVERAGE=authoritative \
  CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=browser_functional \
    NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" playwright-manifest-phase
  helper_status=$?
  set -e
  return "$helper_status"
}

emit_playwright_support_slice() {
  local label="browser-e2e-support raw"
  local phase_dir
  local helper_status
  local support_files

  phase_dir="$(prepare_phase_artifact_dir "$label")"
  support_files=$'apps/web/e2e/phase2.support.spec.ts\napps/web/e2e/phase3.support.spec.ts'

  set +e
  CARTULARY_REPORT_SLICE=1 \
  CARTULARY_PHASE_ACCOUNTING_MODE=derived \
  CARTULARY_PHASE_LABEL="$label" \
  CARTULARY_PHASE_DIR="$phase_dir" \
  CARTULARY_PHASE_COMMAND="$command_text" \
  CARTULARY_PHASE_START_TIME="$start_time" \
  CARTULARY_PHASE_END_TIME="$end_time" \
  CARTULARY_PHASE_DURATION_MS=0 \
  CARTULARY_PHASE_LOGICAL_DURATION_MS=0 \
  CARTULARY_PHASE_EXECUTED_DURATION_MS=0 \
  CARTULARY_PHASE_WALL_DURATION_MS=0 \
  CARTULARY_PHASE_EXIT_STATUS="$run_status" \
  CARTULARY_PHASE_RUNNER_LOG="$run_report" \
  CARTULARY_PHASE_STDOUT_LOG="$stdout_log" \
  CARTULARY_PHASE_STDERR_LOG="$stderr_log" \
  CARTULARY_PLAYWRIGHT_OUTPUT_DIR="$output_dir" \
  CARTULARY_PLAYWRIGHT_FILES="$support_files" \
  CARTULARY_WEB_E2E_SERVER_LOG="${CARTULARY_WEB_E2E_SERVER_LOG:-}" \
  CARTULARY_WEB_E2E_WEB_LOG="${CARTULARY_WEB_E2E_WEB_LOG:-}" \
    NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" playwright-phase
  helper_status=$?
  set -e
  return "$helper_status"
}

overall_status=0
for index in "${!functional_phases[@]}"; do
  phase="${functional_phases[$index]}"
  label="browser-e2e-functional ${phase} authoritative"
  accounting_mode=actual
  logical_ms="$duration_ms"
  executed_ms="$duration_ms"
  wall_ms="${BATCH_WALL_DURATION_MS}"
  if ! emit_playwright_manifest_slice "$label" "$phase" "$accounting_mode" "$logical_ms" "$executed_ms" "$wall_ms"; then
    overall_status=1
  fi
done

if [[ "$mode" == "webserver-backed" ]]; then
  if ! emit_playwright_support_slice; then
    overall_status=1
  fi
fi

if [[ "$overall_status" -eq 0 && "$run_status" -eq 0 ]]; then
  exit 0
fi

emit_target_summary fail || true

if [[ "$overall_status" -ne 0 ]]; then
  exit "$overall_status"
fi

exit "$run_status"
