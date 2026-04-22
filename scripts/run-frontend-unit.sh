#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "${ROOT_DIR}/scripts/lib/run-phase-common.sh"

PNPM_BIN="${PNPM:-}"
NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
NODE_HELPER="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
VITEST_MAX_WORKERS="${VITEST_MAX_WORKERS:-2}"
VITEST_FLAGS_STRING="${VITEST_FLAGS:-}"
export NODE_BIN="${NODE_HELPER}"

if [[ -z "${PNPM_BIN}" ]]; then
  if command -v pnpm >/dev/null 2>&1; then
    PNPM_BIN="$(command -v pnpm)"
  elif [[ -x "${HOME}/.local/share/pnpm/pnpm" ]]; then
    PNPM_BIN="${HOME}/.local/share/pnpm/pnpm"
  else
    echo "pnpm was not provided and could not be discovered" >&2
    exit 1
  fi
fi

raw_dir="$(prepare_target_support_dir raw/frontend-unit)"
run_report="${raw_dir}/runner.json"
stdout_log="${raw_dir}/stdout.log"
stderr_log="${raw_dir}/stderr.log"
output_mode="$(resolve_output_mode)"
path_prefix="${NODE_RUNTIME_DIR}/bin:${PATH}"

command=("${PNPM_BIN}" --dir apps/web exec vitest run)
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

command_text="$(render_command env PATH="${path_prefix}" "${run_command[@]}")"
start_time="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
start_ms="$(date +%s%3N)"

set +e
if [[ "${output_mode}" == "quiet" ]]; then
  env PATH="${path_prefix}" "${run_command[@]}" >"${stdout_log}" 2>"${stderr_log}"
  run_status=$?
else
  env PATH="${path_prefix}" "${run_command[@]}" > >(tee "${stdout_log}") 2> >(tee "${stderr_log}" >&2)
  run_status=$?
fi
set -e

end_time="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
end_ms="$(date +%s%3N)"
duration_ms="$((end_ms - start_ms))"

status=0
export CARTULARY_REPORT_SLICE=1
export CARTULARY_PHASE_RUNNER_LOG="${run_report}"
export CARTULARY_PHASE_STDOUT_LOG="${stdout_log}"
export CARTULARY_PHASE_STDERR_LOG="${stderr_log}"

unset CARTULARY_VITEST_FILES || true
unset CARTULARY_VITEST_TITLES || true
unset CARTULARY_MANIFEST_PHASE || true
unset CARTULARY_MANIFEST_COVERAGE || true
unset CARTULARY_MANIFEST_EXECUTION_DEPENDENCY || true
emit_report_phase_summary vitest-phase "frontend-unit" "${command_text}" "${start_time}" "${end_time}" "${duration_ms}" "${duration_ms}" "${run_status}" || status=$?

export CARTULARY_MANIFEST_PHASE=phase2
export CARTULARY_MANIFEST_COVERAGE=authoritative
export CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=frontend_unit
emit_report_phase_summary vitest-manifest-phase "frontend-unit phase2 authoritative" "${command_text}" "${end_time}" "${end_time}" 0 0 "${run_status}" || status=$?

export CARTULARY_MANIFEST_PHASE=phase3
export CARTULARY_MANIFEST_COVERAGE=authoritative
export CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=frontend_unit
emit_report_phase_summary vitest-manifest-phase "frontend-unit phase3 authoritative" "${command_text}" "${end_time}" "${end_time}" 0 0 "${run_status}" || status=$?

if [[ "${status}" -eq 0 ]]; then
  emit_target_summary pass
  exit 0
fi

emit_target_summary fail || true
exit "${status}"
