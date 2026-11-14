#!/usr/bin/env bash
set -euo pipefail

export CARTULARY_POSTGRES_POSTGRES_PRIMARY_MIGRATION_DSN='postgres://cartulary_migration_login@localhost:5432/cartulary?sslmode=verify-full&require_auth=none&sslrootcert=/fixture/ca.pem&sslcert=/fixture/migration.crt&sslkey=/fixture/migration.key'
export OBJECT_STORE_CORS_PROXY_TLS_ROOT_CERTIFICATE=/fixture/ca.pem
export SEAWEEDFS_S3_ACCESS_KEY_ID=fixture-access
export SEAWEEDFS_S3_SECRET_ACCESS_KEY=fixture-secret

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../../../.." && pwd)"
# shellcheck source=tools/harness/readiness/process-lifecycle.sh
source "$repo_root/tools/harness/readiness/process-lifecycle.sh"
# shellcheck source=tools/harness/test-support/harness-scratch.sh
source "$repo_root/tools/harness/test-support/harness-scratch.sh"
tmp_dir="$(cartulary_harness_mktemp_dir cartulary-dev-services.XXXXXX)"
fake_bin="$tmp_dir/bin"
docker_log="$tmp_dir/docker.log"
go_log="$tmp_dir/go.log"
run_stdout=""
run_stderr=""
run_status=0
command_handles=()

cleanup() {
  local handle status=0 primary=$?
  for handle in "${command_handles[@]}"; do
    stop_owned_command "$handle" || status=$?
  done
  if [[ "$status" -ne 0 ]]; then
    echo "development service fixture cleanup failed; runtime retained" >&2
    if [[ "$primary" -ne 0 ]]; then exit "$primary"; fi
    exit "$status"
  fi
  rm -rf "$tmp_dir"
}
trap cleanup EXIT
lifecycle_install_signal_traps

fail() {
  printf 'test-dev-services-lifecycle: %s\n' "$*" >&2
  if [[ -n "$run_stdout" && -f "$run_stdout" ]]; then
    printf '%s\n' '--- stdout ---' >&2
    cat "$run_stdout" >&2
  fi
  if [[ -n "$run_stderr" && -f "$run_stderr" ]]; then
    printf '%s\n' '--- stderr ---' >&2
    cat "$run_stderr" >&2
  fi
  if [[ -f "$docker_log" ]]; then
    printf '%s\n' '--- docker log ---' >&2
    cat "$docker_log" >&2
  fi
  if [[ -f "$go_log" ]]; then
    printf '%s\n' '--- go log ---' >&2
    cat "$go_log" >&2
  fi
  exit 1
}

mkdir -p "$fake_bin"
: >"$docker_log"
: >"$go_log"

cat >"$fake_bin/docker" <<'FAKE_DOCKER'
#!/usr/bin/env bash
set -euo pipefail

printf '%s\n' "$*" >>"${FAKE_DOCKER_LOG:?}"

if [[ "${1:-}" == "inspect" ]]; then
  printf '%s\n' "running healthy"
  exit 0
fi

if [[ "${1:-}" == "compose" ]]; then
  shift
  if [[ "${1:-}" == "-f" ]]; then
    shift 2
  fi
  case "${1:-}" in
    ps)
      if [[ "${2:-}" == "-q" ]]; then
        printf 'fake-%s\n' "${3:-container}"
      fi
      ;;
    up | down | exec | logs)
      ;;
  esac
fi
FAKE_DOCKER

cat >"$fake_bin/go" <<'FAKE_GO'
#!/usr/bin/env bash
set -euo pipefail

printf '%s\n' "$*" >>"${FAKE_GO_LOG:?}"
printf 'config=%s\n' "${CARTULARY_CONFIG_FILE:-}" >>"${FAKE_GO_LOG:?}"
printf 'managed_dsn=%s\n' "${CARTULARY_POSTGRES_POSTGRES_PRIMARY_MIGRATION_DSN:-}" >>"${FAKE_GO_LOG:?}"
if [[ "${1:-}" == "build" ]]; then
  output=""
  shift
  while [[ "$#" -gt 0 ]]; do
    case "$1" in
      -o)
        output="${2:-}"
        shift 2
        ;;
      *)
        shift
        ;;
    esac
  done
  if [[ -n "$output" ]]; then
    mkdir -p "$(dirname "$output")"
    cat >"$output" <<'FAKE_PROXY'
#!/usr/bin/env bash
set -euo pipefail

command="${1:-}"
shift || true
attempt_file=""
lease_file=""
state_file=""
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --attempt-file)
      attempt_file="$2"
      shift 2
      ;;
    --lease-file)
      lease_file="$2"
      shift 2
      ;;
    --state-file)
      state_file="$2"
      shift 2
      ;;
    *)
      shift
      ;;
  esac
done

case "$command" in
  attempt)
    printf 'launching\n' >"$attempt_file"
    ;;
  serve)
    # This adapter tests orchestration, not physical proxy ownership. Publishing
    # state completes the fake service; no persistent process or PID is needed.
    printf 'bound\n' >"$attempt_file"
    ;;
  status)
    [[ -f "$state_file" && "$(awk '{print $1}' "$state_file")" == "bound" ]]
    ;;
  promote)
    mv "$attempt_file" "$lease_file"
    ;;
  stop | discard)
    rm -f "$state_file"
    ;;
  *)
    exit 2
    ;;
esac
FAKE_PROXY
    chmod +x "$output"
  fi
fi
FAKE_GO

chmod +x "$fake_bin/docker" "$fake_bin/go"

reset_logs() {
  : >"$docker_log"
  : >"$go_log"
}

run_service() {
  local name="$1"
  shift

  run_stdout="$tmp_dir/${name}.stdout"
  run_stderr="$tmp_dir/${name}.stderr"
  set +e
  (
    cd "$repo_root"
    PATH="$fake_bin:$PATH" \
      GO=go \
      FAKE_DOCKER_LOG="$docker_log" \
      FAKE_GO_LOG="$go_log" \
      CARTULARY_RUNTIME_DIR="$tmp_dir/runtime" \
      SEAWEEDFS_S3_PORT=48333 \
      OBJECT_STORE_ENDPOINT=127.0.0.1:48333 \
      OBJECT_STORE_CORS_PROXY_LISTEN=127.0.0.1:48333 \
      CARTULARY_POSTGRES_READY_TIMEOUT_SECONDS=5 \
      CARTULARY_OBJECT_STORE_READY_TIMEOUT_SECONDS=5 \
      "$@"
  ) >"$run_stdout" 2>"$run_stderr"
  run_status=$?
  set -e
}

assert_status() {
  local expected="$1"

  if [[ "$run_status" != "$expected" ]]; then
    fail "expected status $expected, got $run_status"
  fi
}

assert_file_contains() {
  local file="$1"
  local needle="$2"
  local label="$3"

  if ! grep -Fq -- "$needle" "$file"; then
    fail "$label: expected to find [$needle] in $file"
  fi
}

assert_file_not_contains() {
  local file="$1"
  local needle="$2"
  local label="$3"

  if grep -Fq -- "$needle" "$file"; then
    fail "$label: did not expect to find [$needle] in $file"
  fi
}

assert_log_empty() {
  local file="$1"
  local label="$2"

  if [[ -s "$file" ]]; then
    fail "$label: expected empty log"
  fi
}

reset_logs
run_service services_down_dry_run env CARTULARY_CLEANUP_DRY_RUN=1 bash tools/harness/readiness/dev-services.sh services-down
assert_status 0
assert_file_contains "$run_stdout" "DRY-RUN stop-services compose:" "services-down dry-run"
assert_log_empty "$docker_log" "services-down dry-run docker"
assert_log_empty "$go_log" "services-down dry-run go"

reset_logs
mkdir -p "$tmp_dir/runtime"
printf '%s\n' "$$" >"$tmp_dir/runtime/seaweedfs-s3-cors-proxy.pid"
run_service services_down_real bash tools/harness/readiness/dev-services.sh services-down
assert_status 0
assert_file_contains "$docker_log" "down --remove-orphans" "services-down command"
assert_file_not_contains "$docker_log" "--volumes" "services-down volume preservation"
assert_log_empty "$go_log" "services-down real go"
kill -0 "$$" || fail "legacy PID metadata must never authorize signaling"
[[ ! -e "$tmp_dir/runtime/seaweedfs-s3-cors-proxy.pid" ]] ||
  fail "legacy PID metadata must be discarded"

reset_logs
run_service db_down_removed bash tools/harness/readiness/dev-services.sh db-down
assert_status 2
assert_file_contains "$run_stderr" "usage: dev-services.sh" "db-down removed usage"
assert_log_empty "$docker_log" "db-down removed docker"
assert_log_empty "$go_log" "db-down removed go"

reset_logs
run_service db_reset_dry_run env CARTULARY_CLEANUP_DRY_RUN=1 bash tools/harness/readiness/dev-services.sh db-reset
assert_status 0
assert_file_contains "$run_stdout" "DRY-RUN reset-database postgres:cartulary" "db-reset dry-run"
assert_log_empty "$docker_log" "db-reset dry-run docker"
assert_log_empty "$go_log" "db-reset dry-run go"

reset_logs
run_service db_reset_missing_confirm bash tools/harness/readiness/dev-services.sh db-reset
assert_status 2
assert_file_contains "$run_stderr" "refusing db-reset" "db-reset missing confirmation"
assert_log_empty "$docker_log" "db-reset missing confirmation docker"
assert_log_empty "$go_log" "db-reset missing confirmation go"

reset_logs
run_service db_migrate bash tools/harness/readiness/dev-services.sh db-migrate
assert_status 0
assert_file_contains "$run_stdout" "db-migrate: applying local database migrations only" "db-migrate stdout"
assert_file_contains "$docker_log" "up -d postgres" "db-migrate starts postgres"
assert_file_not_contains "$docker_log" "DROP DATABASE IF EXISTS cartulary;" "db-migrate does not drop database"
assert_file_not_contains "$docker_log" "CREATE DATABASE cartulary;" "db-migrate does not create database"
assert_file_not_contains "$docker_log" "seaweedfs-s3" "db-migrate does not start object store"
assert_file_contains "$go_log" "run ./cmd/migrate up" "db-migrate runs migrations"
assert_file_contains "$go_log" "config=$repo_root/configs/dev/config.toml" "db-migrate passes default config"
assert_file_contains "$go_log" "managed_dsn=$CARTULARY_POSTGRES_POSTGRES_PRIMARY_MIGRATION_DSN" "db-migrate passes the explicit purpose binding"
assert_file_not_contains "$docker_log" "provision.sh" "ordinary migration does not reprovision retained state"

reset_logs
run_service db_migrate_missing_dsn env -u CARTULARY_POSTGRES_POSTGRES_PRIMARY_MIGRATION_DSN bash tools/harness/readiness/dev-services.sh db-migrate
[[ "$run_status" != 0 ]] || fail "missing migration identity accepted"
assert_log_empty "$docker_log" "missing migration identity fails before service acquisition"
assert_log_empty "$go_log" "missing migration identity fails before migration"

reset_logs
run_service db_reset_confirmed env CARTULARY_DESTRUCTIVE_CONFIRM=db-reset bash tools/harness/readiness/dev-services.sh db-reset
assert_status 0
assert_file_contains "$docker_log" "up -d postgres" "db-reset starts postgres"
assert_file_contains "$docker_log" "DROP DATABASE IF EXISTS cartulary;" "db-reset drops database"
assert_file_contains "$docker_log" "CREATE DATABASE cartulary;" "db-reset creates database"
assert_file_contains "$go_log" "run ./cmd/migrate up" "db-reset runs migrations"

reset_logs
run_service object_store_reset_dry_run env CARTULARY_CLEANUP_DRY_RUN=1 bash tools/harness/readiness/dev-services.sh object-store-reset
assert_status 0
assert_file_contains "$run_stdout" "DRY-RUN reset-object-store object-store-bucket:cartulary" "object-store-reset dry-run"
assert_log_empty "$docker_log" "object-store-reset dry-run docker"
assert_log_empty "$go_log" "object-store-reset dry-run go"

reset_logs
run_service object_store_reset_missing_confirm bash tools/harness/readiness/dev-services.sh object-store-reset
assert_status 2
assert_file_contains "$run_stderr" "refusing object-store-reset" "object-store-reset missing confirmation"
assert_log_empty "$docker_log" "object-store-reset missing confirmation docker"
assert_log_empty "$go_log" "object-store-reset missing confirmation go"

reset_logs
run_service object_store_reset_confirmed env CARTULARY_DESTRUCTIVE_CONFIRM=object-store-reset OBJECT_STORE_BUCKET=ct-test bash tools/harness/readiness/dev-services.sh object-store-reset
assert_status 0
assert_file_contains "$docker_log" "up -d --remove-orphans seaweedfs-s3" "object-store-reset starts seaweedfs-s3"
assert_file_contains "$go_log" "build -o" "object-store-reset builds CORS proxy"
assert_file_contains "$go_log" "./tools/s3corsproxy" "object-store-reset builds proxy helper"
assert_file_contains "$go_log" "run ./tools/objectstoreprobe" "object-store-reset runs object-store probe"
assert_file_contains "$go_log" "--mode reset" "object-store-reset selects reset mode"
assert_file_contains "$go_log" "--bucket ct-test" "object-store-reset passes configured bucket"

# Repeated startup reuses the owned proxy and performs exactly one readiness pass.
reset_logs
for attempt in 1 2; do
  run_service "services_up_${attempt}" bash tools/harness/readiness/dev-services.sh up
  assert_status 0
done
[[ "$(grep -c 'up -d --remove-orphans postgres seaweedfs-s3$' "$docker_log")" == 2 ]] || fail "one Compose startup per invocation"
[[ "$(grep -c 'pg_isready' "$docker_log")" == 2 ]] || fail "one Postgres readiness pass per invocation"
[[ "$(grep -c -- '--mode probe' "$go_log")" == 2 ]] || fail "one object-store readiness pass per invocation"
assert_file_not_contains "$go_log" "build -o" "repeated startup reuses the owned proxy"
assert_file_not_contains "$go_log" "cmd/migrate" "startup does not migrate"
assert_file_not_contains "$go_log" "--mode reset" "startup does not reset object storage"
assert_file_not_contains "$docker_log" "DROP DATABASE" "startup does not reset the database"
assert_file_not_contains "$docker_log" "provision.sh" "startup does not reprovision retained state"

reset_logs
concurrent_runtime="$tmp_dir/concurrent-runtime"
for index in 1 2; do
  concurrent_handle=""
  acquisition_status=0
  # The positional parameters belong to the owned child shell, not this caller.
  # shellcheck disable=SC2016
  start_owned_command concurrent_handle "$tmp_dir/concurrent-${index}.log" env \
      PATH="$fake_bin:$PATH" \
      GO=go \
      FAKE_DOCKER_LOG="$docker_log" \
      FAKE_GO_LOG="$go_log" \
      CARTULARY_RUNTIME_DIR="$concurrent_runtime" \
      SEAWEEDFS_S3_PORT=48334 \
      OBJECT_STORE_ENDPOINT=127.0.0.1:48334 \
      OBJECT_STORE_CORS_PROXY_LISTEN=127.0.0.1:48334 \
      CARTULARY_POSTGRES_READY_TIMEOUT_SECONDS=5 \
      CARTULARY_OBJECT_STORE_READY_TIMEOUT_SECONDS=5 \
      CARTULARY_DESTRUCTIVE_CONFIRM=object-store-reset \
      bash -c '
        cd "$1"
        : >"$2"
        while [[ ! -f "$3" ]]; do sleep 0.02; done
        exec bash tools/harness/readiness/dev-services.sh object-store-reset
      ' fixture "$repo_root" "$tmp_dir/concurrent-${index}.ready" "$tmp_dir/concurrent.release" || acquisition_status=$?
  if [[ -n "$concurrent_handle" ]]; then command_handles+=("$concurrent_handle"); fi
  [[ "$acquisition_status" -eq 0 ]] || fail "concurrent caller acquisition failed"
done
concurrent_deadline=$((SECONDS + 10))
while [[ ! -f "$tmp_dir/concurrent-1.ready" || ! -f "$tmp_dir/concurrent-2.ready" ]]; do
  (( SECONDS < concurrent_deadline )) || fail "concurrent callers did not reach the release barrier"
  sleep 0.02
done
: >"$tmp_dir/concurrent.release"
concurrent_status=0
for concurrent_handle in "${command_handles[@]}"; do
  while owned_command_running "$concurrent_handle"; do
    (( SECONDS < concurrent_deadline )) || fail "concurrent lifecycle command did not finish"
    sleep 0.02
  done
  stop_owned_command "$concurrent_handle" || fail "concurrent caller settlement failed"
  if ! owned_command_result "$concurrent_handle"; then
    concurrent_status=1
  fi
done
[[ "$concurrent_status" -eq 0 ]] || fail "concurrent proxy lifecycle operation failed"
build_count="$(grep -c '^build -o ' "$go_log" || true)"
[[ "$build_count" == "1" ]] ||
  fail "concurrent proxy starts must publish exactly one owned instance, got $build_count builds"
[[ -f "$concurrent_runtime/object-store-proxy/ready-lease.json" ]] ||
  fail "concurrent proxy starts must publish one ready lease"
