#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RUNTIME_ROOT_BASE="/tmp/cartulary-web-e2e-runtime"

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
      cat /tmp/cartulary-e2e-server.log >&2 || true
      return 1
    fi
    if [[ -n "${VITE_PID:-}" ]] && ! kill -0 "${VITE_PID}" >/dev/null 2>&1; then
      echo "frontend exited before ${name} readiness" >&2
      cat /tmp/cartulary-e2e-web.log >&2 || true
      return 1
    fi
    sleep 1
  done

  echo "timed out waiting for ${name} at ${url}" >&2
  cat /tmp/cartulary-e2e-server.log >&2 || true
  cat /tmp/cartulary-e2e-web.log >&2 || true
  return 1
}

make -C "${ROOT_DIR}" frontend-toolchain
make -C "${ROOT_DIR}" db-up
make -C "${ROOT_DIR}" db-reset

(
  cd "${ROOT_DIR}"
  CARTULARY_CONFIG_FILE="${ROOT_DIR}/configs/dev/config.toml" \
  CARTULARY__BOOTSTRAP__FIRST_ADMIN_MANIFEST_PATH="${ROOT_DIR}/configs/dev/bootstrap-admin.json" \
  CARTULARY__ROOTS__DATABASE_STORAGE__PATH="${RUNTIME_ROOT_BASE}/database-storage" \
  CARTULARY__ROOTS__OBJECT_STORAGE__PATH="${RUNTIME_ROOT_BASE}/object-storage" \
  CARTULARY__ROOTS__BACKUP_STORAGE__PATH="${RUNTIME_ROOT_BASE}/backup-storage" \
  CARTULARY__ROOTS__REFERENCE_PACK_STORAGE__PATH="${RUNTIME_ROOT_BASE}/reference-pack-storage" \
  CARTULARY__ROOTS__TEMPORARY_WORK__PATH="${RUNTIME_ROOT_BASE}/temporary-work" \
  CARTULARY__ROOTS__EXPORT_OUTPUTS__PATH="${RUNTIME_ROOT_BASE}/export-outputs" \
  GOCACHE=/tmp/cartulary-go-build \
  GOMODCACHE=/tmp/cartulary-go-mod \
  go run ./cmd/server
) >/tmp/cartulary-e2e-server.log 2>&1 &
SERVER_PID=$!

(
  cd "${ROOT_DIR}"
  export PATH="${ROOT_DIR}/tmp/node-runtime/bin:${PATH}"
  corepack pnpm --dir apps/web dev --host 127.0.0.1 --port 4173 --strictPort
) >/tmp/cartulary-e2e-web.log 2>&1 &
VITE_PID=$!

wait_for_http "http://127.0.0.1:8080/readyz" "backend"
wait_for_http "http://127.0.0.1:4173" "frontend"

wait
