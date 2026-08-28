#!/usr/bin/env bash
# shellcheck shell=bash
# shellcheck disable=SC2154

web_e2e_reset_mark_tainted() {
  local taint_marker_file="$1"
  local runtime_root="$2"

  printf '%s\n' "reset_failed" >"$taint_marker_file"
  if [[ -n "$runtime_root" ]]; then
    printf '%s\n' "reset_failed" >"${runtime_root%/}/stack.tainted"
  fi
}

web_e2e_reset_database() {
  local root_dir="$1"
  local reset_id="$2"
  local result_file="$3"
  local runtime_root="${CARTULARY_WEB_E2E_RUNTIME_ROOT:-}"
  local test_services_bin="${CARTULARY_TEST_SERVICES_BIN:-}"
  local metadata_file="${CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE:-}"

  if [[ -z "$runtime_root" ]]; then
    echo "CARTULARY_WEB_E2E_RUNTIME_ROOT is required for Recovery-purpose reset" >&2
    return 2
  fi
  if [[ -z "$test_services_bin" || ! -x "$test_services_bin" ]]; then
    echo "CARTULARY_TEST_SERVICES_BIN must name an executable for Recovery-purpose reset" >&2
    return 2
  fi
  if [[ -z "$metadata_file" ]]; then
    echo "CARTULARY_WEB_E2E_TEST_SERVICES_METADATA_FILE is required for reset" >&2
    return 2
  fi
  "$test_services_bin" reset-web-e2e \
    --credential-root "$runtime_root" \
    --bootstrap-manifest "$root_dir/configs/dev/bootstrap-admin.json" \
    --metadata-file "$metadata_file" \
    --reset-id "$reset_id" \
    --result-file "$result_file"
}

web_e2e_reset_clear_playwright_state() {
  local state_dir="${CARTULARY_PLAYWRIGHT_STATE_DIR:-}"
  local state_marker_file="$1"

  if [[ -n "$state_dir" ]]; then
    mkdir -p "$state_dir"
    find "$state_dir" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
  fi
  printf '%s\n' "cleared" >"$state_marker_file"
}

web_e2e_reset_stack() {
  local root_dir="$1"
  local label="$2"
  local support_dir
  local attempt_file
  local database_file
  local object_store_marker_file
  local state_marker_file
  local backend_ready_marker_file
  local taint_marker_file
  local lease_file="${CARTULARY_WEB_E2E_SESSION_LEASE_FILE:-}"
  local node_bin="${NODE_BIN:-${root_dir}/tmp/node-runtime/bin/node}"
  local generation_before
  local started_ms
  local finished_ms
  local duration_ms
  local status=0
  local outcome="pass"

  support_dir="$(prepare_target_support_dir reset-boundary)"
  attempt_file="${support_dir}/${label}.attempt.json"
  database_file="${support_dir}/${label}.database-reset.json"
  object_store_marker_file="${support_dir}/${label}.object-store-reset"
  state_marker_file="${support_dir}/${label}.state-reset"
  backend_ready_marker_file="${support_dir}/${label}.backend-ready"
  taint_marker_file="${support_dir}/${label}.tainted"
  if [[ ! -x "$node_bin" ]]; then
    node_bin="node"
  fi
  if [[ -z "$lease_file" || ! -f "$lease_file" ]]; then
    echo "failure_class=artifact reason=artifact_error browser reset session lease is missing" >&2
    return 11
  fi
  if ! generation_before="$($node_bin - "$lease_file" <<'EOF'
const fs = require("node:fs");
const lease = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
if (!Number.isSafeInteger(lease.backend_generation) || lease.backend_generation < 1) process.exit(1);
process.stdout.write(String(lease.backend_generation));
EOF
  )"; then
    echo "failure_class=artifact reason=artifact_error browser reset lease has no valid backend generation" >&2
    return 11
  fi

  started_ms="$(step_now_monotonic_ms)"
  if "$root_dir/tools/harness/browser/start-web-e2e.sh" \
    --session-reset-backend \
    --lease-file "$lease_file" \
    --label "$label" \
    --database-result-file "$database_file" \
    --object-store-marker-file "$object_store_marker_file" \
    --state-marker-file "$state_marker_file" \
    --backend-ready-marker-file "$backend_ready_marker_file"; then
    :
  else
    status=$?
    outcome="fail"
    web_e2e_reset_mark_tainted "$taint_marker_file" "${CARTULARY_WEB_E2E_RUNTIME_ROOT:-}"
  fi
  finished_ms="$(step_now_monotonic_ms)"
  duration_ms="$(step_elapsed_ms "$started_ms" "$finished_ms")"

  if ! "$node_bin" "$root_dir/tools/harness/browser/browser-reset-attempt.mjs" \
    --label "$label" \
    --status "$outcome" \
    --exit-code "$status" \
    --attempt-file "$attempt_file" \
    --database-file "$database_file" \
    --object-store-marker-file "$object_store_marker_file" \
    --state-marker-file "$state_marker_file" \
    --backend-ready-marker-file "$backend_ready_marker_file" \
    --lease-file "$lease_file" \
    --generation-before "$generation_before" \
    --duration-ms "$duration_ms"; then
    echo "failure_class=artifact reason=artifact_error browser reset attempt evidence was not published" >&2
    return 11
  fi

  if [[ "$status" -ne 0 ]]; then
    case "$status" in
      2)
        echo "failure_class=config reason=configuration_error browser reset failed" >&2
        ;;
      11)
        echo "failure_class=artifact reason=artifact_error browser reset failed" >&2
        ;;
      13)
        echo "failure_class=timing reason=timeout_failure browser reset failed" >&2
        ;;
      *)
        echo "failure_class=harness reason=fixture_error browser reset failed" >&2
        ;;
    esac
    return "$status"
  fi
}
