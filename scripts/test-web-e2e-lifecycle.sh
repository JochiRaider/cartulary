#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
LIFECYCLE_HELPER="$ROOT_DIR/tools/harness/browser/web-e2e-lifecycle.sh"
cleanup_paths=()
cleanup_pgroups=()

cleanup_test_paths() {
  local path
  local pgid
  for pgid in "${cleanup_pgroups[@]}"; do
    stop_process_group "$pgid" >/dev/null 2>&1 || true
  done
  if declare -F release_port_leases >/dev/null 2>&1; then
    release_port_leases >/dev/null 2>&1 || true
  fi
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
  done
}

trap cleanup_test_paths EXIT

fail() {
  echo "$*" >&2
  exit 1
}

assert_equals() {
  local actual="$1"
  local expected="$2"
  local label="$3"

  if [[ "$actual" != "$expected" ]]; then
    fail "$label: expected [$expected], got [$actual]"
  fi
}

assert_process_running() {
  local pid="$1"
  local label="$2"

  if ! kill -0 "$pid" >/dev/null 2>&1; then
    fail "$label: expected pid $pid to still be running"
  fi
}

assert_process_stopped() {
  local pid="$1"
  local label="$2"

  if kill -0 "$pid" >/dev/null 2>&1; then
    fail "$label: expected pid $pid to be stopped"
  fi
}

wait_for_path() {
  local path="$1"
  local label="$2"

  for _ in $(seq 1 50); do
    if [[ -e "$path" ]]; then
      return 0
    fi
    sleep 0.1
  done

  fail "$label: expected path $path to appear"
}

wait_for_process_exit() {
  local pid="$1"
  local label="$2"

  for _ in $(seq 1 100); do
    if ! kill -0 "$pid" >/dev/null 2>&1; then
      return 0
    fi
    sleep 0.1
  done

  fail "$label: expected pid $pid to exit"
}

count_lines() {
  local path="$1"

  if [[ ! -f "$path" ]]; then
    printf '%s\n' 0
    return
  fi

  wc -l <"$path" | tr -d ' '
}

array_has_prefix_entry() {
  local prefix="$1"
  shift
  local entry=""

  for entry in "$@"; do
    if [[ "$entry" == "$prefix"* ]]; then
      return 0
    fi
  done

  return 1
}

assert_file_contains() {
  local path="$1"
  local expected="$2"
  local label="$3"

  if ! grep -Fq "$expected" "$path"; then
    fail "$label: expected $path to contain [$expected]"
  fi
}

assert_file_not_contains() {
  local path="$1"
  local unexpected="$2"
  local label="$3"

  if grep -Fq "$unexpected" "$path"; then
    fail "$label: expected $path to omit [$unexpected]"
  fi
}

mkdir -p "$ROOT_DIR/tmp"
tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/web-e2e-lifecycle-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")

assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'CARTULARY_PHASE_TIMING_BUCKET=service_wait run_phase_command "browser-e2e startup services"' "browser lifecycle service timing bucket"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" "OBJECT_STORE_CORS_ORIGIN=\"\${PUBLIC_ORIGIN}\"" "browser lifecycle object-store CORS probe uses allocated public origin"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" "OBJECT_STORE_CORS_ALLOWED_ORIGINS=\"\${PUBLIC_ORIGIN}\"" "browser lifecycle object-store CORS allows only allocated public origin"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'CARTULARY_PHASE_TIMING_BUCKET=migration run_phase_command "browser-e2e startup database"' "browser lifecycle migration timing bucket"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'run_timing_span "server_startup" "browser-e2e start backend process"' "browser lifecycle backend startup span"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'run_timing_span "frontend_startup" "browser-e2e start frontend process"' "browser lifecycle frontend startup span"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'run_phase_command "browser-e2e validate frontend preview artifact" require_frontend_preview_artifacts' "browser lifecycle validates built preview artifact"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'run_phase_command "browser-e2e startup frontend ready" start_frontend_preview_ready_with_retry' "browser lifecycle proves frontend before backend startup"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'run_phase_command "browser-e2e verify frontend ready" browser_verify_frontend_ready' "browser lifecycle rechecks frontend after backend readiness"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'exec vite preview --host' "browser lifecycle uses vite preview"
assert_file_not_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'apps/web dev --host' "browser lifecycle must not use vite dev server"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'CARTULARY_BROWSER_STAGE' "browser lifecycle uses scheduler browser stage for port windows"
assert_file_not_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'CARTULARY_TEST_TARGET:-}" == *"stateful"*' "browser lifecycle must not use target substring matching for stateful ports"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" '/api/v1/test/runtime/identity' "browser lifecycle backend identity readiness"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" '"Origin": requestOrigin' "browser lifecycle identity probe origin header"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" './tools/webstacklisten' "browser lifecycle inherited listener helper"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'emit_target_timing_span "teardown" "browser-e2e stop owned processes"' "browser lifecycle process teardown span"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'emit_target_timing_span "teardown" "browser-e2e cleanup standalone database"' "browser lifecycle standalone database teardown span"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'emit_target_timing_span "teardown" "browser-e2e remove runtime root"' "browser lifecycle runtime root teardown span"
assert_file_not_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'emit_target_timing_span "teardown" "browser-e2e owned-stack cleanup"' "browser lifecycle inclusive teardown span"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" "CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN=\"\${E2E_DSN}\"" "browser lifecycle managed postgres env"
assert_file_contains "$ROOT_DIR/scripts/start-web-e2e.sh" "CARTULARY_S3_OBJECT_PRIMARY_BUCKET=\"\${CARTULARY_S3_OBJECT_PRIMARY_BUCKET:-cartulary}\"" "browser lifecycle managed object-store env"
assert_file_not_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'CARTULARY__ROOTS__DATABASE_STORAGE__PATH=' "browser lifecycle must not override managed database root path"
assert_file_not_contains "$ROOT_DIR/scripts/start-web-e2e.sh" 'CARTULARY__ROOTS__OBJECT_STORAGE__PATH=' "browser lifecycle must not override managed object-store root path"

signal_recorder="$tmp_dir/signal-recorder.sh"
cat >"$signal_recorder" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

pid_file="${PID_FILE:?}"
term_file="${TERM_FILE:?}"
mode="${MODE:-loop}"
exit_after_seconds="${EXIT_AFTER_SECONDS:-0.1}"
exit_status="${EXIT_STATUS:-0}"

printf '%s\n' "$$" >"$pid_file"

record_term() {
  printf 'TERM\n' >>"$term_file"
  exit 0
}

trap record_term TERM
trap record_term INT

case "$mode" in
  loop)
    while true; do
      sleep 0.1
    done
    ;;
  exit_after)
    sleep "$exit_after_seconds"
    exit "$exit_status"
    ;;
  *)
    echo "unsupported mode $mode" >&2
    exit 2
    ;;
esac
EOF
chmod +x "$signal_recorder"

supervisor_script="$tmp_dir/exercise-supervisor.sh"
cat >"$supervisor_script" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

source "${LIFECYCLE_HELPER:?}"

mode="${MODE:?}"
signal_recorder="${SIGNAL_RECORDER:?}"
tmp_dir="${TMP_DIR:?}"
server_pgid=""
frontend_pgid=""
child_pgid=""
cleanup_done=0

cleanup() {
  if [[ "$cleanup_done" -eq 1 ]]; then
    return 0
  fi
  cleanup_done=1
  printf '1\n' >"${tmp_dir}/cleanup.count"

  if [[ -n "${child_pgid}" ]]; then
    stop_process_group "${child_pgid}" || true
  fi
  if [[ -n "${frontend_pgid}" ]]; then
    stop_process_group "${frontend_pgid}" || true
  fi
  if [[ -n "${server_pgid}" ]]; then
    stop_process_group "${server_pgid}" || true
  fi
}

launch_service() {
  local outvar="$1"
  local label="$2"
  local service_mode="${3:-loop}"
  local exit_after_seconds="${4:-0.1}"
  local exit_status="${5:-0}"

  start_process_group "$outvar" "" \
    env \
    PID_FILE="${tmp_dir}/${label}.pid" \
    TERM_FILE="${tmp_dir}/${label}.term" \
    MODE="${service_mode}" \
    EXIT_AFTER_SECONDS="${exit_after_seconds}" \
    EXIT_STATUS="${exit_status}" \
    "${signal_recorder}"
}

supervise_stack() {
  local child_status=0

  while true; do
    if lifecycle_shutdown_requested; then
      return "$(lifecycle_signal_exit_status)"
    fi

    if ! process_group_running "${server_pgid}"; then
      if wait "${server_pgid}"; then
        :
      fi
      if [[ -n "${child_pgid}" ]]; then
        stop_process_group "${child_pgid}" || true
      fi
      return 1
    fi

    if ! process_group_running "${frontend_pgid}"; then
      if wait "${frontend_pgid}"; then
        :
      fi
      if [[ -n "${child_pgid}" ]]; then
        stop_process_group "${child_pgid}" || true
      fi
      return 1
    fi

    if [[ -n "${child_pgid}" ]] && ! process_group_running "${child_pgid}"; then
      if wait "${child_pgid}"; then
        child_status=0
      else
        child_status=$?
      fi
      return "${child_status}"
    fi

    sleep 0.1
  done
}

trap cleanup EXIT
lifecycle_reset_shutdown_state
lifecycle_install_signal_traps

case "$mode" in
  server-only)
    launch_service server_pgid server
    launch_service frontend_pgid frontend
    ;;
  child-success)
    launch_service server_pgid server
    launch_service frontend_pgid frontend
    launch_service child_pgid child exit_after 0.2 17
    ;;
  server-exits)
    launch_service server_pgid server exit_after 0.2 0
    launch_service frontend_pgid frontend
    launch_service child_pgid child
    ;;
  frontend-exits)
    launch_service server_pgid server
    launch_service frontend_pgid frontend exit_after 0.2 0
    launch_service child_pgid child
    ;;
  *)
    echo "unsupported supervisor mode $mode" >&2
    exit 2
    ;;
esac

supervise_stack
EOF
chmod +x "$supervisor_script"

parent_death_launcher="$tmp_dir/parent-death-launcher.sh"
cat >"$parent_death_launcher" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

source "${LIFECYCLE_HELPER:?}"

signal_recorder="${SIGNAL_RECORDER:?}"
tmp_dir="${TMP_DIR:?}"
service_group=""

start_process_group service_group "" \
  env \
  PID_FILE="${tmp_dir}/orphan.pid" \
  TERM_FILE="${tmp_dir}/orphan.term" \
  MODE=loop \
  "${signal_recorder}"
printf '%s\n' "${service_group}" >"${tmp_dir}/orphan.group"

while true; do
  sleep 1
done
EOF
chmod +x "$parent_death_launcher"

failing_test_services_bin="$tmp_dir/failing-test-services.sh"
cat >"$failing_test_services_bin" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "cleanup-web-e2e" ]]; then
  echo "fixture cleanup failed" >&2
  exit 31
fi

echo "unexpected test-services command: $*" >&2
exit 2
EOF
chmod +x "$failing_test_services_bin"

cleanup_failure_runner="$tmp_dir/exercise-start-web-e2e-cleanup-failure.sh"
cat >"$cleanup_failure_runner" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

source "${ROOT_DIR:?}/scripts/start-web-e2e.sh"

tmp_dir="${TMP_DIR:?}"
signal_recorder="${SIGNAL_RECORDER:?}"
TEST_SERVICES_BIN="${FAILING_TEST_SERVICES_BIN:?}"
TEST_SERVICES_METADATA_FILE="${tmp_dir}/test-services-web-e2e.json"
RUNTIME_ROOT_BASE="${tmp_dir}/runtime-root"
KEEP_RUNTIME_ROOT=0
BACKEND_PORT=1
FRONTEND_PORT=2

mkdir -p "$RUNTIME_ROOT_BASE"
printf '{"database_name":"ct_web","bucket":"ct-web"}\n' >"$TEST_SERVICES_METADATA_FILE"

start_process_group CHILD_PGID "" \
  env \
  PID_FILE="${tmp_dir}/child.pid" \
  TERM_FILE="${tmp_dir}/child.term" \
  MODE=loop \
  "$signal_recorder"

for _ in $(seq 1 50); do
  if [[ -f "${tmp_dir}/child.pid" ]]; then
    break
  fi
  sleep 0.1
done

trap on_exit EXIT
export CARTULARY_TEST_SERVICES_ACTIVE=1
exit 0
EOF
chmod +x "$cleanup_failure_runner"

single_stop_pid_file="$tmp_dir/single-stop.pid"
single_stop_term_file="$tmp_dir/single-stop.term"
# shellcheck source=tools/harness/browser/web-e2e-lifecycle.sh
source "$LIFECYCLE_HELPER"
lifecycle_reset_shutdown_state
single_stop_group=""
start_process_group single_stop_group "" \
  env \
  PID_FILE="$single_stop_pid_file" \
  TERM_FILE="$single_stop_term_file" \
  MODE=loop \
  "$signal_recorder"
wait_for_path "$single_stop_pid_file" "single stop pid file"
stop_process_group "$single_stop_group"
stop_process_group "$single_stop_group"
single_stop_pid="$(tr -d '\n' <"$single_stop_pid_file")"
assert_process_stopped "$single_stop_pid" "single stop process"
assert_equals "$(count_lines "$single_stop_term_file")" "1" "single stop term count"

parent_death_dir="$tmp_dir/parent-death"
mkdir -p "$parent_death_dir"
LIFECYCLE_HELPER="$LIFECYCLE_HELPER" \
SIGNAL_RECORDER="$signal_recorder" \
TMP_DIR="$parent_death_dir" \
  "$parent_death_launcher" >/dev/null 2>&1 &
parent_death_pid=$!
wait_for_path "$parent_death_dir/orphan.pid" "parent-death orphan pid"
orphan_pid="$(tr -d '\n' <"$parent_death_dir/orphan.pid")"
kill -KILL "$parent_death_pid" >/dev/null 2>&1 || true
if wait "$parent_death_pid"; then
  parent_death_status=0
else
  parent_death_status=$?
fi
assert_equals "$parent_death_status" "137" "parent-death launcher status"
wait_for_process_exit "$orphan_pid" "parent-death orphan process"
assert_equals "$(count_lines "$parent_death_dir/orphan.term")" "1" "parent-death orphan term count"

server_only_dir="$tmp_dir/server-only"
mkdir -p "$server_only_dir"
MODE=server-only \
LIFECYCLE_HELPER="$LIFECYCLE_HELPER" \
SIGNAL_RECORDER="$signal_recorder" \
TMP_DIR="$server_only_dir" \
  "$supervisor_script" >/dev/null 2>&1 &
server_only_pid=$!
wait_for_path "$server_only_dir/server.pid" "server-only server pid"
wait_for_path "$server_only_dir/frontend.pid" "server-only frontend pid"
assert_process_running "$server_only_pid" "server-only supervisor"
kill -TERM "$server_only_pid"
if wait "$server_only_pid"; then
  server_only_status=0
else
  server_only_status=$?
fi
assert_equals "$server_only_status" "143" "server-only shutdown status"
assert_equals "$(tr -d '\n' <"$server_only_dir/cleanup.count")" "1" "server-only cleanup count"
assert_equals "$(count_lines "$server_only_dir/server.term")" "1" "server-only server term count"
assert_equals "$(count_lines "$server_only_dir/frontend.term")" "1" "server-only frontend term count"
assert_process_stopped "$(tr -d '\n' <"$server_only_dir/server.pid")" "server-only server process"
assert_process_stopped "$(tr -d '\n' <"$server_only_dir/frontend.pid")" "server-only frontend process"

child_success_dir="$tmp_dir/child-success"
mkdir -p "$child_success_dir"
if MODE=child-success \
  LIFECYCLE_HELPER="$LIFECYCLE_HELPER" \
  SIGNAL_RECORDER="$signal_recorder" \
  TMP_DIR="$child_success_dir" \
  "$supervisor_script" >/dev/null 2>&1; then
  child_success_status=0
else
  child_success_status=$?
fi
assert_equals "$child_success_status" "17" "child-success status"
assert_equals "$(tr -d '\n' <"$child_success_dir/cleanup.count")" "1" "child-success cleanup count"
assert_equals "$(count_lines "$child_success_dir/server.term")" "1" "child-success server term count"
assert_equals "$(count_lines "$child_success_dir/frontend.term")" "1" "child-success frontend term count"
assert_process_stopped "$(tr -d '\n' <"$child_success_dir/server.pid")" "child-success server process"
assert_process_stopped "$(tr -d '\n' <"$child_success_dir/frontend.pid")" "child-success frontend process"
assert_process_stopped "$(tr -d '\n' <"$child_success_dir/child.pid")" "child-success child process"

server_exit_dir="$tmp_dir/server-exits"
mkdir -p "$server_exit_dir"
if MODE=server-exits \
  LIFECYCLE_HELPER="$LIFECYCLE_HELPER" \
  SIGNAL_RECORDER="$signal_recorder" \
  TMP_DIR="$server_exit_dir" \
  "$supervisor_script" >/dev/null 2>&1; then
  server_exit_status=0
else
  server_exit_status=$?
fi
assert_equals "$server_exit_status" "1" "server-exits status"
assert_equals "$(tr -d '\n' <"$server_exit_dir/cleanup.count")" "1" "server-exits cleanup count"
assert_equals "$(count_lines "$server_exit_dir/frontend.term")" "1" "server-exits frontend term count"
assert_equals "$(count_lines "$server_exit_dir/child.term")" "1" "server-exits child term count"
assert_process_stopped "$(tr -d '\n' <"$server_exit_dir/frontend.pid")" "server-exits frontend process"
assert_process_stopped "$(tr -d '\n' <"$server_exit_dir/child.pid")" "server-exits child process"

frontend_exit_dir="$tmp_dir/frontend-exits"
mkdir -p "$frontend_exit_dir"
if MODE=frontend-exits \
  LIFECYCLE_HELPER="$LIFECYCLE_HELPER" \
  SIGNAL_RECORDER="$signal_recorder" \
  TMP_DIR="$frontend_exit_dir" \
  "$supervisor_script" >/dev/null 2>&1; then
  frontend_exit_status=0
else
  frontend_exit_status=$?
fi
assert_equals "$frontend_exit_status" "1" "frontend-exits status"
assert_equals "$(tr -d '\n' <"$frontend_exit_dir/cleanup.count")" "1" "frontend-exits cleanup count"
assert_equals "$(count_lines "$frontend_exit_dir/server.term")" "1" "frontend-exits server term count"
assert_equals "$(count_lines "$frontend_exit_dir/child.term")" "1" "frontend-exits child term count"
assert_process_stopped "$(tr -d '\n' <"$frontend_exit_dir/server.pid")" "frontend-exits server process"
assert_process_stopped "$(tr -d '\n' <"$frontend_exit_dir/child.pid")" "frontend-exits child process"

cleanup_failure_dir="$tmp_dir/cleanup-failure"
mkdir -p "$cleanup_failure_dir"
cleanup_failure_stderr="$cleanup_failure_dir/stderr.log"
if ROOT_DIR="$ROOT_DIR" \
  TMP_DIR="$cleanup_failure_dir" \
  SIGNAL_RECORDER="$signal_recorder" \
  FAILING_TEST_SERVICES_BIN="$failing_test_services_bin" \
  "$cleanup_failure_runner" >/dev/null 2>"$cleanup_failure_stderr"; then
  cleanup_failure_status=0
else
  cleanup_failure_status=$?
fi
assert_equals "$cleanup_failure_status" "31" "start-web-e2e cleanup failure status"
assert_file_contains "$cleanup_failure_stderr" "fixture cleanup failed" "start-web-e2e cleanup failure stderr"
assert_file_contains "$cleanup_failure_stderr" "browser e2e cleanup failed with status 31" "start-web-e2e cleanup failure summary"
assert_process_stopped "$(tr -d '\n' <"$cleanup_failure_dir/child.pid")" "start-web-e2e cleanup failure child process"
if [[ -e "$cleanup_failure_dir/runtime-root" ]]; then
  fail "start-web-e2e cleanup failure must still remove the runtime root"
fi

source "$ROOT_DIR/tools/harness/browser/playwright-owned-stack.sh"
resolve_playwright_owned_stack_env "$ROOT_DIR"
if array_has_prefix_entry "CARTULARY_SERVER_BIN=" "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}"; then
  fail "playwright owned stack env must not inject CARTULARY_SERVER_BIN by default"
fi
if array_has_prefix_entry "CARTULARY_MIGRATE_BIN=" "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}"; then
  fail "playwright owned stack env must not inject CARTULARY_MIGRATE_BIN by default"
fi
CARTULARY_WEB_E2E_API_ORIGIN="http://127.0.0.1:18081"
CARTULARY_WEB_E2E_PUBLIC_ORIGIN="http://127.0.0.1:14173"
resolve_playwright_owned_stack_env "$ROOT_DIR"
if ! array_has_prefix_entry "CARTULARY_WEB_E2E_API_ORIGIN=http://127.0.0.1:18081" "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}"; then
  fail "playwright owned stack env must pass through CARTULARY_WEB_E2E_API_ORIGIN"
fi
if ! array_has_prefix_entry "CARTULARY_WEB_E2E_PUBLIC_ORIGIN=http://127.0.0.1:14173" "${PLAYWRIGHT_OWNED_STACK_COMMON_ENV[@]}"; then
  fail "playwright owned stack env must pass through CARTULARY_WEB_E2E_PUBLIC_ORIGIN"
fi
unset CARTULARY_WEB_E2E_API_ORIGIN
unset CARTULARY_WEB_E2E_PUBLIC_ORIGIN

command_override="$tmp_dir/explicit-runtime-bin.sh"
cat >"$command_override" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
chmod +x "$command_override"

repo_server_artifact="$tmp_dir/repo-server"
cat >"$repo_server_artifact" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
chmod +x "$repo_server_artifact"

repo_migrate_artifact="$tmp_dir/repo-migrate"
cat >"$repo_migrate_artifact" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
chmod +x "$repo_migrate_artifact"

source "$ROOT_DIR/scripts/start-web-e2e.sh"
GO_BIN="go-test-bin"

resolved_command=()
resolve_runtime_command resolved_command "backend" "$repo_server_artifact" "$repo_server_artifact" ./cmd/server
assert_equals "${resolved_command[*]}" "go-test-bin run ./cmd/server" "repo-root backend artifact ignored by default"

resolve_runtime_command resolved_command "migration" "$repo_migrate_artifact" "$repo_migrate_artifact" ./cmd/migrate
assert_equals "${resolved_command[*]}" "go-test-bin run ./cmd/migrate" "repo-root migrate artifact ignored by default"

resolve_runtime_command resolved_command "backend" "$command_override" "$repo_server_artifact" ./cmd/server
assert_equals "${resolved_command[*]}" "$command_override" "explicit backend override honored"

missing_override_stderr="$tmp_dir/missing-runtime-bin.stderr"
if resolve_runtime_command resolved_command "backend" "$tmp_dir/missing-runtime-bin" "$repo_server_artifact" ./cmd/server 2>"$missing_override_stderr"; then
  fail "missing explicit backend override must fail fast"
fi
if ! grep -Fq "is not executable" "$missing_override_stderr"; then
  fail "missing explicit backend override must explain the executable requirement"
fi

export CARTULARY_WEB_E2E_USE_REPO_ROOT_BINARIES=1
resolve_runtime_command resolved_command "backend" "$repo_server_artifact" "$repo_server_artifact" ./cmd/server
assert_equals "${resolved_command[*]}" "$repo_server_artifact" "repo-root backend artifact opt-in honored"
resolve_runtime_command resolved_command "migration" "$repo_migrate_artifact" "$repo_migrate_artifact" ./cmd/migrate
assert_equals "${resolved_command[*]}" "$repo_migrate_artifact" "repo-root migrate artifact opt-in honored"
unset CARTULARY_WEB_E2E_USE_REPO_ROOT_BINARIES

resolve_owned_stack_ports
dynamic_backend_port="$BACKEND_PORT"
dynamic_frontend_port="$FRONTEND_PORT"
if [[ -z "$dynamic_backend_port" || -z "$dynamic_frontend_port" ]]; then
  fail "dynamic browser e2e port allocation must set backend and frontend ports"
fi
if [[ "$dynamic_backend_port" == "$dynamic_frontend_port" ]]; then
  fail "dynamic browser e2e port allocation must choose distinct backend and frontend ports"
fi
assert_equals "$CARTULARY_WEB_E2E_API_ORIGIN" "http://127.0.0.1:${dynamic_backend_port}" "dynamic browser e2e API origin export"
assert_equals "$CARTULARY_WEB_E2E_PUBLIC_ORIGIN" "http://127.0.0.1:${dynamic_frontend_port}" "dynamic browser e2e public origin export"

CARTULARY_WEB_E2E_BACKEND_PORT="$dynamic_backend_port"
CARTULARY_WEB_E2E_FRONTEND_PORT="$dynamic_frontend_port"
resolve_owned_stack_ports
assert_equals "$BACKEND_PORT" "$dynamic_backend_port" "explicit browser e2e backend port override"
assert_equals "$FRONTEND_PORT" "$dynamic_frontend_port" "explicit browser e2e frontend port override"
release_port_leases
unset CARTULARY_WEB_E2E_BACKEND_PORT
unset CARTULARY_WEB_E2E_FRONTEND_PORT

CARTULARY_TEST_SERVICES_ACTIVE=1
CARTULARY_BROWSER_STAGE=stateful
CARTULARY_TEST_TARGET=browser-e2e-webserver-backed
resolve_owned_stack_ports
if (( FRONTEND_PORT < 39100 || FRONTEND_PORT > 39199 )); then
  fail "stateful browser stage must allocate frontend port from 39100-39199, got ${FRONTEND_PORT}"
fi
release_port_leases

CARTULARY_BROWSER_STAGE=webserver-backed
CARTULARY_TEST_TARGET=browser-e2e-stateful
resolve_owned_stack_ports
if (( FRONTEND_PORT < 39000 || FRONTEND_PORT > 39099 )); then
  fail "webserver-backed browser stage must allocate frontend port from 39000-39099, got ${FRONTEND_PORT}"
fi
release_port_leases
unset CARTULARY_TEST_SERVICES_ACTIVE
unset CARTULARY_BROWSER_STAGE
unset CARTULARY_TEST_TARGET
BACKEND_PORT="$dynamic_backend_port"
FRONTEND_PORT="$dynamic_frontend_port"
API_ORIGIN="http://127.0.0.1:${BACKEND_PORT}"
PUBLIC_ORIGIN="http://127.0.0.1:${FRONTEND_PORT}"

STACK_ENV_FILE="$tmp_dir/stack.env"
STACK_JSON_FILE="$tmp_dir/stack.json"
RUNTIME_ROOT_BASE="$tmp_dir/runtime-root"
SERVER_LOG="$tmp_dir/server.log"
WEB_LOG="$tmp_dir/web.log"
STARTUP_DIAGNOSTIC_FILE="$tmp_dir/startup-diagnostics.json"
write_stack_metadata
assert_file_contains "$STACK_ENV_FILE" "CARTULARY_WEB_E2E_API_ORIGIN=http://127.0.0.1:${dynamic_backend_port}" "stack env API origin"
assert_file_contains "$STACK_ENV_FILE" "CARTULARY_WEB_E2E_PUBLIC_ORIGIN=http://127.0.0.1:${dynamic_frontend_port}" "stack env public origin"
assert_file_contains "$STACK_ENV_FILE" "CARTULARY_WEB_E2E_FRONTEND_MODE=preview" "stack env frontend mode"
"${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}" - "$STACK_JSON_FILE" "$dynamic_backend_port" "$dynamic_frontend_port" "$SERVER_LOG" "$WEB_LOG" "$STARTUP_DIAGNOSTIC_FILE" <<'EOF'
const fs = require("node:fs");

const [path, backendPort, frontendPort, serverLog, webLog, startupDiagnostics] = process.argv.slice(2);
const payload = JSON.parse(fs.readFileSync(path, "utf8"));
const failures = [];
if (payload.schema_id !== "cartulary.web_e2e_stack.v3") {
  failures.push(`unexpected schema_id ${payload.schema_id}`);
}
if (payload.api_origin !== `http://127.0.0.1:${backendPort}`) {
  failures.push(`unexpected api_origin ${payload.api_origin}`);
}
if (payload.public_origin !== `http://127.0.0.1:${frontendPort}`) {
  failures.push(`unexpected public_origin ${payload.public_origin}`);
}
if (payload.backend_port !== Number.parseInt(backendPort, 10)) {
  failures.push(`unexpected backend_port ${payload.backend_port}`);
}
if (payload.frontend_port !== Number.parseInt(frontendPort, 10)) {
  failures.push(`unexpected frontend_port ${payload.frontend_port}`);
}
if (payload.frontend_mode !== "preview") {
  failures.push(`unexpected frontend_mode ${payload.frontend_mode}`);
}
if (payload.frontend_command_kind !== "vite-preview") {
  failures.push(`unexpected frontend_command_kind ${payload.frontend_command_kind}`);
}
if (payload.server_log !== serverLog) {
  failures.push(`unexpected server_log ${payload.server_log}`);
}
if (payload.web_log !== webLog) {
  failures.push(`unexpected web_log ${payload.web_log}`);
}
if (payload.startup_diagnostics !== startupDiagnostics) {
  failures.push(`unexpected startup_diagnostics ${payload.startup_diagnostics}`);
}
if (failures.length > 0) {
  console.error(failures.join("\n"));
  process.exit(1);
}
EOF

original_web_dist_index="$WEB_DIST_INDEX"
WEB_DIST_INDEX="$tmp_dir/missing-dist/index.html"
preview_artifact_stderr="$tmp_dir/preview-artifact.stderr"
if require_frontend_preview_artifacts 2>"$preview_artifact_stderr"; then
  fail "missing preview artifact preflight must fail"
else
  preview_artifact_status=$?
  if [[ "$preview_artifact_status" -ne 2 ]]; then
    fail "missing preview artifact preflight must return 2, got $preview_artifact_status"
  fi
fi
assert_file_contains "$preview_artifact_stderr" "run make build-web before browser e2e" "missing preview artifact remediation"
"${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}" - "$STARTUP_DIAGNOSTIC_FILE" <<'EOF'
const fs = require("node:fs");

const payload = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const failures = [];
if (payload.schema_id !== "cartulary.browser_startup_diagnostics.v1") {
  failures.push(`unexpected diagnostic schema ${payload.schema_id}`);
}
if (payload.status !== "fail") {
  failures.push(`unexpected diagnostic status ${payload.status}`);
}
if (payload.failure_class !== "config") {
  failures.push(`unexpected diagnostic failure_class ${payload.failure_class}`);
}
if (payload.failure_reason !== "configuration_error") {
  failures.push(`unexpected diagnostic failure_reason ${payload.failure_reason}`);
}
if (payload.frontend_mode !== "preview") {
  failures.push(`unexpected diagnostic frontend_mode ${payload.frontend_mode}`);
}
if (payload.frontend_command_kind !== "vite-preview") {
  failures.push(`unexpected diagnostic frontend_command_kind ${payload.frontend_command_kind}`);
}
if (failures.length > 0) {
  console.error(failures.join("\n"));
  process.exit(1);
}
EOF
mkdir -p "$(dirname "$WEB_DIST_INDEX")"
touch "$WEB_DIST_INDEX"
require_frontend_preview_artifacts
WEB_DIST_INDEX="$original_web_dist_index"

fake_pnpm="$tmp_dir/fake-pnpm"
cat >"$fake_pnpm" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

port=""
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --port)
      port="${2:-}"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done
if [[ -z "$port" ]]; then
  echo "fake pnpm expected --port" >&2
  exit 2
fi

attempt_file="${FAKE_PREVIEW_ATTEMPT_FILE:?}"
port_file="${FAKE_PREVIEW_PORT_FILE:?}"
mode="${FAKE_PREVIEW_MODE:-retry_once}"
attempt=1
if [[ -f "$attempt_file" ]]; then
  attempt="$(( $(tr -d '\n' <"$attempt_file") + 1 ))"
fi
printf '%s\n' "$attempt" >"$attempt_file"
printf '%s\n' "$port" >>"$port_file"

if [[ "$mode" == "always_conflict" || "$attempt" -eq 1 ]]; then
  echo "Error: Port ${port} is already in use" >&2
  exit 1
fi

exec "${NODE_BIN:-node}" - "$port" <<'JS'
const http = require("node:http");
const port = Number.parseInt(process.argv[2], 10);
const server = http.createServer((_request, response) => {
  response.writeHead(200, { "content-type": "text/plain" });
  response.end("ready\n");
});
server.listen(port, "127.0.0.1");
JS
EOF
chmod +x "$fake_pnpm"

retry_stack_dir="$tmp_dir/retry-owned-stack"
mkdir -p "$retry_stack_dir"
TARGET_ARTIFACT_DIR="$retry_stack_dir"
RUNTIME_ROOT_BASE="$retry_stack_dir/runtime-root"
SERVER_LOG="$retry_stack_dir/server.log"
WEB_LOG="$retry_stack_dir/web.log"
STACK_ENV_FILE="$retry_stack_dir/stack.env"
STACK_JSON_FILE="$retry_stack_dir/stack.json"
STARTUP_DIAGNOSTIC_FILE="$retry_stack_dir/startup-diagnostics.json"
PLAYWRIGHT_STATE_DIR="$retry_stack_dir/playwright-state"
mkdir -p "$RUNTIME_ROOT_BASE" "$PLAYWRIGHT_STATE_DIR"
CARTULARY_TEST_SERVICES_ACTIVE=1
CARTULARY_BROWSER_STAGE=stateful
CARTULARY_TEST_TARGET=browser-e2e-stateful
resolve_owned_stack_ports
retry_first_port="$FRONTEND_PORT"
FAKE_PREVIEW_ATTEMPT_FILE="$tmp_dir/fake-preview-retry-attempt"
FAKE_PREVIEW_PORT_FILE="$tmp_dir/fake-preview-retry-ports"
export FAKE_PREVIEW_ATTEMPT_FILE FAKE_PREVIEW_PORT_FILE
start_frontend_preview_ready_with_retry "$fake_pnpm"
retry_attempts="$(tr -d '\n' <"$FAKE_PREVIEW_ATTEMPT_FILE")"
assert_equals "$retry_attempts" "2" "auto-selected frontend strict-port conflict retries once"
retry_final_port="$(tail -n 1 "$FAKE_PREVIEW_PORT_FILE")"
if [[ "$retry_final_port" == "$retry_first_port" ]]; then
  fail "auto-selected frontend retry must choose a different port"
fi
"${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}" - "$STARTUP_DIAGNOSTIC_FILE" "$retry_final_port" <<'EOF'
const fs = require("node:fs");

const [diagnosticFile, finalPort] = process.argv.slice(2);
const payload = JSON.parse(fs.readFileSync(diagnosticFile, "utf8"));
const failures = [];
if (payload.status !== "pass") {
  failures.push(`unexpected retry diagnostic status ${payload.status}`);
}
if (payload.failure_reason !== undefined) {
  failures.push(`retry success diagnostic must not retain failure_reason ${payload.failure_reason}`);
}
if (payload.frontend_port !== Number.parseInt(finalPort, 10)) {
  failures.push(`unexpected retry diagnostic frontend_port ${payload.frontend_port}`);
}
if (failures.length > 0) {
  console.error(failures.join("\n"));
  process.exit(1);
}
EOF
stop_process_group "$VITE_PGID" || true
VITE_PGID=""
release_port_leases

rm -f "$FAKE_PREVIEW_ATTEMPT_FILE" "$FAKE_PREVIEW_PORT_FILE"
resolve_owned_stack_ports
explicit_retry_backend_port="$BACKEND_PORT"
explicit_retry_frontend_port="$FRONTEND_PORT"
release_port_leases
CARTULARY_WEB_E2E_BACKEND_PORT="$explicit_retry_backend_port"
CARTULARY_WEB_E2E_FRONTEND_PORT="$explicit_retry_frontend_port"
resolve_owned_stack_ports
FAKE_PREVIEW_ATTEMPT_FILE="$tmp_dir/fake-preview-explicit-attempt"
FAKE_PREVIEW_PORT_FILE="$tmp_dir/fake-preview-explicit-ports"
FAKE_PREVIEW_MODE=always_conflict
export FAKE_PREVIEW_ATTEMPT_FILE FAKE_PREVIEW_PORT_FILE FAKE_PREVIEW_MODE
if start_frontend_preview_ready_with_retry "$fake_pnpm" >/dev/null 2>&1; then
  fail "explicit frontend strict-port conflict must not retry to success"
fi
explicit_attempts="$(tr -d '\n' <"$FAKE_PREVIEW_ATTEMPT_FILE")"
assert_equals "$explicit_attempts" "1" "explicit frontend strict-port conflict must not retry"
"${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}" - "$STARTUP_DIAGNOSTIC_FILE" <<'EOF'
const fs = require("node:fs");

const payload = JSON.parse(fs.readFileSync(process.argv[2], "utf8"));
const failures = [];
if (payload.status !== "fail") {
  failures.push(`unexpected explicit diagnostic status ${payload.status}`);
}
if (payload.failure_class !== "infra") {
  failures.push(`unexpected explicit diagnostic class ${payload.failure_class}`);
}
if (payload.failure_reason !== "resource_conflict") {
  failures.push(`unexpected explicit diagnostic reason ${payload.failure_reason}`);
}
if (failures.length > 0) {
  console.error(failures.join("\n"));
  process.exit(1);
}
EOF
stop_process_group "${VITE_PGID:-}" || true
VITE_PGID=""
release_port_leases
unset CARTULARY_TEST_SERVICES_ACTIVE
unset CARTULARY_BROWSER_STAGE
unset CARTULARY_TEST_TARGET
unset CARTULARY_WEB_E2E_BACKEND_PORT
unset CARTULARY_WEB_E2E_FRONTEND_PORT
unset FAKE_PREVIEW_ATTEMPT_FILE
unset FAKE_PREVIEW_PORT_FILE
unset FAKE_PREVIEW_MODE

identity_server="$tmp_dir/identity-server.mjs"
cat >"$identity_server" <<'EOF'
import fs from "node:fs";
import http from "node:http";

const mode = process.env.IDENTITY_MODE ?? "valid";
const token = process.env.IDENTITY_TOKEN ?? "";
const portFile = process.env.IDENTITY_PORT_FILE;

const server = http.createServer((request, response) => {
  if (request.url === "/readyz") {
    response.writeHead(200, { "content-type": "text/plain" });
    response.end("ready\n");
    return;
  }
  if (request.url !== "/api/v1/test/runtime/identity" || mode !== "valid") {
    response.writeHead(404, { "content-type": "text/plain" });
    response.end("not found\n");
    return;
  }
  if (request.headers["x-cartulary-test-route-token"] !== token) {
    response.writeHead(403, { "content-type": "application/json" });
    response.end(JSON.stringify({ error: { code: "test_route_forbidden" } }));
    return;
  }
  response.writeHead(200, { "content-type": "application/json" });
  response.end(JSON.stringify({
    data: {
      schema_id: "cartulary.test.runtime_identity.v1",
      runtime_marker: "harness-owned",
      server_pid: process.pid,
      test_routes_enabled: true
    }
  }));
});

server.listen(0, "127.0.0.1", () => {
  fs.writeFileSync(portFile, `${server.address().port}\n`);
});
EOF

identity_port_file="$tmp_dir/identity.port"
start_process_group IDENTITY_PGID "" \
  env \
  IDENTITY_MODE=valid \
  IDENTITY_TOKEN=0123456789abcdef0123456789abcdef \
  IDENTITY_PORT_FILE="$identity_port_file" \
  "${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}" "$identity_server"
cleanup_pgroups+=("$IDENTITY_PGID")
wait_for_path "$identity_port_file" "identity server port"
BACKEND_PORT="$(tr -d '\n' <"$identity_port_file")"
API_ORIGIN="http://127.0.0.1:${BACKEND_PORT}"
TEST_ROUTE_TOKEN="0123456789abcdef0123456789abcdef"
if ! port_owned_by_process_group "$BACKEND_PORT" "$IDENTITY_PGID"; then
  fail "identity server listener must be owned by its process group"
fi
if port_owned_by_process_group "$BACKEND_PORT" "$$"; then
  fail "identity server listener must not be owned by the test shell process group"
fi
identity_pid="$(probe_backend_identity)"
if [[ -z "$identity_pid" ]]; then
  fail "identity probe must return the backend server pid"
fi

stale_port_file="$tmp_dir/stale.port"
start_process_group STALE_PGID "" \
  env \
  IDENTITY_MODE=ready-only \
  IDENTITY_TOKEN=0123456789abcdef0123456789abcdef \
  IDENTITY_PORT_FILE="$stale_port_file" \
  "${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}" "$identity_server"
cleanup_pgroups+=("$STALE_PGID")
wait_for_path "$stale_port_file" "stale listener port"
BACKEND_PORT="$(tr -d '\n' <"$stale_port_file")"
API_ORIGIN="http://127.0.0.1:${BACKEND_PORT}"
if probe_backend_identity >/dev/null 2>&1; then
  fail "ready-only stale listener must not satisfy backend identity readiness"
fi

occupied_backend_stderr="$tmp_dir/occupied-backend.stderr"
CARTULARY_WEB_E2E_BACKEND_PORT="$BACKEND_PORT"
CARTULARY_WEB_E2E_FRONTEND_PORT=""
if resolve_owned_stack_ports 2>"$occupied_backend_stderr"; then
  fail "occupied explicit backend port must fail before stack startup"
fi
if ! grep -Fq "already in use" "$occupied_backend_stderr"; then
  fail "occupied backend port failure must mention that the port is already in use"
fi
unset CARTULARY_WEB_E2E_BACKEND_PORT
unset CARTULARY_WEB_E2E_FRONTEND_PORT
BACKEND_PORT="$dynamic_backend_port"
FRONTEND_PORT="$dynamic_frontend_port"
API_ORIGIN="http://127.0.0.1:${BACKEND_PORT}"
PUBLIC_ORIGIN="http://127.0.0.1:${FRONTEND_PORT}"

fake_curl_dir="$tmp_dir/fake-curl-bin"
mkdir -p "$fake_curl_dir"
fake_curl_url_file="$tmp_dir/fake-curl-url.txt"
fake_curl_header_file="$tmp_dir/fake-curl-headers.txt"
cat >"$fake_curl_dir/curl" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

output_file=""
url=""
: >"${FAKE_CURL_HEADER_FILE:?}"
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    -o)
      output_file="$2"
      shift 2
      ;;
    -H)
      printf '%s\n' "$2" >>"${FAKE_CURL_HEADER_FILE:?}"
      shift 2
      ;;
    -w|-X|--max-time)
      shift 2
      ;;
    -sS)
      shift
      ;;
    *)
      url="$1"
      shift
      ;;
  esac
done

printf '%s\n' "$url" >"${FAKE_CURL_URL_FILE:?}"
cat >"$output_file" <<JSON
{"data":{"schema_id":"cartulary.test.runtime_reset.v1","reset_id":"reset-smoke","tables_reset":["incidents"],"mutable_table_count":1,"object_count_removed":0,"object_count_after":0,"migration_metadata_preserved":true,"bootstrap_admin_restored":true,"partial_failure":false,"post_reset_counts":{"active_deployment_admins":1,"bootstrap_markers":1,"incidents":0,"records":0,"user_sessions":0,"route_idempotency":0}}}
JSON
printf '200'
EOF
chmod +x "$fake_curl_dir/curl"
FAKE_CURL_URL_FILE="$fake_curl_url_file" \
FAKE_CURL_HEADER_FILE="$fake_curl_header_file" \
PATH="$fake_curl_dir:$PATH" \
CARTULARY_TEST_RESULTS_DIR="$tmp_dir/reset-results" \
CARTULARY_TEST_RUN_ID="reset-smoke" \
CARTULARY_TEST_TARGET="browser-e2e-resettable" \
CARTULARY_WEB_E2E_API_ORIGIN="http://127.0.0.1:${dynamic_backend_port}" \
CARTULARY_WEB_E2E_PUBLIC_ORIGIN="http://127.0.0.1:${dynamic_frontend_port}" \
CARTULARY_TEST_ROUTE_TOKEN="0123456789abcdef0123456789abcdef" \
NODE_BIN="${NODE_BIN:-$ROOT_DIR/tmp/node-runtime/bin/node}" \
  "$ROOT_DIR/scripts/reset-web-e2e-stack.sh" --label dynamic-origin
assert_equals "$(tr -d '\n' <"$fake_curl_url_file")" "http://127.0.0.1:${dynamic_backend_port}/api/v1/test/runtime/reset" "reset helper dynamic API origin"
assert_file_contains "$fake_curl_header_file" "Origin: http://127.0.0.1:${dynamic_frontend_port}" "reset helper declared browser origin"

unset CARTULARY_WEB_E2E_BACKEND_PORT
unset CARTULARY_WEB_E2E_FRONTEND_PORT
unset CARTULARY_WEB_E2E_API_ORIGIN
unset CARTULARY_WEB_E2E_PUBLIC_ORIGIN

CARTULARY_TEST_SERVICES_ACTIVE=1
resolve_owned_stack_ports
if (( FRONTEND_PORT < 39000 || FRONTEND_PORT > 39099 )); then
  fail "active test-service browser frontend port must stay inside the webserver-backed SeaweedFS CORS range, got ${FRONTEND_PORT}"
fi

CARTULARY_TEST_TARGET="browser-e2e-stateful"
unset CARTULARY_WEB_E2E_BACKEND_PORT
unset CARTULARY_WEB_E2E_FRONTEND_PORT
resolve_owned_stack_ports
if (( FRONTEND_PORT < 39100 || FRONTEND_PORT > 39199 )); then
  fail "active stateful test-service browser frontend port must stay inside its SeaweedFS CORS range, got ${FRONTEND_PORT}"
fi
unset CARTULARY_TEST_TARGET

out_of_range_frontend_stderr="$tmp_dir/out-of-range-frontend.stderr"
read -r out_of_range_frontend_start out_of_range_frontend_end < <(service_frontend_port_window)
CARTULARY_WEB_E2E_FRONTEND_PORT=39200
if resolve_owned_stack_ports 2>"$out_of_range_frontend_stderr"; then
  fail "active test-service browser frontend port outside the CORS range must fail"
fi
assert_file_contains "$out_of_range_frontend_stderr" "service-backed browser $(browser_stage_name) CORS range ${out_of_range_frontend_start}-${out_of_range_frontend_end}" "active test-service frontend port range error"
unset CARTULARY_WEB_E2E_FRONTEND_PORT

TEST_SERVICES_BIN="$tmp_dir/missing-test-services"
missing_test_services_stderr="$tmp_dir/missing-test-services.stderr"
if browser_start_services 2>"$missing_test_services_stderr"; then
  fail "active test-service mode must require an executable CARTULARY_TEST_SERVICES_BIN"
fi
if ! grep -Fq "CARTULARY_TEST_SERVICES_BIN" "$missing_test_services_stderr"; then
  fail "missing test-services binary failure must mention CARTULARY_TEST_SERVICES_BIN"
fi

fake_test_services="$tmp_dir/fake-test-services.sh"
cat >"$fake_test_services" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

log_file="${FAKE_TEST_SERVICES_LOG:?}"
printf '%s\n' "$*" >>"$log_file"

case "${1:-}" in
  prepare-web-e2e)
    env_file=""
    metadata_file=""
    shift
    while [[ "$#" -gt 0 ]]; do
      case "$1" in
        --env-file)
          env_file="$2"
          shift 2
          ;;
        --metadata-file)
          metadata_file="$2"
          shift 2
          ;;
        *)
          echo "unexpected prepare arg $1" >&2
          exit 2
          ;;
      esac
    done
    cat >"$env_file" <<ENV
export CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN='postgres://cartulary:cartulary@127.0.0.1:15432/ct_web?sslmode=disable'
export CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT='127.0.0.1:19000'
export CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID='web-access'
export CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY='web-secret'
export CARTULARY_S3_OBJECT_PRIMARY_SECURE='false'
export CARTULARY_S3_OBJECT_PRIMARY_BUCKET='ct-web'
ENV
    printf '{"database_name":"ct_web","bucket":"ct-web"}\n' >"$metadata_file"
    ;;
  cleanup-web-e2e)
    ;;
  *)
    echo "unexpected fake test-services command ${1:-}" >&2
    exit 2
    ;;
esac
EOF
chmod +x "$fake_test_services"

fake_test_services_log="$tmp_dir/fake-test-services.log"
TEST_SERVICES_BIN="$fake_test_services"
TEST_SERVICES_ENV_FILE="$tmp_dir/browser.env"
TEST_SERVICES_METADATA_FILE="$tmp_dir/browser.json"
FAKE_TEST_SERVICES_LOG="$fake_test_services_log"
export FAKE_TEST_SERVICES_LOG
browser_start_services >/dev/null
missing_template_stderr="$tmp_dir/missing-template.stderr"
if browser_prepare_database 2>"$missing_template_stderr"; then
  fail "active test-service browser prepare must require CARTULARY_PGTEST_TEMPLATE_DB"
fi
if ! grep -Fq "CARTULARY_PGTEST_TEMPLATE_DB" "$missing_template_stderr"; then
  fail "missing template database failure must mention CARTULARY_PGTEST_TEMPLATE_DB"
fi
CARTULARY_PGTEST_TEMPLATE_DB="suite_template"
browser_prepare_database
assert_equals "$E2E_DSN" "postgres://cartulary:cartulary@127.0.0.1:15432/ct_web?sslmode=disable" "active test-service browser dsn"
assert_equals "$CARTULARY_S3_OBJECT_PRIMARY_BUCKET" "ct-web" "active test-service browser bucket"
assert_equals "$(head -n 1 "$fake_test_services_log")" "prepare-web-e2e --env-file $TEST_SERVICES_ENV_FILE --metadata-file $TEST_SERVICES_METADATA_FILE" "active test-service prepare command"

cleanup_done=0
SERVER_PGID=""
VITE_PGID=""
CHILD_PGID=""
KEEP_RUNTIME_ROOT=1
cleanup
if ! tail -n 1 "$fake_test_services_log" | grep -Fq "cleanup-web-e2e --metadata-file $TEST_SERVICES_METADATA_FILE"; then
  fail "active test-service cleanup command must use browser metadata"
fi

unset CARTULARY_TEST_SERVICES_ACTIVE
unset CARTULARY_PGTEST_TEMPLATE_DB
unset FAKE_TEST_SERVICES_LOG
