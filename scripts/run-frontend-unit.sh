#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
source "${ROOT_DIR}/scripts/lib/run-phase-common.sh"

NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
PNPM_BIN="${PNPM:-${NODE_RUNTIME_DIR}/bin/pnpm}"
NODE_HELPER="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
VITEST_MAX_WORKERS="${VITEST_MAX_WORKERS:-2}"
VITEST_FLAGS_STRING="${VITEST_FLAGS:-}"
export NODE_BIN="${NODE_HELPER}"

if [[ ! -x "${PNPM_BIN}" ]]; then
  echo "repo-local pnpm was not found at ${PNPM_BIN}; run make frontend-toolchain" >&2
  exit 1
fi

raw_dir="$(prepare_target_support_dir raw/frontend-unit)"
run_report="${raw_dir}/runner.json"
stdout_log="${raw_dir}/stdout.log"
stderr_log="${raw_dir}/stderr.log"
output_mode="$(resolve_output_mode)"
path_prefix="${NODE_RUNTIME_DIR}/bin:${PATH}"
corepack_home="${NODE_RUNTIME_DIR}/corepack"

command=("${PNPM_BIN}" --dir apps/web exec vitest run)
command+=(--project=browser-unit --project=harness-node)
if [[ -n "${VITEST_FLAGS_STRING}" ]]; then
  # shellcheck disable=SC2206
  vitest_flag_parts=(${VITEST_FLAGS_STRING})
  command+=("${vitest_flag_parts[@]}")
fi
command+=(--maxWorkers="${VITEST_MAX_WORKERS}")

if [[ "${output_mode}" == "quiet" ]]; then
  run_command=("${command[@]}" --reporter=json --outputFile="${run_report}")
else
  run_command=("${command[@]}" --reporter=dot --reporter=json --outputFile.json="${run_report}")
fi

command_text="$(render_command env PATH="${path_prefix}" COREPACK_HOME="${corepack_home}" "${run_command[@]}")"
phase_capture_start PHASE

set +e
run_vitest_command_with_watchdog "frontend-unit" "${raw_dir}" "${stdout_log}" "${stderr_log}" "${output_mode}" env PATH="${path_prefix}" COREPACK_HOME="${corepack_home}" "${run_command[@]}"
run_status=$?
set -e

phase_capture_finish PHASE
start_time="${PHASE_START_TIME}"
end_time="${PHASE_END_TIME}"
duration_ms="${PHASE_DURATION_MS}"

status=0
export CARTULARY_REPORT_SLICE=1
export CARTULARY_PHASE_RUNNER_LOG="${run_report}"
export CARTULARY_PHASE_STDOUT_LOG="${stdout_log}"
export CARTULARY_PHASE_STDERR_LOG="${stderr_log}"
export CARTULARY_PHASE_WATCHDOG_LOG="${CARTULARY_VITEST_WATCHDOG_LOG:-}"

unset CARTULARY_VITEST_FILES || true
unset CARTULARY_VITEST_TITLES || true
unset CARTULARY_VITEST_EXCLUDE_MANIFEST_EXECUTION_DEPENDENCY || true
unset CARTULARY_VITEST_ALLOW_EMPTY_SELECTION || true
unset CARTULARY_MANIFEST_PHASE || true
unset CARTULARY_MANIFEST_COVERAGE || true
unset CARTULARY_MANIFEST_EXECUTION_DEPENDENCY || true
unset CARTULARY_PHASE_ACCOUNTING_MODE || true
export CARTULARY_PHASE_COUNTING_MODE=none
emit_report_phase_summary vitest-phase "frontend-unit" "${command_text}" "${start_time}" "${end_time}" "${duration_ms}" "${duration_ms}" "${run_status}" || status=$?
unset CARTULARY_PHASE_COUNTING_MODE || true

if [[ ! -f "${run_report}" ]]; then
  emit_target_summary fail || true
  exit "${status:-1}"
fi

export CARTULARY_PHASE_ACCOUNTING_MODE=derived
export CARTULARY_MANIFEST_COVERAGE=authoritative
export CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=frontend_unit
mapfile -t frontend_unit_phases < <("${NODE_HELPER}" "${ROOT_DIR}/scripts/lib/phase-manifest.mjs" vitest-phases authoritative frontend_unit)
for manifest_phase in "${frontend_unit_phases[@]}"; do
  export CARTULARY_MANIFEST_PHASE="${manifest_phase}"
  emit_report_phase_summary vitest-manifest-phase "frontend-unit ${manifest_phase} authoritative" "${command_text}" "${end_time}" "${end_time}" 0 0 "${run_status}" || status=$?
done

unset CARTULARY_MANIFEST_PHASE || true
unset CARTULARY_MANIFEST_COVERAGE || true
unset CARTULARY_MANIFEST_EXECUTION_DEPENDENCY || true
export CARTULARY_VITEST_EXCLUDE_MANIFEST_EXECUTION_DEPENDENCY=frontend_unit
export CARTULARY_VITEST_ALLOW_EMPTY_SELECTION=1
emit_report_phase_summary vitest-phase "frontend-unit residual" "${command_text}" "${end_time}" "${end_time}" 0 0 "${run_status}" || status=$?

if [[ "${status}" -eq 0 ]]; then
  emit_target_summary pass
  exit 0
fi

emit_target_summary fail || true
exit "${status}"
