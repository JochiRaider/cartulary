#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "${ROOT_DIR}/scripts/lib/run-phase-common.sh"

GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}"
GO_TEST_SERVICE_PACKAGE_PARALLELISM="${GO_TEST_SERVICE_PACKAGE_PARALLELISM:-1}"
GO_TEST_PACKAGE_PARALLELISM="${GO_TEST_PACKAGE_PARALLELISM:-${GO_TEST_SERVICE_PACKAGE_PARALLELISM}}"
NODE_HELPER="${NODE_BIN:-}"
MANIFEST_SCRIPT="${ROOT_DIR}/scripts/lib/phase-manifest.mjs"

if [[ -z "${NODE_HELPER}" ]]; then
  if [[ -x "${ROOT_DIR}/tmp/node-runtime/bin/node" ]]; then
    NODE_HELPER="${ROOT_DIR}/tmp/node-runtime/bin/node"
  else
    NODE_HELPER="node"
  fi
fi
export NODE_BIN="${NODE_HELPER}"

mkdir -p "${GO_CACHE_DIR}" "${GO_MOD_CACHE_DIR}"

usage() {
  echo "usage: run-go-target.sh <backend-unit|backend-store|backend-integration|backend-integration-support|backend-process|phase0-process-e2e|phase1-process-smoke|phase2-process-smoke>" >&2
  echo "       run-go-target.sh inspect-shared-command <target> <shared-name>" >&2
  exit 2
}

manifest_go_regex() {
  "${NODE_HELPER}" "${MANIFEST_SCRIPT}" go-regex "$@"
}

manifest_go_count() {
  "${NODE_HELPER}" "${MANIFEST_SCRIPT}" go-count "$@"
}

support_go_regex() {
  "${NODE_HELPER}" "${MANIFEST_SCRIPT}" support-go-regex "$@"
}

support_go_count() {
  "${NODE_HELPER}" "${MANIFEST_SCRIPT}" support-go-count "$@"
}

SUPPORT_MANIFEST_PHASES=(phase0 phase1 phase2 phase3 phase4)

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

append_declared_support_regex_components() {
  if [[ "$#" -lt 2 ]]; then
    echo "append_declared_support_regex_components requires <components-var> <target> <packages...>" >&2
    return 2
  fi

  local -n components_ref="$1"
  shift

  local target="$1"
  shift

  local phase
  local count
  for phase in "${SUPPORT_MANIFEST_PHASES[@]}"; do
    count="$(support_go_count "${phase}" "${target}" "$@")"
    if [[ "${count}" == "0" ]]; then
      continue
    fi
    components_ref+=("$(support_go_regex "${phase}" "${target}" "$@")")
  done
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

  render_command env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" "${GO_BIN}" test -json -run "${test_regex}" "$@"
}

emit_declared_support_phase() {
  if [[ "$#" -lt 6 ]]; then
    echo "emit_declared_support_phase requires <label> <actual|reused|derived> <report-dir> <phase> <target> <packages...>" >&2
    return 2
  fi

  local phase_label="$1"
  local duration_mode="$2"
  local report_dir="$3"
  local manifest_phase="$4"
  local target="$5"
  shift 5

  local packages=("$@")
  local support_regex

  support_regex="$(support_go_regex "${manifest_phase}" "${target}" "${packages[@]}")"
  emit_go_raw_phase "${phase_label}" "${duration_mode}" "${report_dir}" "${support_regex}" "${packages[@]}"
}

backend_integration_core_shared_spec() {
  if [[ "$#" -ne 2 ]]; then
    echo "backend_integration_core_shared_spec requires <regex-var> <args-var>" >&2
    return 2
  fi

  local -n regex_ref="$1"
  local -n args_ref="$2"
  local package_patterns=(
    ./internal/platform/...
    ./internal/app
    ./internal/modules/incidents
    ./internal/modules/entities
    ./internal/modules/timeline
  )
  local regex_components=(
    "$(manifest_go_regex phase0 integration authoritative backend_integration ./internal/platform/... ./internal/app)"
    "$(manifest_go_regex phase2 integration authoritative backend_integration ./internal/modules/incidents)"
    "$(manifest_go_regex phase3 integration authoritative backend_integration ./internal/modules/timeline)"
    "$(manifest_go_regex phase4 integration authoritative backend_integration ./internal/modules/entities ./internal/modules/timeline)"
  )
  append_declared_support_regex_components regex_components backend_integration_support "${package_patterns[@]}"
  regex_ref="$(build_union_regex "${regex_components[@]}")"
  args_ref=(
    -p "${GO_TEST_PACKAGE_PARALLELISM}"
    "${package_patterns[@]}"
  )
}

backend_integration_auth_shared_spec() {
  if [[ "$#" -ne 2 ]]; then
    echo "backend_integration_auth_shared_spec requires <regex-var> <args-var>" >&2
    return 2
  fi

  local -n regex_ref="$1"
  local -n args_ref="$2"
  local package_patterns=(
    ./internal/modules/auth
  )
  local regex_components=(
    "$(manifest_go_regex phase1 integration authoritative backend_integration ./internal/modules/auth)"
  )
  append_declared_support_regex_components regex_components backend_integration_support "${package_patterns[@]}"
  regex_ref="$(build_union_regex "${regex_components[@]}")"
  args_ref=(
    -p "${GO_TEST_PACKAGE_PARALLELISM}"
    "${package_patterns[@]}"
  )
}

backend_process_shared_spec() {
  if [[ "$#" -ne 2 ]]; then
    echo "backend_process_shared_spec requires <regex-var> <args-var>" >&2
    return 2
  fi

  local -n regex_ref="$1"
  local -n args_ref="$2"

  regex_ref="$(build_union_regex \
    "$(manifest_go_regex phase0 e2e authoritative backend_process ./cmd/server)" \
    '^(TestPhase1_.*_ProcessSmoke)$' \
    '^(TestPhase2_ProcessSmoke_)')"
  args_ref=(
    -parallel 4
    ./cmd/server
  )
}

# Shared captures are intentionally reused across multiple targets, so both
# execution and inspection resolve from the same named spec helpers.
resolve_target_shared_report_spec() {
  if [[ "$#" -ne 4 ]]; then
    echo "resolve_target_shared_report_spec requires <target> <shared-name> <regex-var> <args-var>" >&2
    return 2
  fi

  local target="$1"
  local shared_name="$2"
  local regex_var="$3"
  local args_var="$4"

  case "${shared_name}" in
    backend-integration-core)
      case "${target}" in
        backend-integration|backend-integration-support) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_integration_core_shared_spec "${regex_var}" "${args_var}"
      ;;
    backend-integration-auth)
      case "${target}" in
        backend-integration|backend-integration-support) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_integration_auth_shared_spec "${regex_var}" "${args_var}"
      ;;
    backend-process-shared)
      case "${target}" in
        backend-process|phase0-process-e2e|phase1-process-smoke|phase2-process-smoke) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_process_shared_spec "${regex_var}" "${args_var}"
      ;;
    *)
      echo "unknown shared report ${shared_name}" >&2
      return 2
      ;;
  esac
}

capture_go_report() {
  if [[ "$#" -lt 4 ]]; then
    echo "capture_go_report requires <shared-name> <regex> -- <go test args...>" >&2
    return 2
  fi

  local shared_name="$1"
  local test_regex="$2"
  shift 2

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

  shared_dir="$(prepare_shared_artifact_dir "${shared_name}")"
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

  output_mode="$(resolve_output_mode)"
  phase_capture_start PHASE

  set +e
  if [[ "${output_mode}" != "quiet" ]]; then
    env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" \
      "${GO_BIN}" test -json -run "${test_regex}" "${test_args[@]}" \
      > >(tee "${runner_log}" | stream_go_json_output >&2) \
      2> >(tee "${stderr_log}" >&2)
    run_status=$?
  else
    env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" \
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
  touch "${complete_file}"

  printf '%s\n' "${shared_dir}"
  printf '%s\n' actual
}

assign_named_shared_report() {
  if [[ "$#" -ne 4 ]]; then
    echo "assign_named_shared_report requires <dir-var> <usage-var> <target> <shared-name>" >&2
    return 2
  fi

  local dir_var="$1"
  local usage_var="$2"
  local target="$3"
  local shared_name="$4"
  local shared_regex
  local shared_args=()

  resolve_target_shared_report_spec "${target}" "${shared_name}" shared_regex shared_args
  assign_captured_report "${dir_var}" "${usage_var}" "${shared_name}" "${shared_regex}" -- "${shared_args[@]}"
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

inspect_shared_command() {
  if [[ "$#" -ne 2 ]]; then
    echo "inspect_shared_command requires <target> <shared-name>" >&2
    return 2
  fi

  local target="$1"
  local shared_name="$2"
  local shared_regex
  local shared_args=()

  resolve_target_shared_report_spec "${target}" "${shared_name}" shared_regex shared_args
  render_go_test_command "${shared_regex}" -- "${shared_args[@]}"
}

load_phase_window() {
  local report_dir="$1"
  local mode="$2"
  local stored_duration_ms

  PHASE_COMMAND_TEXT="$(<"${report_dir}/command.txt")"
  PHASE_EXIT_STATUS="$(<"${report_dir}/exit_status.txt")"
  stored_duration_ms="$(phase_clamp_duration_ms "$(<"${report_dir}/duration_ms.txt")")"
  if [[ "${mode}" == "actual" ]]; then
    PHASE_START_TIME="$(<"${report_dir}/start_time.txt")"
    PHASE_END_TIME="$(<"${report_dir}/end_time.txt")"
    PHASE_DURATION_MS="${stored_duration_ms}"
    PHASE_WALL_DURATION_MS="${PHASE_DURATION_MS}"
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
  unset CARTULARY_MANIFEST_PHASE || true
  unset CARTULARY_MANIFEST_SECTION || true
  unset CARTULARY_MANIFEST_COVERAGE || true
  unset CARTULARY_MANIFEST_EXECUTION_DEPENDENCY || true
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
  export CARTULARY_PHASE_RUNNER_LOG="${report_dir}/runner.jsonl"
  export CARTULARY_PHASE_STDERR_LOG="${report_dir}/stderr.log"
  export CARTULARY_GO_TEST_REGEX="${test_regex}"
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
  if [[ "$#" -lt 8 ]]; then
    echo "emit_go_manifest_phase requires <label> <actual|reused|derived> <report-dir> <phase> <section> <coverage> <execution-dependency> <packages...>" >&2
    return 2
  fi

  local phase_label="$1"
  local duration_mode="$2"
  local report_dir="$3"
  local manifest_phase="$4"
  local section="$5"
  local coverage="$6"
  local execution_dependency="$7"
  shift 7

  load_phase_window "${report_dir}" "${duration_mode}"
  export CARTULARY_REPORT_SLICE=1
  export CARTULARY_PHASE_RUNNER_LOG="${report_dir}/runner.jsonl"
  export CARTULARY_PHASE_STDERR_LOG="${report_dir}/stderr.log"
  export CARTULARY_MANIFEST_PHASE="${manifest_phase}"
  export CARTULARY_MANIFEST_SECTION="${section}"
  export CARTULARY_MANIFEST_COVERAGE="${coverage}"
  export CARTULARY_MANIFEST_EXECUTION_DEPENDENCY="${execution_dependency}"
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

finish_target() {
  local status="$1"
  if [[ "${status}" -eq 0 ]]; then
    emit_target_summary pass
    return 0
  fi

  emit_target_summary fail || true
  return "${status}"
}

run_backend_unit() {
  local core_regex
  local auth_regex
  local core_dir
  local core_usage
  local auth_dir
  local auth_usage
  local config_dir
  local config_usage
  local phase1_platform_count
  local status=0
  local core_package_patterns=(
    ./internal/platform/...
    ./internal/app
    ./internal/modules/incidents
    ./internal/modules/entities
    ./internal/modules/timeline
  )
  local core_regex_components=(
    "$(manifest_go_regex phase0 unit authoritative backend_unit ./internal/platform/...)"
    "$(manifest_go_regex phase0 unit authoritative backend_unit ./internal/app)"
    "$(manifest_go_regex phase2 unit authoritative backend_unit ./internal/modules/incidents)"
    "$(manifest_go_regex phase3 unit authoritative backend_unit ./internal/modules/timeline)"
    '^(TestPhase4_.*_U_4_0[89])'
  )
  local auth_package_patterns=(
    ./internal/modules/auth
  )
  local auth_regex_components=(
    "$(manifest_go_regex phase1 unit authoritative backend_unit ./internal/modules/auth)"
  )

  phase1_platform_count="$(manifest_go_count phase1 unit authoritative backend_unit ./internal/platform/...)"
  if [[ "${phase1_platform_count}" != "0" ]]; then
    core_regex_components+=("$(manifest_go_regex phase1 unit authoritative backend_unit ./internal/platform/...)")
  fi
  append_declared_support_regex_components core_regex_components backend_unit "${core_package_patterns[@]}"
  append_declared_support_regex_components auth_regex_components backend_unit "${auth_package_patterns[@]}"
  core_regex="$(build_union_regex "${core_regex_components[@]}")"
  auth_regex="$(build_union_regex "${auth_regex_components[@]}")"

  assign_captured_report core_dir core_usage backend-unit-core "${core_regex}" -- \
    "${core_package_patterns[@]}"
  assign_captured_report auth_dir auth_usage backend-unit-auth "${auth_regex}" -- \
    "${auth_package_patterns[@]}"
  assign_captured_report config_dir config_usage backend-unit-configtest '^Test' -- ./internal/testutil/configtest

  clear_go_selection_env
  emit_go_manifest_phase "backend-unit phase0 authoritative platform" "${core_usage}" "${core_dir}" phase0 unit authoritative backend_unit ./internal/platform/... || status=$?
  if [[ "${phase1_platform_count}" != "0" ]]; then
    clear_go_selection_env
    emit_go_manifest_phase "backend-unit phase1 authoritative platform" derived "${core_dir}" phase1 unit authoritative backend_unit ./internal/platform/... || status=$?
  fi
  clear_go_selection_env
  emit_go_raw_phase "backend-unit configtest" "${config_usage}" "${config_dir}" '^Test' ./internal/testutil/configtest || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-unit phase0 authoritative app" derived "${core_dir}" phase0 unit authoritative backend_unit ./internal/app || status=$?
  clear_go_selection_env
  emit_declared_support_phase "backend-unit support phase0" derived "${core_dir}" phase0 backend_unit ./internal/platform/... ./internal/app || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-unit phase1 authoritative auth" "${auth_usage}" "${auth_dir}" phase1 unit authoritative backend_unit ./internal/modules/auth || status=$?
  clear_go_selection_env
  emit_declared_support_phase "backend-unit support phase1" derived "${auth_dir}" phase1 backend_unit ./internal/modules/auth || status=$?
  clear_go_selection_env
  emit_declared_support_phase "backend-unit support phase3" derived "${core_dir}" phase3 backend_unit ./internal/modules/timeline || status=$?
  clear_go_selection_env
  emit_go_raw_phase "backend-unit phase4 app" derived "${core_dir}" '^(TestPhase4_.*_U_4_0[89])' ./internal/app ./internal/modules/incidents ./internal/modules/entities ./internal/modules/timeline || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-unit phase2 authoritative" derived "${core_dir}" phase2 unit authoritative backend_unit ./internal/modules/incidents || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-unit phase3 authoritative" derived "${core_dir}" phase3 unit authoritative backend_unit ./internal/modules/timeline || status=$?

  finish_target "${status}"
}

run_backend_store() {
  local shared_regex
  local shared_dir
  local shared_usage
  local status=0

  shared_regex="$(build_union_regex \
    '^(TestPhase4_.*_U_4_0[1-7])' \
    "$(manifest_go_regex phase1 unit authoritative backend_store ./internal/modules/auth)" \
    "$(manifest_go_regex phase2 unit authoritative backend_store ./internal/modules/incidents)" \
    "$(manifest_go_regex phase3 unit authoritative backend_store ./internal/modules/timeline)")"

  assign_captured_report shared_dir shared_usage backend-store-shared "${shared_regex}" -- \
    -p "${GO_TEST_PACKAGE_PARALLELISM}" \
    ./internal/modules/auth \
    ./internal/modules/incidents \
    ./internal/modules/entities \
    ./internal/modules/timeline

  clear_go_selection_env
  emit_go_raw_phase "backend-store" "${shared_usage}" "${shared_dir}" '^(TestPhase4_.*_U_4_0[1-7])' ./internal/modules/entities ./internal/modules/timeline || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-store phase1 authoritative" derived "${shared_dir}" phase1 unit authoritative backend_store ./internal/modules/auth || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-store phase2 authoritative" derived "${shared_dir}" phase2 unit authoritative backend_store ./internal/modules/incidents || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-store phase3 authoritative" derived "${shared_dir}" phase3 unit authoritative backend_store ./internal/modules/timeline || status=$?

  finish_target "${status}"
}

run_backend_integration() {
  local testutil_dir
  local testutil_usage
  local core_dir
  local core_usage
  local auth_dir
  local auth_usage
  local status=0

  assign_captured_report testutil_dir testutil_usage backend-integration-testutil '^Test' -- \
    -p "${GO_TEST_PACKAGE_PARALLELISM}" \
    ./internal/testutil/httptestx \
    ./internal/testutil/pgtest \
    ./internal/testutil/s3test \
    ./internal/testutil/testcontainersx \
    ./internal/testutil/wstest

  assign_named_shared_report core_dir core_usage backend-integration backend-integration-core
  assign_named_shared_report auth_dir auth_usage backend-integration backend-integration-auth

  clear_go_selection_env
  emit_go_raw_phase "backend-integration testutil" "${testutil_usage}" "${testutil_dir}" '^Test' ./internal/testutil/httptestx ./internal/testutil/pgtest ./internal/testutil/s3test ./internal/testutil/testcontainersx ./internal/testutil/wstest || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-integration phase0 authoritative" "${core_usage}" "${core_dir}" phase0 integration authoritative backend_integration ./internal/platform/... ./internal/app || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-integration phase1 authoritative" "${auth_usage}" "${auth_dir}" phase1 integration authoritative backend_integration ./internal/modules/auth || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-integration phase4 authoritative" derived "${core_dir}" phase4 integration authoritative backend_integration ./internal/modules/entities ./internal/modules/timeline || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-integration phase2 authoritative" derived "${core_dir}" phase2 integration authoritative backend_integration ./internal/modules/incidents || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-integration phase3 authoritative" derived "${core_dir}" phase3 integration authoritative backend_integration ./internal/modules/timeline || status=$?

  finish_target "${status}"
}

run_backend_integration_support() {
  local core_dir
  local core_usage
  local auth_dir
  local auth_usage
  local status=0

  assign_named_shared_report core_dir core_usage backend-integration-support backend-integration-core
  assign_named_shared_report auth_dir auth_usage backend-integration-support backend-integration-auth

  clear_go_selection_env
  emit_declared_support_phase "backend-integration support phase0" "${core_usage}" "${core_dir}" phase0 backend_integration_support ./internal/platform/... ./internal/app || status=$?
  clear_go_selection_env
  emit_declared_support_phase "backend-integration support phase1" "${auth_usage}" "${auth_dir}" phase1 backend_integration_support ./internal/modules/auth || status=$?
  clear_go_selection_env
  emit_declared_support_phase "backend-integration support phase2" "${core_usage}" "${core_dir}" phase2 backend_integration_support ./internal/modules/incidents || status=$?
  clear_go_selection_env
  emit_declared_support_phase "backend-integration support phase3" derived "${core_dir}" phase3 backend_integration_support ./internal/modules/timeline || status=$?
  clear_go_selection_env
  emit_declared_support_phase "backend-integration support phase4" derived "${core_dir}" phase4 backend_integration_support ./internal/modules/entities ./internal/modules/timeline || status=$?

  finish_target "${status}"
}

run_backend_process() {
  local shared_dir
  local shared_usage
  local status=0

  assign_named_shared_report shared_dir shared_usage backend-process backend-process-shared

  clear_go_selection_env
  emit_go_manifest_phase "backend-process phase0 authoritative" "${shared_usage}" "${shared_dir}" phase0 e2e authoritative backend_process ./cmd/server || status=$?
  clear_go_selection_env
  emit_go_raw_phase "backend-process phase1 smoke" derived "${shared_dir}" '^(TestPhase1_.*_ProcessSmoke)$' ./cmd/server || status=$?

  finish_target "${status}"
}

run_phase0_process_e2e() {
  local shared_dir
  local shared_usage
  local status=0

  assign_named_shared_report shared_dir shared_usage phase0-process-e2e backend-process-shared

  clear_go_selection_env
  emit_go_manifest_phase "phase0-process-e2e" "${shared_usage}" "${shared_dir}" phase0 e2e authoritative backend_process ./cmd/server || status=$?

  finish_target "${status}"
}

run_phase1_process_smoke() {
  local shared_dir
  local shared_usage
  local status=0

  assign_named_shared_report shared_dir shared_usage phase1-process-smoke backend-process-shared

  clear_go_selection_env
  emit_go_raw_phase "phase1-process-smoke" "${shared_usage}" "${shared_dir}" '^(TestPhase1_.*_ProcessSmoke)$' ./cmd/server || status=$?

  finish_target "${status}"
}

run_phase2_process_smoke() {
  local shared_dir
  local shared_usage
  local status=0

  assign_named_shared_report shared_dir shared_usage phase2-process-smoke backend-process-shared

  clear_go_selection_env
  emit_go_raw_phase "phase2-process-smoke" "${shared_usage}" "${shared_dir}" '^(TestPhase2_ProcessSmoke_)' ./cmd/server || status=$?

  finish_target "${status}"
}

main() {
  if [[ "$#" -eq 0 ]]; then
    usage
  fi

  case "$1" in
    inspect-shared-command)
      if [[ "$#" -ne 3 ]]; then
        usage
      fi
      inspect_shared_command "$2" "$3"
      ;;
    backend-unit)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      run_backend_unit
      ;;
    backend-store)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      run_backend_store
      ;;
    backend-integration)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      run_backend_integration
      ;;
    backend-integration-support)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      run_backend_integration_support
      ;;
    backend-process)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      run_backend_process
      ;;
    phase0-process-e2e)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      run_phase0_process_e2e
      ;;
    phase1-process-smoke)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      run_phase1_process_smoke
      ;;
    phase2-process-smoke)
      if [[ "$#" -ne 1 ]]; then
        usage
      fi
      run_phase2_process_smoke
      ;;
    *)
      usage
      ;;
  esac
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
  main "$@"
fi
