#!/usr/bin/env bash
set -euo pipefail

MAKE_BIN="${MAKE:-make}"
RESULTS_DIR="${RESULTS_DIR:-}"

run_make_target() {
  local target="$1"
  local output

  set +e
  output="$(env -u CARTULARY_TEST_TARGET CARTULARY_SUPPRESS_CHILD_SUCCESS=1 CARTULARY_OUTPUT_MODE=verbose "${MAKE_BIN}" --no-print-directory "${target}" 2>&1)"
  local status=$?
  set -e

  if [[ "${status}" -ne 0 ]]; then
    printf '%s\n' "${output}" >&2
    printf 'agent-finalize: failed at %s\n' "${target}" >&2
    return "${status}"
  fi

  printf '%s\n' "${output}"
}

if [[ -n "${RESULTS_DIR}" ]]; then
  run_make_target go-test-duration-baselines >/dev/null
  run_make_target browser-e2e-duration-baselines >/dev/null
  run_make_target service-backed-make-target-duration-baselines >/dev/null
  run_make_target harness-smoke-duration-baselines >/dev/null
  printf 'agent-finalize: duration baselines refreshed from %s\n' "${RESULTS_DIR}"
else
  printf 'agent-finalize: duration baselines skipped, RESULTS_DIR not set\n'
fi

run_make_target go-test-duration-baseline-coverage >/dev/null

phase_schedule_output="$(run_make_target phase-schedules)"
phase_schedule_summary="$(printf '%s\n' "${phase_schedule_output}" | grep -E '^phase-schedules: (unchanged|updated )' | tail -n 1 || true)"

run_make_target phase-schedule-drift >/dev/null
run_make_target json-shape-check >/dev/null

if [[ -n "${RESULTS_DIR}" ]]; then
  run_make_target go-test-duration-baseline-drift >/dev/null
  run_make_target browser-e2e-duration-baseline-drift >/dev/null
  run_make_target service-backed-make-target-duration-baseline-drift >/dev/null
  run_make_target harness-smoke-duration-baseline-drift >/dev/null
  printf 'agent-finalize: duration baselines checked from %s\n' "${RESULTS_DIR}"
fi

case "${phase_schedule_summary}" in
  "phase-schedules: unchanged")
    printf 'agent-finalize: ran, unchanged\n'
    ;;
  "phase-schedules: updated "*)
    printf 'agent-finalize: ran, updated %s\n' "${phase_schedule_summary#phase-schedules: updated }"
    ;;
  *)
    printf 'agent-finalize: ran, phase-schedules completed without a recognized update summary\n'
    ;;
esac
