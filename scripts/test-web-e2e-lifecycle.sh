#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
LIFECYCLE_HELPER="$ROOT_DIR/scripts/lib/web-e2e-lifecycle.sh"
cleanup_paths=()

cleanup() {
  local path
  for path in "${cleanup_paths[@]}"; do
    rm -rf "$path"
  done
}

trap cleanup EXIT

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

mkdir -p "$ROOT_DIR/tmp"
tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/web-e2e-lifecycle-smoke.XXXXXX")"
cleanup_paths+=("$tmp_dir")

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

single_stop_pid_file="$tmp_dir/single-stop.pid"
single_stop_term_file="$tmp_dir/single-stop.term"
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
