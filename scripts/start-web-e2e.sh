#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COMPOSE_FILE="${ROOT_DIR}/docker-compose.dev.yml"
RUNTIME_ROOT_BASE="$(mktemp -d /tmp/cartulary-web-e2e-runtime-XXXXXX)"
E2E_DB="cartulary_web_e2e_$$"
E2E_DSN="postgres://cartulary:cartulary@localhost:5432/${E2E_DB}?sslmode=disable"
SERVER_LOG="/tmp/cartulary-e2e-server-$$.log"
WEB_LOG="/tmp/cartulary-e2e-web-$$.log"
MINIO_READY_URL="http://127.0.0.1:9000/minio/health/ready"

mkdir -p \
  "${RUNTIME_ROOT_BASE}/database-storage" \
  "${RUNTIME_ROOT_BASE}/object-storage" \
  "${RUNTIME_ROOT_BASE}/backup-storage" \
  "${RUNTIME_ROOT_BASE}/reference-pack-storage" \
  "${RUNTIME_ROOT_BASE}/temporary-work" \
  "${RUNTIME_ROOT_BASE}/export-outputs"

cleanup() {
  if [[ -n "${VITE_PID:-}" ]]; then
    kill "${VITE_PID}" >/dev/null 2>&1 || true
  fi
  if [[ -n "${SERVER_PID:-}" ]]; then
    kill "${SERVER_PID}" >/dev/null 2>&1 || true
  fi
  wait >/dev/null 2>&1 || true
  docker compose -f "${COMPOSE_FILE}" exec -T postgres \
    psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"${E2E_DB}\" WITH (FORCE);" >/dev/null 2>&1 || true
  rm -rf "${RUNTIME_ROOT_BASE}"
}

trap cleanup EXIT INT TERM

wait_for_http() {
  local url="$1"
  local name="$2"

  for _ in $(seq 1 180); do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    if [[ -n "${SERVER_PID:-}" ]] && ! kill -0 "${SERVER_PID}" >/dev/null 2>&1; then
      echo "backend exited before ${name} readiness" >&2
      cat "${SERVER_LOG}" >&2 || true
      return 1
    fi
    if [[ -n "${VITE_PID:-}" ]] && ! kill -0 "${VITE_PID}" >/dev/null 2>&1; then
      echo "frontend exited before ${name} readiness" >&2
      cat "${WEB_LOG}" >&2 || true
      return 1
    fi
    sleep 1
  done

  echo "timed out waiting for ${name} at ${url}" >&2
  cat "${SERVER_LOG}" >&2 || true
  cat "${WEB_LOG}" >&2 || true
  return 1
}

wait_for_postgres() {
  for _ in $(seq 1 30); do
    if docker compose -f "${COMPOSE_FILE}" exec -T postgres pg_isready -U cartulary -d postgres >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  echo "postgres did not become ready for browser e2e" >&2
  return 1
}

wait_for_minio() {
  for _ in $(seq 1 120); do
    if curl -fsS "${MINIO_READY_URL}" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  echo "minio did not become ready for browser e2e" >&2
  docker compose -f "${COMPOSE_FILE}" logs --no-color minio >&2 || true
  return 1
}

make -C "${ROOT_DIR}" frontend-toolchain
docker compose -f "${COMPOSE_FILE}" up -d postgres minio >/dev/null
wait_for_postgres
wait_for_minio
docker compose -f "${COMPOSE_FILE}" exec -T postgres \
  psql -U cartulary -d postgres -c "DROP DATABASE IF EXISTS \"${E2E_DB}\" WITH (FORCE);" >/dev/null
docker compose -f "${COMPOSE_FILE}" exec -T postgres \
  psql -U cartulary -d postgres -c "CREATE DATABASE \"${E2E_DB}\";" >/dev/null

(
  cd "${ROOT_DIR}"
  CARTULARY_CONFIG_FILE="${ROOT_DIR}/configs/dev/config.toml" \
  CARTULARY_POSTGRES_DSN="${E2E_DSN}" \
  GOCACHE=/tmp/cartulary-go-build \
  GOMODCACHE=/tmp/cartulary-go-mod \
  go run ./cmd/migrate up
)

(
  cd "${ROOT_DIR}"
  CARTULARY_CONFIG_FILE="${ROOT_DIR}/configs/dev/config.toml" \
  CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH="${ROOT_DIR}/configs/dev/bootstrap-admin.json" \
  CARTULARY_POSTGRES_DSN="${E2E_DSN}" \
  CARTULARY_ENABLE_TEST_ROUTES=1 \
  CARTULARY__ROOTS__DATABASE_STORAGE__PATH="${RUNTIME_ROOT_BASE}/database-storage" \
  CARTULARY__ROOTS__OBJECT_STORAGE__PATH="${RUNTIME_ROOT_BASE}/object-storage" \
  CARTULARY__ROOTS__BACKUP_STORAGE__PATH="${RUNTIME_ROOT_BASE}/backup-storage" \
  CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH="${RUNTIME_ROOT_BASE}/reference-pack-storage" \
  CARTULARY__ROOTS__TEMPORARY_WORK__PATH="${RUNTIME_ROOT_BASE}/temporary-work" \
  CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH="${RUNTIME_ROOT_BASE}/export-outputs" \
  GOCACHE=/tmp/cartulary-go-build \
  GOMODCACHE=/tmp/cartulary-go-mod \
  go run ./cmd/server
) >"${SERVER_LOG}" 2>&1 &
SERVER_PID=$!

(
  cd "${ROOT_DIR}"
  export PATH="${ROOT_DIR}/tmp/node-runtime/bin:${PATH}"
  corepack pnpm --dir apps/web dev --host 127.0.0.1 --port 4173 --strictPort
) >"${WEB_LOG}" 2>&1 &
VITE_PID=$!

wait_for_http "http://127.0.0.1:8080/readyz" "backend"
wait_for_http "http://127.0.0.1:4173" "frontend"

wait
