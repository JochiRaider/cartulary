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
MANIFEST_SCRIPT="${ROOT_DIR}/scripts/lib/phase-manifest.mjs"
SHARD_PLAN_SCRIPT="${ROOT_DIR}/scripts/lib/go-shard-plan.mjs"

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

manifest_phases() {
  "${NODE_HELPER}" "${MANIFEST_SCRIPT}" list-phases
}

planned_shard_names() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" list-shards "$@"
}

planned_shard_spec() {
  "${NODE_HELPER}" "${SHARD_PLAN_SCRIPT}" shard-spec "$@"
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
  while IFS= read -r phase; do
    if [[ -z "${phase}" ]]; then
      continue
    fi
    count="$(support_go_count "${phase}" "${target}" "$@")"
    if [[ "${count}" == "0" ]]; then
      continue
    fi
    components_ref+=("$(support_go_regex "${phase}" "${target}" "$@")")
  done < <(manifest_phases)
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

backend_integration_phase_shared_spec() {
  if [[ "$#" -lt 4 ]]; then
    echo "backend_integration_phase_shared_spec requires <phase> <regex-var> <args-var> <packages...>" >&2
    return 2
  fi

  local manifest_phase="$1"
  local regex_var="$2"
  local args_var="$3"
  shift 3

  local -n regex_ref="${regex_var}"
  local -n args_ref="${args_var}"
  local package_patterns=("$@")
  local regex_components=()
  local authoritative_count

  authoritative_count="$(manifest_go_count "${manifest_phase}" integration authoritative backend_integration "${package_patterns[@]}")"
  if [[ "${authoritative_count}" != "0" ]]; then
    regex_components+=("$(manifest_go_regex "${manifest_phase}" integration authoritative backend_integration "${package_patterns[@]}")")
  fi
  append_declared_support_regex_components regex_components backend_integration_support "${package_patterns[@]}"
  regex_ref="$(build_union_regex "${regex_components[@]}")"
  args_ref=(
    -p "${GO_TEST_PACKAGE_PARALLELISM}"
    "${package_patterns[@]}"
  )
}

backend_integration_phase0_platform_shared_spec() {
  backend_integration_phase_shared_spec phase0 "$1" "$2" ./internal/platform/...
}

backend_integration_phase0_app_shared_spec() {
  backend_integration_phase_shared_spec phase0 "$1" "$2" ./internal/app
}

backend_integration_phase2_incidents_shared_spec() {
  backend_integration_phase_shared_spec phase2 "$1" "$2" ./internal/modules/incidents
}

backend_integration_phase3_timeline_shared_spec() {
  backend_integration_phase_shared_spec phase3 "$1" "$2" ./internal/modules/timeline
}

backend_integration_phase4_entities_shared_spec() {
  backend_integration_phase_shared_spec phase4 "$1" "$2" ./internal/modules/entities
}

backend_integration_phase4_timeline_shared_spec() {
  backend_integration_phase_shared_spec phase4 "$1" "$2" ./internal/modules/timeline
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

backend_integration_testutil_shared_spec() {
  if [[ "$#" -ne 2 ]]; then
    echo "backend_integration_testutil_shared_spec requires <regex-var> <args-var>" >&2
    return 2
  fi

  local -n regex_ref="$1"
  local -n args_ref="$2"

  regex_ref='^Test'
  args_ref=(
    -p "${GO_TEST_PACKAGE_PARALLELISM}"
    ./internal/testutil/httptestx
    ./internal/testutil/pgtest
    ./internal/testutil/s3test
    ./internal/testutil/testcontainersx
    ./internal/testutil/wstest
  )
}

backend_unit_core_shared_spec() {
  if [[ "$#" -ne 2 ]]; then
    echo "backend_unit_core_shared_spec requires <regex-var> <args-var>" >&2
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
    "$(manifest_go_regex phase0 unit authoritative backend_unit ./internal/platform/...)"
    "$(manifest_go_regex phase0 unit authoritative backend_unit ./internal/app)"
    "$(manifest_go_regex phase2 unit authoritative backend_unit ./internal/modules/incidents)"
    "$(manifest_go_regex phase3 unit authoritative backend_unit ./internal/modules/timeline)"
    "$(manifest_go_regex phase4 unit authoritative backend_unit ./internal/app ./internal/modules/incidents ./internal/modules/entities ./internal/modules/timeline)"
  )
  local phase1_platform_count

  phase1_platform_count="$(manifest_go_count phase1 unit authoritative backend_unit ./internal/platform/...)"
  if [[ "${phase1_platform_count}" != "0" ]]; then
    regex_components+=("$(manifest_go_regex phase1 unit authoritative backend_unit ./internal/platform/...)")
  fi
  append_declared_support_regex_components regex_components backend_unit "${package_patterns[@]}"

  regex_ref="$(build_union_regex "${regex_components[@]}")"
  args_ref=(
    "${package_patterns[@]}"
  )
}

backend_unit_auth_shared_spec() {
  if [[ "$#" -ne 2 ]]; then
    echo "backend_unit_auth_shared_spec requires <regex-var> <args-var>" >&2
    return 2
  fi

  local -n regex_ref="$1"
  local -n args_ref="$2"
  local package_patterns=(
    ./internal/modules/auth
  )
  local regex_components=(
    "$(manifest_go_regex phase1 unit authoritative backend_unit ./internal/modules/auth)"
  )

  append_declared_support_regex_components regex_components backend_unit "${package_patterns[@]}"

  regex_ref="$(build_union_regex "${regex_components[@]}")"
  args_ref=(
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
  local planned_spec=()

  if [[ "${target}" == "backend-integration" || "${target}" == "backend-integration-support" ]] && [[ "${shared_name}" == *"-shard-"* ]] && mapfile -t planned_spec < <(planned_shard_spec "${target}" "${shared_name}" 2>/dev/null); then
    if [[ "${#planned_spec[@]}" -lt 2 ]]; then
      echo "planned shard ${shared_name} for ${target} returned an incomplete spec" >&2
      return 2
    fi
    local -n regex_ref="${regex_var}"
    local -n args_ref="${args_var}"
    regex_ref="${planned_spec[0]}"
    args_ref=(
      -p "${GO_TEST_PACKAGE_PARALLELISM}"
      "${planned_spec[@]:1}"
    )
    return 0
  fi

  case "${shared_name}" in
    backend-unit-core)
      case "${target}" in
        backend-unit) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_unit_core_shared_spec "${regex_var}" "${args_var}"
      ;;
    backend-unit-auth)
      case "${target}" in
        backend-unit) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_unit_auth_shared_spec "${regex_var}" "${args_var}"
      ;;
    backend-integration-phase0-platform)
      case "${target}" in
        backend-integration|backend-integration-support) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_integration_phase0_platform_shared_spec "${regex_var}" "${args_var}"
      ;;
    backend-integration-phase0-app)
      case "${target}" in
        backend-integration|backend-integration-support) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_integration_phase0_app_shared_spec "${regex_var}" "${args_var}"
      ;;
    backend-integration-phase2-incidents)
      case "${target}" in
        backend-integration|backend-integration-support) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_integration_phase2_incidents_shared_spec "${regex_var}" "${args_var}"
      ;;
    backend-integration-phase3-timeline)
      case "${target}" in
        backend-integration|backend-integration-support) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_integration_phase3_timeline_shared_spec "${regex_var}" "${args_var}"
      ;;
    backend-integration-phase4-entities)
      case "${target}" in
        backend-integration|backend-integration-support) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_integration_phase4_entities_shared_spec "${regex_var}" "${args_var}"
      ;;
    backend-integration-phase4-timeline)
      case "${target}" in
        backend-integration|backend-integration-support) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_integration_phase4_timeline_shared_spec "${regex_var}" "${args_var}"
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
    backend-integration-testutil)
      case "${target}" in
        backend-integration) ;;
        *)
          echo "shared report ${shared_name} is not defined for target ${target}" >&2
          return 2
          ;;
      esac
      backend_integration_testutil_shared_spec "${regex_var}" "${args_var}"
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
      assign_named_shared_report report_dir report_usage "${target}" "${shared_name}"
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

  local aggregate_root="${metadata_dir}/aggregate-reports"
  local output_dir="${aggregate_root}/${aggregate_name}"
  local runner_log="${output_dir}/runner.jsonl"
  local stderr_log="${output_dir}/stderr.log"
  local command_file="${output_dir}/command.txt"
  local start_time=""
  local end_time=""
  local duration_ms=0
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
    fi
  done

  usage_ref="reused"
  if [[ "${has_actual}" -eq 1 ]]; then
    usage_ref="actual"
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
  export CARTULARY_PHASE_ACCOUNTING_MODE="${duration_mode}"
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
  export CARTULARY_PHASE_ACCOUNTING_MODE="${duration_mode}"
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
  local core_args=()
  local auth_regex
  local auth_args=()
  local core_dir
  local core_usage
  local auth_dir
  local auth_usage
  local config_dir
  local config_usage
  local status=0

  backend_unit_core_shared_spec core_regex core_args
  backend_unit_auth_shared_spec auth_regex auth_args

  assign_captured_report core_dir core_usage backend-unit-core "${core_regex}" -- \
    "${core_args[@]}"
  assign_captured_report auth_dir auth_usage backend-unit-auth "${auth_regex}" -- \
    "${auth_args[@]}"
  assign_captured_report config_dir config_usage backend-unit-configtest '^Test' -- ./internal/testutil/configtest

  clear_go_selection_env
  emit_go_manifest_phase "backend-unit phase0 authoritative platform" "${core_usage}" "${core_dir}" phase0 unit authoritative backend_unit ./internal/platform/... || status=$?
  if [[ "$(manifest_go_count phase1 unit authoritative backend_unit ./internal/platform/...)" != "0" ]]; then
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
  emit_declared_support_phase "backend-unit support phase2" derived "${core_dir}" phase2 backend_unit ./internal/modules/incidents || status=$?
  clear_go_selection_env
  emit_declared_support_phase "backend-unit support phase3" derived "${core_dir}" phase3 backend_unit ./internal/modules/timeline || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-unit phase4 authoritative" derived "${core_dir}" phase4 unit authoritative backend_unit ./internal/app ./internal/modules/incidents ./internal/modules/entities ./internal/modules/timeline || status=$?
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
    "$(manifest_go_regex phase4 unit authoritative backend_store ./internal/modules/entities ./internal/modules/timeline)" \
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
  emit_go_manifest_phase "backend-store phase4 authoritative" "${shared_usage}" "${shared_dir}" phase4 unit authoritative backend_store ./internal/modules/entities ./internal/modules/timeline || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-store phase1 authoritative" derived "${shared_dir}" phase1 unit authoritative backend_store ./internal/modules/auth || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-store phase2 authoritative" derived "${shared_dir}" phase2 unit authoritative backend_store ./internal/modules/incidents || status=$?
  clear_go_selection_env
  emit_go_manifest_phase "backend-store phase3 authoritative" derived "${shared_dir}" phase3 unit authoritative backend_store ./internal/modules/timeline || status=$?

  finish_target "${status}"
}

run_backend_integration() {
  local metadata_dir
  local shard_names=()
  local aggregate_names=()
  local status=0
  local aggregate_name
  local aggregate_dir
  local aggregate_usage
  local aggregate_mode
  local aggregate_label
  local aggregate_phase
  local aggregate_section
  local aggregate_coverage
  local aggregate_dependency
  local aggregate_regex
  local aggregate_shards=()
  local aggregate_packages=()

  metadata_dir="$(mktemp -d "${TMPDIR:-/tmp}/cartulary-backend-integration-shards.XXXXXX")"
  mapfile -t shard_names < <(planned_shard_names backend-integration)
  capture_named_shared_reports_parallel backend-integration "${BACKEND_INTEGRATION_SHARD_JOBS}" "${metadata_dir}" "${shard_names[@]}" || status=$?

  mapfile -t aggregate_names < <(planned_aggregate_names backend-integration)
  for aggregate_name in "${aggregate_names[@]}"; do
    mapfile -t aggregate_shards < <(planned_aggregate_shards backend-integration "${aggregate_name}")
    mapfile -t aggregate_packages < <(planned_aggregate_packages backend-integration "${aggregate_name}")
    create_aggregate_report aggregate_dir aggregate_usage "${metadata_dir}" "${aggregate_name}" backend-integration "${aggregate_shards[@]}" || status=$?
    aggregate_mode="$(planned_aggregate_field backend-integration "${aggregate_name}" mode)"
    aggregate_label="$(planned_aggregate_field backend-integration "${aggregate_name}" label)"
    clear_go_selection_env
    if [[ "${aggregate_mode}" == "raw" ]]; then
      aggregate_regex="$(planned_aggregate_field backend-integration "${aggregate_name}" raw_selector)"
      emit_go_raw_phase "${aggregate_label}" "${aggregate_usage}" "${aggregate_dir}" "${aggregate_regex}" "${aggregate_packages[@]}" || status=$?
      continue
    fi
    aggregate_phase="$(planned_aggregate_field backend-integration "${aggregate_name}" phase)"
    aggregate_section="$(planned_aggregate_field backend-integration "${aggregate_name}" section)"
    aggregate_coverage="$(planned_aggregate_field backend-integration "${aggregate_name}" coverage)"
    aggregate_dependency="$(planned_aggregate_field backend-integration "${aggregate_name}" execution_dependency)"
    emit_go_manifest_phase "${aggregate_label}" "${aggregate_usage}" "${aggregate_dir}" "${aggregate_phase}" "${aggregate_section}" "${aggregate_coverage}" "${aggregate_dependency}" "${aggregate_packages[@]}" || status=$?
  done
  rm -rf -- "${metadata_dir}"

  finish_target "${status}"
}

run_backend_integration_support() {
  local metadata_dir
  local shard_names=()
  local aggregate_names=()
  local status=0
  local aggregate_name
  local aggregate_dir
  local aggregate_usage
  local aggregate_label
  local aggregate_phase
  local aggregate_shards=()
  local aggregate_packages=()

  metadata_dir="$(mktemp -d "${TMPDIR:-/tmp}/cartulary-backend-integration-support-shards.XXXXXX")"
  mapfile -t shard_names < <(planned_shard_names backend-integration-support)
  capture_named_shared_reports_parallel backend-integration-support "${BACKEND_INTEGRATION_SHARD_JOBS}" "${metadata_dir}" "${shard_names[@]}" || status=$?

  mapfile -t aggregate_names < <(planned_aggregate_names backend-integration-support)
  for aggregate_name in "${aggregate_names[@]}"; do
    mapfile -t aggregate_shards < <(planned_aggregate_shards backend-integration-support "${aggregate_name}")
    mapfile -t aggregate_packages < <(planned_aggregate_packages backend-integration-support "${aggregate_name}")
    create_aggregate_report aggregate_dir aggregate_usage "${metadata_dir}" "${aggregate_name}" backend-integration-support "${aggregate_shards[@]}" || status=$?
    aggregate_label="$(planned_aggregate_field backend-integration-support "${aggregate_name}" label)"
    aggregate_phase="$(planned_aggregate_field backend-integration-support "${aggregate_name}" phase)"
    clear_go_selection_env
    emit_declared_support_phase "${aggregate_label}" "${aggregate_usage}" "${aggregate_dir}" "${aggregate_phase}" backend_integration_support "${aggregate_packages[@]}" || status=$?
  done
  rm -rf -- "${metadata_dir}"

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
