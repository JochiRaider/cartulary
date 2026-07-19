#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
source "${ROOT_DIR}/tools/harness/execution/step-runtime.sh"

NODE_RUNTIME_DIR="${NODE_RUNTIME_DIR:-${ROOT_DIR}/tmp/node-runtime}"
PNPM_BIN="${PNPM:-${NODE_RUNTIME_DIR}/bin/pnpm}"
NODE_HELPER="${NODE_BIN:-${NODE_RUNTIME_DIR}/bin/node}"
VITEST_MAX_WORKERS="${VITEST_MAX_WORKERS:-4}"
export NODE_BIN="${NODE_HELPER}"

if [[ ! "${VITEST_MAX_WORKERS}" =~ ^[0-9]+$ ]] ||
  [[ "${VITEST_MAX_WORKERS}" -lt 1 ]] ||
  [[ "${VITEST_MAX_WORKERS}" -gt 16 ]]; then
  echo "VITEST_MAX_WORKERS must be an integer from 1 through 16" >&2
  exit 2
fi
unset VITEST_FLAGS || true

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

if [[ "${output_mode}" == "quiet" ]]; then
  run_command=("${command[@]}" --reporter=json --outputFile="${run_report}")
else
  run_command=("${command[@]}" --reporter=dot --reporter=json --outputFile.json="${run_report}")
fi

command_text="$(render_command env PATH="${path_prefix}" COREPACK_HOME="${corepack_home}" "${run_command[@]}")"
step_capture_start STEP

set +e
run_vitest_command_with_watchdog "frontend-unit" "${raw_dir}" "${stdout_log}" "${stderr_log}" "${output_mode}" env PATH="${path_prefix}" COREPACK_HOME="${corepack_home}" "${run_command[@]}"
run_status=$?
set -e

step_capture_finish STEP
start_time="${STEP_START_TIME}"
end_time="${STEP_END_TIME}"
duration_ms="${STEP_DURATION_MS}"

if [[ -f "${run_report}" ]]; then
  "${NODE_HELPER}" "${ROOT_DIR}/tools/harness/diagnostics/vitest-failure-details.mjs" \
    "${run_report}" "${failure_details}" "${stdout_log}" "${stderr_log}"
  if [[ "${run_status}" -ne 0 && ! -f "${CARTULARY_VITEST_WATCHDOG_LOG:-}" ]] && vitest_report_succeeded "${run_report}"; then
    run_status=0
  fi
fi

status=0
export CARTULARY_REPORT_SLICE=1
export CARTULARY_STEP_RUNNER_LOG="${run_report}"
export CARTULARY_STEP_VITEST_FAILURE_DETAILS="${failure_details}"
export CARTULARY_STEP_STDOUT_LOG="${stdout_log}"
export CARTULARY_STEP_STDERR_LOG="${stderr_log}"
export CARTULARY_STEP_WATCHDOG_LOG="${CARTULARY_VITEST_WATCHDOG_LOG:-}"
export CARTULARY_STEP_INTERRUPT_SIGNAL="${CARTULARY_VITEST_INTERRUPT_SIGNAL:-}"

unset CARTULARY_VITEST_FILES || true
unset CARTULARY_VITEST_TITLES || true
unset CARTULARY_VITEST_EXCLUDE_MANIFEST_EXECUTION_DEPENDENCY || true
unset CARTULARY_VITEST_ALLOW_EMPTY_SELECTION || true
unset CARTULARY_CATALOG_OWNER_ID || true
unset CARTULARY_MANIFEST_COVERAGE || true
unset CARTULARY_MANIFEST_EXECUTION_DEPENDENCY || true
unset CARTULARY_STEP_ACCOUNTING_MODE || true
export CARTULARY_STEP_COUNTING_MODE=none
emit_report_step_summary vitest-step "frontend-unit vitest" "${command_text}" "${start_time}" "${end_time}" "${duration_ms}" "${duration_ms}" "${run_status}" || status=$?
unset CARTULARY_STEP_COUNTING_MODE || true

if [[ ! -f "${run_report}" ]]; then
  emit_target_summary fail || true
  exit "${status:-1}"
fi

if [[ "${run_status}" -ne 0 && "${status}" -eq 0 ]]; then
  status="${run_status}"
fi

export CARTULARY_TARGET_EVIDENCE_FINALIZE=1
if [[ "${status}" -eq 0 ]]; then
  emit_target_summary pass
  exit 0
fi

emit_target_summary fail || true
exit "${status}"
