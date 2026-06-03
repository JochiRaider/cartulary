#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/cartulary-dev-services.XXXXXX")"
fake_bin="$tmp_dir/bin"
docker_log="$tmp_dir/docker.log"
go_log="$tmp_dir/go.log"
run_stdout=""
run_stderr=""
run_status=0

cleanup() {
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

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
      FAKE_DOCKER_LOG="$docker_log" \
      FAKE_GO_LOG="$go_log" \
      CARTULARY_POSTGRES_READY_TIMEOUT_SECONDS=1 \
      CARTULARY_OBJECT_STORE_READY_TIMEOUT_SECONDS=1 \
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
run_service services_down_dry_run env CARTULARY_CLEANUP_DRY_RUN=1 bash scripts/dev-services.sh services-down
assert_status 0
assert_file_contains "$run_stdout" "DRY-RUN stop-services compose:" "services-down dry-run"
assert_log_empty "$docker_log" "services-down dry-run docker"
assert_log_empty "$go_log" "services-down dry-run go"

reset_logs
run_service services_down_real bash scripts/dev-services.sh services-down
assert_status 0
assert_file_contains "$docker_log" "down --remove-orphans" "services-down command"
assert_file_not_contains "$docker_log" "--volumes" "services-down volume preservation"
assert_log_empty "$go_log" "services-down real go"

reset_logs
run_service db_down_dry_run env CARTULARY_CLEANUP_DRY_RUN=1 bash scripts/dev-services.sh db-down
assert_status 0
assert_file_contains "$run_stdout" "DRY-RUN stop-services compose:" "db-down dry-run alias"
assert_log_empty "$docker_log" "db-down dry-run docker"
assert_log_empty "$go_log" "db-down dry-run go"

reset_logs
run_service db_down_real bash scripts/dev-services.sh db-down
assert_status 0
assert_file_contains "$docker_log" "down --remove-orphans" "db-down command alias"
assert_file_not_contains "$docker_log" "--volumes" "db-down volume preservation"
assert_log_empty "$go_log" "db-down real go"

reset_logs
run_service db_reset_dry_run env CARTULARY_CLEANUP_DRY_RUN=1 bash scripts/dev-services.sh db-reset
assert_status 0
assert_file_contains "$run_stdout" "DRY-RUN reset-database postgres:cartulary" "db-reset dry-run"
assert_log_empty "$docker_log" "db-reset dry-run docker"
assert_log_empty "$go_log" "db-reset dry-run go"

reset_logs
run_service db_reset_missing_confirm bash scripts/dev-services.sh db-reset
assert_status 2
assert_file_contains "$run_stderr" "refusing db-reset" "db-reset missing confirmation"
assert_log_empty "$docker_log" "db-reset missing confirmation docker"
assert_log_empty "$go_log" "db-reset missing confirmation go"

reset_logs
run_service db_reset_confirmed env CARTULARY_DESTRUCTIVE_CONFIRM=db-reset bash scripts/dev-services.sh db-reset
assert_status 0
assert_file_contains "$docker_log" "up -d postgres" "db-reset starts postgres"
assert_file_contains "$docker_log" "DROP DATABASE IF EXISTS cartulary;" "db-reset drops database"
assert_file_contains "$docker_log" "CREATE DATABASE cartulary;" "db-reset creates database"
assert_file_contains "$go_log" "run ./cmd/migrate up" "db-reset runs migrations"

reset_logs
run_service object_store_reset_dry_run env CARTULARY_CLEANUP_DRY_RUN=1 bash scripts/dev-services.sh object-store-reset
assert_status 0
assert_file_contains "$run_stdout" "DRY-RUN reset-object-store object-store-bucket:cartulary" "object-store-reset dry-run"
assert_log_empty "$docker_log" "object-store-reset dry-run docker"
assert_log_empty "$go_log" "object-store-reset dry-run go"

reset_logs
run_service object_store_reset_missing_confirm bash scripts/dev-services.sh object-store-reset
assert_status 2
assert_file_contains "$run_stderr" "refusing object-store-reset" "object-store-reset missing confirmation"
assert_log_empty "$docker_log" "object-store-reset missing confirmation docker"
assert_log_empty "$go_log" "object-store-reset missing confirmation go"

reset_logs
run_service object_store_reset_confirmed env CARTULARY_DESTRUCTIVE_CONFIRM=object-store-reset OBJECT_STORE_BUCKET=ct-test bash scripts/dev-services.sh object-store-reset
assert_status 0
assert_file_contains "$docker_log" "up -d --remove-orphans seaweedfs-s3" "object-store-reset starts seaweedfs-s3"
assert_file_contains "$go_log" "run ./tools/objectstoreprobe" "object-store-reset runs object-store probe"
assert_file_contains "$go_log" "--mode reset" "object-store-reset selects reset mode"
assert_file_contains "$go_log" "--bucket ct-test" "object-store-reset passes configured bucket"
