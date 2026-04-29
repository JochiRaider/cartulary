#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "${ROOT_DIR}/scripts/lib/run-phase-common.sh"

GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}"
GO_TEST_SERVICE_PACKAGE_PARALLELISM="${GO_TEST_SERVICE_PACKAGE_PARALLELISM:-1}"
GO_TEST_PACKAGE_PARALLELISM="${GO_TEST_PACKAGE_PARALLELISM:-${GO_TEST_SERVICE_PACKAGE_PARALLELISM}}"
BACKEND_INTEGRATION_SHARD_JOBS="${BACKEND_INTEGRATION_SHARD_JOBS:-4}"
NODE_HELPER="${NODE_BIN:-}"
SHARD_PLAN_SCRIPT="${ROOT_DIR}/scripts/lib/go-shard-plan.mjs"
TARGET_PLAN_SCRIPT="${ROOT_DIR}/scripts/lib/target-plan.mjs"

if [[ -z "${NODE_HELPER}" ]]; then
  if [[ -x "${ROOT_DIR}/tmp/node-runtime/bin/node" ]]; then
    NODE_HELPER="${ROOT_DIR}/tmp/node-runtime/bin/node"
  else
    NODE_HELPER="node"
  fi
fi
export NODE_BIN="${NODE_HELPER}"

mkdir -p "${GO_CACHE_DIR}" "${GO_MOD_CACHE_DIR}"

hash_go_test_dependency_inputs() {
  {
    "${GO_BIN}" version 2>/dev/null || true
    printf '\n-- go.mod --\n'
    cat "${ROOT_DIR}/go.mod"
    printf '\n-- go.sum --\n'
    cat "${ROOT_DIR}/go.sum"
  } | if command -v sha256sum >/dev/null 2>&1; then
    sha256sum | awk '{print $1}'
  else
    cksum | awk '{print $1 "-" $2}'
  fi
}

warm_go_test_dependencies() {
  local warm_root="${GO_MOD_CACHE_DIR}/.cartulary-go-test-warm"
  local lock_dir="${warm_root}/lock"
  local timeout_seconds="${CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS:-300}"
  local start_ms
  local now_ms
  local elapsed_ms
  local owner_pid
  local warm_key
  local stamp_file
  local stamp_tmp

  if [[ ! "${timeout_seconds}" =~ ^[0-9]+$ ]] || (( timeout_seconds < 1 )); then
    echo "invalid CARTULARY_GO_TEST_WARM_LOCK_TIMEOUT_SECONDS=${timeout_seconds}" >&2
    return 2
  fi

  mkdir -p "${warm_root}"
  warm_key="$(hash_go_test_dependency_inputs)"
  stamp_file="${warm_root}/${warm_key}.stamp"
  if [[ -f "${stamp_file}" ]]; then
    return 0
  fi

  start_ms="$(phase_now_monotonic_ms)"
  while true; do
    if mkdir "${lock_dir}" 2>/dev/null; then
      printf '%s\n' "$$" >"${lock_dir}/pid"
      printf '%s\n' "$(phase_now_utc)" >"${lock_dir}/acquired_at"
      break
    fi

    owner_pid="$(cat "${lock_dir}/pid" 2>/dev/null || true)"
    if [[ "${owner_pid}" =~ ^[0-9]+$ ]] && ! kill -0 "${owner_pid}" 2>/dev/null; then
      rm -rf -- "${lock_dir}"
      continue
    fi

    now_ms="$(phase_now_monotonic_ms)"
    elapsed_ms="$(phase_elapsed_ms "${start_ms}" "${now_ms}")"
    if (( elapsed_ms >= timeout_seconds * 1000 )); then
      echo "go_test_dependency_warm_lock_timeout lock=${lock_dir}" >&2
      return 1
    fi

    sleep 0.1
  done

  if [[ -f "${stamp_file}" ]]; then
    rm -rf -- "${lock_dir}"
    return 0
  fi

  local list_status=0

  set +e
  env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" "${GO_BIN}" mod download
  local download_status=$?
  if [[ "${download_status}" -eq 0 ]]; then
    env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" "${GO_BIN}" list -deps -test ./... >/dev/null
    list_status=$?
  fi
  set -e

  if [[ "${download_status}" -ne 0 || "${list_status}" -ne 0 ]]; then
    rm -rf -- "${lock_dir}"
    if [[ "${download_status}" -ne 0 ]]; then
      return "${download_status}"
    fi
    return "${list_status}"
  fi

  stamp_tmp="${stamp_file}.$$"
  {
    printf 'warmed_at=%s\n' "$(phase_now_utc)"
    printf 'go=%s\n' "$("${GO_BIN}" version 2>/dev/null || true)"
  } >"${stamp_tmp}"
  mv -f -- "${stamp_tmp}" "${stamp_file}"
  rm -rf -- "${lock_dir}"
}

usage() {
  echo "usage: run-go-target.sh <backend-unit|backend-store|backend-integration|backend-integration-support|backend-process>" >&2
  echo "       run-go-target.sh inspect-aggregate-command <target> <execution-family-or-shard>" >&2
  echo "       run-go-target.sh capture-shard <target> <shard-name> <metadata-dir>" >&2
  echo "       run-go-target.sh finalize-shards <target> <metadata-dir>" >&2
  exit 2
}

planned_shard_names() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" list-shards "$@"
}

planned_shard_spec() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" shard-spec "$@"
}

planned_shard_postgres_fixture_policy_tests() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" shard-postgres-fixture-policy-tests "$@"
}

planned_shard_postgres_fixture_policy_packages() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" shard-postgres-fixture-policy-packages "$@"
}

planned_shard_postgres_reset_table_tests() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" shard-postgres-reset-table-tests "$@"
}

planned_shard_postgres_reset_table_packages() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" shard-postgres-reset-table-packages "$@"
}

planned_shard_field() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" shard-field "$@"
}

planned_aggregate_names() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" list-aggregates "$@"
}

planned_aggregate_shards() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" aggregate-shards "$@"
}

planned_aggregate_packages() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" aggregate-packages "$@"
}

planned_aggregate_field() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" aggregate-field "$@"
}

target_field() {
  "${NODE_HELPER}" "${TARGET_PLAN_SCRIPT}" target-field "$@"
}

target_aggregate_names() {
  "${NODE_HELPER}" "${TARGET_PLAN_SCRIPT}" list-aggregates "$@"
}

target_aggregate_spec() {
  "${NODE_HELPER}" "${TARGET_PLAN_SCRIPT}" aggregate-spec "$@"
}

target_aggregate_postgres_fixture_policy_tests() {
  "${NODE_HELPER}" "${TARGET_PLAN_SCRIPT}" aggregate-postgres-fixture-policy-tests "$@"
}

target_aggregate_postgres_fixture_policy_packages() {
  "${NODE_HELPER}" "${TARGET_PLAN_SCRIPT}" aggregate-postgres-fixture-policy-packages "$@"
}

target_aggregate_postgres_reset_table_tests() {
  "${NODE_HELPER}" "${TARGET_PLAN_SCRIPT}" aggregate-postgres-reset-table-tests "$@"
}

target_aggregate_postgres_reset_table_packages() {
  "${NODE_HELPER}" "${TARGET_PLAN_SCRIPT}" aggregate-postgres-reset-table-packages "$@"
}

target_aggregate_emission_count() {
  "${NODE_HELPER}" "${TARGET_PLAN_SCRIPT}" aggregate-emission-count "$@"
}

target_aggregate_emission_field() {
  "${NODE_HELPER}" "${TARGET_PLAN_SCRIPT}" aggregate-emission-field "$@"
}

target_aggregate_emission_packages() {
  "${NODE_HELPER}" "${TARGET_PLAN_SCRIPT}" aggregate-emission-packages "$@"
}

build_union_regex() {
  if [[ "$#" -eq 0 ]]; then
    echo "build_union_regex requires at least one regex component" >&2
    return 2
  fi

  local result=""
  local component
  for component in "$@"; do
    if [[ -z "${component}" ]]; then
      continue
    fi
    if [[ -n "${result}" ]]; then
      result="${result}|"
    fi
    result="${result}(${component})"
  done

  if [[ -z "${result}" ]]; then
    echo "build_union_regex received only empty regex components" >&2
    return 2
  fi

  printf '%s\n' "${result}"
}

join_nonempty_csv() {
  local result=""
  local component
  for component in "$@"; do
    if [[ -z "${component}" ]]; then
      continue
    fi
    if [[ -n "${result}" ]]; then
      result="${result},"
    fi
    result="${result}${component}"
  done
  printf '%s\n' "${result}"
}

set_postgres_fixture_policy_env() {
  POSTGRES_FIXTURE_POLICY_TESTS="${1:-}"
  POSTGRES_FIXTURE_POLICY_PACKAGES="${2:-}"
  POSTGRES_FIXTURE_POLICY_DEFAULT="${3:-}"
  POSTGRES_RESET_TABLES_TESTS="${4:-}"
  POSTGRES_RESET_TABLES_PACKAGES="${5:-}"
}

clear_postgres_fixture_policy_env() {
  unset POSTGRES_FIXTURE_POLICY_TESTS || true
  unset POSTGRES_FIXTURE_POLICY_PACKAGES || true
  unset POSTGRES_FIXTURE_POLICY_DEFAULT || true
  unset POSTGRES_RESET_TABLES_TESTS || true
  unset POSTGRES_RESET_TABLES_PACKAGES || true
}

render_go_test_command() {
  if [[ "$#" -lt 2 ]]; then
    echo "render_go_test_command requires <regex> -- <go test args...>" >&2
    return 2
  fi

  local test_regex="$1"
  shift

  if [[ "$1" != "--" ]]; then
    echo "render_go_test_command requires -- before go test args" >&2
    return 2
  fi
  shift

  local env_args=(
    GOCACHE="${GO_CACHE_DIR}"
    GOMODCACHE="${GO_MOD_CACHE_DIR}"
  )
  if [[ -n "${POSTGRES_FIXTURE_POLICY_TESTS:-}" ]]; then
    env_args+=(CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS="${POSTGRES_FIXTURE_POLICY_TESTS}")
  fi
  if [[ -n "${POSTGRES_FIXTURE_POLICY_PACKAGES:-}" ]]; then
    env_args+=(CARTULARY_POSTGRES_FIXTURE_POLICY_PACKAGES="${POSTGRES_FIXTURE_POLICY_PACKAGES}")
  fi
  if [[ -n "${POSTGRES_FIXTURE_POLICY_DEFAULT:-}" ]]; then
    env_args+=(CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT="${POSTGRES_FIXTURE_POLICY_DEFAULT}")
  fi
  if [[ -n "${POSTGRES_RESET_TABLES_TESTS:-}" ]]; then
    env_args+=(CARTULARY_POSTGRES_RESET_TABLES_TESTS="${POSTGRES_RESET_TABLES_TESTS}")
  fi
  if [[ -n "${POSTGRES_RESET_TABLES_PACKAGES:-}" ]]; then
    env_args+=(CARTULARY_POSTGRES_RESET_TABLES_PACKAGES="${POSTGRES_RESET_TABLES_PACKAGES}")
  fi

  render_command env "${env_args[@]}" "${GO_BIN}" test -json -run "${test_regex}" "$@"
}

acquire_shared_report_lock() {
  if [[ "$#" -ne 2 ]]; then
    echo "acquire_shared_report_lock requires <shared-dir> <shared-name>" >&2
    return 2
  fi

  local shared_dir="$1"
  local shared_name="$2"
  local lock_dir="${shared_dir}/capture.lock"
  local timeout_seconds="${CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS:-300}"
  local start_ms
  local now_ms
  local elapsed_ms
  local owner_pid

  if [[ ! "${timeout_seconds}" =~ ^[0-9]+$ ]] || (( timeout_seconds < 1 )); then
    echo "invalid CARTULARY_SHARED_REPORT_LOCK_TIMEOUT_SECONDS=${timeout_seconds}" >&2
    return 2
  fi

  start_ms="$(phase_now_monotonic_ms)"
  while true; do
    if mkdir "${lock_dir}" 2>/dev/null; then
      printf '%s\n' "$$" >"${lock_dir}/pid"
      printf '%s\n' "${shared_name}" >"${lock_dir}/shared_report"
      printf '%s\n' "$(phase_now_utc)" >"${lock_dir}/acquired_at"
      return 0
    fi

    owner_pid="$(cat "${lock_dir}/pid" 2>/dev/null || true)"
    if [[ "${owner_pid}" =~ ^[0-9]+$ ]] && ! kill -0 "${owner_pid}" 2>/dev/null; then
      rm -rf -- "${lock_dir}"
      continue
    fi

    now_ms="$(phase_now_monotonic_ms)"
    elapsed_ms="$(phase_elapsed_ms "${start_ms}" "${now_ms}")"
    if (( elapsed_ms >= timeout_seconds * 1000 )); then
      echo "shared_go_report_lock_timeout report=${shared_name} lock=${lock_dir}" >&2
      if [[ -f "${lock_dir}/pid" ]]; then
        echo "lock owner pid: $(<"${lock_dir}/pid")" >&2
      fi
      return 1
    fi

    sleep 0.1
  done
}

release_shared_report_lock() {
  if [[ "$#" -ne 1 ]]; then
    echo "release_shared_report_lock requires <shared-dir>" >&2
    return 2
  fi

  rm -rf -- "$1/capture.lock"
}

target_go_test_args() {
  if [[ "$#" -ne 2 ]]; then
    echo "target_go_test_args requires <args-var> <target>" >&2
    return 2
  fi

  local args_var="$1"
  local target="$2"
  local mode
  local -n args_ref="${args_var}"
  args_ref=()

  mode="$(target_field "${target}" goTestParallelism)"
  case "${mode}" in
    none)
      ;;
    package)
      args_ref=(-p "${GO_TEST_PACKAGE_PARALLELISM}")
      ;;
    process)
      args_ref=(-parallel 4)
      ;;
    *)
      echo "unsupported go_test_parallelism ${mode} for ${target}" >&2
      return 2
      ;;
  esac
}

resolve_execution_family_spec() {
  if [[ "$#" -ne 4 ]]; then
    echo "resolve_execution_family_spec requires <target> <execution-family-or-shard> <regex-var> <args-var>" >&2
    return 2
  fi

  local target="$1"
  local family="$2"
  local regex_var="$3"
  local args_var="$4"
  local planned_spec=()
  local target_args=()

  if [[ "${family}" == *"-shard-"* ]] && mapfile -t planned_spec < <(planned_shard_spec "${target}" "${family}" 2>/dev/null); then
    if [[ "${#planned_spec[@]}" -lt 2 ]]; then
      echo "planned shard ${family} for ${target} returned an incomplete spec" >&2
      return 2
    fi
    local -n regex_ref="${regex_var}"
    local -n args_ref="${args_var}"
    target_go_test_args target_args "${target}"
    regex_ref="${planned_spec[0]}"
    args_ref=(
      "${target_args[@]}"
      "${planned_spec[@]:1}"
    )
    return 0
  fi

  mapfile -t planned_spec < <(target_aggregate_spec "${target}" "${family}")
  if [[ "${#planned_spec[@]}" -lt 2 ]]; then
    echo "execution family ${family} for ${target} returned an incomplete spec" >&2
    return 2
  fi

  local -n regex_ref="${regex_var}"
  local -n args_ref="${args_var}"
  target_go_test_args target_args "${target}"
  regex_ref="${planned_spec[0]}"
  args_ref=(
    "${target_args[@]}"
    "${planned_spec[@]:1}"
  )
}

resolve_execution_family_postgres_fixture_policy() {
  if [[ "$#" -ne 6 ]]; then
    echo "resolve_execution_family_postgres_fixture_policy requires <target> <execution-family-or-shard> <test-policy-var> <package-policy-var> <reset-tests-var> <reset-packages-var>" >&2
    return 2
  fi

  local target="$1"
  local family="$2"
  local tests_var="$3"
  local packages_var="$4"
  local reset_tests_var="$5"
  local reset_packages_var="$6"

  local -n tests_ref="${tests_var}"
  local -n packages_ref="${packages_var}"
  local -n reset_tests_ref="${reset_tests_var}"
  local -n reset_packages_ref="${reset_packages_var}"
  tests_ref=""
  packages_ref=""
  reset_tests_ref=""
  reset_packages_ref=""

  if [[ "${family}" == *"-shard-"* ]]; then
    tests_ref="$(planned_shard_postgres_fixture_policy_tests "${target}" "${family}")"
    packages_ref="$(planned_shard_postgres_fixture_policy_packages "${target}" "${family}")"
    reset_tests_ref="$(planned_shard_postgres_reset_table_tests "${target}" "${family}")"
    reset_packages_ref="$(planned_shard_postgres_reset_table_packages "${target}" "${family}")"
    return 0
  fi

  tests_ref="$(target_aggregate_postgres_fixture_policy_tests "${target}" "${family}")"
  packages_ref="$(target_aggregate_postgres_fixture_policy_packages "${target}" "${family}")"
  reset_tests_ref="$(target_aggregate_postgres_reset_table_tests "${target}" "${family}")"
  reset_packages_ref="$(target_aggregate_postgres_reset_table_packages "${target}" "${family}")"
}

capture_go_report() {
  if [[ "$#" -lt 4 ]]; then
    echo "capture_go_report requires <shared-name> <regex> -- <go test args...>" >&2
    return 2
  fi

  local shared_name="$1"
  local shared_dir
  local capture_status

  shared_dir="$(prepare_shared_artifact_dir "${shared_name}")"
  acquire_shared_report_lock "${shared_dir}" "${shared_name}" || return $?

  set +e
  capture_go_report_locked "${shared_dir}" "$@"
  capture_status=$?
  set -e

  release_shared_report_lock "${shared_dir}" || true
  return "${capture_status}"
}

capture_go_report_locked() {
  if [[ "$#" -lt 5 ]]; then
    echo "capture_go_report_locked requires <shared-dir> <shared-name> <regex> -- <go test args...>" >&2
    return 2
  fi

  local shared_dir="$1"
  local shared_name="$2"
  local test_regex="$3"
  shift 3

  if [[ "$1" != "--" ]]; then
    echo "capture_go_report requires -- before go test args" >&2
    return 2
  fi
  shift

  local test_args=("$@")
  local shared_dir
  local runner_log
  local stderr_log
  local command_text
  local start_time
  local end_time
  local duration_ms
  local output_mode
  local run_status
  local complete_file
  local existing_command

  complete_file="${shared_dir}/complete"
  runner_log="${shared_dir}/runner.jsonl"
  stderr_log="${shared_dir}/stderr.log"
  command_text="$(render_go_test_command "${test_regex}" -- "${test_args[@]}")"

  if [[ -f "${complete_file}" ]]; then
    existing_command="$(<"${shared_dir}/command.txt")"
    if [[ "${existing_command}" != "${command_text}" ]]; then
      echo "shared_go_report_command_mismatch report=${shared_name}" >&2
      echo "shared go report ${shared_name} was created with a different command" >&2
      echo "existing: ${existing_command}" >&2
      echo "current:  ${command_text}" >&2
      return 1
    fi
    printf '%s\n' "${shared_dir}"
    printf '%s\n' reused
    return 0
  fi

  warm_go_test_dependencies

  output_mode="$(resolve_output_mode)"
  phase_capture_start PHASE

  set +e
  if [[ "${output_mode}" != "quiet" ]]; then
    env GOCACHE="${GO_CACHE_DIR}" \
      GOMODCACHE="${GO_MOD_CACHE_DIR}" \
      CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS="${POSTGRES_FIXTURE_POLICY_TESTS:-}" \
      CARTULARY_POSTGRES_FIXTURE_POLICY_PACKAGES="${POSTGRES_FIXTURE_POLICY_PACKAGES:-}" \
      CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT="${POSTGRES_FIXTURE_POLICY_DEFAULT:-}" \
      CARTULARY_POSTGRES_RESET_TABLES_TESTS="${POSTGRES_RESET_TABLES_TESTS:-}" \
      CARTULARY_POSTGRES_RESET_TABLES_PACKAGES="${POSTGRES_RESET_TABLES_PACKAGES:-}" \
      "${GO_BIN}" test -json -run "${test_regex}" "${test_args[@]}" \
      > >(tee "${runner_log}" | stream_go_json_output >&2) \
      2> >(tee "${stderr_log}" >&2)
    run_status=$?
  else
    env GOCACHE="${GO_CACHE_DIR}" \
      GOMODCACHE="${GO_MOD_CACHE_DIR}" \
      CARTULARY_POSTGRES_FIXTURE_POLICY_TESTS="${POSTGRES_FIXTURE_POLICY_TESTS:-}" \
      CARTULARY_POSTGRES_FIXTURE_POLICY_PACKAGES="${POSTGRES_FIXTURE_POLICY_PACKAGES:-}" \
      CARTULARY_POSTGRES_FIXTURE_POLICY_DEFAULT="${POSTGRES_FIXTURE_POLICY_DEFAULT:-}" \
      CARTULARY_POSTGRES_RESET_TABLES_TESTS="${POSTGRES_RESET_TABLES_TESTS:-}" \
      CARTULARY_POSTGRES_RESET_TABLES_PACKAGES="${POSTGRES_RESET_TABLES_PACKAGES:-}" \
      "${GO_BIN}" test -json -run "${test_regex}" "${test_args[@]}" \
      >"${runner_log}" \
      2>"${stderr_log}"
    run_status=$?
  fi
  set -e

  phase_capture_finish PHASE
  start_time="${PHASE_START_TIME}"
  end_time="${PHASE_END_TIME}"
  duration_ms="${PHASE_DURATION_MS}"

  printf '%s\n' "${command_text}" >"${shared_dir}/command.txt"
  printf '%s\n' "${start_time}" >"${shared_dir}/start_time.txt"
  printf '%s\n' "${end_time}" >"${shared_dir}/end_time.txt"
  printf '%s\n' "$(phase_clamp_duration_ms "${duration_ms}")" >"${shared_dir}/duration_ms.txt"
  printf '%s\n' "${run_status}" >"${shared_dir}/exit_status.txt"
  write_cross_target_shared_execution_metadata \
    "${shared_dir}" \
    "${shared_name}" \
    "${start_time}" \
    "${end_time}" \
    "${duration_ms}" \
    "${run_status}"
  touch "${complete_file}"

  printf '%s\n' "${shared_dir}"
  printf '%s\n' actual
}

is_cross_target_shared_report() {
  if [[ "$#" -ne 2 ]]; then
    echo "is_cross_target_shared_report requires <target> <shared-name>" >&2
    return 2
  fi

  local target="$1"
  local shared_name="$2"
  if [[ "${shared_name}" != *"-shard-"* ]]; then
    return 1
  fi

  local shared_across_targets
  shared_across_targets="$(planned_shard_field "${target}" "${shared_name}" shared_across_targets 2>/dev/null || true)"
  [[ "${shared_across_targets}" == "true" ]]
}

write_cross_target_shared_execution_metadata() {
  if [[ "$#" -ne 6 ]]; then
    echo "write_cross_target_shared_execution_metadata requires <shared-dir> <shared-name> <start-time> <end-time> <duration-ms> <exit-status>" >&2
    return 2
  fi

  local shared_dir="$1"
  local shared_name="$2"
  local start_time="$3"
  local end_time="$4"
  local duration_ms="$5"
  local exit_status="$6"
  local status="pass"

  if ! is_cross_target_shared_report backend-integration "${shared_name}" && \
     ! is_cross_target_shared_report backend-integration-support "${shared_name}"; then
    return 0
  fi

  if [[ "${exit_status}" != "0" ]]; then
    status="fail"
  fi

  NODE_BIN="${NODE_BIN:-}" "${TEST_OUTPUT_HELPER}" shared-execution \
    backend-integration-shards \
    "${shared_name}" \
    "${status}" \
    "${start_time}" \
    "${end_time}" \
    "$(phase_clamp_duration_ms "${duration_ms}")" \
    "${exit_status}" \
    "${shared_dir}/shared-execution.json" >/dev/null
}

assign_execution_family() {
  if [[ "$#" -ne 4 ]]; then
    echo "assign_execution_family requires <dir-var> <usage-var> <target> <execution-family-or-shard>" >&2
    return 2
  fi

  local dir_var="$1"
  local usage_var="$2"
  local target="$3"
  local shared_name="$4"
  local shared_regex
  local shared_args=()
  local policy_tests=""
  local policy_packages=""
  local reset_tests=""
  local reset_packages=""

  resolve_execution_family_spec "${target}" "${shared_name}" shared_regex shared_args
  resolve_execution_family_postgres_fixture_policy "${target}" "${shared_name}" policy_tests policy_packages reset_tests reset_packages
  set_postgres_fixture_policy_env "${policy_tests}" "${policy_packages}" "" "${reset_tests}" "${reset_packages}"
  assign_captured_report "${dir_var}" "${usage_var}" "${shared_name}" "${shared_regex}" -- "${shared_args[@]}"
  clear_postgres_fixture_policy_env

  local -n assigned_usage_ref="${usage_var}"
  if [[ "${assigned_usage_ref}" == "actual" ]] && is_cross_target_shared_report "${target}" "${shared_name}"; then
    assigned_usage_ref="reused"
  fi
}

assign_captured_report() {
  if [[ "$#" -lt 4 ]]; then
    echo "assign_captured_report requires <dir-var> <usage-var> <shared-name> <regex> -- <go test args...>" >&2
    return 2
  fi

  local -n dir_ref="$1"
  local -n usage_ref="$2"
  shift 2

  local capture_output
  local capture_result=()
  capture_output="$(capture_go_report "$@")" || return $?
  mapfile -t capture_result <<<"${capture_output}"
  if [[ "${#capture_result[@]}" -ne 2 ]]; then
    echo "capture_go_report returned incomplete metadata" >&2
    return 1
  fi

  dir_ref="${capture_result[0]}"
  usage_ref="${capture_result[1]}"
}

wait_for_one_parallel_capture() {
  if [[ "$#" -ne 1 ]]; then
    echo "wait_for_one_parallel_capture requires <pid-array-var>" >&2
    return 2
  fi

  local -n pids_ref="$1"
  local completed_pid=""
  local wait_status
  local pid
  local remaining=()

  set +e
  wait -n -p completed_pid "${pids_ref[@]}"
  wait_status=$?
  set -e

  for pid in "${pids_ref[@]}"; do
    if [[ "${pid}" != "${completed_pid}" ]]; then
      remaining+=("${pid}")
    fi
  done
  if [[ -z "${completed_pid}" ]]; then
    remaining=()
  fi
  pids_ref=("${remaining[@]}")

  return "${wait_status}"
}

capture_named_shared_reports_parallel() {
  if [[ "$#" -lt 4 ]]; then
    echo "capture_named_shared_reports_parallel requires <target> <jobs> <metadata-dir> <shared-name...>" >&2
    return 2
  fi

  local target="$1"
  local jobs="$2"
  local metadata_dir="$3"
  shift 3

  if [[ ! "${jobs}" =~ ^[0-9]+$ ]] || (( jobs < 1 )); then
    echo "invalid shard job count: ${jobs}" >&2
    return 2
  fi

  mkdir -p "${metadata_dir}"

  local pids=()
  local status=0
  local shared_name
  for shared_name in "$@"; do
    while (( ${#pids[@]} >= jobs )); do
      if ! wait_for_one_parallel_capture pids; then
        status=1
      fi
    done

    (
      local report_dir
      local report_usage
      assign_execution_family report_dir report_usage "${target}" "${shared_name}"
      printf '%s\n%s\n' "${report_dir}" "${report_usage}" >"${metadata_dir}/${shared_name}.meta"
    ) &
    pids+=("$!")
  done

  while (( ${#pids[@]} > 0 )); do
    if ! wait_for_one_parallel_capture pids; then
      status=1
    fi
  done

  return "${status}"
}

capture_scheduled_shard() {
  if [[ "$#" -ne 3 ]]; then
    echo "capture_scheduled_shard requires <target> <shared-name> <metadata-dir>" >&2
    return 2
  fi

  local target="$1"
  local shared_name="$2"
  local metadata_dir="$3"
  local report_dir
  local report_usage
  local metadata_file
  local metadata_tmp

  mkdir -p "${metadata_dir}"
  assign_execution_family report_dir report_usage "${target}" "${shared_name}"
  metadata_file="${metadata_dir}/${shared_name}.meta"
  metadata_tmp="${metadata_file}.$$"
  printf '%s\n%s\n' "${report_dir}" "${report_usage}" >"${metadata_tmp}"
  mv -f -- "${metadata_tmp}" "${metadata_file}"
}

read_shared_report_metadata() {
  if [[ "$#" -ne 4 ]]; then
    echo "read_shared_report_metadata requires <dir-var> <usage-var> <metadata-dir> <shared-name>" >&2
    return 2
  fi

  local -n dir_ref="$1"
  local -n usage_ref="$2"
  local metadata_dir="$3"
  local shared_name="$4"
  local metadata_file="${metadata_dir}/${shared_name}.meta"
  local metadata=()

  if [[ ! -f "${metadata_file}" ]]; then
    echo "missing shared report metadata for ${shared_name}" >&2
    return 1
  fi

  mapfile -t metadata <"${metadata_file}"
  if [[ "${#metadata[@]}" -ne 2 ]]; then
    echo "incomplete shared report metadata for ${shared_name}" >&2
    return 1
  fi

  dir_ref="${metadata[0]}"
  usage_ref="${metadata[1]}"
}

iso_window_duration_ms() {
  if [[ "$#" -ne 2 ]]; then
    echo "iso_window_duration_ms requires <start-time> <end-time>" >&2
    return 2
  fi

  "${NODE_HELPER}" -e '
const [start, end] = process.argv.slice(1);
const duration = Date.parse(end) - Date.parse(start);
process.stdout.write(String(Number.isFinite(duration) && duration > 0 ? duration : 0));
' "$1" "$2"
}

create_aggregate_report() {
  if [[ "$#" -lt 6 ]]; then
    echo "create_aggregate_report requires <dir-var> <usage-var> <metadata-dir> <aggregate-name> <target> <shard-name...>" >&2
    return 2
  fi

  local -n dir_ref="$1"
  local -n usage_ref="$2"
  local metadata_dir="$3"
  local aggregate_name="$4"
  local target="$5"
  shift 5

  local aggregate_root="${metadata_dir}/aggregate-reports/${target}"
  local output_dir="${aggregate_root}/${aggregate_name}"
  local runner_log="${output_dir}/runner.jsonl"
  local stderr_log="${output_dir}/stderr.log"
  local command_file="${output_dir}/command.txt"
  local start_time=""
  local end_time=""
  local duration_ms=0
  local actual_start_time=""
  local actual_end_time=""
  local actual_duration_ms=0
  local wall_duration_ms=0
  local exit_status=0
  local has_actual=0
  local shard_name
  local shard_dir
  local shard_usage
  local shard_duration
  local shard_status
  local shard_start
  local shard_end

  mkdir -p "${output_dir}"
  : >"${runner_log}"
  : >"${stderr_log}"
  : >"${command_file}"

  for shard_name in "$@"; do
    read_shared_report_metadata shard_dir shard_usage "${metadata_dir}" "${shard_name}"
    if [[ "${shard_usage}" == "actual" && "${target}" == "backend-integration-support" ]] && is_cross_target_shared_report "${target}" "${shard_name}"; then
      shard_usage="reused"
    fi
    if [[ -f "${shard_dir}/runner.jsonl" ]]; then
      cat "${shard_dir}/runner.jsonl" >>"${runner_log}"
    fi
    if [[ -f "${shard_dir}/stderr.log" ]]; then
      cat "${shard_dir}/stderr.log" >>"${stderr_log}"
    fi
    if [[ -s "${command_file}" ]]; then
      printf '\n' >>"${command_file}"
    fi
    printf '%s: %s\n' "${shard_name}" "$(<"${shard_dir}/command.txt")" >>"${command_file}"

    shard_duration="$(phase_clamp_duration_ms "$(<"${shard_dir}/duration_ms.txt")")"
    duration_ms="$((duration_ms + shard_duration))"
    shard_status="$(<"${shard_dir}/exit_status.txt")"
    if [[ "${shard_status}" != "0" ]]; then
      exit_status="${shard_status}"
    fi
    shard_start="$(<"${shard_dir}/start_time.txt")"
    shard_end="$(<"${shard_dir}/end_time.txt")"
    if [[ -z "${start_time}" || "${shard_start}" < "${start_time}" ]]; then
      start_time="${shard_start}"
    fi
    if [[ -z "${end_time}" || "${shard_end}" > "${end_time}" ]]; then
      end_time="${shard_end}"
    fi
    if [[ "${shard_usage}" == "actual" ]]; then
      has_actual=1
      actual_duration_ms="$((actual_duration_ms + shard_duration))"
      if [[ -z "${actual_start_time}" || "${shard_start}" < "${actual_start_time}" ]]; then
        actual_start_time="${shard_start}"
      fi
      if [[ -z "${actual_end_time}" || "${shard_end}" > "${actual_end_time}" ]]; then
        actual_end_time="${shard_end}"
      fi
    fi
  done

  usage_ref="reused"
  if [[ "${has_actual}" -eq 1 ]]; then
    usage_ref="actual"
    start_time="${actual_start_time}"
    end_time="${actual_end_time}"
    duration_ms="${actual_duration_ms}"
    wall_duration_ms="$(iso_window_duration_ms "${start_time}" "${end_time}")"
  fi

  printf '%s\n' "${start_time}" >"${output_dir}/start_time.txt"
  printf '%s\n' "${end_time}" >"${output_dir}/end_time.txt"
  printf '%s\n' "$(phase_clamp_duration_ms "${duration_ms}")" >"${output_dir}/duration_ms.txt"
  printf '%s\n' "$(phase_clamp_duration_ms "${wall_duration_ms}")" >"${output_dir}/wall_duration_ms.txt"
  printf '%s\n' "${exit_status}" >"${output_dir}/exit_status.txt"
  printf '%s\n' "${target}:${aggregate_name}" >"${output_dir}/aggregate.txt"

  dir_ref="${output_dir}"
}

inspect_aggregate_command() {
  if [[ "$#" -ne 2 ]]; then
    echo "inspect_aggregate_command requires <target> <execution-family-or-shard>" >&2
    return 2
  fi

  local target="$1"
  local shared_name="$2"
  local shared_regex
  local shared_args=()
  local policy_tests=""
  local policy_packages=""
  local reset_tests=""
  local reset_packages=""

  resolve_execution_family_spec "${target}" "${shared_name}" shared_regex shared_args
  resolve_execution_family_postgres_fixture_policy "${target}" "${shared_name}" policy_tests policy_packages reset_tests reset_packages
  set_postgres_fixture_policy_env "${policy_tests}" "${policy_packages}" "" "${reset_tests}" "${reset_packages}"
  render_go_test_command "${shared_regex}" -- "${shared_args[@]}"
  clear_postgres_fixture_policy_env
}

load_phase_window() {
  local report_dir="$1"
  local mode="$2"
  local stored_duration_ms
  local stored_wall_duration_ms

  PHASE_COMMAND_TEXT="$(<"${report_dir}/command.txt")"
  PHASE_EXIT_STATUS="$(<"${report_dir}/exit_status.txt")"
  stored_duration_ms="$(phase_clamp_duration_ms "$(<"${report_dir}/duration_ms.txt")")"
  stored_wall_duration_ms="${stored_duration_ms}"
  if [[ -f "${report_dir}/wall_duration_ms.txt" ]]; then
    stored_wall_duration_ms="$(phase_clamp_duration_ms "$(<"${report_dir}/wall_duration_ms.txt")")"
  fi
  if [[ "${mode}" == "actual" ]]; then
    PHASE_START_TIME="$(<"${report_dir}/start_time.txt")"
    PHASE_END_TIME="$(<"${report_dir}/end_time.txt")"
    PHASE_DURATION_MS="${stored_duration_ms}"
    PHASE_WALL_DURATION_MS="${stored_wall_duration_ms}"
    return 0
  fi

  PHASE_END_TIME="$(phase_now_utc)"
  PHASE_START_TIME="${PHASE_END_TIME}"
  PHASE_WALL_DURATION_MS=0
  if [[ "${mode}" == "reused" ]]; then
    PHASE_DURATION_MS="${stored_duration_ms}"
    return 0
  fi

  PHASE_DURATION_MS=0
}

set_go_package_patterns() {
  if [[ "$#" -eq 0 ]]; then
    export CARTULARY_GO_PACKAGE_PATTERNS=""
    return 0
  fi
  export CARTULARY_GO_PACKAGE_PATTERNS="$(printf '%s\n' "$@")"
}

clear_go_selection_env() {
  unset CARTULARY_GO_TEST_REGEX || true
  unset CARTULARY_ACCOUNTING_COVERAGE || true
  unset CARTULARY_GO_ACCOUNTING_COVERAGE || true
  unset CARTULARY_MANIFEST_PHASE || true
  unset CARTULARY_MANIFEST_SECTION || true
  unset CARTULARY_MANIFEST_COVERAGE || true
  unset CARTULARY_MANIFEST_EXECUTION_DEPENDENCY || true
  unset CARTULARY_EXECUTION_FAMILY || true
  unset CARTULARY_GO_PACKAGE_PATTERNS || true
}

emit_go_raw_phase() {
  if [[ "$#" -lt 5 ]]; then
    echo "emit_go_raw_phase requires <label> <actual|reused|derived> <report-dir> <regex> <packages...>" >&2
    return 2
  fi

  local phase_label="$1"
  local duration_mode="$2"
  local report_dir="$3"
  local test_regex="$4"
  shift 4

  load_phase_window "${report_dir}" "${duration_mode}"
  export CARTULARY_REPORT_SLICE=1
  export CARTULARY_PHASE_ACCOUNTING_MODE="${duration_mode}"
  export CARTULARY_PHASE_RUNNER_LOG="${report_dir}/runner.jsonl"
  export CARTULARY_PHASE_STDERR_LOG="${report_dir}/stderr.log"
  export CARTULARY_GO_TEST_REGEX="${test_regex}"
  export CARTULARY_ACCOUNTING_COVERAGE="${CARTULARY_GO_ACCOUNTING_COVERAGE:-}"
  set_go_package_patterns "$@"

  emit_report_phase_summary \
    go-phase \
    "${phase_label}" \
    "${PHASE_COMMAND_TEXT}" \
    "${PHASE_START_TIME}" \
    "${PHASE_END_TIME}" \
    "${PHASE_DURATION_MS}" \
    "${PHASE_WALL_DURATION_MS}" \
    "${PHASE_EXIT_STATUS}"
}

emit_go_manifest_phase() {
  if [[ "$#" -lt 9 ]]; then
    echo "emit_go_manifest_phase requires <label> <actual|reused|derived> <report-dir> <phase> <section> <coverage> <execution-dependency> <execution-family> <packages...>" >&2
    return 2
  fi

  local phase_label="$1"
  local duration_mode="$2"
  local report_dir="$3"
  local manifest_phase="$4"
  local section="$5"
  local coverage="$6"
  local execution_dependency="$7"
  local execution_family="$8"
  shift 8

  load_phase_window "${report_dir}" "${duration_mode}"
  export CARTULARY_REPORT_SLICE=1
  export CARTULARY_PHASE_ACCOUNTING_MODE="${duration_mode}"
  export CARTULARY_PHASE_RUNNER_LOG="${report_dir}/runner.jsonl"
  export CARTULARY_PHASE_STDERR_LOG="${report_dir}/stderr.log"
  export CARTULARY_MANIFEST_PHASE="${manifest_phase}"
  export CARTULARY_MANIFEST_SECTION="${section}"
  export CARTULARY_MANIFEST_COVERAGE="${coverage}"
  export CARTULARY_MANIFEST_EXECUTION_DEPENDENCY="${execution_dependency}"
  export CARTULARY_EXECUTION_FAMILY="${execution_family}"
  set_go_package_patterns "$@"

  emit_report_phase_summary \
    go-manifest-phase \
    "${phase_label}" \
    "${PHASE_COMMAND_TEXT}" \
    "${PHASE_START_TIME}" \
    "${PHASE_END_TIME}" \
    "${PHASE_DURATION_MS}" \
    "${PHASE_WALL_DURATION_MS}" \
    "${PHASE_EXIT_STATUS}"
}

emit_go_target_invocation_span() {
  local status="$1"
  if [[ -z "${GO_TARGET_INVOCATION_START_TIME:-}" || "${GO_TARGET_INVOCATION_EMITTED:-0}" == "1" ]]; then
    return 0
  fi

  local span_status="pass"
  if [[ "${status}" -ne 0 ]]; then
    span_status="fail"
  fi

  phase_capture_finish GO_TARGET_INVOCATION
  GO_TARGET_INVOCATION_EMITTED=1
  emit_target_timing_span \
    test_command \
    "run-go-target ${CARTULARY_TEST_TARGET:-unknown}" \
    "${GO_TARGET_INVOCATION_START_TIME}" \
    "${GO_TARGET_INVOCATION_END_TIME}" \
    "${GO_TARGET_INVOCATION_DURATION_MS}" \
    "${span_status}" \
    "${status}"
}

finish_target() {
  local status="$1"
  emit_go_target_invocation_span "${status}"
  if [[ "${status}" -eq 0 ]]; then
    emit_target_summary pass
    return 0
  fi

  emit_target_summary fail || true
  return "${status}"
}

emit_execution_family() {
  if [[ "$#" -ne 4 ]]; then
    echo "emit_execution_family requires <target> <execution-family> <usage> <report-dir>" >&2
    return 2
  fi

  local target="$1"
  local family="$2"
  local usage="$3"
  local report_dir="$4"
  local emission_count
  local index
  local mode
  local label
  local phase
  local section
  local coverage
  local dependency
  local support_target
  local regex
  local packages=()
  local emission_usage
  local status=0

  emission_count="$(target_aggregate_emission_count "${target}" "${family}")"
  for ((index = 0; index < emission_count; index += 1)); do
    mode="$(target_aggregate_emission_field "${target}" "${family}" "${index}" mode)"
    label="$(target_aggregate_emission_field "${target}" "${family}" "${index}" label)"
    mapfile -t packages < <(target_aggregate_emission_packages "${target}" "${family}" "${index}")
    emission_usage="derived"
    if [[ "${index}" -eq 0 ]]; then
      emission_usage="${usage}"
    fi

    clear_go_selection_env
    case "${mode}" in
      manifest)
        phase="$(target_aggregate_emission_field "${target}" "${family}" "${index}" phase)"
        section="$(target_aggregate_emission_field "${target}" "${family}" "${index}" section)"
        coverage="$(target_aggregate_emission_field "${target}" "${family}" "${index}" coverage)"
        dependency="$(target_aggregate_emission_field "${target}" "${family}" "${index}" execution_dependency)"
        emit_go_manifest_phase "${label}" "${emission_usage}" "${report_dir}" "${phase}" "${section}" "${coverage}" "${dependency}" "${family}" "${packages[@]}" || status=$?
        ;;
      support)
        phase="$(target_aggregate_emission_field "${target}" "${family}" "${index}" phase)"
        support_target="$(target_aggregate_emission_field "${target}" "${family}" "${index}" support_target)"
        regex="$(target_aggregate_emission_field "${target}" "${family}" "${index}" regex)"
        export CARTULARY_GO_ACCOUNTING_COVERAGE=support
        emit_go_raw_phase "${label}" "${emission_usage}" "${report_dir}" "${regex}" "${packages[@]}" || status=$?
        unset CARTULARY_ACCOUNTING_COVERAGE || true
        unset CARTULARY_GO_ACCOUNTING_COVERAGE || true
        ;;
      raw)
        regex="$(target_aggregate_emission_field "${target}" "${family}" "${index}" regex)"
        export CARTULARY_GO_ACCOUNTING_COVERAGE=raw
        emit_go_raw_phase "${label}" "${emission_usage}" "${report_dir}" "${regex}" "${packages[@]}" || status=$?
        unset CARTULARY_ACCOUNTING_COVERAGE || true
        unset CARTULARY_GO_ACCOUNTING_COVERAGE || true
        ;;
      *)
        echo "unsupported execution family emission mode ${mode}" >&2
        return 2
        ;;
    esac
  done

  return "${status}"
}

run_unsharded_target() {
  if [[ "$#" -ne 1 ]]; then
    echo "run_unsharded_target requires <target>" >&2
    return 2
  fi

  local target="$1"
  local aggregate_names=()
  local aggregate_name
  local aggregate_dir
  local aggregate_usage
  local status=0

  mapfile -t aggregate_names < <(target_aggregate_names "${target}")
  for aggregate_name in "${aggregate_names[@]}"; do
    assign_execution_family aggregate_dir aggregate_usage "${target}" "${aggregate_name}" || status=$?
    if [[ "${status}" -eq 0 ]]; then
      emit_execution_family "${target}" "${aggregate_name}" "${aggregate_usage}" "${aggregate_dir}" || status=$?
    fi
  done

  finish_target "${status}"
}

finalize_scheduled_shards() {
  if [[ "$#" -ne 2 ]]; then
    echo "finalize_scheduled_shards requires <target> <metadata-dir>" >&2
    return 2
  fi

  local target="$1"
  local metadata_dir="$2"
  local aggregate_names=()
  local aggregate_name
  local aggregate_dir
  local aggregate_usage
  local aggregate_shards=()
  local status=0

  mapfile -t aggregate_names < <(planned_aggregate_names "${target}")
  for aggregate_name in "${aggregate_names[@]}"; do
    mapfile -t aggregate_shards < <(planned_aggregate_shards "${target}" "${aggregate_name}")
    create_aggregate_report aggregate_dir aggregate_usage "${metadata_dir}" "${aggregate_name}" "${target}" "${aggregate_shards[@]}" || status=$?
    if [[ "${status}" -eq 0 ]]; then
      emit_execution_family "${target}" "${aggregate_name}" "${aggregate_usage}" "${aggregate_dir}" || status=$?
    fi
  done

  finish_target "${status}"
}

run_sharded_target() {
  if [[ "$#" -ne 1 ]]; then
    echo "run_sharded_target requires <target>" >&2
    return 2
  fi

  local target="$1"
  local metadata_dir
  local shard_names=()
  local status=0

  metadata_dir="$(mktemp -d "${TMPDIR:-/tmp}/cartulary-${target}-shards.XXXXXX")"
  mapfile -t shard_names < <(planned_shard_names "${target}")
  capture_named_shared_reports_parallel "${target}" "${BACKEND_INTEGRATION_SHARD_JOBS}" "${metadata_dir}" "${shard_names[@]}" || status=$?
  if [[ "${status}" -eq 0 ]]; then
    finalize_scheduled_shards "${target}" "${metadata_dir}" || status=$?
  else
    finish_target "${status}"
  fi
  rm -rf -- "${metadata_dir}"
  return "${status}"
}

main() {
  if [[ "$#" -eq 0 ]]; then
    usage
  fi

  case "$1" in
    inspect-aggregate-command)
      if [[ "$#" -ne 3 ]]; then
        usage
      fi
      inspect_aggregate_command "$2" "$3"
      ;;
    capture-shard)
      if [[ "$#" -ne 4 ]]; then
        usage
      fi
      capture_scheduled_shard "$2" "$3" "$4"
      ;;
    finalize-shards)
      if [[ "$#" -ne 3 ]]; then
        usage
      fi
      phase_capture_start GO_TARGET_INVOCATION
      finalize_scheduled_shards "$2" "$3"
      ;;
    backend-unit)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      phase_capture_start GO_TARGET_INVOCATION
      run_unsharded_target backend-unit
      ;;
    backend-store)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      phase_capture_start GO_TARGET_INVOCATION
      run_sharded_target backend-store
      ;;
    backend-integration)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      phase_capture_start GO_TARGET_INVOCATION
      run_sharded_target backend-integration
      ;;
    backend-integration-support)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      phase_capture_start GO_TARGET_INVOCATION
      run_sharded_target backend-integration-support
      ;;
    backend-process)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      phase_capture_start GO_TARGET_INVOCATION
      run_unsharded_target backend-process
      ;;
    *)
      usage
      ;;
  esac
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
