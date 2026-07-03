#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")/.." && pwd)"
DEV_STACK_SCRIPT="$ROOT_DIR/tools/harness/readiness/dev-stack.sh"
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

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" != *"$needle"* ]]; then
    fail "$label: expected output to contain [$needle]"
  fi
}

assert_not_contains() {
  local haystack="$1"
  local needle="$2"
  local label="$3"

  if [[ "$haystack" == *"$needle"* ]]; then
    fail "$label: output must not contain [$needle]"
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

  for _ in $(seq 1 100); do
    if [[ -e "$path" ]]; then
      return 0
    fi
    sleep 0.1
  done

  fail "$label: expected path $path to appear"
}

count_lines() {
  local path="$1"

  if [[ ! -f "$path" ]]; then
    printf '%s\n' 0
    return
  fi

  wc -l <"$path" | tr -d ' '
}

run_dev_stack_case() {
  local case_dir="$1"
  shift

  CARTULARY_DEV_STACK_ARTIFACT_DIR="$case_dir/runtime" \
  CARTULARY_DEV_STACK_SKIP_SERVICE_PREFLIGHT=1 \
  CARTULARY_DEV_STACK_SKIP_READINESS=1 \
    "$DEV_STACK_SCRIPT" "$@" >"$case_dir/stdout" 2>"$case_dir/stderr"
}

make_command() {
  local recorder="$1"
  local pid_file="$2"
  local term_file="$3"
  local mode="${4:-loop}"
  local exit_after_seconds="${5:-0.1}"
  local exit_status="${6:-0}"

  printf 'PID_FILE=%q TERM_FILE=%q MODE=%q EXIT_AFTER_SECONDS=%q EXIT_STATUS=%q %q' \
    "$pid_file" \
    "$term_file" \
    "$mode" \
    "$exit_after_seconds" \
    "$exit_status" \
    "$recorder"
}

mkdir -p "$ROOT_DIR/tmp"
tmp_dir="$(mktemp -d "$ROOT_DIR/tmp/dev-stack-lifecycle-smoke.XXXXXX")"
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
env_file="${ENV_FILE:-}"

printf '%s\n' "$$" >"$pid_file"
if [[ -n "$env_file" ]]; then
  env | sort >"$env_file"
fi

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

preflight_dir="$tmp_dir/preflight"
mkdir -p "$preflight_dir"
if CARTULARY_DEV_STACK_ARTIFACT_DIR="$preflight_dir/runtime" \
  CARTULARY_DEV_STACK_POSTGRES_PORT=1 \
  CARTULARY_DEV_STACK_OBJECT_STORE_PORT=1 \
  CARTULARY_DEV_STACK_BACKEND_COMMAND="printf backend-started >'$preflight_dir/backend.started'" \
  CARTULARY_DEV_STACK_FRONTEND_COMMAND="printf frontend-started >'$preflight_dir/frontend.started'" \
  "$DEV_STACK_SCRIPT" >"$preflight_dir/stdout" 2>"$preflight_dir/stderr"; then
  preflight_status=0
else
  preflight_status=$?
fi
assert_equals "$preflight_status" "1" "service preflight status"
assert_contains "$(cat "$preflight_dir/stderr")" "run make db-up or make services-up before make dev" "service preflight guidance"
if [[ -e "$preflight_dir/backend.started" || -e "$preflight_dir/frontend.started" ]]; then
  fail "service preflight must fail before starting child processes"
fi

term_dir="$tmp_dir/term"
mkdir -p "$term_dir"
backend_command="$(make_command "$signal_recorder" "$term_dir/backend.pid" "$term_dir/backend.term")"
frontend_command="$(make_command "$signal_recorder" "$term_dir/frontend.pid" "$term_dir/frontend.term")"
CARTULARY_DEV_STACK_BACKEND_COMMAND="$backend_command" \
CARTULARY_DEV_STACK_FRONTEND_COMMAND="$frontend_command" \
CARTULARY_DEV_STACK_ARTIFACT_DIR="$term_dir/runtime" \
CARTULARY_DEV_STACK_SKIP_SERVICE_PREFLIGHT=1 \
CARTULARY_DEV_STACK_SKIP_READINESS=1 \
  "$DEV_STACK_SCRIPT" >"$term_dir/stdout" 2>"$term_dir/stderr" &
term_stack_pid=$!
wait_for_path "$term_dir/backend.pid" "term backend pid"
wait_for_path "$term_dir/frontend.pid" "term frontend pid"
kill -TERM "$term_stack_pid"
if wait "$term_stack_pid"; then
  term_status=0
else
  term_status=$?
fi
assert_equals "$term_status" "143" "term shutdown status"
assert_equals "$(count_lines "$term_dir/backend.term")" "1" "term backend term count"
assert_equals "$(count_lines "$term_dir/frontend.term")" "1" "term frontend term count"
assert_process_stopped "$(tr -d '\n' <"$term_dir/backend.pid")" "term backend process"
assert_process_stopped "$(tr -d '\n' <"$term_dir/frontend.pid")" "term frontend process"
assert_contains "$(cat "$term_dir/stdout")" "backend log: $term_dir/runtime/server.log" "term backend log output"
assert_contains "$(cat "$term_dir/stdout")" "frontend log: $term_dir/runtime/web.log" "term frontend log output"

backend_exit_dir="$tmp_dir/backend-exits"
mkdir -p "$backend_exit_dir"
backend_command="$(make_command "$signal_recorder" "$backend_exit_dir/backend.pid" "$backend_exit_dir/backend.term" exit_after 0.2 0)"
frontend_command="$(make_command "$signal_recorder" "$backend_exit_dir/frontend.pid" "$backend_exit_dir/frontend.term")"
if CARTULARY_DEV_STACK_BACKEND_COMMAND="$backend_command" \
  CARTULARY_DEV_STACK_FRONTEND_COMMAND="$frontend_command" \
  run_dev_stack_case "$backend_exit_dir"; then
  backend_exit_status=0
else
  backend_exit_status=$?
fi
assert_equals "$backend_exit_status" "1" "backend-exits status"
assert_equals "$(count_lines "$backend_exit_dir/frontend.term")" "1" "backend-exits frontend term count"
assert_process_stopped "$(tr -d '\n' <"$backend_exit_dir/frontend.pid")" "backend-exits frontend process"

frontend_exit_dir="$tmp_dir/frontend-exits"
mkdir -p "$frontend_exit_dir"
backend_command="$(make_command "$signal_recorder" "$frontend_exit_dir/backend.pid" "$frontend_exit_dir/backend.term")"
frontend_command="$(make_command "$signal_recorder" "$frontend_exit_dir/frontend.pid" "$frontend_exit_dir/frontend.term" exit_after 0.2 0)"
if CARTULARY_DEV_STACK_BACKEND_COMMAND="$backend_command" \
  CARTULARY_DEV_STACK_FRONTEND_COMMAND="$frontend_command" \
  run_dev_stack_case "$frontend_exit_dir"; then
  frontend_exit_status=0
else
  frontend_exit_status=$?
fi
assert_equals "$frontend_exit_status" "1" "frontend-exits status"
assert_equals "$(count_lines "$frontend_exit_dir/backend.term")" "1" "frontend-exits backend term count"
assert_process_stopped "$(tr -d '\n' <"$frontend_exit_dir/backend.pid")" "frontend-exits backend process"

port_conflict_dir="$tmp_dir/port-conflict"
mkdir -p "$port_conflict_dir/bin"
cat >"$port_conflict_dir/bin/ss" <<'EOF'
#!/usr/bin/env bash
printf 'State  Recv-Q Send-Q Local Address:Port Peer Address:Port Process\n'
printf 'LISTEN 0      4096       127.0.0.1:8080      0.0.0.0:*     users:(("fake-owner",pid=123,fd=4))\n'
EOF
chmod +x "$port_conflict_dir/bin/ss"
backend_command="$(make_command "$signal_recorder" "$port_conflict_dir/backend.pid" "$port_conflict_dir/backend.term")"
frontend_command="$(make_command "$signal_recorder" "$port_conflict_dir/frontend.pid" "$port_conflict_dir/frontend.term")"
if PATH="$port_conflict_dir/bin:$PATH" \
  CARTULARY_DEV_STACK_BACKEND_COMMAND="$backend_command" \
  CARTULARY_DEV_STACK_FRONTEND_COMMAND="$frontend_command" \
  run_dev_stack_case "$port_conflict_dir"; then
  port_conflict_status=0
else
  port_conflict_status=$?
fi
assert_equals "$port_conflict_status" "1" "port-conflict status"
assert_contains "$(cat "$port_conflict_dir/stderr")" "backend port 8080 is already in use" "port-conflict message"
assert_contains "$(cat "$port_conflict_dir/stderr")" "fake-owner" "port-conflict owner"
if [[ -e "$port_conflict_dir/backend.pid" || -e "$port_conflict_dir/frontend.pid" ]]; then
  fail "port-conflict must fail before starting child processes"
fi

env_dir="$tmp_dir/env"
mkdir -p "$env_dir"
backend_command="$(printf 'PID_FILE=%q TERM_FILE=%q ENV_FILE=%q MODE=exit_after EXIT_AFTER_SECONDS=0.2 EXIT_STATUS=0 %q' \
  "$env_dir/backend.pid" \
  "$env_dir/backend.term" \
  "$env_dir/backend.env" \
  "$signal_recorder")"
frontend_command="$(make_command "$signal_recorder" "$env_dir/frontend.pid" "$env_dir/frontend.term")"
if CARTULARY_DEV_STACK_BACKEND_COMMAND="$backend_command" \
  CARTULARY_DEV_STACK_FRONTEND_COMMAND="$frontend_command" \
  run_dev_stack_case "$env_dir"; then
  env_status=0
else
  env_status=$?
fi
assert_equals "$env_status" "1" "env status"
backend_env="$(cat "$env_dir/backend.env")"
assert_contains "$backend_env" "CARTULARY_CONFIG_FILE=$ROOT_DIR/configs/dev/config.toml" "env dev config"
assert_contains "$backend_env" "CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH=$ROOT_DIR/configs/dev/bootstrap-admin.json" "env bootstrap manifest"
assert_contains "$backend_env" "CARTULARY_POSTGRES_POSTGRES_PRIMARY_DSN=postgres://cartulary:cartulary@localhost:5432/cartulary?sslmode=disable" "env managed postgres dsn"
assert_contains "$backend_env" "CARTULARY_S3_OBJECT_PRIMARY_BUCKET=cartulary" "env managed object bucket"
assert_not_contains "$backend_env" "CARTULARY_POSTGRES_DSN=" "env e2e dsn"
assert_not_contains "$backend_env" "CARTULARY_ENABLE_TEST_ROUTES=" "env test routes"
assert_not_contains "$backend_env" "CARTULARY_PLAYWRIGHT_STATE_DIR=" "env playwright state"
assert_not_contains "$backend_env" "CARTULARY_WEB_E2E_RUNTIME_ROOT=" "env e2e runtime root"
assert_not_contains "$backend_env" "CARTULARY__ROOTS__DATABASE_STORAGE__PATH=" "env e2e storage root"
