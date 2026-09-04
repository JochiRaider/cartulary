#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")" && pwd)"
PACKAGE_DIR="${CARTULARY_MVP_DIR:-$(unset CDPATH && cd -- "${SCRIPT_DIR}/.." && pwd)}"
ENV_FILE="${CARTULARY_MVP_ENV_FILE:-${PACKAGE_DIR}/.env}"
COMPOSE_FILE="${CARTULARY_MVP_COMPOSE_FILE:-${PACKAGE_DIR}/docker-compose.yml}"
SOURCE_CONFIG_HOST="${CARTULARY_SOURCE_CONFIG:-${PACKAGE_DIR}/config.toml}"
TARGET_CONFIG_HOST="${CARTULARY_RESTORE_VERIFY_TARGET_CONFIG:-${PACKAGE_DIR}/restore-verification-target.toml}"
TARGET_ROOT_HOST="${CARTULARY_RESTORE_VERIFY_TARGET_ROOT:-${PACKAGE_DIR}/runtime/restore-verification-target}"
SOURCE_CONFIG_CONTAINER="${CARTULARY_SOURCE_CONFIG_CONTAINER:-/etc/cartulary/config.toml}"
TARGET_CONFIG_CONTAINER="${CARTULARY_RESTORE_VERIFY_TARGET_CONFIG_CONTAINER:-/etc/cartulary/restore-verification-target.toml}"
TARGET_ROOT_CONTAINER="${CARTULARY_RESTORE_VERIFY_TARGET_ROOT_CONTAINER:-/var/lib/cartulary/restore-verification-target}"
OBJECT_INIT_OUTPUT="${CARTULARY_RESTORE_VERIFY_OBJECT_INIT_OUTPUT:-/dev/null}"

fail() {
  echo "cartulary restore verification failed: $*" >&2
  exit 1
}

require_command() {
  local name="$1"
  if ! command -v "$name" >/dev/null 2>&1; then
    fail "missing required command ${name}"
  fi
}

require_file() {
  local path="$1"
  if [[ ! -f "$path" ]]; then
    fail "missing required file ${path}"
  fi
}

compose() {
  local args=(--env-file "$ENV_FILE" -f "$COMPOSE_FILE")
  if [[ -n "${CARTULARY_MVP_COMPOSE_PROJECT_NAME:-}" ]]; then
    args=(--project-name "$CARTULARY_MVP_COMPOSE_PROJECT_NAME" "${args[@]}")
  fi
  docker compose "${args[@]}" "$@"
}

require_command docker
require_command date
require_command sha256sum
docker compose version >/dev/null 2>&1 || fail "docker compose plugin is not available"
require_file "$ENV_FILE"
require_file "$SOURCE_CONFIG_HOST"
require_file "$TARGET_CONFIG_HOST"

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

target_db="${RESTORE_VERIFY_POSTGRES_DB:-cartulary_restore_verify}"
if [[ ! "$target_db" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then
  fail "RESTORE_VERIFY_POSTGRES_DB must be a simple database identifier"
fi
for name in \
  CARTULARY_S3_RESTORE_VERIFY_ENDPOINT \
  CARTULARY_S3_RESTORE_VERIFY_ACCESS_KEY_ID \
  CARTULARY_S3_RESTORE_VERIFY_SECRET_ACCESS_KEY \
  CARTULARY_S3_RESTORE_VERIFY_BUCKET; do
  if [[ -z "${!name:-}" || "${!name:-}" == replace-* ]]; then
    fail "set ${name} in ${ENV_FILE}"
  fi
done

if ! compose exec -T postgres psql -U "${POSTGRES_USER:-cartulary}" -d postgres -Atc "SELECT 1 FROM pg_database WHERE datname = '${target_db}'" | grep -qx "1"; then
  compose exec -T postgres createdb -U "${POSTGRES_USER:-cartulary}" "$target_db"
fi
compose exec -T -e "PGDATABASE=${target_db}" postgres /docker-entrypoint-initdb.d/010-cartulary-provision.sh >/dev/null

compose run --rm --no-deps \
  --volume "${TARGET_CONFIG_HOST}:${TARGET_CONFIG_CONTAINER}:ro" \
  --volume "${TARGET_ROOT_HOST}:${TARGET_ROOT_CONTAINER}" \
  --entrypoint /usr/local/bin/cartulary-migrate \
  restore-verify-migrate up

compose run --rm --no-deps \
  --volume "${TARGET_CONFIG_HOST}:${TARGET_CONFIG_CONTAINER}:ro" \
  --volume "${TARGET_ROOT_HOST}:${TARGET_ROOT_CONTAINER}" \
  --entrypoint /usr/local/bin/cartulary-operator \
  restore-verify-operator object-store init -config "$TARGET_CONFIG_CONTAINER" >"$OBJECT_INIT_OUTPUT"

generation_source="/proc/sys/kernel/random/uuid"
if [[ ! -r "$generation_source" ]]; then
  fail "target generation source is unavailable"
fi
target_generation_id="$(tr -d '\r\n' <"$generation_source")"
if [[ ! "$target_generation_id" =~ ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$ ]]; then
  fail "target generation source returned an invalid UUID"
fi
database_binding_identity="${CARTULARY_RESTORE_VERIFY_DATABASE_BINDING_IDENTITY:-managed_service:restore_verify}"
object_binding_identity="${CARTULARY_RESTORE_VERIFY_OBJECT_BINDING_IDENTITY:-managed_service:restore_verify}"
database_binding_sha256="$(printf '%s' "$database_binding_identity" | sha256sum | awk '{print $1}')"
object_binding_sha256="$(printf '%s' "$object_binding_identity" | sha256sum | awk '{print $1}')"
issued_at="$(date -u '+%Y-%m-%dT%H:%M:%SZ')"
expires_at="$(date -u -d '+23 hours' '+%Y-%m-%dT%H:%M:%SZ')"
mkdir -p "${TARGET_ROOT_HOST}/backups"
chmod 0755 "${TARGET_ROOT_HOST}/backups"
printf '%s\n' "$target_generation_id" >"${TARGET_ROOT_HOST}/backups/restore-target-generation"
printf '{"schema_id":"cartulary.restore_target_marker.v2","purpose":"restore_verification_target","target_generation_id":"%s","binding_digests":{"database_sha256":"%s","object_store_sha256":"%s"},"issued_at":"%s","expires_at":"%s"}\n' \
  "$target_generation_id" \
  "$database_binding_sha256" \
  "$object_binding_sha256" \
  "$issued_at" \
  "$expires_at" >"${TARGET_ROOT_HOST}/backups/restore-target-marker.json"
chmod 0644 \
  "${TARGET_ROOT_HOST}/backups/restore-target-generation" \
  "${TARGET_ROOT_HOST}/backups/restore-target-marker.json"

compose run --rm --no-deps \
  --volume "${SOURCE_CONFIG_HOST}:${SOURCE_CONFIG_CONTAINER}:ro" \
  --volume "${TARGET_CONFIG_HOST}:${TARGET_CONFIG_CONTAINER}:ro" \
  --volume "${TARGET_ROOT_HOST}:${TARGET_ROOT_CONTAINER}" \
  --entrypoint /usr/local/bin/cartulary-operator \
  restore-verify-operator restore-verify due \
  --source-config-file "$SOURCE_CONFIG_CONTAINER" \
  --target-config-file "$TARGET_CONFIG_CONTAINER"
