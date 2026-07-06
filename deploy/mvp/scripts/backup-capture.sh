#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")" && pwd)"
PACKAGE_DIR="${CARTULARY_MVP_DIR:-$(unset CDPATH && cd -- "${SCRIPT_DIR}/.." && pwd)}"
ENV_FILE="${CARTULARY_MVP_ENV_FILE:-${PACKAGE_DIR}/.env}"
COMPOSE_FILE="${CARTULARY_MVP_COMPOSE_FILE:-${PACKAGE_DIR}/docker-compose.yml}"
SOURCE_CONFIG_HOST="${CARTULARY_SOURCE_CONFIG:-${PACKAGE_DIR}/config.toml}"
SOURCE_CONFIG_CONTAINER="${CARTULARY_SOURCE_CONFIG_CONTAINER:-/etc/cartulary/config.toml}"

fail() {
  echo "cartulary backup create failed: $*" >&2
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
docker compose version >/dev/null 2>&1 || fail "docker compose plugin is not available"
require_file "$ENV_FILE"
require_file "$SOURCE_CONFIG_HOST"

set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a

cleanup() {
  local status=$?
  compose up -d app >/dev/null 2>&1 || true
  exit "$status"
}
trap cleanup EXIT

compose stop app >/dev/null

compose run --rm --no-deps \
  --entrypoint /usr/local/bin/cartulary-operator \
  app \
  backup create \
  --source-config-file "$SOURCE_CONFIG_CONTAINER"
