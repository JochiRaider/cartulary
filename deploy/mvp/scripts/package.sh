#!/usr/bin/env bash
set -euo pipefail
SCRIPT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")" && pwd -P)"
# shellcheck source=deploy/mvp/scripts/package-common.sh
source "$SCRIPT_DIR/package-common.sh"
operation="${1:-}"
case "$operation" in start|stop|backup-create|backup-export|restore-bundle|restore-verify-due) shift;; *) package_fail "supported operations: start, stop, backup-create, backup-export, restore-bundle, restore-verify-due";; esac
export_parent=
export_name=
bundle_directory=
restore_arguments=()
restore_operation_id=
stale_acknowledgement=
if [[ "$operation" == backup-export ]]; then
  [[ $# == 1 && "$1" == /* && "$1" != */ && "$1" != *'/../'* && "$1" != *'/./'* ]] || package_fail "backup-export requires an absolute output directory"
  export_parent="$(dirname -- "$1")"
  export_name="$(basename -- "$1")"
  [[ "$export_name" != . && "$export_name" != .. && "$export_name" != *'$'* && "$export_name" != *'`'* ]] || package_fail "invalid output directory"
  package_path "$export_parent" "export parent"
elif [[ "$operation" == restore-bundle ]]; then
  [[ $# -ge 2 ]] || package_fail "restore-bundle requires an absolute bundle directory and exact backup UUID"
  bundle_directory="$1"
  package_path "$bundle_directory" "bundle directory"
  [[ "$2" =~ ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]] || package_fail "exact backup UUID is required"
  restore_arguments=(--confirm-backup-set-id "$2")
  shift 2
  while [[ $# -gt 0 ]]; do
    [[ $# -ge 2 && ( "$1" == --acknowledge-stale-backup || "$1" == --operation-id ) && "$2" =~ ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]] || package_fail "unsupported restore argument"
    case "$1" in
      --operation-id)
        [[ -z "$restore_operation_id" ]] || package_fail "duplicate restore operation ID"
        restore_operation_id="$2";;
      --acknowledge-stale-backup)
        [[ -z "$stale_acknowledgement" ]] || package_fail "duplicate stale acknowledgement"
        stale_acknowledgement="$2"
        restore_arguments+=("$1" "$2");;
    esac
    shift 2
  done
else
  [[ $# == 0 ]] || package_fail "unexpected operation argument"
fi
package_load
for required in flock jq curl timeout; do command -v "$required" >/dev/null || package_fail "required package runtime tool is unavailable"; done

# A Docker deployment, not a working-directory spelling, owns the host lock.
# Different OS principals cannot operate the same registry accidentally.
lock_root=/tmp/cartulary-package-operations
if [[ ! -e "$lock_root" ]]; then mkdir -m 0700 -- "$lock_root" 2>/dev/null || true; fi
[[ -d "$lock_root" && ! -L "$lock_root" && "$(stat -c %u -- "$lock_root")" == "$UID" && "$(stat -c %a -- "$lock_root")" == 700 ]] || package_fail "package operation lock registry is unavailable"
daemon_id="$(docker info --format '{{.ID}}')"
[[ -n "$daemon_id" ]] || package_fail "Docker deployment identity is unavailable"
lock_identity="$(printf '%s\n%s\n' "$daemon_id" "$CARTULARY_MVP_COMPOSE_PROJECT_NAME" | sha256sum | cut -d ' ' -f 1)"
exec {package_lock}>"${lock_root}/${lock_identity}.lock"
flock -n "$package_lock" || package_fail "another package operation is active"
pending_file="${lock_root}/${lock_identity}.pending"
if [[ -e "$pending_file" || -L "$pending_file" ]]; then
  package_file "$pending_file"
  pending_id="$(cat "$pending_file")"
  [[ "$pending_id" =~ ^[0-9a-f-]{36}$ ]] || package_fail "interrupted package evidence is invalid"
  [[ -z "$(docker ps -q --filter "label=cartulary.package.operation=${pending_id}")" ]] || package_fail "interrupted operation still owns a running container; inspect it before retrying"
  rm -- "$pending_file"
fi
package_preflight >/dev/null
package_operation_id="$(cat /proc/sys/kernel/random/uuid)"
printf '%s\n' "$package_operation_id" >"$pending_file"
sync -f "$pending_file"

stopped_container=
private_container=
started_container=
starting_public=false
restart_required=false
restart_result=not_required
cleanup_result=succeeded
gate=admitted
cleanup() {
  local status=$?
  local primary_status=$status
  trap - EXIT INT TERM
  if [[ -n "$private_container" ]]; then
    if ! docker rm -f "$private_container" >/dev/null 2>&1; then cleanup_result=failed; status=1; fi
  fi
  if [[ "$operation" == start && "$status" != 0 && "$starting_public" == true ]]; then
    if [[ -z "$started_container" ]]; then started_container="$(package_running)" || cleanup_result=failed; fi
    if [[ -n "$started_container" ]] && ! docker stop "$started_container" >/dev/null 2>&1; then cleanup_result=failed; fi
  fi
  local active_owned
  if ! active_owned="$(docker ps -q --filter "label=cartulary.package.operation=${package_operation_id}")" || [[ -n "$active_owned" ]]; then
    cleanup_result=failed
    restart_result=blocked
    restart_required=false
    status=1
  fi
  if [[ "$restart_required" == true && -n "$stopped_container" ]]; then
    restart_result=failed
    if docker start "$stopped_container" >/dev/null 2>&1 && package_ready_public; then restart_result=succeeded; else status=1; fi
  fi
  if [[ "$cleanup_result" == succeeded ]]; then rm -f -- "$pending_file"; fi
  printf '{"schema_id":"cartulary.package_operation_result.v1","operation":"%s","gate":"%s","operation_exit_code":%d,"exit_code":%d,"restart":"%s","cleanup":"%s"}\n' "$operation" "$gate" "$primary_status" "$status" "$restart_result" "$cleanup_result" >&2
  exit "$status"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

package_running() {
  local id
  id="$(compose ps --status running -q app)"
  [[ -z "$id" || "$id" =~ ^[0-9a-f]{64}$ ]] || package_fail "ambiguous application container identity"
  printf '%s' "$id"
}
package_stop_app() {
  stopped_container="$(package_running)"
  if [[ -n "$stopped_container" ]]; then docker stop "$stopped_container" >/dev/null; fi
}
package_ready_public() {
  # Arguments are intentionally expanded by the bounded child shell.
  # shellcheck disable=SC2016
  timeout 120 bash -c '
    until curl --cacert "$1/ca.pem" --tlsv1.3 --tls-max 1.3 --max-time 5 --silent --fail "https://localhost:$2/readyz" >/dev/null 2>&1; do sleep 1; done
  ' _ "$CARTULARY_TLS_DIR" "$CARTULARY_HTTP_PORT"
}
package_backup() {
  compose run --rm --no-deps --entrypoint /usr/local/bin/cartulary-operator recovery-operator \
    backup create --source-config-file "$SOURCE_CONFIG_CONTAINER"
}

# Package provisioning is private; Recovery owns proof policy and publication.
# shellcheck source=deploy/mvp/scripts/package-verification.sh
source "$SCRIPT_DIR/package-verification.sh"

case "$operation" in
  stop) package_stop_app;;
  backup-create)
    package_stop_app
    restart_required=true
    package_backup;;
  backup-export)
    compose run --rm --no-deps --volume "${export_parent}:/transfer-output" \
      --entrypoint /usr/local/bin/cartulary-operator recovery-operator \
      backup export latest --source-config-file "$SOURCE_CONFIG_CONTAINER" --output-directory "/transfer-output/${export_name}";;
  restore-bundle)
    [[ -z "$(package_running)" ]] || package_fail "restore target application must be stopped"
    gate=dependencies
    compose up -d --wait --force-recreate postgres seaweedfs-s3 >/dev/null
    # Existing targets must be admitted by Recovery without migration writes.
    initialized="$(compose exec -T --user postgres postgres psql -U postgres -d "$POSTGRES_DB" -Atc "SELECT to_regclass('public.goose_db_version') IS NOT NULL")"
    if [[ "$initialized" != t ]]; then compose run --rm --no-deps migrate up >/dev/null; fi
    gate=object_initialization
    package_initialize_objects run --rm --no-deps object-store-init >/dev/null
    compose run --rm --no-deps --volume "${bundle_directory}:/transfer-input:ro" \
      --entrypoint /usr/local/bin/cartulary-operator recovery-operator \
      restore bundle --bundle-directory /transfer-input --target-config-file "$SOURCE_CONFIG_CONTAINER" \
      --operation-id "${restore_operation_id:-$package_operation_id}" "${restore_arguments[@]}"
    echo 'restore completed with application stopped; run package start to establish fresh protection and readiness' >&2;;
  restore-verify-due) package_verify;;
  start)
    package_stop_app
    gate=dependencies
    compose up -d --wait --force-recreate postgres seaweedfs-s3 >/dev/null
    gate=migration
    compose run --rm --no-deps migrate up >/dev/null
    gate=object_initialization
    package_initialize_objects run --rm --no-deps object-store-init >/dev/null
    # Compose run does not publish the service ports. Bootstrap and startup
    # recovery finish privately before the stopped-app baseline is captured.
    gate=private_bootstrap
    private_container="$(compose run -d --no-deps --name "${CARTULARY_MVP_COMPOSE_PROJECT_NAME}-bootstrap" --label cartulary.package.bootstrap=true app)"
    [[ "$private_container" =~ ^[0-9a-f]{64}$ ]] || package_fail "private bootstrap container identity is unavailable"
    # shellcheck disable=SC2016
    timeout 120 bash -c '
      until docker exec "$1" /usr/local/bin/cartulary-operator package readiness >/dev/null 2>&1; do
        [[ "$(docker inspect --format "{{.State.Running}}" "$1")" == true ]] || exit 1
        sleep 1
      done
    ' _ "$private_container"
    docker stop "$private_container" >/dev/null
    docker rm "$private_container" >/dev/null
    private_container=
    gate=backup_freshness
    inspection=
    if ! inspection="$(compose run --rm --no-deps --entrypoint /usr/local/bin/cartulary-operator recovery-operator backup inspect latest)"; then
      jq -e '.error.code == "backup_set_not_found" and .error.reason_code == "no_successful_retained_backup"' <<<"$inspection" >/dev/null || exit 1
      package_backup >/dev/null
    fi
    gate=due_verification
    package_verify >/dev/null
    gate=serving
    starting_public=true
    compose up -d --no-deps --force-recreate app >/dev/null
    started_container="$(package_running)"
    [[ -n "$started_container" ]] || package_fail "application container did not start"
    gate=public_readiness
    package_ready_public
    gate=complete
    echo 'package ready; start the guest maintenance timers for this operating session';;
esac
