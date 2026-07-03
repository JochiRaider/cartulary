#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
source "${ROOT_DIR}/tools/harness/core/run-phase-common.sh"

NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
PNPM_BIN="${PNPM:-${NODE_RUNTIME_DIR}/bin/pnpm}"
NODE_HELPER="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
VITEST_MAX_WORKERS="${VITEST_MAX_WORKERS:-2}"
VITEST_FLAGS_STRING="${VITEST_FLAGS:-}"
export NODE_BIN="${NODE_HELPER}"

vitest_flag_parts=()
vitest_has_path_filter=0
if [[ -n "${VITEST_FLAGS_STRING}" ]]; then
  # shellcheck disable=SC2206
  vitest_flag_parts=(${VITEST_FLAGS_STRING})
  for vitest_flag_part in "${vitest_flag_parts[@]}"; do
    if [[ "${vitest_flag_part}" != -* ]] && {
      [[ "${vitest_flag_part}" == */* ]] ||
        [[ "${vitest_flag_part}" == *.test.* ]] ||
        [[ "${vitest_flag_part}" == *.spec.* ]]
    }; then
      vitest_has_path_filter=1
    fi
  done
fi

if [[ "${vitest_has_path_filter}" -eq 1 && -z "${CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE:-}" ]]; then
  export CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE=disabled
  export CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE_NAMESPACE=base
  export CARTULARY_FRONTEND_ROW_ACCOUNTING_PHASE="${CARTULARY_PHASE_SLICE_PHASE:-phase1}"
fi

if [[ ! -x "${PNPM_BIN}" ]]; then
  echo "repo-local pnpm was not found at ${PNPM_BIN}; run make frontend-toolchain" >&2
  exit 1
fi

raw_dir="$(prepare_target_support_dir raw/frontend-unit)"
run_report="${raw_dir}/runner.json"
failure_details="${raw_dir}/vitest-failure-details.json"
stdout_log="${raw_dir}/stdout.log"
stderr_log="${raw_dir}/stderr.log"
output_mode="$(resolve_output_mode)"
path_prefix="${NODE_RUNTIME_DIR}/bin:${PATH}"
corepack_home="${NODE_RUNTIME_DIR}/corepack"

command=("${PNPM_BIN}" --dir apps/web exec vitest run)
command+=(--project=browser-unit --project=harness-node)
if [[ -n "${VITEST_FLAGS_STRING}" ]]; then
  for vitest_flag_part in "${vitest_flag_parts[@]}"; do
    if [[ "${vitest_flag_part}" == apps/web/* ]]; then
      command+=("${vitest_flag_part#apps/web/}")
    else
      command+=("${vitest_flag_part}")
    fi
  done
fi
if [[ "${CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE:-}" == "selected_rows" ]]; then
  frontend_row_ids="${CARTULARY_FRONTEND_ROW_ACCOUNTING_ROW_IDS:-}"
  if [[ -z "${frontend_row_ids}" ]]; then
    echo "selected frontend row accounting requires CARTULARY_FRONTEND_ROW_ACCOUNTING_ROW_IDS" >&2
    exit 2
  fi
  frontend_grep="$(
    "${NODE_HELPER}" "${ROOT_DIR}/scripts/lib/frontend-phase-manifest.mjs" \
      title-grep frontend-unit --row-ids "${frontend_row_ids}"
  )"
  if [[ -z "${frontend_grep}" ]]; then
    echo "no frontend-unit scenarios found for selected frontend rows: ${frontend_row_ids}" >&2
    exit 2
  fi
  command+=("-t" "${frontend_grep}")
fi
command+=(--maxWorkers="${VITEST_MAX_WORKERS}")

vitest_report_succeeded() {
  local report_file="$1"
  "${NODE_HELPER}" - "$report_file" <<'NODE'
const fs = require("node:fs");

const [reportFile] = process.argv.slice(2);
let report;
try {
  report = JSON.parse(fs.readFileSync(reportFile, "utf8"));
} catch {
  process.exit(1);
}

const numeric = (value) =>
  typeof value === "number" && Number.isFinite(value) ? value : 0;

if (
  report?.success === true &&
  numeric(report.numFailedTests) === 0 &&
  numeric(report.numFailedTestSuites) === 0
) {
  process.exit(0);
}
process.exit(1);
NODE
}

if [[ -n "${CARTULARY_PHASE_SLICE_PHASE:-}" ]]; then
  phase_status=0
  phase_label="frontend-unit ${CARTULARY_PHASE_SLICE_PHASE} authoritative"
  if ! "${ROOT_DIR}/scripts/lib/run-vitest-manifest-phase.sh" \
    "${phase_label}" \
    "${CARTULARY_PHASE_SLICE_PHASE}" \
    authoritative \
    frontend_unit \
    -- \
    env PATH="${path_prefix}" COREPACK_HOME="${corepack_home}" "${command[@]}"; then
    phase_status=1
  fi
  if [[ "${phase_status}" -eq 0 ]]; then
    emit_target_summary pass
    exit 0
  fi
  emit_target_summary fail || true
  exit "${phase_status}"
fi

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

if [[ -f "${run_report}" ]]; then
  "${NODE_HELPER}" "${ROOT_DIR}/scripts/lib/vitest-failure-details.mjs" \
    "${run_report}" "${failure_details}" "${stdout_log}" "${stderr_log}"
  if [[ "${run_status}" -ne 0 && ! -f "${CARTULARY_VITEST_WATCHDOG_LOG:-}" ]] && vitest_report_succeeded "${run_report}"; then
    run_status=0
  fi
fi

status=0
export CARTULARY_REPORT_SLICE=1
export CARTULARY_PHASE_RUNNER_LOG="${run_report}"
export CARTULARY_PHASE_VITEST_FAILURE_DETAILS="${failure_details}"
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
emit_report_phase_summary vitest-phase "frontend-unit vitest" "${command_text}" "${start_time}" "${end_time}" "${duration_ms}" "${duration_ms}" "${run_status}" || status=$?
unset CARTULARY_PHASE_COUNTING_MODE || true

if [[ ! -f "${run_report}" ]]; then
  emit_target_summary fail || true
  exit "${status:-1}"
fi

if [[ "${CARTULARY_FRONTEND_ROW_ACCOUNTING_SCOPE:-}" == "selected_rows" ]]; then
  if [[ "${status}" -eq 0 && "${run_status}" -eq 0 ]]; then
    emit_target_summary pass
    exit 0
  fi
  selected_exit_status="${status}"
  if [[ "${selected_exit_status}" -eq 0 ]]; then
    selected_exit_status="${run_status}"
  fi
  if [[ "${selected_exit_status}" -eq 0 ]]; then
    selected_exit_status=1
  fi
  emit_target_summary fail || true
  exit "${selected_exit_status}"
fi

export CARTULARY_PHASE_ACCOUNTING_MODE=derived
export CARTULARY_MANIFEST_COVERAGE=authoritative
export CARTULARY_MANIFEST_EXECUTION_DEPENDENCY=frontend_unit
if [[ "${vitest_has_path_filter}" -eq 1 ]]; then
  export CARTULARY_VITEST_ALLOW_EMPTY_SELECTION=1
fi
mapfile -t frontend_unit_phases < <("${NODE_HELPER}" "${ROOT_DIR}/scripts/lib/phase-manifest.mjs" vitest-phases authoritative frontend_unit)
for manifest_phase in "${frontend_unit_phases[@]}"; do
  export CARTULARY_MANIFEST_PHASE="${manifest_phase}"
  emit_report_phase_summary vitest-manifest-phase "frontend-unit ${manifest_phase} authoritative" "${command_text}" "${end_time}" "${end_time}" 0 0 "${run_status}" || status=$?
done
if [[ "${vitest_has_path_filter}" -eq 1 ]]; then
  unset CARTULARY_VITEST_ALLOW_EMPTY_SELECTION || true
fi

unset CARTULARY_MANIFEST_PHASE || true
unset CARTULARY_MANIFEST_COVERAGE || true
unset CARTULARY_MANIFEST_EXECUTION_DEPENDENCY || true
export CARTULARY_VITEST_EXCLUDE_MANIFEST_EXECUTION_DEPENDENCY=frontend_unit
export CARTULARY_VITEST_ALLOW_EMPTY_SELECTION=1
emit_report_phase_summary vitest-phase "frontend-unit residual" "${command_text}" "${end_time}" "${end_time}" 0 0 "${run_status}" || status=$?

if [[ "${run_status}" -ne 0 && "${status}" -eq 0 ]]; then
  status="${run_status}"
fi

if [[ "${status}" -eq 0 ]]; then
  emit_target_summary pass
  exit 0
fi

emit_target_summary fail || true
exit "${status}"
