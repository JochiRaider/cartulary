#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"

unset VERBOSE CI_VERBOSE CARTULARY_OUTPUT_MODE CARTULARY_SUPPRESS_CHILD_SUCCESS

fail() {
  echo "$*" >&2
  exit 1
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "${haystack}" != *"${needle}"* ]]; then
    fail "${label}: expected output to contain [${needle}]"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "${haystack}" == *"${needle}"* ]]; then
    fail "${label}: expected output not to contain [${needle}]"
  fi
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "${actual}" != "${expected}" ]]; then
    fail "${label}: expected [${expected}], got [${actual}]"
  fi
}

run_make_capture() {
  local stdout_file="$1"
  local stderr_file="$2"
  shift 2

  set +e
  "$@" >"${stdout_file}" 2>"${stderr_file}"
  local status=$?
  set -e
  printf '%s' "${status}"
}

tmp_dir="$(mktemp -d "${ROOT_DIR}/tmp/public-make-wrapper.XXXXXX")"
trap 'rm -rf "${tmp_dir}"' EXIT

success_stdout="${tmp_dir}/target-plan.stdout"
success_stderr="${tmp_dir}/target-plan.stderr"
success_status="$(
  PHASE=phase4 \
  DETAIL=logs \
  TASK_SURFACE_MANIFEST=/tmp/cartulary-ignored-task-surface.json \
    run_make_capture "${success_stdout}" "${success_stderr}" make --no-print-directory target-plan
)"
assert_equals "${success_status}" "0" "inherited undeclared env status"
assert_contains "$(cat "${success_stdout}")" "backend-unit" "target-plan public Make output"
assert_equals "$(cat "${success_stderr}")" "" "inherited undeclared env stderr"

wrong_target_stdout="${tmp_dir}/wrong-target.stdout"
wrong_target_stderr="${tmp_dir}/wrong-target.stderr"
wrong_target_status="$(
  run_make_capture "${wrong_target_stdout}" "${wrong_target_stderr}" make --no-print-directory target-plan PHASE=phase4
)"
assert_equals "${wrong_target_status}" "2" "wrong-target Make variable status"
assert_contains "$(cat "${wrong_target_stderr}")" "PHASE is not declared for target target-plan" "wrong-target Make variable diagnostic"
assert_equals "$(cat "${wrong_target_stdout}")" "" "wrong-target Make variable stdout"

internal_stdout="${tmp_dir}/internal.stdout"
internal_stderr="${tmp_dir}/internal.stderr"
internal_status="$(
  run_make_capture "${internal_stdout}" "${internal_stderr}" make --no-print-directory target-plan TASK_SURFACE_MANIFEST=/tmp/cartulary-override.json
)"
assert_equals "${internal_status}" "2" "internal manifest override status"
assert_contains "$(cat "${internal_stderr}")" "TASK_SURFACE_MANIFEST is an internal harness input" "internal manifest override diagnostic"
assert_equals "$(cat "${internal_stdout}")" "" "internal manifest override stdout"

declared_stdout="${tmp_dir}/declared.stdout"
declared_stderr="${tmp_dir}/declared.stderr"
declared_status="$(
  run_make_capture "${declared_stdout}" "${declared_stderr}" make --no-print-directory task-guide PHASE_NAMESPACE=frontend PHASE=FE-P3
)"
assert_equals "${declared_status}" "0" "declared target-local input status"
assert_contains "$(cat "${declared_stdout}")" "phase_namespace=frontend" "declared phase namespace output"
assert_not_contains "$(cat "${declared_stderr}")" "[FAIL]" "declared target-local input stderr"
