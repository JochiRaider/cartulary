#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
source "${ROOT_DIR}/scripts/lib/run-phase-common.sh"

GO_BIN="${GO:-go}"
GO_CACHE_DIR="${GO_CACHE_DIR:-/tmp/cartulary-go-build}"
GO_MOD_CACHE_DIR="${GO_MOD_CACHE_DIR:-/tmp/cartulary-go-mod}"
GO_TEST_SERVICE_PACKAGE_PARALLELISM="${GO_TEST_SERVICE_PACKAGE_PARALLELISM:-1}"
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
  exit 2
}

manifest_go_regex() {
  "${NODE_HELPER}" "${MANIFEST_SCRIPT}" go-regex "$@"
}

manifest_go_count() {
  "${NODE_HELPER}" "${MANIFEST_SCRIPT}" go-count "$@"
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
  local start_ms
  local end_ms
  local duration_ms
  local output_mode
  local run_status
  local complete_file
  local existing_command

  shared_dir="$(prepare_shared_artifact_dir "${shared_name}")"
  complete_file="${shared_dir}/complete"
  runner_log="${shared_dir}/runner.jsonl"
  stderr_log="${shared_dir}/stderr.log"
  command_text="$(render_command env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" "${GO_BIN}" test -json -run "${test_regex}" "${test_args[@]}")"

  if [[ -f "${complete_file}" ]]; then
    existing_command="$(<"${shared_dir}/command.txt")"
    if [[ "${existing_command}" != "${command_text}" ]]; then
      echo "shared go report ${shared_name} was created with a different command" >&2
      echo "existing: ${existing_command}" >&2
      echo "current:  ${command_text}" >&2
      return 1
    fi
    printf '%s\n' "${shared_dir}"
    return 0
  fi

  output_mode="$(resolve_output_mode)"
  start_time="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  start_ms="$(date +%s%3N)"

  set +e
  if [[ "${output_mode}" != "quiet" ]]; then
    env GOCACHE="${GO_CACHE_DIR}" GOMODCACHE="${GO_MOD_CACHE_DIR}" \
      "${GO_BIN}" test -json -run "${test_regex}" "${test_args[@]}" \
      > >(tee "${runner_log}") \
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

  end_time="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  end_ms="$(date +%s%3N)"
  duration_ms="$((end_ms - start_ms))"

  printf '%s\n' "${command_text}" >"${shared_dir}/command.txt"
  printf '%s\n' "${start_time}" >"${shared_dir}/start_time.txt"
  printf '%s\n' "${end_time}" >"${shared_dir}/end_time.txt"
  printf '%s\n' "${duration_ms}" >"${shared_dir}/duration_ms.txt"
  printf '%s\n' "${run_status}" >"${shared_dir}/exit_status.txt"
  touch "${complete_file}"

  printf '%s\n' "${shared_dir}"
}

load_phase_window() {
  local report_dir="$1"
  local mode="$2"

  PHASE_COMMAND_TEXT="$(<"${report_dir}/command.txt")"
  PHASE_EXIT_STATUS="$(<"${report_dir}/exit_status.txt")"
  if [[ "${mode}" == "actual" ]]; then
    PHASE_START_TIME="$(<"${report_dir}/start_time.txt")"
    PHASE_END_TIME="$(<"${report_dir}/end_time.txt")"
    PHASE_DURATION_MS="$(<"${report_dir}/duration_ms.txt")"
    return 0
  fi

  PHASE_END_TIME="$(<"${report_dir}/end_time.txt")"
  PHASE_START_TIME="${PHASE_END_TIME}"
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
    echo "emit_go_raw_phase requires <label> <actual|derived> <report-dir> <regex> <packages...>" >&2
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
    "${PHASE_EXIT_STATUS}"
}

emit_go_manifest_phase() {
  if [[ "$#" -lt 8 ]]; then
    echo "emit_go_manifest_phase requires <label> <actual|derived> <report-dir> <phase> <section> <coverage> <execution-dependency> <packages...>" >&2
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
  local auth_dir
  local config_dir
  local phase1_platform_count
  local status=0

  phase1_platform_count="$(manifest_go_count phase1 unit authoritative backend_unit ./internal/platform/...)"
  core_regex="$(build_union_regex \
    "$(manifest_go_regex phase0 unit authoritative backend_unit ./internal/platform/...)" \
    "$(manifest_go_regex phase0 unit authoritative backend_unit ./internal/app)" \
    "$(manifest_go_regex phase2 unit authoritative backend_unit ./internal/modules/incidents)" \
    "$(manifest_go_regex phase3 unit authoritative backend_unit ./internal/modules/timeline)" \
    '^(TestPhase4_.*_U_4_0[89])')"
  if [[ "${phase1_platform_count}" != "0" ]]; then
    core_regex="$(build_union_regex "${core_regex}" "$(manifest_go_regex phase1 unit authoritative backend_unit ./internal/platform/...)" )"
  fi
  auth_regex="$(build_union_regex \
    "$(manifest_go_regex phase1 unit authoritative backend_unit ./internal/modules/auth)" \
    '^(TestSupportPhase1_)')"

  core_dir="$(capture_go_report backend-unit-core "${core_regex}" -- \
    ./internal/platform/... \
    ./internal/app \
    ./internal/modules/incidents \
    ./internal/modules/entities \
    ./internal/modules/timeline)"
  auth_dir="$(capture_go_report backend-unit-auth "${auth_regex}" -- ./internal/modules/auth)"
  config_dir="$(capture_go_report backend-unit-configtest '^Test' -- ./internal/testutil/configtest)"

  clear_go_selection_env
  emit_go_manifest_phase "backend-unit phase0 authoritative platform" actual "${core_dir}" phase0 unit authoritative backend_unit ./internal/platform/... || status=$?
  if [[ "${phase1_platform_count}" != "0" ]]; then
    clear_go_selection_env
    emit_go_manifest_phase "backend-unit phase1 authoritative platform" derived "${core_dir}" phase1 unit authoritative backend_unit ./internal/platform/... || status=$?
  fi
  clear_go_selection_env
  emit_go_raw_phase "backend-unit configtest" actual "${config_dir}" '^Test' ./internal/testutil/configtest || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-unit phase0 authoritative app" derived "${core_dir}" phase0 unit authoritative backend_unit ./internal/app || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-unit phase1 authoritative auth" actual "${auth_dir}" phase1 unit authoritative backend_unit ./internal/modules/auth || status=$?
  clear_go_selection_env
  emit_go_raw_phase "backend-unit support phase1" derived "${auth_dir}" '^(TestSupportPhase1_)' ./internal/modules/auth || status=$?
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
  local status=0

  shared_regex="$(build_union_regex \
    '^(TestPhase4_.*_U_4_0[1-7])' \
    "$(manifest_go_regex phase2 unit authoritative backend_store ./internal/modules/incidents)" \
    "$(manifest_go_regex phase3 unit authoritative backend_store ./internal/modules/timeline)")"

  shared_dir="$(capture_go_report backend-store-shared "${shared_regex}" -- \
    -p "${GO_TEST_SERVICE_PACKAGE_PARALLELISM}" \
    ./internal/modules/incidents \
    ./internal/modules/entities \
    ./internal/modules/timeline)"

  clear_go_selection_env
  emit_go_raw_phase "backend-store" actual "${shared_dir}" '^(TestPhase4_.*_U_4_0[1-7])' ./internal/modules/entities ./internal/modules/timeline || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-store phase2 authoritative" derived "${shared_dir}" phase2 unit authoritative backend_store ./internal/modules/incidents || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-store phase3 authoritative" derived "${shared_dir}" phase3 unit authoritative backend_store ./internal/modules/timeline || status=$?

  finish_target "${status}"
}

run_backend_integration() {
  local testutil_dir
  local core_dir
  local auth_dir
  local core_regex
  local auth_regex
  local status=0

  testutil_dir="$(capture_go_report backend-integration-testutil '^Test' -- \
    -p "${GO_TEST_SERVICE_PACKAGE_PARALLELISM}" \
    ./internal/testutil/httptestx \
    ./internal/testutil/pgtest \
    ./internal/testutil/s3test \
    ./internal/testutil/testcontainersx \
    ./internal/testutil/wstest)"

  core_regex="$(build_union_regex \
    "$(manifest_go_regex phase0 integration authoritative backend_integration ./internal/platform/... ./internal/app)" \
    "$(manifest_go_regex phase2 integration authoritative backend_integration ./internal/modules/incidents)" \
    "$(manifest_go_regex phase3 integration authoritative backend_integration ./internal/modules/timeline)" \
    '^(TestPhase4_.*_I_4_)' \
    '^(TestSupportPhase2_)' \
    '^(TestSupportPhase3_)')"
  auth_regex="$(build_union_regex \
    "$(manifest_go_regex phase1 integration authoritative backend_integration ./internal/modules/auth)" \
    '^(TestSupportPhase1_)')"

  core_dir="$(capture_go_report backend-integration-core "${core_regex}" -- \
    -p "${GO_TEST_SERVICE_PACKAGE_PARALLELISM}" \
    ./internal/platform/... \
    ./internal/app \
    ./internal/modules/incidents \
    ./internal/modules/entities \
    ./internal/modules/timeline)"
  auth_dir="$(capture_go_report backend-integration-auth "${auth_regex}" -- \
    -p "${GO_TEST_SERVICE_PACKAGE_PARALLELISM}" \
    ./internal/modules/auth)"

  clear_go_selection_env
  emit_go_raw_phase "backend-integration testutil" actual "${testutil_dir}" '^Test' ./internal/testutil/httptestx ./internal/testutil/pgtest ./internal/testutil/s3test ./internal/testutil/testcontainersx ./internal/testutil/wstest || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-integration phase0 authoritative" actual "${core_dir}" phase0 integration authoritative backend_integration ./internal/platform/... ./internal/app || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-integration phase1 authoritative" actual "${auth_dir}" phase1 integration authoritative backend_integration ./internal/modules/auth || status=$?
  clear_go_selection_env
  emit_go_raw_phase "backend-integration phase4" derived "${core_dir}" '^(TestPhase4_.*_I_4_)' ./internal/platform/... ./internal/app ./internal/modules/entities ./internal/modules/timeline || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-integration phase2 authoritative" derived "${core_dir}" phase2 integration authoritative backend_integration ./internal/modules/incidents || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-integration phase3 authoritative" derived "${core_dir}" phase3 integration authoritative backend_integration ./internal/modules/timeline || status=$?

  finish_target "${status}"
}

run_backend_integration_support() {
  local core_dir
  local auth_dir
  local core_regex
  local auth_regex
  local status=0

  core_regex="$(build_union_regex \
    "$(manifest_go_regex phase0 integration authoritative backend_integration ./internal/platform/... ./internal/app)" \
    "$(manifest_go_regex phase2 integration authoritative backend_integration ./internal/modules/incidents)" \
    "$(manifest_go_regex phase3 integration authoritative backend_integration ./internal/modules/timeline)" \
    '^(TestPhase4_.*_I_4_)' \
    '^(TestSupportPhase2_)' \
    '^(TestSupportPhase3_)')"
  auth_regex="$(build_union_regex \
    "$(manifest_go_regex phase1 integration authoritative backend_integration ./internal/modules/auth)" \
    '^(TestSupportPhase1_)')"

  core_dir="$(capture_go_report backend-integration-core "${core_regex}" -- \
    -p "${GO_TEST_SERVICE_PACKAGE_PARALLELISM}" \
    ./internal/platform/... \
    ./internal/app \
    ./internal/modules/incidents \
    ./internal/modules/entities \
    ./internal/modules/timeline)"
  auth_dir="$(capture_go_report backend-integration-auth "${auth_regex}" -- \
    -p "${GO_TEST_SERVICE_PACKAGE_PARALLELISM}" \
    ./internal/modules/auth)"

  clear_go_selection_env
  emit_go_raw_phase "backend-integration support phase1" actual "${auth_dir}" '^(TestSupportPhase1_)' ./internal/modules/auth || status=$?
  clear_go_selection_env
  emit_go_raw_phase "backend-integration support phase2" actual "${core_dir}" '^(TestSupportPhase2_)' ./internal/modules/incidents || status=$?
  clear_go_selection_env
  emit_go_raw_phase "backend-integration support phase3" derived "${core_dir}" '^(TestSupportPhase3_)' ./internal/modules/timeline || status=$?

  finish_target "${status}"
}

run_backend_process() {
  local shared_regex
  local shared_dir
  local status=0

  shared_regex="$(build_union_regex \
    "$(manifest_go_regex phase0 e2e authoritative backend_process ./cmd/server)" \
    '^(TestPhase1_.*_ProcessSmoke)$' \
    '^(TestPhase2_ProcessSmoke_)')"
  shared_dir="$(capture_go_report backend-process-shared "${shared_regex}" -- -parallel 4 ./cmd/server)"

  clear_go_selection_env
  emit_go_manifest_phase "backend-process phase0 authoritative" actual "${shared_dir}" phase0 e2e authoritative backend_process ./cmd/server || status=$?
  clear_go_selection_env
  emit_go_raw_phase "backend-process phase1 smoke" derived "${shared_dir}" '^(TestPhase1_.*_ProcessSmoke)$' ./cmd/server || status=$?

  finish_target "${status}"
}

run_phase0_process_e2e() {
  local shared_regex
  local shared_dir
  local status=0

  shared_regex="$(build_union_regex \
    "$(manifest_go_regex phase0 e2e authoritative backend_process ./cmd/server)" \
    '^(TestPhase1_.*_ProcessSmoke)$' \
    '^(TestPhase2_ProcessSmoke_)')"
  shared_dir="$(capture_go_report backend-process-shared "${shared_regex}" -- -parallel 4 ./cmd/server)"

  clear_go_selection_env
  emit_go_manifest_phase "phase0-process-e2e" actual "${shared_dir}" phase0 e2e authoritative backend_process ./cmd/server || status=$?

  finish_target "${status}"
}

run_phase1_process_smoke() {
  local shared_regex
  local shared_dir
  local status=0

  shared_regex="$(build_union_regex \
    "$(manifest_go_regex phase0 e2e authoritative backend_process ./cmd/server)" \
    '^(TestPhase1_.*_ProcessSmoke)$' \
    '^(TestPhase2_ProcessSmoke_)')"
  shared_dir="$(capture_go_report backend-process-shared "${shared_regex}" -- -parallel 4 ./cmd/server)"

  clear_go_selection_env
  emit_go_raw_phase "phase1-process-smoke" actual "${shared_dir}" '^(TestPhase1_.*_ProcessSmoke)$' ./cmd/server || status=$?

  finish_target "${status}"
}

run_phase2_process_smoke() {
  local shared_regex
  local shared_dir
  local status=0

  shared_regex="$(build_union_regex \
    "$(manifest_go_regex phase0 e2e authoritative backend_process ./cmd/server)" \
    '^(TestPhase1_.*_ProcessSmoke)$' \
    '^(TestPhase2_ProcessSmoke_)')"
  shared_dir="$(capture_go_report backend-process-shared "${shared_regex}" -- -parallel 4 ./cmd/server)"

  clear_go_selection_env
  emit_go_raw_phase "phase2-process-smoke" actual "${shared_dir}" '^(TestPhase2_ProcessSmoke_)' ./cmd/server || status=$?

  finish_target "${status}"
}

if [[ "$#" -ne 1 ]]; then
  usage
fi

case "$1" in
  backend-unit)
    run_backend_unit
    ;;
  backend-store)
    run_backend_store
    ;;
  backend-integration)
    run_backend_integration
    ;;
  backend-integration-support)
    run_backend_integration_support
    ;;
  backend-process)
    run_backend_process
    ;;
  phase0-process-e2e)
    run_phase0_process_e2e
    ;;
  phase1-process-smoke)
    run_phase1_process_smoke
    ;;
  phase2-process-smoke)
    run_phase2_process_smoke
    ;;
  *)
    usage
    ;;
esac
