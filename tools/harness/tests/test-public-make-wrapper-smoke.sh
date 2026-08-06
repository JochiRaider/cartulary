#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/../../.." && pwd)"
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "${ROOT_DIR}/tools/harness/test-support/harness-scratch.sh"

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
  env \
    -u OWNER \
    -u ROWS \
    -u JSON \
    -u PLAYWRIGHT_WORKERS \
    -u VITEST_MAX_WORKERS \
    -u CARTULARY_MAKE_INPUT_SOURCES \
    -u CARTULARY_TEST_CATALOG_ROW_IDS \
    -u CARTULARY_TEST_OWNER \
    -u CARTULARY_TEST_RUN_ID \
    -u CARTULARY_TEST_TARGET \
    -u CARTULARY_STEP_ARTIFACT_DIR \
    -u CARTULARY_HARNESS_IDENTITY_PREPARED \
    CARTULARY_TEST_RESULTS_DIR="${tmp_dir}/child-results" \
    "$@" >"${stdout_file}" 2>"${stderr_file}"
  local status=$?
  set -e
  printf '%s' "${status}"
}

tmp_dir="$(cartulary_harness_mktemp_dir "public-make-wrapper.XXXXXX")"
trap 'rm -rf "${tmp_dir}"' EXIT

success_stdout="${tmp_dir}/target-plan.stdout"
success_stderr="${tmp_dir}/target-plan.stderr"
success_status="$(
  DETAIL=logs \
  TASK_SURFACE_MANIFEST=/tmp/cartulary-ignored-task-surface.json \
    run_make_capture "${success_stdout}" "${success_stderr}" make --no-print-directory target-plan
)"
assert_equals "${success_status}" "0" "inherited undeclared env status"
assert_contains "$(cat "${success_stdout}")" "target=check" "target-plan public Make output"
assert_contains "$(cat "${success_stdout}")" "digest=sha256:" "target-plan graph digest output"
assert_equals "$(cat "${success_stderr}")" "" "inherited undeclared env stderr"

target_filter_stdout="${tmp_dir}/target-filter.stdout"
target_filter_stderr="${tmp_dir}/target-filter.stderr"
target_filter_status="$(
  run_make_capture "${target_filter_stdout}" "${target_filter_stderr}" make --no-print-directory target-plan TARGET=backend-unit
)"
assert_equals "${target_filter_status}" "0" "target-plan TARGET backend-unit status"
assert_contains "$(cat "${target_filter_stdout}")" "target=backend-unit" "target-plan TARGET backend-unit output"
assert_equals "$(cat "${target_filter_stderr}")" "" "target-plan TARGET backend-unit stderr"

make_target_stdout="${tmp_dir}/make-target.stdout"
make_target_stderr="${tmp_dir}/make-target.stderr"
make_target_status="$(
  run_make_capture "${make_target_stdout}" "${make_target_stderr}" make --no-print-directory target-plan TARGET=check
)"
assert_equals "${make_target_status}" "0" "target-plan public Make target status"
assert_contains "$(cat "${make_target_stdout}")" "target=check" "target-plan public Make target output"
assert_equals "$(cat "${make_target_stderr}")" "" "target-plan public Make target stderr"

wrong_target_stdout="${tmp_dir}/wrong-target.stdout"
wrong_target_stderr="${tmp_dir}/wrong-target.stderr"
wrong_target_status="$(
  run_make_capture "${wrong_target_stdout}" "${wrong_target_stderr}" make --no-print-directory target-plan UNDECLARED_INPUT=unexpected
)"
assert_equals "${wrong_target_status}" "2" "wrong-target Make variable status"
assert_contains "$(cat "${wrong_target_stderr}")" "CARTULARY_MAKE_INPUT_SOURCES contains unknown input UNDECLARED_INPUT" "wrong-target Make variable diagnostic"
assert_equals "$(cat "${wrong_target_stdout}")" "" "wrong-target Make variable stdout"

machine_state_root="${tmp_dir}/machine-state"
machine_state_stdout="${tmp_dir}/machine-state.stdout"
machine_state_stderr="${tmp_dir}/machine-state.stderr"
machine_state_status="$(
  run_make_capture "${machine_state_stdout}" "${machine_state_stderr}" \
    make --no-print-directory help \
      CARTULARY_MACHINE_CACHE_DIR="${machine_state_root}" \
      GO_CACHE_DIR="${machine_state_root}/go/build" \
      GO_MOD_CACHE_DIR="${machine_state_root}/go/mod" \
      GO_TMP_DIR="${machine_state_root}/go/tmp"
)"
assert_equals "${machine_state_status}" "0" "global machine-state Make inputs status"
assert_contains "$(cat "${machine_state_stdout}")" "Cartulary compact workflow task surface" "global machine-state Make inputs output"
assert_equals "$(cat "${machine_state_stderr}")" "" "global machine-state Make inputs stderr"
[[ ! -e "${machine_state_root}" ]] || fail "public preflight created machine-state paths"

native_alias_stdout="${tmp_dir}/native-alias.stdout"
native_alias_stderr="${tmp_dir}/native-alias.stderr"
native_alias_status="$(
  run_make_capture "${native_alias_stdout}" "${native_alias_stderr}" \
    make --no-print-directory help GOCACHE="${machine_state_root}/native"
)"
assert_equals "${native_alias_status}" "2" "retired native Go cache alias status"
assert_contains "$(cat "${native_alias_stderr}")" "GOCACHE is not declared for target help" "retired native Go cache alias diagnostic"
assert_equals "$(cat "${native_alias_stdout}")" "" "retired native Go cache alias stdout"

internal_stdout="${tmp_dir}/internal.stdout"
internal_stderr="${tmp_dir}/internal.stderr"
internal_status="$(
  run_make_capture "${internal_stdout}" "${internal_stderr}" make --no-print-directory target-plan TASK_SURFACE_MANIFEST=/tmp/cartulary-override.json
)"
assert_equals "${internal_status}" "2" "internal manifest override status"
assert_contains "$(cat "${internal_stderr}")" "TASK_SURFACE_MANIFEST is an internal harness input" "internal manifest override diagnostic"
assert_equals "$(cat "${internal_stdout}")" "" "internal manifest override stdout"

override_stdout="${tmp_dir}/override.stdout"
override_stderr="${tmp_dir}/override.stderr"
override_status="$(
  run_make_capture "${override_stdout}" "${override_stderr}" make --no-print-directory go-gosec-targeted GOSEC_FLAGS=-quiet
)"
assert_equals "${override_status}" "2" "undeclared static/security override status"
assert_contains "$(cat "${override_stderr}")" "GOSEC_FLAGS is not declared for target go-gosec-targeted" "undeclared static/security override diagnostic"
assert_equals "$(cat "${override_stdout}")" "" "undeclared static/security override stdout"

govuln_flags_stdout="${tmp_dir}/govuln-flags.stdout"
govuln_flags_stderr="${tmp_dir}/govuln-flags.stderr"
govuln_flags_status="$(
  run_make_capture "${govuln_flags_stdout}" "${govuln_flags_stderr}" make --no-print-directory go-vulncheck GOVULNCHECK_FLAGS=-scan=module
)"
assert_equals "${govuln_flags_status}" "2" "undeclared govulncheck flags override status"
assert_contains "$(cat "${govuln_flags_stderr}")" "GOVULNCHECK_FLAGS is not declared for target go-vulncheck" "undeclared govulncheck flags diagnostic"
assert_equals "$(cat "${govuln_flags_stdout}")" "" "undeclared govulncheck flags stdout"

govuln_db_stdout="${tmp_dir}/govuln-db.stdout"
govuln_db_stderr="${tmp_dir}/govuln-db.stderr"
govuln_db_status="$(
  run_make_capture "${govuln_db_stdout}" "${govuln_db_stderr}" env -i PATH="${PATH}" HOME="${HOME:-}" GOVULNCHECK_DB=/tmp/cartulary-vulndb CARTULARY_MAKE_INPUT_SOURCES="GOVULNCHECK_DB=cli" "${NODE_BIN:-node}" "${ROOT_DIR}/tools/harness/contract/harness-contract-cli.mjs" preflight go-vulncheck
)"
assert_equals "${govuln_db_status}" "0" "declared GOVULNCHECK_DB preflight status"
assert_equals "$(cat "${govuln_db_stdout}")" "" "declared GOVULNCHECK_DB preflight stdout"
assert_equals "$(cat "${govuln_db_stderr}")" "" "declared GOVULNCHECK_DB preflight stderr"

declared_stdout="${tmp_dir}/declared.stdout"
declared_stderr="${tmp_dir}/declared.stderr"
declared_status="$(
  run_make_capture "${declared_stdout}" "${declared_stderr}" make --no-print-directory task-guide ROLE=module-author OWNER=module.networkflow
)"
assert_equals "${declared_status}" "0" "declared target-local input status"
assert_contains "$(cat "${declared_stdout}")" "owner module.networkflow" "declared owner output"
assert_not_contains "$(cat "${declared_stderr}")" "[FAIL]" "declared target-local input stderr"
