#!/usr/bin/env bash
# Package composition only. Application configuration and Recovery stay in Operator.

package_fail() { echo "cartulary package: $1" >&2; exit 2; }

package_file() {
  [[ -f "$1" && ! -L "$1" ]] || package_fail "required regular input is unavailable"
}

package_path() {
  [[ "$1" == /* && "$1" != *'/../'* && "$1" != */.. && "$1" != *'/./'* && "$1" != */. && "$1" != *'$'* && "$1" != *'`'* ]] || package_fail "path must be literal and absolute"
  [[ "$(realpath -e -- "$1" 2>/dev/null)" == "$1" ]] || package_fail "path must be canonical without symlinks"
}

package_load() {
  PACKAGE_DIR="${CARTULARY_MVP_DIR:-$(cd -- "${SCRIPT_DIR}/../../.." && pwd -P)}"
  RELEASE_DIR="${PACKAGE_DIR}/release"
  ASSET_DIR="${RELEASE_DIR}/assets"
  package_path "$PACKAGE_DIR"
  ENV_FILE="${CARTULARY_MVP_ENV_FILE:-${PACKAGE_DIR}/.env}"
  package_path "$ENV_FILE"
  package_file "$ENV_FILE"
  # Only input locations are selected by the caller. All Compose values below
  # come from this one file; neither shell evaluation nor dotenv interpolation
  # participates. Compose receives explicit exported values and /dev/null.
  local key value line
  local -A seen=()
  local -a keys=(CARTULARY_HTTP_PORT CARTULARY_PUBLIC_ORIGIN POSTGRES_DB RESTORE_VERIFY_POSTGRES_DB CARTULARY_TLS_DIR CARTULARY_AUTH_MASTER_KEY CARTULARY_RECOVERY_MASTER_KEY CARTULARY_SECRET_REVISIONS_CONFLICT_TOKEN_ACTIVE CARTULARY_S3_PRIMARY_ACCESS_KEY_ID CARTULARY_S3_PRIMARY_SECRET_ACCESS_KEY CARTULARY_S3_PRIMARY_BUCKET CARTULARY_S3_RESTORE_VERIFY_ENDPOINT CARTULARY_S3_RESTORE_VERIFY_ACCESS_KEY_ID CARTULARY_S3_RESTORE_VERIFY_SECRET_ACCESS_KEY CARTULARY_S3_RESTORE_VERIFY_BUCKET CARTULARY_MVP_COMPOSE_PROJECT_NAME CARTULARY_REFERENCE_PACKS_ENABLED)
  for key in "${keys[@]}"; do unset "$key"; done
  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line%$'\r'}"
    [[ -z "$line" || "$line" == \#* ]] && continue
    [[ "$line" =~ ^([A-Z][A-Z0-9_]*)=(.*)$ ]] || package_fail "settings require literal KEY=value lines"
    key="${BASH_REMATCH[1]}"; value="${BASH_REMATCH[2]}"
    [[ " ${keys[*]} " == *" ${key} "* ]] || package_fail "unknown package setting"
    [[ ! -v "seen[$key]" ]] || package_fail "duplicate package setting"
    [[ -n "$value" && "$value" != replace-* && "$value" != *'$'* && "$value" != *'`'* && "$value" != *$'\r'* ]] || package_fail "unresolved or executable package setting"
    seen["$key"]=1
    printf -v "$key" '%s' "$value"
    export "${key?}"
  done <"$ENV_FILE"
  for key in "${keys[@]}"; do
    case "$key" in CARTULARY_REFERENCE_PACKS_ENABLED|CARTULARY_MVP_COMPOSE_PROJECT_NAME) continue;; esac
    [[ -n "${!key:-}" ]] || package_fail "required package setting is missing"
  done
  CARTULARY_MVP_COMPOSE_PROJECT_NAME="${CARTULARY_MVP_COMPOSE_PROJECT_NAME:-cartulary-mvp}"
  CARTULARY_REFERENCE_PACKS_ENABLED="${CARTULARY_REFERENCE_PACKS_ENABLED:-false}"
  [[ "$CARTULARY_MVP_COMPOSE_PROJECT_NAME" =~ ^[a-z0-9][a-z0-9_-]{0,62}$ ]] || package_fail "invalid deployment identity"
  [[ "$POSTGRES_DB" =~ ^[A-Za-z_][A-Za-z0-9_]*$ && "$RESTORE_VERIFY_POSTGRES_DB" =~ ^[A-Za-z_][A-Za-z0-9_]*$ && "$POSTGRES_DB" != "$RESTORE_VERIFY_POSTGRES_DB" ]] || package_fail "database namespaces must be valid and distinct"
  [[ "$CARTULARY_S3_PRIMARY_BUCKET" != "$CARTULARY_S3_RESTORE_VERIFY_BUCKET" ]] || package_fail "object namespaces must be distinct"
  [[ "$CARTULARY_S3_RESTORE_VERIFY_ENDPOINT" == seaweedfs-s3:8333 ]] || package_fail "unsupported package object service"
  if [[ ! "$CARTULARY_HTTP_PORT" =~ ^[0-9]{1,5}$ ]] || ((10#$CARTULARY_HTTP_PORT == 0 || 10#$CARTULARY_HTTP_PORT > 65535)); then
    package_fail "invalid HTTPS port"
  fi
  [[ "$CARTULARY_PUBLIC_ORIGIN" == https://* ]] || package_fail "HTTPS origin is required"
  package_path "$CARTULARY_TLS_DIR"
  for key in ca.pem postgres.crt postgres.key seaweed.crt seaweed.key migration.crt migration.key recovery.crt recovery.key runtime.crt runtime.key restore-migration.crt restore-migration.key restore-recovery.crt restore-recovery.key application.crt application.key; do
    package_file "${CARTULARY_TLS_DIR}/${key}"
  done
  SOURCE_CONFIG_HOST="${PACKAGE_DIR}/config.toml"
  TARGET_CONFIG_HOST="${PACKAGE_DIR}/restore-verification-target.toml"
  SOURCE_CONFIG_CONTAINER=/etc/cartulary/config.toml
  TARGET_CONFIG_CONTAINER=/etc/cartulary/restore-verification-target.toml
  TARGET_ROOT_HOST="${PACKAGE_DIR}/runtime/restore-verification-target"
  TARGET_ROOT_CONTAINER=/var/lib/cartulary/restore-verification-target
  export TARGET_ROOT_HOST TARGET_ROOT_CONTAINER
  [[ "$(realpath -m -- "$TARGET_ROOT_HOST" 2>/dev/null)" == "$TARGET_ROOT_HOST" ]] || package_fail "target root must not traverse symlinks"
  package_file "$SOURCE_CONFIG_HOST"
  package_file "$TARGET_CONFIG_HOST"
  COMPOSE_FILE="${ASSET_DIR}/docker-compose.yml"
  package_file "$COMPOSE_FILE"
  COMPOSE_ARGS=(--project-name "$CARTULARY_MVP_COMPOSE_PROJECT_NAME" --project-directory "$ASSET_DIR" --env-file /dev/null -f "$COMPOSE_FILE")
  case "$CARTULARY_REFERENCE_PACKS_ENABLED" in
    true)
      package_file "${PACKAGE_DIR}/reference-pack-trust.json"
      COMPOSE_ARGS+=(-f "${ASSET_DIR}/docker-compose.reference-packs.yml");;
    false) ;;
    *) package_fail "invalid Reference Pack selection";;
  esac
  # Never inherit Compose's ambient file/profile/project selection.
  unset COMPOSE_FILE COMPOSE_PROFILES COMPOSE_PROJECT_NAME COMPOSE_ENV_FILES
  export COMPOSE_DISABLE_ENV_FILE=1
  export CARTULARY_INSTALLATION_DIR="$PACKAGE_DIR"
  # Verify immutable bytes and exact local image IDs before any service mutation.
  "$ASSET_DIR/scripts/release-verify.sh" "$RELEASE_DIR" images >/dev/null
  docker info >/dev/null 2>&1 || package_fail "start Docker Desktop with its WSL2 backend and retry"
  docker compose version >/dev/null 2>&1 || package_fail "Docker Compose is unavailable"
}

# One command builder owns deployment selection and operation labels, including
# bounded provisioning calls. No caller or ambient Compose overrides enter it.
compose_execute() {
  local bound="$1"
  shift
  local -a command=(docker compose "${COMPOSE_ARGS[@]}")
  if [[ "${1:-}" == run && -n "${package_operation_id:-}" ]]; then
    shift
    command+=(run --label "cartulary.package.operation=${package_operation_id}")
  fi
  command+=("$@")
  if [[ -n "$bound" ]]; then timeout "$bound" "${command[@]}"; else "${command[@]}"; fi
}
compose() { compose_execute "" "$@"; }

# Master health does not establish object metadata readiness. Repeat only this
# idempotent provisioning operation after the owner's transient attempts exhaust.
package_initialize_objects() {
  local deadline=$((SECONDS + 120)) remaining output status
  while true; do
    remaining=$((deadline - SECONDS))
    ((remaining > 0)) || return 1
    if output="$(compose_execute "$remaining" "$@" 2>&1)"; then
      printf '%s\n' "$output"
      return 0
    else status=$?; fi
    if [[ "$status" == 124 || "$output" != *'object-store init failed: reason_code=retry_exhausted'* ]] || ((SECONDS + 2 >= deadline)); then
      printf '%s\n' "$output" >&2
      return "$status"
    fi
    echo 'waiting for admitted object metadata service during provisioning' >&2
    sleep 2
  done
}

package_preflight() {
  compose run --rm --no-deps \
    --volume "${SOURCE_CONFIG_HOST}:${SOURCE_CONFIG_CONTAINER}:ro" \
    --volume "${TARGET_CONFIG_HOST}:${TARGET_CONFIG_CONTAINER}:ro" \
    --entrypoint /usr/local/bin/cartulary-operator restore-verify-operator \
    package preflight --source-config-file "$SOURCE_CONFIG_CONTAINER" --target-config-file "$TARGET_CONFIG_CONTAINER"
}
