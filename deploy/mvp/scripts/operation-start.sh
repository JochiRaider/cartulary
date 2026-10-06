#!/usr/bin/env bash
# Explicit operator entry point after Windows/WSL/Docker Desktop interruption.
set -euo pipefail
SCRIPT_DIR="$(unset CDPATH && cd -- "$(dirname "$0")" && pwd)"
PACKAGE_DIR="${CARTULARY_MVP_DIR:-$(unset CDPATH && cd -- "${SCRIPT_DIR}/.." && pwd)}"
ENV_FILE="${CARTULARY_MVP_ENV_FILE:-${PACKAGE_DIR}/.env}"
COMPOSE_FILE="${CARTULARY_MVP_COMPOSE_FILE:-${PACKAGE_DIR}/docker-compose.yml}"
[[ -f "$ENV_FILE" ]] || { echo "missing package environment" >&2; exit 1; }
set -a
# shellcheck disable=SC1090
source "$ENV_FILE"
set +a
docker info >/dev/null 2>&1 || { echo "start Docker Desktop with its WSL2 backend and retry" >&2; exit 1; }
compose_args=(--env-file "$ENV_FILE" -f "$COMPOSE_FILE")
if [[ -n "${CARTULARY_MVP_COMPOSE_PROJECT_NAME:-}" ]]; then compose_args=(--project-name "$CARTULARY_MVP_COMPOSE_PROJECT_NAME" "${compose_args[@]}"); fi
if [[ -n "${CARTULARY_MVP_COMPOSE_OVERLAY:-}" ]]; then compose_args+=(-f "$CARTULARY_MVP_COMPOSE_OVERLAY"); fi
docker compose "${compose_args[@]}" up -d app
status=0
# The owner checks completion and consistency-point age, including downtime.
if ! docker compose "${compose_args[@]}" run --rm --no-deps \
  --entrypoint /usr/local/bin/cartulary-operator recovery-operator backup inspect latest; then
  echo "backup freshness check failed; inspect the reported reason, capture a current backup and retry" >&2
  status=1
fi
if ! "$SCRIPT_DIR/restore-verify-due.sh"; then status=1; fi
if [[ "$status" == 0 ]]; then
  echo "restart checks passed; explicitly start the guest backup and restore-verification timers for this operating session"
fi
exit "$status"
